package storefront

import (
	"encoding/json"
	"strings"
	"testing"

	restaurantentity "github.com/ChristianDenniss/go-data-model/restaurant/entity"
)

func TestMapRestaurantsHours(t *testing.T) {
	out := mapRestaurants([]restaurantentity.Restaurant{{
		ID: "rest_slice",
		Hours: []restaurantentity.Hours{
			{Service: restaurantentity.HoursStore, DayOfWeek: 5, Opens: "11:00", Closes: "02:00"},
			{Service: restaurantentity.HoursDelivery, DayOfWeek: 5, Opens: "11:00", Closes: "01:00"},
		},
	}})

	raw, err := json.Marshal(out[0].Hours)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"service":"store","dayOfWeek":5,"opens":"11:00","closes":"02:00"},` +
		`{"service":"delivery","dayOfWeek":5,"opens":"11:00","closes":"01:00"}]`
	if string(raw) != want {
		t.Fatalf("hours JSON = %s, want %s", raw, want)
	}
}

// The web Restaurant type requires `hours`, so unknown hours must serialize as [] rather than null.
func TestMapRestaurantsUnknownHoursIsEmptyArray(t *testing.T) {
	raw, err := json.Marshal(mapRestaurants([]restaurantentity.Restaurant{{ID: "rest_green"}})[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"hours":[]`) {
		t.Fatalf("restaurant JSON = %s, want \"hours\":[]", raw)
	}
}
