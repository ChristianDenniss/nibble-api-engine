package entity

import location "github.com/ChristianDenniss/go-data-model/location/entity"

type Place struct {
	ID       string
	BrandID  string
	Name     string
	Location location.Location
}

type PurchaseOption struct {
	ID                string
	PlaceID           string
	ChannelID         string
	FulfillmentMode   string
	DeliveryExecutor  string
	SourceStoreID     string
}
