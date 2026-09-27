package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/promotions"
	"github.com/ChristianDenniss/api-engine/internal/session"
	postgres "github.com/ChristianDenniss/go-data-store"
)

type Controller struct {
	db         *postgres.DB
	promotions *promotions.Controller
}

func NewController(db *postgres.DB, promotionController *promotions.Controller) *Controller {
	return &Controller{db: db, promotions: promotionController}
}

func (c *Controller) RegisterSource(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	var body postgres.DealSMSSource
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PhoneNumber == "" || body.ProviderID == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "phoneNumber and providerId are required")
		return
	}
	if body.ExpiryPolicy != "fixed" {
		body.ExpiryPolicy = "parsed"
	}
	if body.FixedExpiryHours < 1 {
		body.FixedExpiryHours = 24
	}
	if err := c.db.UpsertDealSMSSource(r.Context(), body); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"phoneNumber": body.PhoneNumber, "providerId": body.ProviderID})
}

func (c *Controller) ListSources(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	sources, err := c.db.ListDealSMSSources(r.Context())
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not load SMS sources")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sources": sources})
}

func (c *Controller) ListMessages(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	messages, err := c.db.ListDealSMSMessages(r.Context(), 100)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not load SMS messages")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (c *Controller) RetryMessage(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	externalID := r.PathValue("externalID")
	message, err := c.db.GetDealSMSMessage(r.Context(), externalID)
	if err != nil {
		httpx.WriteFailure(w, http.StatusNotFound, "SMS message not found")
		return
	}
	source, err := c.db.GetDealSMSSource(r.Context(), message.ToNumber)
	if err != nil || !source.Active {
		httpx.WriteFailure(w, http.StatusBadRequest, "no active source for this message")
		return
	}
	parsed, parseErr := promotions.ParseTextDeal(promotions.TextDealInput{Source: "twilio", ChannelID: source.ProviderID, Text: message.Body, ReceivedAt: message.ReceivedAt, ExpiryPolicy: source.ExpiryPolicy, FixedExpiryHours: source.FixedExpiryHours})
	if parseErr != nil {
		_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "parse_failed", "", parseErr.Error())
		httpx.WriteFailure(w, http.StatusUnprocessableEntity, parseErr.Error())
		return
	}
	if err := c.promotions.StoreParsed(r.Context(), parsed); err != nil {
		_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "store_failed", parsed.Promotion.ID, err.Error())
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "processed", parsed.Promotion.ID, "")
	httpx.WriteJSON(w, http.StatusAccepted, parsed)
}

// ReceiveTwilio accepts Twilio's application/x-www-form-urlencoded inbound
// webhook. Raw messages are persisted before parsing so no code is lost when
// a parser rule needs to be improved later.
func (c *Controller) ReceiveTwilio(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid form")
		return
	}
	if secret := os.Getenv("TWILIO_AUTH_TOKEN"); secret != "" && !validTwilioSignature(r, secret) {
		httpx.WriteFailure(w, http.StatusUnauthorized, "invalid signature")
		return
	}
	externalID, from, to, body := r.FormValue("MessageSid"), r.FormValue("From"), r.FormValue("To"), strings.TrimSpace(r.FormValue("Body"))
	if externalID == "" || from == "" || to == "" || body == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "MessageSid, From, To, and Body are required")
		return
	}
	source, err := c.db.GetDealSMSSource(r.Context(), to)
	providerID := ""
	status := "unmatched_source"
	if err == nil && source.Active {
		providerID, status = source.ProviderID, "received"
	}
	message := postgres.DealSMSMessage{ID: "sms_" + externalID, ExternalID: externalID, FromNumber: from, ToNumber: to, ProviderID: providerID, Body: body, ReceivedAt: time.Now(), Status: status}
	if err := c.db.InsertDealSMSMessage(r.Context(), message); err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not log message")
		return
	}
	if providerID != "" {
		parsed, parseErr := promotions.ParseTextDeal(promotions.TextDealInput{Source: "twilio", ChannelID: providerID, Text: body, ReceivedAt: message.ReceivedAt, ExpiryPolicy: source.ExpiryPolicy, FixedExpiryHours: source.FixedExpiryHours})
		if parseErr != nil {
			_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "parse_failed", "", parseErr.Error())
		} else if err := c.promotions.StoreParsed(r.Context(), parsed); err != nil {
			_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "store_failed", parsed.Promotion.ID, err.Error())
		} else {
			_ = c.db.UpdateDealSMSMessage(r.Context(), externalID, "processed", parsed.Promotion.ID, "")
		}
	}
	w.Header().Set("Content-Type", "application/xml")
	_, _ = w.Write([]byte("<Response></Response>"))
}

func (c *Controller) Send(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	var body struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.To == "" || body.Message == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "to and message are required")
		return
	}
	if err := SendTwilio(r.Context(), body.To, body.Message); err != nil {
		httpx.WriteFailure(w, http.StatusBadGateway, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "sent"})
}

func (c *Controller) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	accountID := session.AccountID(r.Context())
	admin, err := postgres.NewAccountRepository(c.db).IsAdmin(r.Context(), accountID)
	if accountID == "" || err != nil || !admin {
		httpx.WriteFailure(w, http.StatusForbidden, "admin access required")
		return false
	}
	return true
}

func validTwilioSignature(r *http.Request, authToken string) bool {
	signature := r.Header.Get("X-Twilio-Signature")
	if signature == "" {
		return false
	}
	urlValue := "https://" + r.Host + r.URL.RequestURI()
	keys := make([]string, 0, len(r.PostForm))
	for key := range r.PostForm {
		keys = append(keys, key)
	}
	sortStrings(keys)
	for _, key := range keys {
		urlValue += key + r.PostForm.Get(key)
	}
	h := hmac.New(sha1.New, []byte(authToken))
	_, _ = h.Write([]byte(urlValue))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

// SendTwilio sends an outbound subscription/test message using Twilio's REST
// API. Credentials stay server-side; the UI never receives the auth token.
func SendTwilio(ctx context.Context, to, body string) error {
	account, token, from := os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN"), os.Getenv("TWILIO_FROM_NUMBER")
	if account == "" || token == "" || from == "" {
		return errors.New("Twilio credentials are not configured")
	}
	form := url.Values{"To": {to}, "From": {from}, "Body": {body}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.twilio.com/2010-04-01/Accounts/"+account+"/Messages.json", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(account, token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("Twilio returned %s", res.Status)
	}
	return nil
}
