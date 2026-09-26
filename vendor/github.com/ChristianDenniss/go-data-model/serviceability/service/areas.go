package service

import (
	"context"
	"errors"
	"strings"

	fulfillmententity "github.com/ChristianDenniss/go-data-model/fulfillment/entity"
	serviceabilityentity "github.com/ChristianDenniss/go-data-model/serviceability/entity"
)

// Deliverable checks store status and optional service_areas geohash prefixes (comma-separated).
// geometry "*" or empty with a row present means entire area; no rows means unknown (allow).
func (s *Service) Deliverable(ctx context.Context, sourceStoreID, fulfillment, deliveryExecutor, dropoffGeohash string) (bool, string, error) {
	if st, err := s.status.Get(ctx, sourceStoreID); err == nil {
		if st.Paused || !st.OpenNow {
			return false, "store_not_available", nil
		}
	} else if !errors.Is(err, serviceabilityentity.ErrNotFound) {
		return false, "", err
	}

	if fulfillment != fulfillmententity.ModeDelivery {
		return true, "", nil
	}
	if dropoffGeohash == "" {
		return false, "dropoff_required", nil
	}

	areas, err := s.areas.ListForPath(ctx, sourceStoreID, fulfillment, deliveryExecutor)
	if err != nil {
		return false, "", err
	}
	if len(areas) == 0 {
		return true, "", nil
	}
	for _, a := range areas {
		if geohashCovered(a.Geometry, dropoffGeohash) {
			return true, "", nil
		}
	}
	return false, "outside_service_area", nil
}

func geohashCovered(geometry, geohash string) bool {
	geometry = strings.TrimSpace(geometry)
	if geometry == "" || geometry == "*" {
		return true
	}
	for _, p := range strings.Split(geometry, ",") {
		p = strings.TrimSpace(p)
		if p != "" && strings.HasPrefix(geohash, p) {
			return true
		}
	}
	return false
}
