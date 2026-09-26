package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/ingest/entity"
)

type RunRepository interface {
	GetByID(ctx context.Context, id string) (entity.IngestRun, error)
	Upsert(ctx context.Context, run entity.IngestRun) error
}

type SnapshotRepository interface {
	GetByID(ctx context.Context, id string) (entity.SourceSnapshot, error)
	Insert(ctx context.Context, snap entity.SourceSnapshot) error
}
