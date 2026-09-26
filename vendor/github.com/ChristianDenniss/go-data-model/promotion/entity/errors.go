package entity

import "errors"

var (
	ErrIDRequired = errors.New("promotion.id is required")
	ErrNotFound   = errors.New("promotion not found")
)
