package streamstest

import (
	"context"
	"testing"
	"time"
)

// TestConnection tests basic WebSocket connection functionality
func TestConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection tests in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	// Test connection status
	if !client.IsConnected() {
		t.Fatal("Client should be connected after setup")
	}

	t.Log("✅ Successfully established WebSocket connection to Binance Options Streams")

	// Test server info
	activeServer := client.client.GetActiveServer()
	if activeServer == nil {
		t.Fatal("No active server found")
	}

	t.Logf("✅ Active server: %s (%s)", activeServer.Name, activeServer.URL)
	t.Logf("   Title: %s", activeServer.Title)
	t.Logf("   Description: %s", activeServer.Description)

	// Test current URL
	currentURL := client.client.GetCurrentURL()
	if currentURL == "" {
		t.Fatal("Current URL should not be empty")
	}

	t.Logf("✅ Current server URL: %s", currentURL)

	// Test connection health
	if !client.client.IsConnected() {
		t.Fatal("Client should report as connected")
	}

	t.Log("✅ Connection health check passed")
}

// TestServerManagement tests server management functionality
func TestServerManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping server management tests in short mode")
	}

	// Create a dedicated client for server management tests (avoid affecting shared client)
	config := getTestConfig()
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}

	// Test listing servers
	servers := client.client.ListServers()
	if len(servers) == 0 {
		t.Fatal("No servers found")
	}

	t.Logf("✅ Found %d available servers:", len(servers))
	for name, server := range servers {
		t.Logf("   - %s: %s (Active: %v)", name, server.URL, server.Active)
	}

	// Test getting active server
	activeServer := client.client.GetActiveServer()
	if activeServer == nil {
		t.Fatal("No active server found")
	}

	t.Logf("✅ Active server: %s", activeServer.Name)

	// Test getting specific server
	server := client.client.GetServer(activeServer.Name)
	if server == nil {
		t.Fatalf("Failed to get server by name: %s", activeServer.Name)
	}

	if server.Name != activeServer.Name {
		t.Fatalf("Server name mismatch: expected %s, got %s", activeServer.Name, server.Name)
	}

	t.Logf("✅ Successfully retrieved server info for: %s", server.Name)

	// Test adding a custom server (client must be disconnected first)
	customURL := "wss://custom.binance.com/eoptionsws"
	err = client.client.AddServer("custom", customURL, "Custom Options Server", "Custom test server")
	if err != nil {
		t.Fatalf("Failed to add custom server: %v", err)
	}

	t.Log("✅ Successfully added custom server")

	// Verify custom server was added
	customServer := client.client.GetServer("custom")
	if customServer == nil {
		t.Fatal("Custom server not found after adding")
	}

	if customServer.URL != customURL {
		t.Fatalf("Custom server URL mismatch: expected %s, got %s", customURL, customServer.URL)
	}

	t.Log("✅ Custom server verification passed")

	// Test removing custom server
	err = client.client.RemoveServer("custom")
	if err != nil {
		t.Fatalf("Failed to remove custom server: %v", err)
	}

	t.Log("✅ Successfully removed custom server")

	// Verify custom server was removed
	removedServer := client.client.GetServer("custom")
	if removedServer != nil {
		t.Fatal("Custom server still exists after removal")
	}

	t.Log("✅ Custom server removal verification passed")

	// Test connection with valid server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	t.Log("✅ Successfully connected to server")

	// Test connection status
	if !client.IsConnected() {
		t.Fatal("Client should be connected")
	}

	t.Log("✅ Connection status verification passed")

	// Test graceful disconnection
	err = client.Disconnect()
	if err != nil {
		// Log the error but don't fail the test - some disconnect errors are expected
		t.Logf("Note: Error during disconnect (may be expected): %v", err)
	}

	t.Log("✅ Successfully disconnected from server")

	// Test connection status after disconnect
	if client.IsConnected() {
		t.Error("Client should not be connected after disconnect")
	}

	t.Log("✅ Disconnect status verification passed")
}

// TestConnectionResilience tests connection error handling and recovery
func TestConnectionResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping connection resilience tests in short mode")
	}

	// Create a dedicated client for resilience testing
	config := getTestConfig()
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	defer client.Disconnect()

	// Test connection with invalid URL
	invalidURL := "wss://invalid.binance.com/eoptionsws"
	err = client.client.AddOrUpdateServer("invalid", invalidURL, "Invalid Server", "Invalid test server")
	if err != nil {
		t.Fatalf("Failed to add invalid server: %v", err)
	}

	err = client.client.SetActiveServer("invalid")
	if err != nil {
		t.Fatalf("Failed to set invalid server as active: %v", err)
	}

	// Attempt connection with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connection should have failed with invalid URL")
	} else {
		t.Logf("✅ Connection properly failed with invalid URL: %v", err)
	}

	// Test connection status after failed connection
	if client.IsConnected() {
		t.Error("Client should not be connected after failed connection")
	}

	t.Log("✅ Connection failure handling verification passed")

	// Switch back to valid server
	err = client.client.SetActiveServer("mainnet1")
	if err != nil {
		t.Fatalf("Failed to switch back to valid server: %v", err)
	}

	// Test successful connection after failure
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to valid server after failure: %v", err)
	}

	t.Log("✅ Successfully recovered connection after failure")

	// Verify connection is working
	if !client.IsConnected() {
		t.Fatal("Client should be connected after recovery")
	}

	t.Log("✅ Connection recovery verification passed")
}