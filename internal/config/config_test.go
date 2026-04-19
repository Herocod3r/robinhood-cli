package config

import (
	"path/filepath"
	"testing"
)

func TestConfigDir_UsesHome(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")
	t.Setenv("XDG_CONFIG_HOME", "")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/testhome", ".config", "robinhood-cli")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestConfigDir_UsesXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/xdg", "robinhood-cli")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestEnvNamesAreStable(t *testing.T) {
	// These names are a public contract; changing them breaks users.
	if EnvAccessToken != "ROBINHOOD_ACCESS_TOKEN" {
		t.Error("EnvAccessToken changed")
	}
	if EnvRefreshToken != "ROBINHOOD_REFRESH_TOKEN" {
		t.Error("EnvRefreshToken changed")
	}
	if EnvDeviceToken != "ROBINHOOD_DEVICE_TOKEN" {
		t.Error("EnvDeviceToken changed")
	}
}
