package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/itemprice/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Observation, error)
	LatestByItem(ctx context.Context, sourceItemID, fulfillmentMode, deliveryExecutor string) (entity.Observation, error)
	Insert(ctx context.Context, obs entity.Observation) error
}
