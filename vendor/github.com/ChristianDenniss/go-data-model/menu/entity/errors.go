package entity

import "errors"

var (
	ErrIDRequired = errors.New("menu_item.id is required")
	ErrNotFound   = errors.New("menu item not found")
)
