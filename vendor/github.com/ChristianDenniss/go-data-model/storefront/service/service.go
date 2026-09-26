package service

import (
	"context"
	"errors"

	accountentity "github.com/ChristianDenniss/go-data-model/account/entity"
	"github.com/ChristianDenniss/go-data-model/storefront/entity"
	"github.com/ChristianDenniss/go-data-model/storefront/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LoadCatalog(ctx context.Context, accountID string) (entity.Catalog, error) {
	if accountID == "" {
		return entity.Catalog{}, entity.ErrAccountIDRequired
	}
	catalog, err := s.repo.LoadCatalog(ctx, accountID)
	if err != nil {
		if errors.Is(err, accountentity.ErrNotFound) {
			return entity.Catalog{}, entity.ErrNotFound
		}
		return entity.Catalog{}, err
	}
	return catalog, nil
}
