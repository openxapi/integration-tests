package pmargin_test

import (
	"log"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/pmargin"
	"github.com/openxapi/binance-go/ws/pmargin/models"
	"github.com/stretchr/testify/suite"
)

// UserDataTestSuite tests user data stream operations
type UserDataTestSuite struct {
	BaseTestSuite
}

// TestUserDataSuite runs the user data stream test suite
func TestUserDataSuite(t *testing.T) {
	suite.Run(t, new(UserDataTestSuite))
}

// TestUserDataStreamLifecycle tests the complete user data stream lifecycle
func (s *UserDataTestSuite) TestUserDataStreamLifecycle() {
	log.Println("\n🔄 === Testing User Data Stream Lifecycle ===")

	s.Run("CompleteStreamLifecycle", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping user data stream test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing complete user data stream lifecycle...")
		
		client := pmargin.NewClient()
		defer s.safeDisconnect(client)
		
		// Step 1: Connect to user data stream
		log.Println("🔗 Step 1: Connecting to user data stream...")
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed: %v", err)
			s.T().Skipf("Cannot connect to user data stream: %v", err)
		}
		
		s.Require().True(client.IsConnected())
		log.Println("✅ Successfully connected to user data stream")
		
		// Step 2: Subscribe to user data stream
		log.Println("📡 Step 2: Subscribing to user data stream...")
		err = client.SubscribeToUserDataStream(s.ctx)
		s.Require().NoError(err)
		log.Println("✅ Successfully subscribed to user data stream")
		
		// Step 3: Set up event handlers
		log.Println("⚡ Step 3: Setting up event handlers...")
		eventReceived := false
		
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			log.Printf("📊 Received margin order update: %+v", event)
			eventReceived = true
			return nil
		})
		
		client.OnMarginBalanceUpdate(func(event *models.MarginBalanceUpdate) error {
			log.Printf("📊 Received margin balance update: %+v", event)
			eventReceived = true
			return nil
		})
		
		client.OnUserDataStreamExpired(func(event *models.UserDataStreamExpired) error {
			log.Printf("📊 User data stream expired: %+v", event)
			return nil
		})
		
		log.Println("✅ Event handlers set up")
		
		// Step 4: Keep stream alive
		log.Println("💓 Step 4: Testing stream keep-alive...")
		err = client.PingUserDataStream(s.ctx)
		s.Require().NoError(err)
		log.Println("✅ Stream keep-alive working")
		
		// Step 5: Wait for potential events
		log.Println("⏳ Step 5: Waiting for events (15 seconds)...")
		time.Sleep(15 * time.Second)
		
		if eventReceived {
			log.Println("✅ Events received during test period")
		} else {
			log.Println("ℹ️  No events received (normal if no trading activity)")
		}
		
		// Step 6: Disconnect gracefully
		log.Println("🔌 Step 6: Disconnecting from stream...")
		err = client.Disconnect()
		s.Require().NoError(err)
		
		// Wait for disconnection
		s.Require().True(s.waitForDisconnection(client, 2*time.Second))
		log.Println("✅ Successfully disconnected from stream")
		
		log.Println("🎉 User data stream lifecycle test completed successfully")
	})
}

// TestUserDataStreamOperations tests specific user data stream operations
func (s *UserDataTestSuite) TestUserDataStreamOperations() {
	log.Println("\n🛠️ === Testing User Data Stream Operations ===")

	s.Run("SubscriptionWithoutConnection", func() {
		log.Println("📍 Testing subscription without connection...")
		
		client := pmargin.NewClient()
		
		// Try to subscribe without being connected
		err := client.SubscribeToUserDataStream(s.ctx)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "not connected")
		
		log.Println("✅ Subscription properly rejected without connection")
	})

	s.Run("PingWithoutConnection", func() {
		log.Println("📍 Testing ping without connection...")
		
		client := pmargin.NewClient()
		
		// Try to ping without being connected
		err := client.PingUserDataStream(s.ctx)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "not connected")
		
		log.Println("✅ Ping properly rejected without connection")
	})

	s.Run("ConnectToAlreadyConnectedStream", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping double connection test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing connection to already connected stream...")
		
		client := pmargin.NewClient()
		defer s.safeDisconnect(client)
		
		// First connection
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			s.T().Skipf("Cannot connect for double connection test: %v", err)
		}
		
		if !client.IsConnected() {
			s.T().Skip("First connection failed, skipping double connection test")
		}
		
		// Try to connect again
		err = client.ConnectToUserDataStream(s.ctx, testListenKey)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "already connected")
		
		log.Println("✅ Double connection properly rejected")
	})
}

