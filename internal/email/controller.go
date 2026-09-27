package email

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/promotions"
	"github.com/ChristianDenniss/api-engine/internal/session"
	postgres "github.com/ChristianDenniss/go-data-store"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const gmailScope = "https://www.googleapis.com/auth/gmail.readonly"

type Config struct {
	ClientID     string
	ClientSecret string
	CallbackBase string
	WebBase      string
}

func ConfigFromEnv() Config {
	return Config{ClientID: os.Getenv("GOOGLE_CLIENT_ID"), ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), CallbackBase: strings.TrimRight(getenv("OAUTH_CALLBACK_BASE_URL", "http://localhost:5174/api"), "/"), WebBase: strings.TrimRight(os.Getenv("WEB_BASE_URL"), "/")}
}

type Controller struct {
	db          *postgres.DB
	promotions  *promotions.Controller
	config      Config
	adminLookup *postgres.AccountRepository
}

func NewController(db *postgres.DB, promotionController *promotions.Controller, config Config) *Controller {
	return &Controller{db: db, promotions: promotionController, config: config, adminLookup: postgres.NewAccountRepository(db)}
}

func (c *Controller) oauthConfig() *oauth2.Config {
	return &oauth2.Config{ClientID: c.config.ClientID, ClientSecret: c.config.ClientSecret, Endpoint: google.Endpoint, RedirectURL: c.config.CallbackBase + "/v1/email/gmail/callback", Scopes: []string{gmailScope}}
}

func (c *Controller) Start(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	if c.config.ClientID == "" || c.config.ClientSecret == "" {
		httpx.WriteFailure(w, http.StatusServiceUnavailable, "Google OAuth is not configured")
		return
	}
	state, err := randomState()
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not create OAuth state")
		return
	}
	if err := c.db.CreateGmailOAuthState(r.Context(), postgres.GmailOAuthState{State: state, AccountID: session.AccountID(r.Context()), ExpiresAt: time.Now().Add(10 * time.Minute)}); err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not create OAuth state")
		return
	}
	http.Redirect(w, r, c.oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), http.StatusFound)
}

func (c *Controller) Callback(w http.ResponseWriter, r *http.Request) {
	state, err := c.db.ConsumeGmailOAuthState(r.Context(), r.URL.Query().Get("state"), time.Now())
	if err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid or expired Gmail OAuth state")
		return
	}
	if code := r.URL.Query().Get("error"); code != "" {
		c.redirect(w, r, "gmail_error")
		return
	}
	token, err := c.oauthConfig().Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil || token.RefreshToken == "" {
		c.redirect(w, r, "gmail_token_error")
		return
	}
	client := c.oauthConfig().Client(r.Context(), token)
	service, err := gmail.NewService(r.Context(), option.WithHTTPClient(client))
	if err != nil {
		c.redirect(w, r, "gmail_service_error")
		return
	}
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		c.redirect(w, r, "gmail_profile_error")
		return
	}
	refresh, err := encrypt(token.RefreshToken)
	if err != nil {
		c.redirect(w, r, "gmail_token_storage_error")
		return
	}
	if err := c.db.UpsertGmailConnection(r.Context(), postgres.GmailConnection{AccountID: state.AccountID, Email: profile.EmailAddress, RefreshToken: refresh, Active: true}); err != nil {
		c.redirect(w, r, "gmail_storage_error")
		return
	}
	c.redirect(w, r, "gmail_connected")
}

func (c *Controller) Status(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	connection, err := c.db.GmailConnection(r.Context(), session.AccountID(r.Context()))
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"connected": false})
		return
	}
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not load Gmail connection")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"connected": connection.Active, "email": connection.Email, "lastSyncAt": connection.LastSyncAt})
}

