package streamstest

import (
	"context"
	"testing"
	"time"
)

// TestSubscriptionManagement tests basic subscription management
func TestSubscriptionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping subscription management test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Test single subscription
	stream := "btcusdt@aggTrade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Verify subscription
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != 1 || activeStreams[0] != stream {
		t.Errorf("Expected 1 active stream (%s), got %v", stream, activeStreams)
	}

	t.Logf("✅ Successfully subscribed to %s", stream)

	// Test list subscriptions
	if err := client.ListSubscriptions(ctx); err != nil {
		t.Errorf("Failed to list subscriptions: %v", err)
	}

	// Wait for some events
	if err := client.WaitForEventsByType("aggTrade", 2, 10*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Test unsubscription
	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
	}

	// Verify unsubscription
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after unsubscribe, got %v", activeStreams)
	}

	t.Logf("✅ Successfully unsubscribed from %s", stream)
}

// TestMultipleStreamsSubscription tests subscribing to multiple streams
func TestMultipleStreamsSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple streams subscription test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple streams at once
	streams := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"btcusdt@ticker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to multiple streams: %v", err)
	}

	// Verify all subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != len(streams) {
		t.Errorf("Expected %d active streams, got %d", len(streams), len(activeStreams))
	}

	t.Logf("✅ Successfully subscribed to %d streams", len(streams))

	// Wait for events from different streams
	time.Sleep(5 * time.Second)

	// Check for events
	aggTradeEvents := client.GetEventsByType("aggTrade")
	tickerEvents := client.GetEventsByType("ticker")

	t.Logf("Received %d aggTrade events and %d ticker events", len(aggTradeEvents), len(tickerEvents))

	// Unsubscribe from all
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from multiple streams: %v", err)
	}

	// Verify all unsubscriptions
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after unsubscribe, got %v", activeStreams)
	}

	t.Log("✅ Successfully unsubscribed from all streams")
}

// TestStreamUnsubscription tests partial unsubscription
func TestStreamUnsubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stream unsubscription test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Subscribe to multiple streams
	streams := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"btcusdt@ticker",
		"ethusdt@ticker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Verify all subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != len(streams) {
		t.Errorf("Expected %d active streams, got %d", len(streams), len(activeStreams))
	}

	// Unsubscribe from some streams
	unsubscribeStreams := []string{
		"btcusdt@aggTrade",
		"ethusdt@ticker",
	}

	if err := client.Unsubscribe(ctx, unsubscribeStreams); err != nil {
		t.Errorf("Failed to unsubscribe from partial streams: %v", err)
	}

	// Verify partial unsubscription
	activeStreams = client.GetActiveStreams()
	expectedRemaining := 2
	if len(activeStreams) != expectedRemaining {
		t.Errorf("Expected %d active streams after partial unsubscribe, got %d", expectedRemaining, len(activeStreams))
	}

	// Check that the correct streams remain
	remainingStreams := map[string]bool{
		"ethusdt@aggTrade": false,
		"btcusdt@ticker":   false,
	}

	for _, stream := range activeStreams {
		if _, exists := remainingStreams[stream]; exists {
			remainingStreams[stream] = true
		}
	}

	for stream, found := range remainingStreams {
		if !found {
			t.Errorf("Expected stream %s to remain active", stream)
		}
	}

	t.Log("✅ Partial unsubscription successful")

	// Clean up - unsubscribe from remaining streams
	remainingStreamsList := make([]string, 0, len(remainingStreams))
	for stream := range remainingStreams {
		remainingStreamsList = append(remainingStreamsList, stream)
	}

	if err := client.Unsubscribe(ctx, remainingStreamsList); err != nil {
		t.Errorf("Failed to unsubscribe from remaining streams: %v", err)
	}
}

// TestResubscription tests resubscribing to the same stream
func TestResubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resubscription test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()
	stream := "btcusdt@aggTrade"

	// First subscription
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe initially: %v", err)
	}

	// Wait for events
	if err := client.WaitForEventsByType("aggTrade", 2, 10*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}

	// Clear events
	client.ClearEvents()

	// Resubscribe to the same stream
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to resubscribe: %v", err)
	}

	// Verify subscription is active
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != 1 || activeStreams[0] != stream {
		t.Errorf("Resubscription failed: expected 1 active stream (%s), got %v", stream, activeStreams)
	}

	// Wait for events from resubscription
	if err := client.WaitForEventsByType("aggTrade", 2, 10*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	events := client.GetEventsByType("aggTrade")
	if len(events) > 0 {
		t.Logf("✅ Resubscription successful: received %d events", len(events))
	}

	// Clean up
	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Errorf("Failed to unsubscribe after resubscription test: %v", err)
	}
}

// TestBatchSubscription tests subscribing to streams in batches
func TestBatchSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch subscription test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// First batch
	batch1 := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
	}

	if err := client.Subscribe(ctx, batch1); err != nil {
		t.Fatalf("Failed to subscribe to first batch: %v", err)
	}

	// Second batch
	batch2 := []string{
		"btcusdt@ticker",
		"ethusdt@ticker",
	}

	if err := client.Subscribe(ctx, batch2); err != nil {
		t.Errorf("Failed to subscribe to second batch: %v", err)
	}

	// Verify all subscriptions
	activeStreams := client.GetActiveStreams()
	expectedTotal := len(batch1) + len(batch2)
	if len(activeStreams) != expectedTotal {
		t.Errorf("Expected %d active streams, got %d", expectedTotal, len(activeStreams))
	}

	t.Logf("✅ Batch subscription successful: %d total streams", len(activeStreams))

	// Wait for events from different types
	time.Sleep(5 * time.Second)

	// Check event counts
	aggTradeEvents := client.GetEventsByType("aggTrade")
	tickerEvents := client.GetEventsByType("ticker")

	t.Logf("Received %d aggTrade events and %d ticker events from batch subscriptions", 
		len(aggTradeEvents), len(tickerEvents))

	// Clean up all subscriptions
	allStreams := append(batch1, batch2...)
	if err := client.Unsubscribe(ctx, allStreams); err != nil {
		t.Errorf("Failed to unsubscribe from all batch streams: %v", err)
	}
}

// TestSubscriptionTracking tests internal subscription tracking
func TestSubscriptionTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping subscription tracking test in short mode")
	}

	client, isDedicated := setupTestClient(t)
	if isDedicated {
		defer client.Disconnect()
	}

	ctx := context.Background()

	// Start with no subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 initial active streams, got %d", len(activeStreams))
	}

	// Add subscriptions one by one and verify tracking
	streams := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"btcusdt@ticker",
	}

	for i, stream := range streams {
		if err := client.Subscribe(ctx, []string{stream}); err != nil {
			t.Errorf("Failed to subscribe to %s: %v", stream, err)
			continue
		}

		activeStreams = client.GetActiveStreams()
		expectedCount := i + 1
		if len(activeStreams) != expectedCount {
			t.Errorf("After subscribing to %s, expected %d active streams, got %d", 
				stream, expectedCount, len(activeStreams))
		}
	}

	// Remove subscriptions one by one and verify tracking
	for i, stream := range streams {
		if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
			t.Errorf("Failed to unsubscribe from %s: %v", stream, err)
			continue
		}

		activeStreams = client.GetActiveStreams()
		expectedCount := len(streams) - i - 1
		if len(activeStreams) != expectedCount {
			t.Errorf("After unsubscribing from %s, expected %d active streams, got %d", 
				stream, expectedCount, len(activeStreams))
		}
	}

	// Verify all streams are removed
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 final active streams, got %d", len(activeStreams))
	}

	t.Log("✅ Subscription tracking working correctly")
}