package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/provider/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Provider, error)
	Upsert(ctx context.Context, provider entity.Provider) error
}
