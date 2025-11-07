package streamstest

import (
	"encoding/json"
	"testing"
)

func logJSON(t *testing.T, label string, payload interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Logf("%s: <marshal error: %v>", label, err)
		return
	}
	const maxLen = 4000
	if len(data) > maxLen {
		t.Logf("%s: %s ... (truncated)", label, string(data[:maxLen]))
		return
	}
	t.Logf("%s: %s", label, string(data))
}
