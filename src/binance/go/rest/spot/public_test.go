package main

import (
	"context"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestExchangeInfo tests the exchange info endpoint
func TestExchangeInfo(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "ExchangeInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Test with no parameters (get all symbols)
				req := client.SpotTradingAPI.GetExchangeInfoV3(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get exchange info: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Timezone == nil || *resp.Timezone == "" {
					t.Error("Expected timezone in response")
				}
				
				if resp.ServerTime == nil || *resp.ServerTime == 0 {
					t.Error("Expected server time in response")
				}
				
				if resp.Symbols == nil || len(resp.Symbols) == 0 {
					t.Error("Expected symbols in response")
				}
				
				// Test with specific symbol
				rateLimiter.WaitForRateLimit()
				req2 := client.SpotTradingAPI.GetExchangeInfoV3(ctx).Symbol("BTCUSDT")
				resp2, _, err2 := req2.Execute()
				if err2 != nil {
					checkAPIError(t, err2)
					t.Fatalf("Failed to get exchange info for BTCUSDT: %v", err2)
				}
				
				if resp2.Symbols == nil || len(resp2.Symbols) != 1 {
					t.Error("Expected exactly one symbol in response")
				}
				
				if len(resp2.Symbols) > 0 && resp2.Symbols[0].Symbol != nil && *resp2.Symbols[0].Symbol != "BTCUSDT" {
					t.Errorf("Expected symbol BTCUSDT, got %s", *resp2.Symbols[0].Symbol)
				}
			})
		})
	}
}

// TestServerTime tests the server time endpoint
func TestServerTime(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "ServerTime", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetTimeV3(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get server time: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.ServerTime == nil || *resp.ServerTime == 0 {
					t.Error("Expected server time in response")
				}
				
				// Check that server time is reasonable (within 5 minutes of local time)
				serverTime := time.Unix(*resp.ServerTime/1000, 0)
				timeDiff := time.Since(serverTime).Abs()
				if timeDiff > 5*time.Minute {
					t.Errorf("Server time differs from local time by %v", timeDiff)
				}
			})
		})
	}
}

// TestMarketDepth tests the order book depth endpoint
func TestMarketDepth(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "MarketDepth", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Test with default limit
				req := client.SpotTradingAPI.GetDepthV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get market depth: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.LastUpdateId == nil || *resp.LastUpdateId == 0 {
					t.Error("Expected lastUpdateId in response")
				}
				
				if resp.Bids == nil || len(resp.Bids) == 0 {
					t.Error("Expected bids in response")
				}
				
				if resp.Asks == nil || len(resp.Asks) == 0 {
					t.Error("Expected asks in response")
				}
				
				// Test with specific limit
				rateLimiter.WaitForRateLimit()
				req2 := client.SpotTradingAPI.GetDepthV3(ctx).Symbol("BTCUSDT").Limit(5)
				resp2, _, err2 := req2.Execute()
				if err2 != nil {
					checkAPIError(t, err2)
					t.Fatalf("Failed to get market depth with limit: %v", err2)
				}
				
				if len(resp2.Bids) > 5 {
					t.Errorf("Expected max 5 bids, got %d", len(resp2.Bids))
				}
				
				if len(resp2.Asks) > 5 {
					t.Errorf("Expected max 5 asks, got %d", len(resp2.Asks))
				}
			})
		})
	}
}

// TestRecentTrades tests the recent trades endpoint
func TestRecentTrades(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "RecentTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetTradesV3(ctx).Symbol("BTCUSDT").Limit(10)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get recent trades: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected trades in response")
				}
				
				if len(resp) > 10 {
					t.Errorf("Expected max 10 trades, got %d", len(resp))
				}
				
				// Check first trade structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.Id == nil || *trade.Id == 0 {
						t.Error("Expected trade ID")
					}
					if trade.Price == nil || *trade.Price == "" {
						t.Error("Expected trade price")
					}
					if trade.Qty == nil || *trade.Qty == "" {
						t.Error("Expected trade quantity")
					}
					if trade.Time == nil || *trade.Time == 0 {
						t.Error("Expected trade time")
					}
				}
			})
		})
	}
}

