package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/offer/entity"
	"github.com/ChristianDenniss/go-data-model/offer/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, o entity.Offer) error {
	if o.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, o)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Offer, error) {
	if id == "" {
		return entity.Offer{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}
