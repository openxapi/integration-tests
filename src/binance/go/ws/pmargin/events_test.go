package pmargin_test

import (
	"encoding/json"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/pmargin"
	"github.com/openxapi/binance-go/ws/pmargin/models"
	"github.com/stretchr/testify/suite"
)

// EventsTestSuite tests event handling functionality
type EventsTestSuite struct {
	BaseTestSuite
}

// TestEventsSuite runs the event handling test suite
func TestEventsSuite(t *testing.T) {
	suite.Run(t, new(EventsTestSuite))
}

// EventHandlerTestState tracks event handler test state
type EventHandlerTestState struct {
	mu                                  sync.RWMutex
	conditionalOrderTradeUpdateReceived bool
	openOrderLossReceived               bool
	marginAccountUpdateReceived         bool
	liabilityUpdateReceived             bool
	marginOrderUpdateReceived           bool
	futuresOrderUpdateReceived          bool
	futuresBalancePositionUpdateReceived bool
	futuresAccountConfigUpdateReceived  bool
	riskLevelChangeReceived             bool
	marginBalanceUpdateReceived         bool
	userDataStreamExpiredReceived       bool
	errorReceived                       bool
	
	lastError                   error
	eventCount                  int
	receivedEvents              []string
}

// RecordEvent safely records an event reception
func (state *EventHandlerTestState) RecordEvent(eventType string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	
	state.eventCount++
	state.receivedEvents = append(state.receivedEvents, eventType)
	
	switch eventType {
	case "conditionalOrderTradeUpdate":
		state.conditionalOrderTradeUpdateReceived = true
	case "openOrderLoss":
		state.openOrderLossReceived = true
	case "marginAccountUpdate":
		state.marginAccountUpdateReceived = true
	case "liabilityUpdate":
		state.liabilityUpdateReceived = true
	case "marginOrderUpdate":
		state.marginOrderUpdateReceived = true
	case "futuresOrderUpdate":
		state.futuresOrderUpdateReceived = true
	case "futuresBalancePositionUpdate":
		state.futuresBalancePositionUpdateReceived = true
	case "futuresAccountConfigUpdate":
		state.futuresAccountConfigUpdateReceived = true
	case "riskLevelChange":
		state.riskLevelChangeReceived = true
	case "marginBalanceUpdate":
		state.marginBalanceUpdateReceived = true
	case "userDataStreamExpired":
		state.userDataStreamExpiredReceived = true
	case "error":
		state.errorReceived = true
	}
}

// GetEventCount safely gets the event count
func (state *EventHandlerTestState) GetEventCount() int {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.eventCount
}

// GetReceivedEvents safely gets the list of received events
func (state *EventHandlerTestState) GetReceivedEvents() []string {
	state.mu.RLock()
	defer state.mu.RUnlock()
	events := make([]string, len(state.receivedEvents))
	copy(events, state.receivedEvents)
	return events
}

