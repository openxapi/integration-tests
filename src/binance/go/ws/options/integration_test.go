package options_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/options"
	"github.com/stretchr/testify/suite"
)

// FullIntegrationTestSuite runs all integration tests together
type FullIntegrationTestSuite struct {
	suite.Suite
	client *options.Client
	auth   *options.Auth
	ctx    context.Context
}

// TestFullIntegrationSuite is the main entry point for all tests
func TestFullIntegrationSuite(t *testing.T) {
	log.Println("🚀 === Starting Binance Options WebSocket Integration Tests ===")
	log.Printf("🌐 Server: mainnet1 (nbstream.binance.com/eoptions/ws/{listenKey})")
	log.Println("⚠️  Note: Updated SDK provides generic WebSocket client with authentication")
	log.Println("📝 Tests focus on: Connection management, Authentication, Event handling")
	log.Println("================================================================")

	// Run individual test suites
	suites := []struct {
		name  string
		suite suite.TestingSuite
	}{
		{"Connection Management Tests", new(ConnectionTestSuite)},
		{"Authentication Tests", new(AuthenticationTestSuite)},
		{"Event Handling Tests", new(EventsTestSuite)},
	}

	allPassed := true
	for _, s := range suites {
		log.Printf("\n📋 --- Running %s ---", s.name)
		// Use t.Run to properly run the suite
		success := t.Run(s.name, func(t *testing.T) {
			suite.Run(t, s.suite)
		})
		if !success {
			allPassed = false
		}
		// Delay between test suites
		time.Sleep(1 * time.Second)
	}

	// Run the comprehensive integration test if all individual suites passed
	if allPassed {
		log.Println("\n🎯 --- Running Comprehensive Integration Test ---")
		suite.Run(t, new(FullIntegrationTestSuite))
	} else {
		t.Error("❌ Some test suites failed, skipping comprehensive integration test")
	}

	// Print test summary
	log.Println("\n📊 === Test Summary ===")
	log.Println("✅ All Options WebSocket integration tests completed!")
	log.Println("🔧 SDK Architecture: Generic WebSocket client with authentication")
	log.Println("🔗 Connection Pattern: listenKey-based URL templates")
	log.Println("🔐 Authentication: HMAC, RSA, Ed25519 support")
	log.Println("⚡ Event Handling: Generic event processing infrastructure")
	log.Println("======================")
}

// SetupSuite runs before the test suite
func (s *FullIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Create auth if available
	if testAPIKey != "" && testSecretKey != "" {
		s.auth = options.NewAuth(testAPIKey)
		s.auth.SetSecretKey(testSecretKey)
		s.client = options.NewClientWithAuth(s.auth)
	} else {
		s.client = options.NewClient()
	}

	log.Println("Full integration test suite setup completed")
}

// TearDownSuite runs after the test suite
func (s *FullIntegrationTestSuite) TearDownSuite() {
	if s.client != nil && s.client.IsConnected() {
		err := s.client.Disconnect()
		if err != nil {
			log.Printf("Error disconnecting client: %v", err)
		}
	}
	log.Println("Full integration test suite teardown completed")
}

// TestComprehensiveWorkflow tests the complete SDK workflow
func (s *FullIntegrationTestSuite) TestComprehensiveWorkflow() {
	log.Println("\n🔄 === Starting Comprehensive SDK Workflow Test ===")

	// Step 1: Test basic client creation and configuration
	s.Run("1_ClientCreationAndConfig", func() {
		log.Println("📍 Step 1: Testing client creation and configuration...")
		s.testClientCreation()
	})

	// Step 2: Test server management
	s.Run("2_ServerManagement", func() {
		log.Println("🌐 Step 2: Testing server management...")
		s.testServerManagement()
	})

	// Step 3: Test authentication setup
	s.Run("3_AuthenticationSetup", func() {
		log.Println("🔐 Step 3: Testing authentication setup...")
		s.testAuthenticationSetup()
	})

	// Step 4: Test connection scenarios
	s.Run("4_ConnectionScenarios", func() {
		log.Println("🔗 Step 4: Testing connection scenarios...")
		s.testConnectionScenarios()
	})

	// Step 5: Test event handling infrastructure
	s.Run("5_EventHandling", func() {
		log.Println("⚡ Step 5: Testing event handling infrastructure...")
		s.testEventHandling()
	})

	// Step 6: Test error handling
	s.Run("6_ErrorHandling", func() {
		log.Println("⚠️ Step 6: Testing error handling...")
		s.testErrorHandling()
	})

	// Step 7: Test performance and concurrency
	s.Run("7_PerformanceAndConcurrency", func() {
		log.Println("🚀 Step 7: Testing performance and concurrency...")
		s.testPerformanceAndConcurrency()
	})

	log.Println("\n✅ === Comprehensive SDK Workflow Test Completed ===")
}

