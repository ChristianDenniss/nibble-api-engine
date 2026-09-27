package serviceability

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	compareService "github.com/ChristianDenniss/go-data-model/compare/service"
	serviceabilityentity "github.com/ChristianDenniss/go-data-model/serviceability/entity"
	serviceabilityrepo "github.com/ChristianDenniss/go-data-model/serviceability/repository"
	serviceabilitysvc "github.com/ChristianDenniss/go-data-model/serviceability/service"
)

type Controller struct {
	svc     *serviceabilitysvc.Service
	paths   serviceabilityrepo.RestaurantStorePathRepository
	cacheMu sync.Mutex
	cache   map[string]cachedBatch
}

type cachedBatch struct {
	expires time.Time
	value   batchResponse
}

func NewController(svc *serviceabilitysvc.Service, paths ...serviceabilityrepo.RestaurantStorePathRepository) *Controller {
	var pathRepo serviceabilityrepo.RestaurantStorePathRepository
	if len(paths) > 0 {
		pathRepo = paths[0]
	}
	return &Controller{svc: svc, paths: pathRepo, cache: make(map[string]cachedBatch)}
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	storeID := query.Get("source_store_id")
	fulfillment := query.Get("fulfillment_mode")
	executor := query.Get("delivery_executor")
	dropoffGeohash := query.Get("dropoff_geohash")
	if storeID == "" || fulfillment == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "source_store_id and fulfillment_mode are required"})
		return
	}
	var latitude, longitude float64
	hasCoordinates := query.Get("latitude") != "" || query.Get("longitude") != ""
	if hasCoordinates {
		var latitudeErr, longitudeErr error
		latitude, latitudeErr = strconv.ParseFloat(query.Get("latitude"), 64)
		longitude, longitudeErr = strconv.ParseFloat(query.Get("longitude"), 64)
		if latitudeErr != nil || longitudeErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "latitude and longitude must be valid coordinates"})
			return
		}
		if dropoffGeohash == "" {
			dropoffGeohash = compareService.EncodeGeohash(latitude, longitude, 5)
		}
	}

	var deliverable bool
	var reason string
	var err error
	if hasCoordinates {
		deliverable, reason, err = c.svc.DeliverableAt(r.Context(), storeID, fulfillment, executor, dropoffGeohash, latitude, longitude)
	} else {
		deliverable, reason, err = c.svc.Deliverable(r.Context(), storeID, fulfillment, executor, dropoffGeohash)
	}
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "serviceability lookup failed"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"sourceStoreId":    storeID,
		"fulfillmentMode":  fulfillment,
		"deliveryExecutor": executor,
		"dropoffGeohash":   dropoffGeohash,
		"deliverable":      deliverable,
		"status":           decisionStatus(deliverable, reason),
		"reason":           reason,
	})
}

type batchResponse struct {
	Restaurants []restaurantCoverage `json:"restaurants"`
}

type restaurantCoverage struct {
	RestaurantID string         `json:"restaurantId"`
	Paths        []pathDecision `json:"paths"`
}

type pathDecision struct {
	ProviderID    string `json:"providerId"`
	SourceStoreID string `json:"sourceStoreId"`
	Deliverable   bool   `json:"deliverable"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
}

func (c *Controller) GetRestaurants(w http.ResponseWriter, r *http.Request) {
	if c.paths == nil {
		httpx.WriteJSON(w, http.StatusNotImplemented, map[string]string{"error": "restaurant serviceability mapping is not configured"})
		return
	}
	query := r.URL.Query()
	ids := splitIDs(query.Get("restaurant_ids"))
	if len(ids) == 0 {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "restaurant_ids is required"})
		return
	}
	latitude, latitudeErr := strconv.ParseFloat(query.Get("latitude"), 64)
	longitude, longitudeErr := strconv.ParseFloat(query.Get("longitude"), 64)
	if latitudeErr != nil || longitudeErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "latitude and longitude must be valid coordinates"})
		return
	}
	fulfillment := query.Get("fulfillment_mode")
	if fulfillment == "" {
		fulfillment = "delivery"
	}
	executor := query.Get("delivery_executor")
	geohash := compareService.EncodeGeohash(latitude, longitude, 5)
	cacheKey := strings.Join([]string{strings.Join(ids, ","), fulfillment, executor, geohash}, "|")
	if cached, ok := c.getCached(cacheKey); ok {
		httpx.WriteJSON(w, http.StatusOK, cached)
		return
	}

	paths, err := c.paths.ListByRestaurants(r.Context(), ids)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "serviceability mapping lookup failed"})
		return
	}
	byRestaurant := make(map[string][]serviceabilityentity.RestaurantStorePath)
	for _, path := range paths {
		byRestaurant[path.RestaurantID] = append(byRestaurant[path.RestaurantID], path)
	}
	response := batchResponse{Restaurants: make([]restaurantCoverage, 0, len(ids))}
	for _, id := range ids {
		coverage := restaurantCoverage{RestaurantID: id}
		for _, path := range byRestaurant[id] {
			ok, reason, err := c.svc.DeliverableAt(r.Context(), path.SourceStoreID, fulfillment, executor, geohash, latitude, longitude)
			if err != nil {
				httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "serviceability lookup failed"})
				return
			}
			coverage.Paths = append(coverage.Paths, pathDecision{ProviderID: path.ProviderID, SourceStoreID: path.SourceStoreID, Deliverable: ok, Status: decisionStatus(ok, reason), Reason: reason})
		}
		response.Restaurants = append(response.Restaurants, coverage)
	}
	c.putCached(cacheKey, response)
	httpx.WriteJSON(w, http.StatusOK, response)
}

func decisionStatus(deliverable bool, reason string) string {
	if reason == "coverage_unknown" || reason == "status_stale" {
		return "unknown"
	}
	if deliverable {
		return "covered"
	}
	return "unavailable"
}

func splitIDs(raw string) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			if _, ok := seen[value]; !ok {
				seen[value] = struct{}{}
				ids = append(ids, value)
			}
		}
	}
	return ids
}

func (c *Controller) getCached(key string) (batchResponse, bool) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	entry, ok := c.cache[key]
	if !ok || time.Now().After(entry.expires) {
		if ok {
			delete(c.cache, key)
		}
		return batchResponse{}, false
	}
	return entry.value, true
}

func (c *Controller) putCached(key string, value batchResponse) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache[key] = cachedBatch{expires: time.Now().Add(30 * time.Second), value: value}
}
