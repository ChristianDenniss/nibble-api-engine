package ingest

import (
	"context"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/store"
	model "github.com/ChristianDenniss/go-data-model"
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) RecordRestaurant(ctx context.Context, restaurant model.Restaurant) error {
	if restaurant.ID == "" {
		return ErrRestaurantIDRequired
	}
	return s.store.UpsertRestaurant(ctx, restaurant)
}

func (s *Service) RecordMenuItem(ctx context.Context, item model.MenuItem) error {
	if item.ID == "" {
		return ErrMenuItemIDRequired
	}
	return s.store.UpsertMenuItem(ctx, item)
}

func (s *Service) RecordOffer(ctx context.Context, offer model.Offer) error {
	if offer.ID == "" {
		return ErrOfferIDRequired
	}
	return s.store.UpsertOffer(ctx, offer)
}

func (s *Service) RecordPriceObservation(ctx context.Context, obs model.PriceObservation) error {
	if obs.ID == "" {
		return ErrPriceObservationIDRequired
	}
	if obs.ObservedAt.IsZero() {
		obs.ObservedAt = time.Now().UTC()
	}
	return s.store.UpsertPriceObservation(ctx, obs)
}
