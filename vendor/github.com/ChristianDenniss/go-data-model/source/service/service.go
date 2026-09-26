package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/source/entity"
	"github.com/ChristianDenniss/go-data-model/source/repository"
)

type Service struct {
	stores     repository.StoreRepository
	menus      repository.MenuRepository
	categories repository.CategoryRepository
	items      repository.ItemRepository
}

func New(
	stores repository.StoreRepository,
	menus repository.MenuRepository,
	categories repository.CategoryRepository,
	items repository.ItemRepository,
) *Service {
	return &Service{
		stores:     stores,
		menus:      menus,
		categories: categories,
		items:      items,
	}
}

func (s *Service) RecordStore(ctx context.Context, store entity.Store) error {
	if store.ID == "" {
		return entity.ErrIDRequired
	}
	return s.stores.Upsert(ctx, store)
}

func (s *Service) RecordMenu(ctx context.Context, menu entity.Menu) error {
	if menu.ID == "" {
		return entity.ErrIDRequired
	}
	return s.menus.Upsert(ctx, menu)
}

func (s *Service) RecordCategory(ctx context.Context, cat entity.Category) error {
	if cat.ID == "" {
		return entity.ErrIDRequired
	}
	return s.categories.Upsert(ctx, cat)
}

func (s *Service) RecordItem(ctx context.Context, item entity.Item) error {
	if item.ID == "" {
		return entity.ErrIDRequired
	}
	return s.items.Upsert(ctx, item)
}

func (s *Service) GetStore(ctx context.Context, id string) (entity.Store, error) {
	if id == "" {
		return entity.Store{}, entity.ErrIDRequired
	}
	return s.stores.GetByID(ctx, id)
}

func (s *Service) GetItem(ctx context.Context, id string) (entity.Item, error) {
	if id == "" {
		return entity.Item{}, entity.ErrIDRequired
	}
	return s.items.GetByID(ctx, id)
}
