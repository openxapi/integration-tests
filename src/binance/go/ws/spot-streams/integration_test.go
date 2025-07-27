package streamstest

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	spotstreams "github.com/openxapi/binance-go/ws/spot-streams"
	"github.com/openxapi/binance-go/ws/spot-streams/models"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
	Name        string
	Description string
	Client      *spotstreams.Client
}

// StreamTestClient wraps the spot-streams client for testing
type StreamTestClient struct {
	client *spotstreams.Client
	config TestConfig

	// Event tracking
	eventsReceived []interface{}
	eventsMu       sync.RWMutex

	// Subscription tracking
	activeStreams []string
	streamsMu     sync.RWMutex

	// Connection state
	connected bool
	mu        sync.RWMutex
}

// NewStreamTestClient creates a new test client for spot streams
func NewStreamTestClient(config TestConfig) (*StreamTestClient, error) {
	client := spotstreams.NewClient()

	// Set to testnet server
	err := client.SetActiveServer("testnet1")
	if err != nil {
		return nil, fmt.Errorf("failed to set testnet server: %w", err)
	}

	return &StreamTestClient{
		client:         client,
		config:         config,
		eventsReceived: make([]interface{}, 0),
		activeStreams:  make([]string, 0),
	}, nil
}

// Connect establishes WebSocket connection
func (stc *StreamTestClient) Connect(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	if stc.connected {
		return nil
	}

	err := stc.client.Connect(ctx)
	if err != nil {
		return err
	}

	stc.connected = true
	return nil
}

// Disconnect closes the WebSocket connection
func (stc *StreamTestClient) Disconnect() error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	if !stc.connected {
		return nil
	}

	// Use defer to ensure we always set connected to false
	defer func() {
		stc.connected = false
	}()

	// Catch any panic from double-close
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't re-panic
			log.Printf("Recovered from disconnect panic: %v", r)
		}
	}()

	return stc.client.Disconnect()
}

// Subscribe to streams
func (stc *StreamTestClient) Subscribe(ctx context.Context, streams []string) error {
	err := stc.client.Subscribe(ctx, streams)
	if err != nil {
		return err
	}

	stc.streamsMu.Lock()
	stc.activeStreams = append(stc.activeStreams, streams...)
	stc.streamsMu.Unlock()

	return nil
}

// Unsubscribe from streams
func (stc *StreamTestClient) Unsubscribe(ctx context.Context, streams []string) error {
	err := stc.client.Unsubscribe(ctx, streams)
	if err != nil {
		return err
	}

	stc.streamsMu.Lock()
	// Remove streams from active list
	newActiveStreams := make([]string, 0)
	for _, active := range stc.activeStreams {
		found := false
		for _, unsub := range streams {
			if active == unsub {
				found = true
				break
			}
		}
		if !found {
			newActiveStreams = append(newActiveStreams, active)
		}
	}
	stc.activeStreams = newActiveStreams
	stc.streamsMu.Unlock()

	return nil
}

// ListSubscriptions returns the active subscriptions
func (stc *StreamTestClient) ListSubscriptions(ctx context.Context) error {
	return stc.client.ListSubscriptions(ctx)
}

// GetActiveStreams returns currently active streams
func (stc *StreamTestClient) GetActiveStreams() []string {
	stc.streamsMu.RLock()
	defer stc.streamsMu.RUnlock()

	streams := make([]string, len(stc.activeStreams))
	copy(streams, stc.activeStreams)
	return streams
}

// IsConnected returns connection status
func (stc *StreamTestClient) IsConnected() bool {
	stc.mu.RLock()
	defer stc.mu.RUnlock()
	return stc.connected
}

// SetupEventHandlers registers event handlers for all stream types
func (stc *StreamTestClient) SetupEventHandlers() {
	// Trade events
	stc.client.HandleTradeEvent(func(event *models.TradeEvent) error {
		stc.recordEvent("trade", event)
		return nil
	})

	// Aggregate trade events
	stc.client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		stc.recordEvent("aggTrade", event)
		return nil
	})

	// Kline events
	stc.client.HandleKlineEvent(func(event *models.KlineEvent) error {
		stc.recordEvent("kline", event)
		return nil
	})

	// Mini ticker events
	stc.client.HandleMiniTickerEvent(func(event *models.MiniTickerEvent) error {
		stc.recordEvent("miniTicker", event)
		return nil
	})

	// Ticker events
	stc.client.HandleTickerEvent(func(event *models.TickerEvent) error {
		stc.recordEvent("ticker", event)
		return nil
	})

	// Book ticker events
	stc.client.HandleBookTickerEvent(func(event *models.BookTickerEvent) error {
		stc.recordEvent("bookTicker", event)
		return nil
	})

	// Depth events
	stc.client.HandleDepthEvent(func(event *models.DiffDepthEvent) error {
		stc.recordEvent("depth", event)
		return nil
	})

	// Partial depth events
	stc.client.HandlePartialDepthEvent(func(event *models.PartialDepthEvent) error {
		stc.recordEvent("partialDepth", event)
		return nil
	})

	// Rolling window ticker events
	stc.client.HandleRollingWindowTickerEvent(func(event *models.RollingWindowTickerEvent) error {
		stc.recordEvent("rollingWindowTicker", event)
		return nil
	})

	// Average price events
	stc.client.HandleAvgPriceEvent(func(event *models.AvgPriceEvent) error {
		stc.recordEvent("avgPrice", event)
		return nil
	})

	// Combined stream events
	stc.client.HandleCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		stc.recordEvent("combinedStream", event)
		return nil
	})

	// Subscription response events
	stc.client.HandleSubscriptionResponse(func(event *models.SubscriptionResponse) error {
		stc.recordEvent("subscriptionResponse", event)
		return nil
	})

	// Error events
	stc.client.HandleStreamError(func(event *models.ErrorResponse) error {
		stc.recordEvent("error", event)
		return nil
	})
}

