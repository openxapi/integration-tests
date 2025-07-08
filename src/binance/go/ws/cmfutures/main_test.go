package cmfutures_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/cmfutures"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Global variables for test configuration
var (
	testAPIKey    string
	testSecretKey string
	testSymbol    string
	testClient    *cmfutures.Client
	testAuth      *cmfutures.Auth
	testCtx       context.Context
	testVerbose   bool
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load environment variables
	testAPIKey = os.Getenv("BINANCE_API_KEY")
	testSecretKey = os.Getenv("BINANCE_SECRET_KEY")
	testSymbol = os.Getenv("TEST_SYMBOL")
	testVerbose = os.Getenv("TEST_VERBOSE") == "true"

	// Set defaults
	if testSymbol == "" {
		testSymbol = "BTCUSD_PERP" // Default CMFUTURES perpetual contract (Coin-M uses USD not USDT)
	}

	// Validate required environment variables
	if testAPIKey == "" || testSecretKey == "" {
		log.Println("WARNING: BINANCE_API_KEY and BINANCE_SECRET_KEY not set")
		log.Println("Only public endpoint tests will run (Note: CMFUTURES has no public WebSocket endpoints)")
		log.Println("To run all tests, set these environment variables with your testnet credentials")
	}

	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

// BaseTestSuite provides common functionality for all test suites
type BaseTestSuite struct {
	suite.Suite
	client *cmfutures.Client
	auth   *cmfutures.Auth
	ctx    context.Context
}

// SetupSuite runs before the test suite
func (s *BaseTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Create auth if credentials are available
	if testAPIKey != "" && testSecretKey != "" {
		s.auth = cmfutures.NewAuth(testAPIKey)
		s.auth.SetSecretKey(testSecretKey)
		s.client = cmfutures.NewClientWithAuth(s.auth)
		log.Println("Created authenticated client for CMFUTURES WebSocket tests")
	} else {
		s.client = cmfutures.NewClient()
		log.Println("Created unauthenticated client (limited functionality)")
	}

	// Set testnet server
	err := s.client.SetActiveServer("testnet1")
	require.NoError(s.T(), err, "Failed to set testnet server")

	// Connect to WebSocket
	err = s.client.Connect(s.ctx)
	require.NoError(s.T(), err, "Failed to connect to WebSocket")

	// Allow connection to stabilize
	time.Sleep(500 * time.Millisecond)

	log.Println("Test suite setup completed")
}

// TearDownSuite runs after the test suite
func (s *BaseTestSuite) TearDownSuite() {
	if s.client != nil {
		// Try to disconnect only if still connected
		if s.client.IsConnected() {
			err := s.client.Disconnect()
			if err != nil && err.Error() != "use of closed network connection" {
				log.Printf("Error disconnecting client: %v", err)
			}
		} else {
			log.Println("Client already disconnected")
		}
	}
	log.Println("Test suite teardown completed")
}

// Helper functions

// requireAuth skips the test if authentication is not available
func (s *BaseTestSuite) requireAuth() {
	if s.auth == nil {
		s.T().Skip("Skipping test: authentication required but not configured")
	}
}

// logVerbose logs a message if verbose mode is enabled
func (s *BaseTestSuite) logVerbose(format string, args ...interface{}) {
	if testVerbose {
		log.Printf("[VERBOSE] "+format, args...)
	}
}

// handleAPIError processes API errors with detailed logging
func (s *BaseTestSuite) handleAPIError(err error, operation string) {
	if err == nil {
		return
	}

	if apiErr, ok := cmfutures.IsAPIError(err); ok {
		s.T().Errorf("%s failed with API error: Status=%d, Code=%d, Message=%s, ID=%s",
			operation, apiErr.Status, apiErr.Code, apiErr.Message, apiErr.ID)
		
		// Log specific error guidance
		switch apiErr.Status {
		case 400:
			s.T().Log("Bad Request: Check request parameters")
		case 401:
			s.T().Log("Unauthorized: Check API key and signature")
		case 403:
			s.T().Log("Forbidden: Check API permissions or IP whitelist")
		case 429:
			s.T().Log("Rate Limit: Slow down request frequency")
		}
	} else {
		s.T().Errorf("%s failed with error: %v", operation, err)
	}
}

// waitForResponse waits for a response with timeout
func (s *BaseTestSuite) waitForResponse(done chan bool, timeout time.Duration, operation string) {
	select {
	case <-done:
		s.logVerbose("%s completed successfully", operation)
	case <-time.After(timeout):
		s.T().Errorf("%s timed out after %v", operation, timeout)
	}
}

// getTestContext returns a context with authentication if available
func (s *BaseTestSuite) getTestContext() context.Context {
	if s.auth != nil {
		authCtx, err := s.auth.ContextWithValue(s.ctx)
		if err != nil {
			s.T().Fatalf("Failed to create auth context: %v", err)
		}
		return authCtx
	}
	return s.ctx
}

// formatJSON formats a response for logging
func formatJSON(v interface{}) string {
	return fmt.Sprintf("%+v", v)
}

// Test helpers for common operations

// placeTestOrder places a small test order
func (s *BaseTestSuite) placeTestOrder() (int64, error) {
	// Implementation will be in trading_test.go
	return 0, fmt.Errorf("not implemented")
}

// cancelAllOrders cancels all open orders for the test symbol
func (s *BaseTestSuite) cancelAllOrders() {
	// Implementation will be in trading_test.go
}

// Constants for test operations
const (
	defaultTimeout    = 10 * time.Second
	rateLimitDelay    = 100 * time.Millisecond
	testOrderQuantity = "1"     // Minimum contract quantity for CMFUTURES
	testOrderPrice    = "20000" // Lower test price for limit orders to avoid rejection
)

// Helper functions removed - using fluent API with SetX methods instead