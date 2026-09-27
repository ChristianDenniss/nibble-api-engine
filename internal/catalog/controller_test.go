package catalog

import (
	"context"
	"encoding/json"
	domain "github.com/ChristianDenniss/go-data-model/catalog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type catalogRepo struct{}

func (catalogRepo) Import(context.Context, domain.Bundle, []byte) error { return nil }
func (catalogRepo) Load(context.Context) (domain.Bundle, error) {
	n := int64(839)
	return domain.Bundle{Version: 1, Providers: map[string]domain.Snapshot{"Uber Eats": {Stores: []domain.Store{{ID: "test", Name: "Restaurant", URL: "https://www.ubereats.com/ca/store/test/test", Items: []domain.MenuItem{{Name: "Big Mac", Amount: &n, Currency: "CAD"}}}}}}}, nil
}
func TestCartComparisonHTTP(t *testing.T) {
	mux := http.NewServeMux()
	Mount(mux, domain.New(catalogRepo{}))
	valid := `{"restaurantId":"catalog-test","lines":[{"itemId":"Big Mac","quantity":2}]}`
	for _, tc := range []struct {
		body   string
		status int
	}{{valid, 200}, {strings.Replace(valid, `"quantity":2`, `"quantity":0`, 1), 400}, {strings.Replace(valid, `"quantity":2`, `"quantity":2,"amountCents":1`, 1), 400}, {valid + `{}`, 400}, {`{}`, 400}} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/v1/catalog/cart-comparison", strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Fatalf("got %d want %d: %s", w.Code, tc.status, w.Body.String())
		}
		if w.Code == 200 {
			var result domain.CartComparison
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if *result.Providers[0].Subtotal != 1678 || result.Providers[0].Total != nil {
				t.Fatal("incorrect basket total")
			}
		}
	}
}

func TestHandoffDoesNotPretendToCreateProviderCart(t *testing.T) {
	mux := http.NewServeMux()
	Mount(mux, domain.New(catalogRepo{}))
	for _, provider := range []string{"Uber Eats", "SkipTheDishes"} {
		w := httptest.NewRecorder()
		body := `{"restaurantId":"catalog-test","provider":"` + provider + `","lines":[{"itemId":"Big Mac","quantity":2}]}`
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/v1/catalog/cart-handoff", strings.NewReader(body)))
		if provider == "SkipTheDishes" {
			if w.Code != 400 {
				t.Fatal("unlinked provider accepted")
			}
			continue
		}
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		var result domain.CartHandoff
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.CartTransferred || result.Mode != "menu_link" || !strings.Contains(result.CartText, "2 × Big Mac") || result.URL != "https://www.ubereats.com/ca/store/test/test" {
			t.Fatal("invalid handoff", result)
		}
	}
}
