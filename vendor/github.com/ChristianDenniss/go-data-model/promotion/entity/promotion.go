package entity

import (
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
)

type Promotion struct {
	ID        string
	ChannelID string
	Name      string
	Kind      string
	Value     money.Money
	ValueBPS  int
	StartsAt  time.Time
	EndsAt    time.Time
}

type Constraint struct {
	ID                 string
	PromotionID        string
	MinSubtotalCents   int64
	Code               string
	MembershipRequired bool
	MaxDiscountCents   int64
}

type Target struct {
	ID            string
	PromotionID   string
	PlaceID       string
	SourceStoreID string
	SourceItemID  string
	DishID        string
	BrandID       string
}

type MembershipProduct struct {
	ID        string
	ChannelID string
	Name      string
	Slug      string
}
