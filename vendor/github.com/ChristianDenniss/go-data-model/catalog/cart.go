package catalog

import (
	"context"
	"errors"
	"fmt"
	"math"
)

var ErrInvalidCart = errors.New("invalid cart")

type CartLine struct {
	ItemID   string `json:"itemId"`
	Quantity int64  `json:"quantity"`
}
type CartRequest struct {
	RestaurantID string     `json:"restaurantId"`
	Lines        []CartLine `json:"lines"`
}
type CartComparison struct {
	RestaurantID   string         `json:"restaurantId"`
	RestaurantName string         `json:"restaurantName"`
	Address        string         `json:"address"`
	Providers      []ProviderCart `json:"providers"`
}
type ProviderCart struct {
	Provider      string           `json:"provider"`
	SourceURL     string           `json:"sourceUrl"`
	Currency      string           `json:"currency"`
	Lines         []PricedCartLine `json:"lines"`
	Complete      bool             `json:"complete"`
	StartingPrice bool             `json:"startingPrice"`
	Subtotal      *int64           `json:"subtotalCents"`
	KnownSubtotal int64            `json:"knownSubtotalCents"`
	Delivery      *int64           `json:"deliveryCents"`
	Service       *int64           `json:"serviceCents"`
	Tax           *int64           `json:"taxCents"`
	Total         *int64           `json:"totalCents"`
	Promotions    []Promotion      `json:"promotions"`
}
type PricedCartLine struct {
	ItemID    string `json:"itemId"`
	Name      string `json:"name"`
	Quantity  int64  `json:"quantity"`
	UnitPrice *int64 `json:"unitPriceCents"`
	LineTotal *int64 `json:"lineTotalCents"`
	PriceKind string `json:"priceKind"`
	Status    string `json:"status"`
}

func (r CartRequest) Validate() error {
	if r.RestaurantID == "" || len(r.Lines) == 0 || len(r.Lines) > 100 {
		return fmt.Errorf("%w: choose 1–100 items from one restaurant", ErrInvalidCart)
	}
	seen := map[string]bool{}
	for _, l := range r.Lines {
		if l.ItemID == "" || l.Quantity < 1 || l.Quantity > 50 || seen[l.ItemID] {
			return fmt.Errorf("%w: quantities must be 1–50 with no duplicate items", ErrInvalidCart)
		}
		seen[l.ItemID] = true
	}
	return nil
}
func (s *Service) CompareCart(ctx context.Context, req CartRequest) (CartComparison, error) {
	if err := req.Validate(); err != nil {
		return CartComparison{}, err
	}
	c, err := s.Read(ctx)
	if err != nil {
		return CartComparison{}, err
	}
	return CompareCart(c, req)
}

// CompareCart prices the same complete basket on each linked provider. Item
// offers do not establish basket/address-specific fees, so fees and totals stay
// unknown until a whole-cart quote exists. Never sum per-item delivery charges.
func CompareCart(c PublicCatalog, req CartRequest) (CartComparison, error) {
	result := CartComparison{Providers: []ProviderCart{}}
	if err := req.Validate(); err != nil {
		return result, err
	}
	var restaurant *Restaurant
	for _, r := range c.Restaurants {
		if r.ID == req.RestaurantID {
			restaurant = r
			break
		}
	}
	if restaurant == nil {
		return result, fmt.Errorf("%w: restaurant unavailable", ErrInvalidCart)
	}
	result.RestaurantID = restaurant.ID
	result.RestaurantName = restaurant.Name
	result.Address = restaurant.Address
	items := map[string]*Item{}
	for _, it := range restaurant.Items {
		items[it.ID] = it
	}
	for _, line := range req.Lines {
		if items[line.ItemID] == nil {
			return result, fmt.Errorf("%w: an item is no longer on this menu; edit your cart", ErrInvalidCart)
		}
	}
	for _, source := range restaurant.Sources {
		p := ProviderCart{Provider: source.Provider, SourceURL: source.URL, Currency: "CAD", Complete: true, Lines: []PricedCartLine{}, Promotions: []Promotion{}}
		promos := map[string]bool{}
		for _, line := range req.Lines {
			item := items[line.ItemID]
			priced := PricedCartLine{ItemID: item.ID, Name: item.Name, Quantity: line.Quantity, PriceKind: "options", Status: "unavailable"}
			for _, offer := range item.Offers {
				if offer.Provider != source.Provider {
					continue
				}
				priced.PriceKind = offer.PriceKind
				priced.Status = offer.Status
				if offer.Amount != nil && offer.Status == "priced" && offer.Currency == "CAD" && *offer.Amount >= 0 {
					if *offer.Amount > math.MaxInt64/line.Quantity {
						return result, fmt.Errorf("%w: price out of range", ErrInvalidCart)
					}
					total := *offer.Amount * line.Quantity
					if p.KnownSubtotal > math.MaxInt64-total {
						return result, fmt.Errorf("%w: subtotal out of range", ErrInvalidCart)
					}
					priced.UnitPrice = offer.Amount
					priced.LineTotal = &total
					p.KnownSubtotal += total
					p.StartingPrice = p.StartingPrice || offer.PriceKind == "from"
				}
				for _, promo := range offer.Promotions {
					key := promo.Title + "\x00" + promo.Conditions
					if !promos[key] {
						promos[key] = true
						p.Promotions = append(p.Promotions, promo)
					}
				}
				break
			}
			if priced.LineTotal == nil {
				p.Complete = false
			}
			p.Lines = append(p.Lines, priced)
		}
		if p.Complete {
			subtotal := p.KnownSubtotal
			p.Subtotal = &subtotal
		}
		result.Providers = append(result.Providers, p)
	}
	return result, nil
}
