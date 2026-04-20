package main

import (
	"net/http"
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// newAuthedClient returns a Robinhood client with a Session loaded per the
// documented precedence: env vars > keychain > fail.
//
// Per Fix P, the ROBINHOOD_API_HOST env var (when set) redirects all three
// Robinhood hosts (api / nummus / phoenix) at the single override URL —
// useful for local dev, mitmproxy captures, and integration tests.
func newAuthedClient() (*robinhood.Client, error) {
	// Fix K: reject invalid profile names before they touch the keychain.
	if err := config.ValidProfile(gflags.Profile); err != nil {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
	}

	var sess *robinhood.Session
	if os.Getenv(config.EnvAccessToken) != "" || os.Getenv(config.EnvRefreshToken) != "" {
		s, err := robinhood.LoadSessionFromEnv()
		if err != nil {
			return nil, err
		}
		sess = s
	} else {
		s, err := robinhood.LoadFromKeychain(gflags.Profile)
		if err != nil {
			return nil, err
		}
		sess = s
	}

	c := clientForEnv()
	c.SetSession(sess)
	// Fix C/P: thread the active --profile flag to the Client so the refresh
	// path reads/writes the correct keychain entry.
	c.SetProfile(gflags.Profile)
	return c, nil
}

// clientForEnv constructs the Client, honoring ROBINHOOD_API_HOST (Fix P).
func clientForEnv() *robinhood.Client {
	h := os.Getenv("ROBINHOOD_API_HOST")
	if h == "" {
		return robinhood.NewClient()
	}
	return robinhood.NewClientWithHosts(h, h, h, &http.Client{Timeout: 30 * time.Second})
}
