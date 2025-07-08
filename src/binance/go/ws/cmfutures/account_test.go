package cmfutures_test

import (
	"log"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/cmfutures"
	"github.com/openxapi/binance-go/ws/cmfutures/models"
	"github.com/stretchr/testify/suite"
)

// AccountTestSuite tests account-related APIs
type AccountTestSuite struct {
	BaseTestSuite
}

// TestAccountTestSuite runs the account test suite
func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

// TestAccountBalance tests the account.balance endpoint
func (s *AccountTestSuite) TestAccountBalance() {
	s.requireAuth()

	done := make(chan bool)
	request := models.NewAccountBalanceRequest()

	s.logVerbose("Sending account.balance request")

	err := s.client.SendAccountBalance(s.getTestContext(), request, func(response *models.AccountBalanceResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "account.balance")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log balance information
		s.logVerbose("Account balance response: %s", formatJSON(response))

		if response.Result != nil {
			log.Printf("💰 Found %d asset balances", len(response.Result))
			for _, balance := range response.Result {
				if balance.Balance != "0" || balance.AvailableBalance != "0" {
					log.Printf("  💎 Asset: %s, Balance: %s, Available: %s", 
						balance.Asset, balance.Balance, balance.AvailableBalance)
				}
			}
		}

		return nil
	})

	s.Require().NoError(err, "Failed to send account.balance request")
	s.waitForResponse(done, defaultTimeout, "account.balance")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestAccountPosition tests the account.position endpoint
func (s *AccountTestSuite) TestAccountPosition() {
	s.requireAuth()

	done := make(chan bool)
	request := models.NewAccountPositionRequest()

	s.logVerbose("Sending account.position request")

	err := s.client.SendAccountPosition(s.getTestContext(), request, func(response *models.AccountPositionResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "account.position")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log position information
		s.logVerbose("Account position response: %s", formatJSON(response))

		if response.Result != nil {
			log.Printf("📈 Found %d positions", len(response.Result))
			for _, position := range response.Result {
				if position.PositionAmt != "0" {
					log.Printf("  📊 Symbol: %s, Position: %s, Entry Price: %s, Mark Price: %s, PNL: %s",
						position.Symbol, position.PositionAmt, position.EntryPrice, 
						position.MarkPrice, position.UnRealizedProfit)
				}
			}
		}

		return nil
	})

	s.Require().NoError(err, "Failed to send account.position request")
	s.waitForResponse(done, defaultTimeout, "account.position")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestAccountStatus tests the account.status endpoint
func (s *AccountTestSuite) TestAccountStatus() {
	s.requireAuth()

	done := make(chan bool)
	request := models.NewAccountStatusRequest()

	s.logVerbose("Sending account.status request")

	err := s.client.SendAccountStatus(s.getTestContext(), request, func(response *models.AccountStatusResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "account.status")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")

		// Log status information
		s.logVerbose("Account status response: %s", formatJSON(response))

		log.Printf("🔒 Account Status:")
		log.Printf("  ✅ Can Trade: %v", response.Result.CanTrade)
		log.Printf("  💸 Can Withdraw: %v", response.Result.CanWithdraw)
		log.Printf("  💵 Can Deposit: %v", response.Result.CanDeposit)
		log.Printf("  🎯 Fee Tier: %d", response.Result.FeeTier)

		// Check if there are any trading restrictions
		if response.Result.Assets != nil {
			for _, asset := range response.Result.Assets {
				if asset.MarginBalance != "0" {
					log.Printf("  📄 Asset with margin - Asset: %s, Margin Balance: %s", 
						asset.Asset, asset.MarginBalance)
				}
			}
		}

		return nil
	})

	s.Require().NoError(err, "Failed to send account.status request")
	s.waitForResponse(done, defaultTimeout, "account.status")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestAccountMultipleRequests tests multiple account requests in sequence
func (s *AccountTestSuite) TestAccountMultipleRequests() {
	s.requireAuth()

	// Test rapid sequential requests with proper rate limiting
	operations := []struct {
		name string
		fn   func()
	}{
		{"balance", func() { s.TestAccountBalance() }},
		{"position", func() { s.TestAccountPosition() }},
		{"status", func() { s.TestAccountStatus() }},
	}

	for _, op := range operations {
		s.Run(op.name, op.fn)
		// Additional delay between different operations
		time.Sleep(200 * time.Millisecond)
	}
}

// TestAccountErrorHandling tests error handling for account endpoints
func (s *AccountTestSuite) TestAccountErrorHandling() {
	// Skip this test if no auth is available to avoid connection issues
	if testAPIKey == "" || testSecretKey == "" {
		s.T().Skip("Skipping error handling test: requires authentication")
	}

	// Test with invalid context (no auth)
	done := make(chan bool)
	request := models.NewAccountBalanceRequest()
	
	// Create a client without auth
	unauthClient := cmfutures.NewClient()
	err := unauthClient.SetActiveServer("testnet1")
	s.Require().NoError(err)
	
	err = unauthClient.Connect(s.ctx)
	s.Require().NoError(err)
	defer unauthClient.Disconnect()

	// This should fail with authentication error
	err = unauthClient.SendAccountBalance(s.ctx, request, func(response *models.AccountBalanceResponse, err error) error {
		defer close(done)

		// We expect an error here
		s.Require().Error(err, "Should receive authentication error")
		s.logVerbose("Received expected error: %v", err)

		return nil
	})

	// The initial send might fail if no auth is configured
	if err != nil {
		s.logVerbose("Send failed as expected: %v", err)
		s.Assert().Contains(err.Error(), "authentication required", "Error should mention authentication")
	} else {
		// Wait for the response handler to process the error
		s.waitForResponse(done, defaultTimeout, "account.balance error test")
	}
}