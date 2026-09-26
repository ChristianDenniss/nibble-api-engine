package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/channel/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Channel, error)
	GetBySlug(ctx context.Context, slug string) (entity.Channel, error)
	Upsert(ctx context.Context, ch entity.Channel) error
}
