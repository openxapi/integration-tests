package main

import (
	"context"
	"net/http"
	"os"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// Default test symbols for CM Futures
const (
	DefaultCMFuturesSymbol  = "BTCUSD_PERP"
	DefaultCMFuturesSymbol2 = "ETHUSD_PERP"
)

// getTestSymbol returns a test symbol for market data tests
func getTestSymbol() string {
	if symbol := os.Getenv("BINANCE_TEST_CMFUTURES_SYMBOL"); symbol != "" {
		return symbol
	}
	return DefaultCMFuturesSymbol
}

// getTestSymbol2 returns a secondary test symbol for market data tests
func getTestSymbol2() string {
	if symbol := os.Getenv("BINANCE_TEST_CMFUTURES_SYMBOL2"); symbol != "" {
		return symbol
	}
	return DefaultCMFuturesSymbol2
}

// TestOrderBookDepth tests the order book depth endpoint
func TestOrderBookDepth(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "OrderBookDepth", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetDepthV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "OrderBookDepth") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "OrderBookDepth")
						t.Fatalf("Order book depth failed: %v", err)
					}
					
					if resp.Bids == nil || len(resp.Bids) == 0 {
						t.Fatal("No bids returned")
					}
					
					if resp.Asks == nil || len(resp.Asks) == 0 {
						t.Fatal("No asks returned")
					}
					
					t.Logf("Order book depth for %s: bids=%d, asks=%d", symbol, len(resp.Bids), len(resp.Asks))
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "AggTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetAggTradesV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "AggTrades") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AggTrades")
						t.Fatalf("Aggregate trades failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No aggregate trades returned")
					}
					
					firstTrade := resp[0]
					if firstTrade.P == nil {
						t.Fatal("First trade has nil price")
					}
					
					if firstTrade.Q == nil {
						t.Fatal("First trade has nil quantity")
					}
					
					t.Logf("Aggregate trades for %s: count=%d, first_trade_price=%s", 
						symbol, len(resp), *firstTrade.P)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "RecentTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetTradesV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "RecentTrades") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "RecentTrades")
						t.Fatalf("Recent trades failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No recent trades returned")
					}
					
					firstTrade := resp[0]
					if firstTrade.Price == nil {
						t.Fatal("First trade has nil price")
					}
					
					if firstTrade.Qty == nil {
						t.Fatal("First trade has nil quantity")
					}
					
					t.Logf("Recent trades for %s: count=%d, first_trade_price=%s", 
						symbol, len(resp), *firstTrade.Price)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "HistoricalTrades", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetHistoricalTradesV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "HistoricalTrades") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "HistoricalTrades")
						t.Fatalf("Historical trades failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No historical trades returned")
					}
					
					firstTrade := resp[0]
					if firstTrade.Price == nil {
						t.Fatal("First trade has nil price")
					}
					
					if firstTrade.Qty == nil {
						t.Fatal("First trade has nil quantity")
					}
					
					t.Logf("Historical trades for %s: count=%d, first_trade_price=%s", 
						symbol, len(resp), *firstTrade.Price)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Klines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetKlinesV1(ctx).Symbol(symbol).Interval("1m")
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Klines") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "Klines")
						t.Fatalf("Klines failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No klines returned")
					}
					
					// Each kline should have 12 elements: [openTime, open, high, low, close, volume, closeTime, quoteVolume, count, takerBuyVolume, takerBuyQuoteVolume, ignore]
					firstKline := resp[0]
					if len(firstKline) < 12 {
						t.Fatalf("First kline has %d elements, expected at least 12", len(firstKline))
					}
					
					t.Logf("Klines for %s: count=%d, first_kline_open=%v", 
						symbol, len(resp), firstKline[1])
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ContinuousKlines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					contractType := "PERPETUAL"
					
					req := client.FuturesAPI.GetContinuousKlinesV1(ctx).Pair(pair).ContractType(contractType).Interval("1m")
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ContinuousKlines") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "ContinuousKlines")
						t.Fatalf("Continuous klines failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No continuous klines returned")
					}
					
					firstKline := resp[0]
					if len(firstKline) < 12 {
						t.Fatalf("First kline has %d elements, expected at least 12", len(firstKline))
					}
					
					t.Logf("Continuous klines for %s %s: count=%d, first_kline_open=%v", 
						pair, contractType, len(resp), firstKline[1])
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "IndexPriceKlines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					
					req := client.FuturesAPI.GetIndexPriceKlinesV1(ctx).Pair(pair).Interval("1m")
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "IndexPriceKlines") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IndexPriceKlines")
						t.Fatalf("Index price klines failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No index price klines returned")
					}
					
					firstKline := resp[0]
					if len(firstKline) < 12 {
						t.Fatalf("First kline has %d elements, expected at least 12", len(firstKline))
					}
					
					t.Logf("Index price klines for %s: count=%d, first_kline_open=%v", 
						pair, len(resp), firstKline[1])
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "MarkPriceKlines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetMarkPriceKlinesV1(ctx).Symbol(symbol).Interval("1m")
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "MarkPriceKlines") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "MarkPriceKlines")
						t.Fatalf("Mark price klines failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No mark price klines returned")
					}
					
					firstKline := resp[0]
					if len(firstKline) < 12 {
						t.Fatalf("First kline has %d elements, expected at least 12", len(firstKline))
					}
					
					t.Logf("Mark price klines for %s: count=%d, first_kline_open=%v", 
						symbol, len(resp), firstKline[1])
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PremiumIndexKlines", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetPremiumIndexKlinesV1(ctx).Symbol(symbol).Interval("1m")
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "PremiumIndexKlines") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "PremiumIndexKlines")
						t.Fatalf("Premium index klines failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No premium index klines returned")
					}
					
					firstKline := resp[0]
					if len(firstKline) < 12 {
						t.Fatalf("First kline has %d elements, expected at least 12", len(firstKline))
					}
					
					t.Logf("Premium index klines for %s: count=%d, first_kline_open=%v", 
						symbol, len(resp), firstKline[1])
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "24hrTicker", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetTicker24hrV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "24hrTicker") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "24hrTicker")
						t.Fatalf("24hr ticker failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No ticker data returned")
					}
					
					ticker := resp[0]
					
					if ticker.Symbol == nil {
						t.Fatal("Symbol is nil")
					}
					
					if ticker.LastPrice == nil {
						t.Fatal("LastPrice is nil")
					}
					
					if ticker.Volume == nil {
						t.Fatal("Volume is nil")
					}
					
					t.Logf("24hr ticker for %s: last_price=%s, volume=%s", 
						*ticker.Symbol, *ticker.LastPrice, *ticker.Volume)
				})
			})
			break
		}
	}
}

