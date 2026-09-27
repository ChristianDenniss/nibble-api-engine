package auth

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/session"
	accountentity "github.com/ChristianDenniss/go-data-model/account/entity"
	accountsvc "github.com/ChristianDenniss/go-data-model/account/service"
	authentity "github.com/ChristianDenniss/go-data-model/auth/entity"
	authsvc "github.com/ChristianDenniss/go-data-model/auth/service"
)

const oauthStateTTL = 10 * time.Minute

type Controller struct {
	svc       *authsvc.Service
	accounts  *accountsvc.Service
	cfg       Config
	providers map[string]oauthProvider
}

func NewController(svc *authsvc.Service, accounts *accountsvc.Service, cfg Config) *Controller {
	providers := map[string]oauthProvider{}
	for _, p := range []oauthProvider{googleProvider{cfg: cfg}, appleProvider{cfg: cfg}} {
		providers[p.id()] = p
	}
	return &Controller{svc: svc, accounts: accounts, cfg: cfg, providers: providers}
}

type accountResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type sessionResponse struct {
	Account *accountResponse `json:"account"`
}

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type providerResponse struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

func (c *Controller) SignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	signed, err := c.svc.SignUp(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	c.respondSignedIn(w, r, http.StatusCreated, signed)
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	signed, err := c.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	c.respondSignedIn(w, r, http.StatusOK, signed)
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := c.svc.Logout(r.Context(), cookie.Value); err != nil {
			log.Printf("auth: logout: %v", err)
		}
	}
	clearSessionCookie(w, c.cfg)
	w.WriteHeader(http.StatusNoContent)
}

// Me reports the signed-in account, or {"account": null} for guests.
func (c *Controller) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	accountID := session.AccountID(r.Context())
	if accountID == "" {
		httpx.WriteJSON(w, http.StatusOK, sessionResponse{})
		return
	}
	account, err := c.accounts.GetByID(r.Context(), accountID)
	if errors.Is(err, accountentity.ErrNotFound) {
		clearSessionCookie(w, c.cfg)
		httpx.WriteJSON(w, http.StatusOK, sessionResponse{})
		return
	}
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to load account")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{Account: toAccountResponse(account)})
}

func (c *Controller) Providers(w http.ResponseWriter, _ *http.Request) {
	out := []providerResponse{}
	for _, id := range []string{authentity.ProviderGoogle, authentity.ProviderApple} {
		out = append(out, providerResponse{ID: id, Enabled: c.providers[id].enabled()})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"providers": out})
}

func (c *Controller) OAuthStart(w http.ResponseWriter, r *http.Request) {
	provider, ok := c.providers[r.PathValue("provider")]
	if !ok {
		httpx.WriteFailure(w, http.StatusNotFound, "unknown provider")
		return
	}
	if !provider.enabled() {
		c.redirectToLogin(w, r, "sso_unavailable")
		return
	}
	state, err := randomString(24)
	if err != nil {
		c.redirectToLogin(w, r, "sso_failed")
		return
	}
	nonce, err := randomString(24)
	if err != nil {
		c.redirectToLogin(w, r, "sso_failed")
		return
	}
	http.SetCookie(w, c.stateCookie(provider, state+"."+nonce, int(oauthStateTTL.Seconds())))
	http.Redirect(w, r, provider.authURL(state, nonce), http.StatusFound)
}

func (c *Controller) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	provider, ok := c.providers[r.PathValue("provider")]
	if !ok {
		httpx.WriteFailure(w, http.StatusNotFound, "unknown provider")
		return
	}
	if !provider.enabled() {
		c.redirectToLogin(w, r, "sso_unavailable")
		return
	}

	cookie, cookieErr := r.Cookie(stateCookieName(provider))
	http.SetCookie(w, c.stateCookie(provider, "", -1))
	if r.FormValue("error") != "" {
		c.redirectToLogin(w, r, "sso_cancelled")
		return
	}
	if cookieErr != nil {
		c.redirectToLogin(w, r, "sso_expired")
		return
	}
	state, nonce, found := strings.Cut(cookie.Value, ".")
	if !found || subtle.ConstantTimeCompare([]byte(state), []byte(r.FormValue("state"))) != 1 {
		c.redirectToLogin(w, r, "sso_expired")
		return
	}

	profile, err := provider.exchange(r.Context(), r, nonce)
	if err != nil {
		log.Printf("auth: %s exchange: %v", provider.id(), err)
		c.redirectToLogin(w, r, "sso_failed")
		return
	}
	signed, err := c.svc.SignInWithProvider(r.Context(), profile)
	if err != nil {
		code := "sso_failed"
		switch {
		case errors.Is(err, authentity.ErrEmailUnverified):
			code = "sso_email_unverified"
		case errors.Is(err, authentity.ErrEmailRequired), errors.Is(err, authentity.ErrEmailInvalid):
			code = "sso_email_missing"
		default:
			log.Printf("auth: %s sign-in: %v", provider.id(), err)
		}
		c.redirectToLogin(w, r, code)
		return
	}
	setSessionCookie(w, c.cfg, signed.Token, signed.ExpiresAt)
	http.Redirect(w, r, c.cfg.WebBaseURL+"/", http.StatusFound)
}

func (c *Controller) respondSignedIn(w http.ResponseWriter, r *http.Request, status int, signed authentity.SignIn) {
	account, err := c.accounts.GetByID(r.Context(), signed.AccountID)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to load account")
		return
	}
	setSessionCookie(w, c.cfg, signed.Token, signed.ExpiresAt)
	httpx.WriteJSON(w, status, sessionResponse{Account: toAccountResponse(account)})
}

func (c *Controller) redirectToLogin(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, c.cfg.WebBaseURL+"/login?error="+url.QueryEscape(code), http.StatusFound)
}

func stateCookieName(p oauthProvider) string {
	return "nibble_oauth_" + p.id()
}

func (c *Controller) stateCookie(p oauthProvider, value string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     stateCookieName(p),
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	if p.formPost() {
		// A cross-site form POST only carries SameSite=None cookies, which browsers require to be Secure.
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = true
	}
	return cookie
}

func toAccountResponse(a accountentity.Account) *accountResponse {
	return &accountResponse{ID: a.ID, Name: a.Name, Email: a.Email}
}

// decodeJSON requires a JSON content type so cross-site HTML forms cannot post
// credentials (browsers preflight JSON requests from other origins).
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		httpx.WriteFailure(w, http.StatusUnsupportedMediaType, "expected application/json")
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(dst); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authentity.ErrNameRequired),
		errors.Is(err, authentity.ErrEmailRequired),
		errors.Is(err, authentity.ErrEmailInvalid),
		errors.Is(err, authentity.ErrPasswordTooShort),
		errors.Is(err, authentity.ErrPasswordTooLong):
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, authentity.ErrEmailTaken):
		httpx.WriteFailure(w, http.StatusConflict, err.Error())
	case errors.Is(err, authentity.ErrInvalidCredentials):
		httpx.WriteFailure(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("auth: %v", err)
		httpx.WriteFailure(w, http.StatusInternalServerError, "sign-in failed")
	}
}
