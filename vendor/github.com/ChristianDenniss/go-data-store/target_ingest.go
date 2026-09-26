package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/ingest/entity"
	"github.com/ChristianDenniss/go-data-model/ingest/repository"
)

var (
	_ repository.RunRepository      = (*IngestRunRepository)(nil)
	_ repository.SnapshotRepository = (*SourceSnapshotRepository)(nil)
)

type IngestRunRepository struct {
	db *DB
}

func NewIngestRunRepository(db *DB) *IngestRunRepository {
	return &IngestRunRepository{db: db}
}

func (r *IngestRunRepository) GetByID(ctx context.Context, id string) (entity.IngestRun, error) {
	var out entity.IngestRun
	var finished sql.NullTime
	var channelID sql.NullString
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, job_type, channel_id, started_at, finished_at, parser_version
		FROM ingest_runs WHERE id = $1`, id).
		Scan(&out.ID, &out.JobType, &channelID, &out.StartedAt, &finished, &out.ParserVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.IngestRun{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.IngestRun{}, err
	}
	if channelID.Valid {
		out.ChannelID = channelID.String
	}
	if finished.Valid {
		t := finished.Time
		out.FinishedAt = &t
	}
	return out, nil
}

func (r *IngestRunRepository) Upsert(ctx context.Context, run entity.IngestRun) error {
	var finished interface{}
	if run.FinishedAt != nil {
		finished = *run.FinishedAt
	}
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO ingest_runs (id, job_type, channel_id, started_at, finished_at, parser_version)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			job_type = EXCLUDED.job_type,
			channel_id = EXCLUDED.channel_id,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at,
			parser_version = EXCLUDED.parser_version`,
		run.ID, run.JobType, nullString(run.ChannelID), run.StartedAt, finished, run.ParserVersion)
	return err
}

type SourceSnapshotRepository struct {
	db *DB
}

func NewSourceSnapshotRepository(db *DB) *SourceSnapshotRepository {
	return &SourceSnapshotRepository{db: db}
}

func (r *SourceSnapshotRepository) GetByID(ctx context.Context, id string) (entity.SourceSnapshot, error) {
	var out entity.SourceSnapshot
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, ingest_run_id, external_store_id, content_type, raw_json, checksum, observed_at
		FROM source_snapshots WHERE id = $1`, id).
		Scan(&out.ID, &out.IngestRunID, &out.ExternalStoreID, &out.ContentType, &out.RawJSON, &out.Checksum, &out.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.SourceSnapshot{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceSnapshotRepository) Insert(ctx context.Context, snap entity.SourceSnapshot) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_snapshots (id, ingest_run_id, external_store_id, content_type, raw_json, checksum, observed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		snap.ID, snap.IngestRunID, snap.ExternalStoreID, snap.ContentType, snap.RawJSON, snap.Checksum, snap.ObservedAt)
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
