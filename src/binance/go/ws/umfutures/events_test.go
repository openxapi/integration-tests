package wstest

import (
	"context"
	"fmt"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
)


func TestUserDataEventHandlers(t *testing.T) {
	for _, config := range getTestConfigs() {
		// Test event handler registration - this doesn't require auth
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UserDataEventHandlers", testUserDataEventHandlers)
		})
		break // Only run once since event handler registration doesn't depend on config
	}
}

// testUserDataEventHandlers tests that all event handlers can be registered
func testUserDataEventHandlers(client *umfuturesws.Client, config TestConfig) error {
	// Register all event handlers to ensure they can be registered without errors
	// These are silent handlers that prevent "No handler found" warnings

	client.HandleAccountConfigUpdateEvent(func(event *models.AccountConfigUpdateEvent) error {
		return nil
	})

	client.HandleAccountUpdateEvent(func(event *models.AccountUpdateEvent) error {
		return nil
	})

	client.HandleOrderTradeUpdateEvent(func(event *models.OrderTradeUpdateEvent) error {
		return nil
	})

	client.HandleConditionalOrderTriggerRejectEvent(func(event *models.ConditionalOrderTriggerRejectEvent) error {
		return nil
	})

	client.HandleGridUpdateEvent(func(event *models.GridUpdateEvent) error {
		return nil
	})

	client.HandleListenKeyExpiredEvent(func(event *models.ListenKeyExpiredEvent) error {
		return nil
	})

	client.HandleMarginCallEvent(func(event *models.MarginCallEvent) error {
		return nil
	})

	client.HandleStrategyUpdateEvent(func(event *models.StrategyUpdateEvent) error {
		return nil
	})

	client.HandleTradeLiteEvent(func(event *models.TradeLiteEvent) error {
		return nil
	})

	return nil
}

func TestSessionLogon(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.KeyType != KeyTypeED25519 {
			continue // SessionLogon requires Ed25519 keys only
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "SessionLogon", testSessionLogon, 20*time.Second)
		})
	}
}

func TestSessionLogout(t *testing.T) {
	for _, config := range getTestConfigs() {
		// Session logout requires NONE authentication (no API key needed)
		// But we still need a client setup, so using USER_STREAM configs
		if config.AuthType != AuthTypeUSER_STREAM {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "SessionLogout", testSessionLogout, 30*time.Second)
		})
	}
}

func TestSessionStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		// Session status requires NONE authentication (no API key needed)
		// But we still need a client setup, so using USER_STREAM configs
		if config.AuthType != AuthTypeUSER_STREAM {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "SessionStatus", testSessionStatus, 30*time.Second)
		})
	}
}

// Implementation functions

func testSessionLogon(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	// Generate timestamp and signature for session logon
	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("apiKey=%s&timestamp=%d", config.APIKey, timestamp)
	signature, err := generateSignature(config, queryString)
	if err != nil {
		return fmt.Errorf("failed to generate signature: %w", err)
	}

	err = client.SendSessionLogon(ctx,
		models.NewSessionLogonRequest().
			SetApiKey(config.APIKey).
			SetTimestamp(timestamp).
			SetSignature(signature),
		func(response *models.SessionLogonResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send session logon request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testSessionLogout(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Session logout requires NONE authentication - test it directly
	responseChan := make(chan error, 1)

	err := client.SendSessionLogout(ctx, models.NewSessionLogoutRequest(),
		func(response *models.SessionLogoutResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send session logout request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testSessionStatus(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Session status requires NONE authentication - test it directly
	responseChan := make(chan error, 1)

	err := client.SendSessionStatus(ctx, models.NewSessionStatusRequest(),
		func(response *models.SessionStatusResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send session status request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

