package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	compareentity "github.com/ChristianDenniss/go-data-model/compare/entity"
	fulfillmententity "github.com/ChristianDenniss/go-data-model/fulfillment/entity"
	itempriceentity "github.com/ChristianDenniss/go-data-model/itemprice/entity"
	money "github.com/ChristianDenniss/go-data-model/money/entity"
	placeentity "github.com/ChristianDenniss/go-data-model/place/entity"
	quoteentity "github.com/ChristianDenniss/go-data-model/quoteobs/entity"
	resolutionentity "github.com/ChristianDenniss/go-data-model/resolution/entity"
)

type pathSpec struct {
	Option      placeentity.PurchaseOption
	Fulfillment string
	Executor    string
}

type rankOutcome struct {
	Ranked      []compareentity.PathRank
	Unavailable []compareentity.UnavailablePath
}

func (s *Service) rankPaths(ctx context.Context, opts []placeentity.PurchaseOption, req compareentity.Request, channelKinds map[string]string) rankOutcome {
	var ranked []compareentity.PathRank
	var unavailable []compareentity.UnavailablePath

	for _, o := range opts {
		spec := resolvePath(o, channelKinds)
		geohash := quoteGeohash(req)
		if s.serviceability != nil {
			ok, code, err := s.serviceability.Deliverable(ctx, spec.Option.SourceStoreID, spec.Fulfillment, spec.Executor, geohash)
			if err != nil {
				unavailable = append(unavailable, compareentity.UnavailablePath{
					PurchaseOptionID: spec.Option.ID,
					Code:             "serviceability_error",
					Message:          err.Error(),
				})
				continue
			}
			if !ok {
				unavailable = append(unavailable, compareentity.UnavailablePath{
					PurchaseOptionID: spec.Option.ID,
					Code:             code,
					Message:          "path not serviceable for this query",
				})
				continue
			}
		}
		pr, unavail, ok := s.pricePath(ctx, spec, req)
		if !ok {
			unavailable = append(unavailable, unavail)
			continue
		}
		ranked = append(ranked, pr)
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].AllIn.AmountCents < ranked[j].AllIn.AmountCents
	})
	for i := range ranked {
		ranked[i].Rank = i + 1
	}
	return rankOutcome{Ranked: ranked, Unavailable: unavailable}
}

func resolvePath(opt placeentity.PurchaseOption, channelKinds map[string]string) pathSpec {
	kind := channelKinds[opt.ChannelID]
	f, e := fulfillmententity.CanonicalizeForChannel(opt.FulfillmentMode, opt.DeliveryExecutor, kind)
	return pathSpec{Option: opt, Fulfillment: f, Executor: e}
}

func (s *Service) pricePath(ctx context.Context, spec pathSpec, req compareentity.Request) (compareentity.PathRank, compareentity.UnavailablePath, bool) {
	opt := spec.Option
	subtotal, currency, itemBullets, err := s.basketSubtotal(ctx, spec, req)
	if err != nil {
		return compareentity.PathRank{}, compareentity.UnavailablePath{
			PurchaseOptionID: opt.ID,
			Code:             "basket_unpriced",
			Message:          err.Error(),
		}, false
	}

	fees := int64(0)
	quoteBullets := []string{}
	confidence := "medium"
	dropoffGeohash := quoteGeohash(req)

	if fulfillmententity.NeedsDropoffForQuote(spec.Fulfillment) && req.FulfillmentContext.Dropoff == nil {
		quoteBullets = append(quoteBullets, "delivery quote skipped: dropoff required")
		confidence = "low"
	} else {
		tier := membershipTier(req.Memberships)
		quote, qErr := s.quotes.Latest(ctx, opt.SourceStoreID, dropoffGeohash, spec.Fulfillment, spec.Executor, tier, subtotal)
		if qErr != nil {
			if !errors.Is(qErr, quoteentity.ErrNotFound) {
				quoteBullets = append(quoteBullets, fmt.Sprintf("quote lookup failed: %v", qErr))
			} else {
				quoteBullets = append(quoteBullets, "no recent quote observation for this path")
			}
			confidence = "low"
		} else {
			fees = sumQuoteFees(quote)
			quoteBullets = append(quoteBullets, fmt.Sprintf("quote %s (%d fee lines)", quote.ID, len(quote.FeeLines)))
			if confidence != "low" {
				confidence = "high"
			}
		}
	}

	allIn := subtotal + fees
	bullets := append(itemBullets, quoteBullets...)
	bullets = append(bullets, fmt.Sprintf("subtotal %d + fees %d %s", subtotal, fees, currency))

	return compareentity.PathRank{
		PurchaseOptionID: opt.ID,
		Kind:             "lowest_all_in",
		Headline:         headlineForPath(spec.Fulfillment, spec.Executor),
		Confidence:       confidence,
		AllIn:            money.Money{AmountCents: allIn, Currency: currency},
		FulfillmentMode:  spec.Fulfillment,
		DeliveryExecutor: spec.Executor,
		ChannelID:        opt.ChannelID,
		RationaleBullets: bullets,
	}, compareentity.UnavailablePath{}, true
}

