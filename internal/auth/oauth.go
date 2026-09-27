package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	authentity "github.com/ChristianDenniss/go-data-model/auth/entity"
)

type oauthProvider interface {
	id() string
	enabled() bool
	authURL(state, nonce string) string
	// exchange trades the callback code for the provider's verified profile.
	exchange(ctx context.Context, r *http.Request, nonce string) (authentity.ExternalProfile, error)
	// formPost providers call back with a cross-site POST, so their state cookie must be SameSite=None.
	formPost() bool
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func callbackURL(cfg Config, provider string) string {
	return cfg.CallbackBaseURL + "/v1/auth/oauth/" + provider + "/callback"
}

type googleProvider struct{ cfg Config }

func (googleProvider) id() string     { return authentity.ProviderGoogle }
func (googleProvider) formPost() bool { return false }

func (p googleProvider) enabled() bool {
	return p.cfg.GoogleClientID != "" && p.cfg.GoogleClientSecret != ""
}

func (p googleProvider) authURL(state, nonce string) string {
	q := url.Values{
		"client_id":     {p.cfg.GoogleClientID},
		"redirect_uri":  {callbackURL(p.cfg, p.id())},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"nonce":         {nonce},
		"prompt":        {"select_account"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode()
}

func (p googleProvider) exchange(ctx context.Context, r *http.Request, nonce string) (authentity.ExternalProfile, error) {
	claims, err := exchangeCode(ctx, "https://oauth2.googleapis.com/token", url.Values{
		"code":          {r.FormValue("code")},
		"client_id":     {p.cfg.GoogleClientID},
		"client_secret": {p.cfg.GoogleClientSecret},
		"redirect_uri":  {callbackURL(p.cfg, p.id())},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return authentity.ExternalProfile{}, err
	}
	if err := claims.validate([]string{"https://accounts.google.com", "accounts.google.com"}, p.cfg.GoogleClientID, nonce); err != nil {
		return authentity.ExternalProfile{}, err
	}
	return authentity.ExternalProfile{
		Provider:      p.id(),
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: bool(claims.EmailVerified),
		Name:          claims.Name,
	}, nil
}

type appleProvider struct{ cfg Config }

func (appleProvider) id() string     { return authentity.ProviderApple }
func (appleProvider) formPost() bool { return true }

func (p appleProvider) enabled() bool {
	return p.cfg.AppleClientID != "" && p.cfg.AppleTeamID != "" && p.cfg.AppleKeyID != "" && p.cfg.ApplePrivateKey != ""
}

func (p appleProvider) authURL(state, nonce string) string {
	q := url.Values{
		"client_id":     {p.cfg.AppleClientID},
		"redirect_uri":  {callbackURL(p.cfg, p.id())},
		"response_type": {"code"},
		"response_mode": {"form_post"},
		"scope":         {"name email"},
		"state":         {state},
		"nonce":         {nonce},
	}
	return "https://appleid.apple.com/auth/authorize?" + q.Encode()
}

func (p appleProvider) exchange(ctx context.Context, r *http.Request, nonce string) (authentity.ExternalProfile, error) {
	secret, err := p.clientSecret(time.Now())
	if err != nil {
		return authentity.ExternalProfile{}, err
	}
	claims, err := exchangeCode(ctx, "https://appleid.apple.com/auth/token", url.Values{
		"code":          {r.FormValue("code")},
		"client_id":     {p.cfg.AppleClientID},
		"client_secret": {secret},
		"redirect_uri":  {callbackURL(p.cfg, p.id())},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return authentity.ExternalProfile{}, err
	}
	if err := claims.validate([]string{"https://appleid.apple.com"}, p.cfg.AppleClientID, nonce); err != nil {
		return authentity.ExternalProfile{}, err
	}
	return authentity.ExternalProfile{
		Provider:      p.id(),
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: bool(claims.EmailVerified),
		// Apple sends the name only on the very first authorization, in the form body.
		Name: appleUserName(r.FormValue("user")),
	}, nil
}

// clientSecret is the ES256 JWT Apple requires in place of a static secret.
func (p appleProvider) clientSecret(now time.Time) (string, error) {
	block, _ := pem.Decode([]byte(p.cfg.ApplePrivateKey))
	if block == nil {
		return "", errors.New("apple: private key is not PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("apple: parse private key: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("apple: private key is not ECDSA")
	}

	header, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": p.cfg.AppleKeyID})
	payload, _ := json.Marshal(map[string]any{
		"iss": p.cfg.AppleTeamID,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": p.cfg.AppleClientID,
	})
	signingInput := b64(header) + "." + b64(payload)
	digest := sha256.Sum256([]byte(signingInput))
	rInt, sInt, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)
	rInt.FillBytes(sig[:32])
	sInt.FillBytes(sig[32:])
	return signingInput + "." + b64(sig), nil
}

func appleUserName(raw string) string {
	if raw == "" {
		return ""
	}
	var user struct {
		Name struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"name"`
	}
	if json.Unmarshal([]byte(raw), &user) != nil {
		return ""
	}
	return strings.TrimSpace(user.Name.FirstName + " " + user.Name.LastName)
}

type idTokenClaims struct {
	Issuer        string       `json:"iss"`
	Audience      audience     `json:"aud"`
	Subject       string       `json:"sub"`
	Expiry        int64        `json:"exp"`
	Nonce         string       `json:"nonce"`
	Email         string       `json:"email"`
	EmailVerified flexibleBool `json:"email_verified"`
	Name          string       `json:"name"`
}

func (c idTokenClaims) validate(issuers []string, clientID, nonce string) error {
	issuerOK := false
	for _, iss := range issuers {
		if c.Issuer == iss {
			issuerOK = true
		}
	}
	switch {
	case !issuerOK:
		return fmt.Errorf("id_token: unexpected issuer %q", c.Issuer)
	case !c.Audience.contains(clientID):
		return errors.New("id_token: audience mismatch")
	case time.Now().Unix() >= c.Expiry:
		return errors.New("id_token: expired")
	case c.Nonce != nonce:
		return errors.New("id_token: nonce mismatch")
	case c.Subject == "":
		return errors.New("id_token: missing subject")
	}
	return nil
}

// exchangeCode posts to the provider token endpoint and decodes the returned
// id_token. The signature is not checked: the token arrives directly from the
// provider over TLS in exchange for our client secret (OIDC Core 3.1.3.7).
func exchangeCode(ctx context.Context, endpoint string, form url.Values) (idTokenClaims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return idTokenClaims{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return idTokenClaims{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return idTokenClaims{}, err
	}
	if res.StatusCode != http.StatusOK {
		return idTokenClaims{}, fmt.Errorf("token endpoint %s: %d %s", endpoint, res.StatusCode, strings.TrimSpace(string(body)))
	}
	var tokens struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tokens); err != nil || tokens.IDToken == "" {
		return idTokenClaims{}, errors.New("token endpoint: missing id_token")
	}
	parts := strings.Split(tokens.IDToken, ".")
	if len(parts) != 3 {
		return idTokenClaims{}, errors.New("id_token: malformed")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return idTokenClaims{}, fmt.Errorf("id_token: %w", err)
	}
	var claims idTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return idTokenClaims{}, fmt.Errorf("id_token: %w", err)
	}
	return claims, nil
}

// audience accepts the JWT `aud` claim as either a string or an array.
type audience []string

func (a *audience) UnmarshalJSON(raw []byte) error {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		*a = audience{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a audience) contains(want string) bool {
	for _, v := range a {
		if v == want {
			return true
		}
	}
	return false
}

// flexibleBool accepts Apple's "true"/"false" strings as well as JSON booleans.
type flexibleBool bool

func (b *flexibleBool) UnmarshalJSON(raw []byte) error {
	switch strings.Trim(string(raw), `"`) {
	case "true":
		*b = true
	default:
		*b = false
	}
	return nil
}

func b64(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func randomString(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return b64(buf), nil
}
