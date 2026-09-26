package health

import (
	"context"

	"github.com/ChristianDenniss/api-engine/internal/apperror"
	"github.com/ChristianDenniss/api-engine/internal/store"
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) Check(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		return apperror.Unavailable("database unavailable")
	}
	return nil
}
