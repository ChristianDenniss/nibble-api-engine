package storefront

import (
	"time"

	accountentity "github.com/ChristianDenniss/go-data-model/account/entity"
	categoryentity "github.com/ChristianDenniss/go-data-model/category/entity"
	cartentity "github.com/ChristianDenniss/go-data-model/cart/entity"
	cuisineentity "github.com/ChristianDenniss/go-data-model/cuisine/entity"
	menuentity "github.com/ChristianDenniss/go-data-model/menu/entity"
	offerentity "github.com/ChristianDenniss/go-data-model/offer/entity"
	orderentity "github.com/ChristianDenniss/go-data-model/order/entity"
	providerentity "github.com/ChristianDenniss/go-data-model/provider/entity"
	restaurantentity "github.com/ChristianDenniss/go-data-model/restaurant/entity"
	storefrontentity "github.com/ChristianDenniss/go-data-model/storefront/entity"
	locationentity "github.com/ChristianDenniss/go-data-model/location/entity"
)

// Wire types use camelCase JSON keys to match gentypes / the web app.

type catalogResponse struct {
	Account     accountWire      `json:"account"`
	Providers   []providerWire   `json:"providers"`
	Categories  []categoryWire   `json:"categories"`
	Cuisines    []cuisineWire    `json:"cuisines"`
	Restaurants []restaurantWire `json:"restaurants"`
	Items       []itemWire       `json:"items"`
	Offers      []offerWire      `json:"offers"`
	Cart        cartWire         `json:"cart"`
	Orders      []orderWire      `json:"orders"`
}

type moneyWire struct {
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
}

type locationWire struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	Region     string  `json:"region"`
	PostalCode string  `json:"postalCode"`
}

type savedAddressWire struct {
	ID       string       `json:"id"`
	Label    string       `json:"label"`
	Location locationWire `json:"location"`
	Current  bool         `json:"current"`
}

type paymentMethodWire struct {
	ID        string `json:"id"`
	Brand     string `json:"brand"`
	Last4     string `json:"last4"`
	ExpMonth  int    `json:"expMonth"`
	ExpYear   int    `json:"expYear"`
	Default   bool   `json:"default"`
}

type accountWire struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Email          string              `json:"email"`
	Phone          string              `json:"phone"`
	Addresses      []savedAddressWire  `json:"addresses"`
	PaymentMethods []paymentMethodWire `json:"paymentMethods"`
}

type providerWire struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type categoryWire struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type cuisineWire struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ratingWire struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type restaurantWire struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Location    locationWire `json:"location"`
	CuisineIds  []string     `json:"cuisineIds"`
	CategoryIds []string     `json:"categoryIds"`
	Rating      ratingWire   `json:"rating"`
	Phone       string       `json:"phone"`
	AppURL      string       `json:"appURL"`
	Hours       []hoursWire  `json:"hours"`
}

type hoursWire struct {
	Service   string `json:"service"`
	DayOfWeek int    `json:"dayOfWeek"`
	Opens     string `json:"opens"`
	Closes    string `json:"closes"`
}

type itemWire struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurantId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Section      string `json:"section"`
	ImageURL     string `json:"imageURL"`
}

type offerWire struct {
	ID               string    `json:"id"`
	RestaurantID     string    `json:"restaurantId"`
	MenuItemID       string    `json:"menuItemId"`
	ProviderID       string    `json:"providerId"`
	Price            moneyWire `json:"price"`
	EstimatedMinutes int       `json:"estimatedMinutes"`
}

type cartLineWire struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurantId"`
	MenuItemID   string `json:"menuItemId"`
	ProviderID   string `json:"providerId"`
	Quantity     int    `json:"quantity"`
}

type cartWire struct {
	ID        string         `json:"id"`
	AccountID string         `json:"accountId"`
	Lines     []cartLineWire `json:"lines"`
}

