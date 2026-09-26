package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/cuisine/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Cuisine, error)
	Upsert(ctx context.Context, cuisine entity.Cuisine) error
}
