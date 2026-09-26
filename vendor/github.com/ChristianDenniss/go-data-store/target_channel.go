package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/channel/entity"
	"github.com/ChristianDenniss/go-data-model/channel/repository"
)

var _ repository.Repository = (*ChannelRepository)(nil)

type ChannelRepository struct {
	db *DB
}

func NewChannelRepository(db *DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (entity.Channel, error) {
	var out entity.Channel
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, slug, kind, name FROM channels WHERE id = $1`, id).
		Scan(&out.ID, &out.Slug, &out.Kind, &out.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Channel{}, entity.ErrNotFound
	}
	return out, err
}

func (r *ChannelRepository) GetBySlug(ctx context.Context, slug string) (entity.Channel, error) {
	var out entity.Channel
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, slug, kind, name FROM channels WHERE slug = $1`, slug).
		Scan(&out.ID, &out.Slug, &out.Kind, &out.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Channel{}, entity.ErrNotFound
	}
	return out, err
}

func (r *ChannelRepository) Upsert(ctx context.Context, ch entity.Channel) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO channels (id, slug, kind, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET slug = EXCLUDED.slug, kind = EXCLUDED.kind, name = EXCLUDED.name, updated_at = now()`,
		ch.ID, ch.Slug, ch.Kind, ch.Name)
	return err
}
