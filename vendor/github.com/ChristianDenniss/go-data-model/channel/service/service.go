package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/channel/entity"
	"github.com/ChristianDenniss/go-data-model/channel/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, ch entity.Channel) error {
	if ch.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, ch)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Channel, error) {
	if id == "" {
		return entity.Channel{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (entity.Channel, error) {
	if slug == "" {
		return entity.Channel{}, entity.ErrIDRequired
	}
	return s.repo.GetBySlug(ctx, slug)
}
