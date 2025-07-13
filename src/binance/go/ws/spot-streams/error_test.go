package streamstest

import (
	"context"
	"testing"
	"time"
)

// TestErrorHandling tests error handling in stream operations
func TestErrorHandling(t *testing.T) {
	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test subscribing to malformed stream
	invalidStreams := []string{
		"invalid_stream",
		"btcusdt@invalid_type",
		"@trade",
		"btcusdt@",
		"",
	}

	for _, invalidStream := range invalidStreams {
		t.Run("invalid_stream_"+invalidStream, func(t *testing.T) {
			if invalidStream == "" {
				return // Skip empty stream test
			}
			
			// Check if underlying client is connected before attempting subscription
			if !client.client.IsConnected() {
				t.Logf("Client disconnected, reconnecting before testing: %s", invalidStream)
				
				// Force wrapper state to disconnected to allow reconnection
				client.mu.Lock()
				client.connected = false
				client.mu.Unlock()
				
				// Disconnect first to clean up any stale connections
				if err := client.Disconnect(); err != nil {
					t.Logf("Warning: Failed to disconnect client: %v", err)
				}
				
				// Create a new context with timeout for reconnection
				reconnectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				
				if err := client.Connect(reconnectCtx); err != nil {
					t.Fatalf("Failed to reconnect client: %v", err)
				}
				
				// Set up event handlers again after reconnection
				client.SetupEventHandlers()
				
				// Wait a moment for connection to stabilize
				time.Sleep(500 * time.Millisecond)
			}
			
			err := client.Subscribe(ctx, []string{invalidStream})
			// The subscription might not fail immediately at the client level
			// but we should check for error events
			if err != nil {
				t.Logf("Expected error for invalid stream '%s': %v", invalidStream, err)
			}

			// Wait for possible error events
			time.Sleep(2 * time.Second)

			// Check for error events
			errorEvents := client.GetEventsByType("error")
			if len(errorEvents) > 0 {
				t.Logf("Received %d error events for invalid stream '%s'", len(errorEvents), invalidStream)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestInvalidStreamNames tests various invalid stream name formats
func TestInvalidStreamNames(t *testing.T) {
	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test various invalid stream formats
	invalidStreamFormats := []string{
		"BTCUSDT@trade",        // Uppercase symbol
		"btcusdt@TRADE",        // Uppercase stream type
		"btc_usdt@trade",       // Invalid symbol format
		"btcusdt@trade@extra",  // Extra parts
		"btcusdt trade",        // Missing @ (causes connection close)
		"btcusdt@@trade",       // Double @
		"btcusdt@kline_",       // Missing interval
		"btcusdt@kline_999m",   // Invalid interval
		"btcusdt@depth999",     // Invalid depth level
	}

	for _, invalidStream := range invalidStreamFormats {
		t.Run("format_test_"+invalidStream, func(t *testing.T) {
			// Check if underlying client is connected before attempting subscription
			if !client.client.IsConnected() {
				t.Logf("Client disconnected, reconnecting before testing: %s", invalidStream)
				
				// Force wrapper state to disconnected to allow reconnection
				client.mu.Lock()
				client.connected = false
				client.mu.Unlock()
				
				// Disconnect first to clean up any stale connections
				if err := client.Disconnect(); err != nil {
					t.Logf("Warning: Failed to disconnect client: %v", err)
				}
				
				// Create a new context with timeout for reconnection
				reconnectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				
				if err := client.Connect(reconnectCtx); err != nil {
					t.Fatalf("Failed to reconnect client: %v", err)
				}
				
				// Set up event handlers again after reconnection
				client.SetupEventHandlers()
				
				// Wait a moment for connection to stabilize
				time.Sleep(500 * time.Millisecond)
			}

			err := client.Subscribe(ctx, []string{invalidStream})
			
			// Log the attempt
			t.Logf("Attempting to subscribe to invalid stream: %s", invalidStream)
			
			if err != nil {
				t.Logf("Client-level error for '%s': %v", invalidStream, err)
			}

			// Wait for server response
			time.Sleep(2 * time.Second)

			// Check for error events
			errorEvents := client.GetEventsByType("error")
			if len(errorEvents) > 0 {
				t.Logf("Received %d error events for invalid format '%s'", len(errorEvents), invalidStream)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestConnectionErrors tests error handling during connection issues
func TestConnectionErrors(t *testing.T) {
	client := createTestClient(t)
	
	ctx := context.Background()

	// Test operations without connection
	err := client.Subscribe(ctx, []string{"btcusdt@trade"})
	if err == nil {
		t.Error("Expected error when subscribing without connection")
	}

	err = client.Unsubscribe(ctx, []string{"btcusdt@trade"})
	if err == nil {
		t.Error("Expected error when unsubscribing without connection")
	}

	err = client.ListSubscriptions(ctx)
	if err == nil {
		t.Error("Expected error when listing subscriptions without connection")
	}
}

// TestUnsubscribeNonExistentStream tests unsubscribing from non-existent streams
func TestUnsubscribeNonExistentStream(t *testing.T) {
	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try to unsubscribe from a stream that was never subscribed
	nonExistentStream := "btcusdt@trade"
	err := client.Unsubscribe(ctx, []string{nonExistentStream})
	
	// This might not fail at the client level, but server might send error
	if err != nil {
		t.Logf("Client-level error for unsubscribing non-existent stream: %v", err)
	}

	// Wait for server response
	time.Sleep(2 * time.Second)

	// Check for error events
	errorEvents := client.GetEventsByType("error")
	if len(errorEvents) > 0 {
		t.Logf("Received %d error events for unsubscribing non-existent stream", len(errorEvents))
	}
}

// TestEmptyStreamList tests subscribing/unsubscribing with empty stream list
func TestEmptyStreamList(t *testing.T) {
	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test subscribing to empty stream list
	err := client.Subscribe(ctx, []string{})
	if err != nil {
		t.Logf("Error subscribing to empty stream list: %v", err)
	}

	// Test unsubscribing from empty stream list
	err = client.Unsubscribe(ctx, []string{})
	if err != nil {
		t.Logf("Error unsubscribing from empty stream list: %v", err)
	}
}

// TestMaxStreamLimits tests behavior when approaching stream limits
func TestMaxStreamLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping max stream limits test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Generate a large number of streams (approaching Binance limits)
	symbols := []string{"btcusdt", "ethusdt", "adausdt", "dotusdt", "linkusdt"}
	streamTypes := []string{"trade", "depth5", "miniTicker", "bookTicker"}
	
	var streams []string
	for _, symbol := range symbols {
		for _, streamType := range streamTypes {
			streams = append(streams, symbol+"@"+streamType)
		}
	}

	// Add kline streams with different intervals
	intervals := []string{"1m", "5m", "15m", "1h"}
	for _, symbol := range symbols {
		for _, interval := range intervals {
			streams = append(streams, symbol+"@kline_"+interval)
		}
	}

	t.Logf("Attempting to subscribe to %d streams", len(streams))

	// Subscribe to all streams
	err := client.Subscribe(ctx, streams)
	if err != nil {
		t.Logf("Error subscribing to many streams: %v", err)
	}

	// Wait for server response
	time.Sleep(5 * time.Second)

	// Check for error events
	errorEvents := client.GetEventsByType("error")
	if len(errorEvents) > 0 {
		t.Logf("Received %d error events for many streams", len(errorEvents))
	}

	// Check how many streams are actually active
	activeStreams := client.GetActiveStreams()
	t.Logf("Successfully subscribed to %d out of %d streams", len(activeStreams), len(streams))

	// Cleanup
	if len(activeStreams) > 0 {
		if err := client.Unsubscribe(ctx, activeStreams); err != nil {
			t.Logf("Error unsubscribing from active streams: %v", err)
		}
	}
}

// TestReconnectionAfterError tests behavior after connection errors
func TestReconnectionAfterError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reconnection test in short mode")
	}

	client := setupAndConnectClient(t)

	ctx := context.Background()

	// Subscribe to a stream
	stream := "btcusdt@trade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Wait for some events
	time.Sleep(3 * time.Second)

	// Force disconnect
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Try to subscribe while disconnected (should fail)
	err := client.Subscribe(ctx, []string{"ethusdt@trade"})
	if err == nil {
		t.Error("Expected error when subscribing while disconnected")
	}

	// Reconnect
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}

	// Set up event handlers again
	client.SetupEventHandlers()

	// Try to subscribe again (should work)
	if err := client.Subscribe(ctx, []string{"ethusdt@trade"}); err != nil {
		t.Fatalf("Failed to subscribe after reconnection: %v", err)
	}

	// Wait for events
	time.Sleep(3 * time.Second)

	// Verify we're receiving events
	events := client.GetEventsReceived()
	if len(events) == 0 {
		t.Error("No events received after reconnection")
	}

	// Cleanup
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

// TestConcurrentSubscriptions tests concurrent subscription operations
func TestConcurrentSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent subscriptions test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test concurrent subscriptions
	symbols := []string{"btcusdt", "ethusdt", "adausdt"}
	streamTypes := []string{"trade", "depth5", "miniTicker"}

	var streams []string
	for _, symbol := range symbols {
		for _, streamType := range streamTypes {
			streams = append(streams, symbol+"@"+streamType)
		}
	}

	// Subscribe to streams with synchronization to avoid concurrent writes
	errChan := make(chan error, len(streams))
	semaphore := make(chan struct{}, 1) // Limit to 1 concurrent subscription
	
	for _, stream := range streams {
		go func(s string) {
			semaphore <- struct{}{} // Acquire lock
			defer func() { <-semaphore }() // Release lock
			time.Sleep(100 * time.Millisecond) // Small delay between subscriptions
			errChan <- client.Subscribe(ctx, []string{s})
		}(stream)
	}

	// Collect errors
	var errors []error
	for i := 0; i < len(streams); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Logf("Received %d errors from concurrent subscriptions", len(errors))
		for _, err := range errors {
			t.Logf("Error: %v", err)
		}
	}

	// Wait for events
	time.Sleep(5 * time.Second)

	// Check active streams
	activeStreams := client.GetActiveStreams()
	t.Logf("Active streams after concurrent subscription: %d", len(activeStreams))

	// Cleanup
	if len(activeStreams) > 0 {
		if err := client.Unsubscribe(ctx, activeStreams); err != nil {
			t.Logf("Error unsubscribing: %v", err)
		}
	}
}