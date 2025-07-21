package main

import (
	"context"
	"fmt"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestCreateOrderOco tests creating an OCO order
func TestCreateOrderOco(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateOrderOco", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get current price
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				// For SELL OCO: need BTC balance but more typical use case
				// Take profit (limit order) above market, stop loss below market
				takeProfitPrice := price * 1.05  // Take profit above market
				takeProfitPriceStr := fmt.Sprintf("%.2f", takeProfitPrice)
				
				stopLossPrice := price * 0.95   // Stop loss below market
				stopLossPriceStr := fmt.Sprintf("%.2f", stopLossPrice)
				
				stopLimitPrice := price * 0.94  // Stop limit below stop price
				stopLimitPriceStr := fmt.Sprintf("%.2f", stopLimitPrice)
				
				req := client.SpotTradingAPI.CreateOrderOcoV3(ctx).
					Symbol("BTCUSDT").
					Side("SELL").
					Quantity("0.0001").
					Price(takeProfitPriceStr).
					StopPrice(stopLossPriceStr).
					StopLimitPrice(stopLimitPriceStr).
					StopLimitTimeInForce("GTC").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to create OCO order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderListId == nil || *resp.OrderListId == 0 {
					t.Error("Expected order list ID")
				}
				
				if resp.ContingencyType == nil || *resp.ContingencyType != "OCO" {
					t.Errorf("Expected contingency type OCO, got %v", resp.ContingencyType)
				}
				
				if resp.ListStatusType == nil || *resp.ListStatusType == "" {
					t.Error("Expected list status type")
				}
				
				if resp.Orders == nil || len(resp.Orders) != 2 {
					t.Error("Expected 2 orders in OCO response")
				}
				
				// Cancel the OCO order
				if resp.OrderListId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.SpotTradingAPI.DeleteOrderListV3(ctx).
						Symbol("BTCUSDT").
						OrderListId(*resp.OrderListId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel OCO order: %v", cancelErr)
					}
				}
			})
		})
	}
}

// TestCreateOrderListOco tests creating an order list OCO
func TestCreateOrderListOco(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateOrderListOco", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get current price
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				// For SELL OCO: above order (take profit) above market, below order (stop loss) below market
				takeProfitPrice := price * 1.05  // Take profit above market
				takeProfitPriceStr := fmt.Sprintf("%.2f", takeProfitPrice)
				
				stopLossPrice := price * 0.95    // Stop loss below market
				stopLossPriceStr := fmt.Sprintf("%.2f", stopLossPrice)
				
				stopLimitPrice := price * 0.94   // Stop limit below stop price
				stopLimitPriceStr := fmt.Sprintf("%.2f", stopLimitPrice)
				
				req := client.SpotTradingAPI.CreateOrderListOcoV3(ctx).
					Symbol("BTCUSDT").
					Side("SELL").
					Quantity("0.0001").
					AboveType("LIMIT_MAKER").
					AbovePrice(takeProfitPriceStr).
					BelowType("STOP_LOSS_LIMIT").
					BelowPrice(stopLimitPriceStr).
					BelowStopPrice(stopLossPriceStr).
					BelowTimeInForce("GTC").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to create order list OCO: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderListId == nil || *resp.OrderListId == 0 {
					t.Error("Expected order list ID")
				}
				
				if resp.ContingencyType == nil || *resp.ContingencyType != "OCO" {
					t.Errorf("Expected contingency type OCO, got %v", resp.ContingencyType)
				}
				
				// Cancel the order list
				if resp.OrderListId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.SpotTradingAPI.DeleteOrderListV3(ctx).
						Symbol("BTCUSDT").
						OrderListId(*resp.OrderListId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel order list: %v", cancelErr)
					}
				}
			})
		})
	}
}

// TestGetOrderList tests retrieving an order list
func TestGetOrderList(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetOrderList", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// First create an OCO order to query
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				takeProfitPrice := price * 1.05
				takeProfitPriceStr := fmt.Sprintf("%.2f", takeProfitPrice)
				
				stopLossPrice := price * 0.95
				stopLossPriceStr := fmt.Sprintf("%.2f", stopLossPrice)
				
				stopLimitPrice := price * 0.94
				stopLimitPriceStr := fmt.Sprintf("%.2f", stopLimitPrice)
				
				createReq := client.SpotTradingAPI.CreateOrderOcoV3(ctx).
					Symbol("BTCUSDT").
					Side("SELL").
					Quantity("0.0001").
					Price(takeProfitPriceStr).
					StopPrice(stopLossPriceStr).
					StopLimitPrice(stopLimitPriceStr).
					StopLimitTimeInForce("GTC").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				createResp, _, err := createReq.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to create OCO order for query test: %v", err)
				}
				
				if createResp.OrderListId == nil {
					t.Fatal("No order list ID returned")
				}
				
				// Query the order list
				rateLimiter.WaitForRateLimit()
				req := client.SpotTradingAPI.GetOrderListV3(ctx).
					OrderListId(*createResp.OrderListId).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get order list: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderListId == nil || *resp.OrderListId != *createResp.OrderListId {
					t.Errorf("Order list ID mismatch: expected %d, got %v", 
						*createResp.OrderListId, resp.OrderListId)
				}
				
				if resp.ContingencyType == nil || *resp.ContingencyType != "OCO" {
					t.Errorf("Expected contingency type OCO, got %v", resp.ContingencyType)
				}
				
				if resp.Orders == nil || len(resp.Orders) != 2 {
					t.Error("Expected 2 orders in order list")
				}
				
				// Clean up - cancel the order list
				rateLimiter.WaitForRateLimit()
				cancelReq := client.SpotTradingAPI.DeleteOrderListV3(ctx).
					Symbol("BTCUSDT").
					OrderListId(*createResp.OrderListId).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				_, _, cancelErr := cancelReq.Execute()
				if cancelErr != nil {
					t.Logf("Warning: Failed to cancel order list: %v", cancelErr)
				}
			})
		})
	}
}

