package postgres

import (
	"context"
	"database/sql"
	"errors"

	brandentity "github.com/ChristianDenniss/go-data-model/brand/entity"
	brandrepo "github.com/ChristianDenniss/go-data-model/brand/repository"
	dishentity "github.com/ChristianDenniss/go-data-model/dish/entity"
	dishrepo "github.com/ChristianDenniss/go-data-model/dish/repository"
	placeentity "github.com/ChristianDenniss/go-data-model/place/entity"
	placerepo "github.com/ChristianDenniss/go-data-model/place/repository"
)

var (
	_ brandrepo.Repository           = (*BrandRepository)(nil)
	_ placerepo.PlaceRepository      = (*PlaceRepository)(nil)
	_ placerepo.PurchaseOptionRepository = (*PurchaseOptionRepository)(nil)
	_ dishrepo.Repository            = (*DishRepository)(nil)
)

type BrandRepository struct {
	db *DB
}

func NewBrandRepository(db *DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) GetByID(ctx context.Context, id string) (brandentity.Brand, error) {
	var out brandentity.Brand
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, slug, name FROM brands WHERE id = $1`, id).
		Scan(&out.ID, &out.Slug, &out.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return brandentity.Brand{}, brandentity.ErrNotFound
	}
	return out, err
}

func (r *BrandRepository) Upsert(ctx context.Context, b brandentity.Brand) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO brands (id, slug, name) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET slug = EXCLUDED.slug, name = EXCLUDED.name, updated_at = now()`,
		b.ID, b.Slug, b.Name)
	return err
}

type PlaceRepository struct {
	db *DB
}

func NewPlaceRepository(db *DB) *PlaceRepository {
	return &PlaceRepository{db: db}
}

func (r *PlaceRepository) GetByID(ctx context.Context, id string) (placeentity.Place, error) {
	var out placeentity.Place
	var brandID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, brand_id, name, latitude, longitude, address, city, region, postal_code
		FROM places WHERE id = $1`, id).
		Scan(&out.ID, &brandID, &out.Name,
			&out.Location.Latitude, &out.Location.Longitude, &out.Location.Address,
			&out.Location.City, &out.Location.Region, &out.Location.PostalCode)
	if errors.Is(err, sql.ErrNoRows) {
		return placeentity.Place{}, placeentity.ErrNotFound
	}
	if err != nil {
		return placeentity.Place{}, err
	}
	if brandID.Valid {
		out.BrandID = brandID.String
	}
	return out, nil
}

func (r *PlaceRepository) Upsert(ctx context.Context, p placeentity.Place) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO places (id, brand_id, name, latitude, longitude, address, city, region, postal_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			brand_id = EXCLUDED.brand_id, name = EXCLUDED.name,
			latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
			address = EXCLUDED.address, city = EXCLUDED.city, region = EXCLUDED.region,
			postal_code = EXCLUDED.postal_code, updated_at = now()`,
		p.ID, nullString(p.BrandID), p.Name,
		p.Location.Latitude, p.Location.Longitude, p.Location.Address,
		p.Location.City, p.Location.Region, p.Location.PostalCode)
	return err
}

type PurchaseOptionRepository struct {
	db *DB
}

func NewPurchaseOptionRepository(db *DB) *PurchaseOptionRepository {
	return &PurchaseOptionRepository{db: db}
}

func (r *PurchaseOptionRepository) GetByID(ctx context.Context, id string) (placeentity.PurchaseOption, error) {
	var out placeentity.PurchaseOption
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, place_id, channel_id, fulfillment_mode, delivery_executor, source_store_id
		FROM place_purchase_options WHERE id = $1`, id).
		Scan(&out.ID, &out.PlaceID, &out.ChannelID, &out.FulfillmentMode, &out.DeliveryExecutor, &out.SourceStoreID)
	if errors.Is(err, sql.ErrNoRows) {
		return placeentity.PurchaseOption{}, placeentity.ErrNotFound
	}
	return out, err
}

func (r *PurchaseOptionRepository) ListByPlace(ctx context.Context, placeID string) ([]placeentity.PurchaseOption, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, place_id, channel_id, fulfillment_mode, delivery_executor, source_store_id
		FROM place_purchase_options WHERE place_id = $1`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []placeentity.PurchaseOption
	for rows.Next() {
		var o placeentity.PurchaseOption
		if err := rows.Scan(&o.ID, &o.PlaceID, &o.ChannelID, &o.FulfillmentMode, &o.DeliveryExecutor, &o.SourceStoreID); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *PurchaseOptionRepository) Upsert(ctx context.Context, opt placeentity.PurchaseOption) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO place_purchase_options (id, place_id, channel_id, fulfillment_mode, delivery_executor, source_store_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			place_id = EXCLUDED.place_id, channel_id = EXCLUDED.channel_id,
			fulfillment_mode = EXCLUDED.fulfillment_mode, delivery_executor = EXCLUDED.delivery_executor,
			source_store_id = EXCLUDED.source_store_id,
			updated_at = now()`,
		opt.ID, opt.PlaceID, opt.ChannelID, opt.FulfillmentMode, opt.DeliveryExecutor, opt.SourceStoreID)
	return err
}

type DishRepository struct {
	db *DB
}

func NewDishRepository(db *DB) *DishRepository {
	return &DishRepository{db: db}
}

func (r *DishRepository) GetByID(ctx context.Context, id string) (dishentity.Dish, error) {
	var out dishentity.Dish
	var brandID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, brand_id, name, canonical_name, description FROM dishes WHERE id = $1`, id).
		Scan(&out.ID, &brandID, &out.Name, &out.CanonicalName, &out.Description)
	if errors.Is(err, sql.ErrNoRows) {
		return dishentity.Dish{}, dishentity.ErrNotFound
	}
	if err != nil {
		return dishentity.Dish{}, err
	}
	if brandID.Valid {
		out.BrandID = brandID.String
	}
	return out, nil
}

func (r *DishRepository) Upsert(ctx context.Context, d dishentity.Dish) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO dishes (id, brand_id, name, canonical_name, description)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			brand_id = EXCLUDED.brand_id, name = EXCLUDED.name,
			canonical_name = EXCLUDED.canonical_name, description = EXCLUDED.description,
			updated_at = now()`,
		d.ID, nullString(d.BrandID), d.Name, d.CanonicalName, d.Description)
	return err
}
