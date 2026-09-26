package entity

import location "github.com/ChristianDenniss/go-data-model/location/entity"

type Restaurant struct {
	ID          string
	Name        string
	Location    location.Location
	CuisineIDs  []string
	CategoryIDs []string
	Rating      Rating
}
