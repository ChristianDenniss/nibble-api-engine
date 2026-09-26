package ingest

import (
	"context"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/store"
	model "github.com/ChristianDenniss/go-data-model"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ingestv1.UnimplementedIngestServiceServer
	store *store.Store
}

func NewServer(s *store.Store) *Server {
	return &Server{store: s}
}

func (s *Server) RecordRestaurant(ctx context.Context, req *ingestv1.RecordRestaurantRequest) (*ingestv1.RecordRestaurantResponse, error) {
	r := req.GetRestaurant()
	if r == nil || r.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "restaurant.id is required")
	}
	loc := r.GetLocation()
	err := s.store.UpsertRestaurant(ctx, model.Restaurant{
		ID:   r.GetId(),
		Name: r.GetName(),
		Location: model.Location{
			Latitude:  loc.GetLatitude(),
			Longitude: loc.GetLongitude(),
			Address:   loc.GetAddress(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "record restaurant: %v", err)
	}
	return &ingestv1.RecordRestaurantResponse{Id: r.GetId()}, nil
}

func (s *Server) RecordMenuItem(ctx context.Context, req *ingestv1.RecordMenuItemRequest) (*ingestv1.RecordMenuItemResponse, error) {
	item := req.GetMenuItem()
	if item == nil || item.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "menu_item.id is required")
	}
	err := s.store.UpsertMenuItem(ctx, model.MenuItem{
		ID:           item.GetId(),
		RestaurantID: item.GetRestaurantId(),
		Name:         item.GetName(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "record menu item: %v", err)
	}
	return &ingestv1.RecordMenuItemResponse{Id: item.GetId()}, nil
}

func (s *Server) RecordOffer(ctx context.Context, req *ingestv1.RecordOfferRequest) (*ingestv1.RecordOfferResponse, error) {
	offer := req.GetOffer()
	if offer == nil || offer.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "offer.id is required")
	}
	price := offer.GetPrice()
	err := s.store.UpsertOffer(ctx, model.Offer{
		ID:           offer.GetId(),
		RestaurantID: offer.GetRestaurantId(),
		ProviderID:   offer.GetProviderId(),
		MenuItemID:   offer.GetMenuItemId(),
		Price: model.Money{
			AmountCents: price.GetAmountCents(),
			Currency:    price.GetCurrency(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "record offer: %v", err)
	}
	return &ingestv1.RecordOfferResponse{Id: offer.GetId()}, nil
}

func (s *Server) RecordPriceObservation(ctx context.Context, req *ingestv1.RecordPriceObservationRequest) (*ingestv1.RecordPriceObservationResponse, error) {
	obs := req.GetPriceObservation()
	if obs == nil || obs.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "price_observation.id is required")
	}
	price := obs.GetPrice()
	observedAt := time.UnixMilli(obs.GetObservedAtUnixMs())
	if obs.GetObservedAtUnixMs() == 0 {
		observedAt = time.Now().UTC()
	}
	err := s.store.UpsertPriceObservation(ctx, model.PriceObservation{
		ID:      obs.GetId(),
		OfferID: obs.GetOfferId(),
		Price: model.Money{
			AmountCents: price.GetAmountCents(),
			Currency:    price.GetCurrency(),
		},
		ObservedAt: observedAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "record price observation: %v", err)
	}
	return &ingestv1.RecordPriceObservationResponse{Id: obs.GetId()}, nil
}
