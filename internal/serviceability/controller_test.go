package serviceability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceabilityentity "github.com/ChristianDenniss/go-data-model/serviceability/entity"
	serviceabilityrepo "github.com/ChristianDenniss/go-data-model/serviceability/repository"
	serviceabilitysvc "github.com/ChristianDenniss/go-data-model/serviceability/service"
)

type fakeStatusRepo struct{}

func (fakeStatusRepo) Get(context.Context, string) (serviceabilityentity.StoreStatus, error) {
	return serviceabilityentity.StoreStatus{OpenNow: true}, nil
}

func (fakeStatusRepo) Upsert(context.Context, serviceabilityentity.StoreStatus) error { return nil }

type fakeAreaRepo struct {
	areas []serviceabilityentity.ServiceArea
}

type fakePathRepo struct{}

func (fakePathRepo) ListByRestaurants(context.Context, []string) ([]serviceabilityentity.RestaurantStorePath, error) {
	return []serviceabilityentity.RestaurantStorePath{{RestaurantID: "rest_1", ProviderID: "prov_1", SourceStoreID: "store_1"}}, nil
}

func (fakePathRepo) Upsert(context.Context, serviceabilityentity.RestaurantStorePath) error {
	return nil
}

func (r fakeAreaRepo) ListForPath(context.Context, string, string, string) ([]serviceabilityentity.ServiceArea, error) {
	return r.areas, nil
}

func (fakeAreaRepo) Upsert(context.Context, serviceabilityentity.ServiceArea) error { return nil }

var _ serviceabilityrepo.StoreStatusRepository = fakeStatusRepo{}
var _ serviceabilityrepo.ServiceAreaRepository = fakeAreaRepo{}

func TestGetReturnsCoverageDecision(t *testing.T) {
	svc := serviceabilitysvc.New(fakeStatusRepo{}, fakeAreaRepo{areas: []serviceabilityentity.ServiceArea{{Geometry: "f80t7"}}})
	c := NewController(svc)
	r := httptest.NewRequest(http.MethodGet, "/v1/serviceability?source_store_id=store_1&fulfillment_mode=delivery&delivery_executor=third_party&dropoff_geohash=f80t7s", nil)
	w := httptest.NewRecorder()
	c.Get(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); !strings.Contains(body, `"deliverable":true`) || !strings.Contains(body, `"reason":""`) {
		t.Fatalf("body = %s", body)
	}
}

func TestGetRequiresPathIdentity(t *testing.T) {
	c := NewController(serviceabilitysvc.New(fakeStatusRepo{}, fakeAreaRepo{}))
	r := httptest.NewRequest(http.MethodGet, "/v1/serviceability", nil)
	w := httptest.NewRecorder()
	c.Get(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetBuildsGeohashFromCoordinates(t *testing.T) {
	svc := serviceabilitysvc.New(fakeStatusRepo{}, fakeAreaRepo{areas: []serviceabilityentity.ServiceArea{{Geometry: "f80t7"}}})
	c := NewController(svc)
	r := httptest.NewRequest(http.MethodGet, "/v1/serviceability?source_store_id=store_1&fulfillment_mode=delivery&delivery_executor=third_party&latitude=45.9458&longitude=-66.6414", nil)
	w := httptest.NewRecorder()
	c.Get(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"deliverable":true`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestGetRestaurantsReturnsProviderPaths(t *testing.T) {
	svc := serviceabilitysvc.New(fakeStatusRepo{}, fakeAreaRepo{areas: []serviceabilityentity.ServiceArea{{Geometry: "f80t7"}}})
	c := NewController(svc, fakePathRepo{})
	r := httptest.NewRequest(http.MethodGet, "/v1/serviceability/restaurants?restaurant_ids=rest_1&latitude=45.9458&longitude=-66.6414", nil)
	w := httptest.NewRecorder()
	c.GetRestaurants(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"providerId":"prov_1"`) || !strings.Contains(w.Body.String(), `"status":"covered"`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