// TestKlines tests the klines/candlestick endpoint
func TestKlines(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetKlinesV3(ctx).
					Symbol("BTCUSDT").
					Interval("1m").
					Limit(10)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get klines: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected klines in response")
				}
				
				if len(resp) > 10 {
					t.Errorf("Expected max 10 klines, got %d", len(resp))
				}
				
				// Check kline structure
				if len(resp) > 0 {
					kline := resp[0]
					if len(kline) < 12 {
						t.Errorf("Expected kline to have at least 12 fields, got %d", len(kline))
					}
				}
			})
		})
	}
}

// Test24hrTicker tests the 24hr ticker statistics endpoint
func Test24hrTicker(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "24hrTicker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Test 24hr ticker with single symbol
				req := client.SpotTradingAPI.GetTicker24hrV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get 24hr ticker: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp == nil {
					t.Fatal("Expected ticker response")
				}
				
				// The response can be either a single ticker or an array
				// Check if it's a valid response by trying to access the raw value
				if resp.SpotGetTicker24hrV3RespItem != nil {
					ticker := resp.SpotGetTicker24hrV3RespItem
					if ticker.Symbol == nil || *ticker.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT")
					}
					if ticker.LastPrice == nil || *ticker.LastPrice == "" {
						t.Error("Expected last price")
					}
					if ticker.Volume == nil || *ticker.Volume == "" {
						t.Error("Expected volume")
					}
				} else if resp.ArrayOfSpotGetTicker24hrV3RespItem != nil && len(*resp.ArrayOfSpotGetTicker24hrV3RespItem) > 0 {
					tickers := *resp.ArrayOfSpotGetTicker24hrV3RespItem
					if len(tickers) != 1 {
						t.Errorf("Expected 1 ticker, got %d", len(tickers))
					}
					ticker := tickers[0]
					if ticker.Symbol == nil || *ticker.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT")
					}
				} else {
					t.Error("Unexpected response format")
				}
			})
		})
	}
}

// TestAveragePrice tests the average price endpoint
func TestAveragePrice(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AveragePrice", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetAvgPriceV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get average price: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Mins == nil || *resp.Mins == 0 {
					t.Error("Expected mins in response")
				}
				
				if resp.Price == nil || *resp.Price == "" {
					t.Error("Expected price in response")
				}
				
				// Verify price is a valid number
				if resp.Price != nil {
					price, err := getCurrentPrice(client, ctx, "BTCUSDT")
					if err != nil {
						t.Errorf("Failed to parse price: %v", err)
					}
					if price <= 0 {
						t.Error("Expected positive price")
					}
				}
			})
		})
	}
}

// TestPing tests the ping endpoint
func TestPing(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "Ping", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetPingV3(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to ping server: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Ping endpoint returns empty response
				_ = resp
			})
		})
	}
}

// TestAggTrades tests the aggregated trades endpoint
func TestAggTrades(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AggTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetAggTradesV3(ctx).
					Symbol("BTCUSDT").
					Limit(10)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get aggregated trades: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected aggregated trades in response")
				}
				
				if len(resp) > 10 {
					t.Errorf("Expected max 10 trades, got %d", len(resp))
				}
				
				// Check first trade structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.A == nil || *trade.A == 0 {
						t.Error("Expected aggregate trade ID")
					}
					if trade.P == nil || *trade.P == "" {
						t.Error("Expected price")
					}
					if trade.Q == nil || *trade.Q == "" {
						t.Error("Expected quantity")
					}
					if trade.T == nil || *trade.T == 0 {
						t.Error("Expected timestamp")
					}
				}
			})
		})
	}
}

// TestHistoricalTrades tests the historical trades endpoint
func TestHistoricalTrades(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "HistoricalTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetHistoricalTradesV3(ctx).
					Symbol("BTCUSDT").
					Limit(10)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// Historical trades might require authentication on some exchanges
					if httpResp != nil && httpResp.StatusCode == 401 {
						t.Skip("Historical trades endpoint requires authentication")
					}
					t.Fatalf("Failed to get historical trades: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected historical trades in response")
				}
				
				if len(resp) > 10 {
					t.Errorf("Expected max 10 trades, got %d", len(resp))
				}
			})
		})
	}
}

