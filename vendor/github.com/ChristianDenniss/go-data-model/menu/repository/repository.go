package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/menu/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Item, error)
	Upsert(ctx context.Context, item entity.Item) error
}
