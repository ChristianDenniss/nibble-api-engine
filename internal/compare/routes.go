package compare

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("POST /v1/compare", c.PostCompare)
	mux.HandleFunc("GET /v1/compare/sessions/{id}", c.GetSession)
}
