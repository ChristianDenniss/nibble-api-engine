package ingest

import (
	"context"
	"errors"
	"time"

	model "github.com/ChristianDenniss/go-data-model"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ingestv1.UnimplementedIngestServiceServer
	svc *Service
}

func NewServer(svc *Service) *Server {
	return &Server{svc: svc}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrRestaurantIDRequired) ||
		errors.Is(err, ErrMenuItemIDRequired) ||
		errors.Is(err, ErrOfferIDRequired) ||
		errors.Is(err, ErrPriceObservationIDRequired) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}

func (s *Server) RecordRestaurant(ctx context.Context, req *ingestv1.RecordRestaurantRequest) (*ingestv1.RecordRestaurantResponse, error) {
	r := req.GetRestaurant()
	if r == nil {
		return nil, mapError(ErrRestaurantIDRequired)
	}
	loc := r.GetLocation()
	err := s.svc.RecordRestaurant(ctx, model.Restaurant{
		ID:   r.GetId(),
		Name: r.GetName(),
		Location: model.Location{
			Latitude:  loc.GetLatitude(),
			Longitude: loc.GetLongitude(),
			Address:   loc.GetAddress(),
		},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv1.RecordRestaurantResponse{Id: r.GetId()}, nil
}

func (s *Server) RecordMenuItem(ctx context.Context, req *ingestv1.RecordMenuItemRequest) (*ingestv1.RecordMenuItemResponse, error) {
	item := req.GetMenuItem()
	if item == nil {
		return nil, mapError(ErrMenuItemIDRequired)
	}
	err := s.svc.RecordMenuItem(ctx, model.MenuItem{
		ID:           item.GetId(),
		RestaurantID: item.GetRestaurantId(),
		Name:         item.GetName(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv1.RecordMenuItemResponse{Id: item.GetId()}, nil
}

func (s *Server) RecordOffer(ctx context.Context, req *ingestv1.RecordOfferRequest) (*ingestv1.RecordOfferResponse, error) {
	offer := req.GetOffer()
	if offer == nil {
		return nil, mapError(ErrOfferIDRequired)
	}
	price := offer.GetPrice()
	err := s.svc.RecordOffer(ctx, model.Offer{
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
		return nil, mapError(err)
	}
	return &ingestv1.RecordOfferResponse{Id: offer.GetId()}, nil
}

func (s *Server) RecordPriceObservation(ctx context.Context, req *ingestv1.RecordPriceObservationRequest) (*ingestv1.RecordPriceObservationResponse, error) {
	obs := req.GetPriceObservation()
	if obs == nil {
		return nil, mapError(ErrPriceObservationIDRequired)
	}
	price := obs.GetPrice()
	var observedAt time.Time
	if ms := obs.GetObservedAtUnixMs(); ms != 0 {
		observedAt = time.UnixMilli(ms)
	}
	err := s.svc.RecordPriceObservation(ctx, model.PriceObservation{
		ID:      obs.GetId(),
		OfferID: obs.GetOfferId(),
		Price: model.Money{
			AmountCents: price.GetAmountCents(),
			Currency:    price.GetCurrency(),
		},
		ObservedAt: observedAt,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv1.RecordPriceObservationResponse{Id: obs.GetId()}, nil
}
