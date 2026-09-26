package entity

import "errors"

var (
	ErrIDRequired = errors.New("brand.id is required")
	ErrNotFound   = errors.New("brand not found")
)
