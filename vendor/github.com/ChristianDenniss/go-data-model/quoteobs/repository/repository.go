package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/quoteobs/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Observation, error)
	Latest(ctx context.Context, sourceStoreID, dropoffGeohash, fulfillmentMode, deliveryExecutor, membershipTier string, basketSubtotalCents int64) (entity.Observation, error)
	Insert(ctx context.Context, obs entity.Observation) error
}
