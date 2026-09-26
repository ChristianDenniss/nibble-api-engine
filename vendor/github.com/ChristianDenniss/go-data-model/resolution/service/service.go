package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/resolution/entity"
	"github.com/ChristianDenniss/go-data-model/resolution/repository"
)

type Service struct {
	storeMatches repository.StoreMatchRepository
	itemMatches  repository.ItemMatchRepository
	evidence     repository.EvidenceRepository
}

func New(
	storeMatches repository.StoreMatchRepository,
	itemMatches repository.ItemMatchRepository,
	evidence repository.EvidenceRepository,
) *Service {
	return &Service{
		storeMatches: storeMatches,
		itemMatches:  itemMatches,
		evidence:     evidence,
	}
}

func (s *Service) RecordStoreMatch(ctx context.Context, m entity.StoreMatch) error {
	if m.ID == "" {
		return entity.ErrIDRequired
	}
	return s.storeMatches.Upsert(ctx, m)
}

func (s *Service) RecordItemMatch(ctx context.Context, m entity.ItemMatch) error {
	if m.ID == "" {
		return entity.ErrIDRequired
	}
	return s.itemMatches.Upsert(ctx, m)
}

func (s *Service) ListStoreMatchesByPlace(ctx context.Context, placeID string) ([]entity.StoreMatch, error) {
	return s.storeMatches.ListByPlace(ctx, placeID)
}

func (s *Service) ResolveSourceItemForStoreAndDish(ctx context.Context, sourceStoreID, dishID string) (string, error) {
	if sourceStoreID == "" || dishID == "" {
		return "", entity.ErrIDRequired
	}
	return s.itemMatches.FindSourceItemForStoreAndDish(ctx, sourceStoreID, dishID)
}
