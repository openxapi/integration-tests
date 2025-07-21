package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestAlgoSpotOrders tests algo trading spot order endpoints
func TestAlgoSpotOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetAlgoSpotOpenOrders", func(t *testing.T) {
		resp, httpResp, err := client.AlgoTradingAPI.GetAlgoSpotOpenOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Algo spot open orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if algo trading not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -13003) {
						t.Skip("Algo trading not enabled or not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get algo spot open orders: %v", err)
		}

		t.Logf("Algo spot open orders: %+v", resp)
	})

	t.Run("GetAlgoSpotHistoricalOrders", func(t *testing.T) {
		resp, httpResp, err := client.AlgoTradingAPI.GetAlgoSpotHistoricalOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Algo spot historical orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if algo trading not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -13003) {
						t.Skip("Algo trading not enabled or not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get algo spot historical orders: %v", err)
		}

		t.Logf("Algo spot historical orders: %+v", resp)
	})

	t.Run("GetAlgoSpotSubOrders", func(t *testing.T) {
		// This requires an algo order ID
		algoOrderIdStr := os.Getenv("BINANCE_TEST_ALGO_ORDER_ID")
		if algoOrderIdStr == "" {
			t.Skip("BINANCE_TEST_ALGO_ORDER_ID not set")
		}
		
		algoOrderId, err := strconv.ParseInt(algoOrderIdStr, 10, 64)
		if err != nil {
			t.Fatalf("Invalid algo order ID: %v", err)
		}

		resp, _, err := client.AlgoTradingAPI.GetAlgoSpotSubOrdersV1(ctx).
			AlgoId(algoOrderId).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -13001 {
						t.Skip("Algo order not found")
					}
				}
			}
			t.Fatalf("Failed to get algo spot sub orders: %v", err)
		}

		t.Logf("Algo spot sub orders: %+v", resp)
	})
}

// TestAlgoFuturesOrders tests algo trading futures order endpoints
func TestAlgoFuturesOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetAlgoFuturesOpenOrders", func(t *testing.T) {
		resp, httpResp, err := client.AlgoTradingAPI.GetAlgoFuturesOpenOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Algo futures open orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if algo trading not enabled or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -13003 || code == -11002) {
						t.Skip("Algo trading or futures not enabled on account")
					}
				}
			}
			t.Fatalf("Failed to get algo futures open orders: %v", err)
		}

		t.Logf("Algo futures open orders: %+v", resp)
	})

	t.Run("GetAlgoFuturesHistoricalOrders", func(t *testing.T) {
		resp, httpResp, err := client.AlgoTradingAPI.GetAlgoFuturesHistoricalOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Algo futures historical orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if algo trading not enabled or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -13003 || code == -11002) {
						t.Skip("Algo trading or futures not enabled on account")
					}
				}
			}
			t.Fatalf("Failed to get algo futures historical orders: %v", err)
		}

		t.Logf("Algo futures historical orders: %+v", resp)
	})

	t.Run("GetAlgoFuturesSubOrders", func(t *testing.T) {
		// This requires an algo order ID
		algoOrderIdStr := os.Getenv("BINANCE_TEST_ALGO_FUTURES_ORDER_ID")
		if algoOrderIdStr == "" {
			t.Skip("BINANCE_TEST_ALGO_FUTURES_ORDER_ID not set")
		}
		
		algoOrderId, err := strconv.ParseInt(algoOrderIdStr, 10, 64)
		if err != nil {
			t.Fatalf("Invalid algo order ID: %v", err)
		}

		resp, _, err := client.AlgoTradingAPI.GetAlgoFuturesSubOrdersV1(ctx).
			AlgoId(algoOrderId).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -13001 {
						t.Skip("Algo order not found")
					}
				}
			}
			t.Fatalf("Failed to get algo futures sub orders: %v", err)
		}

		t.Logf("Algo futures sub orders: %+v", resp)
	})
}

