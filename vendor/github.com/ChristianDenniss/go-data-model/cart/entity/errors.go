package entity

import "errors"

var (
	ErrIDRequired = errors.New("cart.id is required")
	ErrNotFound   = errors.New("cart not found")
)
