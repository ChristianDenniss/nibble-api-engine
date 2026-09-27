package auth

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("POST /v1/auth/signup", c.SignUp)
	mux.HandleFunc("POST /v1/auth/login", c.Login)
	mux.HandleFunc("POST /v1/auth/logout", c.Logout)
	mux.HandleFunc("GET /v1/auth/me", c.Me)
	mux.HandleFunc("GET /v1/auth/providers", c.Providers)
	mux.HandleFunc("GET /v1/auth/oauth/{provider}/start", c.OAuthStart)
	mux.HandleFunc("/v1/auth/oauth/{provider}/callback", c.OAuthCallback)
}