// TestTicker24hr tests the 24hr ticker endpoint (alternative endpoint)
func TestTicker24hr(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "Ticker24hr", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetTicker24hrV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get 24hr ticker: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp == nil {
					t.Fatal("Expected ticker response")
				}
				
				// Check response based on type
				if resp.SpotGetTicker24hrV3RespItem != nil {
					ticker := resp.SpotGetTicker24hrV3RespItem
					if ticker.Symbol == nil || *ticker.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT")
					}
					if ticker.PriceChange == nil || *ticker.PriceChange == "" {
						t.Error("Expected price change")
					}
					if ticker.PriceChangePercent == nil || *ticker.PriceChangePercent == "" {
						t.Error("Expected price change percent")
					}
				} else if resp.ArrayOfSpotGetTicker24hrV3RespItem != nil {
					t.Error("Expected single ticker, got array")
				}
			})
		})
	}
}

// TestTickerPrice tests the ticker price endpoint
func TestTickerPrice(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "TickerPrice", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetTickerPriceV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get ticker price: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp == nil {
					t.Fatal("Expected ticker price response")
				}
				
				// Check response based on type
				if resp.SpotGetTickerPriceV3RespItem != nil {
					ticker := resp.SpotGetTickerPriceV3RespItem
					if ticker.Symbol == nil || *ticker.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT")
					}
					if ticker.Price == nil || *ticker.Price == "" {
						t.Error("Expected price")
					}
				} else if resp.ArrayOfSpotGetTickerPriceV3RespItem != nil {
					t.Error("Expected single ticker, got array")
				}
			})
		})
	}
}

// TestTickerBookTicker tests the best price/qty on the order book
func TestTickerBookTicker(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "TickerBookTicker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetTickerBookTickerV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get book ticker: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp == nil {
					t.Fatal("Expected book ticker response")
				}
				
				// Check response based on type
				if resp.SpotGetTickerBookTickerV3RespItem != nil {
					ticker := resp.SpotGetTickerBookTickerV3RespItem
					if ticker.Symbol == nil || *ticker.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT")
					}
					if ticker.BidPrice == nil || *ticker.BidPrice == "" {
						t.Error("Expected bid price")
					}
					if ticker.AskPrice == nil || *ticker.AskPrice == "" {
						t.Error("Expected ask price")
					}
					if ticker.BidQty == nil || *ticker.BidQty == "" {
						t.Error("Expected bid quantity")
					}
					if ticker.AskQty == nil || *ticker.AskQty == "" {
						t.Error("Expected ask quantity")
					}
				} else if resp.ArrayOfSpotGetTickerBookTickerV3RespItem != nil {
					t.Error("Expected single ticker, got array")
				}
			})
		})
	}
}

// TestTickerTradingDay tests the trading day ticker
func TestTickerTradingDay(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "TickerTradingDay", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Test trading day ticker using symbol parameter (either symbol or symbols can be used)
				req := client.SpotTradingAPI.GetTickerTradingDayV3(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on all servers
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Trading day ticker endpoint not available")
					}
					t.Fatalf("Failed to get trading day ticker: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp == nil {
					t.Fatal("Expected trading day ticker response")
				}
			})
		})
	}
}

// TestUiKlines tests the UI klines endpoint
func TestUiKlines(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "UiKlines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetUiKlinesV3(ctx).
					Symbol("BTCUSDT").
					Interval("1m").
					Limit(10)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get UI klines: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected UI klines in response")
				}
				
				if len(resp) > 10 {
					t.Errorf("Expected max 10 klines, got %d", len(resp))
				}
			})
		})
	}
}

// TestRateLimitOrder tests the rate limit endpoint
func TestRateLimitOrder(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "RateLimitOrder", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetRateLimitOrderV3(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get rate limit info: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected rate limit info in response")
				}
				
				// Check rate limit structure
				if len(resp) > 0 {
					limit := resp[0]
					if limit.RateLimitType == nil || *limit.RateLimitType == "" {
						t.Error("Expected rate limit type")
					}
					if limit.Interval == nil || *limit.Interval == "" {
						t.Error("Expected interval")
					}
					if limit.IntervalNum == nil || *limit.IntervalNum == 0 {
						t.Error("Expected interval number")
					}
					if limit.Limit == nil || *limit.Limit == 0 {
						t.Error("Expected limit")
					}
				}
			})
		})
	}
}