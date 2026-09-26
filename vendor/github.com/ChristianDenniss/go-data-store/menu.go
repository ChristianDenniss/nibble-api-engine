package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/menu/entity"
	"github.com/ChristianDenniss/go-data-model/menu/repository"
)

var _ repository.Repository = (*MenuRepository)(nil)

type MenuRepository struct {
	db *DB
}

func NewMenuRepository(db *DB) *MenuRepository {
	return &MenuRepository{db: db}
}

type menuItemRow struct {
	ID           string
	RestaurantID string
	Name         string
	Description  string
	Section      string
}

func (row menuItemRow) toDomain() entity.Item {
	return entity.Item{
		ID:           row.ID,
		RestaurantID: row.RestaurantID,
		Name:         row.Name,
		Description:  row.Description,
		Section:      row.Section,
	}
}

func (r *MenuRepository) GetByID(ctx context.Context, id string) (entity.Item, error) {
	var row menuItemRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, restaurant_id, name, description, section
		FROM menu_items
		WHERE id = $1`, id).Scan(&row.ID, &row.RestaurantID, &row.Name, &row.Description, &row.Section)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Item{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Item{}, err
	}
	return row.toDomain(), nil
}

func (r *MenuRepository) Upsert(ctx context.Context, item entity.Item) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO menu_items (id, restaurant_id, name, description, section)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			restaurant_id = EXCLUDED.restaurant_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			section = EXCLUDED.section,
			updated_at = now()`,
		item.ID, item.RestaurantID, item.Name, item.Description, item.Section)
	return err
}
