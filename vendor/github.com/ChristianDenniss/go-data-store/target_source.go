package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/source/entity"
	"github.com/ChristianDenniss/go-data-model/source/repository"
)

var (
	_ repository.StoreRepository     = (*SourceStoreRepository)(nil)
	_ repository.MenuRepository      = (*SourceMenuRepository)(nil)
	_ repository.CategoryRepository  = (*SourceCategoryRepository)(nil)
	_ repository.ItemRepository      = (*SourceItemRepository)(nil)
)

type SourceStoreRepository struct {
	db *DB
}

func NewSourceStoreRepository(db *DB) *SourceStoreRepository {
	return &SourceStoreRepository{db: db}
}

func (r *SourceStoreRepository) GetByID(ctx context.Context, id string) (entity.Store, error) {
	var out entity.Store
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, channel_id, external_store_id, name, latitude, longitude, address, phone
		FROM source_stores WHERE id = $1`, id).
		Scan(&out.ID, &out.ChannelID, &out.ExternalStoreID, &out.Name,
			&out.Location.Latitude, &out.Location.Longitude, &out.Location.Address, &out.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Store{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceStoreRepository) ListByChannel(ctx context.Context, channelID string) ([]entity.Store, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, channel_id, external_store_id, name, latitude, longitude, address, phone
		FROM source_stores WHERE channel_id = $1 ORDER BY name`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Store
	for rows.Next() {
		var st entity.Store
		if err := rows.Scan(&st.ID, &st.ChannelID, &st.ExternalStoreID, &st.Name,
			&st.Location.Latitude, &st.Location.Longitude, &st.Location.Address, &st.Phone); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (r *SourceStoreRepository) GetByChannelExternal(ctx context.Context, channelID, externalStoreID string) (entity.Store, error) {
	var out entity.Store
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, channel_id, external_store_id, name, latitude, longitude, address, phone
		FROM source_stores WHERE channel_id = $1 AND external_store_id = $2`, channelID, externalStoreID).
		Scan(&out.ID, &out.ChannelID, &out.ExternalStoreID, &out.Name,
			&out.Location.Latitude, &out.Location.Longitude, &out.Location.Address, &out.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Store{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceStoreRepository) Upsert(ctx context.Context, store entity.Store) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_stores (id, channel_id, external_store_id, name, latitude, longitude, address, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			channel_id = EXCLUDED.channel_id,
			external_store_id = EXCLUDED.external_store_id,
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			address = EXCLUDED.address,
			phone = EXCLUDED.phone,
			updated_at = now()`,
		store.ID, store.ChannelID, store.ExternalStoreID, store.Name,
		store.Location.Latitude, store.Location.Longitude, store.Location.Address, store.Phone)
	return err
}

type SourceMenuRepository struct {
	db *DB
}

func NewSourceMenuRepository(db *DB) *SourceMenuRepository {
	return &SourceMenuRepository{db: db}
}

func (r *SourceMenuRepository) GetByID(ctx context.Context, id string) (entity.Menu, error) {
	var out entity.Menu
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_store_id, fulfillment_mode, delivery_executor, external_menu_id FROM source_menus WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceStoreID, &out.FulfillmentMode, &out.DeliveryExecutor, &out.ExternalMenuID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Menu{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceMenuRepository) ListByStore(ctx context.Context, sourceStoreID string) ([]entity.Menu, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_store_id, fulfillment_mode, delivery_executor, external_menu_id FROM source_menus WHERE source_store_id = $1`, sourceStoreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Menu
	for rows.Next() {
		var m entity.Menu
		if err := rows.Scan(&m.ID, &m.SourceStoreID, &m.FulfillmentMode, &m.DeliveryExecutor, &m.ExternalMenuID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *SourceMenuRepository) Upsert(ctx context.Context, menu entity.Menu) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_menus (id, source_store_id, fulfillment_mode, delivery_executor, external_menu_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			source_store_id = EXCLUDED.source_store_id,
			fulfillment_mode = EXCLUDED.fulfillment_mode,
			delivery_executor = EXCLUDED.delivery_executor,
			external_menu_id = EXCLUDED.external_menu_id,
			updated_at = now()`,
		menu.ID, menu.SourceStoreID, menu.FulfillmentMode, menu.DeliveryExecutor, menu.ExternalMenuID)
	return err
}

type SourceCategoryRepository struct {
	db *DB
}

func NewSourceCategoryRepository(db *DB) *SourceCategoryRepository {
	return &SourceCategoryRepository{db: db}
}

func (r *SourceCategoryRepository) GetByID(ctx context.Context, id string) (entity.Category, error) {
	var out entity.Category
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_menu_id, external_category_id, name, sort_order FROM source_categories WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceMenuID, &out.ExternalCategoryID, &out.Name, &out.SortOrder)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Category{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceCategoryRepository) ListByMenu(ctx context.Context, sourceMenuID string) ([]entity.Category, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_menu_id, external_category_id, name, sort_order FROM source_categories WHERE source_menu_id = $1 ORDER BY sort_order`, sourceMenuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.SourceMenuID, &c.ExternalCategoryID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *SourceCategoryRepository) Upsert(ctx context.Context, category entity.Category) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_categories (id, source_menu_id, external_category_id, name, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			source_menu_id = EXCLUDED.source_menu_id,
			external_category_id = EXCLUDED.external_category_id,
			name = EXCLUDED.name,
			sort_order = EXCLUDED.sort_order,
			updated_at = now()`,
		category.ID, category.SourceMenuID, category.ExternalCategoryID, category.Name, category.SortOrder)
	return err
}

type SourceItemRepository struct {
	db *DB
}

func NewSourceItemRepository(db *DB) *SourceItemRepository {
	return &SourceItemRepository{db: db}
}

func (r *SourceItemRepository) GetByID(ctx context.Context, id string) (entity.Item, error) {
	var out entity.Item
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_category_id, external_item_id, name, description, available FROM source_items WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceCategoryID, &out.ExternalItemID, &out.Name, &out.Description, &out.Available)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Item{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceItemRepository) ListByCategory(ctx context.Context, sourceCategoryID string) ([]entity.Item, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_category_id, external_item_id, name, description, available FROM source_items WHERE source_category_id = $1`, sourceCategoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Item
	for rows.Next() {
		var it entity.Item
		if err := rows.Scan(&it.ID, &it.SourceCategoryID, &it.ExternalItemID, &it.Name, &it.Description, &it.Available); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *SourceItemRepository) Upsert(ctx context.Context, item entity.Item) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_items (id, source_category_id, external_item_id, name, description, available)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			source_category_id = EXCLUDED.source_category_id,
			external_item_id = EXCLUDED.external_item_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			available = EXCLUDED.available,
			updated_at = now()`,
		item.ID, item.SourceCategoryID, item.ExternalItemID, item.Name, item.Description, item.Available)
	return err
}
