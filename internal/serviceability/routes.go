package serviceability

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/serviceability", c.Get)
	mux.HandleFunc("GET /v1/serviceability/restaurants", c.GetRestaurants)
}
