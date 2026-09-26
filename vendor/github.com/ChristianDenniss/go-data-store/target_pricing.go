package postgres

import (
	"context"
	"database/sql"
	"errors"

	itempriceentity "github.com/ChristianDenniss/go-data-model/itemprice/entity"
	itempricerepo "github.com/ChristianDenniss/go-data-model/itemprice/repository"
	quoteentity "github.com/ChristianDenniss/go-data-model/quoteobs/entity"
	quoterepo "github.com/ChristianDenniss/go-data-model/quoteobs/repository"
)

var (
	_ itempricerepo.Repository = (*ItemPriceObservationRepository)(nil)
	_ quoterepo.Repository     = (*QuoteObservationRepository)(nil)
)

type ItemPriceObservationRepository struct {
	db *DB
}

func NewItemPriceObservationRepository(db *DB) *ItemPriceObservationRepository {
	return &ItemPriceObservationRepository{db: db}
}

func (r *ItemPriceObservationRepository) GetByID(ctx context.Context, id string) (itempriceentity.Observation, error) {
	var out itempriceentity.Observation
	var ingestID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_item_id, ingest_run_id, amount_cents, currency, fulfillment_mode, delivery_executor, observed_at
		FROM item_price_observations WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceItemID, &ingestID, &out.Price.AmountCents, &out.Price.Currency, &out.FulfillmentMode, &out.DeliveryExecutor, &out.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return itempriceentity.Observation{}, itempriceentity.ErrNotFound
	}
	if err != nil {
		return itempriceentity.Observation{}, err
	}
	if ingestID.Valid {
		out.IngestRunID = ingestID.String
	}
	return out, nil
}

func (r *ItemPriceObservationRepository) LatestByItem(ctx context.Context, sourceItemID, fulfillmentMode, deliveryExecutor string) (itempriceentity.Observation, error) {
	var out itempriceentity.Observation
	var ingestID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_item_id, ingest_run_id, amount_cents, currency, fulfillment_mode, delivery_executor, observed_at
		FROM item_price_observations
		WHERE source_item_id = $1 AND fulfillment_mode = $2 AND delivery_executor = $3
		ORDER BY observed_at DESC LIMIT 1`, sourceItemID, fulfillmentMode, deliveryExecutor).
		Scan(&out.ID, &out.SourceItemID, &ingestID, &out.Price.AmountCents, &out.Price.Currency, &out.FulfillmentMode, &out.DeliveryExecutor, &out.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return itempriceentity.Observation{}, itempriceentity.ErrNotFound
	}
	if err != nil {
		return itempriceentity.Observation{}, err
	}
	if ingestID.Valid {
		out.IngestRunID = ingestID.String
	}
	return out, nil
}

func (r *ItemPriceObservationRepository) Insert(ctx context.Context, obs itempriceentity.Observation) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO item_price_observations (id, source_item_id, ingest_run_id, amount_cents, currency, fulfillment_mode, delivery_executor, observed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		obs.ID, obs.SourceItemID, nullString(obs.IngestRunID), obs.Price.AmountCents, obs.Price.Currency, obs.FulfillmentMode, obs.DeliveryExecutor, obs.ObservedAt)
	return err
}

type QuoteObservationRepository struct {
	db *DB
}

func NewQuoteObservationRepository(db *DB) *QuoteObservationRepository {
	return &QuoteObservationRepository{db: db}
}

func (r *QuoteObservationRepository) GetByID(ctx context.Context, id string) (quoteentity.Observation, error) {
	obs, err := r.loadQuote(ctx, id)
	if err != nil {
		return quoteentity.Observation{}, err
	}
	obs.FeeLines, err = r.loadFeeLines(ctx, id)
	return obs, err
}

func (r *QuoteObservationRepository) Latest(ctx context.Context, sourceStoreID, geohash, mode, deliveryExecutor, tier string, basketSubtotalCents int64) (quoteentity.Observation, error) {
	bucket := quoteentity.BasketSubtotalBucket(basketSubtotalCents)
	var id string
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id FROM quote_observations
		WHERE source_store_id = $1 AND dropoff_geohash = $2 AND fulfillment_mode = $3
			AND delivery_executor = $4 AND membership_tier = $5
			AND basket_subtotal_cents = $6
		ORDER BY observed_at DESC LIMIT 1`,
		sourceStoreID, geohash, mode, deliveryExecutor, tier, bucket).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return quoteentity.Observation{}, quoteentity.ErrNotFound
	}
	if err != nil {
		return quoteentity.Observation{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *QuoteObservationRepository) Insert(ctx context.Context, obs quoteentity.Observation) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO quote_observations (
			id, source_store_id, channel_id, fulfillment_mode, delivery_executor, dropoff_geohash,
			membership_tier, quote_kind, basket_subtotal_cents, observed_at, ingest_run_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		obs.ID, obs.SourceStoreID, obs.ChannelID, obs.FulfillmentMode, obs.DeliveryExecutor, obs.DropoffGeohash,
		obs.MembershipTier, obs.QuoteKind, obs.BasketSubtotalCents, obs.ObservedAt, nullString(obs.IngestRunID))
	if err != nil {
		return err
	}
	for _, line := range obs.FeeLines {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO quote_fee_lines (id, quote_obs_id, kind, amount_cents, currency, percent, threshold_cents)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			line.ID, obs.ID, line.Kind, line.Amount.AmountCents, line.Amount.Currency, line.Percent, line.ThresholdCents)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *QuoteObservationRepository) loadQuote(ctx context.Context, id string) (quoteentity.Observation, error) {
	var out quoteentity.Observation
	var ingestID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_store_id, channel_id, fulfillment_mode, delivery_executor, dropoff_geohash,
			membership_tier, quote_kind, basket_subtotal_cents, observed_at, ingest_run_id
		FROM quote_observations WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceStoreID, &out.ChannelID, &out.FulfillmentMode, &out.DeliveryExecutor, &out.DropoffGeohash,
			&out.MembershipTier, &out.QuoteKind, &out.BasketSubtotalCents, &out.ObservedAt, &ingestID)
	if errors.Is(err, sql.ErrNoRows) {
		return quoteentity.Observation{}, quoteentity.ErrNotFound
	}
	if err != nil {
		return quoteentity.Observation{}, err
	}
	if ingestID.Valid {
		out.IngestRunID = ingestID.String
	}
	return out, nil
}

func (r *QuoteObservationRepository) loadFeeLines(ctx context.Context, quoteID string) ([]quoteentity.FeeLine, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, quote_obs_id, kind, amount_cents, currency, percent, threshold_cents
		FROM quote_fee_lines WHERE quote_obs_id = $1`, quoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []quoteentity.FeeLine
	for rows.Next() {
		var l quoteentity.FeeLine
		if err := rows.Scan(&l.ID, &l.QuoteObsID, &l.Kind, &l.Amount.AmountCents, &l.Amount.Currency, &l.Percent, &l.ThresholdCents); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}
