package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAppleClientSecretIsValidES256(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	p := appleProvider{cfg: Config{
		AppleClientID:   "com.nibble.web",
		AppleTeamID:     "TEAM123",
		AppleKeyID:      "KEY123",
		ApplePrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
	}}

	token, err := p.clientSecret(time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatalf("clientSecret: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts", len(parts))
	}

	var header map[string]string
	decodeSegment(t, parts[0], &header)
	if header["alg"] != "ES256" || header["kid"] != "KEY123" {
		t.Fatalf("header = %v", header)
	}
	var claims map[string]any
	decodeSegment(t, parts[1], &claims)
	if claims["iss"] != "TEAM123" || claims["sub"] != "com.nibble.web" || claims["aud"] != "https://appleid.apple.com" {
		t.Fatalf("claims = %v", claims)
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		t.Fatalf("signature len = %d, err = %v", len(sig), err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r, s := new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(&key.PublicKey, digest[:], r, s) {
		t.Fatal("signature does not verify with the public key")
	}
}

func TestIDTokenClaimsValidate(t *testing.T) {
	var claims idTokenClaims
	raw := `{"iss":"https://appleid.apple.com","aud":["other","com.nibble.web"],"sub":"001",
		"exp":` + jsonInt(time.Now().Add(time.Hour).Unix()) + `,"nonce":"n1","email":"a@b.co","email_verified":"true"}`
	if err := json.Unmarshal([]byte(raw), &claims); err != nil {
		t.Fatal(err)
	}
	if !claims.EmailVerified {
		t.Fatal(`email_verified "true" should parse as true`)
	}
	issuers := []string{"https://appleid.apple.com"}
	if err := claims.validate(issuers, "com.nibble.web", "n1"); err != nil {
		t.Fatalf("valid claims rejected: %v", err)
	}
	if err := claims.validate(issuers, "com.nibble.web", "other-nonce"); err == nil {
		t.Fatal("nonce mismatch accepted")
	}
	if err := claims.validate(issuers, "someone-else", "n1"); err == nil {
		t.Fatal("audience mismatch accepted")
	}
	if err := claims.validate([]string{"https://accounts.google.com"}, "com.nibble.web", "n1"); err == nil {
		t.Fatal("issuer mismatch accepted")
	}
	claims.Expiry = time.Now().Add(-time.Minute).Unix()
	if err := claims.validate(issuers, "com.nibble.web", "n1"); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestOAuthCallbackRejectsStateMismatch(t *testing.T) {
	c := NewController(nil, nil, Config{GoogleClientID: "id", GoogleClientSecret: "secret"})
	mux := http.NewServeMux()
	Mount(mux, c)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/oauth/google/callback?state=attacker&code=x", nil)
	req.AddCookie(&http.Cookie{Name: "nibble_oauth_google", Value: "expected.nonce"})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login?error=sso_expired" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestOAuthStartUnavailableWithoutCredentials(t *testing.T) {
	mux := http.NewServeMux()
	Mount(mux, NewController(nil, nil, Config{}))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/auth/oauth/apple/start", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login?error=sso_unavailable" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}

func decodeSegment(t *testing.T, segment string, dst any) {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatal(err)
	}
}

func jsonInt(v int64) string {
	out, _ := json.Marshal(v)
	return string(out)
}
