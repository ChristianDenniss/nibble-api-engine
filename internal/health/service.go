package health

import (
	"context"

	"github.com/ChristianDenniss/api-engine/internal/apperror"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Service struct {
	db Pinger
}

func NewService(db Pinger) *Service {
	return &Service{db: db}
}

func (s *Service) Check(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		return apperror.Unavailable("database unavailable")
	}
	return nil
}