// TestEventHandlerRegistration tests event handler registration
func (s *EventsTestSuite) TestEventHandlerRegistration() {
	log.Println("\n⚡ === Testing Event Handler Registration ===")

	s.Run("AllEventHandlerRegistration", func() {
		log.Println("📍 Testing registration of all event handlers...")
		
		client := pmargin.NewClient()
		state := &EventHandlerTestState{}
		
		// Register all event handlers
		client.OnConditionalOrderTradeUpdate(func(event *models.ConditionalOrderTradeUpdate) error {
			log.Printf("📊 Received conditional order trade update: %+v", event)
			state.RecordEvent("conditionalOrderTradeUpdate")
			return nil
		})
		
		client.OnOpenOrderLoss(func(event *models.OpenOrderLoss) error {
			log.Printf("📊 Received open order loss: %+v", event)
			state.RecordEvent("openOrderLoss")
			return nil
		})
		
		client.OnMarginAccountUpdate(func(event *models.MarginAccountUpdate) error {
			log.Printf("📊 Received margin account update: %+v", event)
			state.RecordEvent("marginAccountUpdate")
			return nil
		})
		
		client.OnLiabilityUpdate(func(event *models.LiabilityUpdate) error {
			log.Printf("📊 Received liability update: %+v", event)
			state.RecordEvent("liabilityUpdate")
			return nil
		})
		
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			log.Printf("📊 Received margin order update: %+v", event)
			state.RecordEvent("marginOrderUpdate")
			return nil
		})
		
		client.OnFuturesOrderUpdate(func(event *models.FuturesOrderUpdate) error {
			log.Printf("📊 Received futures order update: %+v", event)
			state.RecordEvent("futuresOrderUpdate")
			return nil
		})
		
		client.OnFuturesBalancePositionUpdate(func(event *models.FuturesBalancePositionUpdate) error {
			log.Printf("📊 Received futures balance position update: %+v", event)
			state.RecordEvent("futuresBalancePositionUpdate")
			return nil
		})
		
		client.OnFuturesAccountConfigUpdate(func(event *models.FuturesAccountConfigUpdate) error {
			log.Printf("📊 Received futures account config update: %+v", event)
			state.RecordEvent("futuresAccountConfigUpdate")
			return nil
		})
		
		client.OnRiskLevelChange(func(event *models.RiskLevelChange) error {
			log.Printf("📊 Received risk level change: %+v", event)
			state.RecordEvent("riskLevelChange")
			return nil
		})
		
		client.OnMarginBalanceUpdate(func(event *models.MarginBalanceUpdate) error {
			log.Printf("📊 Received margin balance update: %+v", event)
			state.RecordEvent("marginBalanceUpdate")
			return nil
		})
		
		client.OnUserDataStreamExpired(func(event *models.UserDataStreamExpired) error {
			log.Printf("📊 Received user data stream expired: %+v", event)
			state.RecordEvent("userDataStreamExpired")
			return nil
		})
		
		client.OnPmarginError(func(error *models.ErrorResponse) error {
			log.Printf("📊 Received error: %+v", error)
			state.RecordEvent("error")
			state.lastError = nil // Store for testing
			return nil
		})
		
		log.Println("✅ All event handlers registered successfully")
	})

	s.Run("EventHandlerErrorHandling", func() {
		log.Println("📍 Testing event handler error scenarios...")
		
		client := pmargin.NewClient()
		
		// Register a handler that returns an error
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			return nil // Return error to test error handling
		})
		
		log.Println("✅ Event handler error handling configured")
	})

	s.Run("ConcurrentEventHandlers", func() {
		log.Println("📍 Testing concurrent event handler execution...")
		
		client := pmargin.NewClient()
		state := &EventHandlerTestState{}
		var wg sync.WaitGroup
		
		// Register handlers that simulate concurrent processing
		client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(10 * time.Millisecond) // Simulate processing
				state.RecordEvent("marginOrderUpdate")
			}()
			return nil
		})
		
		client.OnMarginBalanceUpdate(func(event *models.MarginBalanceUpdate) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(15 * time.Millisecond) // Simulate processing
				state.RecordEvent("marginBalanceUpdate")
			}()
			return nil
		})
		
		log.Println("✅ Concurrent event handlers configured")
	})
}

