package postgres

import (
	"context"
	"database/sql"
	"errors"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
	"github.com/ChristianDenniss/go-data-model/offer/entity"
	"github.com/ChristianDenniss/go-data-model/offer/repository"
)

var _ repository.Repository = (*OfferRepository)(nil)

type OfferRepository struct {
	db *DB
}

func NewOfferRepository(db *DB) *OfferRepository {
	return &OfferRepository{db: db}
}

type offerRow struct {
	ID               string
	RestaurantID     string
	ProviderID       string
	MenuItemID       string
	AmountCents      int64
	Currency         string
	EstimatedMinutes int
}

func (row offerRow) toDomain() entity.Offer {
	return entity.Offer{
		ID:               row.ID,
		RestaurantID:     row.RestaurantID,
		ProviderID:       row.ProviderID,
		MenuItemID:       row.MenuItemID,
		EstimatedMinutes: row.EstimatedMinutes,
		Price: money.Money{
			AmountCents: row.AmountCents,
			Currency:    row.Currency,
		},
	}
}

func (r *OfferRepository) GetByID(ctx context.Context, id string) (entity.Offer, error) {
	var row offerRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, restaurant_id, provider_id, menu_item_id, amount_cents, currency, estimated_minutes
		FROM offers
		WHERE id = $1`, id).Scan(
		&row.ID, &row.RestaurantID, &row.ProviderID, &row.MenuItemID, &row.AmountCents, &row.Currency, &row.EstimatedMinutes)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Offer{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Offer{}, err
	}
	return row.toDomain(), nil
}

func (r *OfferRepository) Upsert(ctx context.Context, o entity.Offer) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO offers (id, restaurant_id, provider_id, menu_item_id, amount_cents, currency, estimated_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			restaurant_id = EXCLUDED.restaurant_id,
			provider_id = EXCLUDED.provider_id,
			menu_item_id = EXCLUDED.menu_item_id,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			estimated_minutes = EXCLUDED.estimated_minutes,
			updated_at = now()`,
		o.ID, o.RestaurantID, o.ProviderID, o.MenuItemID, o.Price.AmountCents, o.Price.Currency, o.EstimatedMinutes)
	return err
}
