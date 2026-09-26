package entity

type Cart struct {
	ID        string
	AccountID string
	Lines     []CartLine
}

type CartLine struct {
	ID           string
	RestaurantID string
	MenuItemID   string
	ProviderID   string
	Quantity     int
}
