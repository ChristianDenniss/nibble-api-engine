package entity

import (
	"time"
)

type ServiceArea struct {
	ID               string
	SourceStoreID    string
	FulfillmentMode  string
	DeliveryExecutor string
	Geometry         string
}

type HoursRegular struct {
	ID            string
	SourceStoreID string
	DayOfWeek     int
	Opens         string
	Closes        string
}

type HoursException struct {
	ID            string
	SourceStoreID string
	OnDate        string
	Closed        bool
}

type StoreStatus struct {
	SourceStoreID string
	OpenNow       bool
	Paused        bool
	ObservedAt    time.Time
}
