package auth

import (
	"os"
	"strings"
)

type Config struct {
	// CookieSecure marks cookies Secure. Required in production (HTTPS) and for Apple sign-in.
	CookieSecure bool
	// WebBaseURL prefixes browser redirects after SSO. Empty means same origin as the callback.
	WebBaseURL string
	// CallbackBaseURL is the public origin + prefix providers redirect back to,
	// e.g. http://localhost:5174/api (the Vite proxy strips /api).
	CallbackBaseURL string

	GoogleClientID     string
	GoogleClientSecret string

	AppleClientID   string
	AppleTeamID     string
	AppleKeyID      string
	ApplePrivateKey string
}

func ConfigFromEnv() Config {
	return Config{
		CookieSecure:       os.Getenv("AUTH_COOKIE_SECURE") == "true",
		WebBaseURL:         strings.TrimRight(os.Getenv("WEB_BASE_URL"), "/"),
		CallbackBaseURL:    strings.TrimRight(getenv("OAUTH_CALLBACK_BASE_URL", "http://localhost:5174/api"), "/"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		AppleClientID:      os.Getenv("APPLE_CLIENT_ID"),
		AppleTeamID:        os.Getenv("APPLE_TEAM_ID"),
		AppleKeyID:         os.Getenv("APPLE_KEY_ID"),
		ApplePrivateKey:    applePrivateKey(),
	}
}

func applePrivateKey() string {
	if key := os.Getenv("APPLE_PRIVATE_KEY"); key != "" {
		return strings.ReplaceAll(key, `\n`, "\n")
	}
	if path := os.Getenv("APPLE_PRIVATE_KEY_PATH"); path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			return string(raw)
		}
	}
	return ""
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
