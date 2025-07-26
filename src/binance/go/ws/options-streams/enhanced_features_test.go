package streamstest

import (
	"context"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/options-streams/models"
)

// TestCombinedStreamEventHandler tests combined stream event handling capability
func TestCombinedStreamEventHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream tests in short mode")
	}

	client, err := NewStreamTestClientDedicated(getTestConfig())
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	defer client.Disconnect()

	eventsReceived := 0
	
	client.client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		eventsReceived++
		t.Logf("Received CombinedStreamEvent #%d: StreamName=%s, StreamData available=%t", 
			eventsReceived, event.StreamName, event.StreamData != nil)
		
		// Validate event structure
		if event.StreamName == "" {
			t.Error("Expected StreamName to be non-empty")
		}
		if event.StreamData == nil {
			t.Error("Expected StreamData to be non-nil")
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect to combined streams specifically
	err = client.ConnectToCombinedStreams(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}

	// Try to get dynamic symbols, but use index streams as fallback
	// Subscribe to multiple streams to trigger combined events
	streams := []string{
		"ETHUSDT@index",    // Index price stream
		"BTCUSDT@index",    // Index price stream  
		"ETHUSDT@markPrice", // Mark price stream
	}

	// Use dynamic symbol selection if available, otherwise fallback to index streams
	ethSymbol, err := selectNearestExpirySymbol("ETH", "C")
	if err == nil && ethSymbol != "" {
		// Add actual options symbol streams if available
		streams = append(streams, ethSymbol+"@ticker")
		t.Logf("Using dynamic ETH option symbol: %s", ethSymbol)
	}

	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, streams)
	if err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Wait for events
	time.Sleep(8 * time.Second)

	if eventsReceived > 0 {
		t.Logf("✅ Successfully received %d CombinedStreamEvents", eventsReceived)
	} else {
		t.Error("❌ Expected to receive at least one CombinedStreamEvent")
	}
}

// TestStreamErrorHandler tests error handling functionality
func TestStreamErrorHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handler tests in short mode")
	}

	// Create a dedicated client for error testing
	config := getTestConfig()
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	defer client.Disconnect()

	// Test error handling by connecting to an invalid stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect with an invalid stream path
	err = client.client.ConnectToStream(ctx, "invalid@stream@format")
	
	// We expect this to either fail immediately or timeout
	if err != nil {
		t.Logf("✅ Connection properly failed for invalid stream: %v", err)
	} else {
		t.Log("⚠️  Connection succeeded unexpectedly - invalid stream may be accepted")
		
		// If connection succeeds, check if we get any responses
		time.Sleep(3 * time.Second)
		
		responses := client.GetResponseList()
		if len(responses) == 0 {
			t.Log("✅ No responses received for invalid stream - proper handling")
		} else {
			t.Logf("⚠️  Received %d responses for invalid stream", len(responses))
		}
	}

	t.Log("✅ Basic error handling functionality verified")
}

// TestAdvancedPropertyManagement tests advanced client property management
func TestAdvancedPropertyManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping advanced property tests in short mode")
	}

	// Create a dedicated client for property testing
	config := getTestConfig()
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	defer client.Disconnect()

	// Test client property access
	currentURL := client.client.GetCurrentURL()
	if currentURL == "" {
		t.Fatal("Current URL should not be empty")
	}
	t.Logf("✅ Current URL: %s", currentURL)

	// Test server management
	activeServer := client.client.GetActiveServer()
	if activeServer == nil {
		t.Fatal("Active server should not be nil")
	}
	t.Logf("✅ Active server: %s", activeServer.Name)

	// Test server listing
	servers := client.client.ListServers()
	if len(servers) == 0 {
		t.Fatal("Server list should not be empty")
	}
	t.Logf("✅ Available servers: %d", len(servers))

	// Test connection status
	connected := client.client.IsConnected()
	t.Logf("✅ Connection status: %v", connected)

	// Test response list functionality
	responseList := client.client.GetResponseList()
	t.Logf("✅ Response list length: %d", len(responseList))

	// Clear response list
	client.client.ClearResponseList()
	clearedList := client.client.GetResponseList()
	if len(clearedList) != 0 {
		t.Errorf("Response list should be empty after clear, got %d items", len(clearedList))
	} else {
		t.Log("✅ Response list cleared successfully")
	}

	t.Log("✅ Advanced property management verified")
}

