package entity

import "errors"

var (
	ErrIDRequired = errors.New("resolution.id is required")
	ErrNotFound   = errors.New("match not found")
)
