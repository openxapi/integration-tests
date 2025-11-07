package wstest

import (
	"encoding/json"
	"testing"
)

func logJSON(t *testing.T, label string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: <marshal error: %v>", label, err)
		return
	}
	const maxBytes = 4000
	if len(data) > maxBytes {
		t.Logf("%s: %s … (truncated)", label, string(data[:maxBytes]))
		return
	}
	t.Logf("%s: %s", label, string(data))
}

func logRequestOnFailure(t testing.TB, label string, req interface{}) {
	t.Helper()
	if req == nil {
		return
	}
	t.Cleanup(func() {
		type failed interface{ Failed() bool }
		if ft, ok := t.(failed); ok && ft.Failed() {
			if tt, ok2 := t.(*testing.T); ok2 {
				logJSON(tt, label, req)
			}
		}
	})
}
