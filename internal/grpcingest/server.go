package grpcingest

import (
	"context"
	"errors"
	"time"

	location "github.com/ChristianDenniss/go-data-model/location/entity"
	menuentity "github.com/ChristianDenniss/go-data-model/menu/entity"
	menusvc "github.com/ChristianDenniss/go-data-model/menu/service"
	money "github.com/ChristianDenniss/go-data-model/money/entity"
	obsentity "github.com/ChristianDenniss/go-data-model/observation/entity"
	obssvc "github.com/ChristianDenniss/go-data-model/observation/service"
	offerentity "github.com/ChristianDenniss/go-data-model/offer/entity"
	offersvc "github.com/ChristianDenniss/go-data-model/offer/service"
	restaurantentity "github.com/ChristianDenniss/go-data-model/restaurant/entity"
	restaurantsvc "github.com/ChristianDenniss/go-data-model/restaurant/service"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ingestv1.UnimplementedIngestServiceServer
	restaurants  *restaurantsvc.Service
	menuItems    *menusvc.Service
	offers       *offersvc.Service
	observations *obssvc.Service
}

func NewServer(
	restaurants *restaurantsvc.Service,
	menuItems *menusvc.Service,
	offers *offersvc.Service,
	observations *obssvc.Service,
) *Server {
	return &Server{
		restaurants:  restaurants,
		menuItems:    menuItems,
		offers:       offers,
		observations: observations,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, restaurantentity.ErrIDRequired) ||
		errors.Is(err, menuentity.ErrIDRequired) ||
		errors.Is(err, offerentity.ErrIDRequired) ||
		errors.Is(err, obsentity.ErrIDRequired) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, restaurantentity.ErrNotFound) ||
		errors.Is(err, menuentity.ErrNotFound) ||
		errors.Is(err, offerentity.ErrNotFound) ||
		errors.Is(err, obsentity.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}

func (s *Server) RecordRestaurant(ctx context.Context, req *ingestv1.RecordRestaurantRequest) (*ingestv1.RecordRestaurantResponse, error) {
	r := req.GetRestaurant()
	if r == nil {
		return nil, mapError(restaurantentity.ErrIDRequired)
	}
	loc := r.GetLocation()
	err := s.restaurants.Record(ctx, restaurantentity.Restaurant{
		ID:   r.GetId(),
		Name: r.GetName(),
		Location: location.Location{
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
		return nil, mapError(menuentity.ErrIDRequired)
	}
	err := s.menuItems.Record(ctx, menuentity.Item{
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
	o := req.GetOffer()
	if o == nil {
		return nil, mapError(offerentity.ErrIDRequired)
	}
	price := o.GetPrice()
	err := s.offers.Record(ctx, offerentity.Offer{
		ID:           o.GetId(),
		RestaurantID: o.GetRestaurantId(),
		ProviderID:   o.GetProviderId(),
		MenuItemID:   o.GetMenuItemId(),
		Price: money.Money{
			AmountCents: price.GetAmountCents(),
			Currency:    price.GetCurrency(),
		},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv1.RecordOfferResponse{Id: o.GetId()}, nil
}

func (s *Server) RecordPriceObservation(ctx context.Context, req *ingestv1.RecordPriceObservationRequest) (*ingestv1.RecordPriceObservationResponse, error) {
	obs := req.GetPriceObservation()
	if obs == nil {
		return nil, mapError(obsentity.ErrIDRequired)
	}
	price := obs.GetPrice()
	var observedAt time.Time
	if ms := obs.GetObservedAtUnixMs(); ms != 0 {
		observedAt = time.UnixMilli(ms)
	}
	err := s.observations.Record(ctx, obsentity.Observation{
		ID:      obs.GetId(),
		OfferID: obs.GetOfferId(),
		Price: money.Money{
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
