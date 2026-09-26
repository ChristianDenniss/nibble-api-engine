package storefront

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/storefront", c.Get)
	mux.HandleFunc("GET /storefront", c.Get)
}
