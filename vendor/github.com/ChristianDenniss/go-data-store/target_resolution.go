package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/resolution/entity"
	"github.com/ChristianDenniss/go-data-model/resolution/repository"
)

var (
	_ repository.StoreMatchRepository = (*StoreMatchRepository)(nil)
	_ repository.ItemMatchRepository  = (*ItemMatchRepository)(nil)
	_ repository.EvidenceRepository   = (*MatchEvidenceRepository)(nil)
)

type StoreMatchRepository struct {
	db *DB
}

func NewStoreMatchRepository(db *DB) *StoreMatchRepository {
	return &StoreMatchRepository{db: db}
}

func (r *StoreMatchRepository) GetByID(ctx context.Context, id string) (entity.StoreMatch, error) {
	var out entity.StoreMatch
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_store_id, place_id, confidence, status, method FROM store_matches WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceStoreID, &out.PlaceID, &out.Confidence, &out.Status, &out.Method)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.StoreMatch{}, entity.ErrNotFound
	}
	return out, err
}

func (r *StoreMatchRepository) ListByPlace(ctx context.Context, placeID string) ([]entity.StoreMatch, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_store_id, place_id, confidence, status, method FROM store_matches WHERE place_id = $1`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.StoreMatch
	for rows.Next() {
		var m entity.StoreMatch
		if err := rows.Scan(&m.ID, &m.SourceStoreID, &m.PlaceID, &m.Confidence, &m.Status, &m.Method); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *StoreMatchRepository) Upsert(ctx context.Context, m entity.StoreMatch) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO store_matches (id, source_store_id, place_id, confidence, status, method)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			source_store_id = EXCLUDED.source_store_id, place_id = EXCLUDED.place_id,
			confidence = EXCLUDED.confidence, status = EXCLUDED.status, method = EXCLUDED.method,
			updated_at = now()`,
		m.ID, m.SourceStoreID, m.PlaceID, m.Confidence, m.Status, m.Method)
	return err
}

type ItemMatchRepository struct {
	db *DB
}

func NewItemMatchRepository(db *DB) *ItemMatchRepository {
	return &ItemMatchRepository{db: db}
}

func (r *ItemMatchRepository) GetByID(ctx context.Context, id string) (entity.ItemMatch, error) {
	var out entity.ItemMatch
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, source_item_id, dish_id, confidence, status, method FROM item_matches WHERE id = $1`, id).
		Scan(&out.ID, &out.SourceItemID, &out.DishID, &out.Confidence, &out.Status, &out.Method)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ItemMatch{}, entity.ErrNotFound
	}
	return out, err
}

func (r *ItemMatchRepository) FindSourceItemForStoreAndDish(ctx context.Context, sourceStoreID, dishID string) (string, error) {
	var sourceItemID string
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT im.source_item_id
		FROM item_matches im
		JOIN source_items si ON si.id = im.source_item_id
		JOIN source_categories sc ON sc.id = si.source_category_id
		JOIN source_menus sm ON sm.id = sc.source_menu_id
		WHERE sm.source_store_id = $1 AND im.dish_id = $2
		ORDER BY im.confidence DESC NULLS LAST
		LIMIT 1`, sourceStoreID, dishID).Scan(&sourceItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", entity.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return sourceItemID, nil
}

func (r *ItemMatchRepository) ListByDish(ctx context.Context, dishID string) ([]entity.ItemMatch, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, source_item_id, dish_id, confidence, status, method FROM item_matches WHERE dish_id = $1`, dishID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.ItemMatch
	for rows.Next() {
		var m entity.ItemMatch
		if err := rows.Scan(&m.ID, &m.SourceItemID, &m.DishID, &m.Confidence, &m.Status, &m.Method); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ItemMatchRepository) Upsert(ctx context.Context, m entity.ItemMatch) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO item_matches (id, source_item_id, dish_id, confidence, status, method)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			source_item_id = EXCLUDED.source_item_id, dish_id = EXCLUDED.dish_id,
			confidence = EXCLUDED.confidence, status = EXCLUDED.status, method = EXCLUDED.method,
			updated_at = now()`,
		m.ID, m.SourceItemID, m.DishID, m.Confidence, m.Status, m.Method)
	return err
}

type MatchEvidenceRepository struct {
	db *DB
}

func NewMatchEvidenceRepository(db *DB) *MatchEvidenceRepository {
	return &MatchEvidenceRepository{db: db}
}

func (r *MatchEvidenceRepository) Insert(ctx context.Context, e entity.Evidence) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO match_evidence (id, store_match_id, item_match_id, signal_kind, score, note)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, nullString(e.StoreMatchID), nullString(e.ItemMatchID), e.SignalKind, e.Score, e.Note)
	return err
}
