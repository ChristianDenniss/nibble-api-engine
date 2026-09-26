package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/storefront/entity"
)

type Repository interface {
	LoadCatalog(ctx context.Context, accountID string) (entity.Catalog, error)
}
