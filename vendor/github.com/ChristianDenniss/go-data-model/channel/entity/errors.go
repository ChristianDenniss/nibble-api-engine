package entity

import "errors"

var (
	ErrIDRequired = errors.New("channel.id is required")
	ErrNotFound   = errors.New("channel not found")
)
