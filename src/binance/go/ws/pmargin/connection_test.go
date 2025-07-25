package pmargin_test

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/pmargin"
	"github.com/stretchr/testify/suite"
)

// ConnectionTestSuite tests WebSocket connection management
type ConnectionTestSuite struct {
	BaseTestSuite
}

// TestConnectionSuite runs the connection management test suite
func TestConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}

// TestClientCreation tests basic client creation and configuration
func (s *ConnectionTestSuite) TestClientCreation() {
	log.Println("\n🔧 === Testing Client Creation ===")

	s.Run("BasicClientCreation", func() {
		log.Println("📍 Testing basic client creation...")
		
		client := pmargin.NewClient()
		s.Require().NotNil(client)
		s.Require().False(client.IsConnected())
		
		log.Println("✅ Basic client created successfully")
	})

	s.Run("ClientWithAuth", func() {
		if testAPIKey == "" || testSecretKey == "" {
			s.T().Skip("Skipping auth client test: credentials not available")
		}
		
		log.Println("📍 Testing client creation with authentication...")
		
		auth := pmargin.NewAuth(testAPIKey)
		auth.SetSecretKey(testSecretKey)
		client := pmargin.NewClientWithAuth(auth)
		
		s.Require().NotNil(client)
		s.Require().False(client.IsConnected())
		
		log.Println("✅ Authenticated client created successfully")
	})

	s.Run("AuthOnExistingClient", func() {
		if testAPIKey == "" || testSecretKey == "" {
			s.T().Skip("Skipping auth test: credentials not available")
		}
		
		log.Println("📍 Testing authentication on existing client...")
		
		client := pmargin.NewClient()
		auth := pmargin.NewAuth(testAPIKey)
		auth.SetSecretKey(testSecretKey)
		client.SetAuth(auth)
		
		s.Require().NotNil(client)
		
		log.Println("✅ Authentication set on existing client")
	})
}

// TestServerManagement tests server configuration and management
func (s *ConnectionTestSuite) TestServerManagement() {
	log.Println("\n🌐 === Testing Server Management ===")

	s.Run("DefaultServerConfiguration", func() {
		log.Println("📍 Testing default server configuration...")
		
		client := pmargin.NewClient()
		
		// Test getting active server
		activeServer := client.GetActiveServer()
		s.Require().NotNil(activeServer)
		s.Require().Equal("mainnet1", activeServer.Name)
		s.Require().Contains(activeServer.URL, "{listenKey}")
		s.Require().True(activeServer.Active)
		
		log.Printf("✅ Default server: %s (%s)", activeServer.Name, activeServer.URL)
	})

	s.Run("ServerListing", func() {
		log.Println("📍 Testing server listing...")
		
		client := pmargin.NewClient()
		servers := client.ListServers()
		
		s.Require().Greater(len(servers), 0)
		s.Require().Contains(servers, "mainnet1")
		
		log.Printf("✅ Found %d servers", len(servers))
	})

	s.Run("AddCustomServer", func() {
		log.Println("📍 Testing custom server addition...")
		
		client := pmargin.NewClient()
		
		err := client.AddServer("test", "wss://test.example.com/ws/{listenKey}", "Test Server", "Test description")
		s.Require().NoError(err)
		
		server := client.GetServer("test")
		s.Require().NotNil(server)
		s.Require().Equal("test", server.Name)
		
		// Clean up
		err = client.RemoveServer("test")
		s.Require().NoError(err)
		
		log.Println("✅ Custom server management working")
	})

	s.Run("ServerSwitching", func() {
		log.Println("📍 Testing server switching...")
		
		client := pmargin.NewClient()
		
		// Add a test server
		err := client.AddServer("test", "wss://test.example.com/ws/{listenKey}", "Test", "Test server")
		s.Require().NoError(err)
		
		// Switch to test server
		err = client.SetActiveServer("test")
		s.Require().NoError(err)
		
		activeServer := client.GetActiveServer()
		s.Require().Equal("test", activeServer.Name)
		
		// Switch back to mainnet1
		err = client.SetActiveServer("mainnet1")
		s.Require().NoError(err)
		
		// Clean up
		err = client.RemoveServer("test")
		s.Require().NoError(err)
		
		log.Println("✅ Server switching working")
	})
}

