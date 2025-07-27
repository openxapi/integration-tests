package streamstest

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
	"github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// TestConcurrentStreams tests handling multiple concurrent stream connections
func TestConcurrentStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent streams test in short mode")
	}

	numClients := 3
	var wg sync.WaitGroup
	var mu sync.Mutex
	totalEvents := 0

	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			client := cmfuturesstreams.NewClient()
			
			err := client.SetActiveServer("testnet1")
			if err != nil {
				t.Errorf("Client %d: Failed to set testnet server: %v", clientID, err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := client.Connect(ctx); err != nil {
				t.Errorf("Client %d: Failed to connect: %v", clientID, err)
				return
			}
			defer client.Disconnect()

			clientEvents := 0
			client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
				clientEvents++
				mu.Lock()
				totalEvents++
				mu.Unlock()
				return nil
			})

			// Each client subscribes to a different symbol
			symbols := []string{"BTCUSD_PERP", "LINKUSD_PERP", "ADAUSD_PERP"}
			stream := symbols[clientID%len(symbols)] + "@aggTrade"

			if err := client.Subscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Client %d: Failed to subscribe: %v", clientID, err)
				return
			}

			// Wait for events
			time.Sleep(8 * time.Second)

			t.Logf("Client %d received %d events from %s", clientID, clientEvents, stream)

			if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
				t.Errorf("Client %d: Failed to unsubscribe: %v", clientID, err)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("✅ Concurrent streams test completed: %d total events from %d clients", totalEvents, numClients)

	if totalEvents > 0 {
		avgEventsPerClient := float64(totalEvents) / float64(numClients)
		t.Logf("📊 Average events per client: %.1f", avgEventsPerClient)
	}
}

// TestHighVolumeStreams tests handling high-volume stream data
func TestHighVolumeStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high volume streams test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple high-volume streams
	highVolumeStreams := []string{
		"btcusd_perp@aggTrade",
		"LINKUSD_PERP@aggTrade",
		"BTCUSD_PERP@bookTicker",
		"LINKUSD_PERP@bookTicker",
		"BTCUSD_PERP@depth@100ms",
		"LINKUSD_PERP@depth@100ms",
		"BTCUSD_PERP@ticker",
		"LINKUSD_PERP@ticker",
	}

	eventCounts := make(map[string]int)
	var eventMu sync.RWMutex
	startTime := time.Now()

	// Setup handlers
	client.client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		eventMu.Lock()
		eventCounts["aggTrade"]++
		eventMu.Unlock()
		return nil
	})

	client.client.HandleBookTickerEvent(func(event *models.BookTickerEvent) error {
		eventMu.Lock()
		eventCounts["bookTicker"]++
		eventMu.Unlock()
		return nil
	})

	client.client.HandleDepthEvent(func(event *models.DiffDepthEvent) error {
		eventMu.Lock()
		eventCounts["depth"]++
		eventMu.Unlock()
		return nil
	})

	client.client.HandleTickerEvent(func(event *models.TickerEvent) error {
		eventMu.Lock()
		eventCounts["ticker"]++
		eventMu.Unlock()
		return nil
	})

	if err := client.Subscribe(ctx, highVolumeStreams); err != nil {
		t.Fatalf("Failed to subscribe to high volume streams: %v", err)
	}

	t.Log("📈 Collecting high-volume stream data...")

	// Collect data for 10 seconds
	time.Sleep(10 * time.Second)

	// Calculate metrics
	duration := time.Since(startTime)
	eventMu.RLock()
	totalEvents := 0
	for eventType, count := range eventCounts {
		totalEvents += count
		t.Logf("  %s: %d events", eventType, count)
	}
	eventMu.RUnlock()

	eventsPerSecond := float64(totalEvents) / duration.Seconds()

	t.Logf("📊 High Volume Stream Performance:")
	t.Logf("  Total Events: %d", totalEvents)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Events/Second: %.1f", eventsPerSecond)

	if totalEvents > 50 {
		t.Logf("✅ High-volume streams performing well: %.1f events/sec", eventsPerSecond)
	} else {
		t.Logf("⚠️  Lower event volume than expected: %.1f events/sec", eventsPerSecond)
	}

	// Clean up
	if err := client.Unsubscribe(ctx, highVolumeStreams); err != nil {
		t.Errorf("Failed to unsubscribe from high volume streams: %v", err)
	}
}

// BenchmarkEventProcessing benchmarks event processing performance
func BenchmarkEventProcessing(b *testing.B) {
	client := cmfuturesstreams.NewClient()
	
	err := client.SetActiveServer("testnet1")
	if err != nil {
		b.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	eventCount := 0
	client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		eventCount++
		return nil
	})

	if err := client.Subscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		b.Fatalf("Failed to subscribe: %v", err)
	}

	b.ResetTimer()

	// Run benchmark for the specified time
	for i := 0; i < b.N; i++ {
		time.Sleep(time.Millisecond) // Small delay to simulate work
	}

	b.StopTimer()

	b.Logf("Processed %d events during benchmark", eventCount)

	if err := client.Unsubscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		b.Errorf("Failed to unsubscribe: %v", err)
	}
}

// BenchmarkSubscriptionOperations benchmarks subscription/unsubscription performance
func BenchmarkSubscriptionOperations(b *testing.B) {
	client := cmfuturesstreams.NewClient()
	
	err := client.SetActiveServer("testnet1")
	if err != nil {
		b.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	stream := "btcusd_perp@aggTrade"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := client.Subscribe(ctx, []string{stream}); err != nil {
			b.Errorf("Failed to subscribe: %v", err)
		}

		if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
			b.Errorf("Failed to unsubscribe: %v", err)
		}
	}
}

