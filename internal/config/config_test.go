package config

import (
	"os"
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

func TestConfigDir_IgnoresRelativeXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "tmp-relative")
	t.Setenv("HOME", "/tmp/testhome")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/testhome", ".config", "robinhood-cli")
	if dir != want {
		t.Errorf("relative XDG should be ignored: got %q, want %q", dir, want)
	}
}

func TestEnsureConfigDir_TightensExistingPerms(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	preexisting := filepath.Join(home, ".config", "robinhood-cli")
	if err := os.MkdirAll(preexisting, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureConfigDir(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(preexisting)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("perms = %o, want 0o700", perm)
	}
}

func TestEnvNamesAreStable(t *testing.T) {
	cases := map[string]string{
		"EnvUsername":     EnvUsername,
		"EnvPassword":     EnvPassword,
		"EnvTOTPSecret":   EnvTOTPSecret,
		"EnvDeviceToken":  EnvDeviceToken,
		"EnvAccessToken":  EnvAccessToken,
		"EnvRefreshToken": EnvRefreshToken,
		"EnvProfile":      EnvProfile,
	}
	want := map[string]string{
		"EnvUsername":     "ROBINHOOD_USERNAME",
		"EnvPassword":     "ROBINHOOD_PASSWORD",
		"EnvTOTPSecret":   "ROBINHOOD_TOTP_SECRET",
		"EnvDeviceToken":  "ROBINHOOD_DEVICE_TOKEN",
		"EnvAccessToken":  "ROBINHOOD_ACCESS_TOKEN",
		"EnvRefreshToken": "ROBINHOOD_REFRESH_TOKEN",
		"EnvProfile":      "ROBINHOOD_PROFILE",
	}
	for k, got := range cases {
		if got != want[k] {
			t.Errorf("%s = %q, want %q", k, got, want[k])
		}
	}
}
