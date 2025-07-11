package wstest

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
)

func TestOrderPlace(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "OrderPlace", testOrderPlace)
		})
	}
}

func TestOrderStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "OrderStatus", testOrderStatus, 20*time.Second)
		})
	}
}

func TestOrderCancel(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "OrderCancel", testOrderCancel, 20*time.Second)
		})
	}
}

func TestOrderModify(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpointWithTimeout(t, config, "OrderModify", testOrderModify, 20*time.Second)
		})
	}
}

// getAccountBalanceForSymbol gets the available balance for a specific asset
func getAccountBalanceForSymbol(client *umfuturesws.Client, asset string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	responseChan := make(chan float64, 1)
	errChan := make(chan error, 1)

	err := client.SendAccountBalance(ctx,
		models.NewAccountBalanceRequest(),
		func(response *models.AccountBalanceResponse, err error) error {
			if err != nil {
				errChan <- err
				return err
			}

			// Find the USDT balance
			for _, balance := range response.Result {
				if balance.Asset == asset {
					availableBalance, parseErr := strconv.ParseFloat(balance.AvailableBalance, 64)
					if parseErr != nil {
						errChan <- fmt.Errorf("failed to parse balance: %w", parseErr)
						return nil
					}
					responseChan <- availableBalance
					return nil
				}
			}
			errChan <- fmt.Errorf("asset %s not found in account", asset)
			return nil
		})

	if err != nil {
		return 0, fmt.Errorf("failed to send account balance request: %w", err)
	}

	select {
	case balance := <-responseChan:
		return balance, nil
	case err := <-errChan:
		return 0, err
	case <-ctx.Done():
		return 0, fmt.Errorf("account balance timeout")
	}
}

// calculateOrderQuantity calculates an appropriate order quantity based on minimum notional
func calculateOrderQuantity(price float64, minNotional float64) string {
	// Calculate quantity needed for minimum notional + 10% buffer
	requiredQuantity := (minNotional * 1.1) / price
	
	// Round to step size of 0.001 for BTCUSDT
	return roundQuantity(requiredQuantity, 0.001)
}

// roundToTickSize rounds a price to the correct tick size
// For BTCUSDT futures, the tick size is typically 0.01
func roundToTickSize(price float64, tickSize float64) string {
	// Round to the nearest tick
	rounded := math.Round(price/tickSize) * tickSize
	
	// Determine decimal places from tick size
	decimalPlaces := 0
	temp := tickSize
	for temp < 1 && decimalPlaces < 8 {
		temp *= 10
		decimalPlaces++
	}
	
	// Format with exact decimal places
	format := fmt.Sprintf("%%.%df", decimalPlaces)
	return fmt.Sprintf(format, rounded)
}

// roundQuantity rounds a quantity to the correct step size
// For BTCUSDT futures, the quantity step size is typically 0.001
func roundQuantity(quantity float64, stepSize float64) string {
	// Round to the nearest step
	rounded := math.Round(quantity/stepSize) * stepSize
	
	// For BTCUSDT, quantity precision is typically 3 decimal places
	return fmt.Sprintf("%.3f", rounded)
}

