package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/offer/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Offer, error)
	Upsert(ctx context.Context, offer entity.Offer) error
}