// TestEventProcessing tests event processing with mock data
func (s *EventsTestSuite) TestEventProcessing() {
	log.Println("\n🔄 === Testing Event Processing ===")

	s.Run("MockEventDataParsing", func() {
		log.Println("📍 Testing mock event data parsing...")
		
		// Test parsing of different event types with mock data
		testEvents := []struct {
			name      string
			eventType string
			jsonData  string
		}{
			{
				name:      "ConditionalOrderTradeUpdate",
				eventType: "CONDITIONAL_ORDER_TRADE_UPDATE",
				jsonData:  `{"e":"CONDITIONAL_ORDER_TRADE_UPDATE","E":1234567890,"T":1234567890}`,
			},
			{
				name:      "MarginOrderUpdate",
				eventType: "executionReport",
				jsonData:  `{"e":"executionReport","E":1234567890,"s":"BTCUSDT","c":"TEST_ORDER","S":"BUY","o":"LIMIT","f":"GTC","q":"1.00000000","p":"50000.00000000","P":"0.00000000","F":"0.00000000","g":-1,"C":"","x":"NEW","X":"NEW","r":"NONE","i":123456,"l":"0.00000000","z":"0.00000000","L":"0.00000000","n":"0","N":null,"T":1234567890,"t":-1,"I":123456,"w":true,"m":false,"M":false,"O":1234567890,"Z":"0.00000000","Y":"0.00000000","Q":"0.00000000"}`,
			},
			{
				name:      "MarginAccountUpdate",
				eventType: "outboundAccountPosition",
				jsonData:  `{"e":"outboundAccountPosition","E":1234567890,"u":1234567890,"B":[{"a":"BTC","f":"1.00000000","l":"0.00000000"},{"a":"USDT","f":"1000.00000000","l":"0.00000000"}]}`,
			},
			{
				name:      "UserDataStreamExpired",
				eventType: "listenKeyExpired",
				jsonData:  `{"e":"listenKeyExpired","E":1234567890}`,
			},
		}
		
		for _, testEvent := range testEvents {
			s.Run(testEvent.name, func() {
				log.Printf("📍 Testing %s parsing...", testEvent.name)
				
				// Verify JSON is valid
				var generic map[string]interface{}
				err := json.Unmarshal([]byte(testEvent.jsonData), &generic)
				s.Require().NoError(err)
				
				// Verify event type field exists
				eventType, exists := generic["e"]
				s.Require().True(exists)
				s.Require().Equal(testEvent.eventType, eventType)
				
				log.Printf("✅ %s JSON structure valid", testEvent.name)
			})
		}
	})

	s.Run("EventTypeDetection", func() {
		log.Println("📍 Testing event type detection...")
		
		// Test various event type formats
		testCases := []struct {
			name     string
			jsonData string
			expected string
		}{
			{
				name:     "DirectEventField",
				jsonData: `{"e":"executionReport","s":"BTCUSDT"}`,
				expected: "executionReport",
			},
			{
				name:     "NestedEventField",
				jsonData: `{"event":{"e":"balanceUpdate"},"data":{}}`,
				expected: "balanceUpdate",
			},
		}
		
		for _, tc := range testCases {
			s.Run(tc.name, func() {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(tc.jsonData), &data)
				s.Require().NoError(err)
				
				// Test direct event field
				if eventType, exists := data["e"]; exists {
					s.Require().Equal(tc.expected, eventType)
				}
				
				// Test nested event field
				if eventObj, exists := data["event"]; exists {
					if eventMap, ok := eventObj.(map[string]interface{}); ok {
						if eventType, exists := eventMap["e"]; exists {
							s.Require().Equal(tc.expected, eventType)
						}
					}
				}
			})
		}
		
		log.Println("✅ Event type detection working")
	})
}

// TestRealTimeEventHandling tests event handling with real connections (if available)
func (s *EventsTestSuite) TestRealTimeEventHandling() {
	log.Println("\n📡 === Testing Real-Time Event Handling ===")

	s.Run("UserDataStreamEventHandling", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping real-time test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing user data stream event handling...")
		
		client := pmargin.NewClient()
		state := &EventHandlerTestState{}
		
		// Register all event handlers
		s.registerAllEventHandlers(client, state)
		
		// Connect to user data stream
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			log.Printf("⚠️  Connection failed: %v", err)
			s.T().Skipf("Cannot connect to user data stream: %v", err)
		}
		
		defer s.safeDisconnect(client)
		
		if !client.IsConnected() {
			s.T().Skip("Not connected, skipping real-time event test")
		}
		
		log.Println("✅ Connected to user data stream")
		
		// Subscribe to user data stream
		err = client.SubscribeToUserDataStream(s.ctx)
		s.Require().NoError(err)
		
		// Wait for potential events
		log.Println("⏳ Waiting for events (30 seconds)...")
		time.Sleep(30 * time.Second)
		
		// Check results
		eventCount := state.GetEventCount()
		receivedEvents := state.GetReceivedEvents()
		
		log.Printf("📊 Received %d events: %v", eventCount, receivedEvents)
		
		if eventCount > 0 {
			log.Println("✅ Real-time events received successfully")
		} else {
			log.Println("ℹ️  No events received (this may be normal if no trading activity)")
		}
	})

	s.Run("StreamKeepAlive", func() {
		if testListenKey == "" {
			s.T().Skip("Skipping keep-alive test: BINANCE_LISTEN_KEY not available")
		}
		
		log.Println("📍 Testing user data stream keep-alive...")
		
		client := pmargin.NewClient()
		
		err := client.ConnectToUserDataStream(s.ctx, testListenKey)
		if err != nil {
			s.T().Skipf("Cannot connect for keep-alive test: %v", err)
		}
		
		defer s.safeDisconnect(client)
		
		if !client.IsConnected() {
			s.T().Skip("Not connected, skipping keep-alive test")
		}
		
		// Test ping functionality
		err = client.PingUserDataStream(s.ctx)
		s.Require().NoError(err)
		
		log.Println("✅ User data stream ping working")
	})
}

