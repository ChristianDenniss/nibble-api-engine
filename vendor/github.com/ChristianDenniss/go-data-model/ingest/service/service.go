package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/ingest/entity"
	"github.com/ChristianDenniss/go-data-model/ingest/repository"
)

type Service struct {
	runs      repository.RunRepository
	snapshots repository.SnapshotRepository
}

func New(runs repository.RunRepository, snapshots repository.SnapshotRepository) *Service {
	return &Service{runs: runs, snapshots: snapshots}
}

func (s *Service) StartRun(ctx context.Context, run entity.IngestRun) error {
	if run.ID == "" {
		return entity.ErrIDRequired
	}
	return s.runs.Upsert(ctx, run)
}

func (s *Service) FinishRun(ctx context.Context, run entity.IngestRun) error {
	if run.ID == "" {
		return entity.ErrIDRequired
	}
	return s.runs.Upsert(ctx, run)
}

func (s *Service) RecordSnapshot(ctx context.Context, snap entity.SourceSnapshot) error {
	if snap.ID == "" {
		return entity.ErrIDRequired
	}
	return s.snapshots.Insert(ctx, snap)
}
