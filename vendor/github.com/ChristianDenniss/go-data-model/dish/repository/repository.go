package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/dish/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Dish, error)
	Upsert(ctx context.Context, dish entity.Dish) error
}
