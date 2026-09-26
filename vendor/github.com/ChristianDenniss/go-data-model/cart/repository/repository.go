package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/cart/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Cart, error)
	Upsert(ctx context.Context, cart entity.Cart) error
}