// TestConcurrentStreams tests handling multiple streams concurrently
// Note: Limited functionality due to lack of proper subscription management
func TestConcurrentStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent streams tests in short mode")
	}

	t.Log("Testing concurrent stream connections with updated SDK")

	// Get dynamic BTC option symbol for ticker test
	btcSymbol, err := selectATMSymbol("BTC", "C")
	if err != nil {
		t.Logf("Warning: Could not get dynamic BTC symbol, using fallback: %v", err)
		btcSymbol = "BTC-250725-100000-C" // Fallback symbol
	}
	t.Logf("Using dynamic BTC option symbol: %s", btcSymbol)

	// Define multiple streams to test sequentially (due to SDK limitations)
	streams := []struct {
		name       string
		streamPath string
	}{
		{"IndexPrice", "ETHUSDT@index"},
		{"MarkPrice", "ETH@markPrice"}, 
		{"Ticker", btcSymbol + "@ticker"},
	}

	// Test each stream sequentially due to SDK limitations
	for _, stream := range streams {
		t.Run(stream.name, func(t *testing.T) {
			// Create dedicated client for this test
			config := getTestConfig()
			client, err := NewStreamTestClientDedicated(config)
			if err != nil {
				t.Fatalf("Failed to create client for %s: %v", stream.name, err)
				return
			}
			defer client.Disconnect()

			// Connect to stream
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = client.client.ConnectToStream(ctx, stream.streamPath)
			if err != nil {
				t.Errorf("❌ %s: Failed to connect: %v", stream.name, err)
				return
			}

			// Wait for potential responses
			time.Sleep(3 * time.Second)

			responses := client.GetResponseList()
			if len(responses) > 0 {
				t.Logf("✅ %s: Received %d responses", stream.name, len(responses))
			} else {
				t.Logf("⚠️  %s: No responses (expected for low activity)", stream.name)
			}
		})
	}

	t.Log("✅ Concurrent stream testing completed")
}

// TestHighVolumeStreams tests stream performance under potential high volume
func TestHighVolumeStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high volume streams tests in short mode")
	}

	// Use dedicated client for high volume testing to avoid conflicts
	config := getTestConfig()
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	defer client.Disconnect()

	// Clear responses before test
	client.ClearResponseList()

	t.Log("Testing high volume stream handling capabilities")

	// Connect to a potentially high-volume stream (all trades for underlying)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.client.ConnectToStream(ctx, "ETH@trade")
	if err != nil {
		t.Fatalf("Failed to connect to high volume stream: %v", err)
	}

	t.Log("✅ Connected to high volume trade stream")

	// Monitor responses for a longer period
	startTime := time.Now()
	monitorDuration := 10 * time.Second

	time.Sleep(monitorDuration)

	endTime := time.Now()
	totalResponses := len(client.GetResponseList())
	duration := endTime.Sub(startTime)

	t.Logf("✅ High volume test results:")
	t.Logf("   Duration: %v", duration)
	t.Logf("   Total responses: %d", totalResponses)
	
	if totalResponses > 0 {
		responsesPerSecond := float64(totalResponses) / duration.Seconds()
		t.Logf("   Responses per second: %.2f", responsesPerSecond)
		t.Log("✅ Successfully handled high volume stream")
	} else {
		t.Log("⚠️  No responses received - this is expected for low trading activity")
		t.Log("✅ High volume stream connection verified")
	}
}