package service

import (
	"context"
	"errors"

	"github.com/ChristianDenniss/go-data-model/source/entity"
	"github.com/ChristianDenniss/go-data-model/source/repository"
)

type Service struct {
	stores     repository.StoreRepository
	menus      repository.MenuRepository
	categories repository.CategoryRepository
	items      repository.ItemRepository
	browse     repository.BrowseRepository
}

func New(
	stores repository.StoreRepository,
	menus repository.MenuRepository,
	categories repository.CategoryRepository,
	items repository.ItemRepository,
	browse repository.BrowseRepository,
) *Service {
	return &Service{
		stores:     stores,
		menus:      menus,
		categories: categories,
		items:      items,
		browse:     browse,
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

func (s *Service) LoadMenuBrowse(ctx context.Context, sourceStoreID, fulfillmentMode, deliveryExecutor string) (entity.MenuBrowse, error) {
	if sourceStoreID == "" || fulfillmentMode == "" {
		return entity.MenuBrowse{}, entity.ErrIDRequired
	}
	out, err := s.browse.LoadMenuBrowse(ctx, sourceStoreID, fulfillmentMode, deliveryExecutor)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.MenuBrowse{}, entity.ErrNotFound
		}
		return entity.MenuBrowse{}, err
	}
	return out, nil
}
