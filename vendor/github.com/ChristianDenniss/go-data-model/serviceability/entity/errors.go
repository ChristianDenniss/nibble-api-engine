package entity

import "errors"

var (
	ErrIDRequired = errors.New("serviceability.id is required")
	ErrNotFound   = errors.New("serviceability record not found")
)
