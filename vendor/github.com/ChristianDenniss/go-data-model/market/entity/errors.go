package entity

import "errors"

var (
	ErrIDRequired    = errors.New("market.id is required")
	ErrNotFound      = errors.New("market not found")
	ErrSlugRequired  = errors.New("market.slug is required")
	ErrStatusInvalid = errors.New("market coverage status is invalid")
)
