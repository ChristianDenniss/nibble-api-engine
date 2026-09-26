package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/quoteobs/entity"
	"github.com/ChristianDenniss/go-data-model/quoteobs/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, obs entity.Observation) error {
	if obs.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Insert(ctx, obs)
}

func (s *Service) Latest(ctx context.Context, sourceStoreID, geohash, mode, deliveryExecutor, tier string, basketSubtotalCents int64) (entity.Observation, error) {
	return s.repo.Latest(ctx, sourceStoreID, geohash, mode, deliveryExecutor, tier, basketSubtotalCents)
}
