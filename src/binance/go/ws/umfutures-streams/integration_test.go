package streamstest

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
	"github.com/openxapi/binance-go/ws/umfutures-streams/models"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
	Name        string
	Description string
	Client      *umfuturesstreams.Client
}

// StreamTestClient wraps the umfutures-streams client for testing
type StreamTestClient struct {
	client *umfuturesstreams.Client
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

// NewStreamTestClient creates a new test client for USD-M futures streams
func NewStreamTestClient(config TestConfig) (*StreamTestClient, error) {
	client := umfuturesstreams.NewClient()

	// Set to testnet server
	err := client.SetActiveServer("testnet1")
	if err != nil {
		return nil, fmt.Errorf("failed to set testnet server: %v", err)
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

// ConnectToCombinedStreams establishes WebSocket connection to combined streams endpoint
func (stc *StreamTestClient) ConnectToCombinedStreams(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	if stc.connected {
		return nil
	}

	err := stc.client.ConnectToCombinedStreams(ctx, "")
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
		// Check if the error indicates a closed connection
		if strings.Contains(err.Error(), "websocket not connected") || 
		   strings.Contains(err.Error(), "policy violation") {
			stc.mu.Lock()
			stc.connected = false
			stc.mu.Unlock()
		}
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
		// Check if the error indicates a closed connection
		if strings.Contains(err.Error(), "websocket not connected") || 
		   strings.Contains(err.Error(), "policy violation") {
			stc.mu.Lock()
			stc.connected = false
			stc.mu.Unlock()
			// If connection is closed, still update our local state
		} else {
			return err
		}
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
	stc.setupEventHandlers(true)
}

// SetupEventHandlersForCombinedStreams registers event handlers without combined stream handler
// This allows individual event handlers to be called when using combined streams connection
func (stc *StreamTestClient) SetupEventHandlersForCombinedStreams() {
	stc.setupEventHandlers(false)
}

// setupEventHandlers is the internal method that sets up event handlers
func (stc *StreamTestClient) setupEventHandlers(includeCombinedHandler bool) {
	// Aggregate trade events
	stc.client.OnAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		stc.recordEvent("aggTrade", event)
		return nil
	})

	// Mark price events (futures-specific)
	stc.client.OnMarkPriceEvent(func(event *models.MarkPriceEvent) error {
		stc.recordEvent("markPrice", event)
		return nil
	})

	// Kline events
	stc.client.OnKlineEvent(func(event *models.KlineEvent) error {
		stc.recordEvent("kline", event)
		return nil
	})

	// Continuous kline events (futures-specific)
	stc.client.OnContinuousKlineEvent(func(event *models.ContinuousKlineEvent) error {
		stc.recordEvent("continuousKline", event)
		return nil
	})

	// Mini ticker events
	stc.client.OnMiniTickerEvent(func(event *models.MiniTickerEvent) error {
		stc.recordEvent("miniTicker", event)
		return nil
	})

	// Ticker events
	stc.client.OnTickerEvent(func(event *models.TickerEvent) error {
		stc.recordEvent("ticker", event)
		return nil
	})

	// Book ticker events
	stc.client.OnBookTickerEvent(func(event *models.BookTickerEvent) error {
		stc.recordEvent("bookTicker", event)
		return nil
	})

	// Liquidation events (futures-specific)
	stc.client.OnLiquidationEvent(func(event *models.LiquidationEvent) error {
		stc.recordEvent("forceOrder", event)
		return nil
	})

	// Diff depth events
	stc.client.OnDiffDepthEvent(func(event *models.DiffDepthEvent) error {
		stc.recordEvent("depthUpdate", event)
		return nil
	})

	// Composite index events (futures-specific)
	stc.client.OnCompositeIndexEvent(func(event *models.CompositeIndexEvent) error {
		stc.recordEvent("compositeIndex", event)
		return nil
	})

	// Contract info events (futures-specific)
	stc.client.OnContractInfoEvent(func(event *models.ContractInfoEvent) error {
		stc.recordEvent("contractInfo", event)
		return nil
	})

	// Asset index events (futures-specific)
	stc.client.OnAssetIndexEvent(func(event *models.AssetIndexEvent) error {
		stc.recordEvent("assetIndexUpdate", event)
		return nil
	})

	// Subscription response events
	stc.client.OnSubscriptionResponse(func(event *models.SubscriptionResponse) error {
		stc.recordEvent("subscriptionResponse", event)
		return nil
	})

	// Error events
	stc.client.OnStreamError(func(event *models.ErrorResponse) error {
		stc.recordEvent("error", event)
		return nil
	})

	// Combined stream events (only register if requested)
	if includeCombinedHandler {
		stc.client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
			stc.recordEvent("combinedStream", event)
			return nil
		})
	}
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
		Name:        "UMFuturesStreamsTest",
		Description: "Test configuration for Binance USD-M futures streams",
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

// Helper function to setup and connect client to combined streams
func setupAndConnectCombinedStreamsClient(t *testing.T) *StreamTestClient {
	client := createTestClient(t)
	client.SetupEventHandlersForCombinedStreams() // Don't register combined handler to allow individual handlers

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
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

	// Wait for events
	t.Logf("Waiting for %s events...", eventType)
	err := client.WaitForEventsByType(eventType, eventCount, 15*time.Second)
	if err != nil {
		t.Fatalf("Failed to receive %s events: %v", eventType, err)
	}

	// Check received events
	events := client.GetEventsByType(eventType)
	t.Logf("Received %d %s events", len(events), eventType)

	// Fail if no events received
	if len(events) == 0 {
		t.Fatalf("No %s events received - stream may not be working properly", eventType)
	}

	t.Logf("✅ Successfully received %d %s events", len(events), eventType)

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