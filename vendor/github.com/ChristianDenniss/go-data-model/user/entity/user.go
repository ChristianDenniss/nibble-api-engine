package entity

import (
	"encoding/json"

	location "github.com/ChristianDenniss/go-data-model/location/entity"
)

type User struct {
	ID    string
	Name  string
	Email string
	Phone string
}

type ComparePrefs struct {
	AllowedFulfillmentModes   []string `json:"allowed_fulfillment_modes"`
	WillingToUseAggregator  bool     `json:"willing_to_use_aggregator"`
	AllowedChannelIDs       []string `json:"allowed_channel_ids"`
	BlockedChannelIDs       []string `json:"blocked_channel_ids"`
	DriveThruOK             bool     `json:"drive_thru_ok"`
	MerchantDirectOK        bool     `json:"merchant_direct_ok"`
}

type Settings struct {
	UserID                string
	PreferredCurrency     string
	NotificationsEnabled  bool
	ComparePrefs          ComparePrefs
}

type Dropoff struct {
	ID       string
	UserID   string
	Label    string
	Location location.Location
	Current  bool
}

type Membership struct {
	ID                  string
	UserID              string
	MembershipProductID string
}

type Session struct {
	ID              string
	UserID          string
	PlaceID         string
	QuerySnapshot   json.RawMessage
	BasketSnapshot  json.RawMessage
	ResultSnapshot  json.RawMessage
}

type OutboundClick struct {
	ID               string
	UserID           string
	CompareSessionID string
	PurchaseOptionID string
	ActionKind       string
	TargetURL        string
}
