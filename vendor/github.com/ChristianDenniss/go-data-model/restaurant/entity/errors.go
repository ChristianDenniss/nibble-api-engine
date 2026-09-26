package entity

import "errors"

var (
	ErrIDRequired = errors.New("restaurant.id is required")
	ErrNotFound   = errors.New("restaurant not found")
)