// TestGetOpenOrderList tests retrieving open order lists
func TestGetOpenOrderList(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetOpenOrderList", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetOpenOrderListV3(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get open order lists: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty if no open order lists
				t.Logf("Found %d open order lists", len(resp))
				
				// If there are open order lists, verify structure
				if len(resp) > 0 {
					orderList := resp[0]
					if orderList.OrderListId == nil || *orderList.OrderListId == 0 {
						t.Error("Expected order list ID")
					}
					if orderList.ContingencyType == nil || *orderList.ContingencyType == "" {
						t.Error("Expected contingency type")
					}
					if orderList.ListStatusType == nil || *orderList.ListStatusType == "" {
						t.Error("Expected list status type")
					}
				}
			})
		})
	}
}

// TestGetAllOrderList tests retrieving all order lists
func TestGetAllOrderList(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAllOrderList", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetAllOrderListV3(ctx).
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get all order lists: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty if no order lists exist
				t.Logf("Found %d order lists", len(resp))
				
				// If there are order lists, verify structure
				if len(resp) > 0 {
					orderList := resp[0]
					if orderList.OrderListId == nil || *orderList.OrderListId == 0 {
						t.Error("Expected order list ID")
					}
					if orderList.ContingencyType == nil || *orderList.ContingencyType == "" {
						t.Error("Expected contingency type")
					}
					if orderList.ListStatusType == nil || *orderList.ListStatusType == "" {
						t.Error("Expected list status type")
					}
					if orderList.Orders == nil {
						t.Error("Expected orders in order list")
					}
				}
			})
		})
	}
}

// TestCreateOrderListOto tests creating an OTO (One-Triggers-Other) order
func TestCreateOrderListOto(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateOrderListOto", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get current price
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				// Set working order price below market
				workingPrice := price * 0.95
				workingPriceStr := fmt.Sprintf("%.2f", workingPrice)
				
				// Set pending order price
				pendingPrice := price * 0.90
				pendingPriceStr := fmt.Sprintf("%.2f", pendingPrice)
				
				req := client.SpotTradingAPI.CreateOrderListOtoV3(ctx).
					Symbol("BTCUSDT").
					WorkingType("LIMIT").
					WorkingSide("BUY").
					WorkingPrice(workingPriceStr).
					WorkingQuantity("0.0001").
					WorkingTimeInForce("GTC").
					PendingType("LIMIT").
					PendingSide("SELL").
					PendingPrice(pendingPriceStr).
					PendingQuantity("0.0001").
					PendingTimeInForce("GTC").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Never skip 400 - these are bad requests that need fixing
					// If OTO returns 400, it indicates a real API issue that needs investigation
					t.Fatalf("Failed to create OTO order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderListId == nil || *resp.OrderListId == 0 {
					t.Error("Expected order list ID")
				}
				
				if resp.ContingencyType == nil || *resp.ContingencyType != "OTO" {
					t.Errorf("Expected contingency type OTO, got %v", resp.ContingencyType)
				}
				
				// Cancel the order list
				if resp.OrderListId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.SpotTradingAPI.DeleteOrderListV3(ctx).
						Symbol("BTCUSDT").
						OrderListId(*resp.OrderListId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel OTO order: %v", cancelErr)
					}
				}
			})
		})
	}
}

// TestCreateOrderListOtoco tests creating an OTOCO (One-Triggers-One-Cancels-Other) order
func TestCreateOrderListOtoco(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateOrderListOtoco", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get current price
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				// Set working order price below market
				workingPrice := price * 0.95
				workingPriceStr := fmt.Sprintf("%.2f", workingPrice)
				
				// Set pending limit and stop prices
				pendingLimitPrice := price * 1.05
				pendingLimitPriceStr := fmt.Sprintf("%.2f", pendingLimitPrice)
				
				pendingStopPrice := price * 0.90
				pendingStopPriceStr := fmt.Sprintf("%.2f", pendingStopPrice)
				
				pendingStopLimitPrice := price * 0.89
				pendingStopLimitPriceStr := fmt.Sprintf("%.2f", pendingStopLimitPrice)
				
				req := client.SpotTradingAPI.CreateOrderListOtocoV3(ctx).
					Symbol("BTCUSDT").
					WorkingType("LIMIT").
					WorkingSide("BUY").
					WorkingPrice(workingPriceStr).
					WorkingQuantity("0.0001").
					WorkingTimeInForce("GTC").
					PendingSide("SELL").
					PendingQuantity("0.0001").
					PendingAboveType("LIMIT_MAKER").
					PendingAbovePrice(pendingLimitPriceStr).
					PendingBelowType("STOP_LOSS_LIMIT").
					PendingBelowStopPrice(pendingStopPriceStr).
					PendingBelowPrice(pendingStopLimitPriceStr).
					PendingBelowTimeInForce("GTC").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Never skip 400 - these are bad requests that need fixing
					// If OTOCO returns 400, it indicates a real API issue that needs investigation
					t.Fatalf("Failed to create OTOCO order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderListId == nil || *resp.OrderListId == 0 {
					t.Error("Expected order list ID")
				}
				
				if resp.ContingencyType == nil || *resp.ContingencyType != "OTO" {
					t.Errorf("Expected contingency type OTO, got %v", resp.ContingencyType)
				}
				
				// Cancel the order list
				if resp.OrderListId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.SpotTradingAPI.DeleteOrderListV3(ctx).
						Symbol("BTCUSDT").
						OrderListId(*resp.OrderListId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel OTOCO order: %v", cancelErr)
					}
				}
			})
		})
	}
}