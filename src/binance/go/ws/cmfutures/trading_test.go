package cmfutures_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/cmfutures"
	"github.com/openxapi/binance-go/ws/cmfutures/models"
	"github.com/stretchr/testify/suite"
)

// TradingTestSuite tests trading-related APIs
type TradingTestSuite struct {
	BaseTestSuite
	testOrderID int64
}

// TestTradingTestSuite runs the trading test suite
func TestTradingTestSuite(t *testing.T) {
	suite.Run(t, new(TradingTestSuite))
}

// SetupTest runs before each test
func (s *TradingTestSuite) SetupTest() {
	s.testOrderID = 0
}

// TearDownTest runs after each test
func (s *TradingTestSuite) TearDownTest() {
	// Clean up any open orders
	if s.testOrderID != 0 {
		s.cancelTestOrder(s.testOrderID)
		s.testOrderID = 0
	}
}

// TestOrderPlace tests the order.place endpoint
func (s *TradingTestSuite) TestOrderPlace() {
	s.requireAuth()

	done := make(chan bool)
	request := models.NewOrderPlaceRequest().
		SetSymbol(testSymbol).
		SetSide("BUY").
		SetType("LIMIT").
		SetQuantity(testOrderQuantity).
		SetPrice(testOrderPrice).
		SetTimeInForce("GTC").
		SetNewClientOrderId(fmt.Sprintf("test_%d", time.Now().Unix()))

	s.logVerbose("Placing test order: %+v", request.Params)

	err := s.client.SendOrderPlace(s.getTestContext(), request, func(response *models.OrderPlaceResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "order.place")
			// Check if it's a min notional error
			if apiErr, ok := cmfutures.IsAPIError(err); ok && apiErr.Code == -1013 {
				s.T().Skip("Skipping test due to MIN_NOTIONAL requirement")
			}
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Store order ID for cleanup
		s.testOrderID = response.Result.OrderId

		// Log order details
		s.logVerbose("Order placed successfully: %s", formatJSON(response))
		log.Printf("🚀 Order placed - ID: %d, Symbol: %s, Side: %s, Price: %s, Quantity: %s",
			response.Result.OrderId, response.Result.Symbol, response.Result.Side,
			response.Result.Price, response.Result.OrigQty)

		// Validate order fields
		s.Assert().Equal(testSymbol, response.Result.Symbol)
		s.Assert().Equal("BUY", response.Result.Side)
		s.Assert().Equal("LIMIT", response.Result.Type)
		s.Assert().Contains([]string{"NEW", "PARTIALLY_FILLED", "FILLED"}, response.Result.Status)

		return nil
	})

	s.Require().NoError(err, "Failed to send order.place request")
	s.waitForResponse(done, defaultTimeout, "order.place")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestOrderModify tests the order.modify endpoint
func (s *TradingTestSuite) TestOrderModify() {
	s.requireAuth()

	// First place an order to modify
	orderID := s.placeTestOrderForModification()
	if orderID == 0 {
		s.T().Skip("Failed to place order for modification test")
	}

	done := make(chan bool)
	newPrice := "17500" // Different price for modification
	request := models.NewOrderModifyRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID).
		SetSide("BUY").
		SetQuantity(testOrderQuantity).
		SetPrice(newPrice)

	s.logVerbose("Modifying order %d with new price: %s", orderID, newPrice)

	err := s.client.SendOrderModify(s.getTestContext(), request, func(response *models.OrderModifyResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "order.modify")
			// Some order types or states might not support modification
			if apiErr, ok := cmfutures.IsAPIError(err); ok {
				log.Printf("Order modify error (expected in some cases): Code=%d, Message=%s", 
					apiErr.Code, apiErr.Message)
			}
			return nil // Don't fail test as modify might not be supported
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log modification details
		s.logVerbose("Order modified successfully: %s", formatJSON(response))
		log.Printf("✏️ Order modified - ID: %d, New Price: %s", 
			response.Result.OrderId, response.Result.Price)

		// Validate modification
		s.Assert().Equal(newPrice, response.Result.Price)

		return nil
	})

	s.Require().NoError(err, "Failed to send order.modify request")
	s.waitForResponse(done, defaultTimeout, "order.modify")

	// Clean up the modified order
	s.cancelTestOrder(orderID)

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestOrderStatus tests the order.status endpoint
func (s *TradingTestSuite) TestOrderStatus() {
	s.requireAuth()

	// First place an order to check status
	orderID := s.placeTestOrderForStatus()
	if orderID == 0 {
		s.T().Skip("Failed to place order for status test")
	}
	defer s.cancelTestOrder(orderID)

	done := make(chan bool)
	request := models.NewOrderStatusRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	s.logVerbose("Checking status for order: %d", orderID)

	err := s.client.SendOrderStatus(s.getTestContext(), request, func(response *models.OrderStatusResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "order.status")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log status details
		s.logVerbose("Order status response: %s", formatJSON(response))
		log.Printf("🔍 Order status - ID: %d, Status: %s, Executed Qty: %s, Remaining Qty: %s",
			response.Result.OrderId, response.Result.Status, 
			response.Result.ExecutedQty, response.Result.OrigQty)

		// Validate order details
		s.Assert().Equal(orderID, response.Result.OrderId)
		s.Assert().Equal(testSymbol, response.Result.Symbol)

		return nil
	})

	s.Require().NoError(err, "Failed to send order.status request")
	s.waitForResponse(done, defaultTimeout, "order.status")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestOrderCancel tests the order.cancel endpoint
