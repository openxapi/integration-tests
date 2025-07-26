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

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Test single stream subscription
	stream := "btcusdt@trade"
	if err := client.Subscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", stream, err)
	}

	// Verify subscription
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != 1 || activeStreams[0] != stream {
		t.Errorf("Expected 1 active stream [%s], got %v", stream, activeStreams)
	}

	// Wait for some events
	time.Sleep(3 * time.Second)

	// Test unsubscription
	if err := client.Unsubscribe(ctx, []string{stream}); err != nil {
		t.Fatalf("Failed to unsubscribe from %s: %v", stream, err)
	}

	// Verify unsubscription
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after unsubscribe, got %v", activeStreams)
	}
}

// TestMultipleStreamsSubscription tests subscribing to multiple streams
func TestMultipleStreamsSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple streams subscription test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple streams
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"btcusdt@depth5",
		"ethusdt@miniTicker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to multiple streams: %v", err)
	}

	// Verify subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != len(streams) {
		t.Errorf("Expected %d active streams, got %d", len(streams), len(activeStreams))
	}

	// Wait for events from multiple streams
	t.Log("Waiting for events from multiple streams...")
	if err := client.WaitForEvents(10, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsReceived()
	t.Logf("Received %d events from multiple streams", len(events))

	// Unsubscribe from all streams
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from multiple streams: %v", err)
	}

	// Verify all unsubscribed
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after unsubscribe all, got %v", activeStreams)
	}
}

// TestStreamUnsubscription tests partial unsubscription
func TestStreamUnsubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stream unsubscription test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to multiple streams
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade",
		"btcusdt@depth5",
		"ethusdt@miniTicker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to multiple streams: %v", err)
	}

	// Verify initial subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != len(streams) {
		t.Errorf("Expected %d active streams, got %d", len(streams), len(activeStreams))
	}

	// Unsubscribe from some streams
	partialUnsubscribe := []string{"btcusdt@trade", "ethusdt@trade"}
	if err := client.Unsubscribe(ctx, partialUnsubscribe); err != nil {
		t.Errorf("Failed to unsubscribe from partial streams: %v", err)
	}

	// Verify partial unsubscription
	activeStreams = client.GetActiveStreams()
	expectedRemainingCount := 2
	if len(activeStreams) != expectedRemainingCount {
		t.Errorf("Expected %d active streams after partial unsubscribe, got %d", expectedRemainingCount, len(activeStreams))
	}

	// Check remaining streams
	expectedRemainingStreams := []string{"btcusdt@depth5", "ethusdt@miniTicker"}
	for _, expected := range expectedRemainingStreams {
		found := false
		for _, active := range activeStreams {
			if active == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected stream %s to still be active", expected)
		}
	}

	// Unsubscribe from remaining streams
	if err := client.Unsubscribe(ctx, activeStreams); err != nil {
		t.Errorf("Failed to unsubscribe from remaining streams: %v", err)
	}

	// Verify all unsubscribed
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after unsubscribe all, got %v", activeStreams)
	}
}

// TestListSubscriptions tests listing active subscriptions
func TestListSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping list subscriptions test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to some streams
	streams := []string{
		"btcusdt@trade",
		"ethusdt@miniTicker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Test listing subscriptions
	if err := client.ListSubscriptions(ctx); err != nil {
		t.Errorf("Failed to list subscriptions: %v", err)
	}

	// Wait for subscription response
	time.Sleep(2 * time.Second)

	// Check if subscription response was received
	events := client.GetEventsByType("subscriptionResponse")
	if len(events) == 0 {
		t.Log("No subscription response received (this might be expected)")
	} else {
		t.Logf("Received %d subscription response events", len(events))
	}

	// Cleanup
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestSubscriptionToInvalidStream tests subscribing to invalid streams
func TestSubscriptionToInvalidStream(t *testing.T) {
	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Try to subscribe to invalid stream
	invalidStream := "invalid@stream"
	err := client.Subscribe(ctx, []string{invalidStream})
	
	// The subscription might not fail immediately, but we should get an error event
	if err != nil {
		t.Logf("Expected error for invalid stream: %v", err)
	}

	// Wait for possible error events
	time.Sleep(3 * time.Second)

	// Check for error events
	errorEvents := client.GetEventsByType("error")
	if len(errorEvents) > 0 {
		t.Logf("Received %d error events for invalid stream", len(errorEvents))
	}
}

// TestBatchSubscription tests subscribing to many streams at once
func TestBatchSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch subscription test in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Create a batch of streams
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
	}

	// Subscribe to all streams
	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to batch streams: %v", err)
	}

	// Verify all subscriptions
	activeStreams := client.GetActiveStreams()
	if len(activeStreams) != len(streams) {
		t.Errorf("Expected %d active streams, got %d", len(streams), len(activeStreams))
	}

	// Wait for events from all streams
	t.Log("Waiting for events from batch streams...")
	if err := client.WaitForEvents(20, 30*time.Second); err != nil {
		t.Logf("Warning: %v", err)
	}

	// Check received events
	events := client.GetEventsReceived()
	t.Logf("Received %d events from batch streams", len(events))

	// Unsubscribe from all
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe from batch streams: %v", err)
	}

	// Verify all unsubscribed
	activeStreams = client.GetActiveStreams()
	if len(activeStreams) != 0 {
		t.Errorf("Expected 0 active streams after batch unsubscribe, got %v", activeStreams)
	}
}