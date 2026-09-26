package entity

import "errors"

var (
	ErrIDRequired = errors.New("account.id is required")
	ErrNotFound   = errors.New("account not found")
)