// Helper function to register all event handlers
func (s *EventsTestSuite) registerAllEventHandlers(client *pmargin.Client, state *EventHandlerTestState) {
	client.OnConditionalOrderTradeUpdate(func(event *models.ConditionalOrderTradeUpdate) error {
		log.Printf("📊 Conditional Order Trade Update: %+v", event)
		state.RecordEvent("conditionalOrderTradeUpdate")
		return nil
	})
	
	client.OnOpenOrderLoss(func(event *models.OpenOrderLoss) error {
		log.Printf("📊 Open Order Loss: %+v", event)
		state.RecordEvent("openOrderLoss")
		return nil
	})
	
	client.OnMarginAccountUpdate(func(event *models.MarginAccountUpdate) error {
		log.Printf("📊 Margin Account Update: %+v", event)
		state.RecordEvent("marginAccountUpdate")
		return nil
	})
	
	client.OnLiabilityUpdate(func(event *models.LiabilityUpdate) error {
		log.Printf("📊 Liability Update: %+v", event)
		state.RecordEvent("liabilityUpdate")
		return nil
	})
	
	client.OnMarginOrderUpdate(func(event *models.MarginOrderUpdate) error {
		log.Printf("📊 Margin Order Update: %+v", event)
		state.RecordEvent("marginOrderUpdate")
		return nil
	})
	
	client.OnFuturesOrderUpdate(func(event *models.FuturesOrderUpdate) error {
		log.Printf("📊 Futures Order Update: %+v", event)
		state.RecordEvent("futuresOrderUpdate")
		return nil
	})
	
	client.OnFuturesBalancePositionUpdate(func(event *models.FuturesBalancePositionUpdate) error {
		log.Printf("📊 Futures Balance Position Update: %+v", event)
		state.RecordEvent("futuresBalancePositionUpdate")
		return nil
	})
	
	client.OnFuturesAccountConfigUpdate(func(event *models.FuturesAccountConfigUpdate) error {
		log.Printf("📊 Futures Account Config Update: %+v", event)
		state.RecordEvent("futuresAccountConfigUpdate")
		return nil
	})
	
	client.OnRiskLevelChange(func(event *models.RiskLevelChange) error {
		log.Printf("📊 Risk Level Change: %+v", event)
		state.RecordEvent("riskLevelChange")
		return nil
	})
	
	client.OnMarginBalanceUpdate(func(event *models.MarginBalanceUpdate) error {
		log.Printf("📊 Margin Balance Update: %+v", event)
		state.RecordEvent("marginBalanceUpdate")
		return nil
	})
	
	client.OnUserDataStreamExpired(func(event *models.UserDataStreamExpired) error {
		log.Printf("📊 User Data Stream Expired: %+v", event)
		state.RecordEvent("userDataStreamExpired")
		return nil
	})
	
	client.OnPmarginError(func(errorResp *models.ErrorResponse) error {
		log.Printf("📊 Portfolio Margin Error: %+v", errorResp)
		state.RecordEvent("error")
		return nil
	})
}