package entity

import "errors"

var (
	ErrIDRequired = errors.New("place.id is required")
	ErrNotFound   = errors.New("place not found")
)
