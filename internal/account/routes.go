package account

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("DELETE /v1/accounts/{accountId}/addresses/{addressId}", c.DeleteAddress)
	mux.HandleFunc("PUT /v1/accounts/{accountId}/addresses/{addressId}/current", c.SetCurrentAddress)
}
