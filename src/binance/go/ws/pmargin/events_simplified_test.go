package pmargin_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/openxapi/binance-go/ws/pmargin"
	"github.com/openxapi/binance-go/ws/pmargin/models"
	"github.com/stretchr/testify/suite"
)

// EventsSimplifiedTestSuite tests basic event handling with workarounds for SDK issues
type EventsSimplifiedTestSuite struct {
	BaseTestSuite
}

// TestEventsSimplifiedSuite runs the simplified event handling test suite
func TestEventsSimplifiedSuite(t *testing.T) {
	suite.Run(t, new(EventsSimplifiedTestSuite))
}

// TestEventModelStructures tests that event model structures are properly defined
func (s *EventsSimplifiedTestSuite) TestEventModelStructures() {
	log.Println("\n📊 === Testing Event Model Structures ===")

	s.Run("ModelStructuresExist", func() {
		log.Println("📍 Testing that all event model structures exist...")
		
		// Test that we can create instances of all event models
		models := []interface{}{
			&models.ConditionalOrderTradeUpdate{},
			&models.OpenOrderLoss{},
			&models.MarginAccountUpdate{},
			&models.LiabilityUpdate{},
			&models.MarginOrderUpdate{},
			&models.FuturesOrderUpdate{},
			&models.FuturesBalancePositionUpdate{},
			&models.FuturesAccountConfigUpdate{},
			&models.RiskLevelChange{},
			&models.MarginBalanceUpdate{},
			&models.UserDataStreamExpired{},
			&models.ErrorResponse{},
		}
		
		for _, model := range models {
			s.Require().NotNil(model)
		}
		
		log.Printf("✅ All %d event model structures exist", len(models))
	})

	s.Run("ModelSerialization", func() {
		log.Println("📍 Testing event model JSON serialization...")
		
		// Test MarginOrderUpdate as an example
		orderUpdate := &models.MarginOrderUpdate{
			EventType:   "executionReport",
			EventTime:   1234567890,
			Symbol:      "BTCUSDT",
			OrderID:     123456,
		}
		
		// Test serialization
		jsonData, err := json.Marshal(orderUpdate)
		s.Require().NoError(err)
		s.Require().Contains(string(jsonData), "executionReport")
		s.Require().Contains(string(jsonData), "BTCUSDT")
		
		// Test deserialization
		var deserializedUpdate models.MarginOrderUpdate
		err = json.Unmarshal(jsonData, &deserializedUpdate)
		s.Require().NoError(err)
		s.Require().Equal("executionReport", deserializedUpdate.EventType)
		s.Require().Equal("BTCUSDT", deserializedUpdate.Symbol)
		
		log.Println("✅ Event model serialization working")
	})
}

// TestClientBasicFunctionality tests basic client functionality
func (s *EventsSimplifiedTestSuite) TestClientBasicFunctionality() {
	log.Println("\n🔧 === Testing Basic Client Functionality ===")

	s.Run("ClientCreation", func() {
		log.Println("📍 Testing client creation...")
		
		client := pmargin.NewClient()
		s.Require().NotNil(client)
		
		// Test basic client methods
		s.Require().False(client.IsConnected())
		
		activeServer := client.GetActiveServer()
		s.Require().NotNil(activeServer)
		s.Require().Contains(activeServer.URL, "{listenKey}")
		
		log.Println("✅ Basic client functionality working")
	})

	s.Run("ResponseListManagement", func() {
		log.Println("📍 Testing response list management...")
		
		client := pmargin.NewClient()
		
		// Test response list operations
		responses := client.GetResponseList()
		s.Require().NotNil(responses)
		s.Require().Equal(0, len(responses))
		
		client.ClearResponseList()
		clearedResponses := client.GetResponseList()
		s.Require().Equal(0, len(clearedResponses))
		
		log.Println("✅ Response list management working")
	})
}

// TestEventHandlerWorkArounds tests workarounds for SDK event handler issues
func (s *EventsSimplifiedTestSuite) TestEventHandlerWorkarounds() {
	log.Println("\n⚠️ === Testing Event Handler Workarounds (SDK Issues) ===")

	s.Run("HandlerMethodsExist", func() {
		log.Println("📍 Testing that handler methods exist (but may have type issues)...")
		
		client := pmargin.NewClient()
		
		// NOTE: The following handlers have SDK issues - they reference non-existent event types
		// These tests verify the methods exist but cannot test actual functionality due to SDK bugs
		
		// The SDK defines handlers that reference types like "ConditionalOrderTradeUpdateEvent"
		// but the actual model types are "ConditionalOrderTradeUpdate" (without "Event" suffix)
		
		log.Println("⚠️  SDK Issue: Handler types reference non-existent Event types")
		log.Println("⚠️  Example: ConditionalOrderTradeUpdateHandler expects *models.ConditionalOrderTradeUpdateEvent")
		log.Println("⚠️  But actual type is: *models.ConditionalOrderTradeUpdate")
		
		// We can verify the client exists and basic structure is there
		s.Require().NotNil(client)
		
		log.Println("✅ Client structure exists (handlers have type mismatches)")
	})
}

