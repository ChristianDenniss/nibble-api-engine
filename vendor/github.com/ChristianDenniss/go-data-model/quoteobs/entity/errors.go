package entity

import "errors"

var (
	ErrIDRequired = errors.New("quoteobs.id is required")
	ErrNotFound   = errors.New("quote observation not found")
)