func (s *Service) basketSubtotal(ctx context.Context, spec pathSpec, req compareentity.Request) (int64, string, []string, error) {
	opt := spec.Option
	var subtotal int64
	currency := "CAD"
	var bullets []string

	for _, line := range req.Basket.Lines {
		qty := line.Quantity
		if qty < 1 {
			qty = 1
		}
		sourceItemID := strings.TrimSpace(line.SourceItemID)
		if sourceItemID == "" {
			if line.DishID == "" {
				return 0, "", nil, errors.New("basket line needs dish_id or source_item_id")
			}
			resolved, err := s.resolution.ResolveSourceItemForStoreAndDish(ctx, opt.SourceStoreID, line.DishID)
			if err != nil {
				if errors.Is(err, resolutionentity.ErrNotFound) {
					return 0, "", nil, fmt.Errorf("no item match for dish %s on store %s", line.DishID, opt.SourceStoreID)
				}
				return 0, "", nil, err
			}
			sourceItemID = resolved
		}
		obs, err := s.itemPrices.Latest(ctx, sourceItemID, spec.Fulfillment, spec.Executor)
		if err != nil {
			if errors.Is(err, itempriceentity.ErrNotFound) {
				return 0, "", nil, fmt.Errorf("no price for item %s (%s/%s)", sourceItemID, spec.Fulfillment, spec.Executor)
			}
			return 0, "", nil, err
		}
		if currency == "CAD" && obs.Price.Currency != "" {
			currency = obs.Price.Currency
		}
		lineTotal := obs.Price.AmountCents * int64(qty)
		subtotal += lineTotal
		bullets = append(bullets, fmt.Sprintf("%s x%d @ %d", sourceItemID, qty, obs.Price.AmountCents))
	}
	return subtotal, currency, bullets, nil
}

func sumQuoteFees(quote quoteentity.Observation) int64 {
	var total int64
	for _, line := range quote.FeeLines {
		total += line.Amount.AmountCents
	}
	return total
}

func quoteGeohash(req compareentity.Request) string {
	if req.FulfillmentContext.Dropoff != nil {
		return EncodeGeohash(req.FulfillmentContext.Dropoff.Latitude, req.FulfillmentContext.Dropoff.Longitude, 5)
	}
	return ""
}

func membershipTier(memberships []string) string {
	for _, m := range memberships {
		if strings.TrimSpace(m) != "" {
			return strings.TrimSpace(m)
		}
	}
	return ""
}

func headlineForPath(fulfillment, executor string) string {
	switch fulfillment {
	case fulfillmententity.ModeDelivery:
		switch executor {
		case fulfillmententity.ExecutorThirdParty:
			return "Delivery (3rd party)"
		case fulfillmententity.ExecutorMerchant:
			return "Delivery (merchant)"
		default:
			return "Delivery"
		}
	case fulfillmententity.ModePickup:
		return "Pickup"
	case fulfillmententity.ModeDriveThru:
		return "Drive-thru"
	case fulfillmententity.ModeInStore:
		return "In store"
	case fulfillmententity.ModeDineIn:
		return "Dine in"
	default:
		return fulfillment
	}
}
