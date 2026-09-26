package entity

import "errors"

var (
	ErrIDRequired = errors.New("category.id is required")
	ErrNotFound   = errors.New("category not found")
)
