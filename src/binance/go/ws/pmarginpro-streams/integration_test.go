package streamstest

import (
	"fmt"
	"sync"
	"testing"

	pmarginpro "github.com/openxapi/binance-go/ws/pmarginpro-streams"
)

// TestConfig describes a test scenario (e.g., environment selection)
type TestConfig struct {
	Name        string
	Description string
}

// SharedClientManager manages cached WebSocket clients across tests
type SharedClientManager struct {
	clients   map[string]*pmarginpro.Client
	mutex     sync.RWMutex
	cleanupFn func()
}

var (
	sharedClients *SharedClientManager
	initOnce      sync.Once
)

// initSharedClients initialises the shared client cache once
func initSharedClients() {
	initOnce.Do(func() {
		sharedClients = &SharedClientManager{
			clients: make(map[string]*pmarginpro.Client),
		}
		sharedClients.cleanupFn = func() {
			sharedClients.mutex.Lock()
			defer sharedClients.mutex.Unlock()
			for key := range sharedClients.clients {
				delete(sharedClients.clients, key)
			}
		}
	})
}

// getOrCreateSharedClient returns the cached client or creates one for the config
func getOrCreateSharedClient(t *testing.T, config TestConfig) *pmarginpro.Client {
	initSharedClients()

	sharedClients.mutex.RLock()
	client, ok := sharedClients.clients[config.Name]
	sharedClients.mutex.RUnlock()
	if ok && client != nil {
		return client
	}

	sharedClients.mutex.Lock()
	defer sharedClients.mutex.Unlock()

	if client, ok := sharedClients.clients[config.Name]; ok && client != nil {
		return client
	}

	newClient, err := setupClient(config)
	if err != nil {
		t.Logf("failed to configure shared client for %s: %v", config.Name, err)
		return nil
	}
	sharedClients.clients[config.Name] = newClient
	return newClient
}

// setupClient constructs a fresh pmargin pro streams client for tests
func setupClient(config TestConfig) (*pmarginpro.Client, error) {
	client := pmarginpro.NewClient()
	if err := ensureDefaultServer(client); err != nil {
		return nil, err
	}
	return client, nil
}

// ensureDefaultServer verifies that SDK default servers are available
func ensureDefaultServer(client *pmarginpro.Client) error {
	if client.GetActiveServer() == nil {
		return fmt.Errorf("SDK did not configure a default active server")
	}
	return nil
}

// disconnectAllSharedClients clears cached client references
func disconnectAllSharedClients() {
	if sharedClients == nil {
		return
	}
	if sharedClients.cleanupFn != nil {
		sharedClients.cleanupFn()
	}
}

// getTestConfigs returns supported test scenarios
func getTestConfigs() []TestConfig {
	return []TestConfig{
		{
			Name:        "UserDataPro-Mainnet",
			Description: "Portfolio Margin Pro user data streams on Binance mainnet",
		},
	}
}

// getTestConfig returns the default configuration for tests
func getTestConfig() TestConfig {
	configs := getTestConfigs()
	if len(configs) == 0 {
		panic("no test configs defined")
	}
	return configs[0]
}

// StreamTestClient wraps a pmargin pro streams client with metadata for tests
type StreamTestClient struct {
	client *pmarginpro.Client
	config TestConfig
}

// Client exposes the underlying pmargin pro streams client
func (stc *StreamTestClient) Client() *pmarginpro.Client {
	return stc.client
}

// NewStreamTestClient returns a shared client instance for the given config
func NewStreamTestClient(t *testing.T, config TestConfig) (*StreamTestClient, error) {
	client := getOrCreateSharedClient(t, config)
	if client == nil {
		return nil, fmt.Errorf("shared client unavailable for config %s", config.Name)
	}
	return &StreamTestClient{
		client: client,
		config: config,
	}, nil
}

// NewStreamTestClientDedicated creates a dedicated client (no sharing)
func NewStreamTestClientDedicated(config TestConfig) (*StreamTestClient, error) {
	client, err := setupClient(config)
	if err != nil {
		return nil, err
	}
	return &StreamTestClient{
		client: client,
		config: config,
	}, nil
}