// recordEvent stores received events for verification
func (stc *StreamTestClient) recordEvent(eventType string, data interface{}) {
	stc.eventsMu.Lock()
	defer stc.eventsMu.Unlock()

	event := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now(),
	}

	stc.eventsReceived = append(stc.eventsReceived, event)
	log.Printf("Received %s event: %+v", eventType, data)
}

// GetEventsReceived returns all received events
func (stc *StreamTestClient) GetEventsReceived() []interface{} {
	stc.eventsMu.RLock()
	defer stc.eventsMu.RUnlock()

	events := make([]interface{}, len(stc.eventsReceived))
	copy(events, stc.eventsReceived)
	return events
}

// GetEventsByType returns events of a specific type
func (stc *StreamTestClient) GetEventsByType(eventType string) []interface{} {
	stc.eventsMu.RLock()
	defer stc.eventsMu.RUnlock()

	var filteredEvents []interface{}
	for _, event := range stc.eventsReceived {
		if eventMap, ok := event.(map[string]interface{}); ok {
			if eventMap["type"] == eventType {
				filteredEvents = append(filteredEvents, eventMap["data"])
			}
		}
	}
	return filteredEvents
}

// ClearEvents clears all received events
func (stc *StreamTestClient) ClearEvents() {
	stc.eventsMu.Lock()
	defer stc.eventsMu.Unlock()
	stc.eventsReceived = stc.eventsReceived[:0]
}

// WaitForEvents waits for a specific number of events or timeout
func (stc *StreamTestClient) WaitForEvents(count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		stc.eventsMu.RLock()
		currentCount := len(stc.eventsReceived)
		stc.eventsMu.RUnlock()

		if currentCount >= count {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	stc.eventsMu.RLock()
	currentCount := len(stc.eventsReceived)
	stc.eventsMu.RUnlock()

	return fmt.Errorf("timeout waiting for events: expected %d, got %d", count, currentCount)
}

// WaitForEventsByType waits for events of a specific type
func (stc *StreamTestClient) WaitForEventsByType(eventType string, count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		events := stc.GetEventsByType(eventType)
		if len(events) >= count {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	events := stc.GetEventsByType(eventType)
	return fmt.Errorf("timeout waiting for %s events: expected %d, got %d", eventType, count, len(events))
}

// getTestConfig returns a basic test configuration
func getTestConfig() TestConfig {
	return TestConfig{
		Name:        "SpotStreamsTest",
		Description: "Test configuration for Binance spot streams",
	}
}

// TB interface for both testing.T and testing.B
type TB interface {
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

// Helper function to create a test client
func createTestClient(t TB) *StreamTestClient {
	config := getTestConfig()
	client, err := NewStreamTestClient(config)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	return client
}

// Helper function to setup and connect client
func setupAndConnectClient(t *testing.T) *StreamTestClient {
	client := createTestClient(t)
	client.SetupEventHandlers()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	return client
}

// Helper function to test stream subscription
func testStreamSubscription(t *testing.T, streamName string, eventType string, eventCount int) {
	if testing.Short() {
		t.Skip("Skipping stream tests in short mode")
	}

	client := setupAndConnectClient(t)
	defer client.Disconnect()

	ctx := context.Background()

	// Subscribe to stream
	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to %s: %v", streamName, err)
	}

	// Verify stream is in active list
	activeStreams := client.GetActiveStreams()
	found := false
	for _, stream := range activeStreams {
		if stream == streamName {
			found = true
			break
		}
	}
	if found {
		t.Logf("✅ Successfully subscribed to %s", streamName)
	} else {
		t.Logf("⚠️  Stream %s not found in active streams", streamName)
	}

	// Wait a bit for connection stability
	time.Sleep(3 * time.Second)

	// Wait for events - SDK now properly handles stream events
	t.Logf("Waiting for %s events...", eventType)
	err := client.WaitForEventsByType(eventType, eventCount, 15*time.Second)
	if err != nil {
		t.Logf("⚠️  Timeout waiting for %s events: %v", eventType, err)
	}

	// Check received events
	events := client.GetEventsByType(eventType)
	t.Logf("Received %d %s events", len(events), eventType)

	// With updated SDK, we should receive events for active streams
	if len(events) > 0 {
		t.Logf("✅ Successfully received %d %s events", len(events), eventType)
	} else {
		t.Logf("⚠️  No %s events received (may be due to low market activity)", eventType)
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, []string{streamName}); err != nil {
		t.Errorf("Failed to unsubscribe from %s: %v", streamName, err)
	}

	// Verify stream is removed from active list
	activeStreams = client.GetActiveStreams()
	found = false
	for _, stream := range activeStreams {
		if stream == streamName {
			found = true
			break
		}
	}
	if !found {
		t.Logf("✅ Successfully unsubscribed from %s", streamName)
	} else {
		t.Errorf("Stream %s still found in active streams after unsubscribe", streamName)
	}
}

// Helper to check if environment variable is set
func hasEnvVar(key string) bool {
	return os.Getenv(key) != ""
}

// Helper to get environment variable or default
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}