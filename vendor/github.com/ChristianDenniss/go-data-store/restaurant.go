package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/restaurant/entity"
	"github.com/ChristianDenniss/go-data-model/restaurant/repository"
)

var _ repository.Repository = (*RestaurantRepository)(nil)

type RestaurantRepository struct {
	db *DB
}

func NewRestaurantRepository(db *DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id string) (entity.Restaurant, error) {
	var out entity.Restaurant
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, name, latitude, longitude, address, city, region, postal_code, rating_average, rating_count
		FROM restaurants
		WHERE id = $1`, id).Scan(
		&out.ID, &out.Name,
		&out.Location.Latitude, &out.Location.Longitude, &out.Location.Address,
		&out.Location.City, &out.Location.Region, &out.Location.PostalCode,
		&out.Rating.Average, &out.Rating.Count)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Restaurant{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Restaurant{}, err
	}

	out.CuisineIDs, err = listIDs(ctx, r.db.sql, `SELECT cuisine_id FROM restaurant_cuisines WHERE restaurant_id = $1`, id)
	if err != nil {
		return entity.Restaurant{}, err
	}
	out.CategoryIDs, err = listIDs(ctx, r.db.sql, `SELECT category_id FROM restaurant_categories WHERE restaurant_id = $1`, id)
	if err != nil {
		return entity.Restaurant{}, err
	}
	return out, nil
}

func (r *RestaurantRepository) Upsert(ctx context.Context, in entity.Restaurant) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO restaurants (id, name, latitude, longitude, address, city, region, postal_code, rating_average, rating_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			region = EXCLUDED.region,
			postal_code = EXCLUDED.postal_code,
			rating_average = EXCLUDED.rating_average,
			rating_count = EXCLUDED.rating_count,
			updated_at = now()`,
		in.ID, in.Name,
		in.Location.Latitude, in.Location.Longitude, in.Location.Address,
		in.Location.City, in.Location.Region, in.Location.PostalCode,
		in.Rating.Average, in.Rating.Count)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM restaurant_cuisines WHERE restaurant_id = $1`, in.ID); err != nil {
		return err
	}
	for _, cuisineID := range in.CuisineIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO restaurant_cuisines (restaurant_id, cuisine_id) VALUES ($1, $2)`, in.ID, cuisineID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM restaurant_categories WHERE restaurant_id = $1`, in.ID); err != nil {
		return err
	}
	for _, categoryID := range in.CategoryIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO restaurant_categories (restaurant_id, category_id) VALUES ($1, $2)`, in.ID, categoryID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func listIDs(ctx context.Context, db *sql.DB, query, id string) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		ids = append(ids, value)
	}
	return ids, rows.Err()
}
