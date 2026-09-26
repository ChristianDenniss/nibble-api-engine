package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/dish/entity"
	"github.com/ChristianDenniss/go-data-model/dish/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, d entity.Dish) error {
	if d.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, d)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Dish, error) {
	if id == "" {
		return entity.Dish{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
