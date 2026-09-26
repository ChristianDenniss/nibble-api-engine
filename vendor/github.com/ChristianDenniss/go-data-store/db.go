package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"time"

	"github.com/ChristianDenniss/go-data-store/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	sql *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	log.Printf("postgres: opening")
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for i := 0; i < 30; i++ {
		log.Printf("postgres: ping (attempt %d/30)", i+1)
		if err := sqlDB.PingContext(ctx); err == nil {
			log.Printf("postgres: reachable")
			db := &DB{sql: sqlDB}
			if err := db.migrate(ctx); err != nil {
				_ = sqlDB.Close()
				return nil, err
			}
			return db, nil
		} else {
			lastErr = err
			log.Printf("postgres: ping failed: %v", err)
			time.Sleep(time.Second)
		}
	}
	_ = sqlDB.Close()
	return nil, fmt.Errorf("postgres: %w", lastErr)
}

func (db *DB) Close() error {
	log.Printf("postgres: closing")
	return db.sql.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.sql.PingContext(ctx)
}

func (db *DB) migrate(ctx context.Context) error {
	if _, err := db.sql.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("migrations: create ledger: %w", err)
	}

	files, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("migrations: no sql files embedded")
	}
	log.Printf("postgres: applying %d migration file(s)", len(files))
	for _, name := range files {
		var applied string
		err := db.sql.QueryRowContext(ctx, `SELECT filename FROM schema_migrations WHERE filename = $1`, name).Scan(&applied)
		if err == nil {
			log.Printf("postgres: migration %s already applied", name)
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("migrations: lookup %s: %w", name, err)
		}

		log.Printf("postgres: migration %s starting", name)
		sqlBytes, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return err
		}
		if _, err := db.sql.ExecContext(ctx, string(sqlBytes)); err != nil {
			log.Printf("postgres: migration %s failed: %v", name, err)
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := db.sql.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("migrations: record %s: %w", name, err)
		}
		log.Printf("postgres: migration %s ok", name)
	}
	log.Printf("postgres: migrations complete")
	return nil
}
