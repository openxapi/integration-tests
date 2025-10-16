package streamstest

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	optionsstreams "github.com/openxapi/binance-go/ws/options-streams"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
	Name        string
	Description string
}

// SharedClientManager manages shared WebSocket clients across tests
type SharedClientManager struct {
	clients   map[string]*optionsstreams.Client
	mutex     sync.RWMutex
	cleanupFn func()
}

var (
	sharedClients *SharedClientManager
	once          sync.Once
)

// initSharedClients initializes the shared client manager
func initSharedClients() {
	once.Do(func() {
		sharedClients = &SharedClientManager{
			clients: make(map[string]*optionsstreams.Client),
		}

		// Register cleanup function to clear shared client references at program exit
		sharedClients.cleanupFn = func() {
			sharedClients.mutex.Lock()
			defer sharedClients.mutex.Unlock()

			// Note: SDK client has no Disconnect method; tests handle disconnect via channels.
			// Here we only clear references to allow GC; individual tests should close channels.
			for configName := range sharedClients.clients {
				delete(sharedClients.clients, configName)
			}
		}
	})
}

// getOrCreateSharedClient gets or creates a shared client for the given config
func getOrCreateSharedClient(t *testing.T, config TestConfig) *optionsstreams.Client {
	initSharedClients()

	sharedClients.mutex.RLock()
	client, exists := sharedClients.clients[config.Name]
	sharedClients.mutex.RUnlock()

	if exists && client != nil {
		return client
	}

	// Need to create a new client
	sharedClients.mutex.Lock()
	defer sharedClients.mutex.Unlock()

	// Double-check in case another goroutine created it
	if client, exists := sharedClients.clients[config.Name]; exists && client != nil {
		return client
	}

	// Create new client
	newClient, err := setupClient(config)
	if err != nil {
		t.Logf("Failed to setup shared client for %s: %v", config.Name, err)
		return nil
	}

	sharedClients.clients[config.Name] = newClient
	return newClient
}

// setupClient creates and configures a new client
func setupClient(config TestConfig) (*optionsstreams.Client, error) {
	client := optionsstreams.NewClient()
	// Rely on SDK's built-in default servers; do not manually set servers here
	if err := ensureDefaultServer(client); err != nil {
		return nil, fmt.Errorf("failed to use default server: %w", err)
	}
	return client, nil
}

// ensureDefaultServer sets up a default server if none is present
func ensureDefaultServer(client *optionsstreams.Client) error {
	// Do not manually add or select servers; SDK provides defaults.
	// If no active server is set by SDK, we still avoid overriding.
	// This function remains for compatibility but performs no configuration.
	_ = os.Getenv // keep import usage
	return nil
}

// disconnectAllSharedClients disconnects all shared clients
func disconnectAllSharedClients() {
	if sharedClients == nil {
		return
	}

	if sharedClients.cleanupFn != nil {
		sharedClients.cleanupFn()
	}
}

// getTestConfigs returns all available test configurations
func getTestConfigs() []TestConfig {
	configs := []TestConfig{
		{
			Name:        "Public-NoAuth",
			Description: "Test public endpoints that don't require authentication",
		},
	}
	return configs
}

// StreamTestClient wraps the options-streams client for testing
type StreamTestClient struct {
	client *optionsstreams.Client
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

	// SDK channels
	combinedCh *optionsstreams.CombinedMarketStreamsChannel
	marketCh   *optionsstreams.MarketStreamsChannel
}

// NewStreamTestClient creates a new test client for Options streams using shared client
func NewStreamTestClient(t *testing.T, config TestConfig) (*StreamTestClient, error) {
	client := getOrCreateSharedClient(t, config)
	if client == nil {
		return nil, fmt.Errorf("failed to get shared client for config %s", config.Name)
	}

	return &StreamTestClient{
		client:         client,
		config:         config,
		eventsReceived: make([]interface{}, 0),
		activeStreams:  make([]string, 0),
		combinedCh:     optionsstreams.NewCombinedMarketStreamsChannel(client),
		marketCh:       optionsstreams.NewMarketStreamsChannel(client),
	}, nil
}

// NewStreamTestClientDedicated creates a dedicated (non-shared) test client for specific use cases
func NewStreamTestClientDedicated(config TestConfig) (*StreamTestClient, error) {
	client, err := setupClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dedicated client: %v", err)
	}

	return &StreamTestClient{
		client:         client,
		config:         config,
		eventsReceived: make([]interface{}, 0),
	}, nil
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

// GetResponseList returns all responses received by the client
func (stc *StreamTestClient) GetResponseList() []interface{} {
	// Fallback: return processed events as responses for debugging
	stc.eventsMu.RLock()
	defer stc.eventsMu.RUnlock()
	out := make([]interface{}, len(stc.eventsReceived))
	copy(out, stc.eventsReceived)
	return out
}

// ClearResponseList clears the client's response list
func (stc *StreamTestClient) ClearResponseList() {
	stc.ClearEvents()
}

// WaitForResponses waits for a specific number of responses or timeout
func (stc *StreamTestClient) WaitForResponses(count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		responses := stc.GetResponseList()
		if len(responses) >= count {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	responses := stc.GetResponseList()
	return fmt.Errorf("timeout waiting for responses: expected %d, got %d", count, len(responses))
}

// getTestConfig returns a basic test configuration
func getTestConfig() TestConfig {
	configs := getTestConfigs()
	if len(configs) > 0 {
		return configs[0] // Return the first available config
	}
	// Fallback
	return TestConfig{
		Name:        "Public-NoAuth",
		Description: "Test public endpoints that don't require authentication",
	}
}

// TB interface for both testing.T and testing.B
type TB interface {
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

// Helper function to create a test client using shared client pattern
func createTestClient(t TB) *StreamTestClient {
	config := getTestConfig()

	// We need a testing.T for the shared client, but we have TB interface
	// For now, create a dedicated client if we can't convert to testing.T
	if testingT, ok := t.(*testing.T); ok {
		client, err := NewStreamTestClient(testingT, config)
		if err != nil {
			t.Fatalf("Failed to create shared test client: %v", err)
		}
		return client
	}

	// Fallback to dedicated client for benchmarks or other TB implementations
	client, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("Failed to create dedicated test client: %v", err)
	}
	return client
}
