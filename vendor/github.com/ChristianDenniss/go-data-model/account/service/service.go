package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/account/entity"
	"github.com/ChristianDenniss/go-data-model/account/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, a entity.Account) error {
	if a.ID == "" {
		return entity.ErrIDRequired
	}
	return s.repo.Upsert(ctx, a)
}

func (s *Service) GetByID(ctx context.Context, id string) (entity.Account, error) {
	if id == "" {
		return entity.Account{}, entity.ErrIDRequired
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) RemoveSavedAddress(ctx context.Context, accountID, addressID string) error {
	if accountID == "" {
		return entity.ErrIDRequired
	}
	if addressID == "" {
		return entity.ErrAddressIDRequired
	}
	return s.repo.DeleteSavedAddress(ctx, accountID, addressID)
}

func (s *Service) SelectSavedAddress(ctx context.Context, accountID, addressID string) error {
	if accountID == "" {
		return entity.ErrIDRequired
	}
	if addressID == "" {
		return entity.ErrAddressIDRequired
	}
	return s.repo.SetCurrentSavedAddress(ctx, accountID, addressID)
}
