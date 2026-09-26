package sourcestores

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/channels/{channelId}/source-stores", c.ListByChannel)
}
