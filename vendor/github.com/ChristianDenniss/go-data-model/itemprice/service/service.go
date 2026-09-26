package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/itemprice/entity"
	"github.com/ChristianDenniss/go-data-model/itemprice/repository"
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

func (s *Service) Latest(ctx context.Context, sourceItemID, fulfillmentMode, deliveryExecutor string) (entity.Observation, error) {
	return s.repo.LatestByItem(ctx, sourceItemID, fulfillmentMode, deliveryExecutor)
}
