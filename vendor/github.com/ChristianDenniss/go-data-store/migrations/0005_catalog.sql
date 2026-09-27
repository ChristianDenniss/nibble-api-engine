-- Durable browse read model. Raw immutable imports retain provenance for audit;
-- current restaurants/items are separately queryable and replaced atomically.
CREATE TABLE IF NOT EXISTS catalog_imports (
 checksum text PRIMARY KEY,
 raw_json jsonb NOT NULL,
 imported_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS catalog_providers (
 provider text PRIMARY KEY,
 metadata jsonb NOT NULL,
 checksum text NOT NULL REFERENCES catalog_imports(checksum)
);
CREATE TABLE IF NOT EXISTS catalog_stores (
 id text PRIMARY KEY,
 provider text NOT NULL REFERENCES catalog_providers(provider),
 name text NOT NULL,
 address text NOT NULL,
 url text NOT NULL,
 metadata jsonb NOT NULL,
 sort_order integer NOT NULL
);
CREATE INDEX IF NOT EXISTS catalog_stores_provider ON catalog_stores(provider);
CREATE TABLE IF NOT EXISTS catalog_items (
 store_id text NOT NULL REFERENCES catalog_stores(id) ON DELETE CASCADE,
 ordinal integer NOT NULL,
 name text NOT NULL,
 amount_cents bigint CHECK (amount_cents >= 0),
 currency text NOT NULL CHECK (currency = 'CAD'),
 metadata jsonb NOT NULL,
 PRIMARY KEY(store_id, ordinal)
);
