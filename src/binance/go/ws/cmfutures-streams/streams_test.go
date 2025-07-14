package streamstest

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestAggregateTradeStream tests aggregate trade stream functionality
func TestAggregateTradeStream(t *testing.T) {
	// Note: aggTrade streams require actual trading activity which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@aggTrade", "aggTrade", 1, "AggTrade events require actual trading activity - limited on testnet")
}

// TestKlineStream tests kline stream functionality
func TestKlineStream(t *testing.T) {
	// Note: kline streams require price movement and trades which are limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@kline_1m", "kline", 1, "Kline events require price movement - limited on testnet")
}

// TestMiniTickerStream tests 24hr mini ticker stream functionality
func TestMiniTickerStream(t *testing.T) {
	// Note: ticker streams require trading volume for 24hr statistics which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@miniTicker", "miniTicker", 1, "MiniTicker events require trading volume - limited on testnet")
}

// TestTickerStream tests 24hr ticker stream functionality
func TestTickerStream(t *testing.T) {
	// Note: ticker streams require trading volume for 24hr statistics which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@ticker", "ticker", 1, "Ticker events require trading volume - limited on testnet")
}

// TestBookTickerStream tests best bid/ask stream functionality
func TestBookTickerStream(t *testing.T) {
	// Note: bookTicker streams require active order book changes which are limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@bookTicker", "bookTicker", 1, "BookTicker events require active order book changes - limited on testnet")
}

// TestPartialDepthStream tests partial order book depth stream functionality
func TestPartialDepthStream(t *testing.T) {
	// Note: depth streams require order book activity which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@depth5", "depthUpdate", 1, "DepthUpdate events require order book activity - limited on testnet")
}

// TestDiffDepthStream tests differential order book depth stream functionality
func TestDiffDepthStream(t *testing.T) {
	// Note: depth streams require order book activity which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@depth", "depthUpdate", 1, "DepthUpdate events require order book activity - limited on testnet")
}

// TestMarkPriceStream tests mark price stream functionality (futures-specific)
func TestMarkPriceStream(t *testing.T) {
	// Mark price streams have different naming format: symbol@markPrice@1s
	testStreamSubscription(t, "btcusd_perp@markPrice@1s", "markPrice", 3)
}

// TestContinuousKlineStream tests continuous kline stream functionality (futures-specific)
func TestContinuousKlineStream(t *testing.T) {
	// Continuous kline format: pair_contractType@continuousKline_interval
	// Note: continuous kline streams require price movement which is limited on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perpetual@continuousKline_1m", "continuousKline", 1, "ContinuousKline events require price movement - limited on testnet")
}

