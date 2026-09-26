-- Target domain (catalog, pricing, resolution, user, compare). Legacy 0001/0002 unchanged.

CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ingest_runs (
    id TEXT PRIMARY KEY,
    job_type TEXT NOT NULL,
    channel_id TEXT REFERENCES channels (id),
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    parser_version TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_stores (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    external_store_id TEXT NOT NULL,
    name TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    address TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel_id, external_store_id)
);

CREATE TABLE IF NOT EXISTS source_snapshots (
    id TEXT PRIMARY KEY,
    ingest_run_id TEXT NOT NULL REFERENCES ingest_runs (id) ON DELETE CASCADE,
    provider_id TEXT NOT NULL DEFAULT '',
    external_store_id TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT 'application/json',
    raw_json JSONB NOT NULL,
    checksum TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_menus (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    fulfillment_mode TEXT NOT NULL,
    external_menu_id TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_store_id, fulfillment_mode)
);

CREATE TABLE IF NOT EXISTS source_categories (
    id TEXT PRIMARY KEY,
    source_menu_id TEXT NOT NULL REFERENCES source_menus (id) ON DELETE CASCADE,
    external_category_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_items (
    id TEXT PRIMARY KEY,
    source_category_id TEXT NOT NULL REFERENCES source_categories (id) ON DELETE CASCADE,
    external_item_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    available BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_modifier_groups (
    id TEXT PRIMARY KEY,
    external_id TEXT NOT NULL DEFAULT '',
    min_select INTEGER NOT NULL DEFAULT 0,
    max_select INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_modifier_options (
    id TEXT PRIMARY KEY,
    source_mod_group_id TEXT NOT NULL REFERENCES source_modifier_groups (id) ON DELETE CASCADE,
    external_option_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    price_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_item_modifier_groups (
    source_item_id TEXT NOT NULL REFERENCES source_items (id) ON DELETE CASCADE,
    source_mod_group_id TEXT NOT NULL REFERENCES source_modifier_groups (id) ON DELETE CASCADE,
    PRIMARY KEY (source_item_id, source_mod_group_id)
);

CREATE TABLE IF NOT EXISTS brands (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS places (
    id TEXT PRIMARY KEY,
    brand_id TEXT REFERENCES brands (id),
    name TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    address TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    postal_code TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dishes (
    id TEXT PRIMARY KEY,
    brand_id TEXT REFERENCES brands (id),
    name TEXT NOT NULL,
    canonical_name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS brand_aliases (
    id TEXT PRIMARY KEY,
    brand_id TEXT NOT NULL REFERENCES brands (id) ON DELETE CASCADE,
    alias TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dish_aliases (
    id TEXT PRIMARY KEY,
    dish_id TEXT NOT NULL REFERENCES dishes (id) ON DELETE CASCADE,
    alias TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS brand_cuisines (
    brand_id TEXT NOT NULL REFERENCES brands (id) ON DELETE CASCADE,
    cuisine_id TEXT NOT NULL REFERENCES cuisines (id) ON DELETE CASCADE,
    PRIMARY KEY (brand_id, cuisine_id)
);

CREATE TABLE IF NOT EXISTS brand_categories (
    brand_id TEXT NOT NULL REFERENCES brands (id) ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (brand_id, category_id)
);

CREATE TABLE IF NOT EXISTS dish_categories (
    dish_id TEXT NOT NULL REFERENCES dishes (id) ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (dish_id, category_id)
);

CREATE TABLE IF NOT EXISTS dietary_tags (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dish_dietary_tags (
    dish_id TEXT NOT NULL REFERENCES dishes (id) ON DELETE CASCADE,
    dietary_tag_id TEXT NOT NULL REFERENCES dietary_tags (id) ON DELETE CASCADE,
    PRIMARY KEY (dish_id, dietary_tag_id)
);

CREATE TABLE IF NOT EXISTS place_purchase_options (
    id TEXT PRIMARY KEY,
    place_id TEXT NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    fulfillment_mode TEXT NOT NULL,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (place_id, channel_id, fulfillment_mode)
);

CREATE TABLE IF NOT EXISTS store_matches (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    place_id TEXT NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    method TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_store_id, place_id)
);

CREATE TABLE IF NOT EXISTS item_matches (
    id TEXT PRIMARY KEY,
    source_item_id TEXT NOT NULL REFERENCES source_items (id) ON DELETE CASCADE,
    dish_id TEXT NOT NULL REFERENCES dishes (id) ON DELETE CASCADE,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    method TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_item_id, dish_id)
);

CREATE TABLE IF NOT EXISTS match_evidence (
    id TEXT PRIMARY KEY,
    store_match_id TEXT REFERENCES store_matches (id) ON DELETE CASCADE,
    item_match_id TEXT REFERENCES item_matches (id) ON DELETE CASCADE,
    signal_kind TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS service_areas (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    fulfillment_mode TEXT NOT NULL,
    geometry TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS hours_regular (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL,
    opens TIME NOT NULL,
    closes TIME NOT NULL
);

CREATE TABLE IF NOT EXISTS hours_exceptions (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    on_date DATE NOT NULL,
    closed BOOLEAN NOT NULL DEFAULT false,
    opens TIME,
    closes TIME
);

CREATE TABLE IF NOT EXISTS source_store_status (
    source_store_id TEXT PRIMARY KEY REFERENCES source_stores (id) ON DELETE CASCADE,
    open_now BOOLEAN NOT NULL DEFAULT false,
    paused BOOLEAN NOT NULL DEFAULT false,
    observed_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS item_price_observations (
    id TEXT PRIMARY KEY,
    source_item_id TEXT NOT NULL REFERENCES source_items (id) ON DELETE CASCADE,
    ingest_run_id TEXT REFERENCES ingest_runs (id),
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    fulfillment_mode TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quote_observations (
    id TEXT PRIMARY KEY,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    fulfillment_mode TEXT NOT NULL,
    dropoff_geohash TEXT NOT NULL DEFAULT '',
    membership_tier TEXT NOT NULL DEFAULT '',
    quote_kind TEXT NOT NULL,
    basket_subtotal_cents BIGINT NOT NULL DEFAULT 0,
    observed_at TIMESTAMPTZ NOT NULL,
    ingest_run_id TEXT REFERENCES ingest_runs (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quote_fee_lines (
    id TEXT PRIMARY KEY,
    quote_obs_id TEXT NOT NULL REFERENCES quote_observations (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    percent NUMERIC(8, 4) NOT NULL DEFAULT 0,
    threshold_cents BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS promotions (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    value_cents BIGINT NOT NULL DEFAULT 0,
    value_bps INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS promotion_constraints (
    id TEXT PRIMARY KEY,
    promotion_id TEXT NOT NULL REFERENCES promotions (id) ON DELETE CASCADE,
    min_subtotal_cents BIGINT NOT NULL DEFAULT 0,
    code TEXT NOT NULL DEFAULT '',
    membership_required BOOLEAN NOT NULL DEFAULT false,
    max_discount_cents BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS promotion_targets (
    id TEXT PRIMARY KEY,
    promotion_id TEXT NOT NULL REFERENCES promotions (id) ON DELETE CASCADE,
    place_id TEXT REFERENCES places (id) ON DELETE CASCADE,
    source_store_id TEXT REFERENCES source_stores (id) ON DELETE CASCADE,
    source_item_id TEXT REFERENCES source_items (id) ON DELETE CASCADE,
    dish_id TEXT REFERENCES dishes (id) ON DELETE CASCADE,
    brand_id TEXT REFERENCES brands (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS membership_products (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    UNIQUE (channel_id, slug)
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    preferred_currency TEXT NOT NULL,
    notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    compare_prefs JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS user_dropoffs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    address TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    postal_code TEXT NOT NULL DEFAULT '',
    current BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS user_memberships (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    membership_product_id TEXT NOT NULL REFERENCES membership_products (id) ON DELETE CASCADE,
    UNIQUE (user_id, membership_product_id)
);

CREATE TABLE IF NOT EXISTS user_dietary_prefs (
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (user_id, tag)
);

CREATE TABLE IF NOT EXISTS watchlists (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    place_id TEXT REFERENCES places (id) ON DELETE CASCADE,
    dish_id TEXT REFERENCES dishes (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS price_alerts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    threshold_cents BIGINT NOT NULL,
    currency TEXT NOT NULL,
    filter_snapshot JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS compare_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    place_id TEXT NOT NULL REFERENCES places (id),
    query_snapshot JSONB NOT NULL DEFAULT '{}',
    basket_snapshot JSONB NOT NULL DEFAULT '{}',
    result_snapshot JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbound_clicks (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    compare_session_id TEXT REFERENCES compare_sessions (id) ON DELETE SET NULL,
    purchase_option_id TEXT NOT NULL DEFAULT '',
    action_kind TEXT NOT NULL,
    target_url TEXT NOT NULL DEFAULT '',
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_item_price_obs_item_observed
    ON item_price_observations (source_item_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_quote_obs_store_geohash
    ON quote_observations (source_store_id, dropoff_geohash, fulfillment_mode, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_source_items_category ON source_items (source_category_id);
