package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/market/entity"
)

type MarketRepository interface {
	GetByID(ctx context.Context, id string) (entity.Market, error)
	GetBySlug(ctx context.Context, slug string) (entity.Market, error)
	List(ctx context.Context) ([]entity.Market, error)
	Upsert(ctx context.Context, market entity.Market) error
}

type ProbeDropoffRepository interface {
	ListByMarket(ctx context.Context, marketID string) ([]entity.ProbeDropoff, error)
	Upsert(ctx context.Context, dropoff entity.ProbeDropoff) error
}

type CoverageRepository interface {
	Get(ctx context.Context, channelID, marketID string) (entity.ChannelCoverage, error)
	ListByMarket(ctx context.Context, marketID string) ([]entity.ChannelCoverage, error)
	Upsert(ctx context.Context, coverage entity.ChannelCoverage) error
}
