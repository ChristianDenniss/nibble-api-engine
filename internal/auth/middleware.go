package auth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/session"
	authentity "github.com/ChristianDenniss/go-data-model/auth/entity"
	authsvc "github.com/ChristianDenniss/go-data-model/auth/service"
)

const sessionCookie = "nibble_session"

// Middleware resolves the session cookie to an account on every request.
// Requests without a valid session continue as guests.
func Middleware(svc *authsvc.Service, cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			accountID, err := svc.Authenticate(r.Context(), cookie.Value)
			switch {
			case err == nil:
				r = r.WithContext(session.WithAccountID(r.Context(), accountID))
			case errors.Is(err, authentity.ErrSessionNotFound):
				clearSessionCookie(w, cfg)
			default:
				log.Printf("auth: session lookup failed: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func setSessionCookie(w http.ResponseWriter, cfg Config, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, cfg Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
