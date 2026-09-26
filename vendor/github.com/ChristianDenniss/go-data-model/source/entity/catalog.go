package entity

import location "github.com/ChristianDenniss/go-data-model/location/entity"

type Store struct {
	ID              string
	ChannelID       string
	ExternalStoreID string
	Name            string
	Location        location.Location
	Phone           string
}

type Menu struct {
	ID               string
	SourceStoreID    string
	FulfillmentMode  string
	DeliveryExecutor string
	ExternalMenuID   string
}

type Category struct {
	ID                 string
	SourceMenuID       string
	ExternalCategoryID string
	Name               string
	SortOrder          int
}

type Item struct {
	ID             string
	SourceCategoryID string
	ExternalItemID string
	Name           string
	Description    string
	Available      bool
}

type ModifierGroup struct {
	ID         string
	ExternalID string
	MinSelect  int
	MaxSelect  int
}

type ModifierOption struct {
	ID               string
	SourceModGroupID string
	ExternalOptionID string
	Name             string
	PriceCents       int64
	Currency         string
}

type ItemModifierGroup struct {
	SourceItemID     string
	SourceModGroupID string
}
