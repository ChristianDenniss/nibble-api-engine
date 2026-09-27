package sms

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("POST /v1/integrations/twilio/inbound", c.ReceiveTwilio)
	mux.HandleFunc("POST /v1/sms/send", c.Send)
	mux.HandleFunc("POST /v1/sms/sources", c.RegisterSource)
	mux.HandleFunc("GET /v1/sms/sources", c.ListSources)
	mux.HandleFunc("GET /v1/sms/messages", c.ListMessages)
	mux.HandleFunc("POST /v1/sms/messages/{externalID}/retry", c.RetryMessage)
}
