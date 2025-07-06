//go:build ignore

package wstest

import (
	"context"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
)

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


// Implementation functions
func testAccountBalance(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendAccountBalance(ctx, models.NewAccountBalanceRequest(),
		func(response *models.AccountBalanceResponse, err error) error {
			responseChan <- err
			return err
		})

	if err != nil {
		return err
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
			responseChan <- err
			return err
		})

	if err != nil {
		return err
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
			responseChan <- err
			return err
		})

	if err != nil {
		return err
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
			responseChan <- err
			return err
		})

	if err != nil {
		return err
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
			responseChan <- err
			return err
		})

	if err != nil {
		return err
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
			responseChan <- err
			return err
		})

	if err != nil {
		return err
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

