package systray

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Quit and the platform's own close handling can both call runSystrayExit,
// from different goroutines; the exit callback must still run exactly once.
func TestRunSystrayExitCallsOnExitOnce(t *testing.T) {
	var calls atomic.Int32
	exitMu.Lock()
	systrayExit = func() { calls.Add(1) }
	systrayExitCalled = false
	exitMu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runSystrayExit()
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("onExit called %d times, want 1", got)
	}
}
