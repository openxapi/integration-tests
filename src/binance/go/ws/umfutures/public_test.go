package wstest

import (
	"context"
	"fmt"
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
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			if response.Result == nil {
				responseChan <- fmt.Errorf("received nil result in ticker price response")
				return nil
			}
			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			if response.Result.Price == "" {
				responseChan <- fmt.Errorf("price is empty")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send ticker price request: %w", err)
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
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

			if response.Result == nil {
				responseChan <- fmt.Errorf("received nil result in response")
				return nil
			}
			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			if response.Result.BidPrice == "" || response.Result.BidQty == "" {
				responseChan <- fmt.Errorf("bid price or quantity is empty")
				return nil
			}

			if response.Result.AskPrice == "" || response.Result.AskQty == "" {
				responseChan <- fmt.Errorf("ask price or quantity is empty")
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send book ticker request: %w", err)
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
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

	
			if len(response.Result.Bids) == 0 {
				responseChan <- fmt.Errorf("no bids in depth response")
				return nil
			}

			if len(response.Result.Asks) == 0 {
				responseChan <- fmt.Errorf("no asks in depth response")
				return nil
			}

			// Validate that bids and asks have price and quantity
			for i, bid := range response.Result.Bids {
				if len(bid) < 2 {
					responseChan <- fmt.Errorf("bid at index %d has insufficient data", i)
					return nil
				}
			}

			for i, ask := range response.Result.Asks {
				if len(ask) < 2 {
					responseChan <- fmt.Errorf("ask at index %d has insufficient data", i)
					return nil
				}
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send depth request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}