package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	compareentity "github.com/ChristianDenniss/go-data-model/compare/entity"
	channelsvc "github.com/ChristianDenniss/go-data-model/channel/service"
	channelentity "github.com/ChristianDenniss/go-data-model/channel/entity"
	fulfillmententity "github.com/ChristianDenniss/go-data-model/fulfillment/entity"
	itempricesvc "github.com/ChristianDenniss/go-data-model/itemprice/service"
	placeentity "github.com/ChristianDenniss/go-data-model/place/entity"
	placesvc "github.com/ChristianDenniss/go-data-model/place/service"
	quoteobssvc "github.com/ChristianDenniss/go-data-model/quoteobs/service"
	resolutionsvc "github.com/ChristianDenniss/go-data-model/resolution/service"
	serviceabilitysvc "github.com/ChristianDenniss/go-data-model/serviceability/service"
	userentity "github.com/ChristianDenniss/go-data-model/user/entity"
	usersvc "github.com/ChristianDenniss/go-data-model/user/service"
)

type Service struct {
	places     *placesvc.Service
	users      *usersvc.Service
	channels   *channelsvc.Service
	resolution *resolutionsvc.Service
	itemPrices *itempricesvc.Service
	quotes           *quoteobssvc.Service
	serviceability   *serviceabilitysvc.Service
}

func New(
	places *placesvc.Service,
	users *usersvc.Service,
	channels *channelsvc.Service,
	resolution *resolutionsvc.Service,
	itemPrices *itempricesvc.Service,
	quotes *quoteobssvc.Service,
	serviceability *serviceabilitysvc.Service,
) *Service {
	return &Service{
		places:           places,
		users:            users,
		channels:         channels,
		resolution:       resolution,
		itemPrices:       itemPrices,
		quotes:           quotes,
		serviceability:   serviceability,
	}
}

func (s *Service) Compare(ctx context.Context, userID string, req compareentity.Request) (compareentity.Result, userentity.Session, error) {
	if req.PlaceID == "" {
		return compareentity.Result{}, userentity.Session{}, compareentity.ErrPlaceRequired
	}
	if len(req.Basket.Lines) == 0 {
		return compareentity.Result{}, userentity.Session{}, compareentity.ErrBasketRequired
	}

	req, err := s.applyUserDefaults(ctx, userID, req)
	if err != nil {
		return compareentity.Result{}, userentity.Session{}, err
	}

	place, err := s.places.GetPlace(ctx, req.PlaceID)
	if err != nil {
		return compareentity.Result{}, userentity.Session{}, err
	}

	options, err := s.places.ListPurchaseOptions(ctx, req.PlaceID)
	if err != nil {
		return compareentity.Result{}, userentity.Session{}, err
	}

	channelKinds, err := s.channelKindsForOptions(ctx, options)
	if err != nil {
		return compareentity.Result{}, userentity.Session{}, err
	}
	filtered := filterOptions(options, req.Filters, channelKinds)
	outcome := s.rankPaths(ctx, filtered, req, channelKinds)

	queryJSON, _ := json.Marshal(req)
	basketJSON, _ := json.Marshal(req.Basket)
	sessionID := newSessionID()
	resultPayload := buildResultPayload(sessionID, place, outcome, req)
	resultJSON, _ := json.Marshal(resultPayload)

	session := userentity.Session{
		ID:             sessionID,
		UserID:         userID,
		PlaceID:        req.PlaceID,
		QuerySnapshot:  queryJSON,
		BasketSnapshot: basketJSON,
		ResultSnapshot: resultJSON,
	}
	if err := s.users.SaveSession(ctx, session); err != nil {
		return compareentity.Result{}, userentity.Session{}, err
	}

	return compareentity.Result{
		CompareSessionID: session.ID,
		Recommendation:   firstRank(outcome.Ranked),
		RunnersUp:        restRanks(outcome.Ranked),
		UnavailablePaths: outcome.Unavailable,
		Raw:              resultJSON,
	}, session, nil
}

func (s *Service) channelKindsForOptions(ctx context.Context, opts []placeentity.PurchaseOption) (map[string]string, error) {
	kinds := make(map[string]string)
	for _, o := range opts {
		if _, seen := kinds[o.ChannelID]; seen {
			continue
		}
		ch, err := s.channels.GetByID(ctx, o.ChannelID)
		if err != nil {
			if errors.Is(err, channelentity.ErrNotFound) {
				kinds[o.ChannelID] = ""
				continue
			}
			return nil, err
		}
		kinds[o.ChannelID] = ch.Kind
	}
	return kinds, nil
}

func filterOptions(opts []placeentity.PurchaseOption, prefs userentity.ComparePrefs, channelKinds map[string]string) []placeentity.PurchaseOption {
	var out []placeentity.PurchaseOption
	for _, o := range opts {
		if !prefs.WillingToUseAggregator && channelKinds[o.ChannelID] == fulfillmententity.ChannelKindAggregator {
			continue
		}
		if len(prefs.AllowedFulfillmentModes) > 0 {
			allowed := false
			for _, mode := range prefs.AllowedFulfillmentModes {
				if fulfillmententity.OptionMatchesFilter(o.FulfillmentMode, o.DeliveryExecutor, mode) {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		if len(prefs.BlockedChannelIDs) > 0 && contains(prefs.BlockedChannelIDs, o.ChannelID) {
			continue
		}
		if len(prefs.AllowedChannelIDs) > 0 && !contains(prefs.AllowedChannelIDs, o.ChannelID) {
			continue
		}
		out = append(out, o)
	}
	return out
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func buildResultPayload(sessionID string, place placeentity.Place, outcome rankOutcome, req compareentity.Request) map[string]interface{} {
	return map[string]interface{}{
		"compare_session_id": sessionID,
		"observed_at":        time.Now().UTC().Format(time.RFC3339),
		"place": map[string]string{
			"id":   place.ID,
			"name": place.Name,
		},
		"paths_ranked":      len(outcome.Ranked),
		"recommendation":    firstRank(outcome.Ranked),
		"runners_up":        restRanks(outcome.Ranked),
		"unavailable_paths": outcome.Unavailable,
		"query":             req,
	}
}

func firstRank(ranked []compareentity.PathRank) *compareentity.PathRank {
	if len(ranked) == 0 {
		return nil
	}
	return &ranked[0]
}

func restRanks(ranked []compareentity.PathRank) []compareentity.PathRank {
	if len(ranked) <= 1 {
		return nil
	}
	return ranked[1:]
}

func newSessionID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return "cmp_" + hex.EncodeToString(b[:])
}

func (s *Service) applyUserDefaults(ctx context.Context, userID string, req compareentity.Request) (compareentity.Request, error) {
	if userID == "" {
		return req, nil
	}
	if comparePrefsEmpty(req.Filters) {
		settings, err := s.users.GetSettings(ctx, userID)
		if err == nil {
			req.Filters = settings.ComparePrefs
		}
	}
	if len(req.Memberships) == 0 {
		slugs, err := s.users.ListMembershipSlugs(ctx, userID)
		if err != nil {
			return req, err
		}
		req.Memberships = slugs
	}
	return req, nil
}

func comparePrefsEmpty(p userentity.ComparePrefs) bool {
	return len(p.AllowedFulfillmentModes) == 0 &&
		len(p.AllowedChannelIDs) == 0 &&
		len(p.BlockedChannelIDs) == 0 &&
		!p.WillingToUseAggregator &&
		!p.DriveThruOK &&
		!p.MerchantDirectOK
}
