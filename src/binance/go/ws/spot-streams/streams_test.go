package streamstest

import (
	"context"
	"testing"
	"time"
)

// TestTradeStream tests individual trade stream functionality
func TestTradeStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@trade", "trade", 3)
}

// TestAggregateTradeStream tests aggregate trade stream functionality
func TestAggregateTradeStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@aggTrade", "aggTrade", 3)
}

// TestKlineStream tests kline stream functionality
func TestKlineStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@kline_1m", "kline", 2)
}

// TestTickerStream tests 24hr ticker stream functionality
func TestTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@ticker", "ticker", 1)
}

// TestMiniTickerStream tests 24hr mini ticker stream functionality
func TestMiniTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@miniTicker", "miniTicker", 1)
}

// TestBookTickerStream tests best bid/ask stream functionality
func TestBookTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@bookTicker", "bookTicker", 3)
}

// TestDepthStream tests order book depth stream functionality
func TestDepthStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@depth", "depth", 5)
}

// TestPartialDepthStream tests partial order book depth stream functionality
func TestPartialDepthStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@depth5", "partialDepth", 5)
}

// TestRollingWindowTickerStream tests rolling window ticker stream functionality
func TestRollingWindowTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@ticker_1h", "rollingWindowTicker", 1)
}

// TestAvgPriceStream tests average price stream functionality
func TestAvgPriceStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@avgPrice", "avgPrice", 1)
}

// TestMultipleSymbolStreams tests streams for multiple symbols
func TestMultipleSymbolStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple symbol streams test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple symbols
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"adausdt@trade",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to multiple streams: %v", err)
	}

	// Wait for events
	t.Log("Waiting for events from multiple symbols...")
	if err := client.WaitForEventsByType("trade", 10, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsByType("trade")
	t.Logf("Received %d trade events from multiple symbols", len(events))

	if len(events) == 0 {
		t.Error("No trade events received from multiple symbols")
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from multiple streams: %v", err)
	}
}

// TestMultipleStreamTypes tests multiple stream types simultaneously
func TestMultipleStreamTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined streams test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple streams of different types to test combined functionality
	streams := []string{
		"btcusdt@trade",
		"btcusdt@ticker",
		"ethusdt@trade",
		"ethusdt@miniTicker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to combined streams: %v", err)
	}

	t.Log("✅ Successfully subscribed to combined streams")

	// Wait for events from multiple stream types
	t.Log("Waiting for combined stream events...")
	
	// Test for different event types
	eventTypes := []string{"trade", "ticker", "miniTicker"}
	receivedEvents := make(map[string]int)
	
	for _, eventType := range eventTypes {
		// Wait for events of this type
		if err := client.WaitForEventsByType(eventType, 2, 15*time.Second); err != nil {
			t.Logf("⚠️  Timeout waiting for %s events: %v", eventType, err)
		}

		// Check received events
		events := client.GetEventsByType(eventType)
		receivedEvents[eventType] = len(events)
		t.Logf("Received %d %s events", len(events), eventType)
	}

	// Check for combined stream events specifically
	combinedEvents := client.GetEventsByType("combinedStream")
	t.Logf("Received %d combined stream events", len(combinedEvents))

	// Verify we received events from multiple stream types
	totalEvents := 0
	for eventType, count := range receivedEvents {
		totalEvents += count
		if count > 0 {
			t.Logf("✅ Successfully received %d %s events", count, eventType)
		}
	}

	if totalEvents == 0 {
		t.Error("❌ No events received from combined streams")
	} else {
		t.Logf("✅ Combined streams test successful: %d total events received", totalEvents)
	}

	// Test concurrent stream processing
	allEvents := client.GetEventsReceived()
	if len(allEvents) > 0 {
		t.Logf("✅ Combined stream event processing working: %d total events processed", len(allEvents))
		
		// Verify event integrity
		for i, event := range allEvents {
			if i >= 5 { // Just check first 5 events
				break
			}
			if eventMap, ok := event.(map[string]interface{}); ok {
				if eventType, exists := eventMap["type"]; exists {
					t.Logf("Event %d: type=%s", i+1, eventType)
				}
			}
		}
	}

	// Unsubscribe from all streams
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from combined streams: %v", err)
	}

	t.Log("✅ Successfully unsubscribed from combined streams")
}

// TestDifferentKlineIntervals tests different kline intervals
func TestDifferentKlineIntervals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping kline intervals test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test different intervals
	intervals := []string{"1m", "5m", "15m", "1h"}
	
	for _, interval := range intervals {
		t.Run(interval, func(t *testing.T) {
			stream := "btcusdt@kline_" + interval
			
			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Fatalf("Failed to subscribe to %s: %v", stream, err)
			}

			// Wait for at least one event
			if err := client.WaitForEventsByType("kline", 1, 20*time.Second); err != nil {
				t.Logf("Warning: %v", err)
			}

			events := client.GetEventsByType("kline")
			t.Logf("Received %d kline events for %s interval", len(events), interval)

			if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestDifferentDepthLevels tests different depth levels
func TestDifferentDepthLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping depth levels test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test different depth levels
	depthLevels := []string{"5", "10", "20"}
	
	for _, level := range depthLevels {
		t.Run("depth"+level, func(t *testing.T) {
			stream := "btcusdt@depth" + level
			
			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Fatalf("Failed to subscribe to %s: %v", stream, err)
			}

			// Wait for events
			if err := client.WaitForEventsByType("partialDepth", 3, 20*time.Second); err != nil {
				t.Logf("Warning: %v", err)
			}

			events := client.GetEventsByType("partialDepth")
			t.Logf("Received %d partial depth events for level %s", len(events), level)

			if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestAllTickerStream tests all symbols ticker stream
func TestAllTickerStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all ticker stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to all symbols ticker
	stream := "!ticker@arr"
	
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Wait for events (all symbols ticker updates less frequently)
	t.Log("Waiting for all ticker events...")
	if err := client.WaitForEventsByType("ticker", 1, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("ticker")
	t.Logf("Received %d all ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
	}
}

// TestAllMiniTickerStream tests all symbols mini ticker stream
func TestAllMiniTickerStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all mini ticker stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to all symbols mini ticker
	stream := "!miniTicker@arr"
	
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Wait for events
	t.Log("Waiting for all mini ticker events...")
	if err := client.WaitForEventsByType("miniTicker", 1, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("miniTicker")
	t.Logf("Received %d all mini ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
	}
}

// TestAllBookTickerStream tests all symbols book ticker stream
func TestAllBookTickerStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all book ticker stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to all symbols book ticker
	stream := "!bookTicker"
	
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Wait for events
	t.Log("Waiting for all book ticker events...")
	if err := client.WaitForEventsByType("bookTicker", 5, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("bookTicker")
	t.Logf("Received %d all book ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
	}
}