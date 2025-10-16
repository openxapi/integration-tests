package streamstest

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

func eventWaitLong() time.Duration {
	// Roughly double the base wait
	return eventWait() * 2
}
