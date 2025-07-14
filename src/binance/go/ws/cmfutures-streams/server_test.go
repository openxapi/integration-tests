package streamstest

import (
	"testing"

	cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
)

// TestWebSocketAPIServers verifies that the new WebSocket API servers are properly configured
func TestWebSocketAPIServers(t *testing.T) {
	client := cmfuturesstreams.NewClient()

	// Test that all expected servers are available
	servers := client.ListServers()
	
	expectedServers := []string{
		"mainnet1",         // Market data mainnet
		"testnet1",         // Market data testnet  
		"userDataMainnet1", // User data streams mainnet
		"userDataTestnet1", // User data streams testnet
	}
	
	for _, expectedServer := range expectedServers {
		server, found := servers[expectedServer]
		if !found {
			t.Errorf("Expected server '%s' not found", expectedServer)
			continue
		}
		
		t.Logf("✅ Server %s: %s", expectedServer, server.URL)
	}

	// Test setting active server to testnet1
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Errorf("Failed to set testnet1 as active server: %v", err)
	}
	
	activeServer := client.GetActiveServer()
	if activeServer.Name != "testnet1" {
		t.Errorf("Expected active server to be 'testnet1', got '%s'", activeServer.Name)
	}
	
	t.Logf("✅ Successfully set testnet1 as active server")
	t.Logf("📊 Summary: All %d servers configured correctly", len(expectedServers))
	t.Logf("🔧 Using testnet1 for user data stream management commands")
}