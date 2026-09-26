package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/restaurant/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Restaurant, error)
	Upsert(ctx context.Context, restaurant entity.Restaurant) error
}
