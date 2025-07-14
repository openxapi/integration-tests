package streamstest

import (
	"context"
	"testing"
	"time"

	cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
)

// TestConnection tests basic connection functionality
func TestConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection tests in short mode")
	}

	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Verify connection
	if !client.IsConnected() {
		t.Error("Client reports not connected after successful connect")
	}

	t.Log("✅ Successfully connected to COIN-M futures streams")

	// Test disconnection
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Verify disconnection
	if client.IsConnected() {
		t.Error("Client reports connected after disconnect")
	}

	t.Log("✅ Successfully disconnected from COIN-M futures streams")
}

// TestConnectionTimeout tests connection with timeout
func TestConnectionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection timeout test in short mode")
	}

	client := createTestClient(t)

	// Use a very short timeout to test timeout behavior
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should timeout immediately
	err := client.Connect(ctx)
	if err == nil {
		client.Disconnect() // Clean up if somehow connected
		t.Error("Expected connection to timeout, but it succeeded")
	} else {
		t.Logf("✅ Connection properly timed out: %v", err)
	}
}

// TestMultipleConnections tests multiple connection attempts
func TestMultipleConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple connections test in short mode")
	}

	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect first time: %v", err)
	}

	// Second connection attempt should not fail
	if err := client.Connect(ctx); err != nil {
		t.Errorf("Second connection attempt failed: %v", err)
	}

	t.Log("✅ Multiple connection attempts handled correctly")

	// Clean up
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}

// TestServerManagement tests server management functionality
func TestServerManagement(t *testing.T) {
	client := cmfuturesstreams.NewClient()

	// Test listing servers
	servers := client.ListServers()
	if len(servers) == 0 {
		t.Error("No servers found in client")
	}

	t.Logf("✅ Found %d predefined servers", len(servers))

	// Test getting active server
	activeServer := client.GetActiveServer()
	if activeServer == nil {
		t.Error("No active server found")
	} else {
		t.Logf("✅ Active server: %s (%s)", activeServer.Name, activeServer.URL)
	}

	// Test getting specific server
	testnetServer := client.GetServer("testnet1")
	if testnetServer == nil {
		t.Error("Testnet server not found")
	} else {
		t.Logf("✅ Testnet server: %s", testnetServer.URL)
	}

	// Test adding a new server
	testServerURL := "wss://example.com/ws"
	err := client.AddServer("test-server", testServerURL, "Test Server", "Test server for testing")
	if err != nil {
		t.Errorf("Failed to add server: %v", err)
	} else {
		t.Log("✅ Successfully added test server")
	}

	// Test updating server
	updatedURL := "wss://updated.example.com/ws"
	err = client.UpdateServer("test-server", updatedURL, "Updated Test Server", "Updated test server")
	if err != nil {
		t.Errorf("Failed to update server: %v", err)
	} else {
		t.Log("✅ Successfully updated test server")
	}

	// Test AddOrUpdate server
	err = client.AddOrUpdateServer("test-server2", testServerURL, "Test Server 2", "Another test server")
	if err != nil {
		t.Errorf("Failed to add/update server: %v", err)
	} else {
		t.Log("✅ Successfully added/updated test server 2")
	}

	// Test removing server
	err = client.RemoveServer("test-server")
	if err != nil {
		t.Errorf("Failed to remove server: %v", err)
	} else {
		t.Log("✅ Successfully removed test server")
	}

	// Clean up
	client.RemoveServer("test-server2")

	// Test setting active server
	originalActive := client.GetActiveServer()
	err = client.SetActiveServer("testnet1")
	if err != nil {
		t.Errorf("Failed to set active server: %v", err)
	} else {
		newActive := client.GetActiveServer()
		if newActive.Name != "testnet1" {
			t.Errorf("Active server not set correctly: expected testnet1, got %s", newActive.Name)
		} else {
			t.Log("✅ Successfully set active server")
		}
	}

	// Restore original active server
	if originalActive != nil {
		client.SetActiveServer(originalActive.Name)
	}
}

// TestConnectToServer tests connecting to specific servers
func TestConnectToServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connect to server test in short mode")
	}

	client := cmfuturesstreams.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connecting to testnet server
	if err := client.ConnectToServer(ctx, "testnet1"); err != nil {
		t.Fatalf("Failed to connect to testnet1: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client not connected after ConnectToServer")
	}

	activeServer := client.GetActiveServer()
	if activeServer == nil || activeServer.Name != "testnet1" {
		t.Error("Active server not set correctly after ConnectToServer")
	}

	t.Log("✅ Successfully connected to specific server")

	// Clean up
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}

// TestConnectionRecovery tests connection recovery scenarios
func TestConnectionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection recovery test in short mode")
	}

	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initial connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect initially: %v", err)
	}

	// Force disconnect
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Reconnect
	if err := client.Connect(ctx); err != nil {
		t.Errorf("Failed to reconnect: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client not connected after reconnection")
	}

	t.Log("✅ Connection recovery successful")

	// Clean up
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}