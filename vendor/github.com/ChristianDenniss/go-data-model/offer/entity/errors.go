package entity

import "errors"

var (
	ErrIDRequired = errors.New("offer.id is required")
	ErrNotFound   = errors.New("offer not found")
)
