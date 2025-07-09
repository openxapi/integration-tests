package streamstest

import (
	"context"
	"testing"
	"time"

	spotstreams "github.com/openxapi/binance-go/ws/spot-streams"
)

// TestConnection tests basic WebSocket connection functionality
func TestConnection(t *testing.T) {
	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Verify connection status
	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// Test disconnection
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Verify disconnection
	if client.IsConnected() {
		t.Error("Client should be disconnected")
	}
}

// TestServerManagement tests server management functionality
func TestServerManagement(t *testing.T) {
	client := spotstreams.NewClient()

	// Test listing servers
	servers := client.ListServers()
	if len(servers) == 0 {
		t.Error("Should have at least one server")
	}

	// Test getting active server
	activeServer := client.GetActiveServer()
	if activeServer == nil {
		t.Error("Should have an active server")
	}

	t.Logf("Active server: %s (%s)", activeServer.Name, activeServer.URL)

	// Verify testnet server is available
	testnetServer := client.GetServer("testnet1")
	if testnetServer == nil {
		t.Error("Testnet server should exist")
	} else {
		t.Logf("Testnet server: %s (%s)", testnetServer.Name, testnetServer.URL)
	}

	// Test adding a new server
	testURL := "wss://test.example.com/ws"
	if err := client.AddServer("test", testURL, "Test Server", "Test server for unit tests"); err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Test getting the added server
	testServer := client.GetServer("test")
	if testServer == nil {
		t.Error("Test server should exist")
	}

	if testServer.URL != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, testServer.URL)
	}

	// Test removing the server
	if err := client.RemoveServer("test"); err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	// Verify removal
	if client.GetServer("test") != nil {
		t.Error("Test server should be removed")
	}
}

// TestConnectionTimeout tests connection timeout behavior
func TestConnectionTimeout(t *testing.T) {
	client := createTestClient(t)

	// Test with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected timeout error")
		client.Disconnect()
	}
}

// TestMultipleConnections tests behavior with multiple connection attempts
func TestMultipleConnections(t *testing.T) {
	client := createTestClient(t)

	ctx := context.Background()

	// First connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("First connection failed: %v", err)
	}

	// Second connection (should not fail)
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Second connection failed: %v", err)
	}

	// Verify still connected
	if !client.IsConnected() {
		t.Error("Client should still be connected")
	}

	// Cleanup
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

// TestConnectToSpecificServer tests connecting to a specific server
func TestConnectToSpecificServer(t *testing.T) {
	client := spotstreams.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to testnet1
	if err := client.ConnectToServer(ctx, "testnet1"); err != nil {
		t.Fatalf("Failed to connect to testnet1: %v", err)
	}

	// Verify active server
	activeServer := client.GetActiveServer()
	if activeServer == nil || activeServer.Name != "testnet1" {
		t.Error("Should be connected to testnet1")
	}

	// Cleanup
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

// TestConnectionRecovery tests connection recovery after network issues
func TestConnectionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection recovery test in short mode")
	}

	client := createTestClient(t)
	client.SetupEventHandlers()

	ctx := context.Background()

	// Connect
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to a stream
	if err := client.Subscribe(ctx, []string{"btcusdt@trade"}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Wait for some events
	time.Sleep(5 * time.Second)

	// Disconnect
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Reconnect
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}

	// Verify connection is working
	if !client.IsConnected() {
		t.Error("Client should be connected after recovery")
	}

	// Cleanup
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}