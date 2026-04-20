package main

import (
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestNewAuthedClient_PrefersEnv(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "env-access")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "env-refresh")
	// keychain has a different token:
	kc := &robinhood.Session{AccessToken: "kc", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
	if err := kc.SaveToKeychain("default"); err != nil {
		t.Fatal(err)
	}
	gflags.Profile = "default"
	c, err := newAuthedClient()
	if err != nil {
		t.Fatalf("newAuthedClient: %v", err)
	}
	s := c.Session()
	if s.AccessToken != "env-access" {
		t.Fatalf("got %q, want env-access", s.AccessToken)
	}
	// Fix N: env-loaded sessions must be marked Ephemeral so that refresh
	// never writes them back to the keychain.
	if !s.Ephemeral {
		t.Fatalf("expected env-loaded session to have Ephemeral=true")
	}
}

func TestNewAuthedClient_FallsBackToKeychain(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "")
	kc := &robinhood.Session{AccessToken: "kc", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
	if err := kc.SaveToKeychain("default"); err != nil {
		t.Fatal(err)
	}
	gflags.Profile = "default"
	c, err := newAuthedClient()
	if err != nil {
		t.Fatalf("newAuthedClient: %v", err)
	}
	if c.Session().AccessToken != "kc" {
		t.Fatalf("got %q, want kc", c.Session().AccessToken)
	}
	if c.Session().Ephemeral {
		t.Fatalf("keychain-loaded session should not be Ephemeral")
	}
	if c.Profile() != "default" {
		t.Fatalf("client profile = %q, want default", c.Profile())
	}
}

func TestNewAuthedClient_NothingAvailable(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "")
	gflags.Profile = "default"
	_, err := newAuthedClient()
	apiErr, _ := err.(*robinhood.APIError)
	if apiErr == nil || apiErr.Code != robinhood.CodeUnauthenticated {
		t.Fatalf("err = %v; want CodeUnauthenticated", err)
	}
}

func TestNewAuthedClient_InvalidProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "a")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "r")
	gflags.Profile = "../../etc/passwd"
	defer func() { gflags.Profile = "default" }()
	_, err := newAuthedClient()
	apiErr, _ := err.(*robinhood.APIError)
	if apiErr == nil || apiErr.Code != robinhood.CodeValidation {
		t.Fatalf("err = %v; want CodeValidation", err)
	}
}
