package entity

import "errors"

var (
	ErrIDRequired        = errors.New("account.id is required")
	ErrNotFound          = errors.New("account not found")
	ErrAddressIDRequired = errors.New("saved address id is required")
	ErrAddressNotFound   = errors.New("saved address not found")
)
