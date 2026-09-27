package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDPreservesIncomingID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "request-123" {
			t.Fatalf("request id in context = %q", got)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "request-123")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if got := res.Header().Get("X-Request-ID"); got != "request-123" {
		t.Fatalf("response request id = %q", got)
	}
}

func TestRateLimitOnlyAppliesToProtectedBuckets(t *testing.T) {
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
		req.RemoteAddr = "192.0.2.10:1234"
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusNoContent {
			t.Fatalf("request %d returned %d", i+1, res.Code)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("limited request returned %d", res.Code)
	}

	public := httptest.NewRequest(http.MethodGet, "/v1/restaurants", nil)
	public.RemoteAddr = "192.0.2.10:1234"
	publicRes := httptest.NewRecorder()
	handler.ServeHTTP(publicRes, public)
	if publicRes.Code != http.StatusNoContent {
		t.Fatalf("public request returned %d", publicRes.Code)
	}
}