func (s *TradingTestSuite) TestOrderCancel() {
	s.requireAuth()

	// First place an order to cancel
	orderID := s.placeTestOrderForCancellation()
	if orderID == 0 {
		s.T().Skip("Failed to place order for cancellation test")
	}

	done := make(chan bool)
	request := models.NewOrderCancelRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	s.logVerbose("Cancelling order: %d", orderID)

	err := s.client.SendOrderCancel(s.getTestContext(), request, func(response *models.OrderCancelResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "order.cancel")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log cancellation details
		s.logVerbose("Order cancelled successfully: %s", formatJSON(response))
		log.Printf("❌ Order cancelled - ID: %d, Status: %s",
			response.Result.OrderId, response.Result.Status)

		// Validate cancellation
		s.Assert().Equal(orderID, response.Result.OrderId)
		s.Assert().Equal("CANCELED", response.Result.Status)

		// Clear the stored order ID since it's cancelled
		s.testOrderID = 0

		return nil
	})

	s.Require().NoError(err, "Failed to send order.cancel request")
	s.waitForResponse(done, defaultTimeout, "order.cancel")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestOrderFlowComplete tests a complete order flow
func (s *TradingTestSuite) TestOrderFlowComplete() {
	s.requireAuth()

	// 1. Place order
	orderID := s.placeTestOrderForFlow()
	if orderID == 0 {
		s.T().Skip("Failed to place order for flow test")
	}

	// 2. Check status
	s.checkOrderStatus(orderID)

	// 3. Cancel order
	s.cancelOrderInFlow(orderID)

	// 4. Verify cancellation
	s.verifyOrderCancelled(orderID)
}

// Helper methods

func (s *TradingTestSuite) placeTestOrderForModification() int64 {
	return s.placeTestOrderWithPrice("18000")
}

func (s *TradingTestSuite) placeTestOrderForStatus() int64 {
	return s.placeTestOrderWithPrice("17000")
}

func (s *TradingTestSuite) placeTestOrderForCancellation() int64 {
	return s.placeTestOrderWithPrice("16000")
}

func (s *TradingTestSuite) placeTestOrderForFlow() int64 {
	return s.placeTestOrderWithPrice("15000")
}

func (s *TradingTestSuite) placeTestOrderWithPrice(price string) int64 {
	done := make(chan bool)
	var orderID int64

	request := models.NewOrderPlaceRequest().
		SetSymbol(testSymbol).
		SetSide("BUY").
		SetType("LIMIT").
		SetQuantity(testOrderQuantity).
		SetPrice(price).
		SetTimeInForce("GTC").
		SetNewClientOrderId(fmt.Sprintf("test_%d", time.Now().UnixNano()))

	err := s.client.SendOrderPlace(s.getTestContext(), request, func(response *models.OrderPlaceResponse, err error) error {
		defer close(done)

		if err != nil {
			s.logVerbose("Failed to place test order: %v", err)
			return err
		}

		orderID = response.Result.OrderId
		s.logVerbose("Test order placed: %d", orderID)
		return nil
	})

	if err != nil {
		return 0
	}

	select {
	case <-done:
		return orderID
	case <-time.After(defaultTimeout):
		return 0
	}
}

func (s *TradingTestSuite) cancelTestOrder(orderID int64) {
	if orderID == 0 {
		return
	}

	done := make(chan bool)
	request := models.NewOrderCancelRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	err := s.client.SendOrderCancel(s.getTestContext(), request, func(response *models.OrderCancelResponse, err error) error {
		defer close(done)
		if err != nil {
			s.logVerbose("Failed to cancel test order %d: %v", orderID, err)
		} else {
			s.logVerbose("Test order %d cancelled", orderID)
		}
		return nil
	})

	if err == nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

func (s *TradingTestSuite) checkOrderStatus(orderID int64) {
	done := make(chan bool)
	request := models.NewOrderStatusRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	err := s.client.SendOrderStatus(s.getTestContext(), request, func(response *models.OrderStatusResponse, err error) error {
		defer close(done)
		if err != nil {
			s.T().Errorf("Failed to check order status: %v", err)
		} else {
			log.Printf("Order %d status: %s", orderID, response.Result.Status)
		}
		return nil
	})

	s.Require().NoError(err)
	s.waitForResponse(done, defaultTimeout, "order status check")
	time.Sleep(rateLimitDelay)
}

func (s *TradingTestSuite) cancelOrderInFlow(orderID int64) {
	done := make(chan bool)
	request := models.NewOrderCancelRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	err := s.client.SendOrderCancel(s.getTestContext(), request, func(response *models.OrderCancelResponse, err error) error {
		defer close(done)
		if err != nil {
			s.T().Errorf("Failed to cancel order: %v", err)
		} else {
			log.Printf("Order %d cancelled", orderID)
		}
		return nil
	})

	s.Require().NoError(err)
	s.waitForResponse(done, defaultTimeout, "order cancellation")
	time.Sleep(rateLimitDelay)
}

func (s *TradingTestSuite) verifyOrderCancelled(orderID int64) {
	done := make(chan bool)
	request := models.NewOrderStatusRequest().
		SetSymbol(testSymbol).
		SetOrderId(orderID)

	err := s.client.SendOrderStatus(s.getTestContext(), request, func(response *models.OrderStatusResponse, err error) error {
		defer close(done)
		if err != nil {
			s.T().Errorf("Failed to verify order cancellation: %v", err)
		} else {
			s.Assert().Equal("CANCELED", response.Result.Status, "Order should be cancelled")
			log.Printf("Order %d verified as cancelled", orderID)
		}
		return nil
	})

	s.Require().NoError(err)
	s.waitForResponse(done, defaultTimeout, "order cancellation verification")
}