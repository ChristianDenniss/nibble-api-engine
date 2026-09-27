package promotions

import (
	"testing"
	"time"
)

func TestParseTextDeal(t *testing.T) {
	at := time.Date(2026, 9, 26, 9, 0, 0, 0, time.FixedZone("ADT", -3*60*60))
	deal, err := ParseTextDeal(TextDealInput{Text: "Use code XYZ123 for 40% off all orders TODAY only on Skip the Dishes", ReceivedAt: at})
	if err != nil { t.Fatal(err) }
	if deal.Promotion.ChannelID != "ch_skip" || deal.Promotion.Kind != "percent_off" || deal.Promotion.ValueBPS != 4000 { t.Fatalf("unexpected promotion: %#v", deal.Promotion) }
	if deal.Constraint.Code != "XYZ123" { t.Fatalf("expected code, got %q", deal.Constraint.Code) }
	if !deal.Promotion.EndsAt.Equal(time.Date(2026, 9, 26, 23, 59, 59, 0, at.Location())) { t.Fatalf("unexpected expiry: %s", deal.Promotion.EndsAt) }
}

func TestParseTextDealRequiresDiscount(t *testing.T) {
	_, err := ParseTextDeal(TextDealInput{ChannelID: "ch_skip", Text: "Use code ABCD on Skip", ReceivedAt: time.Now()})
	if err == nil { t.Fatal("expected unsupported message to be rejected") }
}
