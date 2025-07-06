//go:build ignore

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
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UserDataStreamStart", testUserDataStreamStart)
		})
	}
}

func TestUserDataStreamPing(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UserDataStreamPing", testUserDataStreamPing)
		})
	}
}

func TestUserDataStreamStop(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UserDataStreamStop", testUserDataStreamStop)
		})
	}
}

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
			testEndpoint(t, config, "OrderStatus", testOrderStatus)
		})
	}
}

func TestOrderCancel(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "OrderCancel", testOrderCancel)
		})
	}
}

func TestOrderModify(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "OrderModify", testOrderModify)
		})
	}
}

// Implementation functions
func testUserDataStreamStart(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendUserDataStreamStart(ctx, models.NewUserDataStreamStartRequest(),
		func(response *models.UserDataStreamStartResponse, err error) error {
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

func testUserDataStreamPing(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First start a user data stream to get a listen key
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
	var listenKey string
	select {
	case response := <-startChan:
		listenKey = response.Result.ListenKey
	case err := <-startErrChan:
		return fmt.Errorf("failed to start user data stream: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("start user data stream timeout")
	}

	// Now ping the user data stream
	responseChan := make(chan error, 1)

	err = client.SendUserDataStreamPing(ctx,
		models.NewUserDataStreamPingRequest().SetListenKey(listenKey),
		func(response *models.UserDataStreamPingResponse, err error) error {
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

func testUserDataStreamStop(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First start a user data stream to get a listen key
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
	var listenKey string
	select {
	case response := <-startChan:
		listenKey = response.Result.ListenKey
	case err := <-startErrChan:
		return fmt.Errorf("failed to start user data stream: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("start user data stream timeout")
	}

	// Now stop the user data stream
	responseChan := make(chan error, 1)

	err = client.SendUserDataStreamStop(ctx,
		models.NewUserDataStreamStopRequest().SetListenKey(listenKey),
		func(response *models.UserDataStreamStopResponse, err error) error {
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

func testOrderPlace(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responseChan := make(chan error, 1)

	err := client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity("0.001").
			SetPrice("30000.00"),
		func(response *models.OrderPlaceResponse, err error) error {
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

func testOrderStatus(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First create an order to get an order ID
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err := client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity("0.001").
			SetPrice("30000.00"),
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

func testOrderCancel(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First create an order to cancel
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err := client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity("0.001").
			SetPrice("30000.00"),
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

func testOrderModify(client *umfuturesws.Client, config TestConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// First create an order to modify
	orderChan := make(chan *models.OrderPlaceResponse, 1)
	orderErrChan := make(chan error, 1)

	err := client.SendOrderPlace(ctx,
		models.NewOrderPlaceRequest().
			SetSymbol("BTCUSDT").
			SetSide("BUY").
			SetType("LIMIT").
			SetTimeInForce("GTC").
			SetQuantity("0.001").
			SetPrice("30000.00"),
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
			SetOrderId(orderID).
			SetQuantity("0.002").
			SetPrice("29000.00"),
		func(response *models.OrderModifyResponse, err error) error {
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