package entity

import (
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type Order struct {
	ID           string
	AccountID    string
	RestaurantID string
	ProviderID   string
	PlacedAt     time.Time
	Status       Status
	Total        money.Money
	Lines        []OrderLine
}

type OrderLine struct {
	ID         string
	MenuItemID string
	Name       string
	Quantity   int
}
