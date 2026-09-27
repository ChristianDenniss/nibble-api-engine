package promotions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
	promotionentity "github.com/ChristianDenniss/go-data-model/promotion/entity"
)

var (
	codePattern    = regexp.MustCompile(`(?i)\b(?:(?:use|with|using)\s+(?:promo(?:tion)?\s+)?code|promo(?:tion)?\s+code|code)\s*[:#-]?\s*([A-Z0-9][A-Z0-9_-]{3,24})\b`)
	percentPattern = regexp.MustCompile(`(?i)\b(\d{1,3})\s*%\s*(?:off|discount)\b`)
	amountPattern  = regexp.MustCompile(`(?i)(?:\$|CAD\s*)(\d+(?:\.\d{1,2})?)\s*off\b`)
	channelPattern = regexp.MustCompile(`(?i)\b(skip(?:\s+the\s+dishes)?|doordash|uber\s*eats|instacart|fantuan)\b`)
)

// TextDealInput is the transport shape for daily SMS/email/chat collectors.
// Scope metadata is optional; it is intentionally explicit so a parser never
// silently guesses a restaurant or region from free text.
type TextDealInput struct {
	Source           string                   `json:"source"`
	ChannelID        string                   `json:"channelId"`
	Text             string                   `json:"text"`
	ReceivedAt       time.Time                `json:"receivedAt"`
	Targets          []promotionentity.Target `json:"targets"`
	FulfillmentMode  string                   `json:"fulfillmentMode"`
	ExpiryPolicy     string                   `json:"expiryPolicy"`
	FixedExpiryHours int                      `json:"fixedExpiryHours"`
}

type ParsedTextDeal struct {
	Promotion  promotionentity.Promotion  `json:"promotion"`
	Constraint promotionentity.Constraint `json:"constraint"`
	Targets    []promotionentity.Target   `json:"targets"`
	Source     string                     `json:"source"`
	RawText    string                     `json:"rawText"`
	Warnings   []string                   `json:"warnings"`
}

func parseTextDeal(input TextDealInput) (ParsedTextDeal, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return ParsedTextDeal{}, fmt.Errorf("text is required")
	}
	at := input.ReceivedAt
	if at.IsZero() {
		at = time.Now()
	}
	channelID := strings.TrimSpace(input.ChannelID)
	if channelID == "" {
		match := channelPattern.FindStringSubmatch(text)
		if len(match) == 0 {
			return ParsedTextDeal{}, fmt.Errorf("channelId is required when the message does not mention a platform")
		}
		channelID = channelSlug(match[1])
	}

	endsAt := endOfDealDay(at)
	if input.ExpiryPolicy == "fixed" && input.FixedExpiryHours > 0 {
		endsAt = at.Add(time.Duration(input.FixedExpiryHours) * time.Hour)
	}
	p := promotionentity.Promotion{
		ID: stableTextDealID(channelID, text, at), ChannelID: channelID,
		Name: firstSentence(text), Description: text, StartsAt: at,
		EndsAt: endsAt, Value: moneyForCurrency("CAD"),
	}
	warnings := []string{}
	if match := percentPattern.FindStringSubmatch(text); len(match) > 1 {
		value, _ := strconv.Atoi(match[1])
		if value > 100 {
			return ParsedTextDeal{}, fmt.Errorf("percentage discount cannot exceed 100")
		}
		p.Kind, p.ValueBPS = promotionentity.KindPercentOff, value*100
	} else if match := amountPattern.FindStringSubmatch(text); len(match) > 1 {
		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return ParsedTextDeal{}, fmt.Errorf("invalid amount discount")
		}
		p.Kind, p.Value.AmountCents = promotionentity.KindAmountOff, int64(value*100)
	} else if strings.Contains(strings.ToLower(text), "free delivery") || strings.Contains(strings.ToLower(text), "free shipping") {
		p.Kind = promotionentity.KindFreeDelivery
	} else {
		return ParsedTextDeal{}, fmt.Errorf("could not find a supported discount: percentage, amount off, or free delivery")
	}
	if match := codePattern.FindStringSubmatch(text); len(match) > 1 {
		return withCode(ParsedTextDeal{Promotion: p, Constraint: promotionentity.Constraint{ID: p.ID + "_constraint", PromotionID: p.ID, Code: strings.ToUpper(match[1])}, Targets: input.Targets, Source: input.Source, RawText: text, Warnings: warnings})
	}
	warnings = append(warnings, "no promo code found; this deal will be displayed as an offer")
	return ParsedTextDeal{Promotion: p, Targets: input.Targets, Source: input.Source, RawText: text, Warnings: warnings}, nil
}

func withCode(parsed ParsedTextDeal) (ParsedTextDeal, error) { return parsed, nil }

// ParseTextDeal is the reusable normalizer used by SMS and non-SMS collectors.
func ParseTextDeal(input TextDealInput) (ParsedTextDeal, error) { return parseTextDeal(input) }
func moneyForCurrency(currency string) money.Money              { return money.Money{Currency: currency} }

func endOfDealDay(at time.Time) time.Time {
	y, m, d := at.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, at.Location())
}
func firstSentence(text string) string {
	if i := strings.IndexAny(text, ".!\n"); i > 0 {
		text = text[:i]
	}
	if len(text) > 120 {
		return text[:120]
	}
	return text
}
func channelSlug(value string) string {
	v := strings.ToLower(strings.ReplaceAll(value, " ", ""))
	if strings.Contains(v, "skip") {
		return "ch_skip"
	}
	if strings.Contains(v, "door") {
		return "ch_doordash"
	}
	if strings.Contains(v, "uber") {
		return "ch_ubereats"
	}
	if strings.Contains(v, "instacart") {
		return "ch_instacart"
	}
	return "ch_" + v
}
func stableTextDealID(channel, text string, at time.Time) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(text), " "))
	sum := 0
	for _, r := range normalized {
		sum = (sum*31 + int(r)) % 1000000007
	}
	return fmt.Sprintf("text_%s_%s_%d", channel, at.Format("20060102"), sum)
}
