package streamstest

import (
	"encoding/json"
	"testing"
)

func logJSON(t *testing.T, label string, v interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: <marshal error: %v>", label, err)
		return
	}
	// Truncate very large payloads to keep logs readable
	const max = 4000
	if len(b) > max {
		t.Logf("%s: %s ... (truncated)", label, string(b[:max]))
		return
	}
	t.Logf("%s: %s", label, string(b))
}
