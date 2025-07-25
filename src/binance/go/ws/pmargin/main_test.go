package pmargin_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/pmargin"
	"github.com/stretchr/testify/suite"
)

var (
	testAPIKey      string
	testSecretKey   string
	testListenKey   string
	testServerURL   string
)

func init() {
	// Load environment variables for testing
	testAPIKey = os.Getenv("BINANCE_API_KEY")
	testSecretKey = os.Getenv("BINANCE_SECRET_KEY")
	testListenKey = os.Getenv("BINANCE_LISTEN_KEY")
	testServerURL = os.Getenv("BINANCE_WS_SERVER_URL")

	// Set default server URL if not provided
	if testServerURL == "" {
		testServerURL = "wss://fstream.binance.com/pm/ws/{listenKey}"
	}

	// Configure logging for tests
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// BaseTestSuite provides common functionality for all test suites
type BaseTestSuite struct {
	suite.Suite
	client *pmargin.Client
	auth   *pmargin.Auth
	ctx    context.Context
}

// SetupSuite runs before each test suite
func (s *BaseTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Create auth if credentials available
	if testAPIKey != "" && testSecretKey != "" {
		s.auth = pmargin.NewAuth(testAPIKey)
		s.auth.SetSecretKey(testSecretKey)
		s.client = pmargin.NewClientWithAuth(s.auth)
	} else {
		s.client = pmargin.NewClient()
	}

	log.Printf("Portfolio Margin test suite setup completed")
}

// TearDownSuite runs after each test suite
func (s *BaseTestSuite) TearDownSuite() {
	if s.client != nil && s.client.IsConnected() {
		err := s.client.Disconnect()
		if err != nil {
			log.Printf("Error disconnecting client: %v", err)
		}
	}
	log.Printf("Portfolio Margin test suite teardown completed")
}

// TearDownTest runs after each individual test
func (s *BaseTestSuite) TearDownTest() {
	if s.client != nil && s.client.IsConnected() {
		err := s.client.Disconnect()
		if err != nil {
			log.Printf("Error disconnecting client after test: %v", err)
		}
		// Wait a bit for clean disconnection
		time.Sleep(100 * time.Millisecond)
	}
}

// SetupTest runs before each individual test
func (s *BaseTestSuite) SetupTest() {
	// Ensure we start each test with a clean client state
	if s.client != nil && s.client.IsConnected() {
		err := s.client.Disconnect()
		if err != nil {
			log.Printf("Warning: Error disconnecting client before test: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Helper functions for common test operations

// requireCredentials skips test if API credentials are not available
func (s *BaseTestSuite) requireCredentials() {
	if testAPIKey == "" || testSecretKey == "" {
		s.T().Skip("Skipping test: API credentials not available")
	}
}

// requireListenKey skips test if listen key is not available
func (s *BaseTestSuite) requireListenKey() {
	if testListenKey == "" {
		s.T().Skip("Skipping test: Listen key not available")
	}
}

// createTestClient creates a fresh client for testing
func (s *BaseTestSuite) createTestClient() *pmargin.Client {
	if testAPIKey != "" && testSecretKey != "" {
		auth := pmargin.NewAuth(testAPIKey)
		auth.SetSecretKey(testSecretKey)
		return pmargin.NewClientWithAuth(auth)
	}
	return pmargin.NewClient()
}

// waitForConnection waits for connection to be established
func (s *BaseTestSuite) waitForConnection(client *pmargin.Client, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if client.IsConnected() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// waitForDisconnection waits for connection to be closed
func (s *BaseTestSuite) waitForDisconnection(client *pmargin.Client, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !client.IsConnected() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// safeDisconnect safely disconnects a client with timeout
func (s *BaseTestSuite) safeDisconnect(client *pmargin.Client) {
	if client != nil && client.IsConnected() {
		err := client.Disconnect()
		if err != nil {
			log.Printf("Error disconnecting client: %v", err)
		}
		
		// Wait for disconnection with timeout
		if !s.waitForDisconnection(client, 2*time.Second) {
			log.Printf("Warning: Client did not disconnect within timeout")
		}
	}
}

// Test environment validation
func TestEnvironmentSetup(t *testing.T) {
	log.Println("🔍 === Validating Test Environment ===")
	
	// Check for API credentials
	if testAPIKey == "" {
		log.Println("⚠️  BINANCE_API_KEY not set - some tests will be skipped")
	} else {
		log.Println("✅ BINANCE_API_KEY available")
	}
	
	if testSecretKey == "" {
		log.Println("⚠️  BINANCE_SECRET_KEY not set - some tests will be skipped")
	} else {
		log.Println("✅ BINANCE_SECRET_KEY available")
	}
	
	if testListenKey == "" {
		log.Println("⚠️  BINANCE_LISTEN_KEY not set - user data stream tests will be skipped")
		log.Println("💡 To test user data streams, obtain a listen key from the REST API")
	} else {
		log.Println("✅ BINANCE_LISTEN_KEY available")
	}
	
	log.Printf("🌐 Using server URL: %s", testServerURL)
	log.Println("✅ Environment validation completed")
}