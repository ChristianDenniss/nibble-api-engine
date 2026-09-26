package entity

import location "github.com/ChristianDenniss/go-data-model/location/entity"

type SavedAddress struct {
	ID       string
	Label    string
	Location location.Location
	Current  bool
}
