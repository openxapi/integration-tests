package streamstest

import (
	"testing"
	"time"
)

func waitForEvent[T any](t *testing.T, name string, ch <-chan T) (T, bool) {
	t.Helper()
	select {
	case ev := <-ch:
		return ev, true
	case <-time.After(eventWait()):
		t.Logf("timeout waiting %s (activity may be limited on live streams)", name)
		var zero T
		return zero, false
	}
}
