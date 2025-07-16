package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/umfutures"
)

// TestPing tests the ping endpoint
func TestPing(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Ping", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetPingV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetPingV1")
					t.Fatalf("Error calling GetPingV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				t.Logf("Ping successful: %+v", resp)
			})
			break
		}
	}
}

// TestServerTime tests the server time endpoint
func TestServerTime(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Server Time", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetTimeV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetTimeV1")
					t.Fatalf("Error calling GetTimeV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.ServerTime == nil {
					t.Fatal("ServerTime should not be nil")
				}
				
				serverTime := *resp.ServerTime
				now := time.Now().UnixMilli()
				
				// Check if server time is within 10 seconds of local time
				if abs(serverTime-now) > 10000 {
					t.Logf("Warning: Server time (%d) differs significantly from local time (%d)", serverTime, now)
				}
				
				t.Logf("Server time: %d", serverTime)
			})
			break
		}
	}
}

// TestExchangeInfo tests the exchange info endpoint
func TestExchangeInfo(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Exchange Info", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetExchangeInfoV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					// Known issue: deliveryDate field type mismatch (int32 vs int64)
					if strings.Contains(err.Error(), "cannot unmarshal number") && strings.Contains(err.Error(), "deliveryDate") {
						t.Logf("Known SDK issue detected: deliveryDate field type mismatch (int32 vs int64): %v", err)
						logResponseBody(t, httpResp, "GetExchangeInfoV1")
						t.Fatalf("SDK Error - deliveryDate field type mismatch (int32 vs int64): %v", err)
					}
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetExchangeInfoV1")
					t.Fatalf("Error calling GetExchangeInfoV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.Symbols == nil || len(resp.Symbols) == 0 {
					t.Fatal("Symbols should not be empty")
				}
				
				if resp.Assets == nil || len(resp.Assets) == 0 {
					t.Fatal("Assets should not be empty")
				}
				
				t.Logf("Found %d symbols and %d assets", len(resp.Symbols), len(resp.Assets))
				
				// Check first symbol structure
				if len(resp.Symbols) > 0 {
					symbol := resp.Symbols[0]
					if symbol.Symbol == nil || *symbol.Symbol == "" {
						t.Fatal("Symbol name should not be empty")
					}
					if symbol.Status == nil || *symbol.Status == "" {
						t.Fatal("Symbol status should not be empty")
					}
					t.Logf("First symbol: %s, Status: %s", *symbol.Symbol, *symbol.Status)
				}
			})
			break
		}
	}
}

// TestOrderBook tests the order book endpoint
func TestOrderBook(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Order Book", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetDepthV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					// Known issue: timestamp field type mismatch (int32 vs int64)
					if strings.Contains(err.Error(), "cannot unmarshal number") && strings.Contains(err.Error(), "int32") {
						t.Logf("Known SDK issue detected: timestamp field type mismatch (int32 vs int64): %v", err)
						logResponseBody(t, httpResp, "GetDepthV1")
						t.Fatalf("SDK Error - timestamp field type mismatch (int32 vs int64): %v", err)
					}
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetDepthV1")
					t.Fatalf("Error calling GetDepthV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.Bids == nil || len(resp.Bids) == 0 {
					t.Fatal("Bids should not be empty")
				}
				
				if resp.Asks == nil || len(resp.Asks) == 0 {
					t.Fatal("Asks should not be empty")
				}
				
				t.Logf("Order book for BTCUSDT - Bids: %d, Asks: %d", len(resp.Bids), len(resp.Asks))
			})
			break
		}
	}
}

// TestRecentTrades tests the recent trades endpoint
func TestRecentTrades(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Recent Trades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetTradesV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetTradesV1")
					t.Fatalf("Error calling GetTradesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Trades should not be empty")
				}
				
				t.Logf("Found %d recent trades for BTCUSDT", len(resp))
				
				// Check first trade structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.Id == nil {
						t.Fatal("Trade ID should not be nil")
					}
					if trade.Price == nil || *trade.Price == "" {
						t.Fatal("Trade price should not be empty")
					}
					if trade.Qty == nil || *trade.Qty == "" {
						t.Fatal("Trade quantity should not be empty")
					}
					t.Logf("First trade: ID=%d, Price=%s, Qty=%s", *trade.Id, *trade.Price, *trade.Qty)
				}
			})
			break
		}
	}
}

