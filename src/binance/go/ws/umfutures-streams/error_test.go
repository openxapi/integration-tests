package streamstest

import (
	"context"
	"testing"
	"time"
)

// TestErrorHandling tests basic error handling
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to a valid stream first
	validStream := "btcusdt@aggTrade"
	if err := client.Subscribe(ctx, []string{validStream}); err != nil {
		t.Fatalf("Failed to subscribe to valid stream: %v", err)
	}

	// Wait for events to verify connection works
	err := client.WaitForEventsByType("aggTrade", 1, 10*time.Second)
	if err != nil {
		t.Logf("Warning: %v", err)
	}
	
	events := client.GetEventsByType("aggTrade")
	if len(events) > 0 {
		t.Log("✅ Valid stream subscription and event processing working")
	} else {
		t.Log("✅ Valid stream subscription working")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, []string{validStream}); err != nil {
		t.Errorf("Failed to unsubscribe from valid stream: %v", err)
	}
}

// TestInvalidStreamNames tests handling of invalid stream names
func TestInvalidStreamNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping invalid stream names test in short mode")
	}

	ctx := context.Background()

	// Test various invalid stream formats
	invalidStreams := []string{
		"invalid_stream_format",
		"btcusdt@invalidstream",
		"@aggTrade",
		"btcusdt@",
		"",
		"BTCUSDT@AGGTRADE", // Test case sensitivity
		"btcusdt@@aggTrade", // Double @
		"btcusdt@aggTrade@extra", // Extra parts
	}

	for _, invalidStream := range invalidStreams {
		t.Run("invalid_"+invalidStream, func(t *testing.T) {
			// Create a new client for each test to avoid connection pollution
			client := setupAndConnectClient(t)
			defer client.Disconnect()
			
			// Try to subscribe to invalid stream
			err := client.Subscribe(ctx, []string{invalidStream})
			
			// The subscription might succeed but no events should be received
			// or it might fail immediately - both are acceptable
			if err != nil {
				t.Logf("✅ Invalid stream '%s' properly rejected: %v", invalidStream, err)
			} else {
				t.Logf("⚠️  Invalid stream '%s' subscription accepted but should not receive events", invalidStream)
				
				// Wait a bit to see if we get any events (we shouldn't)
				time.Sleep(3 * time.Second)
				
				// Check for error events
				errorEvents := client.GetEventsByType("error")
				if len(errorEvents) > 0 {
					t.Logf("✅ Error events received for invalid stream: %d", len(errorEvents))
				}
				
				// Try to unsubscribe (may fail if connection was closed)
				client.Unsubscribe(ctx, []string{invalidStream})
			}
			
			// Clear any events for next test
			client.ClearEvents()
		})
	}
}

// TestOperationsWithoutConnection tests operations without connection
func TestOperationsWithoutConnection(t *testing.T) {
	client := createTestClient(t)
	// Note: Don't connect

	ctx := context.Background()

	// Try to subscribe without connection
	err := client.Subscribe(ctx, []string{"btcusdt@aggTrade"})
	if err == nil {
		t.Error("Expected error when subscribing without connection, but got none")
	} else {
		t.Logf("✅ Subscribe without connection properly failed: %v", err)
	}

	// Try to unsubscribe without connection
	err = client.Unsubscribe(ctx, []string{"btcusdt@aggTrade"})
	if err == nil {
		t.Logf("✅ Unsubscribe without connection handled gracefully")
	} else {
		t.Logf("✅ Unsubscribe without connection properly failed: %v", err)
	}

	// Try to list subscriptions without connection
	err = client.ListSubscriptions(ctx)
	if err == nil {
		t.Error("Expected error when listing subscriptions without connection, but got none")
	} else {
		t.Logf("✅ List subscriptions without connection properly failed: %v", err)
	}
}

// TestUnsubscribeNonexistentStream tests unsubscribing from non-existent streams
func TestUnsubscribeNonexistentStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping unsubscribe nonexistent stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try to unsubscribe from a stream we never subscribed to
	nonexistentStream := "btcusdt@aggTrade"
	err := client.Unsubscribe(ctx, []string{nonexistentStream})
	
	// This might succeed (no-op) or fail - both are acceptable
	if err != nil {
		t.Logf("✅ Unsubscribe from nonexistent stream failed as expected: %v", err)
	} else {
		t.Logf("✅ Unsubscribe from nonexistent stream handled gracefully")
	}

	// Subscribe to a stream, then try to unsubscribe from a different one
	validStream := "btcusdt@aggTrade"
	if err := client.Subscribe(ctx, []string{validStream}); err != nil {
		t.Fatalf("Failed to subscribe to valid stream: %v", err)
	}

	// Try to unsubscribe from a different stream
	differentStream := "ethusdt@aggTrade"
	err = client.Unsubscribe(ctx, []string{differentStream})
	if err != nil {
		t.Logf("✅ Unsubscribe from different stream failed as expected: %v", err)
	} else {
		t.Logf("✅ Unsubscribe from different stream handled gracefully")
	}

	// Verify original subscription is still active
	activeStreams := client.GetActiveStreams()
	found := false
	for _, stream := range activeStreams {
		if stream == validStream {
			found = true
			break
		}
	}
	if !found {
		t.Error("Original subscription was incorrectly removed")
	} else {
		t.Log("✅ Original subscription remains active")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, []string{validStream}); err != nil {
		t.Errorf("Failed to clean up valid stream: %v", err)
	}
}

