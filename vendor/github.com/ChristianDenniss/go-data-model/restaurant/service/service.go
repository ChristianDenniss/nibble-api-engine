package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/restaurant/entity"
	"github.com/ChristianDenniss/go-data-model/restaurant/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, r entity.Restaurant) error {
	if r.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, r)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Restaurant, error) {
	if id == "" {
		return entity.Restaurant{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
