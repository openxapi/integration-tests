package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestGetMarginAccount tests getting margin account details
func TestGetMarginAccount(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginAccount", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginAccountV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin account not available on testnet")
					}
					t.Fatalf("Failed to get margin account: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.MarginLevel == nil || *resp.MarginLevel == "" {
					t.Error("Expected margin level")
				}
				
				if resp.TotalAssetOfBtc == nil || *resp.TotalAssetOfBtc == "" {
					t.Error("Expected total asset in BTC")
				}
				
				if resp.TotalLiabilityOfBtc == nil || *resp.TotalLiabilityOfBtc == "" {
					t.Error("Expected total liability in BTC")
				}
				
				if resp.TradeEnabled == nil {
					t.Error("Expected trade enabled flag")
				}
				
				t.Logf("Margin Level: %s, Trade Enabled: %v", 
					*resp.MarginLevel, *resp.TradeEnabled)
			})
		})
	}
}

// TestGetMarginAllAssets tests getting all margin assets
func TestGetMarginAllAssets(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginAllAssets", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginAllAssetsV1(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin assets endpoint not available on testnet")
					}
					t.Fatalf("Failed to get margin assets: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected margin assets in response")
				}
				
				// Check for common assets
				foundBTC := false
				foundUSDT := false
				
				for _, asset := range resp {
					if asset.AssetFullName != nil && asset.AssetName != nil {
						if *asset.AssetName == "BTC" {
							foundBTC = true
							if asset.IsBorrowable == nil {
								t.Error("Expected borrowable flag for BTC")
							}
							if asset.IsMortgageable == nil {
								t.Error("Expected mortgageable flag for BTC")
							}
						} else if *asset.AssetName == "USDT" {
							foundUSDT = true
						}
					}
				}
				
				if !foundBTC {
					t.Error("Expected to find BTC in margin assets")
				}
				if !foundUSDT {
					t.Error("Expected to find USDT in margin assets")
				}
				
				t.Logf("Found %d margin assets", len(resp))
			})
		})
	}
}

// TestGetMarginAllPairs tests getting all margin pairs
func TestGetMarginAllPairs(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginAllPairs", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginAllPairsV1(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin pairs endpoint not available on testnet")
					}
					t.Fatalf("Failed to get margin pairs: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected margin pairs in response")
				}
				
				// Check for BTCUSDT pair
				foundBTCUSDT := false
				
				for _, pair := range resp {
					if pair.Symbol != nil && *pair.Symbol == "BTCUSDT" {
						foundBTCUSDT = true
						if pair.Base == nil || *pair.Base != "BTC" {
							t.Error("Expected base BTC for BTCUSDT")
						}
						if pair.Quote == nil || *pair.Quote != "USDT" {
							t.Error("Expected quote USDT for BTCUSDT")
						}
						if pair.IsMarginTrade == nil || !*pair.IsMarginTrade {
							t.Error("Expected margin trade enabled for BTCUSDT")
						}
						if pair.IsBuyAllowed == nil {
							t.Error("Expected buy allowed flag")
						}
						if pair.IsSellAllowed == nil {
							t.Error("Expected sell allowed flag")
						}
						break
					}
				}
				
				if !foundBTCUSDT {
					t.Error("Expected to find BTCUSDT in margin pairs")
				}
				
				t.Logf("Found %d margin pairs", len(resp))
			})
		})
	}
}

// TestGetMarginPriceIndex tests getting margin price index
func TestGetMarginPriceIndex(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginPriceIndex", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginPriceIndexV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin price index endpoint not available on testnet")
					}
					t.Fatalf("Failed to get margin price index: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Symbol == nil || *resp.Symbol != "BTCUSDT" {
					t.Errorf("Expected symbol BTCUSDT, got %v", resp.Symbol)
				}
				
				if resp.Price == nil || *resp.Price == "" {
					t.Error("Expected price in response")
				}
				
				if resp.CalcTime == nil || *resp.CalcTime == 0 {
					t.Error("Expected calculation time")
				}
				
				t.Logf("BTCUSDT margin price index: %s", *resp.Price)
			})
		})
	}
}

// TestCreateMarginOrder tests creating a margin order
func TestCreateMarginOrder(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateMarginOrder", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get current price
				price, err := getCurrentPrice(client, ctx, "BTCUSDT")
				if err != nil {
					t.Fatalf("Failed to get current price: %v", err)
				}
				
				// Set limit price below market
				limitPrice := price * 0.5
				limitPriceStr := fmt.Sprintf("%.2f", limitPrice)
				
				req := client.MarginTradingAPI.CreateMarginOrderV1(ctx).
					Symbol("BTCUSDT").
					Side("BUY").
					Type_("LIMIT").
					TimeInForce("GTC").
					Quantity("0.0001").
					Price(limitPriceStr).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if handleTestnetError(t, err, httpResp, "Create margin order") {
					return
				}
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin trading might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin trading not available on this account")
					}
					t.Fatalf("Failed to create margin order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderId == nil || *resp.OrderId == 0 {
					t.Error("Expected order ID")
				}
				
				if resp.Symbol == nil || *resp.Symbol != "BTCUSDT" {
					t.Errorf("Expected symbol BTCUSDT, got %v", resp.Symbol)
				}
				
				if resp.Status == nil {
					t.Error("Expected order status")
				}
				
				if resp.IsIsolated == nil {
					t.Error("Expected isolated flag")
				}
				
				// Cancel the order
				if resp.OrderId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.MarginTradingAPI.DeleteMarginOrderV1(ctx).
						Symbol("BTCUSDT").
						OrderId(*resp.OrderId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel margin order: %v", cancelErr)
					}
				}
			})
		})
	}
}

