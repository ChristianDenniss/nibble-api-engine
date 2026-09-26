package entity

import "errors"

var (
	ErrIDRequired = errors.New("itemprice.id is required")
	ErrNotFound   = errors.New("item price observation not found")
)
