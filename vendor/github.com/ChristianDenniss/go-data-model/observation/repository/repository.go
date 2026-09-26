package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/observation/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Observation, error)
	Upsert(ctx context.Context, observation entity.Observation) error
}
