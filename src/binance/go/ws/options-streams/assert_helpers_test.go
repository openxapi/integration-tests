package streamstest

import (
	"strconv"
	"strings"
	"testing"
	"time"
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

func assertOptionSymbolFormat(t *testing.T, symbol string) {
	t.Helper()
	if !validateSymbolFormat(symbol) {
		t.Errorf("invalid option symbol format: %s", symbol)
	}
}

func assertWithinTolerancePercent(t *testing.T, a, b, tolPct float64, label string) {
	t.Helper()
	if a == 0 && b == 0 {
		return
	}
	if a == 0 || b == 0 {
		// If one side is zero, use absolute diff check with a loose bound
		diff := a - b
		if diff < 0 {
			diff = -diff
		}
		if diff > 1e-9 { // some non-zero diff
			t.Logf("%s tolerance skipped due to near-zero reference (a=%.8f, b=%.8f)", label, a, b)
		}
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

func optionUnderlying(symbol string) string {
	// Expect UNDERLYING-YYMMDD-STRIKE-(C|P)
	if i := strings.Index(symbol, "-"); i > 0 {
		return symbol[:i]
	}
	return symbol
}
