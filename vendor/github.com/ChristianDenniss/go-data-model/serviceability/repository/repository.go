package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/serviceability/entity"
)

type StoreStatusRepository interface {
	Get(ctx context.Context, sourceStoreID string) (entity.StoreStatus, error)
	Upsert(ctx context.Context, status entity.StoreStatus) error
}

type ServiceAreaRepository interface {
	ListForPath(ctx context.Context, sourceStoreID, fulfillmentMode, deliveryExecutor string) ([]entity.ServiceArea, error)
	Upsert(ctx context.Context, area entity.ServiceArea) error
}
