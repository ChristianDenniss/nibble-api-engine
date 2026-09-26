package entity

import (
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
)

type Observation struct {
	ID                   string
	SourceStoreID        string
	ChannelID            string
	FulfillmentMode      string
	DeliveryExecutor     string
	DropoffGeohash       string
	MembershipTier       string
	QuoteKind            string
	BasketSubtotalCents  int64
	ObservedAt           time.Time
	IngestRunID          string
	FeeLines             []FeeLine
}

type FeeLine struct {
	ID             string
	QuoteObsID     string
	Kind           string
	Amount         money.Money
	Percent        float64
	ThresholdCents int64
}
