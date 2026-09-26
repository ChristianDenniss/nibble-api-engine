package ingest

import "errors"

var (
	ErrRestaurantIDRequired       = errors.New("restaurant.id is required")
	ErrMenuItemIDRequired         = errors.New("menu_item.id is required")
	ErrOfferIDRequired            = errors.New("offer.id is required")
	ErrPriceObservationIDRequired = errors.New("price_observation.id is required")
)
