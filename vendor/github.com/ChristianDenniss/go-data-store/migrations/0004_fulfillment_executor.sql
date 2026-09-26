-- Split legacy combined fulfillment strings (e.g. delivery_3p) into fulfillment_mode + delivery_executor.

ALTER TABLE place_purchase_options
    ADD COLUMN IF NOT EXISTS delivery_executor TEXT NOT NULL DEFAULT '';

ALTER TABLE source_menus
    ADD COLUMN IF NOT EXISTS delivery_executor TEXT NOT NULL DEFAULT '';

ALTER TABLE service_areas
    ADD COLUMN IF NOT EXISTS delivery_executor TEXT NOT NULL DEFAULT '';

ALTER TABLE item_price_observations
    ADD COLUMN IF NOT EXISTS delivery_executor TEXT NOT NULL DEFAULT '';

ALTER TABLE quote_observations
    ADD COLUMN IF NOT EXISTS delivery_executor TEXT NOT NULL DEFAULT '';

UPDATE place_purchase_options
SET fulfillment_mode = 'delivery', delivery_executor = 'third_party'
WHERE fulfillment_mode = 'delivery_3p';

UPDATE place_purchase_options
SET fulfillment_mode = 'delivery', delivery_executor = 'merchant'
WHERE fulfillment_mode IN ('delivery_merchant', 'merchant_delivery');

UPDATE place_purchase_options
SET fulfillment_mode = 'in_store', delivery_executor = ''
WHERE fulfillment_mode = 'in_person';

UPDATE source_menus
SET fulfillment_mode = 'delivery', delivery_executor = 'third_party'
WHERE fulfillment_mode = 'delivery_3p';

UPDATE source_menus
SET fulfillment_mode = 'delivery', delivery_executor = 'merchant'
WHERE fulfillment_mode IN ('delivery_merchant', 'merchant_delivery');

UPDATE source_menus
SET fulfillment_mode = 'in_store', delivery_executor = ''
WHERE fulfillment_mode = 'in_person';

UPDATE service_areas
SET fulfillment_mode = 'delivery', delivery_executor = 'third_party'
WHERE fulfillment_mode = 'delivery_3p';

UPDATE service_areas
SET fulfillment_mode = 'delivery', delivery_executor = 'merchant'
WHERE fulfillment_mode IN ('delivery_merchant', 'merchant_delivery');

UPDATE item_price_observations
SET fulfillment_mode = 'delivery', delivery_executor = 'third_party'
WHERE fulfillment_mode = 'delivery_3p';

UPDATE item_price_observations
SET fulfillment_mode = 'delivery', delivery_executor = 'merchant'
WHERE fulfillment_mode IN ('delivery_merchant', 'merchant_delivery');

UPDATE quote_observations
SET fulfillment_mode = 'delivery', delivery_executor = 'third_party'
WHERE fulfillment_mode = 'delivery_3p';

UPDATE quote_observations
SET fulfillment_mode = 'delivery', delivery_executor = 'merchant'
WHERE fulfillment_mode IN ('delivery_merchant', 'merchant_delivery');

ALTER TABLE place_purchase_options
    DROP CONSTRAINT IF EXISTS place_purchase_options_place_id_channel_id_fulfillment_mode_key;

ALTER TABLE source_menus
    DROP CONSTRAINT IF EXISTS source_menus_source_store_id_fulfillment_mode_key;

ALTER TABLE place_purchase_options
    ADD CONSTRAINT place_purchase_options_path_key
    UNIQUE (place_id, channel_id, fulfillment_mode, delivery_executor);

ALTER TABLE source_menus
    ADD CONSTRAINT source_menus_store_fulfillment_key
    UNIQUE (source_store_id, fulfillment_mode, delivery_executor);

CREATE INDEX IF NOT EXISTS idx_item_price_obs_lookup
    ON item_price_observations (source_item_id, fulfillment_mode, delivery_executor, observed_at DESC);

DROP INDEX IF EXISTS idx_quote_obs_store_geohash;
CREATE INDEX IF NOT EXISTS idx_quote_obs_store_geohash_path
    ON quote_observations (source_store_id, dropoff_geohash, fulfillment_mode, delivery_executor, observed_at DESC);
