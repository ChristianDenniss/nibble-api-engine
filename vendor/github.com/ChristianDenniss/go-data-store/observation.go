package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	money "github.com/ChristianDenniss/go-data-model/money/entity"
	"github.com/ChristianDenniss/go-data-model/observation/entity"
	"github.com/ChristianDenniss/go-data-model/observation/repository"
)

var _ repository.Repository = (*ObservationRepository)(nil)

type ObservationRepository struct {
	db *DB
}

func NewObservationRepository(db *DB) *ObservationRepository {
	return &ObservationRepository{db: db}
}

type observationRow struct {
	ID          string
	OfferID     string
	AmountCents int64
	Currency    string
	ObservedAt  time.Time
}

func (row observationRow) toDomain() entity.Observation {
	return entity.Observation{
		ID:      row.ID,
		OfferID: row.OfferID,
		Price: money.Money{
			AmountCents: row.AmountCents,
			Currency:    row.Currency,
		},
		ObservedAt: row.ObservedAt,
	}
}

func (r *ObservationRepository) GetByID(ctx context.Context, id string) (entity.Observation, error) {
	var row observationRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, offer_id, amount_cents, currency, observed_at
		FROM price_observations
		WHERE id = $1`, id).Scan(
		&row.ID, &row.OfferID, &row.AmountCents, &row.Currency, &row.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Observation{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Observation{}, err
	}
	return row.toDomain(), nil
}

func (r *ObservationRepository) Upsert(ctx context.Context, obs entity.Observation) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO price_observations (id, offer_id, amount_cents, currency, observed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			offer_id = EXCLUDED.offer_id,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			observed_at = EXCLUDED.observed_at`,
		obs.ID, obs.OfferID, obs.Price.AmountCents, obs.Price.Currency, obs.ObservedAt)
	return err
}