// TestAlgoSpotTWAPOrder tests TWAP (Time-Weighted Average Price) orders for spot
func TestAlgoSpotTWAPOrder(t *testing.T) {
	// Skip by default to avoid creating real orders
	if os.Getenv("BINANCE_TEST_ALGO_ORDERS") != "true" {
		t.Skip("Set BINANCE_TEST_ALGO_ORDERS=true to test algo order creation")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateAlgoSpotTWAPOrder", func(t *testing.T) {
		// TWAP order configuration
		symbol := "BTCUSDT"
		side := "BUY"
		quantity := "0.001"
		duration := int64(300) // 5 minutes

		resp, httpResp, err := client.AlgoTradingAPI.CreateAlgoSpotNewOrderTwapV1(ctx).
			Symbol(symbol).
			Side(side).
			Quantity(quantity).
			Duration(duration).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create algo spot TWAP order") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or algo trading not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2010 || code == -13003) {
						t.Skip("Insufficient balance or algo trading not enabled")
					}
				}
			}
			t.Fatalf("Failed to create algo spot TWAP order: %v", err)
		}

		t.Logf("Algo spot TWAP order created: %+v", resp)

		// Test canceling the order
		if resp.ClientAlgoId != nil && resp.Success != nil && *resp.Success {
			t.Run("CancelAlgoSpotOrder", func(t *testing.T) {
				// Note: The cancel endpoint might require a different ID
				// This is a placeholder implementation
				t.Logf("Order created successfully with ClientAlgoId: %s", *resp.ClientAlgoId)
				t.Skip("Cancel implementation depends on SDK structure")
			})
		}
	})
}

// TestAlgoFuturesTWAPOrder tests TWAP orders for futures
func TestAlgoFuturesTWAPOrder(t *testing.T) {
	// Skip by default to avoid creating real orders
	if os.Getenv("BINANCE_TEST_ALGO_ORDERS") != "true" {
		t.Skip("Set BINANCE_TEST_ALGO_ORDERS=true to test algo order creation")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateAlgoFuturesTWAPOrder", func(t *testing.T) {
		// TWAP order configuration
		symbol := "BTCUSDT"
		side := "BUY"
		quantity := "0.001"
		duration := int64(300) // 5 minutes

		resp, httpResp, err := client.AlgoTradingAPI.CreateAlgoFuturesNewOrderTwapV1(ctx).
			Symbol(symbol).
			Side(side).
			Quantity(quantity).
			Duration(duration).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create algo futures TWAP order") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance, algo trading not enabled, or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2010 || code == -13003 || code == -11002) {
						t.Skip("Insufficient balance, algo trading not enabled, or futures not enabled")
					}
				}
			}
			t.Fatalf("Failed to create algo futures TWAP order: %v", err)
		}

		t.Logf("Algo futures TWAP order created: %+v", resp)

		// Test canceling the order
		if resp.ClientAlgoId != nil && resp.Success != nil && *resp.Success {
			t.Run("CancelAlgoFuturesOrder", func(t *testing.T) {
				// Note: The cancel endpoint might require a different ID
				// This is a placeholder implementation
				t.Logf("Order created successfully with ClientAlgoId: %s", *resp.ClientAlgoId)
				t.Skip("Cancel implementation depends on SDK structure")
			})
		}
	})
}

// TestAlgoFuturesVPOrder tests VP (Volume Participation) orders for futures
func TestAlgoFuturesVPOrder(t *testing.T) {
	// Skip by default to avoid creating real orders
	if os.Getenv("BINANCE_TEST_ALGO_ORDERS") != "true" {
		t.Skip("Set BINANCE_TEST_ALGO_ORDERS=true to test algo order creation")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateAlgoFuturesVPOrder", func(t *testing.T) {
		// VP order configuration
		symbol := "BTCUSDT"
		side := "BUY"
		quantity := "0.001"
		urgency := "LOW"

		resp, httpResp, err := client.AlgoTradingAPI.CreateAlgoFuturesNewOrderVpV1(ctx).
			Symbol(symbol).
			Side(side).
			Quantity(quantity).
			Urgency(urgency).
			ClientAlgoId("test_vp_" + strconv.FormatInt(timestamp, 10)).
			ReduceOnly(false).
			LimitPrice("40000").
			PositionSide("BOTH").
			RecvWindow(5000).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create algo futures VP order") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance, algo trading not enabled, or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2010 || code == -13003 || code == -11002) {
						t.Skip("Insufficient balance, algo trading not enabled, or futures not enabled")
					}
				}
			}
			t.Fatalf("Failed to create algo futures VP order: %v", err)
		}

		t.Logf("Algo futures VP order created: %+v", resp)

		// Test canceling the order if created
		if resp.ClientAlgoId != nil && resp.Success != nil && *resp.Success {
			t.Logf("VP order created successfully with ClientAlgoId: %s", *resp.ClientAlgoId)
			// Note: Cancellation would require the actual order ID from the system
		}
	})
}