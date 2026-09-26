package entity

import "errors"

var (
	ErrIDRequired = errors.New("price_observation.id is required")
	ErrNotFound   = errors.New("price observation not found")
)
