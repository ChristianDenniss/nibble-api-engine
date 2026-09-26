package postgres

import (
	"context"
	"database/sql"
	"errors"

	itempriceentity "github.com/ChristianDenniss/go-data-model/itemprice/entity"
	"github.com/ChristianDenniss/go-data-model/source/entity"
	"github.com/ChristianDenniss/go-data-model/source/repository"
)

var _ repository.BrowseRepository = (*SourceBrowseRepository)(nil)

type SourceBrowseRepository struct {
	db *DB
}

func NewSourceBrowseRepository(db *DB) *SourceBrowseRepository {
	return &SourceBrowseRepository{db: db}
}

func (r *SourceBrowseRepository) LoadMenuBrowse(ctx context.Context, sourceStoreID, fulfillmentMode, deliveryExecutor string) (entity.MenuBrowse, error) {
	storeRepo := NewSourceStoreRepository(r.db)
	catRepo := NewSourceCategoryRepository(r.db)
	itemRepo := NewSourceItemRepository(r.db)
	priceRepo := NewItemPriceObservationRepository(r.db)

	store, err := storeRepo.GetByID(ctx, sourceStoreID)
	if err != nil {
		return entity.MenuBrowse{}, err
	}

	var menu entity.Menu
	err = r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_store_id, fulfillment_mode, delivery_executor, external_menu_id
		FROM source_menus
		WHERE source_store_id = $1 AND fulfillment_mode = $2 AND delivery_executor = $3`,
		sourceStoreID, fulfillmentMode, deliveryExecutor).
		Scan(&menu.ID, &menu.SourceStoreID, &menu.FulfillmentMode, &menu.DeliveryExecutor, &menu.ExternalMenuID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.MenuBrowse{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.MenuBrowse{}, err
	}

	categories, err := catRepo.ListByMenu(ctx, menu.ID)
	if err != nil {
		return entity.MenuBrowse{}, err
	}

	out := entity.MenuBrowse{Store: store, Menu: menu}
	for _, cat := range categories {
		items, err := itemRepo.ListByCategory(ctx, cat.ID)
		if err != nil {
			return entity.MenuBrowse{}, err
		}
		views := make([]entity.MenuItemView, 0, len(items))
		for _, it := range items {
			view := entity.MenuItemView{Item: it}
			obs, err := priceRepo.LatestByItem(ctx, it.ID, fulfillmentMode, deliveryExecutor)
			if err == nil {
				view.PriceCents = obs.Price.AmountCents
				view.Currency = obs.Price.Currency
			} else if !errors.Is(err, itempriceentity.ErrNotFound) {
				return entity.MenuBrowse{}, err
			}
			views = append(views, view)
		}
		out.Categories = append(out.Categories, entity.CategoryWithItems{
			Category: cat,
			Items:    views,
		})
	}
	return out, nil
}
