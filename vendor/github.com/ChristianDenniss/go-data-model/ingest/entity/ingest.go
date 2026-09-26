package entity

import (
	"encoding/json"
	"time"
)

type IngestRun struct {
	ID             string
	JobType        string
	ChannelID      string
	StartedAt      time.Time
	FinishedAt     *time.Time
	ParserVersion  string
}

type SourceSnapshot struct {
	ID              string
	IngestRunID     string
	ExternalStoreID string
	ContentType     string
	RawJSON         json.RawMessage
	Checksum        string
	ObservedAt      time.Time
}