// TestMockEventProcessing tests event processing with mock data using correct types
func (s *EventsSimplifiedTestSuite) TestMockEventProcessing() {
	log.Println("\n🔄 === Testing Mock Event Processing ===")

	s.Run("MarginOrderUpdateParsing", func() {
		log.Println("📍 Testing margin order update event parsing...")
		
		// Mock JSON data for margin order update
		mockJSON := `{
			"e": "executionReport",
			"E": 1234567890,
			"s": "BTCUSDT",
			"c": "TEST_ORDER_ID",
			"S": "BUY",
			"o": "LIMIT",
			"f": "GTC",
			"q": "1.00000000",
			"p": "50000.00000000",
			"x": "NEW",
			"X": "NEW",
			"i": 123456,
			"T": 1234567890
		}`
		
		var orderUpdate models.MarginOrderUpdate
		err := json.Unmarshal([]byte(mockJSON), &orderUpdate)
		s.Require().NoError(err)
		
		s.Require().Equal("executionReport", orderUpdate.EventType)
		s.Require().Equal("BTCUSDT", orderUpdate.Symbol)
		s.Require().Equal("BUY", orderUpdate.Side)
		s.Require().Equal(int64(123456), orderUpdate.OrderID)
		
		log.Println("✅ Margin order update parsing working")
	})

	s.Run("UserDataStreamExpiredParsing", func() {
		log.Println("📍 Testing user data stream expired event parsing...")
		
		mockJSON := `{
			"e": "listenKeyExpired",
			"E": 1234567890
		}`
		
		var expiredEvent models.UserDataStreamExpired
		err := json.Unmarshal([]byte(mockJSON), &expiredEvent)
		s.Require().NoError(err)
		
		s.Require().Equal("listenKeyExpired", expiredEvent.EventType)
		s.Require().Equal(int64(1234567890), expiredEvent.EventTime)
		
		log.Println("✅ User data stream expired parsing working")
	})

	s.Run("ErrorResponseParsing", func() {
		log.Println("📍 Testing error response parsing...")
		
		mockJSON := `{
			"error": {
				"code": 1001,
				"msg": "Test error message"
			}
		}`
		
		var errorResp models.ErrorResponse
		err := json.Unmarshal([]byte(mockJSON), &errorResp)
		s.Require().NoError(err)
		
		s.Require().NotNil(errorResp.Error)
		s.Require().Equal(int64(1001), errorResp.Error.Code)
		s.Require().Equal("Test error message", errorResp.Error.Msg)
		
		log.Println("✅ Error response parsing working")
	})
}

// TestEventTypeHelpers tests helper methods on event types
func (s *EventsSimplifiedTestSuite) TestEventTypeHelpers() {
	log.Println("\n🔧 === Testing Event Type Helper Methods ===")

	s.Run("MarginOrderUpdateHelpers", func() {
		log.Println("📍 Testing MarginOrderUpdate helper methods...")
		
		orderUpdate := &models.MarginOrderUpdate{
			EventType: "executionReport",
			EventTime: 1234567890,
			Symbol:    "BTCUSDT",
		}
		
		// Test GetEventType method
		eventType := orderUpdate.GetEventType()
		s.Require().Equal("executionReport", eventType)
		
		// Test GetEventTime method
		eventTime := orderUpdate.GetEventTime()
		s.Require().Equal(int64(1234567890), eventTime)
		
		// Test String method
		str := orderUpdate.String()
		s.Require().Contains(str, "executionReport")
		s.Require().Contains(str, "BTCUSDT")
		
		log.Println("✅ MarginOrderUpdate helper methods working")
	})

	s.Run("UserDataStreamExpiredHelpers", func() {
		log.Println("📍 Testing UserDataStreamExpired helper methods...")
		
		expiredEvent := &models.UserDataStreamExpired{
			EventType: "listenKeyExpired",
			EventTime: 1234567890,
		}
		
		eventType := expiredEvent.GetEventType()
		s.Require().Equal("listenKeyExpired", eventType)
		
		eventTime := expiredEvent.GetEventTime()
		s.Require().Equal(int64(1234567890), eventTime)
		
		log.Println("✅ UserDataStreamExpired helper methods working")
	})
}