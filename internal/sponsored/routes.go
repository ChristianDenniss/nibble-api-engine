package sponsored

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("POST /v1/sponsored/events", c.RecordEvent)
}