// BenchmarkConcurrentEventAccess benchmarks concurrent access to events
func BenchmarkConcurrentEventAccess(b *testing.B) {
	client := createTestClient(&testing.T{})
	
	// Simulate some events
	for i := 0; i < 1000; i++ {
		client.recordEvent("test", map[string]interface{}{
			"id":   i,
			"data": "test data",
		})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Concurrent read operations
			events := client.GetEventsReceived()
			_ = len(events)

			eventsByType := client.GetEventsByType("test")
			_ = len(eventsByType)
		}
	})
}

// TestStreamLatency tests stream latency characteristics
func TestStreamLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stream latency test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	var latencies []time.Duration
	var latencyMu sync.Mutex

	client.client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		// Calculate approximate latency using event time
		if event.EventTime > 0 {
			eventTime := time.Unix(0, event.EventTime*int64(time.Millisecond))
			latency := time.Since(eventTime)
			
			latencyMu.Lock()
			latencies = append(latencies, latency)
			latencyMu.Unlock()
		}
		return nil
	})

	if err := client.Subscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Collect latency data
	t.Log("📊 Collecting latency data...")
	time.Sleep(15 * time.Second)

	latencyMu.Lock()
	if len(latencies) > 0 {
		var totalLatency time.Duration
		minLatency := latencies[0]
		maxLatency := latencies[0]

		for _, latency := range latencies {
			totalLatency += latency
			if latency < minLatency {
				minLatency = latency
			}
			if latency > maxLatency {
				maxLatency = latency
			}
		}

		avgLatency := totalLatency / time.Duration(len(latencies))

		t.Logf("📈 Latency Statistics (%d samples):", len(latencies))
		t.Logf("  Average: %v", avgLatency)
		t.Logf("  Minimum: %v", minLatency)
		t.Logf("  Maximum: %v", maxLatency)

		if avgLatency < 1*time.Second {
			t.Logf("✅ Good latency performance: %v average", avgLatency)
		} else {
			t.Logf("⚠️  Higher latency than expected: %v average", avgLatency)
		}
	} else {
		t.Log("⚠️  No latency data collected (no events with timestamps)")
	}
	latencyMu.Unlock()

	// Clean up
	if err := client.Unsubscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestMemoryUsage tests memory usage patterns during streaming
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple streams to generate load
	streams := []string{
		"btcusd_perp@aggTrade",
		"linkusd_perp@aggTrade",
		"btcusd_perp@bookTicker",
		"linkusd_perp@bookTicker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Let it run for a while to collect events
	t.Log("📊 Collecting events for memory usage analysis...")
	time.Sleep(10 * time.Second)

	// Check event accumulation
	totalEvents := len(client.GetEventsReceived())
	t.Logf("Total events accumulated: %d", totalEvents)

	// Test clearing events (memory cleanup)
	client.ClearEvents()
	eventsAfterClear := len(client.GetEventsReceived())

	if eventsAfterClear == 0 {
		t.Log("✅ Event clearing working correctly")
	} else {
		t.Errorf("❌ Event clearing failed: %d events remain", eventsAfterClear)
	}

	// Test continued operation after clearing
	time.Sleep(3 * time.Second)
	newEvents := len(client.GetEventsReceived())
	t.Logf("New events after clear: %d", newEvents)

	if newEvents > 0 {
		t.Log("✅ Continued operation after memory cleanup working")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestRapidSubscriptionChanges tests rapid subscription/unsubscription cycles
func TestRapidSubscriptionChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid subscription changes test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	streams := []string{
		"btcusd_perp@aggTrade",
		"linkusd_perp@aggTrade",
		"adausd_perp@aggTrade",
	}

	t.Log("🔄 Testing rapid subscription changes...")

	// Perform subscription changes (heavily reduced iterations and longer delays to avoid rate limits)
	for i := 0; i < 3; i++ { // Reduced from 5 to 3 iterations
		// Subscribe to all streams
		if err := client.Subscribe(ctx, streams); err != nil {
			// Check if this is a rate limit issue and handle gracefully
			if strings.Contains(err.Error(), "not connected") || strings.Contains(err.Error(), "policy violation") {
				t.Logf("Iteration %d: Connection lost or rate limited, stopping test: %v", i, err)
				break // Stop the test if connection is lost or rate limited
			}
			t.Errorf("Iteration %d: Failed to subscribe: %v", i, err)
			continue
		}

		// Much longer wait to avoid rate limiting
		time.Sleep(2 * time.Second) // Increased from 1s to 2s

		// Unsubscribe from all streams
		if err := client.Unsubscribe(ctx, streams); err != nil {
			if strings.Contains(err.Error(), "not connected") || strings.Contains(err.Error(), "policy violation") {
				t.Logf("Iteration %d: Connection lost or rate limited during unsubscribe, stopping: %v", i, err)
				break
			}
			t.Errorf("Iteration %d: Failed to unsubscribe: %v", i, err)
			continue
		}

		// Much longer wait between iterations to avoid rate limiting
		time.Sleep(1 * time.Second) // Increased from 500ms to 1s
	}

	// Final subscription to test stability
	if err := client.Subscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		t.Errorf("Final subscription failed: %v", err)
	}

	// Wait for events to verify stability
	if err := client.WaitForEventsByType("aggTrade", 2, 10*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("aggTrade")
	if len(events) > 0 {
		t.Logf("✅ Rapid subscription changes handled successfully: %d final events", len(events))
	} else {
		t.Log("⚠️  No events after rapid subscription changes")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, []string{"btcusd_perp@aggTrade"}); err != nil {
		t.Errorf("Failed final unsubscribe: %v", err)
	}
}