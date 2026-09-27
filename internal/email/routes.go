package email

import "net/http"

func Mount(mux *http.ServeMux, controller *Controller) {
	mux.HandleFunc("GET /v1/email/gmail/start", controller.Start)
	mux.HandleFunc("GET /v1/email/gmail/callback", controller.Callback)
	mux.HandleFunc("GET /v1/email/gmail/status", controller.Status)
	mux.HandleFunc("POST /v1/email/gmail/sync", controller.Sync)
	mux.HandleFunc("GET /v1/email/messages", controller.ListMessages)
}
