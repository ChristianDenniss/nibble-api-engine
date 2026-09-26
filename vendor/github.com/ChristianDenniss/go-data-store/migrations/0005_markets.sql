-- Nibble operating geography: markets, probe dropoffs, channel-in-market coverage.
-- Store-level delivery polygons stay on service_areas.

CREATE TABLE IF NOT EXISTS markets (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    country TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT '',
    currency TEXT NOT NULL,
    timezone TEXT NOT NULL,
    status TEXT NOT NULL,
    geohash_prefixes TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS market_probe_dropoffs (
    id TEXT PRIMARY KEY,
    market_id TEXT NOT NULL REFERENCES markets (id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    address TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    postal_code TEXT NOT NULL DEFAULT '',
    geohash TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS channel_market_coverage (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    market_id TEXT NOT NULL REFERENCES markets (id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    store_count INTEGER NOT NULL DEFAULT 0,
    ingest_run_id TEXT REFERENCES ingest_runs (id) ON DELETE SET NULL,
    note TEXT NOT NULL DEFAULT '',
    last_observed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel_id, market_id)
);

CREATE INDEX IF NOT EXISTS idx_probe_dropoffs_market ON market_probe_dropoffs (market_id);
CREATE INDEX IF NOT EXISTS idx_channel_market_coverage_market ON channel_market_coverage (market_id);