func (c *Controller) Sync(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	accountID := session.AccountID(r.Context())
	connection, err := c.db.GmailConnection(r.Context(), accountID)
	if err != nil || !connection.Active {
		httpx.WriteFailure(w, http.StatusBadRequest, "Gmail is not connected")
		return
	}
	refresh, err := decrypt(connection.RefreshToken)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not unlock Gmail connection")
		return
	}
	token := &oauth2.Token{RefreshToken: refresh}
	service, err := gmail.NewService(r.Context(), option.WithHTTPClient(c.oauthConfig().Client(r.Context(), token)))
	if err != nil {
		httpx.WriteFailure(w, http.StatusBadGateway, "could not connect to Gmail")
		return
	}
	list, err := service.Users.Messages.List("me").Q("is:inbox newer_than:30d").MaxResults(25).Do()
	if err != nil {
		httpx.WriteFailure(w, http.StatusBadGateway, "could not list Gmail messages")
		return
	}
	processed, failed := 0, 0
	for _, summary := range list.Messages {
		message, getErr := service.Users.Messages.Get("me", summary.Id).Format("full").Do()
		if getErr != nil {
			failed++
			continue
		}
		parsed := extractMessage(message)
		row := postgres.DealEmailMessage{ID: "email_" + summary.Id, AccountID: accountID, ExternalID: summary.Id, Sender: parsed.sender, Subject: parsed.subject, Body: parsed.body, ReceivedAt: parsed.receivedAt, Status: "received"}
		inserted, insertErr := c.db.InsertDealEmailMessage(r.Context(), row)
		if insertErr != nil {
			failed++
			continue
		}
		if !inserted {
			continue
		}
		deal, parseErr := promotions.ParseTextDeal(promotions.TextDealInput{Source: "gmail", ChannelID: "", Text: parsed.subject + "\n" + parsed.body, ReceivedAt: parsed.receivedAt})
		if parseErr != nil {
			_ = c.db.UpdateDealEmailMessage(r.Context(), accountID, summary.Id, "parse_failed", "", parseErr.Error())
			failed++
			continue
		}
		if storeErr := c.promotions.StoreParsed(r.Context(), deal); storeErr != nil {
			_ = c.db.UpdateDealEmailMessage(r.Context(), accountID, summary.Id, "store_failed", deal.Promotion.ID, storeErr.Error())
			failed++
			continue
		}
		_ = c.db.UpdateDealEmailMessage(r.Context(), accountID, summary.Id, "processed", deal.Promotion.ID, "")
		processed++
	}
	now := time.Now()
	_ = c.db.SetGmailLastSync(r.Context(), accountID, now)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"processed": processed, "failed": failed, "checked": len(list.Messages), "syncedAt": now})
}

func (c *Controller) ListMessages(w http.ResponseWriter, r *http.Request) {
	if !c.requireAdmin(w, r) {
		return
	}
	messages, err := c.db.ListDealEmailMessages(r.Context(), session.AccountID(r.Context()), 50)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "could not load email messages")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (c *Controller) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	accountID := session.AccountID(r.Context())
	admin, err := c.adminLookup.IsAdmin(r.Context(), accountID)
	if accountID == "" || err != nil || !admin {
		httpx.WriteFailure(w, http.StatusForbidden, "admin access required")
		return false
	}
	return true
}

func (c *Controller) redirect(w http.ResponseWriter, r *http.Request, result string) {
	target := c.config.WebBase + "/admin?gmail=" + result
	if c.config.WebBase == "" {
		target = "/admin?gmail=" + result
	}
	http.Redirect(w, r, target, http.StatusFound)
}

type extractedMessage struct {
	sender, subject, body string
	receivedAt            time.Time
}

func extractMessage(message *gmail.Message) extractedMessage {
	out := extractedMessage{receivedAt: time.UnixMilli(message.InternalDate)}
	for _, header := range message.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "from":
			out.sender = header.Value
		case "subject":
			out.subject = header.Value
		}
	}
	out.body = plainTextPart(message.Payload)
	if out.receivedAt.IsZero() {
		out.receivedAt = time.Now()
	}
	return out
}

func plainTextPart(part *gmail.MessagePart) string {
	if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
		raw, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return strings.TrimSpace(string(raw))
		}
	}
	for _, child := range part.Parts {
		if body := plainTextPart(child); body != "" {
			return body
		}
	}
	if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
		raw, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return stripHTML(string(raw))
		}
	}
	return ""
}

var tagPattern = regexp.MustCompile(`<[^>]*>`)

func stripHTML(value string) string {
	return strings.TrimSpace(html.UnescapeString(tagPattern.ReplaceAllString(value, " ")))
}
func randomState() (string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func encrypt(value string) (string, error) {
	block, err := cipherBlock()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, block.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := block.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}
func decrypt(value string) (string, error) {
	block, err := cipherBlock()
	if err != nil {
		return "", err
	}
	raw, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil || len(raw) < block.NonceSize() {
		return "", errors.New("invalid encrypted token")
	}
	nonce, ciphertext := raw[:block.NonceSize()], raw[block.NonceSize():]
	plain, err := block.Open(nil, nonce, ciphertext, nil)
	return string(plain), err
}
func cipherBlock() (cipher.AEAD, error) {
	key := os.Getenv("GMAIL_TOKEN_ENCRYPTION_KEY")
	raw, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(raw) != 32 {
		return nil, fmt.Errorf("GMAIL_TOKEN_ENCRYPTION_KEY must be base64 for 32 bytes")
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
