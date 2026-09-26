package entity

import "errors"

var (
	ErrIDRequired = errors.New("ingest.id is required")
	ErrNotFound   = errors.New("ingest record not found")
)
