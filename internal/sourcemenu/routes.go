package sourcemenu

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/source-stores/{storeId}/menu", c.GetMenu)
}
