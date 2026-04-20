// Package keychain stores and retrieves secret blobs. It tries the OS keychain
// (macOS Keychain, GNOME libsecret, Windows Credential Manager) and falls back
// to a 0600 JSON file under the config dir when the OS keychain is unavailable
// (CI, headless Linux without a session bus, some containers).
package keychain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/zalando/go-keyring"
)

// ErrNotFound is returned when no entry exists for the given service+profile.
var ErrNotFound = errors.New("keychain: not found")

// maxBlobBytes caps file-backed reads so a corrupted/hostile session file
// can't blow out memory (Fix M).
const maxBlobBytes = 64 * 1024

// Store is the minimal contract Session uses.
type Store interface {
	Set(service, profile string, data []byte) error
	Get(service, profile string) ([]byte, error)
	Delete(service, profile string) error
}

// Open picks a backend. Precedence:
//  1. ROBINHOOD_KEYCHAIN_BACKEND=file — force file (for CI + tests)
//  2. ROBINHOOD_KEYCHAIN_BACKEND=system — force keyring (fail hard if unavailable)
//  3. default: try keyring with a one-shot, deadline-bounded probe; fall back to file
//     in $XDG_CONFIG_HOME/robinhood-cli/. If the config dir can't be created we
//     return an error — we never fall back to /tmp for secrets (Fix L).
func Open() (Store, error) {
	switch os.Getenv("ROBINHOOD_KEYCHAIN_BACKEND") {
	case "file":
		return fileStoreAtConfigDir()
	case "system":
		return &SystemStore{}, nil
	}
	if systemAvailable() {
		return &SystemStore{}, nil
	}
	return fileStoreAtConfigDir()
}

// fileStoreAtConfigDir returns a FileStore rooted at the application config dir.
// Returns an error if the config dir cannot be created — never falls back to
// os.TempDir() for secret storage (Fix L).
func fileStoreAtConfigDir() (*FileStore, error) {
	dir, err := config.EnsureConfigDir()
	if err != nil {
		return nil, fmt.Errorf("keychain: could not create config dir: %w", err)
	}
	return &FileStore{Dir: dir}, nil
}

var (
	probeOnce sync.Once
	probeRes  bool
)

// systemAvailable returns whether the OS keychain is usable. The probe runs
// at most once per process, with a 1500ms deadline; on headless Linux with
// no dbus session it short-circuits to false without probing (Fix J).
func systemAvailable() bool {
	probeOnce.Do(func() {
		if runtime.GOOS == "linux" && os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
			probeRes = false
			return
		}
		done := make(chan bool, 1)
		go func() { done <- probeSystem(&SystemStore{}) }()
		select {
		case ok := <-done:
			probeRes = ok
		case <-time.After(1500 * time.Millisecond):
			probeRes = false
		}
	})
	return probeRes
}

func probeSystem(s *SystemStore) bool {
	const probeService = "robinhood-cli-probe"
	const probeKey = "probe"
	if err := s.Set(probeService, probeKey, []byte("x")); err != nil {
		return false
	}
	defer func() { _ = s.Delete(probeService, probeKey) }()
	got, err := s.Get(probeService, probeKey)
	if err != nil || string(got) != "x" {
		return false
	}
	return true
}

// SystemStore uses the OS keychain via zalando/go-keyring.
type SystemStore struct{}

func (s *SystemStore) Set(service, profile string, data []byte) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	return keyring.Set(service, profile, string(data))
}

func (s *SystemStore) Get(service, profile string) ([]byte, error) {
	if err := config.ValidProfile(profile); err != nil {
		return nil, err
	}
	v, err := keyring.Get(service, profile)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return []byte(v), nil
}

func (s *SystemStore) Delete(service, profile string) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	err := keyring.Delete(service, profile)
	if err != nil && errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// FileStore writes session blobs to <Dir>/<service>-<profile>.json at 0600.
type FileStore struct {
	Dir string
}

func (f *FileStore) path(service, profile string) string {
	return filepath.Join(f.Dir, fmt.Sprintf("%s-%s.json", service, profile))
}

func (f *FileStore) Set(service, profile string, data []byte) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	if err := os.MkdirAll(f.Dir, 0o700); err != nil {
		return err
	}
	p := f.path(service, profile)
	tmp := p + ".tmp"
	// Best-effort cleanup of a prior aborted write so O_EXCL can succeed.
	_ = os.Remove(tmp)
	fh, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, werr := fh.Write(data); werr != nil {
		_ = fh.Close()
		_ = os.Remove(tmp)
		return werr
	}
	if cerr := fh.Close(); cerr != nil {
		_ = os.Remove(tmp)
		return cerr
	}
	if rerr := os.Rename(tmp, p); rerr != nil {
		_ = os.Remove(tmp)
		return rerr
	}
	return nil
}

func (f *FileStore) Get(service, profile string) ([]byte, error) {
	if err := config.ValidProfile(profile); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(f.path(service, profile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if len(data) > maxBlobBytes {
		return nil, fmt.Errorf("session file too large (%d bytes, cap %d)", len(data), maxBlobBytes)
	}
	return data, nil
}

func (f *FileStore) Delete(service, profile string) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	err := os.Remove(f.path(service, profile))
	if err != nil && os.IsNotExist(err) {
		return ErrNotFound
	}
	return err
}