type orderLineWire struct {
	ID         string `json:"id"`
	MenuItemID string `json:"menuItemId"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
}

type orderWire struct {
	ID           string          `json:"id"`
	AccountID    string          `json:"accountId"`
	RestaurantID string          `json:"restaurantId"`
	ProviderID   string          `json:"providerId"`
	PlacedAt     string          `json:"placedAt"`
	Status       string          `json:"status"`
	Total        moneyWire       `json:"total"`
	Lines        []orderLineWire `json:"lines"`
}

func toCatalogResponse(c storefrontentity.Catalog) catalogResponse {
	return catalogResponse{
		Account:     toAccountWire(c.Account),
		Providers:   mapProviders(c.Providers),
		Categories:  mapCategories(c.Categories),
		Cuisines:    mapCuisines(c.Cuisines),
		Restaurants: mapRestaurants(c.Restaurants),
		Items:       mapItems(c.Items),
		Offers:      mapOffers(c.Offers),
		Cart:        toCartWire(c.Cart),
		Orders:      mapOrders(c.Orders),
	}
}

func toAccountWire(a accountentity.Account) accountWire {
	addrs := make([]savedAddressWire, 0, len(a.Addresses))
	for _, addr := range a.Addresses {
		addrs = append(addrs, savedAddressWire{
			ID: addr.ID, Label: addr.Label, Current: addr.Current,
			Location: locationWire{
				Latitude: addr.Location.Latitude, Longitude: addr.Location.Longitude,
				Address: addr.Location.Address, City: addr.Location.City,
				Region: addr.Location.Region, PostalCode: addr.Location.PostalCode,
			},
		})
	}
	pays := make([]paymentMethodWire, 0, len(a.PaymentMethods))
	for _, pm := range a.PaymentMethods {
		pays = append(pays, paymentMethodWire{
			ID: pm.ID, Brand: pm.Brand, Last4: pm.Last4,
			ExpMonth: pm.ExpMonth, ExpYear: pm.ExpYear, Default: pm.Default,
		})
	}
	return accountWire{
		ID: a.ID, Name: a.Name, Email: a.Email, Phone: a.Phone,
		Addresses: addrs, PaymentMethods: pays,
	}
}

func mapProviders(in []providerentity.Provider) []providerWire {
	out := make([]providerWire, len(in))
	for i, p := range in {
		out[i] = providerWire{ID: p.ID, Name: p.Name}
	}
	return out
}

func mapCategories(in []categoryentity.Category) []categoryWire {
	out := make([]categoryWire, len(in))
	for i, c := range in {
		out[i] = categoryWire{ID: c.ID, Slug: c.Slug, Name: c.Name, Description: c.Description}
	}
	return out
}

func mapCuisines(in []cuisineentity.Cuisine) []cuisineWire {
	out := make([]cuisineWire, len(in))
	for i, c := range in {
		out[i] = cuisineWire{ID: c.ID, Slug: c.Slug, Name: c.Name}
	}
	return out
}

func mapRestaurants(in []restaurantentity.Restaurant) []restaurantWire {
	out := make([]restaurantWire, len(in))
	for i, r := range in {
		out[i] = restaurantWire{
			ID: r.ID, Name: r.Name,
			Location: locationFromDomain(r.Location),
			CuisineIds: r.CuisineIDs, CategoryIds: r.CategoryIDs,
			Rating: ratingWire{Average: r.Rating.Average, Count: r.Rating.Count},
			Phone: r.Phone, AppURL: r.AppURL,
			Hours: mapHours(r.Hours),
		}
	}
	return out
}

func mapHours(in []restaurantentity.Hours) []hoursWire {
	out := make([]hoursWire, len(in))
	for i, h := range in {
		out[i] = hoursWire{Service: h.Service, DayOfWeek: h.DayOfWeek, Opens: h.Opens, Closes: h.Closes}
	}
	return out
}

func mapItems(in []menuentity.Item) []itemWire {
	out := make([]itemWire, len(in))
	for i, item := range in {
		out[i] = toItemWire(item)
	}
	return out
}

func toItemWire(item menuentity.Item) itemWire {
	return itemWire{
		ID: item.ID, RestaurantID: item.RestaurantID,
		Name: item.Name, Description: item.Description, Section: item.Section,
		ImageURL: item.ImageURL,
	}
}

func mapOffers(in []offerentity.Offer) []offerWire {
	out := make([]offerWire, len(in))
	for i, o := range in {
		out[i] = offerWire{
			ID: o.ID, RestaurantID: o.RestaurantID, MenuItemID: o.MenuItemID, ProviderID: o.ProviderID,
			Price: moneyWire{AmountCents: o.Price.AmountCents, Currency: o.Price.Currency},
			EstimatedMinutes: o.EstimatedMinutes,
		}
	}
	return out
}

func toCartWire(c cartentity.Cart) cartWire {
	lines := make([]cartLineWire, len(c.Lines))
	for i, line := range c.Lines {
		lines[i] = cartLineWire{
			ID: line.ID, RestaurantID: line.RestaurantID,
			MenuItemID: line.MenuItemID, ProviderID: line.ProviderID, Quantity: line.Quantity,
		}
	}
	return cartWire{ID: c.ID, AccountID: c.AccountID, Lines: lines}
}

func mapOrders(in []orderentity.Order) []orderWire {
	out := make([]orderWire, len(in))
	for i, o := range in {
		lines := make([]orderLineWire, len(o.Lines))
		for j, line := range o.Lines {
			lines[j] = orderLineWire{
				ID: line.ID, MenuItemID: line.MenuItemID, Name: line.Name, Quantity: line.Quantity,
			}
		}
		out[i] = orderWire{
			ID: o.ID, AccountID: o.AccountID, RestaurantID: o.RestaurantID, ProviderID: o.ProviderID,
			PlacedAt: o.PlacedAt.Format(time.RFC3339), Status: string(o.Status),
			Total: moneyWire{AmountCents: o.Total.AmountCents, Currency: o.Total.Currency},
			Lines: lines,
		}
	}
	return out
}

func locationFromDomain(loc locationentity.Location) locationWire {
	return locationWire{
		Latitude: loc.Latitude, Longitude: loc.Longitude,
		Address: loc.Address, City: loc.City, Region: loc.Region, PostalCode: loc.PostalCode,
	}
}
