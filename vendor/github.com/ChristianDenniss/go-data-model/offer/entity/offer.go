package entity

import money "github.com/ChristianDenniss/go-data-model/money/entity"

type Offer struct {
	ID               string
	RestaurantID     string
	ProviderID       string
	MenuItemID       string
	Price            money.Money
	EstimatedMinutes int
}
