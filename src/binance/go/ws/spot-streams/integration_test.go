package streamstest

import (
	"fmt"
	"os"
	"sync"
	"testing"

	spotstreams "github.com/openxapi/binance-go/ws/spot-streams"
)

// TestConfig holds configuration for different test scenarios
type TestConfig struct {
	Name        string
	Description string
}

// SharedClientManager manages shared WebSocket clients across tests
type SharedClientManager struct {
	clients   map[string]*spotstreams.Client
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
		sharedClients = &SharedClientManager{clients: make(map[string]*spotstreams.Client)}
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
func getOrCreateSharedClient(t *testing.T, config TestConfig) *spotstreams.Client {
	initSharedClients()
	sharedClients.mutex.RLock()
	c, ok := sharedClients.clients[config.Name]
	sharedClients.mutex.RUnlock()
	if ok && c != nil {
		return c
	}
	sharedClients.mutex.Lock()
	defer sharedClients.mutex.Unlock()
	if c, ok := sharedClients.clients[config.Name]; ok && c != nil {
		return c
	}
	nc, err := setupClient(config)
	if err != nil {
		t.Logf("Failed to setup shared client for %s: %v", config.Name, err)
		return nil
	}
	sharedClients.clients[config.Name] = nc
	return nc
}

// setupClient creates and configures a new client
func setupClient(config TestConfig) (*spotstreams.Client, error) {
	client := spotstreams.NewClient()
	if err := ensureDefaultServer(client); err != nil {
		return nil, fmt.Errorf("failed to select default server: %w", err)
	}
	return client, nil
}

// ensureDefaultServer selects testnet unless overridden via BINANCE_SPOT_WS_SERVER
func ensureDefaultServer(client *spotstreams.Client) error {
	if v := os.Getenv("BINANCE_SPOT_WS_SERVER"); v != "" {
		if err := client.AddOrUpdateServer("override", v, "override", "override"); err == nil {
			if err2 := client.SetActiveServer("override"); err2 != nil {
				return fmt.Errorf("failed to set override server active: %w", err2)
			}
			return nil
		}
	}
	if err := client.SetActiveServer("testnet"); err != nil {
		return fmt.Errorf("failed to select testnet server 'testnet': %w", err)
	}
	return nil
}

// disconnectAllSharedClients clears shared references (channels close per-suite)
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
	return []TestConfig{{Name: "Public-NoAuth", Description: "Test public endpoints that don't require authentication"}}
}

// StreamTestClient wraps the spot-streams client for testing
type StreamTestClient struct {
	client *spotstreams.Client
	config TestConfig
}

// NewStreamTestClient creates a new test client using shared client
func NewStreamTestClient(t *testing.T, config TestConfig) (*StreamTestClient, error) {
	client := getOrCreateSharedClient(t, config)
	if client == nil {
		return nil, fmt.Errorf("failed to get shared client for %s", config.Name)
	}
	return &StreamTestClient{client: client, config: config}, nil
}

// NewStreamTestClientDedicated creates a dedicated (non-shared) client
func NewStreamTestClientDedicated(config TestConfig) (*StreamTestClient, error) {
	client, err := setupClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dedicated client: %v", err)
	}
	return &StreamTestClient{client: client, config: config}, nil
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
		if err != nil {
			t.Fatalf("failed to create shared test client: %v", err)
		}
		return c
	}
	c, err := NewStreamTestClientDedicated(cfg)
	if err != nil {
		t.Fatalf("failed to create dedicated test client: %v", err)
	}
	return c
}

// getTestConfig returns a basic test configuration
func getTestConfig() TestConfig {
	cfgs := getTestConfigs()
	if len(cfgs) > 0 {
		return cfgs[0]
	}
	return TestConfig{Name: "Public-NoAuth", Description: "Test public endpoints that don't require authentication"}
}
