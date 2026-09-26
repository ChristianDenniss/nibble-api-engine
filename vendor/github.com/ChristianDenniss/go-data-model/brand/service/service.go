package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/brand/entity"
	"github.com/ChristianDenniss/go-data-model/brand/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, b entity.Brand) error {
	if b.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, b)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Brand, error) {
	if id == "" {
		return entity.Brand{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
