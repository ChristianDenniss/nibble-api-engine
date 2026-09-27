package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/market/entity"
	"github.com/ChristianDenniss/go-data-model/market/repository"
)

var (
	_ repository.MarketRepository       = (*MarketRepository)(nil)
	_ repository.ProbeDropoffRepository = (*ProbeDropoffRepository)(nil)
	_ repository.CoverageRepository     = (*ChannelCoverageRepository)(nil)
)

type MarketRepository struct {
	db *DB
}

func NewMarketRepository(db *DB) *MarketRepository {
	return &MarketRepository{db: db}
}

func scanMarket(row interface {
	Scan(dest ...any) error
}) (entity.Market, error) {
	var out entity.Market
	err := row.Scan(&out.ID, &out.Slug, &out.Name, &out.Country, &out.Region, &out.Currency, &out.Timezone, &out.Status, &out.GeohashPrefixes)
	return out, err
}

func (r *MarketRepository) GetByID(ctx context.Context, id string) (entity.Market, error) {
	out, err := scanMarket(r.db.sql.QueryRowContext(ctx, `
		SELECT id, slug, name, country, region, currency, timezone, status, geohash_prefixes
		FROM markets WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Market{}, entity.ErrNotFound
	}
	return out, err
}

func (r *MarketRepository) GetBySlug(ctx context.Context, slug string) (entity.Market, error) {
	out, err := scanMarket(r.db.sql.QueryRowContext(ctx, `
		SELECT id, slug, name, country, region, currency, timezone, status, geohash_prefixes
		FROM markets WHERE slug = $1`, slug))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Market{}, entity.ErrNotFound
	}
	return out, err
}

func (r *MarketRepository) List(ctx context.Context) ([]entity.Market, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, slug, name, country, region, currency, timezone, status, geohash_prefixes
		FROM markets ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Market
	for rows.Next() {
		m, err := scanMarket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MarketRepository) Upsert(ctx context.Context, m entity.Market) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO markets (id, slug, name, country, region, currency, timezone, status, geohash_prefixes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug, name = EXCLUDED.name, country = EXCLUDED.country,
			region = EXCLUDED.region, currency = EXCLUDED.currency, timezone = EXCLUDED.timezone,
			status = EXCLUDED.status, geohash_prefixes = EXCLUDED.geohash_prefixes, updated_at = now()`,
		m.ID, m.Slug, m.Name, m.Country, m.Region, m.Currency, m.Timezone, m.Status, m.GeohashPrefixes)
	return err
}

type ProbeDropoffRepository struct {
	db *DB
}

func NewProbeDropoffRepository(db *DB) *ProbeDropoffRepository {
	return &ProbeDropoffRepository{db: db}
}

func (r *ProbeDropoffRepository) ListByMarket(ctx context.Context, marketID string) ([]entity.ProbeDropoff, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, market_id, label, latitude, longitude, address, city, region, postal_code, geohash
		FROM market_probe_dropoffs WHERE market_id = $1 ORDER BY label`, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.ProbeDropoff
	for rows.Next() {
		var d entity.ProbeDropoff
		if err := rows.Scan(
			&d.ID, &d.MarketID, &d.Label,
			&d.Location.Latitude, &d.Location.Longitude, &d.Location.Address,
			&d.Location.City, &d.Location.Region, &d.Location.PostalCode, &d.Geohash,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *ProbeDropoffRepository) Upsert(ctx context.Context, d entity.ProbeDropoff) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO market_probe_dropoffs (
			id, market_id, label, latitude, longitude, address, city, region, postal_code, geohash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			market_id = EXCLUDED.market_id, label = EXCLUDED.label,
			latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
			address = EXCLUDED.address, city = EXCLUDED.city, region = EXCLUDED.region,
			postal_code = EXCLUDED.postal_code, geohash = EXCLUDED.geohash, updated_at = now()`,
		d.ID, d.MarketID, d.Label,
		d.Location.Latitude, d.Location.Longitude, d.Location.Address,
		d.Location.City, d.Location.Region, d.Location.PostalCode, d.Geohash)
	return err
}

type ChannelCoverageRepository struct {
	db *DB
}

func NewChannelCoverageRepository(db *DB) *ChannelCoverageRepository {
	return &ChannelCoverageRepository{db: db}
}

func scanCoverage(row interface {
	Scan(dest ...any) error
}) (entity.ChannelCoverage, error) {
	var out entity.ChannelCoverage
	var ingest sql.NullString
	var observed sql.NullTime
	err := row.Scan(&out.ID, &out.ChannelID, &out.MarketID, &out.Status, &out.StoreCount, &ingest, &out.Note, &observed)
	if err != nil {
		return entity.ChannelCoverage{}, err
	}
	if ingest.Valid {
		out.IngestRunID = ingest.String
	}
	if observed.Valid {
		t := observed.Time
		out.LastObservedAt = &t
	}
	return out, nil
}

func (r *ChannelCoverageRepository) Get(ctx context.Context, channelID, marketID string) (entity.ChannelCoverage, error) {
	out, err := scanCoverage(r.db.sql.QueryRowContext(ctx, `
		SELECT id, channel_id, market_id, status, store_count, ingest_run_id, note, last_observed_at
		FROM channel_market_coverage WHERE channel_id = $1 AND market_id = $2`, channelID, marketID))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ChannelCoverage{}, entity.ErrNotFound
	}
	return out, err
}

func (r *ChannelCoverageRepository) ListByMarket(ctx context.Context, marketID string) ([]entity.ChannelCoverage, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, channel_id, market_id, status, store_count, ingest_run_id, note, last_observed_at
		FROM channel_market_coverage WHERE market_id = $1 ORDER BY channel_id`, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.ChannelCoverage
	for rows.Next() {
		c, err := scanCoverage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ChannelCoverageRepository) Upsert(ctx context.Context, c entity.ChannelCoverage) error {
	var ingest any
	if c.IngestRunID != "" {
		ingest = c.IngestRunID
	}
	var observed any
	if c.LastObservedAt != nil {
		observed = c.LastObservedAt.UTC()
	}
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO channel_market_coverage (
			id, channel_id, market_id, status, store_count, ingest_run_id, note, last_observed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (channel_id, market_id) DO UPDATE SET
			id = EXCLUDED.id, status = EXCLUDED.status, store_count = EXCLUDED.store_count,
			ingest_run_id = EXCLUDED.ingest_run_id, note = EXCLUDED.note,
			last_observed_at = EXCLUDED.last_observed_at, updated_at = now()`,
		c.ID, c.ChannelID, c.MarketID, c.Status, c.StoreCount, ingest, c.Note, observed)
	return err
}