// TestAggTrades tests the aggregate trades endpoint
func TestAggTrades(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Aggregate Trades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetAggTradesV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetAggTradesV1")
					t.Fatalf("Error calling GetAggTradesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Aggregate trades should not be empty")
				}
				
				t.Logf("Found %d aggregate trades for BTCUSDT", len(resp))
				
				// Check first trade structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.A == nil {
						t.Fatal("Aggregate trade ID should not be nil")
					}
					if trade.P == nil || *trade.P == "" {
						t.Fatal("Trade price should not be empty")
					}
					if trade.Q == nil || *trade.Q == "" {
						t.Fatal("Trade quantity should not be empty")
					}
					t.Logf("First agg trade: ID=%d, Price=%s, Qty=%s", *trade.A, *trade.P, *trade.Q)
				}
			})
			break
		}
	}
}

// TestKlines tests the klines endpoint
func TestKlines(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetKlinesV1(ctx).Symbol("BTCUSDT").Interval("1m")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetKlinesV1")
					t.Fatalf("Error calling GetKlinesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Klines should not be empty")
				}
				
				t.Logf("Found %d klines for BTCUSDT 1m", len(resp))
				
				// Check first kline structure
				if len(resp) > 0 {
					kline := resp[0]
					if len(kline) < 6 {
						t.Fatal("Kline should have at least 6 elements")
					}
					t.Logf("First kline: %+v", kline)
				}
			})
			break
		}
	}
}

// Test24hrTicker tests the 24hr ticker endpoint
func Test24hrTicker(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "24hr Ticker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetTicker24hrV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetTicker24hrV1")
					t.Fatalf("Error calling GetTicker24hrV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Handle both single and array response
				if resp.UmfuturesGetTicker24hrV1RespItem != nil {
					// Single item response
					ticker := resp.UmfuturesGetTicker24hrV1RespItem
					if ticker.Symbol == nil || *ticker.Symbol == "" {
						t.Fatal("Ticker symbol should not be empty")
					}
					if ticker.LastPrice == nil || *ticker.LastPrice == "" {
						t.Fatal("Last price should not be empty")
					}
					t.Logf("24hr ticker: %s, LastPrice=%s", *ticker.Symbol, *ticker.LastPrice)
				} else if resp.ArrayOfUmfuturesGetTicker24hrV1RespItem != nil {
					// Array response
					tickers := *resp.ArrayOfUmfuturesGetTicker24hrV1RespItem
					if len(tickers) == 0 {
						t.Fatal("24hr ticker array should not be empty")
					}
					t.Logf("Found %d 24hr tickers", len(tickers))
					
					// Check first ticker structure
					if len(tickers) > 0 {
						ticker := tickers[0]
						if ticker.Symbol == nil || *ticker.Symbol == "" {
							t.Fatal("Ticker symbol should not be empty")
						}
						if ticker.LastPrice == nil || *ticker.LastPrice == "" {
							t.Fatal("Last price should not be empty")
						}
						t.Logf("First ticker: %s, LastPrice=%s", *ticker.Symbol, *ticker.LastPrice)
					}
				} else {
					t.Fatal("No valid response received")
				}
			})
			break
		}
	}
}

// TestPriceTicker tests the price ticker endpoint
func TestPriceTicker(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Price Ticker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetTickerPriceV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetTickerPriceV1")
					t.Fatalf("Error calling GetTickerPriceV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Handle both single and array response
				if resp.UmfuturesGetTickerPriceV1RespItem != nil {
					// Single item response
					item := resp.UmfuturesGetTickerPriceV1RespItem
					if item.Symbol == nil || *item.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					if item.Price == nil || *item.Price == "" {
						t.Fatal("Price should not be empty")
					}
					t.Logf("Price ticker: %s = %s", *item.Symbol, *item.Price)
				} else if resp.ArrayOfUmfuturesGetTickerPriceV1RespItem != nil {
					// Array response
					items := *resp.ArrayOfUmfuturesGetTickerPriceV1RespItem
					if len(items) == 0 {
						t.Fatal("Price ticker array should not be empty")
					}
					t.Logf("Found %d price tickers", len(items))
					
					// Check first item
					if len(items) > 0 {
						item := items[0]
						if item.Symbol == nil || *item.Symbol == "" {
							t.Fatal("Symbol should not be empty")
						}
						if item.Price == nil || *item.Price == "" {
							t.Fatal("Price should not be empty")
						}
						t.Logf("First price ticker: %s = %s", *item.Symbol, *item.Price)
					}
				} else {
					t.Fatal("No valid response received")
				}
			})
			break
		}
	}
}

