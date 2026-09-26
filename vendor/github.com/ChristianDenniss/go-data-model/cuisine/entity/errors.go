package entity

import "errors"

var (
	ErrIDRequired = errors.New("cuisine.id is required")
	ErrNotFound   = errors.New("cuisine not found")
)
