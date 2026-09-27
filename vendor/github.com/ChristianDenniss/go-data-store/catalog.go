package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ChristianDenniss/go-data-model/catalog"
	"time"
)

type CatalogRepository struct{ db *DB }

func NewCatalogRepository(db *DB) *CatalogRepository { return &CatalogRepository{db} }

func (r *CatalogRepository) Import(ctx context.Context, b catalog.Bundle, raw []byte) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize import batches; readers see either the previous or the new catalog.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(7349921)`); err != nil {
		return err
	}
	checksum := fmt.Sprintf("%x", sha256.Sum256(raw))
	res, err := tx.ExecContext(ctx, `INSERT INTO catalog_imports(checksum,raw_json) VALUES($1,$2) ON CONFLICT DO NOTHING`, checksum, raw)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return tx.Commit()
	}
	for provider, snapshot := range b.Providers {
		stores := snapshot.Stores
		snapshot.Stores = nil
		meta, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO catalog_providers(provider,metadata,checksum) VALUES($1,$2,$3) ON CONFLICT(provider) DO UPDATE SET metadata=excluded.metadata,checksum=excluded.checksum`, provider, meta, checksum); err != nil {
			return err
		}
		// A restarted worker with older archives cannot roll back observed prices.
		for _, incoming := range stores {
			var previous []byte
			lookupErr := tx.QueryRowContext(ctx, `SELECT metadata FROM catalog_stores WHERE id=$1`, incoming.ID).Scan(&previous)
			if lookupErr != nil && lookupErr != sql.ErrNoRows {
				return lookupErr
			}
			if lookupErr == nil {
				var old catalog.Store
				if err := json.Unmarshal(previous, &old); err != nil {
					return err
				}
				oldTime, _ := time.Parse(time.RFC3339Nano, old.ObservedAt)
				nextTime, _ := time.Parse(time.RFC3339Nano, incoming.ObservedAt)
				if oldTime.After(nextTime) {
					return fmt.Errorf("older menu rejected for %s", incoming.ID)
				}
			}
		}
		// Each provider envelope is a complete retained directory, not a delta.
		if _, err = tx.ExecContext(ctx, `DELETE FROM catalog_stores WHERE provider=$1`, provider); err != nil {
			return err
		}
		for order, st := range stores {
			items := st.Items
			st.Items = nil
			meta, err = json.Marshal(st)
			if err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO catalog_stores(id,provider,name,address,url,metadata,sort_order) VALUES($1,$2,$3,$4,$5,$6,$7)`, st.ID, provider, st.Name, st.Address, st.URL, meta, order); err != nil {
				return err
			}
			for ordinal, item := range items {
				meta, err = json.Marshal(item)
				if err != nil {
					return err
				}
				if _, err = tx.ExecContext(ctx, `INSERT INTO catalog_items(store_id,ordinal,name,amount_cents,currency,metadata) VALUES($1,$2,$3,$4,$5,$6)`, st.ID, ordinal, item.Name, item.Amount, item.Currency, meta); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}
func (r *CatalogRepository) Load(ctx context.Context) (catalog.Bundle, error) {
	b := catalog.Bundle{Version: 1, Providers: map[string]catalog.Snapshot{}}
	tx, err := r.db.sql.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return b, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT provider,metadata FROM catalog_providers ORDER BY provider`)
	if err != nil {
		return b, err
	}
	for rows.Next() {
		var p string
		var raw []byte
		var snap catalog.Snapshot
		if err = rows.Scan(&p, &raw); err != nil {
			rows.Close()
			return b, err
		}
		if err = json.Unmarshal(raw, &snap); err != nil {
			rows.Close()
			return b, err
		}
		snap.Stores = []catalog.Store{}
		b.Providers[p] = snap
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return b, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT s.provider,s.metadata,COALESCE(jsonb_agg(i.metadata ORDER BY i.ordinal) FILTER(WHERE i.store_id IS NOT NULL),'[]') FROM catalog_stores s LEFT JOIN catalog_items i ON i.store_id=s.id GROUP BY s.id ORDER BY s.provider,s.sort_order`)
	if err != nil {
		return b, err
	}
	for rows.Next() {
		var p string
		var raw, items []byte
		var st catalog.Store
		if err = rows.Scan(&p, &raw, &items); err != nil {
			rows.Close()
			return b, err
		}
		if err = json.Unmarshal(raw, &st); err != nil {
			rows.Close()
			return b, err
		}
		if err = json.Unmarshal(items, &st.Items); err != nil {
			rows.Close()
			return b, err
		}
		snap := b.Providers[p]
		snap.Stores = append(snap.Stores, st)
		b.Providers[p] = snap
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return b, err
	}
	return b, tx.Commit()
}
