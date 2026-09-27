package entity

import "errors"

var (
	ErrAccountIDRequired = errors.New("account id required")
	ErrNotFound          = errors.New("storefront catalog not found")
)
