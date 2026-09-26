package entity

import (
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
)

type Observation struct {
	ID         string
	OfferID    string
	Price      money.Money
	ObservedAt time.Time
}
