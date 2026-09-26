package entity

import money "github.com/ChristianDenniss/go-data-model/money/entity"

// HTTP/OpenAPI-aligned DTOs (also emitted to TypeScript via gentypes).

type APIPathRank struct {
	PurchaseOptionID string
	Rank             int
	Kind             string
	Headline         string
	Confidence       string
	AllIn            money.Money
	FulfillmentMode  string
	DeliveryExecutor string
	ChannelID        string
	RationaleBullets []string
}

type APIUnavailablePath struct {
	PurchaseOptionID string
	Code             string
	Message          string
}

type APICompareResponse struct {
	CompareSessionID string
	ObservedAt       string
	PathsRanked      int
	Recommendation   *APIPathRank
	RunnersUp        []APIPathRank
	UnavailablePaths []APIUnavailablePath
}
