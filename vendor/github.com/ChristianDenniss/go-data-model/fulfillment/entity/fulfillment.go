package entity

import "strings"

// How food reaches the customer (orthogonal to channel / surface).
const (
	ModeDelivery  = "delivery"
	ModePickup    = "pickup"
	ModeDriveThru = "drive_thru"
	ModeDineIn    = "dine_in"
	ModeInStore   = "in_store"
)

// Who performs delivery when ModeDelivery (empty for non-delivery paths).
const (
	ExecutorThirdParty = "third_party"
	ExecutorMerchant   = "merchant"
)

const (
	ChannelKindAggregator  = "aggregator"
	ChannelKindMerchantApp = "merchant_app"
	ChannelKindMerchantWeb = "merchant_web"
	ChannelKindPhone       = "phone"
	ChannelKindInPerson    = "in_person"
)

// Canonicalize splits legacy combined mode strings and merges with an explicit executor.
func Canonicalize(fulfillmentMode, deliveryExecutor string) (fulfillment string, executor string) {
	fulfillmentMode = strings.TrimSpace(fulfillmentMode)
	deliveryExecutor = strings.TrimSpace(deliveryExecutor)

	if legacy, exec := splitLegacyMode(fulfillmentMode); legacy != "" {
		fulfillment = legacy
		if deliveryExecutor == "" {
			deliveryExecutor = exec
		}
	} else {
		fulfillment = fulfillmentMode
	}
	if fulfillment != ModeDelivery {
		return fulfillment, ""
	}
	return fulfillment, deliveryExecutor
}

func splitLegacyMode(mode string) (fulfillment string, executor string) {
	switch mode {
	case "delivery_3p":
		return ModeDelivery, ExecutorThirdParty
	case "delivery_merchant", "merchant_delivery":
		return ModeDelivery, ExecutorMerchant
	case "in_person":
		return ModeInStore, ""
	default:
		return "", ""
	}
}

// InferDeliveryExecutor fills executor for delivery when ingest omitted it.
func InferDeliveryExecutor(fulfillment, deliveryExecutor, channelKind string) string {
	fulfillment, deliveryExecutor = Canonicalize(fulfillment, deliveryExecutor)
	if fulfillment != ModeDelivery {
		return ""
	}
	if deliveryExecutor != "" {
		return deliveryExecutor
	}
	if channelKind == ChannelKindAggregator {
		return ExecutorThirdParty
	}
	return ExecutorMerchant
}

// NeedsDropoffForQuote is true when a quote observation should be keyed by dropoff geohash.
func NeedsDropoffForQuote(fulfillment string) bool {
	fulfillment, _ = Canonicalize(fulfillment, "")
	return fulfillment == ModeDelivery
}

// OptionMatchesFilter returns whether a path matches one allowed_fulfillment_modes entry.
// Allowed value may be legacy (delivery_3p) or canonical (delivery); executor omitted matches any delivery executor.
// CanonicalizeForChannel is used at ingest/write boundaries.
func CanonicalizeForChannel(fulfillmentMode, deliveryExecutor, channelKind string) (fulfillment string, executor string) {
	fulfillment, executor = Canonicalize(fulfillmentMode, deliveryExecutor)
	executor = InferDeliveryExecutor(fulfillment, executor, channelKind)
	return fulfillment, executor
}

func OptionMatchesFilter(optFulfillment, optExecutor, allowed string) bool {
	optF, optE := Canonicalize(optFulfillment, optExecutor)
	allowF, allowE := Canonicalize(allowed, "")
	if optF != allowF {
		return false
	}
	if allowE == "" {
		return true
	}
	return optE == allowE
}
