package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/place/entity"
)

type PlaceRepository interface {
	GetByID(ctx context.Context, id string) (entity.Place, error)
	Upsert(ctx context.Context, place entity.Place) error
}

type PurchaseOptionRepository interface {
	GetByID(ctx context.Context, id string) (entity.PurchaseOption, error)
	ListByPlace(ctx context.Context, placeID string) ([]entity.PurchaseOption, error)
	Upsert(ctx context.Context, opt entity.PurchaseOption) error
}
