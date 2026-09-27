package home

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/home", c.Get)
}
