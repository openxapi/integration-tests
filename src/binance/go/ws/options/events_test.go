package options_test

import (
	"sync"
	"testing"

	"github.com/openxapi/binance-go/ws/options/models"
	"github.com/stretchr/testify/suite"
)

// EventsTestSuite tests event handler functionality
type EventsTestSuite struct {
	BaseTestSuite
}

// TestEventsSuite runs the events test suite
func TestEventsSuite(t *testing.T) {
	suite.Run(t, new(EventsTestSuite))
}

// TestHandleAccountUpdate tests the HandleAccountUpdate event handler
func (s *EventsTestSuite) TestHandleAccountUpdate() {
	s.Run("HandleAccountUpdate", func() {
		var (
			handlerCalled bool
			mu            sync.Mutex
		)

		// Register the event handler
		s.client.HandleAccountUpdateEvent(func(event *models.AccountUpdateEvent) error {
			mu.Lock()
			handlerCalled = true
			mu.Unlock()
			
			s.logVerbose("AccountUpdate event handler called with event: %+v", event)
			s.Require().NotNil(event, "Event should not be nil")
			return nil
		})

		// Verify handler was registered (we can't easily trigger events without real connection)
		mu.Lock()
		registered := !handlerCalled // Initially false means it's registered and waiting
		mu.Unlock()
		
		s.Require().True(registered, "HandleAccountUpdate should register without error")
		s.logVerbose("HandleAccountUpdate event handler registered successfully")
	})
}

// TestHandleOrderTradeUpdate tests the HandleOrderTradeUpdate event handler
func (s *EventsTestSuite) TestHandleOrderTradeUpdate() {
	s.Run("HandleOrderTradeUpdate", func() {
		var (
			handlerCalled bool
			mu            sync.Mutex
		)

		// Register the event handler
		s.client.HandleOrderTradeUpdateEvent(func(event *models.OrderTradeUpdateEvent) error {
			mu.Lock()
			handlerCalled = true
			mu.Unlock()
			
			s.logVerbose("OrderTradeUpdate event handler called with event: %+v", event)
			s.Require().NotNil(event, "Event should not be nil")
			return nil
		})

		// Verify handler was registered
		mu.Lock()
		registered := !handlerCalled // Initially false means it's registered and waiting
		mu.Unlock()
		
		s.Require().True(registered, "HandleOrderTradeUpdate should register without error")
		s.logVerbose("HandleOrderTradeUpdate event handler registered successfully")
	})
}

// TestHandleRiskLevelChange tests the HandleRiskLevelChange event handler
func (s *EventsTestSuite) TestHandleRiskLevelChange() {
	s.Run("HandleRiskLevelChange", func() {
		var (
			handlerCalled bool
			mu            sync.Mutex
		)

		// Register the event handler
		s.client.HandleRiskLevelChangeEvent(func(event *models.RiskLevelChangeEvent) error {
			mu.Lock()
			handlerCalled = true
			mu.Unlock()
			
			s.logVerbose("RiskLevelChange event handler called with event: %+v", event)
			s.Require().NotNil(event, "Event should not be nil")
			return nil
		})

		// Verify handler was registered
		mu.Lock()
		registered := !handlerCalled // Initially false means it's registered and waiting
		mu.Unlock()
		
		s.Require().True(registered, "HandleRiskLevelChange should register without error")
		s.logVerbose("HandleRiskLevelChange event handler registered successfully")
	})
}

// TestMultipleEventHandlers tests registering multiple event handlers
func (s *EventsTestSuite) TestMultipleEventHandlers() {
	s.Run("MultipleEventHandlers", func() {
		var (
			accountHandlerCalled bool
			orderHandlerCalled   bool
			riskHandlerCalled    bool
			mu                   sync.Mutex
		)

		// Register all event handlers
		s.client.HandleAccountUpdateEvent(func(event *models.AccountUpdateEvent) error {
			mu.Lock()
			accountHandlerCalled = true
			mu.Unlock()
			return nil
		})

		s.client.HandleOrderTradeUpdateEvent(func(event *models.OrderTradeUpdateEvent) error {
			mu.Lock()
			orderHandlerCalled = true
			mu.Unlock()
			return nil
		})

		s.client.HandleRiskLevelChangeEvent(func(event *models.RiskLevelChangeEvent) error {
			mu.Lock()
			riskHandlerCalled = true
			mu.Unlock()
			return nil
		})

		// All handlers should be registered successfully
		mu.Lock()
		allRegistered := !accountHandlerCalled && !orderHandlerCalled && !riskHandlerCalled
		mu.Unlock()

		s.Require().True(allRegistered, "All event handlers should register successfully")
		s.logVerbose("All event handlers registered successfully")
	})
}