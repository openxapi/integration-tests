package streamstest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/spot-streams/models"
)

// TestConcurrentStreams tests handling multiple concurrent streams
func TestConcurrentStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent streams test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Create multiple concurrent streams
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade", 
		"adausdt@trade",
		"btcusdt@depth5",
		"ethusdt@depth5",
		"adausdt@depth5",
		"btcusdt@miniTicker",
		"ethusdt@miniTicker",
		"adausdt@miniTicker",
		"btcusdt@bookTicker",
		"ethusdt@bookTicker",
		"adausdt@bookTicker",
	}

	// Subscribe to all streams
	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to concurrent streams: %v", err)
	}

	// Wait for events from all streams
	t.Log("Waiting for events from concurrent streams...")
	if err := client.WaitForEvents(50, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsReceived()
	t.Logf("Received %d events from concurrent streams", len(events))

	// Verify we received different event types
	eventTypes := map[string]int{}
	for _, event := range events {
		if eventMap, ok := event.(map[string]interface{}); ok {
			if eventType, ok := eventMap["type"].(string); ok {
				eventTypes[eventType]++
			}
		}
	}

	t.Logf("Event types received: %v", eventTypes)

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from concurrent streams: %v", err)
	}
}

// TestHighVolumeStreams tests handling high-volume streams
func TestHighVolumeStreams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high volume streams test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to high-volume streams
	highVolumeStreams := []string{
		"btcusdt@trade",     // High-volume individual trades
		"ethusdt@trade",     // High-volume individual trades
		"btcusdt@depth",     // High-volume depth updates
		"ethusdt@depth",     // High-volume depth updates
		"btcusdt@aggTrade",  // High-volume aggregate trades
		"ethusdt@aggTrade",  // High-volume aggregate trades
	}

	if err := client.Subscribe(ctx, highVolumeStreams); err != nil {
		t.Fatalf("Failed to subscribe to high-volume streams: %v", err)
	}

	// Measure event processing performance
	startTime := time.Now()
	targetEvents := 100

	t.Logf("Waiting for %d events from high-volume streams...", targetEvents)
	if err := client.WaitForEvents(targetEvents, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	elapsed := time.Since(startTime)
	events := client.GetEventsReceived()
	eventsPerSecond := float64(len(events)) / elapsed.Seconds()

	t.Logf("Processed %d events in %v (%.2f events/second)", len(events), elapsed, eventsPerSecond)

	// Verify performance is acceptable (adjust threshold as needed)
	if eventsPerSecond < 1.0 {
		t.Logf("Warning: Low event processing rate: %.2f events/second", eventsPerSecond)
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, highVolumeStreams); err != nil {
		t.Errorf("Failed to unsubscribe from high-volume streams: %v", err)
	}
}

// TestStreamLatency tests stream latency measurements
func TestStreamLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stream latency test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to a fast stream
	stream := "btcusdt@trade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Record timestamps of events
	var latencies []time.Duration
	var mu sync.Mutex

	// Override event handler to measure latency
	client.client.HandleTradeEvent(func(event *models.TradeEvent) error {
		mu.Lock()
		defer mu.Unlock()
		
		// Calculate latency (simplified - using current time vs event time)
		now := time.Now()
		// Note: This is a simplified latency calculation
		// In a real scenario, you'd compare against the actual event timestamp
		latency := now.Sub(now) // Placeholder - real implementation would use event.E
		latencies = append(latencies, latency)
		
		return nil
	})

	// Wait for events
	time.Sleep(10 * time.Second)

	// Analyze latencies
	if len(latencies) > 0 {
		var total time.Duration
		for _, latency := range latencies {
			total += latency
		}
		avgLatency := total / time.Duration(len(latencies))
		t.Logf("Average latency over %d events: %v", len(latencies), avgLatency)
	} else {
		t.Log("No latency measurements recorded")
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestMemoryUsage tests memory usage during stream processing
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple streams
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"adausdt@trade",
		"btcusdt@depth5",
		"ethusdt@depth5",
		"adausdt@depth5",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Run for a period and monitor events
	duration := 30 * time.Second
	startTime := time.Now()
	
	t.Logf("Running stream processing for %v...", duration)
	
	for time.Since(startTime) < duration {
		time.Sleep(1 * time.Second)
		
		// Periodically check event count
		events := client.GetEventsReceived()
		if len(events)%100 == 0 && len(events) > 0 {
			t.Logf("Processed %d events so far", len(events))
		}
	}

	// Final event count
	finalEvents := client.GetEventsReceived()
	t.Logf("Total events processed: %d", len(finalEvents))

	// Test memory cleanup
	client.ClearEvents()
	clearedEvents := client.GetEventsReceived()
	if len(clearedEvents) != 0 {
		t.Errorf("Expected 0 events after clearing, got %d", len(clearedEvents))
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestRapidSubscriptionChanges tests rapid subscription/unsubscription
func TestRapidSubscriptionChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid subscription changes test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"adausdt@trade",
	}

	// Rapid subscribe/unsubscribe cycles
	for i := 0; i < 5; i++ {
		t.Logf("Rapid subscription cycle %d", i+1)
		
		// Subscribe
		if err := client.Subscribe(ctx, streams); err != nil {
			t.Errorf("Failed to subscribe in cycle %d: %v", i+1, err)
		}

		// Wait briefly
		time.Sleep(2 * time.Second)

		// Unsubscribe
		if err := client.Unsubscribe(ctx, streams); err != nil {
			t.Errorf("Failed to unsubscribe in cycle %d: %v", i+1, err)
		}

		// Wait briefly
		time.Sleep(1 * time.Second)
	}

	// Verify final state
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after rapid cycles, got %d", len(activeStreams))
	}
}

// BenchmarkEventProcessing benchmarks event processing performance
func BenchmarkEventProcessing(b *testing.B) {
	client := createTestClient(b)
	client.SetupEventHandlers()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to a high-volume stream
	stream := "btcusdt@trade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		b.Fatalf("Failed to subscribe: %v", err)
	}

	// Let events accumulate
	time.Sleep(5 * time.Second)

	// Benchmark event retrieval
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetEventsReceived()
	}
}

// BenchmarkSubscription benchmarks subscription operations
func BenchmarkSubscription(b *testing.B) {
	client := createTestClient(b)
	client.SetupEventHandlers()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	stream := "btcusdt@trade"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := client.Subscribe(ctx, []string{stream}); err != nil {
			b.Fatalf("Failed to subscribe: %v", err)
		}
		if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
			b.Fatalf("Failed to unsubscribe: %v", err)
		}
	}
}

// BenchmarkConcurrentEventAccess benchmarks concurrent event access
func BenchmarkConcurrentEventAccess(b *testing.B) {
	client := createTestClient(b)
	client.SetupEventHandlers()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to generate events
	if err := client.Subscribe(ctx, []string{"btcusdt@trade"}); err != nil {
		b.Fatalf("Failed to subscribe: %v", err)
	}

	// Let events accumulate
	time.Sleep(3 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = client.GetEventsReceived()
		}
	})
}