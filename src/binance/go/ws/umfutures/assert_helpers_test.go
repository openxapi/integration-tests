package wstest

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func assertNonEmpty(t *testing.T, value string, name string) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Errorf("%s is empty", name)
	}
}

func mustParseFloat(t *testing.T, value string, name string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Fatalf("%s parse float: %v (value=%q)", name, err, value)
	}
	return f
}

func tryParseFloat(t *testing.T, value string, name string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Errorf("%s parse float: %v (value=%q)", name, err, value)
		return 0
	}
	return f
}

func assertRecentMs(t *testing.T, tsMs int64, within time.Duration, name string) {
	t.Helper()
	if tsMs <= 0 {
		t.Errorf("%s <= 0", name)
		return
	}
	now := time.Now().UnixMilli()
	diff := now - tsMs
	if diff < 0 {
		diff = -diff
	}
	if diff > within.Milliseconds() {
		t.Logf("%s diff=%dms (> %v)", name, diff, within)
	}
}