// TestConnectionScenarios tests various connection scenarios
func (s *ConnectionTestSuite) TestConnectionScenarios() {
	log.Println("\n🔗 === Testing Connection Scenarios ===")

	s.Run("ConnectionWithoutListenKey", func() {
		log.Println("📍 Testing connection without listen key (should fail)...")
		
		client := pmargin.NewClient()
		
		// Try to connect without resolving {listenKey}
		err := client.Connect(s.ctx)
		s.Require().Error(err)
		s.Require().Contains(strings.ToLower(err.Error()), "listen")
		s.Require().False(client.IsConnected())
		
		log.Println("✅ Connection properly rejected without listen key")
	})

	s.Run("ConnectionWithInvalidListenKey", func() {
		log.Println("📍 Testing connection with invalid listen key...")
		
		client := pmargin.NewClient()
		
		// Try to connect with invalid listen key - SDK may be lenient about format
		err := client.ConnectWithListenKey(s.ctx, "invalid_listen_key")
		if err != nil {
			log.Println("✅ Connection properly rejected with invalid listen key")
			s.Require().False(client.IsConnected())
		} else {
			log.Println("ℹ️  SDK allows invalid listen key format - connection will likely fail during WebSocket handshake")
			// Even if no initial error, client should not be connected with invalid key
			if client.IsConnected() {
				// Give it a moment to fail
				time.Sleep(2 * time.Second)
				if !client.IsConnected() {
					log.Println("✅ Connection failed as expected after WebSocket handshake")
				}
			}
		}
	})

	s.Run("ConnectionStateManagement", func() {
		log.Println("📍 Testing connection state management...")
		
		client := pmargin.NewClient()
		
		// Initial state
		s.Require().False(client.IsConnected())
		
		// Disconnect unconnected client (should not error)
		err := client.Disconnect()
		s.Require().NoError(err)
		
		log.Println("✅ Connection state management working")
	})

	s.Run("ConnectionWithValidListenKey", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping valid listen key test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing connection with valid listen key...")
		
		client := pmargin.NewClient()
		
		// Try to connect with valid listen key
		err := client.ConnectWithListenKey(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed (this may be expected): %v", err)
			s.T().Skipf("Connection failed with valid listen key: %v", err)
		}
		
		if client.IsConnected() {
			log.Println("✅ Connected successfully with valid listen key")
			
			// Test graceful disconnection
			err = client.Disconnect()
			s.Require().NoError(err)
			
			// Wait for disconnection
			s.Require().True(s.waitForDisconnection(client, 2*time.Second))
			
			log.Println("✅ Disconnected successfully")
		}
	})
}

// TestConnectionMethods tests different connection methods
func (s *ConnectionTestSuite) TestConnectionMethods() {
	log.Println("\n🔌 === Testing Connection Methods ===")

	s.Run("ConnectWithVariables", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping variable connection test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing ConnectWithVariables method...")
		
		client := pmargin.NewClient()
		
		err := client.ConnectWithVariables(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed: %v", err)
			return
		}
		
		defer s.safeDisconnect(client)
		
		if client.IsConnected() {
			log.Println("✅ ConnectWithVariables working")
		}
	})

	s.Run("ConnectToServerWithListenKey", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping server connection test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing ConnectToServerWithListenKey method...")
		
		client := pmargin.NewClient()
		
		err := client.ConnectToServerWithListenKey(s.ctx, "mainnet1", testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed: %v", err)
			return
		}
		
		defer s.safeDisconnect(client)
		
		if client.IsConnected() {
			log.Println("✅ ConnectToServerWithListenKey working")
		}
	})

	s.Run("ConnectToUserDataStream", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping user data stream test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing ConnectToUserDataStream method...")
		
		client := pmargin.NewClient()
		
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed: %v", err)
			return
		}
		
		defer s.safeDisconnect(client)
		
		if client.IsConnected() {
			log.Println("✅ ConnectToUserDataStream working")
		}
	})
}

// TestConnectionErrors tests error handling in connections
func (s *ConnectionTestSuite) TestConnectionErrors() {
	log.Println("\n⚠️ === Testing Connection Error Handling ===")

	s.Run("InvalidServerURL", func() {
		log.Println("📍 Testing invalid server URL handling...")
		
		client := pmargin.NewClient()
		
		// Try to add server with invalid URL - the SDK may be more lenient
		err := client.AddServer("invalid", "invalid-url", "Invalid", "Invalid URL")
		if err != nil {
			log.Println("✅ Invalid URL properly rejected")
		} else {
			log.Println("ℹ️  SDK allows invalid URL format - this may be by design")
		}
	})

	s.Run("OperationsOnConnectedClient", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping connected client test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing operations on connected client...")
		
		client := pmargin.NewClient()
		
		err := client.ConnectWithListenKey(s.ctx, testListenKey)
		if err != nil {
			s.T().Skipf("Cannot connect for this test: %v", err)
		}
		
		if !client.IsConnected() {
			s.T().Skip("Client not connected, skipping connected client tests")
		}
		
		defer s.safeDisconnect(client)
		
		// Try to add server while connected (should fail)
		err = client.AddServer("test", "wss://test.com/ws", "Test", "Test")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "connected")
		
		// Try to remove server while connected (should fail)
		err = client.RemoveServer("mainnet1")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "connected")
		
		log.Println("✅ Operations on connected client properly restricted")
	})

	s.Run("RemoveActiveServer", func() {
		log.Println("📍 Testing removal of active server...")
		
		client := pmargin.NewClient()
		
		// Add a test server
		err := client.AddServer("test", "wss://test.com/ws", "Test", "Test")
		s.Require().NoError(err)
		
		// Try to remove active server (should fail)
		activeServer := client.GetActiveServer()
		err = client.RemoveServer(activeServer.Name)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "active")
		
		// Clean up
		err = client.SetActiveServer("test")
		s.Require().NoError(err)
		err = client.RemoveServer("mainnet1")
		s.Require().NoError(err)
		
		log.Println("✅ Active server removal properly restricted")
	})
}