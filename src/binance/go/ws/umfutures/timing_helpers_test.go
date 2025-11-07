package wstest

import (
	"os"
	"strconv"
	"time"
)

func eventWait() time.Duration {
	if s := os.Getenv("EVENT_WAIT_SECS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return time.Duration(v) * time.Second
		}
	}
	return 20 * time.Second
}

func eventWaitLong() time.Duration { return eventWait() * 2 }

// throttleWS sleeps briefly to avoid hitting WS request rate limits.
// Override with WS_THROTTLE_MS env var (default: 300ms).
func throttleWS() {
	d := 300 * time.Millisecond
	if s := os.Getenv("WS_THROTTLE_MS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			d = time.Duration(v) * time.Millisecond
		}
	}
	if d > 0 {
		time.Sleep(d)
	}
}
