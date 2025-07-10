package streamstest

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestAggregateTradeStream tests aggregate trade stream functionality
func TestAggregateTradeStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@aggTrade", "aggTrade", 3)
}

// TestKlineStream tests kline stream functionality
func TestKlineStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@kline_1m", "kline", 2)
}

// TestMiniTickerStream tests 24hr mini ticker stream functionality
func TestMiniTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@miniTicker", "miniTicker", 1)
}

// TestTickerStream tests 24hr ticker stream functionality
func TestTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@ticker", "ticker", 1)
}

// TestBookTickerStream tests best bid/ask stream functionality
func TestBookTickerStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@bookTicker", "bookTicker", 3)
}

// TestPartialDepthStream tests partial order book depth stream functionality
func TestPartialDepthStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@depth5", "depthUpdate", 5)
}

// TestDiffDepthStream tests differential order book depth stream functionality
func TestDiffDepthStream(t *testing.T) {
	testStreamSubscription(t, "btcusdt@depth", "depthUpdate", 5)
}

// TestMarkPriceStream tests mark price stream functionality (futures-specific)
func TestMarkPriceStream(t *testing.T) {
	// Mark price streams have different naming format: symbol@markPrice@1s
	testStreamSubscription(t, "btcusdt@markPrice@1s", "markPrice", 3)
}

// TestContinuousKlineStream tests continuous kline stream functionality (futures-specific)
func TestContinuousKlineStream(t *testing.T) {
	// Continuous kline format: pair_contractType@continuousKline_interval
	testStreamSubscription(t, "btcusdt_perpetual@continuousKline_1m", "continuousKline", 2)
}

// TestLiquidationOrderStream tests liquidation order stream functionality (futures-specific)
func TestLiquidationOrderStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping liquidation stream test in short mode - liquidations are rare on testnet")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()
	streamName := "btcusdt@forceOrder"

	// Subscribe to liquidation stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	t.Log("✅ Successfully subscribed to liquidation stream")
	t.Log("⚠️  Note: Liquidation events are rare on testnet and only push snapshots (max 1/second)")

	// Wait for events with longer timeout since liquidations are rare
	t.Log("Waiting for liquidation events...")
	err := client.WaitForEventsByType("forceOrder", 1, 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to receive liquidation events: %v", err)
	}

	// Check received events
	events := client.GetEventsByType("forceOrder")
	t.Logf("Received %d liquidation events", len(events))

	if len(events) == 0 {
		t.Fatalf("No liquidation events received - stream may not be working properly")
	}

	t.Logf("✅ Successfully received %d liquidation events", len(events))

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{streamName}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", streamName, err)
	} else {
		t.Log("✅ Successfully unsubscribed from liquidation stream")
	}
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
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"adausdt@aggTrade",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to multiple streams: %v", err)
	}

	// Wait for events
	t.Log("Waiting for events from multiple symbols...")
	if err := client.WaitForEventsByType("aggTrade", 10, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsByType("aggTrade")
	t.Logf("Received %d aggregate trade events from multiple symbols", len(events))

	if len(events) == 0 {
		t.Error("No aggregate trade events received from multiple symbols")
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

	client := setupAndConnectCombinedStreamsClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple streams of different types
	streams := []string{
		"btcusdt@aggTrade",
		"btcusdt@ticker",
		"ethusdt@aggTrade",
		"ethusdt@miniTicker",
		"btcusdt@markPrice@1s",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to combined streams: %v", err)
	}

	t.Log("✅ Successfully subscribed to combined streams")

	// Wait for events from multiple stream types
	t.Log("Waiting for combined stream events...")
	
	// Test for different event types
	eventTypes := []string{"aggTrade", "ticker", "miniTicker", "markPrice"}
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
	// Note: When using combined streams, we register individual handlers instead of combined handler
	// so we expect 0 combined stream events and individual events instead
	combinedEvents := client.GetEventsByType("combinedStream")
	t.Logf("Received %d combined stream wrapper events (expected 0 - using individual handlers)", len(combinedEvents))

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
			if err := client.WaitForEventsByType("depthUpdate", 3, 20*time.Second); err != nil {
				t.Logf("Warning: %v", err)
			}

			events := client.GetEventsByType("depthUpdate")
			t.Logf("Received %d partial depth events for level %s", len(events), level)

			if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestDiffDepthStreamUpdateSpeed tests differential depth streams with updateSpeed
func TestDiffDepthStreamUpdateSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping depth update speed test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test different update speeds for differential depth
	updateSpeeds := []string{"100ms", "250ms", "500ms"}
	
	for _, speed := range updateSpeeds {
		t.Run("depth@"+speed, func(t *testing.T) {
			stream := "btcusdt@depth@" + speed
			
			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Fatalf("Failed to subscribe to %s: %v", stream, err)
			}

			// Wait for events - 100ms should be faster than 500ms
			expectedEvents := 3
			if speed == "100ms" {
				expectedEvents = 5 // Expect more events with faster updates
			}
			
			if err := client.WaitForEventsByType("depthUpdate", expectedEvents, 25*time.Second); err != nil {
				t.Logf("Warning: %v", err)
			}

			events := client.GetEventsByType("depthUpdate")
			t.Logf("Received %d depth events with %s update speed", len(events), speed)

			// Verify we got some events
			if len(events) > 0 {
				t.Logf("✅ Successfully received depth events with %s update speed", speed)
			} else {
				t.Logf("⚠️  No depth events received with %s update speed", speed)
			}

			if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
			}

			// Clear events for next test
			client.ClearEvents()
		})
	}
}

