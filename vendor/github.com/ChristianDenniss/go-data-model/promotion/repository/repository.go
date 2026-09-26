package repository

import (
	"context"
	"time"

	"github.com/ChristianDenniss/go-data-model/promotion/entity"
)

type PromotionRepository interface {
	GetByID(ctx context.Context, id string) (entity.Promotion, error)
	ListActiveByChannel(ctx context.Context, channelID string, at time.Time) ([]entity.Promotion, error)
	Upsert(ctx context.Context, p entity.Promotion) error
}

type MembershipProductRepository interface {
	GetByID(ctx context.Context, id string) (entity.MembershipProduct, error)
	Upsert(ctx context.Context, p entity.MembershipProduct) error
}
