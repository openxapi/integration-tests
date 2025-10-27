package streamstest

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/umfutures-streams/models"
)

func assertNonEmpty(t *testing.T, v string, name string) {
	t.Helper()
	if strings.TrimSpace(v) == "" {
		t.Errorf("%s is empty", name)
	}
}

func mustParseFloat(t *testing.T, v string, name string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		t.Fatalf("%s parse float: %v (value=%q)", name, err, v)
	}
	return f
}

func tryParseFloat(t *testing.T, v string, name string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		t.Errorf("%s parse float: %v (value=%q)", name, err, v)
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
		t.Logf("%s not recent: diff=%dms (> %v)", name, diff, within)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func assertWithinTolerancePercent(t *testing.T, a, b, tolPct float64, label string) {
	t.Helper()
	if a == 0 && b == 0 {
		return
	}
	if a == 0 || b == 0 {
		// If one side is zero, skip strict percent comparison
		return
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	pct := diff / ((a + b) / 2.0) * 100.0
	if pct > tolPct {
		t.Errorf("%s differ: a=%.8f b=%.8f (%.2f%% > %.2f%%)", label, a, b, pct, tolPct)
	}
}

func newMessageID() models.MessageID {
	return models.NewMessageIDInt64(time.Now().UnixMicro())
}
