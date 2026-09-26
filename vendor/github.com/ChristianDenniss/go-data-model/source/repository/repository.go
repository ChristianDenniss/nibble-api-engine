package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/source/entity"
)

type StoreRepository interface {
	GetByID(ctx context.Context, id string) (entity.Store, error)
	GetByChannelExternal(ctx context.Context, channelID, externalStoreID string) (entity.Store, error)
	Upsert(ctx context.Context, store entity.Store) error
}

type MenuRepository interface {
	GetByID(ctx context.Context, id string) (entity.Menu, error)
	ListByStore(ctx context.Context, sourceStoreID string) ([]entity.Menu, error)
	Upsert(ctx context.Context, menu entity.Menu) error
}

type CategoryRepository interface {
	GetByID(ctx context.Context, id string) (entity.Category, error)
	ListByMenu(ctx context.Context, sourceMenuID string) ([]entity.Category, error)
	Upsert(ctx context.Context, category entity.Category) error
}

type ItemRepository interface {
	GetByID(ctx context.Context, id string) (entity.Item, error)
	ListByCategory(ctx context.Context, sourceCategoryID string) ([]entity.Item, error)
	Upsert(ctx context.Context, item entity.Item) error
}
