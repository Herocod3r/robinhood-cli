// Package config holds env-var names and config-path resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Env-var names — public contract with users and the skill.
// gosec G101 fires on names containing "SECRET"/"TOKEN"; these are env-var
// identifiers, not embedded credentials.
const (
	EnvUsername     = "ROBINHOOD_USERNAME"
	EnvPassword     = "ROBINHOOD_PASSWORD"      //nolint:gosec // env-var name, not a credential
	EnvTOTPSecret   = "ROBINHOOD_TOTP_SECRET"   //nolint:gosec // env-var name, not a credential
	EnvDeviceToken  = "ROBINHOOD_DEVICE_TOKEN"  //nolint:gosec // env-var name, not a credential
	EnvAccessToken  = "ROBINHOOD_ACCESS_TOKEN"  //nolint:gosec // env-var name, not a credential
	EnvRefreshToken = "ROBINHOOD_REFRESH_TOKEN" //nolint:gosec // env-var name, not a credential
	EnvProfile      = "ROBINHOOD_PROFILE"
)

// AppName is the application directory name under the config root.
const AppName = "robinhood-cli"

// ConfigDir returns $XDG_CONFIG_HOME/robinhood-cli, falling back to $HOME/.config/robinhood-cli.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && filepath.IsAbs(xdg) {
		return filepath.Join(xdg, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", AppName), nil
}

// profileRe restricts profile names to safe characters for use as filesystem paths
// (path-traversal guard — Fix K).
var profileRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// ValidProfile reports whether name is a safe profile identifier.
// Returns an error with a human-readable message when it is not.
func ValidProfile(name string) error {
	if !profileRe.MatchString(name) {
		return fmt.Errorf("invalid profile name: %q (allowed: [A-Za-z0-9_-], 1-64 chars)", name)
	}
	return nil
}

// EnsureConfigDir creates the config dir with 0700 perms if missing.
func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", fmt.Errorf("chmod %s: %w", dir, err)
	}
	return dir, nil
}
