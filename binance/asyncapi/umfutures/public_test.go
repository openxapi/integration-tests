//go:build ignore

package wstest

import (
	"context"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
)

func TestTickerPrice(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue // Skip non-public configs - ticker price is a public endpoint
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "TickerPrice", testTickerPrice)
		})
	}
}

func TestBookTicker(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue // Skip non-public configs - book ticker is a public endpoint
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "BookTicker", testBookTicker)
		})
	}
}

func TestDepth(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue // Skip non-public configs - depth is a public endpoint
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "Depth", testDepth)
		})
	}
}

// Implementation functions
func testTickerPrice(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendTickerPrice(ctx,
		models.NewTickerPriceRequest().SetSymbol("BTCUSDT"),
		func(response *models.TickerPriceResponse, err error) error {
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

func testBookTicker(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendTickerBook(ctx,
		models.NewTickerBookRequest().SetSymbol("BTCUSDT"),
		func(response *models.TickerBookResponse, err error) error {
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

func testDepth(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendDepth(ctx,
		models.NewDepthRequest().
			SetSymbol("BTCUSDT").
			SetLimit(100),
		func(response *models.DepthResponse, err error) error {
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