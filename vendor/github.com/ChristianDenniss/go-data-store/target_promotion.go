package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ChristianDenniss/go-data-model/promotion/entity"
	"github.com/ChristianDenniss/go-data-model/promotion/repository"
)

var (
	_ repository.PromotionRepository         = (*PromotionRepository)(nil)
	_ repository.MembershipProductRepository = (*MembershipProductRepository)(nil)
)

type PromotionRepository struct {
	db *DB
}

func NewPromotionRepository(db *DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func (r *PromotionRepository) GetByID(ctx context.Context, id string) (entity.Promotion, error) {
	var out entity.Promotion
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, channel_id, name, kind, value_cents, currency, value_bps, starts_at, ends_at
		FROM promotions WHERE id = $1`, id).
		Scan(&out.ID, &out.ChannelID, &out.Name, &out.Kind, &out.Value.AmountCents, &out.Value.Currency, &out.ValueBPS, &out.StartsAt, &out.EndsAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Promotion{}, entity.ErrNotFound
	}
	return out, err
}

func (r *PromotionRepository) ListActiveByChannel(ctx context.Context, channelID string, at time.Time) ([]entity.Promotion, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, channel_id, name, kind, value_cents, currency, value_bps, starts_at, ends_at
		FROM promotions WHERE channel_id = $1 AND starts_at <= $2 AND ends_at >= $2`, channelID, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Promotion
	for rows.Next() {
		var p entity.Promotion
		if err := rows.Scan(&p.ID, &p.ChannelID, &p.Name, &p.Kind, &p.Value.AmountCents, &p.Value.Currency, &p.ValueBPS, &p.StartsAt, &p.EndsAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PromotionRepository) Upsert(ctx context.Context, p entity.Promotion) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO promotions (id, channel_id, name, kind, value_cents, currency, value_bps, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			channel_id = EXCLUDED.channel_id, name = EXCLUDED.name, kind = EXCLUDED.kind,
			value_cents = EXCLUDED.value_cents, currency = EXCLUDED.currency, value_bps = EXCLUDED.value_bps,
			starts_at = EXCLUDED.starts_at, ends_at = EXCLUDED.ends_at, updated_at = now()`,
		p.ID, p.ChannelID, p.Name, p.Kind, p.Value.AmountCents, p.Value.Currency, p.ValueBPS, p.StartsAt, p.EndsAt)
	return err
}

type MembershipProductRepository struct {
	db *DB
}

func NewMembershipProductRepository(db *DB) *MembershipProductRepository {
	return &MembershipProductRepository{db: db}
}

func (r *MembershipProductRepository) GetByID(ctx context.Context, id string) (entity.MembershipProduct, error) {
	var out entity.MembershipProduct
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, channel_id, name, slug FROM membership_products WHERE id = $1`, id).
		Scan(&out.ID, &out.ChannelID, &out.Name, &out.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.MembershipProduct{}, entity.ErrNotFound
	}
	return out, err
}

func (r *MembershipProductRepository) Upsert(ctx context.Context, p entity.MembershipProduct) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO membership_products (id, channel_id, name, slug) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET channel_id = EXCLUDED.channel_id, name = EXCLUDED.name, slug = EXCLUDED.slug`,
		p.ID, p.ChannelID, p.Name, p.Slug)
	return err
}