// Helper methods for comprehensive test

func (s *FullIntegrationTestSuite) testClientCreation() {
	// Test basic client creation
	basicClient := options.NewClient()
	s.Require().NotNil(basicClient)
	
	// Test client with auth
	if testAPIKey != "" && testSecretKey != "" {
		auth := options.NewAuth(testAPIKey)
		auth.SetSecretKey(testSecretKey)
		authClient := options.NewClientWithAuth(auth)
		s.Require().NotNil(authClient)
	}
	
	log.Printf("✅ Client creation test completed")
}

func (s *FullIntegrationTestSuite) testServerManagement() {
	// Test server configuration
	activeServer := s.client.GetActiveServer()
	s.Require().NotNil(activeServer)
	s.Require().Contains(activeServer.URL, "{listenKey}")
	
	// Test server listing
	servers := s.client.ListServers()
	s.Require().Contains(servers, "mainnet1")
	
	// Test adding custom server
	err := s.client.AddServer("test", "wss://test.com/ws/{listenKey}", "Test", "Test server")
	s.Require().NoError(err)
	
	// Clean up
	err = s.client.SetActiveServer("mainnet1")
	s.Require().NoError(err)
	err = s.client.RemoveServer("test")
	s.Require().NoError(err)
	
	log.Printf("✅ Server management test completed")
}

func (s *FullIntegrationTestSuite) testAuthenticationSetup() {
	// Test HMAC auth
	hmacAuth := options.NewAuth("test_key")
	hmacAuth.SetSecretKey("test_secret")
	_, err := hmacAuth.ContextWithValue(s.ctx)
	s.Require().NoError(err)
	
	// Test auth on client
	s.client.SetAuth(hmacAuth)
	
	// Test auth validation
	emptyAuth := options.NewAuth("")
	_, err = emptyAuth.ContextWithValue(s.ctx)
	s.Require().Error(err)
	
	log.Printf("✅ Authentication setup test completed")
}

func (s *FullIntegrationTestSuite) testConnectionScenarios() {
	// Test connection without listenKey (should fail)
	err := s.client.Connect(s.ctx)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "{listenKey}")
	
	// Test connection state
	s.Require().False(s.client.IsConnected())
	
	// Test disconnect on unconnected client
	err = s.client.Disconnect()
	s.Require().NoError(err)
	
	log.Printf("✅ Connection scenarios test completed")
}

func (s *FullIntegrationTestSuite) testEventHandling() {
	// Test response list functionality
	responses := s.client.GetResponseList()
	s.Require().NotNil(responses)
	
	// Test clearing
	s.client.ClearResponseList()
	clearedResponses := s.client.GetResponseList()
	s.Require().Equal(0, len(clearedResponses))
	
	log.Printf("✅ Event handling test completed")
}

func (s *FullIntegrationTestSuite) testErrorHandling() {
	// Test API error detection
	normalErr := fmt.Errorf("normal error")
	_, isAPI := options.IsAPIError(normalErr)
	s.Require().False(isAPI)
	
	// Test invalid operations - try to remove active server (should fail)
	invalidClient := options.NewClient()
	activeServer := invalidClient.GetActiveServer()
	s.Require().NotNil(activeServer)
	
	err := invalidClient.RemoveServer(activeServer.Name)
	s.Require().Error(err)
	
	log.Printf("✅ Error handling test completed")
}

func (s *FullIntegrationTestSuite) testPerformanceAndConcurrency() {
	// Test concurrent response list access
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				responses := s.client.GetResponseList()
				_ = responses
			}
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	log.Printf("✅ Performance and concurrency test completed")
}

func (s *FullIntegrationTestSuite) waitForOperation(timeout time.Duration, operation string) {
	select {
	case <-time.After(timeout):
		s.T().Errorf("%s timed out after %v", operation, timeout)
	default:
		// Operation completed
	}
}