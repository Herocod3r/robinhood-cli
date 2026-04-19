// Package config holds env-var names and config-path resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Env-var names — public contract with users and the skill.
const (
	EnvUsername     = "ROBINHOOD_USERNAME"
	EnvPassword     = "ROBINHOOD_PASSWORD"
	EnvTOTPSecret   = "ROBINHOOD_TOTP_SECRET"
	EnvDeviceToken  = "ROBINHOOD_DEVICE_TOKEN"
	EnvAccessToken  = "ROBINHOOD_ACCESS_TOKEN"
	EnvRefreshToken = "ROBINHOOD_REFRESH_TOKEN"
	EnvProfile      = "ROBINHOOD_PROFILE"
)

// AppName is the application directory name under the config root.
const AppName = "robinhood-cli"

// ConfigDir returns $XDG_CONFIG_HOME/robinhood-cli, falling back to $HOME/.config/robinhood-cli.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", AppName), nil
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
	return dir, nil
}