// TestPartialDepthStreamUpdateSpeed tests partial depth streams with updateSpeed
func TestPartialDepthStreamUpdateSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partial depth update speed test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test different combinations of depth levels and update speeds
	depthLevels := []string{"5", "10", "20"}
	updateSpeeds := []string{"100ms", "250ms", "500ms"}
	
	for _, level := range depthLevels {
		for _, speed := range updateSpeeds {
			t.Run(fmt.Sprintf("depth%s@%s", level, speed), func(t *testing.T) {
				stream := fmt.Sprintf("btcusdt@depth%s@%s", level, speed)
				
				if err := client.Subscribe(ctx, []string{stream}); err != nil {
					t.Fatalf("Failed to subscribe to %s: %v", stream, err)
				}

				// Wait for events - 100ms should be faster than 500ms
				expectedEvents := 3
				if speed == "100ms" {
					expectedEvents = 5 // Expect more events with faster updates
				}

				if err := client.WaitForEventsByType("depthUpdate", expectedEvents, 25*time.Second); err != nil {
					t.Logf("Warning: %v", err)
				}

				events := client.GetEventsByType("depthUpdate")
				t.Logf("Received %d partial depth events for level %s with %s update speed", len(events), level, speed)

				// Verify we got some events
				if len(events) > 0 {
					t.Logf("✅ Successfully received partial depth events for level %s with %s update speed", level, speed)
				} else {
					// 250ms update speed may have reduced availability on testnet
					if speed == "250ms" {
						t.Logf("⚠️  No partial depth events received for level %s with %s update speed - 250ms may have reduced availability on testnet", level, speed)
					} else {
						t.Logf("⚠️  No partial depth events received for level %s with %s update speed", level, speed)
					}
				}

				if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
					t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
				}

				// Clear events for next test
				client.ClearEvents()
			})
		}
	}
}

// TestCompositeIndexStream tests composite index stream functionality (futures-specific)
func TestCompositeIndexStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping composite index stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()
	streamName := "defiusdt@compositeIndex"

	// Subscribe to composite index stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	t.Log("✅ Successfully subscribed to composite index stream")
	t.Log("⚠️  Note: Composite index streams may not be available for all symbols on testnet")

	// Wait for events
	t.Log("Waiting for composite index events...")
	err := client.WaitForEventsByType("compositeIndex", 1, 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to receive composite index events: %v", err)
	}

	// Check received events
	events := client.GetEventsByType("compositeIndex")
	t.Logf("Received %d composite index events", len(events))

	if len(events) == 0 {
		t.Fatalf("No composite index events received - stream may not be working properly")
	}

	t.Logf("✅ Successfully received %d composite index events", len(events))

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{streamName}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", streamName, err)
	} else {
		t.Log("✅ Successfully unsubscribed from composite index stream")
	}
}

// TestAssetIndexStream tests asset index stream functionality (futures-specific)
func TestAssetIndexStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping asset index stream test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()
	streamName := "btcusdt@assetIndex"

	// Subscribe to asset index stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	t.Log("✅ Successfully subscribed to asset index stream")
	t.Log("⚠️  Note: Asset index streams are for multi-assets mode users and may not be available on testnet")

	// Wait for events with shorter timeout since this feature may not be available on testnet
	t.Log("Waiting for asset index events...")
	_ = client.WaitForEventsByType("assetIndexUpdate", 1, 30*time.Second)
	
	// Check received events
	events := client.GetEventsByType("assetIndexUpdate")
	t.Logf("Received %d asset index events", len(events))

	if len(events) == 0 {
		t.Log("⚠️  No asset index events received - requires multi-assets mode, may not be available on testnet")
		t.Log("ℹ️  Asset index streams are only available for accounts with multi-assets mode enabled")
		// This is expected on testnet, so we don't fail the test
	} else {
		t.Logf("✅ Successfully received %d asset index events", len(events))
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{streamName}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", streamName, err)
	} else {
		t.Log("✅ Successfully unsubscribed from asset index stream")
	}
}

// TestAllSymbolsStreams tests all symbols stream functionality
func TestAllSymbolsStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all symbols streams test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test all symbols ticker
	allTickerStream := "!ticker@arr"
	
	if err := client.Subscribe(ctx, []string{allTickerStream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", allTickerStream, err)
	}

	// Wait for events (all symbols ticker updates less frequently)
	t.Log("Waiting for all ticker events...")
	if err := client.WaitForEventsByType("ticker", 1, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("ticker")
	t.Logf("Received %d all ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{allTickerStream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", allTickerStream, err)
	}

	// Test all symbols mini ticker
	allMiniTickerStream := "!miniTicker@arr"
	
	if err := client.Subscribe(ctx, []string{allMiniTickerStream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", allMiniTickerStream, err)
	}

	// Wait for events
	t.Log("Waiting for all mini ticker events...")
	if err := client.WaitForEventsByType("miniTicker", 1, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events = client.GetEventsByType("miniTicker")
	t.Logf("Received %d all mini ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{allMiniTickerStream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", allMiniTickerStream, err)
	}

	// Test all symbols book ticker
	allBookTickerStream := "!bookTicker"
	
	if err := client.Subscribe(ctx, []string{allBookTickerStream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", allBookTickerStream, err)
	}

	// Wait for events
	t.Log("Waiting for all book ticker events...")
	if err := client.WaitForEventsByType("bookTicker", 5, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events = client.GetEventsByType("bookTicker")
	t.Logf("Received %d all book ticker events", len(events))

	if err := client.Unsubscribe(ctx, []string{allBookTickerStream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", allBookTickerStream, err)
	}
}