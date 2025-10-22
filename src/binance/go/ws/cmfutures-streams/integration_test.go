package streamstest

import (
    "fmt"
    "os"
    "sync"
    "testing"

    cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
    Name        string
    Description string
}

// SharedClientManager manages shared WebSocket clients across tests
type SharedClientManager struct {
    clients   map[string]*cmfuturesstreams.Client
    mutex     sync.RWMutex
    cleanupFn func()
}

var (
    sharedClients *SharedClientManager
    once          sync.Once
)

func initSharedClients() {
    once.Do(func() {
        sharedClients = &SharedClientManager{clients: make(map[string]*cmfuturesstreams.Client)}
        sharedClients.cleanupFn = func() {
            sharedClients.mutex.Lock()
            defer sharedClients.mutex.Unlock()
            for k := range sharedClients.clients { delete(sharedClients.clients, k) }
        }
    })
}

func getOrCreateSharedClient(t *testing.T, config TestConfig) *cmfuturesstreams.Client {
    initSharedClients()
    sharedClients.mutex.RLock()
    c, ok := sharedClients.clients[config.Name]
    sharedClients.mutex.RUnlock()
    if ok && c != nil { return c }

    sharedClients.mutex.Lock()
    defer sharedClients.mutex.Unlock()
    if c, ok := sharedClients.clients[config.Name]; ok && c != nil { return c }
    nc, err := setupClient(config)
    if err != nil { t.Logf("Failed to setup shared client for %s: %v", config.Name, err); return nil }
    sharedClients.clients[config.Name] = nc
    return nc
}

func setupClient(config TestConfig) (*cmfuturesstreams.Client, error) {
    client := cmfuturesstreams.NewClient()
    // Allow override via env, else use testnet
    if v := os.Getenv("BINANCE_CMFUTURES_WS_SERVER"); v != "" {
        if err := client.AddOrUpdateServer("override", v, "override", "override"); err == nil {
            if err2 := client.SetActiveServer("override"); err2 != nil { return nil, fmt.Errorf("failed to set override server: %w", err2) }
            return client, nil
        }
    }
    if err := client.SetActiveServer("testnet"); err != nil {
        return nil, fmt.Errorf("failed to select testnet server 'testnet': %w", err)
    }
    return client, nil
}

func disconnectAllSharedClients() {
    if sharedClients == nil { return }
    if sharedClients.cleanupFn != nil { sharedClients.cleanupFn() }
}

func getTestConfigs() []TestConfig {
    return []TestConfig{{
        Name:        "Public-NoAuth",
        Description: "Test public endpoints that don't require authentication",
    }}
}

// getTestConfig returns the default single test configuration for convenience
func getTestConfig() TestConfig {
    cfgs := getTestConfigs()
    if len(cfgs) > 0 {
        return cfgs[0]
    }
    return TestConfig{Name: "Public-NoAuth", Description: "Default config"}
}

// StreamTestClient wraps the client for test setup
type StreamTestClient struct {
    client *cmfuturesstreams.Client
    config TestConfig
}

func NewStreamTestClient(t *testing.T, config TestConfig) (*StreamTestClient, error) {
    client := getOrCreateSharedClient(t, config)
    if client == nil { return nil, fmt.Errorf("failed to get shared client for %s", config.Name) }
    return &StreamTestClient{client: client, config: config}, nil
}

func NewStreamTestClientDedicated(config TestConfig) (*StreamTestClient, error) {
    client, err := setupClient(config)
    if err != nil { return nil, fmt.Errorf("failed to setup dedicated client: %v", err) }
    return &StreamTestClient{client: client, config: config}, nil
}
