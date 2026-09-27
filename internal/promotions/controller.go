package promotions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	money "github.com/ChristianDenniss/go-data-model/money/entity"
	promotionentity "github.com/ChristianDenniss/go-data-model/promotion/entity"
	promotionsvc "github.com/ChristianDenniss/go-data-model/promotion/service"
)

// Controller is the small, idempotent insertion boundary used by deal
// collectors. A collector can send a single payload and the service stores
// the promotion, its code/limits, and every scope target together.
type Controller struct{ svc *promotionsvc.Service }

func NewController(svc *promotionsvc.Service) *Controller { return &Controller{svc: svc} }

type insertRequest struct {
	ID              string                      `json:"id"`
	ChannelID       string                      `json:"channelId"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	Kind            string                      `json:"kind"`
	FulfillmentMode string                      `json:"fulfillmentMode"`
	ValueCents      int64                       `json:"valueCents"`
	Currency        string                      `json:"currency"`
	ValueBPS        int                         `json:"valueBPS"`
	StartsAt        time.Time                   `json:"startsAt"`
	EndsAt          time.Time                   `json:"endsAt"`
	Constraint      *promotionentity.Constraint `json:"constraint"`
	Targets         []promotionentity.Target    `json:"targets"`
}

func (c *Controller) InsertText(w http.ResponseWriter, r *http.Request) {
	var input TextDealInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid json")
		return
	}
	parsed, err := parseTextDeal(input)
	if err != nil {
		httpx.WriteFailure(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if parsed.Promotion.FulfillmentMode == "" {
		parsed.Promotion.FulfillmentMode = input.FulfillmentMode
	}
	if err := c.upsert(r, parsed); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, parsed)
}

func (c *Controller) Insert(w http.ResponseWriter, r *http.Request) {
	var body insertRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" || body.ChannelID == "" || body.Kind == "" || body.EndsAt.IsZero() {
		httpx.WriteFailure(w, http.StatusBadRequest, "id, channelId, kind, and endsAt are required")
		return
	}
	if body.Currency == "" {
		body.Currency = "CAD"
	}
	if body.StartsAt.IsZero() {
		body.StartsAt = time.Now()
	}
	p := promotionentity.Promotion{ID: body.ID, ChannelID: body.ChannelID, Name: body.Name, Description: body.Description, Kind: body.Kind, FulfillmentMode: body.FulfillmentMode, Value: money.Money{AmountCents: body.ValueCents, Currency: body.Currency}, ValueBPS: body.ValueBPS, StartsAt: body.StartsAt, EndsAt: body.EndsAt}
	parsed := ParsedTextDeal{Promotion: p, Targets: body.Targets}
	if body.Constraint != nil {
		parsed.Constraint = *body.Constraint
	}
	if err := c.upsert(r, parsed); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"id": p.ID, "status": "upserted"})
}

func (c *Controller) upsert(r *http.Request, parsed ParsedTextDeal) error {
	return c.StoreParsed(r.Context(), parsed)
}

func (c *Controller) StoreParsed(ctx context.Context, parsed ParsedTextDeal) error {
	p := parsed.Promotion
	if err := validatePromotion(p); err != nil {
		return err
	}
	if err := c.svc.RecordPromotion(ctx, p); err != nil {
		return err
	}
	if parsed.Constraint.ID != "" {
		parsed.Constraint.PromotionID = p.ID
		if err := c.svc.RecordConstraint(ctx, parsed.Constraint); err != nil {
			return err
		}
	}
	for i, target := range parsed.Targets {
		target.PromotionID = p.ID
		if target.ID == "" {
			target.ID = p.ID + "_target_" + strconv.Itoa(i)
		}
		if err := c.svc.RecordTarget(ctx, target); err != nil {
			return err
		}
	}
	return nil
}

func validatePromotion(p promotionentity.Promotion) error {
	if p.ID == "" || p.ChannelID == "" {
		return errors.New("promotion id and channelId are required")
	}
	if p.EndsAt.IsZero() || !p.EndsAt.After(p.StartsAt) {
		return errors.New("promotion must expire after it starts")
	}
	switch p.Kind {
	case promotionentity.KindPercentOff:
		if p.ValueBPS < 1 || p.ValueBPS > 10000 {
			return errors.New("percentage discount must be between 1% and 100%")
		}
	case promotionentity.KindAmountOff:
		if p.Value.AmountCents < 1 {
			return errors.New("amount discount must be positive")
		}
	case promotionentity.KindFreeDelivery:
	case promotionentity.KindFixedPrice:
	default:
		return errors.New("unsupported promotion kind")
	}
	return nil
}
