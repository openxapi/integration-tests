package streamstest

import (
	"bytes"
	"sync"
)

// unhandledCatcher captures SDK log entries containing "unhandled message:"
type unhandledCatcher struct {
	matches []string
	mu      sync.Mutex
}

func (c *unhandledCatcher) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("unhandled message:")) {
		c.mu.Lock()
		c.matches = append(c.matches, string(p))
		c.mu.Unlock()
	}
	return len(p), nil
}
