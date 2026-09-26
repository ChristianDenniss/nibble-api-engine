package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/resolution/entity"
)

type StoreMatchRepository interface {
	GetByID(ctx context.Context, id string) (entity.StoreMatch, error)
	ListByPlace(ctx context.Context, placeID string) ([]entity.StoreMatch, error)
	Upsert(ctx context.Context, m entity.StoreMatch) error
}

type ItemMatchRepository interface {
	GetByID(ctx context.Context, id string) (entity.ItemMatch, error)
	ListByDish(ctx context.Context, dishID string) ([]entity.ItemMatch, error)
	FindSourceItemForStoreAndDish(ctx context.Context, sourceStoreID, dishID string) (string, error)
	Upsert(ctx context.Context, m entity.ItemMatch) error
}

type EvidenceRepository interface {
	Insert(ctx context.Context, e entity.Evidence) error
}