// TestBookTicker tests the book ticker endpoint
func TestBookTicker(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Book Ticker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetTickerBookTickerV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetTickerBookTickerV1")
					t.Fatalf("Error calling GetTickerBookTickerV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Handle both single and array response
				if resp.UmfuturesGetTickerBookTickerV1RespItem != nil {
					// Single item response
					item := resp.UmfuturesGetTickerBookTickerV1RespItem
					if item.Symbol == nil || *item.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					if item.BidPrice == nil || *item.BidPrice == "" {
						t.Fatal("Bid price should not be empty")
					}
					if item.AskPrice == nil || *item.AskPrice == "" {
						t.Fatal("Ask price should not be empty")
					}
					t.Logf("Book ticker: %s, Bid=%s, Ask=%s", *item.Symbol, *item.BidPrice, *item.AskPrice)
				} else if resp.ArrayOfUmfuturesGetTickerBookTickerV1RespItem != nil {
					// Array response
					items := *resp.ArrayOfUmfuturesGetTickerBookTickerV1RespItem
					if len(items) == 0 {
						t.Fatal("Book ticker array should not be empty")
					}
					t.Logf("Found %d book tickers", len(items))
					
					// Check first item
					if len(items) > 0 {
						item := items[0]
						if item.Symbol == nil || *item.Symbol == "" {
							t.Fatal("Symbol should not be empty")
						}
						if item.BidPrice == nil || *item.BidPrice == "" {
							t.Fatal("Bid price should not be empty")
						}
						if item.AskPrice == nil || *item.AskPrice == "" {
							t.Fatal("Ask price should not be empty")
						}
						t.Logf("First book ticker: %s, Bid=%s, Ask=%s", *item.Symbol, *item.BidPrice, *item.AskPrice)
					}
				} else {
					t.Fatal("No valid response received")
				}
			})
			break
		}
	}
}

// TestOpenInterest tests the open interest endpoint
func TestOpenInterest(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Open Interest", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetOpenInterestV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetOpenInterestV1")
					t.Fatalf("Error calling GetOpenInterestV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.OpenInterest == nil || *resp.OpenInterest == "" {
					t.Fatal("Open interest should not be empty")
				}
				
				t.Logf("Open interest for BTCUSDT: %s", *resp.OpenInterest)
			})
			break
		}
	}
}

// TestPremiumIndex tests the premium index endpoint
func TestPremiumIndex(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Premium Index", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetPremiumIndexV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetPremiumIndexV1")
					t.Fatalf("Error calling GetPremiumIndexV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Handle both single and array response
				if resp.UmfuturesGetPremiumIndexV1RespItem != nil {
					// Single item response
					item := resp.UmfuturesGetPremiumIndexV1RespItem
					if item.Symbol == nil || *item.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					if item.MarkPrice == nil || *item.MarkPrice == "" {
						t.Fatal("Mark price should not be empty")
					}
					t.Logf("Premium index: %s, MarkPrice=%s", *item.Symbol, *item.MarkPrice)
				} else if resp.ArrayOfUmfuturesGetPremiumIndexV1RespItem != nil {
					// Array response
					items := *resp.ArrayOfUmfuturesGetPremiumIndexV1RespItem
					if len(items) == 0 {
						t.Fatal("Premium index array should not be empty")
					}
					t.Logf("Found %d premium index entries", len(items))
					
					// Check first entry structure
					if len(items) > 0 {
						entry := items[0]
						if entry.Symbol == nil || *entry.Symbol == "" {
							t.Fatal("Symbol should not be empty")
						}
						if entry.MarkPrice == nil || *entry.MarkPrice == "" {
							t.Fatal("Mark price should not be empty")
						}
						t.Logf("First premium index: %s, MarkPrice=%s", *entry.Symbol, *entry.MarkPrice)
					}
				} else {
					t.Fatal("No valid response received")
				}
			})
			break
		}
	}
}

