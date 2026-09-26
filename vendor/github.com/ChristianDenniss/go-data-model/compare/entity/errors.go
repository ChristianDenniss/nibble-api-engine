package entity

import "errors"

var (
	ErrPlaceRequired  = errors.New("compare.place_id is required")
	ErrBasketRequired = errors.New("compare.basket is required")
)