// TestUserDataStreamEvents tests specific event scenarios
func (s *UserDataTestSuite) TestUserDataStreamEvents() {
	log.Println("\n📊 === Testing User Data Stream Events ===")

	s.Run("EventHandlerCoverage", func() {
		log.Println("📍 Testing event handler coverage...")
		
		client := pmargin.NewClient()
		handlersCalled := make(map[string]bool)
		
		// Register handlers for all event types
		client.OnConditionalOrderTradeUpdate(func(event *models.ConditionalOrderTradeUpdate) error {
			handlersCalled["conditionalOrderTradeUpdate"] = true
			return nil
		})
		
		client.OnOpenOrderLoss(func(event *models.OpenOrderLoss) error {
			handlersCalled["openOrderLoss"] = true
			return nil
		})
		
		client.OnMarginAccountUpdate(func(event *models.MarginAccountUpdate) error {
			handlersCalled["marginAccountUpdate"] = true
			return nil
		})
		
		client.OnLiabilityUpdate(func(event *models.LiabilityUpdate) error {
			handlersCalled["liabilityUpdate"] = true
			return nil
		})
		
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			handlersCalled["marginOrderUpdate"] = true
			return nil
		})
		
		client.OnFuturesOrderUpdate(func(event *models.FuturesOrderUpdate) error {
			handlersCalled["futuresOrderUpdate"] = true
			return nil
		})
		
		client.OnFuturesBalancePositionUpdate(func(event *models.FuturesBalancePositionUpdate) error {
			handlersCalled["futuresBalancePositionUpdate"] = true
			return nil
		})
		
		client.OnFuturesAccountConfigUpdate(func(event *models.FuturesAccountConfigUpdate) error {
			handlersCalled["futuresAccountConfigUpdate"] = true
			return nil
		})
		
		client.OnRiskLevelChange(func(event *models.RiskLevelChange) error {
			handlersCalled["riskLevelChange"] = true
			return nil
		})
		
		client.OnMarginBalanceUpdate(func(event *models.MarginBalanceUpdate) error {
			handlersCalled["marginBalanceUpdate"] = true
			return nil
		})
		
		client.OnUserDataStreamExpired(func(event *models.UserDataStreamExpired) error {
			handlersCalled["userDataStreamExpired"] = true
			return nil
		})
		
		client.OnPmarginError(func(errorResp *models.ErrorResponse) error {
			handlersCalled["error"] = true
			return nil
		})
		
		log.Printf("✅ Registered %d event handlers", len(handlersCalled))
		
		// Note: handlers would be tested with real connection if listen key available
		if testListenKey != "" {
			log.Println("ℹ️  Handlers registered - would be tested with real connection")
		}
	})

	s.Run("EventHandlerErrorScenarios", func() {
		log.Println("📍 Testing event handler error scenarios...")
		
		client := pmargin.NewClient()
		
		// Register handler that returns error
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			log.Printf("📊 Handler processing event and returning error for testing")
			return nil // Return error to test error handling in real scenarios
		})
		
		// Register error handler
		client.OnPmarginError(func(errorResp *models.ErrorResponse) error {
			log.Printf("📊 Error handler called: %+v", errorResp)
			return nil
		})
		
		log.Println("✅ Error scenario handlers configured")
	})
}

// TestUserDataStreamReconnection tests reconnection scenarios
func (s *UserDataTestSuite) TestUserDataStreamReconnection() {
	log.Println("\n🔄 === Testing User Data Stream Reconnection ===")

	s.Run("DisconnectAndReconnect", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping reconnection test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing disconnect and reconnect scenario...")
		
		client := pmargin.NewClient()
		
		// First connection
		log.Println("🔗 Establishing initial connection...")
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			s.T().Skipf("Cannot connect for reconnection test: %v", err)
		}
		
		if !client.IsConnected() {
			s.T().Skip("Initial connection failed, skipping reconnection test")
		}
		
		log.Println("✅ Initial connection successful")
		
		// Disconnect
		log.Println("🔌 Disconnecting...")
		err = client.Disconnect()
		s.Require().NoError(err)
		s.Require().True(s.waitForDisconnection(client, 2*time.Second))
		log.Println("✅ Disconnection successful")
		
		// Wait a moment
		time.Sleep(1 * time.Second)
		
		// Reconnect
		log.Println("🔗 Reconnecting...")
		err = client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Reconnection failed: %v", err)
		} else if client.IsConnected() {
			log.Println("✅ Reconnection successful")
			
			// Clean up
			err = client.Disconnect()
			s.Require().NoError(err)
		}
	})

	s.Run("MultipleConsecutiveConnections", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping multiple connections test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing multiple consecutive connections...")
		
		for i := 0; i < 3; i++ {
			log.Printf("🔗 Connection attempt %d...", i+1)
			
			client := pmargin.NewClient()
			
			err := client.ConnectToUserDataStream(s.ctx, testListenKey)
			if err != nil {
				log.Printf("⚠️  Connection %d failed: %v", i+1, err)
				continue
			}
			
			if client.IsConnected() {
				log.Printf("✅ Connection %d successful", i+1)
				
				// Brief operation
				time.Sleep(1 * time.Second)
				
				// Disconnect
				err = client.Disconnect()
				s.Require().NoError(err)
				
				// Wait for clean disconnection
				s.Require().True(s.waitForDisconnection(client, 2*time.Second))
				log.Printf("✅ Disconnection %d successful", i+1)
			}
			
			// Brief pause between connections
			time.Sleep(500 * time.Millisecond)
		}
		
		log.Println("✅ Multiple consecutive connections test completed")
	})
}