package streamstest

import (
	"testing"
)

// TestGetSymbols tests getting symbols from the REST API
func TestGetSymbols(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symbol lookup test in short mode")
	}

	t.Log("=== Testing Symbol Retrieval ===")
	
	// Log all available symbols for debugging
	logAvailableSymbols(t)
	
	// Get test symbols
	testSymbols, err := getTestSymbols(t)
	if err != nil {
		t.Fatalf("Failed to get test symbols: %v", err)
	}
	
	t.Log("=== Test Symbols Found ===")
	for key, symbol := range testSymbols {
		if symbol != "" {
			t.Logf("%s: %s", key, symbol)
		} else {
			t.Logf("%s: NOT FOUND", key)
		}
	}
	
	// Verify we have at least BTC perpetual
	if testSymbols["btc_perp"] == "" {
		t.Error("No BTC perpetual symbol found")
	} else {
		t.Logf("✅ BTC perpetual symbol: %s", testSymbols["btc_perp"])
	}
}