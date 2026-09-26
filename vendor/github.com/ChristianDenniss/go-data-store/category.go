package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/category/entity"
	"github.com/ChristianDenniss/go-data-model/category/repository"
)

var _ repository.Repository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	db *DB
}

func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (entity.Category, error) {
	var out entity.Category
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, slug, name, description
		FROM categories
		WHERE id = $1`, id).Scan(&out.ID, &out.Slug, &out.Name, &out.Description)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Category{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Category{}, err
	}
	return out, nil
}

func (r *CategoryRepository) Upsert(ctx context.Context, c entity.Category) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO categories (id, slug, name, description)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = now()`,
		c.ID, c.Slug, c.Name, c.Description)
	return err
}
