package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/serviceability/entity"
	"github.com/ChristianDenniss/go-data-model/serviceability/repository"
)

var (
	_ repository.StoreStatusRepository = (*SourceStoreStatusRepository)(nil)
	_ repository.ServiceAreaRepository = (*ServiceAreaRepository)(nil)
)

type SourceStoreStatusRepository struct {
	db *DB
}

func NewSourceStoreStatusRepository(db *DB) *SourceStoreStatusRepository {
	return &SourceStoreStatusRepository{db: db}
}

func (r *SourceStoreStatusRepository) Get(ctx context.Context, sourceStoreID string) (entity.StoreStatus, error) {
	var out entity.StoreStatus
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT source_store_id, open_now, paused, observed_at FROM source_store_status WHERE source_store_id = $1`, sourceStoreID).
		Scan(&out.SourceStoreID, &out.OpenNow, &out.Paused, &out.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.StoreStatus{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceStoreStatusRepository) Upsert(ctx context.Context, st entity.StoreStatus) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_store_status (source_store_id, open_now, paused, observed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_store_id) DO UPDATE SET
			open_now = EXCLUDED.open_now, paused = EXCLUDED.paused, observed_at = EXCLUDED.observed_at`,
		st.SourceStoreID, st.OpenNow, st.Paused, st.ObservedAt)
	return err
}

type ServiceAreaRepository struct {
	db *DB
}

func NewServiceAreaRepository(db *DB) *ServiceAreaRepository {
	return &ServiceAreaRepository{db: db}
}

func (r *ServiceAreaRepository) ListForPath(ctx context.Context, sourceStoreID, fulfillmentMode, deliveryExecutor string) ([]entity.ServiceArea, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_store_id, fulfillment_mode, delivery_executor, geometry
		FROM service_areas
		WHERE source_store_id = $1 AND fulfillment_mode = $2 AND delivery_executor = $3`,
		sourceStoreID, fulfillmentMode, deliveryExecutor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.ServiceArea
	for rows.Next() {
		var a entity.ServiceArea
		if err := rows.Scan(&a.ID, &a.SourceStoreID, &a.FulfillmentMode, &a.DeliveryExecutor, &a.Geometry); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *ServiceAreaRepository) Upsert(ctx context.Context, area entity.ServiceArea) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO service_areas (id, source_store_id, fulfillment_mode, delivery_executor, geometry)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			source_store_id = EXCLUDED.source_store_id,
			fulfillment_mode = EXCLUDED.fulfillment_mode,
			delivery_executor = EXCLUDED.delivery_executor,
			geometry = EXCLUDED.geometry,
			updated_at = now()`,
		area.ID, area.SourceStoreID, area.FulfillmentMode, area.DeliveryExecutor, area.Geometry)
	return err
}
