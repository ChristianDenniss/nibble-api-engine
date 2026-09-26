package service

import (
	"context"
	"time"

	"github.com/ChristianDenniss/go-data-model/observation/entity"
	"github.com/ChristianDenniss/go-data-model/observation/repository"
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
	if obs.ObservedAt.IsZero() {
		obs.ObservedAt = time.Now().UTC()
	}
	return s.repo.Upsert(ctx, obs)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Observation, error) {
	if id == "" {
		return entity.Observation{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