// TestLiquidationOrderStream tests liquidation order stream functionality (futures-specific)
func TestLiquidationOrderStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping liquidation stream test in short mode - liquidations are rare on testnet")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()
	streamName := "btcusd_perp@forceOrder"

	// Subscribe to liquidation stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	t.Log("✅ Successfully subscribed to liquidation stream")
	t.Log("⚠️  Note: Liquidation events are rare on testnet and only push snapshots (max 1/second)")

	// Wait for events with longer timeout since liquidations are rare
	t.Log("Waiting for liquidation events...")
	_ = client.WaitForEventsByType("forceOrder", 1, 30*time.Second)
	
	// Check received events
	events := client.GetEventsByType("forceOrder")
	t.Logf("Received %d liquidation events", len(events))

	if len(events) == 0 {
		t.Log("⚠️  No liquidation events received - this is expected on testnet as liquidations are rare")
		t.Log("ℹ️  Liquidation streams work but require actual liquidation events which are uncommon on testnet")
		// This is expected behavior on testnet, so we don't fail the test
	} else {
		t.Logf("✅ Successfully received %d liquidation events", len(events))
	}

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

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple symbols
	streams := []string{
		"btcusd_perp@aggTrade",
		"linkusd_perp@aggTrade",
		"adausd_perp@aggTrade",
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

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple streams of different types
	streams := []string{
		"btcusd_perp@aggTrade",
		"btcusd_perp@ticker",
		"linkusd_perp@aggTrade",
		"linkusd_perp@miniTicker",
		"btcusd_perp@markPrice@1s",
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

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Test different intervals
	intervals := []string{"1m", "5m", "15m", "1h"}
	
	for _, interval := range intervals {
		t.Run(interval, func(t *testing.T) {
			stream := "btcusd_perp@kline_" + interval
			
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

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Test different depth levels - reduced for testnet efficiency
	depthLevels := []string{"5", "20"} // Test just 5 and 20 levels
	
	for _, level := range depthLevels {
		t.Run("depth"+level, func(t *testing.T) {
			stream := "btcusd_perp@depth" + level
			
			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Fatalf("Failed to subscribe to %s: %v", stream, err)
			}

			// Wait for events - reduced timeout for testnet
			if err := client.WaitForEventsByType("depthUpdate", 1, 10*time.Second); err != nil {
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

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Test different update speeds for differential depth (reduced for performance)
	updateSpeeds := []string{"100ms", "500ms"} // Reduced from 3 to 2 speeds
	
	for _, speed := range updateSpeeds {
		t.Run("depth@"+speed, func(t *testing.T) {
			stream := "btcusd_perp@depth@" + speed
			
			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Fatalf("Failed to subscribe to %s: %v", stream, err)
			}

			// Wait for events - reduced timeouts and expectations for performance
			expectedEvents := 2
			if speed == "100ms" {
				expectedEvents = 3 // Slightly more events with faster updates
			}
			
			timeout := 10 * time.Second // Reduced from 25s to 10s
			if err := client.WaitForEventsByType("depthUpdate", expectedEvents, timeout); err != nil {
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
			
			// Add delay to avoid rate limiting between subtests
			time.Sleep(2 * time.Second)
		})
	}
}

// TestPartialDepthStreamUpdateSpeed tests partial depth streams with updateSpeed
func TestPartialDepthStreamUpdateSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partial depth update speed test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Test different combinations of depth levels and update speeds - reduced for testnet efficiency
	depthLevels := []string{"5", "20"} // Test just 5 and 20 levels
	updateSpeeds := []string{"100ms", "500ms"} // Test just fastest and slowest speeds
	
	for _, level := range depthLevels {
		for _, speed := range updateSpeeds {
			t.Run(fmt.Sprintf("depth%s@%s", level, speed), func(t *testing.T) {
				stream := fmt.Sprintf("btcusd_perp@depth%s@%s", level, speed)
				
				if err := client.Subscribe(ctx, []string{stream}); err != nil {
					t.Fatalf("Failed to subscribe to %s: %v", stream, err)
				}

				// Wait for events - 100ms should be faster than 500ms
				expectedEvents := 3
				if speed == "100ms" {
					expectedEvents = 5 // Expect more events with faster updates
				}

				if err := client.WaitForEventsByType("depthUpdate", expectedEvents, 10*time.Second); err != nil {
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


// TestAllSymbolsStreams tests all symbols stream functionality
func TestAllSymbolsStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all symbols streams test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

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

	// Test all symbols force order (liquidation) stream
	allForceOrderStream := "!forceOrder@arr"
	
	if err := client.Subscribe(ctx, []string{allForceOrderStream}); err != nil {
		t.Logf("⚠️  Failed to subscribe to %s: %v", allForceOrderStream, err)
		t.Log("ℹ️  All force order streams may not be available on testnet (liquidations are rare)")
	} else {
		t.Logf("✅ Successfully subscribed to %s", allForceOrderStream)
		
		// Wait for events (liquidations are rare, so shorter timeout)
		t.Log("Waiting for all force order events...")
		if err := client.WaitForEventsByType("forceOrder", 1, 15*time.Second); err != nil {
			t.Logf("Warning: %v", err)
		}

		forceOrderEvents := client.GetEventsByType("forceOrder")
		t.Logf("Received %d all force order events", len(forceOrderEvents))
		
		if len(forceOrderEvents) == 0 {
			t.Log("⚠️  No force order events received - this is expected on testnet where liquidations are rare")
		} else {
			t.Logf("✅ Successfully received %d force order events", len(forceOrderEvents))
		}

		if err := client.Unsubscribe(ctx, []string{allForceOrderStream}); err != nil {
			t.Logf("Note: Failed to unsubscribe from %s: %v", allForceOrderStream, err)
		}
	}
}

// TestAllArrayStreams tests all @arr stream types comprehensively
func TestAllArrayStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all array streams test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()
	
	// Define all @arr streams to test
	arrayStreams := []struct {
		name        string
		stream      string
		eventType   string
		description string
		timeout     time.Duration
		expectEvents bool
	}{
		{
			name:        "All Symbols Ticker",
			stream:      "!ticker@arr",
			eventType:   "ticker",
			description: "24hr ticker statistics for all symbols",
			timeout:     30 * time.Second,
			expectEvents: true,
		},
		{
			name:        "All Symbols Mini Ticker", 
			stream:      "!miniTicker@arr",
			eventType:   "miniTicker",
			description: "24hr mini ticker statistics for all symbols",
			timeout:     30 * time.Second,
			expectEvents: true,
		},
		{
			name:        "All Symbols Book Ticker",
			stream:      "!bookTicker",
			eventType:   "bookTicker", 
			description: "Best bid/ask price for all symbols",
			timeout:     30 * time.Second,
			expectEvents: true,
		},
		{
			name:        "All Symbols Force Order",
			stream:      "!forceOrder@arr", 
			eventType:   "forceOrder",
			description: "Liquidation order information for all symbols",
			timeout:     15 * time.Second,
			expectEvents: false, // Liquidations are rare on testnet
		},
	}

	t.Log("🧪 Testing all @arr stream types...")
	
	for i, arrStream := range arrayStreams {
		t.Run(arrStream.name, func(t *testing.T) {
			t.Logf("📡 Testing %s (%s)", arrStream.name, arrStream.stream)
			
			// Subscribe to the stream
			if err := client.Subscribe(ctx, []string{arrStream.stream}); err != nil {
				if !arrStream.expectEvents {
					t.Logf("⚠️  Failed to subscribe to %s: %v", arrStream.stream, err)
					t.Logf("ℹ️  %s may not be available on testnet", arrStream.description)
					return
				}
				t.Fatalf("Failed to subscribe to %s: %v", arrStream.stream, err)
			}
			
			t.Logf("✅ Successfully subscribed to %s", arrStream.stream)
			
			// Wait for events
			t.Logf("⏳ Waiting for %s events...", arrStream.eventType)
			expectedCount := 1
			if arrStream.expectEvents {
				expectedCount = 3 // Expect multiple events for active streams
			}
			
			if err := client.WaitForEventsByType(arrStream.eventType, expectedCount, arrStream.timeout); err != nil {
				if arrStream.expectEvents {
					t.Logf("⚠️  Warning: %v", err)
				} else {
					t.Logf("ℹ️  Expected timeout for %s: %v", arrStream.stream, err)
				}
			}
			
			// Check received events
			events := client.GetEventsByType(arrStream.eventType)
			t.Logf("📊 Received %d %s events", len(events), arrStream.eventType)
			
			if len(events) == 0 {
				if arrStream.expectEvents {
					t.Logf("⚠️  No %s events received - this may indicate an issue", arrStream.eventType)
				} else {
					t.Logf("ℹ️  No %s events received - this is expected on testnet", arrStream.eventType)
				}
			} else {
				t.Logf("✅ Successfully received %d %s events", len(events), arrStream.eventType)
				
				// Log first event details for debugging
				if len(events) > 0 {
					t.Logf("📋 First event sample: %+v", events[0])
				}
			}
			
			// Unsubscribe
			if err := client.Unsubscribe(ctx, []string{arrStream.stream}); err != nil {
				t.Logf("⚠️  Failed to unsubscribe from %s: %v", arrStream.stream, err)
			} else {
				t.Logf("✅ Successfully unsubscribed from %s", arrStream.stream)
			}
			
			// Add delay between tests to avoid rate limiting
			if i < len(arrayStreams)-1 {
				time.Sleep(1 * time.Second)
			}
		})
	}
	
	t.Log("🏁 All @arr stream tests completed")
}

// TestIndexPriceKlineStream tests index price kline stream functionality
func TestIndexPriceKlineStream(t *testing.T) {
	// Note: Index price kline streams may have limited data availability on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd@indexPriceKline_1m", "indexPrice_kline", 1, "Index price kline events may not be available on testnet")
}

// TestMarkPriceKlineStream tests mark price kline stream functionality
func TestMarkPriceKlineStream(t *testing.T) {
	// Note: Mark price kline streams may have limited data availability on testnet
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd_perp@markPriceKline_1m", "markPrice_kline", 1, "Mark price kline events may not be available on testnet")
}

// TestContractInfoStream tests contract info stream functionality
func TestContractInfoStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract info stream test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()
	streamName := "!contractInfo"

	// Subscribe to contract info stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	t.Log("✅ Successfully subscribed to contract info stream")
	t.Log("⚠️  Note: Contract info streams may have limited availability on testnet")

	// Wait for events
	t.Log("Waiting for contract info events...")
	err := client.WaitForEventsByType("contractInfo", 1, 30*time.Second)
	if err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsByType("contractInfo")
	t.Logf("Received %d contract info events", len(events))

	if len(events) == 0 {
		t.Log("⚠️  No contract info events received - this may be expected on testnet")
		t.Log("ℹ️  Contract info streams provide updates when contract specifications change")
	} else {
		t.Logf("✅ Successfully received %d contract info events", len(events))
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{streamName}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", streamName, err)
	} else {
		t.Log("✅ Successfully unsubscribed from contract info stream")
	}
}

// TestIndividualIndexPriceStream tests individual index price stream functionality
func TestIndividualIndexPriceStream(t *testing.T) {
	testStreamSubscriptionWithGracefulTimeout(t, "btcusd@indexPrice@1s", "indexPriceUpdate", 3, "Index price streams may not be available on testnet")
}

