package streamstest

import (
    "os"
    "testing"
)

// Test_UsesTestnetServerByDefault ensures integration tests use testnet server 'testnet1'
// unless explicitly overridden by BINANCE_UMFUTURES_WS_SERVER.
func Test_UsesTestnetServerByDefault(t *testing.T) {
    cfg := getTestConfig()
    stc, err := NewStreamTestClientDedicated(cfg)
    if err != nil { t.Fatalf("failed to create client: %v", err) }
    as := stc.client.GetActiveServer()
    if as == nil { t.Fatalf("active server is nil; expected 'testnet1' or 'override'") }

    if os.Getenv("BINANCE_UMFUTURES_WS_SERVER") != "" {
        if as.Name != "override" {
            t.Fatalf("expected active server name 'override' when BINANCE_UMFUTURES_WS_SERVER is set, got %q (url=%s)", as.Name, as.URL)
        }
    } else {
        if as.Name != "testnet1" {
            t.Fatalf("expected active server name 'testnet1' by default, got %q (url=%s)", as.Name, as.URL)
        }
    }
}

