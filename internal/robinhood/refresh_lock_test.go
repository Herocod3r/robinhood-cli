package robinhood

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRefreshLock_Serializes(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var active int32
	var maxSeen int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := WithRefreshLock(func() error {
				n := atomic.AddInt32(&active, 1)
				for {
					m := atomic.LoadInt32(&maxSeen)
					if n <= m || atomic.CompareAndSwapInt32(&maxSeen, m, n) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&active, -1)
				return nil
			})
			if err != nil {
				t.Errorf("WithRefreshLock: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := atomic.LoadInt32(&maxSeen); got != 1 {
		t.Fatalf("max concurrent critical sections = %d, want 1", got)
	}
}
