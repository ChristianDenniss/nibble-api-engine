package entity

import "errors"

var (
	ErrIDRequired = errors.New("source.id is required")
	ErrNotFound   = errors.New("source record not found")
)