// TestGetMarginAllOrders tests getting all margin orders
func TestGetMarginAllOrders(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginAllOrders", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginAllOrdersV1(ctx).
					Symbol("BTCUSDT").
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin orders not available on this account")
					}
					t.Fatalf("Failed to get margin orders: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty
				t.Logf("Found %d margin orders", len(resp))
				
				// If there are orders, verify structure
				if len(resp) > 0 {
					order := resp[0]
					if order.OrderId == nil || *order.OrderId == 0 {
						t.Error("Expected order ID")
					}
					if order.Symbol == nil || *order.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT, got %v", order.Symbol)
					}
					if order.IsIsolated == nil {
						t.Error("Expected isolated flag")
					}
				}
			})
		})
	}
}

// TestGetMarginMyTrades tests getting margin trades
func TestGetMarginMyTrades(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginMyTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginMyTradesV1(ctx).
					Symbol("BTCUSDT").
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin trades not available on this account")
					}
					t.Fatalf("Failed to get margin trades: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty
				t.Logf("Found %d margin trades", len(resp))
				
				// If there are trades, verify structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.Id == nil || *trade.Id == 0 {
						t.Error("Expected trade ID")
					}
					if trade.Symbol == nil || *trade.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT, got %v", trade.Symbol)
					}
					if trade.Price == nil || *trade.Price == "" {
						t.Error("Expected price")
					}
					if trade.Qty == nil || *trade.Qty == "" {
						t.Error("Expected quantity")
					}
					if trade.IsIsolated == nil {
						t.Error("Expected isolated flag")
					}
				}
			})
		})
	}
}

// TestGetMarginMaxBorrowable tests getting max borrowable amount
func TestGetMarginMaxBorrowable(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginMaxBorrowable", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginMaxBorrowableV1(ctx).
					Asset("USDT").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin max borrowable not available")
					}
					t.Fatalf("Failed to get max borrowable: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Amount == nil || *resp.Amount == "" {
					t.Error("Expected borrowable amount")
				} else {
					t.Logf("Max borrowable USDT: %s", *resp.Amount)
				}
				
				if resp.BorrowLimit != nil && *resp.BorrowLimit != "" {
					t.Logf("Borrow limit: %s", *resp.BorrowLimit)
				}
			})
		})
	}
}

// TestGetMarginInterestHistory tests getting interest history
func TestGetMarginInterestHistory(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMarginInterestHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.GetMarginInterestHistoryV1(ctx).
					Size(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin interest history not available")
					}
					t.Fatalf("Failed to get interest history: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Total == nil {
					t.Error("Expected total count")
				} else {
					t.Logf("Total interest records: %d", *resp.Total)
				}
				
				if resp.Rows != nil && len(resp.Rows) > 0 {
					interest := resp.Rows[0]
					if interest.Asset == nil || *interest.Asset == "" {
						t.Error("Expected asset")
					}
					if interest.Interest == nil || *interest.Interest == "" {
						t.Error("Expected interest amount")
					}
					if interest.InterestRate == nil || *interest.InterestRate == "" {
						t.Error("Expected interest rate")
					}
					if interest.InterestAccuredTime == nil || *interest.InterestAccuredTime == 0 {
						t.Error("Expected interest time")
					}
				}
			})
		})
	}
}

// TestCreateMarginListenKey tests creating a margin listen key
func TestCreateMarginListenKey(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateMarginListenKey", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.MarginTradingAPI.CreateMarginListenKeyV1(ctx)
				
				resp, httpResp, err := req.Execute()
				if handleTestnetError(t, err, httpResp, "Create margin listen key") {
					return
				}
				if err != nil {
					checkAPIError(t, err)
					// Only skip 404 - margin might not be enabled on testnet
					// Never skip 400 - these are bad requests that need fixing
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Margin listen key not available")
					}
					t.Fatalf("Failed to create margin listen key: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.ListenKey == nil || *resp.ListenKey == "" {
					t.Error("Expected listen key")
				} else {
					listenKey := *resp.ListenKey
					t.Logf("Created margin listen key: %s", listenKey)
					
					// Update the listen key
					time.Sleep(1 * time.Second)
					rateLimiter.WaitForRateLimit()
					
					updateReq := client.MarginTradingAPI.UpdateMarginListenKeyV1(ctx).
						ListenKey(listenKey)
					
					_, httpResp, err = updateReq.Execute()
					if err != nil {
						t.Logf("Warning: Failed to update margin listen key: %v", err)
					} else if httpResp.StatusCode == 200 {
						t.Log("Successfully updated margin listen key")
					}
					
					// Delete the listen key
					time.Sleep(1 * time.Second)
					rateLimiter.WaitForRateLimit()
					
					deleteReq := client.MarginTradingAPI.DeleteMarginListenKeyV1(ctx)
					
					_, _, deleteErr := deleteReq.Execute()
					if deleteErr != nil {
						t.Logf("Warning: Failed to delete margin listen key: %v", deleteErr)
					}
				}
			})
		})
	}
}