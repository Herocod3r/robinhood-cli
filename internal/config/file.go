package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// File is the on-disk config at $XDG_CONFIG_HOME/robinhood-cli/config.yaml.
type File struct {
	DefaultProfile string             `yaml:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

// Profile is a per-account block.
type Profile struct {
	AccountNumber string `yaml:"account_number,omitempty"`
}

// ConfigFilePath returns the absolute path to config.yaml (never creates it).
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadFile reads config.yaml if present, or returns zero-value defaults.
// Missing file is NOT an error — a fresh install has no config.
// Profile keys are validated via ValidProfile (Fix K).
func LoadFile() (*File, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}
	f := &File{DefaultProfile: "default", Profiles: map[string]Profile{"default": {}}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return f, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.DefaultProfile == "" {
		f.DefaultProfile = "default"
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{"default": {}}
	}
	for k := range f.Profiles {
		if verr := ValidProfile(k); verr != nil {
			return nil, fmt.Errorf("%s: %w", path, verr)
		}
	}
	if f.DefaultProfile != "" {
		if verr := ValidProfile(f.DefaultProfile); verr != nil {
			return nil, fmt.Errorf("%s: default_profile: %w", path, verr)
		}
	}
	return f, nil
}

// SaveFile writes the config atomically at 0600. Creates the config dir with 0700 if missing.
// Uses O_EXCL for the temp file to prevent TOCTOU symlink races (Fix L).
func SaveFile(f *File) error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	if f.DefaultProfile != "" {
		if verr := ValidProfile(f.DefaultProfile); verr != nil {
			return verr
		}
	}
	for k := range f.Profiles {
		if verr := ValidProfile(k); verr != nil {
			return verr
		}
	}
	out, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	final := filepath.Join(dir, "config.yaml")
	tmp := final + ".tmp"
	// Clean up any leftover tmp from a prior aborted write.
	_ = os.Remove(tmp)
	fh, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, werr := fh.Write(out); werr != nil {
		_ = fh.Close()
		_ = os.Remove(tmp)
		return werr
	}
	if cerr := fh.Close(); cerr != nil {
		_ = os.Remove(tmp)
		return cerr
	}
	if rerr := os.Rename(tmp, final); rerr != nil {
		_ = os.Remove(tmp)
		return rerr
	}
	return nil
}
