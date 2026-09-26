package entity

import (
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
)

type Observation struct {
	ID              string
	SourceItemID    string
	IngestRunID     string
	Price           money.Money
	FulfillmentMode  string
	DeliveryExecutor string
	ObservedAt       time.Time
}
