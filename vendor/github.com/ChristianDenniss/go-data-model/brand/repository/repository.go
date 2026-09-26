package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/brand/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Brand, error)
	Upsert(ctx context.Context, brand entity.Brand) error
}