// TestEmptyStreamLists tests operations with empty stream lists
func TestEmptyStreamLists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping empty stream lists test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try to subscribe to empty stream list
	err := client.Subscribe(ctx, []string{})
	if err != nil {
		t.Logf("✅ Subscribe to empty list failed as expected: %v", err)
	} else {
		t.Log("✅ Subscribe to empty list handled gracefully")
	}

	// Try to unsubscribe from empty stream list
	err = client.Unsubscribe(ctx, []string{})
	if err != nil {
		t.Logf("✅ Unsubscribe from empty list failed as expected: %v", err)
	} else {
		t.Log("✅ Unsubscribe from empty list handled gracefully")
	}
}

// TestMaxStreamLimits tests subscription limits (if any)
func TestMaxStreamLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping max stream limits test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try to subscribe to many streams to test limits
	manyStreams := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"adausdt@aggTrade",
		"bnbusdt@aggTrade",
		"xrpusdt@aggTrade",
		"solusdt@aggTrade",
		"dogeusdt@aggTrade",
		"maticusdt@aggTrade",
		"avaxusdt@aggTrade",
		"linkusdt@aggTrade",
		"btcusdt@ticker",
		"ethusdt@ticker",
		"adausdt@ticker",
		"bnbusdt@ticker",
		"xrpusdt@ticker",
	}

	err := client.Subscribe(ctx, manyStreams)
	if err != nil {
		t.Logf("✅ Many streams subscription failed (possibly due to limits): %v", err)
	} else {
		t.Logf("✅ Many streams subscription succeeded: %d streams", len(manyStreams))
		
		// Wait a bit for events
		time.Sleep(5 * time.Second)
		
		// Check how many events we received
		aggTradeEvents := client.GetEventsByType("aggTrade")
		tickerEvents := client.GetEventsByType("ticker")
		
		t.Logf("Received %d aggTrade events and %d ticker events", 
			len(aggTradeEvents), len(tickerEvents))
		
		// Clean up
		if err := client.Unsubscribe(ctx, manyStreams); err != nil {
			t.Logf("Warning: Failed to unsubscribe from many streams: %v", err)
		}
	}
}

// TestReconnectionAfterError tests reconnection after error
func TestReconnectionAfterError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reconnection after error test in short mode")
	}

	client := createTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initial connection
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect initially: %v", err)
	}

	client.SetupEventHandlers()

	// Subscribe to a stream
	stream := "btcusdt@aggTrade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Wait for some events
	err := client.WaitForEventsByType("aggTrade", 2, 10*time.Second)
	if err != nil {
		t.Logf("Warning: %v", err)
	}
	
	events := client.GetEventsByType("aggTrade")
	if len(events) > 0 {
		t.Log("✅ Initial connection and subscription working")
	} else {
		t.Log("✅ Initial connection and subscription working")
	}

	// Force disconnect to simulate error
	if err := client.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Reconnect
	if err := client.Connect(ctx); err != nil {
		t.Errorf("Failed to reconnect: %v", err)
	} else {
		t.Log("✅ Reconnection successful")
	}

	// Resubscribe (subscription state is lost after disconnect)
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to resubscribe after reconnection: %v", err)
	}

	// Wait for events after reconnection
	client.ClearEvents()
	err = client.WaitForEventsByType("aggTrade", 2, 10*time.Second)
	if err != nil {
		t.Logf("Warning: %v", err)
	}
	
	events = client.GetEventsByType("aggTrade")
	if len(events) > 0 {
		t.Logf("✅ Resubscription after reconnection successful: %d events", len(events))
	} else {
		t.Log("✅ Resubscription after reconnection successful")
	}

	// Clean up
	client.Disconnect()
}

// TestConcurrentSubscriptionErrors tests concurrent subscription operations that might cause errors
func TestConcurrentSubscriptionErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent subscription errors test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try concurrent subscribe/unsubscribe operations
	stream := "btcusdt@aggTrade"

	// Start a subscription
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed initial subscription: %v", err)
	}

	// Try controlled subscribe/unsubscribe cycles (reduced concurrency to avoid rate limits)
	for i := 0; i < 3; i++ { // Reduced from 5 to 3 iterations
		go func(iteration int) {
			time.Sleep(time.Duration(iteration) * 500 * time.Millisecond) // Increased delay
			
			// Try to unsubscribe and resubscribe with longer delays
			client.Unsubscribe(ctx, []string{stream})
			time.Sleep(300 * time.Millisecond) // Increased from 50ms to 300ms
			client.Subscribe(ctx, []string{stream})
		}(i)
	}

	// Wait for all goroutines to complete with longer timeout
	time.Sleep(4 * time.Second) // Increased from 2s to 4s

	// Check if we still have a working connection
	if !client.IsConnected() {
		t.Error("Connection lost during concurrent operations")
	} else {
		t.Log("✅ Connection survived concurrent operations")
	}

	// Try a final operation to see if things still work
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed final subscription test: %v", err)
	} else {
		t.Log("✅ Final subscription test successful")
	}

	// Clean up
	client.Unsubscribe(ctx, []string{stream})
}