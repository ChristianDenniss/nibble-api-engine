package promotions

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("POST /v1/deals", c.Insert)
	mux.HandleFunc("POST /v1/deals/text", c.InsertText)
}
