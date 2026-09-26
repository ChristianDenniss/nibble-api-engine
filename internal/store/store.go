package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/ChristianDenniss/api-engine/migrations"
	model "github.com/ChristianDenniss/go-data-model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for i := 0; i < 30; i++ {
		if err := db.PingContext(ctx); err == nil {
			s := &Store{db: db}
			if err := s.migrate(ctx); err != nil {
				_ = db.Close()
				return nil, err
			}
			return s, nil
		} else {
			lastErr = err
			time.Sleep(time.Second)
		}
	}
	_ = db.Close()
	return nil, fmt.Errorf("postgres: %w", lastErr)
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	files, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("migrations: no sql files embedded")
	}
	for _, name := range files {
		sqlBytes, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func (s *Store) UpsertRestaurant(ctx context.Context, r model.Restaurant) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO restaurants (id, name, latitude, longitude, address)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			address = EXCLUDED.address,
			updated_at = now()`,
		r.ID, r.Name, r.Location.Latitude, r.Location.Longitude, r.Location.Address)
	return err
}

func (s *Store) UpsertMenuItem(ctx context.Context, item model.MenuItem) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO menu_items (id, restaurant_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			restaurant_id = EXCLUDED.restaurant_id,
			name = EXCLUDED.name,
			updated_at = now()`,
		item.ID, item.RestaurantID, item.Name)
	return err
}

func (s *Store) UpsertOffer(ctx context.Context, offer model.Offer) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO offers (id, restaurant_id, provider_id, menu_item_id, amount_cents, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			restaurant_id = EXCLUDED.restaurant_id,
			provider_id = EXCLUDED.provider_id,
			menu_item_id = EXCLUDED.menu_item_id,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			updated_at = now()`,
		offer.ID, offer.RestaurantID, offer.ProviderID, offer.MenuItemID, offer.Price.AmountCents, offer.Price.Currency)
	return err
}

func (s *Store) UpsertPriceObservation(ctx context.Context, obs model.PriceObservation) error {
	_, err := s.db.ExecContext(ctx, `
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