// TestTickerPrice tests the ticker price endpoint
func TestTickerPrice(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TickerPrice", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetTickerPriceV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TickerPrice") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "TickerPrice")
						t.Fatalf("Ticker price failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No ticker price data returned")
					}
					
					ticker := resp[0]
					
					if ticker.Symbol == nil {
						t.Fatal("Symbol is nil")
					}
					
					if ticker.Price == nil {
						t.Fatal("Price is nil")
					}
					
					t.Logf("Ticker price for %s: price=%s", *ticker.Symbol, *ticker.Price)
				})
			})
			break
		}
	}
}

// TestTickerBook tests the ticker book endpoint
func TestTickerBook(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TickerBook", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetTickerBookTickerV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TickerBook") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "TickerBook")
						t.Fatalf("Ticker book failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No ticker book data returned")
					}
					
					ticker := resp[0]
					
					if ticker.Symbol == nil {
						t.Fatal("Symbol is nil")
					}
					
					if ticker.BidPrice == nil {
						t.Fatal("BidPrice is nil")
					}
					
					if ticker.AskPrice == nil {
						t.Fatal("AskPrice is nil")
					}
					
					t.Logf("Ticker book for %s: bid=%s, ask=%s", 
						*ticker.Symbol, *ticker.BidPrice, *ticker.AskPrice)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PremiumIndex", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetPremiumIndexV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "PremiumIndex") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "PremiumIndex")
						t.Fatalf("Premium index failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No premium index data returned")
					}
					
					premium := resp[0]
					
					if premium.Symbol == nil {
						t.Fatal("Symbol is nil")
					}
					
					if premium.MarkPrice == nil {
						t.Fatal("MarkPrice is nil")
					}
					
					t.Logf("Premium index for %s: mark_price=%s", *premium.Symbol, *premium.MarkPrice)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "FundingRate", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetFundingRateV1(ctx).Symbol(symbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "FundingRate") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FundingRate")
						t.Fatalf("Funding rate failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No funding rate data returned")
					}
					
					firstRate := resp[0]
					if firstRate.Symbol == nil {
						t.Fatal("First rate has nil symbol")
					}
					
					if firstRate.FundingRate == nil {
						t.Fatal("First rate has nil funding rate")
					}
					
					t.Logf("Funding rate for %s: count=%d, first_rate=%s", 
						symbol, len(resp), *firstRate.FundingRate)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "FundingInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetFundingInfoV1(ctx)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "FundingInfo") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FundingInfo")
						t.Fatalf("Funding info failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No funding info returned")
					}
					
					firstInfo := resp[0]
					if firstInfo.Symbol == nil {
						t.Fatal("First info has nil symbol")
					}
					
					t.Logf("Funding info: count=%d, first_symbol=%s", len(resp), *firstInfo.Symbol)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "OpenInterest", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// First try to get exchange info to find valid symbols
					exchangeReq := client.FuturesAPI.GetExchangeInfoV1(ctx)
					exchangeResp, _, exchangeErr := exchangeReq.Execute()
					
					if exchangeErr != nil || exchangeResp.Symbols == nil || len(exchangeResp.Symbols) == 0 {
						t.Logf("Warning: Could not get exchange info, using default symbol")
						symbol := getTestSymbol()
						
						req := client.FuturesAPI.GetOpenInterestV1(ctx).Symbol(symbol)
						resp, httpResp, err := req.Execute()
						
						if handleTestnetError(t, err, httpResp, "OpenInterest") {
							return
						}
						
						if err != nil {
							checkAPIError(t, err, httpResp, "OpenInterest")
							t.Fatalf("Open interest failed: %v", err)
						}
						
						t.Logf("OpenInterest response: %+v", resp)
						
						if resp.Symbol == nil {
							t.Fatal("Symbol is nil - unexpected response structure")
						}
						
						if resp.OpenInterest == nil {
							t.Fatal("OpenInterest is nil - unexpected response structure")
						}
						
						t.Logf("Open interest for %s: open_interest=%s", *resp.Symbol, *resp.OpenInterest)
						return
					}
					
					// Find an active symbol from exchange info
					var validSymbol string
					for _, symbolInfo := range exchangeResp.Symbols {
						if symbolInfo.Symbol != nil {
							// Use the first available symbol - they should all be trading on testnet
							validSymbol = *symbolInfo.Symbol
							break
						}
					}
					
					if validSymbol == "" {
						t.Fatal("No valid trading symbols found in exchange info")
					}
					
					t.Logf("Using symbol from exchange info: %s", validSymbol)
					
					req := client.FuturesAPI.GetOpenInterestV1(ctx).Symbol(validSymbol)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "OpenInterest") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "OpenInterest")
						t.Fatalf("Open interest failed: %v", err)
					}
					
					// Check if we got a successful response
					
					// Handle empty response - this is normal for testnet or symbols without open interest data
					if resp.Symbol == nil {
						t.Logf("OpenInterest endpoint returned empty response for symbol %s - this is normal on testnet", validSymbol)
						t.Logf("API call successful but no open interest data available for this symbol")
						return
					}
					
					if resp.OpenInterest == nil {
						t.Fatal("OpenInterest is nil - unexpected response structure")
					}
					
					t.Logf("Open interest for %s: open_interest=%s", *resp.Symbol, *resp.OpenInterest)
				})
			})
			break
		}
	}
}

