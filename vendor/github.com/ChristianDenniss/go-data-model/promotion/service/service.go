package service

import (
	"context"
	"time"

	"github.com/ChristianDenniss/go-data-model/promotion/entity"
	"github.com/ChristianDenniss/go-data-model/promotion/repository"
)

type Service struct {
	promos      repository.PromotionRepository
	memberships repository.MembershipProductRepository
}

func New(promos repository.PromotionRepository, memberships repository.MembershipProductRepository) *Service {
	return &Service{promos: promos, memberships: memberships}
}

func (s *Service) RecordPromotion(ctx context.Context, p entity.Promotion) error {
	if p.ID == "" {
		return entity.ErrIDRequired
	}
	return s.promos.Upsert(ctx, p)
}

func (s *Service) ActivePromotions(ctx context.Context, channelID string, at time.Time) ([]entity.Promotion, error) {
	return s.promos.ListActiveByChannel(ctx, channelID, at)
}
