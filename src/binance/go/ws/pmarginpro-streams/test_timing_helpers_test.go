package streamstest

import (
	"os"
	"strconv"
	"time"
)

func eventWait() time.Duration {
	if override := os.Getenv("EVENT_WAIT_SECS"); override != "" {
		if secs, err := strconv.Atoi(override); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return 20 * time.Second
}

func eventWaitLong() time.Duration {
	return eventWait() * 2
}