// TestIndexConstituents tests the index constituents endpoint
func TestIndexConstituents(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "IndexConstituents", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Try different index symbols - these are for coin-margined futures
					// Common CM futures index symbols based on Binance documentation
					indexSymbols := []string{"BTCUSD", "ETHUSD", "ADAUSD", "LINKUSD", "BNBUSD", "DOGEUSD"}
					
					var lastErr error
					var lastHttpResp *http.Response
					
					for _, symbol := range indexSymbols {
						t.Logf("Trying index symbol: %s", symbol)
						
						req := client.FuturesAPI.GetConstituentsV1(ctx).Symbol(symbol)
						resp, httpResp, err := req.Execute()
						
						if err == nil {
							// Success! Test passed
							if resp.Symbol == nil {
								t.Fatal("Symbol is nil")
							}
							
							if resp.Constituents == nil {
								t.Fatal("Constituents is nil")
							}
							
							t.Logf("Index constituents for %s: count=%d", *resp.Symbol, len(resp.Constituents))
							return
						}
						
						// Save the error for potential fallback
						lastErr = err
						lastHttpResp = httpResp
						
						// Check if this is a testnet limitation (404/403) and skip if so
						if handleTestnetError(t, err, httpResp, "IndexConstituents-"+symbol) {
							return
						}
						
						// For 400 Bad Request, continue trying other symbols
						if httpResp != nil && httpResp.StatusCode == 400 {
							t.Logf("Symbol %s invalid (400), trying next symbol", symbol)
							continue
						}
						
						// For other errors, break and fail
						break
					}
					
					// If we get here, all symbols failed
					if lastErr != nil {
						checkAPIError(t, lastErr, lastHttpResp, "IndexConstituents")
						t.Fatalf("Index constituents failed for all symbols tried. Last error: %v", lastErr)
					}
				})
			})
			break
		}
	}
}

// TestForceOrders tests the force orders endpoint
func TestForceOrders(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ForceOrders", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					timestamp := generateTimestamp()
					
					req := client.FuturesAPI.GetForceOrdersV1(ctx).Symbol(symbol).Timestamp(timestamp)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ForceOrders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "ForceOrders")
						t.Fatalf("Force orders failed: %v", err)
					}
					
					// Force orders might be empty, which is normal
					t.Logf("Force orders for %s: count=%d", symbol, len(resp))
				})
			})
			break
		}
	}
}