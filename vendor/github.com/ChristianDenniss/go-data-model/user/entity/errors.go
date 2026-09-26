package entity

import "errors"

var (
	ErrIDRequired = errors.New("user.id is required")
	ErrNotFound   = errors.New("user not found")
)
