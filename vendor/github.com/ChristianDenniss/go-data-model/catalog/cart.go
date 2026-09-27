package catalog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
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
	LowestListedSubtotal bool             `json:"lowestListedSubtotal"`
	HandoffMode          string           `json:"handoffMode"`
	Provider             string           `json:"provider"`
	SourceURL            string           `json:"sourceUrl"`
	Currency             string           `json:"currency"`
	Lines                []PricedCartLine `json:"lines"`
	Complete             bool             `json:"complete"`
	StartingPrice        bool             `json:"startingPrice"`
	Subtotal             *int64           `json:"subtotalCents"`
	KnownSubtotal        int64            `json:"knownSubtotalCents"`
	Delivery             *int64           `json:"deliveryCents"`
	Service              *int64           `json:"serviceCents"`
	Tax                  *int64           `json:"taxCents"`
	Total                *int64           `json:"totalCents"`
	Promotions           []Promotion      `json:"promotions"`
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
		p := ProviderCart{HandoffMode: "menu_link", Provider: source.Provider, SourceURL: source.URL, Currency: "CAD", Complete: true, Lines: []PricedCartLine{}, Promotions: []Promotion{}}
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
	RankCarts(result.Providers)
	return result, nil
}

// Rank only complete fixed-price item baskets. This never ranks delivered cost.
func RankCarts(providers []ProviderCart) {
	eligible := func(p ProviderCart) bool { return p.Complete && !p.StartingPrice && p.Subtotal != nil }
	count := 0
	var lowest int64
	for i := range providers {
		p := &providers[i]
		p.LowestListedSubtotal = false
		if eligible(*p) {
			if count == 0 || *p.Subtotal < lowest {
				lowest = *p.Subtotal
			}
			count++
		}
	}
	if count > 1 {
		for i := range providers {
			p := &providers[i]
			p.LowestListedSubtotal = eligible(*p) && *p.Subtotal == lowest
		}
	}
	sort.SliceStable(providers, func(i, j int) bool {
		a, b := providers[i], providers[j]
		if eligible(a) != eligible(b) {
			return eligible(a)
		}
		if eligible(a) && *a.Subtotal != *b.Subtotal {
			return *a.Subtotal < *b.Subtotal
		}
		if a.Complete != b.Complete {
			return a.Complete
		}
		return a.Provider < b.Provider
	})
}

type HandoffRequest struct {
	RestaurantID string     `json:"restaurantId"`
	Lines        []CartLine `json:"lines"`
	Provider     string     `json:"provider"`
}
type CartHandoff struct {
	Provider        string `json:"provider"`
	Mode            string `json:"mode"`
	URL             string `json:"url"`
	CartTransferred bool   `json:"cartTransferred"`
	CartText        string `json:"cartText"`
}

// Public menu links cannot create authenticated consumer carts. This explicit
// fallback preserves the basket and never claims an external cart was created.
func (s *Service) PrepareHandoff(ctx context.Context, req HandoffRequest) (CartHandoff, error) {
	c, err := s.CompareCart(ctx, CartRequest{RestaurantID: req.RestaurantID, Lines: req.Lines})
	if err != nil {
		return CartHandoff{}, err
	}
	for _, p := range c.Providers {
		if p.Provider == req.Provider {
			lines := []string{c.RestaurantName, c.Address}
			for _, line := range p.Lines {
				lines = append(lines, fmt.Sprintf("%d × %s", line.Quantity, line.Name))
			}
			return CartHandoff{Provider: p.Provider, Mode: "menu_link", URL: p.SourceURL, CartTransferred: false, CartText: strings.Join(lines, "\n")}, nil
		}
	}
	return CartHandoff{}, fmt.Errorf("%w: provider is not linked to this restaurant", ErrInvalidCart)
}
