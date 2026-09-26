package entity

import "errors"

var (
	ErrIDRequired = errors.New("order.id is required")
	ErrNotFound   = errors.New("order not found")
)
