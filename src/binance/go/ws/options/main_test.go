package options_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/options"
	"github.com/stretchr/testify/suite"
)

// Global variables for test configuration
var (
	testAPIKey    string
	testSecretKey string
	testClient    *options.Client
	testAuth      *options.Auth
	testCtx       context.Context
	testVerbose   bool
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load environment variables
	testAPIKey = os.Getenv("BINANCE_API_KEY")
	testSecretKey = os.Getenv("BINANCE_SECRET_KEY")
	testVerbose = os.Getenv("TEST_VERBOSE") == "true"

	// Validate required environment variables
	if testAPIKey == "" || testSecretKey == "" {
		log.Println("WARNING: BINANCE_API_KEY and BINANCE_SECRET_KEY not set")
		log.Println("Only basic connection tests will run")
		log.Println("To run all tests, set these environment variables with your credentials")
	}

	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

// BaseTestSuite provides common functionality for all test suites
type BaseTestSuite struct {
	suite.Suite
	client *options.Client
	auth   *options.Auth
	ctx    context.Context
}

// SetupSuite runs before the test suite
func (s *BaseTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Create auth if credentials are available
	if testAPIKey != "" && testSecretKey != "" {
		s.auth = options.NewAuth(testAPIKey)
		s.auth.SetSecretKey(testSecretKey)
		s.client = options.NewClientWithAuth(s.auth)
		log.Println("Created authenticated client for Options WebSocket tests")
	} else {
		s.client = options.NewClient()
		log.Println("Created unauthenticated client (limited functionality)")
	}

	// Log server configuration
	activeServer := s.client.GetActiveServer()
	if activeServer != nil {
		log.Printf("Using server: %s (%s)", activeServer.Name, activeServer.URL)
	}

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

	if apiErr, ok := options.IsAPIError(err); ok {
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

// waitForConnection waits for WebSocket connection or timeout
func (s *BaseTestSuite) waitForConnection(timeout time.Duration, operation string) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if s.client.IsConnected() {
			s.logVerbose("%s completed successfully", operation)
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.T().Errorf("%s timed out after %v", operation, timeout)
	return false
}

// waitForDisconnection waits for WebSocket disconnection or timeout
func (s *BaseTestSuite) waitForDisconnection(timeout time.Duration, operation string) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if !s.client.IsConnected() {
			s.logVerbose("%s completed successfully", operation)
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.T().Errorf("%s timed out after %v", operation, timeout)
	return false
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

// Constants for test operations
const (
	defaultTimeout = 10 * time.Second
	rateLimitDelay = 100 * time.Millisecond
)