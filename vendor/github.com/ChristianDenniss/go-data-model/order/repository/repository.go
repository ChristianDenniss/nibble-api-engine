package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/order/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Order, error)
	Upsert(ctx context.Context, order entity.Order) error
}
