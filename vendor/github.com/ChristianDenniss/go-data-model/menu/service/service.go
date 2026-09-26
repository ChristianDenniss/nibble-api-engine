package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/menu/entity"
	"github.com/ChristianDenniss/go-data-model/menu/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, item entity.Item) error {
	if item.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, item)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Item, error) {
	if id == "" {
		return entity.Item{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
