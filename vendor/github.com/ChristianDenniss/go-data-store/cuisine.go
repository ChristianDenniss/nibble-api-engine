package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/cuisine/entity"
	"github.com/ChristianDenniss/go-data-model/cuisine/repository"
)

var _ repository.Repository = (*CuisineRepository)(nil)

type CuisineRepository struct {
	db *DB
}

func NewCuisineRepository(db *DB) *CuisineRepository {
	return &CuisineRepository{db: db}
}

func (r *CuisineRepository) GetByID(ctx context.Context, id string) (entity.Cuisine, error) {
	var out entity.Cuisine
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, slug, name
		FROM cuisines
		WHERE id = $1`, id).Scan(&out.ID, &out.Slug, &out.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Cuisine{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Cuisine{}, err
	}
	return out, nil
}

func (r *CuisineRepository) Upsert(ctx context.Context, c entity.Cuisine) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO cuisines (id, slug, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			updated_at = now()`,
		c.ID, c.Slug, c.Name)
	return err
}
