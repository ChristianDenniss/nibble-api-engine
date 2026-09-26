package entity

import "errors"

var (
	ErrIDRequired = errors.New("provider.id is required")
	ErrNotFound   = errors.New("provider not found")
)
