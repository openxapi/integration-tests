package streamstest

import (
    "fmt"
    "os"
    "sync"
    "testing"
    "time"

    umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
    Name        string
    Description string
}

// SharedClientManager manages shared WebSocket clients across tests
type SharedClientManager struct {
    clients   map[string]*umfuturesstreams.Client
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
            clients: make(map[string]*umfuturesstreams.Client),
        }
        // Clear references at program exit; individual tests close channels
        sharedClients.cleanupFn = func() {
            sharedClients.mutex.Lock()
            defer sharedClients.mutex.Unlock()
            for k := range sharedClients.clients {
                delete(sharedClients.clients, k)
            }
        }
    })
}

// getOrCreateSharedClient gets or creates a shared client for the given config
func getOrCreateSharedClient(t *testing.T, config TestConfig) *umfuturesstreams.Client {
    initSharedClients()
    sharedClients.mutex.RLock()
    c, ok := sharedClients.clients[config.Name]
    sharedClients.mutex.RUnlock()
    if ok && c != nil { return c }

    sharedClients.mutex.Lock()
    defer sharedClients.mutex.Unlock()
    if c, ok := sharedClients.clients[config.Name]; ok && c != nil { return c }
    nc, err := setupClient(config)
    if err != nil {
        t.Logf("Failed to setup shared client for %s: %v", config.Name, err)
        return nil
    }
    sharedClients.clients[config.Name] = nc
    return nc
}

// setupClient creates and configures a new client
func setupClient(config TestConfig) (*umfuturesstreams.Client, error) {
    client := umfuturesstreams.NewClient()
    if err := ensureDefaultServer(client); err != nil {
        return nil, fmt.Errorf("failed to select default server: %w", err)
    }
    return client, nil
}

// ensureDefaultServer selects testnet unless overridden
func ensureDefaultServer(client *umfuturesstreams.Client) error {
    if v := os.Getenv("BINANCE_UMFUTURES_WS_SERVER"); v != "" {
        if err := client.AddOrUpdateServer("override", v, "override", "override"); err == nil {
            if err2 := client.SetActiveServer("override"); err2 != nil {
                return fmt.Errorf("failed to set override server active: %w", err2)
            }
            return nil
        }
    }
    if err := client.SetActiveServer("testnet1"); err != nil {
        return fmt.Errorf("failed to select testnet server 'testnet1': %w", err)
    }
    return nil
}

// disconnectAllSharedClients disconnects all shared clients (ref clearing handled above)
func disconnectAllSharedClients() {
    if sharedClients == nil { return }
    if sharedClients.cleanupFn != nil { sharedClients.cleanupFn() }
}

// getTestConfigs returns all available test configurations
func getTestConfigs() []TestConfig {
    return []TestConfig{{
        Name:        "Public-NoAuth",
        Description: "Test public endpoints that don't require authentication",
    }}
}

// StreamTestClient wraps the umfutures-streams client for testing
type StreamTestClient struct {
    client     *umfuturesstreams.Client
    config     TestConfig
    combinedCh *umfuturesstreams.CombinedMarketStreamsChannel
    marketCh   *umfuturesstreams.MarketStreamsChannel

    eventsMu       sync.RWMutex
    eventsReceived []interface{}
}

// NewStreamTestClient creates a new test client using shared client
func NewStreamTestClient(t *testing.T, config TestConfig) (*StreamTestClient, error) {
    client := getOrCreateSharedClient(t, config)
    if client == nil { return nil, fmt.Errorf("failed to get shared client for %s", config.Name) }
    return &StreamTestClient{client: client, config: config, eventsReceived: make([]interface{}, 0)}, nil
}

// NewStreamTestClientDedicated creates a dedicated (non-shared) client
func NewStreamTestClientDedicated(config TestConfig) (*StreamTestClient, error) {
    client, err := setupClient(config)
    if err != nil { return nil, fmt.Errorf("failed to setup dedicated client: %v", err) }
    return &StreamTestClient{client: client, config: config, eventsReceived: make([]interface{}, 0)}, nil
}

// recordEvent stores received events for verification
func (stc *StreamTestClient) recordEvent(eventType string, data interface{}) {
    stc.eventsMu.Lock()
    stc.eventsReceived = append(stc.eventsReceived, map[string]interface{}{"type": eventType, "data": data, "ts": time.Now()})
    stc.eventsMu.Unlock()
}

// GetResponseList returns all captured events (for debugging)
func (stc *StreamTestClient) GetResponseList() []interface{} {
    stc.eventsMu.RLock()
    defer stc.eventsMu.RUnlock()
    out := make([]interface{}, len(stc.eventsReceived))
    copy(out, stc.eventsReceived)
    return out
}

// ClearResponseList clears captured events
func (stc *StreamTestClient) ClearResponseList() {
    stc.eventsMu.Lock()
    stc.eventsReceived = stc.eventsReceived[:0]
    stc.eventsMu.Unlock()
}

// getTestConfig returns a basic test configuration
func getTestConfig() TestConfig {
    cfgs := getTestConfigs()
    if len(cfgs) > 0 { return cfgs[0] }
    return TestConfig{Name: "Public-NoAuth", Description: "Test public endpoints that don't require authentication"}
}

// TB interface for both testing.T and testing.B
type TB interface {
    Fatalf(format string, args ...interface{})
    Logf(format string, args ...interface{})
}

// Helper function to create a test client using shared client pattern
func createTestClient(t TB) *StreamTestClient {
    cfg := getTestConfig()
    if testingT, ok := t.(*testing.T); ok {
        c, err := NewStreamTestClient(testingT, cfg)
        if err != nil { t.Fatalf("failed to create shared test client: %v", err) }
        return c
    }
    c, err := NewStreamTestClientDedicated(cfg)
    if err != nil { t.Fatalf("failed to create dedicated test client: %v", err) }
    return c
}
