package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestLogout_ClearsKeychain(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

	in := &robinhood.Session{Username: "u", AccessToken: "a", RefreshToken: "r", DeviceToken: "d", ExpiresAt: time.Now().Add(time.Hour)}
	if err := in.SaveToKeychain("default"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RunLogout(LogoutOpts{Profile: "default", Out: &out}); err != nil {
		t.Fatalf("RunLogout: %v", err)
	}
	if _, err := robinhood.LoadFromKeychain("default"); err == nil {
		t.Fatalf("expected LoadFromKeychain to fail after logout")
	}
}

func TestLogout_AllProfiles(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

	// Seed config.yaml with both profiles so RunLogout(AllProfiles) iterates them.
	f := &config.File{
		DefaultProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {},
			"work":    {},
		},
	}
	if err := config.SaveFile(f); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{"default", "work"} {
		s := &robinhood.Session{Username: p, AccessToken: "a" + p, RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
		if err := s.SaveToKeychain(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := RunLogout(LogoutOpts{AllProfiles: true, Out: &bytes.Buffer{}}); err != nil {
		t.Fatalf("RunLogout: %v", err)
	}
	for _, p := range []string{"default", "work"} {
		if _, err := robinhood.LoadFromKeychain(p); err == nil {
			t.Fatalf("expected no session for profile %q", p)
		}
	}
}

func TestLogout_InvalidProfileRejected(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	err := RunLogout(LogoutOpts{Profile: "../../etc/passwd", Out: &bytes.Buffer{}})
	apiErr, _ := err.(*robinhood.APIError)
	if apiErr == nil || apiErr.Code != robinhood.CodeValidation {
		t.Fatalf("err = %v, want CodeValidation", err)
	}
}
