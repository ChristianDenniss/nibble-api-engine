package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/place/entity"
	"github.com/ChristianDenniss/go-data-model/place/repository"
)

type Service struct {
	places  repository.PlaceRepository
	options repository.PurchaseOptionRepository
}

func New(places repository.PlaceRepository, options repository.PurchaseOptionRepository) *Service {
	return &Service{places: places, options: options}
}

func (s *Service) RecordPlace(ctx context.Context, p entity.Place) error {
	if p.ID == "" {
		return entity.ErrIDRequired
	}
	return s.places.Upsert(ctx, p)
}

func (s *Service) RecordPurchaseOption(ctx context.Context, opt entity.PurchaseOption) error {
	if opt.ID == "" {
		return entity.ErrIDRequired
	}
	return s.options.Upsert(ctx, opt)
}

func (s *Service) GetPlace(ctx context.Context, id string) (entity.Place, error) {
	if id == "" {
		return entity.Place{}, entity.ErrIDRequired
	}
	return s.places.GetByID(ctx, id)
}

func (s *Service) ListPurchaseOptions(ctx context.Context, placeID string) ([]entity.PurchaseOption, error) {
	if placeID == "" {
		return nil, entity.ErrIDRequired
	}
	return s.options.ListByPlace(ctx, placeID)
}