// TestFundingRate tests the funding rate endpoint
func TestFundingRate(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Funding Rate", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetFundingRateV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetFundingRateV1")
					t.Fatalf("Error calling GetFundingRateV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Funding rate should not be empty")
				}
				
				t.Logf("Found %d funding rate entries for BTCUSDT", len(resp))
				
				// Check first entry structure
				if len(resp) > 0 {
					entry := resp[0]
					if entry.Symbol == nil || *entry.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					if entry.FundingRate == nil || *entry.FundingRate == "" {
						t.Fatal("Funding rate should not be empty")
					}
					t.Logf("First funding rate: %s, Rate=%s", *entry.Symbol, *entry.FundingRate)
				}
			})
			break
		}
	}
}

// TestFundingInfo tests the funding info endpoint
func TestFundingInfo(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Funding Info", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetFundingInfoV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					// Check if this is the known issue that the endpoint doesn't exist on Binance 
					if httpResp != nil && httpResp.StatusCode == 404 {
						logResponseBody(t, httpResp, "GetFundingInfoV1")
						t.Skip("GetFundingInfoV1 endpoint not supported by Binance API (404 Not Found)")
						return
					}
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetFundingInfoV1")
					t.Fatalf("Error calling GetFundingInfoV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Funding info should not be empty")
				}
				
				t.Logf("Found %d funding info entries", len(resp))
				
				// Check first entry structure
				if len(resp) > 0 {
					entry := resp[0]
					if entry.Symbol == nil || *entry.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					t.Logf("First funding info: %s", *entry.Symbol)
				}
			})
			break
		}
	}
}

// TestIndexInfo tests the index info endpoint
func TestIndexInfo(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Index Info", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetIndexInfoV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetIndexInfoV1")
					t.Fatalf("Error calling GetIndexInfoV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Index info should not be empty")
				}
				
				t.Logf("Found %d index info entries", len(resp))
				
				// Check first entry structure
				if len(resp) > 0 {
					entry := resp[0]
					if entry.Symbol == nil || *entry.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					t.Logf("First index info: %s", *entry.Symbol)
				}
			})
			break
		}
	}
}

// TestConstituents tests the constituents endpoint
func TestConstituents(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Constituents", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetConstituentsV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetConstituentsV1")
					t.Fatalf("Error calling GetConstituentsV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if resp.Symbol == nil || *resp.Symbol == "" {
					t.Fatal("Symbol should not be empty")
				}
				
				t.Logf("Constituents for %s: %+v", *resp.Symbol, resp)
			})
			break
		}
	}
}

// TestAssetIndex tests the asset index endpoint
func TestAssetIndex(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Asset Index", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetAssetIndexV1(ctx)
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetAssetIndexV1")
					t.Fatalf("Error calling GetAssetIndexV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Handle both single and array response
				if resp.UmfuturesGetAssetIndexV1RespItem != nil {
					// Single item response
					item := resp.UmfuturesGetAssetIndexV1RespItem
					if item.Symbol == nil || *item.Symbol == "" {
						t.Fatal("Symbol should not be empty")
					}
					t.Logf("Asset index: %s", *item.Symbol)
				} else if resp.ArrayOfUmfuturesGetAssetIndexV1RespItem != nil {
					// Array response
					items := *resp.ArrayOfUmfuturesGetAssetIndexV1RespItem
					if len(items) == 0 {
						t.Fatal("Asset index array should not be empty")
					}
					t.Logf("Found %d asset index entries", len(items))
					
					// Check first entry structure
					if len(items) > 0 {
						entry := items[0]
						if entry.Symbol == nil || *entry.Symbol == "" {
							t.Fatal("Symbol should not be empty")
						}
						t.Logf("First asset index: %s", *entry.Symbol)
					}
				} else {
					t.Fatal("No valid response received")
				}
			})
			break
		}
	}
}

