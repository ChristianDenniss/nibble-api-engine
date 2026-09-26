package entity

import "errors"

var (
	ErrIDRequired = errors.New("dish.id is required")
	ErrNotFound   = errors.New("dish not found")
)
