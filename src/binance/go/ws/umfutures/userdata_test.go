package wstest

import (
	"context"
	"fmt"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
)

func TestUserDataStreamStart(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_STREAM {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UserDataStreamStart", testUserDataStreamStart)
		})
	}
}

func TestUserDataStreamPing(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_STREAM {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "UserDataStreamPing", testUserDataStreamPing, 20*time.Second)
		})
	}
}

func TestUserDataStreamStop(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_STREAM {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "UserDataStreamStop", testUserDataStreamStop, 20*time.Second)
		})
	}
}

func TestAccountBalance(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountBalance", testAccountBalance)
		})
	}
}

func TestAccountPosition(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountPosition", testAccountPosition)
		})
	}
}

func TestAccountStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountStatus", testAccountStatus)
		})
	}
}

func TestV2AccountBalance(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "V2AccountBalance", testV2AccountBalance)
		})
	}
}

func TestV2AccountPosition(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "V2AccountPosition", testV2AccountPosition)
		})
	}
}

func TestV2AccountStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeUSER_DATA {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "V2AccountStatus", testV2AccountStatus)
		})
	}
}

// Implementation functions for user stream management
func testUserDataStreamStart(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendUserDataStreamStart(ctx, models.NewUserDataStreamStartRequest(),
		func(response *models.UserDataStreamStartResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			if response.Result.ListenKey == "" {
				responseChan <- fmt.Errorf("listen key is empty")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send user data stream start request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testUserDataStreamPing(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First start a user data stream
	startChan := make(chan *models.UserDataStreamStartResponse, 1)
	startErrChan := make(chan error, 1)

	err := client.SendUserDataStreamStart(ctx, models.NewUserDataStreamStartRequest(),
		func(response *models.UserDataStreamStartResponse, err error) error {
			if err != nil {
				startErrChan <- err
			} else {
				startChan <- response
			}
			return err
		})

	if err != nil {
		return fmt.Errorf("failed to send start request: %w", err)
	}

	// Wait for start response
	select {
	case <-startChan:
		// User data stream started successfully
	case err := <-startErrChan:
		return fmt.Errorf("failed to start user data stream: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("start user data stream timeout")
	}

	// Now ping the user data stream
	responseChan := make(chan error, 1)

	err = client.SendUserDataStreamPing(ctx,
		models.NewUserDataStreamPingRequest(),
		func(response *models.UserDataStreamPingResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response - ping usually returns empty success response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send user data stream ping request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testUserDataStreamStop(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First start a user data stream
	startChan := make(chan *models.UserDataStreamStartResponse, 1)
	startErrChan := make(chan error, 1)

	err := client.SendUserDataStreamStart(ctx, models.NewUserDataStreamStartRequest(),
		func(response *models.UserDataStreamStartResponse, err error) error {
			if err != nil {
				startErrChan <- err
			} else {
				startChan <- response
			}
			return err
		})

	if err != nil {
		return fmt.Errorf("failed to send start request: %w", err)
	}

	// Wait for start response
	select {
	case <-startChan:
		// User data stream started successfully
	case err := <-startErrChan:
		return fmt.Errorf("failed to start user data stream: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("start user data stream timeout")
	}

	// Now stop the user data stream
	responseChan := make(chan error, 1)

	err = client.SendUserDataStreamStop(ctx,
		models.NewUserDataStreamStopRequest(),
		func(response *models.UserDataStreamStopResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response - stop usually returns empty success response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send user data stream stop request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Implementation functions for account data
func testAccountBalance(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendAccountBalance(ctx, models.NewAccountBalanceRequest(),
		func(response *models.AccountBalanceResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send account balance request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testAccountPosition(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendAccountPosition(ctx,
		models.NewAccountPositionRequest().SetSymbol("BTCUSDT"),
		func(response *models.AccountPositionResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send account position request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testAccountStatus(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendAccountStatus(ctx, models.NewAccountStatusRequest(),
		func(response *models.AccountStatusResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send account status request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// V2 method implementations
func testV2AccountBalance(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendV2AccountBalance(ctx, models.NewV2AccountBalanceRequest(),
		func(response *models.V2AccountBalanceResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send v2 account balance request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testV2AccountPosition(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendV2AccountPosition(ctx,
		models.NewV2AccountPositionRequest().SetSymbol("BTCUSDT"),
		func(response *models.V2AccountPositionResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send v2 account position request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testV2AccountStatus(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendV2AccountStatus(ctx, models.NewV2AccountStatusRequest(),
		func(response *models.V2AccountStatusResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			// Result is a value type, not a pointer

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send v2 account status request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}