// TestContinuousKlines tests the continuous klines endpoint
func TestContinuousKlines(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Continuous Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetContinuousKlinesV1(ctx).Pair("BTCUSDT").ContractType("PERPETUAL").Interval("1m")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetContinuousKlinesV1")
					t.Fatalf("Error calling GetContinuousKlinesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Continuous klines should not be empty")
				}
				
				t.Logf("Found %d continuous klines for BTCUSDT perpetual", len(resp))
			})
			break
		}
	}
}

// TestIndexPriceKlines tests the index price klines endpoint
func TestIndexPriceKlines(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Index Price Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetIndexPriceKlinesV1(ctx).Pair("BTCUSDT").Interval("1m")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetIndexPriceKlinesV1")
					t.Fatalf("Error calling GetIndexPriceKlinesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Index price klines should not be empty")
				}
				
				t.Logf("Found %d index price klines for BTCUSDT", len(resp))
			})
			break
		}
	}
}

// TestMarkPriceKlines tests the mark price klines endpoint
func TestMarkPriceKlines(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Mark Price Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetMarkPriceKlinesV1(ctx).Symbol("BTCUSDT").Interval("1m")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetMarkPriceKlinesV1")
					t.Fatalf("Error calling GetMarkPriceKlinesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Mark price klines should not be empty")
				}
				
				t.Logf("Found %d mark price klines for BTCUSDT", len(resp))
			})
			break
		}
	}
}

// TestPremiumIndexKlines tests the premium index klines endpoint
func TestPremiumIndexKlines(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			testEndpoint(t, config, "Premium Index Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetPremiumIndexKlinesV1(ctx).Symbol("BTCUSDT").Interval("1m")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetPremiumIndexKlinesV1")
					t.Fatalf("Error calling GetPremiumIndexKlinesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Premium index klines should not be empty")
				}
				
				t.Logf("Found %d premium index klines for BTCUSDT", len(resp))
			})
			break
		}
	}
}

// TestHistoricalTrades tests the historical trades endpoint
func TestHistoricalTrades(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			testEndpoint(t, config, "Historical Trades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.FuturesAPI.GetHistoricalTradesV1(ctx).Symbol("BTCUSDT")
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					logResponseBody(t, httpResp, "GetHistoricalTradesV1")
					t.Fatalf("Error calling GetHistoricalTradesV1: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				if len(resp) == 0 {
					t.Fatal("Historical trades should not be empty")
				}
				
				t.Logf("Found %d historical trades for BTCUSDT", len(resp))
				
				// Check first trade structure
				if len(resp) > 0 {
					trade := resp[0]
					if trade.Id == nil {
						t.Fatal("Trade ID should not be nil")
					}
					if trade.Price == nil || *trade.Price == "" {
						t.Fatal("Trade price should not be empty")
					}
					t.Logf("First historical trade: ID=%d, Price=%s", *trade.Id, *trade.Price)
				}
			})
			break
		}
	}
}

// logResponseBody logs the HTTP response body for debugging failed API calls
func logResponseBody(t *testing.T, httpResp *http.Response, endpoint string) {
	if httpResp != nil && httpResp.Body != nil {
		// Read the response body
		bodyBytes, err := io.ReadAll(httpResp.Body)
		if err != nil {
			t.Logf("Failed to read response body for %s: %v", endpoint, err)
			return
		}
		
		// Log the response details
		t.Logf("=== %s Response Details ===", endpoint)
		t.Logf("Status Code: %d %s", httpResp.StatusCode, httpResp.Status)
		t.Logf("Response Headers: %v", httpResp.Header)
		
		if len(bodyBytes) > 0 {
			t.Logf("Response Body: %s", string(bodyBytes))
		} else {
			t.Logf("Response Body: (empty)")
		}
		t.Logf("=== End %s Response ===", endpoint)
		
		// Try to restore the body for potential further processing
		// Note: This creates a new ReadCloser from the read bytes
		httpResp.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	} else {
		t.Logf("No HTTP response available for %s", endpoint)
	}
}

// Helper function to get absolute value
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}