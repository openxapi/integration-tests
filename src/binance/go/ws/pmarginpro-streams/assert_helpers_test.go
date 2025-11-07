package streamstest

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func assertNonEmpty(t *testing.T, value, name string) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Errorf("%s is empty", name)
	}
}

func tryParseFloat(t *testing.T, value, name string) float64 {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		return 0
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Errorf("%s parse float: %v (value=%q)", name, err, value)
		return 0
	}
	return f
}

func tryParseInt64(t *testing.T, value, name string) int64 {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		return 0
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Errorf("%s parse int64: %v (value=%q)", name, err, value)
		return 0
	}
	return i
}

func assertRecentMs(t *testing.T, ts int64, within time.Duration, name string) {
	t.Helper()
	if ts <= 0 {
		t.Errorf("%s <= 0", name)
		return
	}
	now := time.Now().UnixMilli()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}
	if diff > within.Milliseconds() {
		t.Logf("%s outside expected freshness: %dms (> %v)", name, diff, within)
	}
}

func assertPositiveFloat(t *testing.T, value float64, name string) {
	t.Helper()
	if value <= 0 {
		t.Errorf("%s <= 0 (%.8f)", name, value)
	}
}
