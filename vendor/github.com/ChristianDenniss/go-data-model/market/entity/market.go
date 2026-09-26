package entity

import (
	"time"

	location "github.com/ChristianDenniss/go-data-model/location/entity"
)

const (
	StatusPlanned = "planned"
	StatusActive  = "active"
	StatusPaused  = "paused"
)

const (
	CoverageExpected = "expected"
	CoverageObserved = "observed"
	CoverageAbsent   = "absent"
	CoverageUnknown  = "unknown"
)

// Market is a Nibble launch geography (metro), not a provider delivery zone.
type Market struct {
	ID              string
	Slug            string
	Name            string
	Country         string
	Region          string
	Currency        string
	Timezone        string
	Status          string
	GeohashPrefixes string
}

// ProbeDropoff is a representative address used to check channel coverage.
type ProbeDropoff struct {
	ID       string
	MarketID string
	Label    string
	Geohash  string
	Location location.Location
}

// ChannelCoverage is whether a channel operates in a market (not store-level zones).
type ChannelCoverage struct {
	ID             string
	ChannelID      string
	MarketID       string
	Status         string
	StoreCount     int
	IngestRunID    string
	Note           string
	LastObservedAt *time.Time
}
