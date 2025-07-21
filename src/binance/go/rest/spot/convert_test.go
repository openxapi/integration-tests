package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestConvertInfo tests convert information endpoints
func TestConvertInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetConvertExchangeInfo", func(t *testing.T) {
		resp, httpResp, err := client.ConvertAPI.GetConvertExchangeInfoV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Convert exchange info") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if convert not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Convert not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get convert exchange info: %v", err)
		}

		t.Logf("Convert exchange info: %+v", resp)
	})

	t.Run("GetConvertAssetInfo", func(t *testing.T) {
		resp, httpResp, err := client.ConvertAPI.GetConvertAssetInfoV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Convert asset info") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if convert not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Convert not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get convert asset info: %v", err)
		}

		t.Logf("Convert asset info: %+v", resp)
	})

	t.Run("GetConvertTradeFlow", func(t *testing.T) {
		// Get trade flow for the last 30 days
		startTime := time.Now().Add(-30 * 24 * time.Hour).UnixMilli()
		endTime := time.Now().UnixMilli()

		resp, httpResp, err := client.ConvertAPI.GetConvertTradeFlowV1(ctx).
			StartTime(startTime).
			EndTime(endTime).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Convert trade flow") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get convert trade flow: %v", err)
		}

		t.Logf("Convert trade flow: %+v", resp)
	})
}

// TestConvertQuote tests convert quote endpoints
func TestConvertQuote(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetConvertQuote", func(t *testing.T) {
		resp, httpResp, err := client.ConvertAPI.CreateConvertGetQuoteV1(ctx).
			FromAsset("USDT").
			ToAsset("BTC").
			FromAmount("10").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get convert quote") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if convert not available or invalid pair
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -1000) {
						t.Skip("Convert not available on testnet or invalid pair")
					}
				}
			}
			t.Fatalf("Failed to get convert quote: %v", err)
		}

		t.Logf("Convert quote: %+v", resp)
		
		// Store quote ID for accept test
		if resp.QuoteId != nil {
			ctx = context.WithValue(ctx, "quoteId", *resp.QuoteId)
		}
	})
}

// TestConvertOrders tests convert order endpoints
func TestConvertOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	t.Run("GetConvertOrderStatus", func(t *testing.T) {
		// This requires an order ID
		orderId := os.Getenv("BINANCE_TEST_CONVERT_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_CONVERT_ORDER_ID not set")
		}

		resp, _, err := client.ConvertAPI.GetConvertOrderStatusV1(ctx).
			OrderId(orderId).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1001 {
						t.Skip("Convert order not found")
					}
				}
			}
			t.Fatalf("Failed to get convert order status: %v", err)
		}

		t.Logf("Convert order status: %+v", resp)
	})
}

// TestConvertLimitOrders tests convert limit order endpoints
func TestConvertLimitOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("QueryConvertLimitOpenOrders", func(t *testing.T) {
		resp, httpResp, err := client.ConvertAPI.CreateConvertLimitQueryOpenOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Query convert limit open orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if convert limit orders not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Convert limit orders not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to query convert limit open orders: %v", err)
		}

		t.Logf("Convert limit open orders: %+v", resp)
	})
}

// TestConvertOperations tests convert trading operations (use with caution)
func TestConvertOperations(t *testing.T) {
	// Skip by default to avoid actual conversions
	if os.Getenv("BINANCE_TEST_CONVERT_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_CONVERT_OPERATIONS=true to test convert operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	var quoteId string

	t.Run("CreateConvertQuote", func(t *testing.T) {
		resp, httpResp, err := client.ConvertAPI.CreateConvertGetQuoteV1(ctx).
			FromAsset("USDT").
			ToAsset("BUSD").
			FromAmount("1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create convert quote") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or invalid pair
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -2010) {
						t.Skip("Convert not available, insufficient balance, or invalid pair")
					}
				}
			}
			t.Fatalf("Failed to create convert quote: %v", err)
		}

		t.Logf("Convert quote created: %+v", resp)
		
		if resp.QuoteId != nil {
			quoteId = *resp.QuoteId
		}
	})

	t.Run("AcceptConvertQuote", func(t *testing.T) {
		if quoteId == "" {
			t.Skip("No quote ID available")
		}

		resp, _, err := client.ConvertAPI.CreateConvertAcceptQuoteV1(ctx).
			QuoteId(quoteId).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if quote expired or invalid
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1002 || code == -1003) {
						t.Skip("Quote expired or invalid")
					}
				}
			}
			t.Fatalf("Failed to accept convert quote: %v", err)
		}

		t.Logf("Convert quote accepted: %+v", resp)
	})

	t.Run("PlaceConvertLimitOrder", func(t *testing.T) {
		// Get current price first
		priceResp, _, err := client.ConvertAPI.CreateConvertGetQuoteV1(ctx).
			FromAsset("USDT").
			ToAsset("BTC").
			FromAmount("10").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			t.Skip("Failed to get price quote for limit order")
		}

		var limitPrice string
		if priceResp.ToAmount != nil {
			// Set limit price slightly below current price
			price, _ := strconv.ParseFloat(*priceResp.ToAmount, 64)
			limitPrice = strconv.FormatFloat(price*0.95, 'f', 8, 64)
		} else {
			t.Skip("No price available for limit order")
		}

		resp, _, err := client.ConvertAPI.CreateConvertLimitPlaceOrderV1(ctx).
			BaseAsset("USDT").
			QuoteAsset("BTC").
			LimitPrice(limitPrice).
			Side("BUY").
			ExpiredType("1_D"). // 1 day expiry
			BaseAmount("10").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or feature not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2010 || code == -1121) {
						t.Skip("Insufficient balance or convert limit orders not available")
					}
				}
			}
			t.Fatalf("Failed to place convert limit order: %v", err)
		}

		t.Logf("Convert limit order placed: %+v", resp)

		// Note: The response doesn't contain an OrderId
		if resp.QuoteId != nil {
			t.Logf("Convert limit order created with QuoteId: %s", *resp.QuoteId)
		}
	})
}