package streamstest

import (
	"context"
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
	const max = 4000
	if len(b) > max {
		t.Logf("%s: %s ... (truncated)", label, string(b[:max]))
		return
	}
	t.Logf("%s: %s", label, string(b))
}

// ackLogger returns a generic request callback that logs the payload and any RPC error.
func ackLogger[T any](t *testing.T, label string) func(context.Context, *T, error) error {
	t.Helper()
	return func(ctx context.Context, resp *T, rpcErr error) error {
		if rpcErr != nil {
			t.Logf("%s error: %v", label, rpcErr)
			return nil
		}
		if resp != nil {
			logJSON(t, label, resp)
		}
		return nil
	}
}
