package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/category/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Category, error)
	Upsert(ctx context.Context, category entity.Category) error
}
