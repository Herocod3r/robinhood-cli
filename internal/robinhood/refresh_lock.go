package robinhood

import (
	"fmt"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/herocod3r/robinhood-cli/internal/config"
)

// WithRefreshLock runs fn while holding an exclusive lock on
// <ConfigDir>/session.lock. Concurrent `rh` invocations serialize here
// so only one performs the refresh at a time — without this, two
// processes refreshing in parallel can both burn the same refresh
// token and invalidate the other.
func WithRefreshLock(fn func() error) error {
	dir, err := config.EnsureConfigDir()
	if err != nil {
		return err
	}
	lock := flock.New(filepath.Join(dir, "session.lock"))
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("acquire refresh lock: %w", err)
	}
	defer func() { _ = lock.Unlock() }()
	return fn()
}
