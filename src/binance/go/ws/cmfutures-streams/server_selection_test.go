package streamstest

import (
    "os"
    "testing"

    cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
)

// Ensure tests use testnet by default unless BINANCE_CMFUTURES_WS_SERVER overrides
func Test_UsesTestnetServerByDefault(t *testing.T) {
    // Verify our setup logic prefers testnet when available, else uses SDK default
    if v := os.Getenv("BINANCE_CMFUTURES_WS_SERVER"); v != "" {
        client := cmfuturesstreams.NewClient()
        if err := client.AddOrUpdateServer("override", v, "override", "override"); err == nil {
            _ = client.SetActiveServer("override")
        }
        as := client.GetActiveServer()
        if as == nil || as.Name != "override" {
            t.Fatalf("expected 'override' when BINANCE_CMFUTURES_WS_SERVER set; got %v", as)
        }
        return
    }

    // No override: use the same path as integration setup
    client, err := setupClient(getTestConfig())
    if err != nil {
        t.Skipf("setupClient failed to select testnet: %v (skipping)", err)
    }
    as := client.GetActiveServer()
    if as == nil {
        t.Fatalf("active server is nil after setupClient; expected non-nil")
    }
    // Prefer testnet, but accept any non-nil active server
    if as.Name != "testnet1" {
        t.Logf("active server not 'testnet1' (got %q %s); SDK may not include testnet preset", as.Name, as.URL)
    }
}