// Implementation functions
func testOrderPlace(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First check account balance
	usdtBalance, err := getAccountBalanceForSymbol(client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	// Ensure we have enough balance for minimum order
	if usdtBalance < 110 { // $100 minimum + buffer
		return fmt.Errorf("insufficient USDT balance: %.2f (need at least 110 USDT for testing)", usdtBalance)
	}

	// Get current price to place a far-out limit order
	currentPrice, err := getCurrentPrice(client, "BTCUSDT")
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Place a buy order at 50% of current price to ensure it won't execute
	orderPrice := currentPrice * 0.5
	// Use tick size of 0.1 for BTCUSDT futures
	orderPriceStr := roundToTickSize(orderPrice, 0.1)
	
	// Parse the rounded price string back to float for quantity calculation
	orderPriceFloat, _ := strconv.ParseFloat(orderPriceStr, 64)
	
	// Calculate quantity for $110 notional value at the order price
	quantity := calculateOrderQuantity(orderPriceFloat, 100.0)
	
	// Log for debugging
	fmt.Printf("DEBUG OrderPlace: currentPrice=%.2f, orderPrice=%.2f, orderPriceStr=%s, quantity=%s\n", 
		currentPrice, orderPrice, orderPriceStr, quantity)

	responseChan := make(chan error, 1)

	err = client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity(quantity).
			SetPrice(orderPriceStr),
		func(response *models.OrderPlaceResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

	
			if response.Result.OrderId == 0 {
				responseChan <- fmt.Errorf("order ID is 0")
				return nil
			}

			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send order place request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testOrderStatus(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First check account balance
	usdtBalance, err := getAccountBalanceForSymbol(client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	// Ensure we have enough balance for minimum order
	if usdtBalance < 110 { // $100 minimum + buffer
		return fmt.Errorf("insufficient USDT balance: %.2f (need at least 110 USDT for testing)", usdtBalance)
	}

	// Get current price to place a far-out limit order
	currentPrice, err := getCurrentPrice(client, "BTCUSDT")
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Place a buy order at 50% of current price to ensure it won't execute
	orderPrice := currentPrice * 0.5
	// Use tick size of 0.1 for BTCUSDT futures
	orderPriceStr := roundToTickSize(orderPrice, 0.1)
	
	// Parse the rounded price string back to float for quantity calculation
	orderPriceFloat, _ := strconv.ParseFloat(orderPriceStr, 64)
	quantity := calculateOrderQuantity(orderPriceFloat, 100.0)

	// First create an order to get an order ID
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err = client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity(quantity).
			SetPrice(orderPriceStr),
		func(response *models.OrderPlaceResponse, err error) error {
			if err != nil {
				orderErrChan <- err
			} else {
				orderChan <- response
			}
			return err
		})

	if err != nil {
		return fmt.Errorf("failed to send create order request: %w", err)
	}

	// Wait for order creation
	var orderID int64
	select {
	case response := <-orderChan:
		if response.Result == nil {
			return fmt.Errorf("received nil result in order response")
		}
		orderID = response.Result.OrderId
	case err := <-orderErrChan:
		return fmt.Errorf("failed to create order for status test: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("create order timeout")
	}

	// Now check order status
	responseChan := make(chan error, 1)

	err = client.SendOrderStatus(ctx,
		models.NewOrderStatusRequest().
			SetSymbol("BTCUSDT").
			SetOrderId(orderID),
		func(response *models.OrderStatusResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

	
			if response.Result.OrderId != orderID {
				responseChan <- fmt.Errorf("expected order ID %d, got %d", orderID, response.Result.OrderId)
				return nil
			}

			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send order status request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testOrderCancel(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First check account balance
	usdtBalance, err := getAccountBalanceForSymbol(client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	// Ensure we have enough balance for minimum order
	if usdtBalance < 110 { // $100 minimum + buffer
		return fmt.Errorf("insufficient USDT balance: %.2f (need at least 110 USDT for testing)", usdtBalance)
	}

	// Get current price to place a far-out limit order
	currentPrice, err := getCurrentPrice(client, "BTCUSDT")
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Place a buy order at 50% of current price to ensure it won't execute
	orderPrice := currentPrice * 0.5
	// Use tick size of 0.1 for BTCUSDT futures
	orderPriceStr := roundToTickSize(orderPrice, 0.1)
	quantity := calculateOrderQuantity(orderPrice, 100.0)

	// First create an order to cancel
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err = client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity(quantity).
			SetPrice(orderPriceStr),
		func(response *models.OrderPlaceResponse, err error) error {
			if err != nil {
				orderErrChan <- err
			} else {
				orderChan <- response
			}
			return err
		})

	if err != nil {
		return fmt.Errorf("failed to send create order request: %w", err)
	}

	// Wait for order creation
	var orderID int64
	select {
	case response := <-orderChan:
		if response.Result == nil {
			return fmt.Errorf("received nil result in order response")
		}
		orderID = response.Result.OrderId
	case err := <-orderErrChan:
		return fmt.Errorf("failed to create order for cancel test: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("create order timeout")
	}

	// Now cancel the order
	responseChan := make(chan error, 1)

	err = client.SendOrderCancel(ctx,
		models.NewOrderCancelRequest().
			SetSymbol("BTCUSDT").
			SetOrderId(orderID),
		func(response *models.OrderCancelResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

	
			if response.Result.OrderId != orderID {
				responseChan <- fmt.Errorf("expected order ID %d, got %d", orderID, response.Result.OrderId)
				return nil
			}

			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			if response.Result.Status != "CANCELED" {
				responseChan <- fmt.Errorf("expected status CANCELED, got %s", response.Result.Status)
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send order cancel request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func testOrderModify(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First check account balance
	usdtBalance, err := getAccountBalanceForSymbol(client, "USDT")
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	// Ensure we have enough balance for minimum order
	if usdtBalance < 110 { // $100 minimum + buffer
		return fmt.Errorf("insufficient USDT balance: %.2f (need at least 110 USDT for testing)", usdtBalance)
	}

	// Get current price to place a far-out limit order
	currentPrice, err := getCurrentPrice(client, "BTCUSDT")
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Place a buy order at 50% of current price to ensure it won't execute
	orderPrice := currentPrice * 0.5
	// Use tick size of 0.1 for BTCUSDT futures
	orderPriceStr := roundToTickSize(orderPrice, 0.1)
	// For modify, use 48% of current price (not too far from original to avoid issues)
	modifiedPrice := currentPrice * 0.48
	modifiedPriceStr := roundToTickSize(modifiedPrice, 0.1)
	// Parse the rounded price strings back to float for quantity calculation
	orderPriceFloat, _ := strconv.ParseFloat(orderPriceStr, 64)
	quantity := calculateOrderQuantity(orderPriceFloat, 100.0)
	
	modifiedPriceFloat, _ := strconv.ParseFloat(modifiedPriceStr, 64)
	// For modified order, ensure we have enough buffer since we're lowering the price
	modifiedQuantity := calculateOrderQuantity(modifiedPriceFloat, 105.0) // Use $105 minimum for safety
	
	// Log for debugging
	quantityFloat, _ := strconv.ParseFloat(quantity, 64)
	modifiedQuantityFloat, _ := strconv.ParseFloat(modifiedQuantity, 64)
	fmt.Printf("DEBUG OrderModify: currentPrice=%.2f, orderPrice=%.2f, orderPriceStr=%s, quantity=%s, notional=%.2f\n", 
		currentPrice, orderPrice, orderPriceStr, quantity, orderPriceFloat*quantityFloat)
	fmt.Printf("DEBUG OrderModify: modifiedPrice=%.2f, modifiedPriceStr=%s, modifiedQuantity=%s, notional=%.2f\n", 
		modifiedPrice, modifiedPriceStr, modifiedQuantity, modifiedPriceFloat*modifiedQuantityFloat)

	// First create an order to modify
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err = client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity(quantity).
			SetPrice(orderPriceStr),
		func(response *models.OrderPlaceResponse, err error) error {
			if err != nil {
				orderErrChan <- err
			} else {
				orderChan <- response
			}
			return err
		})

	if err != nil {
		return fmt.Errorf("failed to send create order request: %w", err)
	}

	// Wait for order creation
	var orderID int64
	select {
	case response := <-orderChan:
		if response.Result == nil {
			return fmt.Errorf("received nil result in order response")
		}
		orderID = response.Result.OrderId
	case err := <-orderErrChan:
		return fmt.Errorf("failed to create order for modify test: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("create order timeout")
	}

	// Now modify the order
	responseChan := make(chan error, 1)

	err = client.SendOrderModify(ctx,
		models.NewOrderModifyRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetOrderId(orderID).
			SetQuantity(modifiedQuantity).
			SetPrice(modifiedPriceStr),
		func(response *models.OrderModifyResponse, err error) error {
			if err != nil {
				responseChan <- err
				return err
			}

			// Validate response
			if response == nil {
				responseChan <- fmt.Errorf("response is nil")
				return nil
			}

	
			if response.Result.OrderId == 0 {
				responseChan <- fmt.Errorf("modified order ID is 0")
				return nil
			}

			if response.Result.Symbol != "BTCUSDT" {
				responseChan <- fmt.Errorf("expected symbol BTCUSDT, got %s", response.Result.Symbol)
				return nil
			}

			responseChan <- nil
			return nil
		})

	if err != nil {
		return fmt.Errorf("failed to send order modify request: %w", err)
	}

	select {
	case err := <-responseChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}