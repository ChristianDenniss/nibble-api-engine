package sourcemenu

import (
	sourceentity "github.com/ChristianDenniss/go-data-model/source/entity"
)

type menuBrowseResponse struct {
	Store      storeWire              `json:"store"`
	Menu       menuWire               `json:"menu"`
	Categories []categoryWithItemsWire `json:"categories"`
}

type storeWire struct {
	ID              string       `json:"id"`
	ChannelID       string       `json:"channelId"`
	ExternalStoreID string       `json:"externalStoreId"`
	Name            string       `json:"name"`
	Location        locationWire `json:"location"`
	Phone           string       `json:"phone"`
}

type menuWire struct {
	ID               string `json:"id"`
	SourceStoreID    string `json:"sourceStoreId"`
	FulfillmentMode  string `json:"fulfillmentMode"`
	DeliveryExecutor string `json:"deliveryExecutor"`
	ExternalMenuID   string `json:"externalMenuId"`
}

type categoryWire struct {
	ID                 string `json:"id"`
	SourceMenuID       string `json:"sourceMenuId"`
	ExternalCategoryID string `json:"externalCategoryId"`
	Name               string `json:"name"`
	SortOrder          int    `json:"sortOrder"`
}

type menuItemWire struct {
	ID               string    `json:"id"`
	SourceCategoryID string    `json:"sourceCategoryId"`
	ExternalItemID   string    `json:"externalItemId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Available        bool      `json:"available"`
	Price            moneyWire `json:"price"`
}

type categoryWithItemsWire struct {
	Category categoryWire   `json:"category"`
	Items    []menuItemWire `json:"items"`
}

type locationWire struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type moneyWire struct {
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
}

func toMenuBrowseResponse(b sourceentity.MenuBrowse) menuBrowseResponse {
	cats := make([]categoryWithItemsWire, len(b.Categories))
	for i, c := range b.Categories {
		items := make([]menuItemWire, len(c.Items))
		for j, it := range c.Items {
			items[j] = menuItemWire{
				ID: it.Item.ID, SourceCategoryID: it.Item.SourceCategoryID, ExternalItemID: it.Item.ExternalItemID,
				Name: it.Item.Name, Description: it.Item.Description, Available: it.Item.Available,
				Price: moneyWire{AmountCents: it.PriceCents, Currency: it.Currency},
			}
		}
		cats[i] = categoryWithItemsWire{
			Category: categoryWire{
				ID: c.Category.ID, SourceMenuID: c.Category.SourceMenuID,
				ExternalCategoryID: c.Category.ExternalCategoryID, Name: c.Category.Name, SortOrder: c.Category.SortOrder,
			},
			Items: items,
		}
	}
	return menuBrowseResponse{
		Store: storeWire{
			ID: b.Store.ID, ChannelID: b.Store.ChannelID, ExternalStoreID: b.Store.ExternalStoreID,
			Name: b.Store.Name, Phone: b.Store.Phone,
			Location: locationWire{
				Latitude: b.Store.Location.Latitude, Longitude: b.Store.Location.Longitude,
				Address: b.Store.Location.Address,
			},
		},
		Menu: menuWire{
			ID: b.Menu.ID, SourceStoreID: b.Menu.SourceStoreID, FulfillmentMode: b.Menu.FulfillmentMode,
			DeliveryExecutor: b.Menu.DeliveryExecutor, ExternalMenuID: b.Menu.ExternalMenuID,
		},
		Categories: cats,
	}
}
