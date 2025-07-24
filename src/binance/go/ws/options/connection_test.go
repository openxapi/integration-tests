package options_test

import (
	"fmt"
	"testing"

	"github.com/openxapi/binance-go/ws/options"
	"github.com/stretchr/testify/suite"
)

// ConnectionTestSuite tests WebSocket connection functionality
type ConnectionTestSuite struct {
	BaseTestSuite
}

// TestConnectionSuite runs the connection test suite
func TestConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}

// TestBasicConnection tests basic connection scenarios
func (s *ConnectionTestSuite) TestBasicConnection() {
	s.Run("ConnectWithoutListenKey", func() {
		// Should fail since URL contains {listenKey} template
		err := s.client.Connect(s.ctx)
		s.Require().Error(err, "Connection should fail without listenKey")
		s.Require().Contains(err.Error(), "{listenKey}", "Error should mention listenKey template")
		s.logVerbose("Expected connection failure: %v", err)
	})
}

// TestConnectWithListenKey tests the new ConnectWithListenKey method
func (s *ConnectionTestSuite) TestConnectWithListenKey() {
	s.Run("ConnectWithListenKey", func() {
		testClient := options.NewClient()
		mockListenKey := "test_listen_key_12345"
		
		s.logVerbose("Testing ConnectWithListenKey with mock listenKey: %s", mockListenKey)
		err := testClient.ConnectWithListenKey(s.ctx, mockListenKey)
		
		if err != nil {
			s.logVerbose("ConnectWithListenKey failed (expected): %v", err)
			s.Require().NotContains(err.Error(), "{listenKey}", "Error should not mention template")
		} else {
			s.logVerbose("ConnectWithListenKey succeeded - URL template resolved correctly")
			s.Require().True(testClient.IsConnected(), "Should be connected if no error")
		}
		
		// Clean up
		if testClient.IsConnected() {
			testClient.Disconnect()
		}
		
		s.logVerbose("ConnectWithListenKey method test completed")
	})
}

// TestServerManagement tests essential server management
func (s *ConnectionTestSuite) TestServerManagement() {
	s.Run("ServerManagement", func() {
		// Verify default server
		activeServer := s.client.GetActiveServer()
		s.Require().NotNil(activeServer)
		s.Require().Contains(activeServer.URL, "{listenKey}")
		s.Require().Equal("wss://nbstream.binance.com/eoptions/ws/{listenKey}", activeServer.URL)
		
		// Test adding and removing server
		err := s.client.AddServer("test", "wss://test.com/ws/{listenKey}", "Test", "Test server")
		s.Require().NoError(err)
		
		// Switch back to mainnet and remove test server
		err = s.client.SetActiveServer("mainnet1")
		s.Require().NoError(err)
		
		err = s.client.RemoveServer("test")
		s.Require().NoError(err)
		
		s.logVerbose("Server management test completed")
	})
}

// TestAuthentication tests authentication setup
func (s *ConnectionTestSuite) TestAuthentication() {
	s.Run("AuthenticationSetup", func() {
		// Test HMAC auth setup
		auth := options.NewAuth("test_api_key")
		auth.SetSecretKey("test_secret_key")
		
		authCtx, err := auth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		s.Require().NotNil(authCtx)
		
		// Test empty API key validation
		emptyAuth := options.NewAuth("")
		_, err = emptyAuth.ContextWithValue(s.ctx)
		s.Require().Error(err, "Empty API key should cause error")
		
		s.logVerbose("Authentication setup test completed")
	})
}

// TestErrorHandling tests error scenarios
func (s *ConnectionTestSuite) TestErrorHandling() {
	s.Run("ErrorHandling", func() {
		// Test API error detection
		normalError := fmt.Errorf("normal error")
		_, isAPIError := options.IsAPIError(normalError)
		s.Require().False(isAPIError, "Normal error should not be detected as APIError")
		
		// Test invalid operation - try to remove active server
		testClient := options.NewClient()
		activeServer := testClient.GetActiveServer()
		s.Require().NotNil(activeServer)
		
		err := testClient.RemoveServer(activeServer.Name)
		s.Require().Error(err, "Should not be able to remove active server")
		
		s.logVerbose("Error handling test completed")
	})
}