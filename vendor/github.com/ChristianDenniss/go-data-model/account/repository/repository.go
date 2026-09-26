package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/account/entity"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (entity.Account, error)
	Upsert(ctx context.Context, account entity.Account) error
}
