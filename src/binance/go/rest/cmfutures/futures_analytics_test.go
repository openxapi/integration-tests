package main

import (
	"context"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// TestFuturesDataBasis tests querying basis
func TestFuturesDataBasis(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "FuturesDataBasis", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					contractType := "PERPETUAL"
					
					req := client.FuturesAPI.GetFuturesDataBasis(ctx).
						Pair(pair).
						ContractType(contractType).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "FuturesDataBasis") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Futures data basis failed: %v", err)
					}
					
					t.Logf("Futures data basis for %s %s: count=%d", pair, contractType, len(resp))
					
					// Check structure of first basis entry if any exist
					if len(resp) > 0 {
						firstBasis := resp[0]
						if firstBasis.Pair == nil {
							t.Fatal("First basis has nil Pair")
						}
						
						if firstBasis.ContractType == nil {
							t.Fatal("First basis has nil ContractType")
						}
						
						if firstBasis.FuturesPrice == nil {
							t.Fatal("First basis has nil FuturesPrice")
						}
						
						if firstBasis.IndexPrice == nil {
							t.Fatal("First basis has nil IndexPrice")
						}
						
						if firstBasis.Basis == nil {
							t.Fatal("First basis has nil Basis")
						}
						
						if firstBasis.BasisRate == nil {
							t.Fatal("First basis has nil BasisRate")
						}
						
						if firstBasis.Timestamp == nil {
							t.Fatal("First basis has nil Timestamp")
						}
						
						t.Logf("First basis: pair=%s, contractType=%s, futuresPrice=%s, indexPrice=%s, basis=%s, basisRate=%s, timestamp=%d", 
							*firstBasis.Pair, *firstBasis.ContractType, *firstBasis.FuturesPrice, *firstBasis.IndexPrice, 
							*firstBasis.Basis, *firstBasis.BasisRate, *firstBasis.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestGlobalLongShortRatio tests querying symbol Long/Short Ratio
func TestGlobalLongShortRatio(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "GlobalLongShortRatio", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					
					req := client.FuturesAPI.GetFuturesDataGlobalLongShortAccountRatio(ctx).
						Pair(pair).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "GlobalLongShortRatio") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Global long short ratio failed: %v", err)
					}
					
					t.Logf("Global long short ratio for %s: count=%d", pair, len(resp))
					
					// Check structure of first ratio entry if any exist
					if len(resp) > 0 {
						firstRatio := resp[0]
						if firstRatio.Pair == nil {
							t.Fatal("First ratio has nil Pair")
						}
						
						if firstRatio.LongShortRatio == nil {
							t.Fatal("First ratio has nil LongShortRatio")
						}
						
						if firstRatio.LongAccount == nil {
							t.Fatal("First ratio has nil LongAccount")
						}
						
						if firstRatio.ShortAccount == nil {
							t.Fatal("First ratio has nil ShortAccount")
						}
						
						if firstRatio.Timestamp == nil {
							t.Fatal("First ratio has nil Timestamp")
						}
						
						t.Logf("First ratio: pair=%s, longShortRatio=%s, longAccount=%s, shortAccount=%s, timestamp=%d", 
							*firstRatio.Pair, *firstRatio.LongShortRatio, *firstRatio.LongAccount, 
							*firstRatio.ShortAccount, *firstRatio.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestOpenInterestHistory tests querying open interest stats
func TestOpenInterestHistory(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "OpenInterestHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					contractType := "PERPETUAL"
					
					req := client.FuturesAPI.GetFuturesDataOpenInterestHist(ctx).
						Pair(pair).
						ContractType(contractType).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "OpenInterestHistory") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Open interest history failed: %v", err)
					}
					
					t.Logf("Open interest history for %s %s: count=%d", pair, contractType, len(resp))
					
					// Check structure of first open interest entry if any exist
					if len(resp) > 0 {
						firstOI := resp[0]
						if firstOI.Pair == nil {
							t.Fatal("First open interest has nil Pair")
						}
						
						if firstOI.ContractType == nil {
							t.Fatal("First open interest has nil ContractType")
						}
						
						if firstOI.SumOpenInterest == nil {
							t.Fatal("First open interest has nil SumOpenInterest")
						}
						
						if firstOI.SumOpenInterestValue == nil {
							t.Fatal("First open interest has nil SumOpenInterestValue")
						}
						
						if firstOI.Timestamp == nil {
							t.Fatal("First open interest has nil Timestamp")
						}
						
						t.Logf("First open interest: pair=%s, contractType=%s, sumOpenInterest=%s, sumOpenInterestValue=%s, timestamp=%d", 
							*firstOI.Pair, *firstOI.ContractType, *firstOI.SumOpenInterest, 
							*firstOI.SumOpenInterestValue, *firstOI.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestTakerBuySellVolume tests querying taker buy/sell volume
func TestTakerBuySellVolume(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TakerBuySellVolume", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					contractType := "PERPETUAL"
					
					req := client.FuturesAPI.GetFuturesDataTakerBuySellVol(ctx).
						Pair(pair).
						ContractType(contractType).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TakerBuySellVolume") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Taker buy sell volume failed: %v", err)
					}
					
					t.Logf("Taker buy sell volume for %s %s: count=%d", pair, contractType, len(resp))
					
					// Check structure of first volume entry if any exist
					if len(resp) > 0 {
						firstVolume := resp[0]
						if firstVolume.Pair == nil {
							t.Fatal("First volume has nil Pair")
						}
						
						if firstVolume.ContractType == nil {
							t.Fatal("First volume has nil ContractType")
						}
						
						if firstVolume.TakerBuyVol == nil {
							t.Fatal("First volume has nil TakerBuyVol")
						}
						
						if firstVolume.TakerSellVol == nil {
							t.Fatal("First volume has nil TakerSellVol")
						}
						
						if firstVolume.TakerBuyVolValue == nil {
							t.Fatal("First volume has nil TakerBuyVolValue")
						}
						
						if firstVolume.TakerSellVolValue == nil {
							t.Fatal("First volume has nil TakerSellVolValue")
						}
						
						if firstVolume.Timestamp == nil {
							t.Fatal("First volume has nil Timestamp")
						}
						
						t.Logf("First volume: pair=%s, contractType=%s, takerBuyVol=%s, takerSellVol=%s, takerBuyVolValue=%s, takerSellVolValue=%s, timestamp=%d", 
							*firstVolume.Pair, *firstVolume.ContractType, *firstVolume.TakerBuyVol, 
							*firstVolume.TakerSellVol, *firstVolume.TakerBuyVolValue, 
							*firstVolume.TakerSellVolValue, *firstVolume.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestTopTraderLongShortAccountRatio tests querying top trader Long/Short Account Ratio
func TestTopTraderLongShortAccountRatio(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TopTraderLongShortAccountRatio", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					
					req := client.FuturesAPI.GetFuturesDataTopLongShortAccountRatio(ctx).
						Symbol(pair).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TopTraderLongShortAccountRatio") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Top trader long short account ratio failed: %v", err)
					}
					
					t.Logf("Top trader long short account ratio for %s: count=%d", pair, len(resp))
					
					// Check structure of first ratio entry if any exist
					if len(resp) > 0 {
						firstRatio := resp[0]
						if firstRatio.Pair == nil {
							t.Fatal("First ratio has nil Pair")
						}
						
						if firstRatio.LongShortRatio == nil {
							t.Fatal("First ratio has nil LongShortRatio")
						}
						
						if firstRatio.LongAccount == nil {
							t.Fatal("First ratio has nil LongAccount")
						}
						
						if firstRatio.ShortAccount == nil {
							t.Fatal("First ratio has nil ShortAccount")
						}
						
						if firstRatio.Timestamp == nil {
							t.Fatal("First ratio has nil Timestamp")
						}
						
						t.Logf("First ratio: pair=%s, longShortRatio=%s, longAccount=%s, shortAccount=%s, timestamp=%d", 
							*firstRatio.Pair, *firstRatio.LongShortRatio, *firstRatio.LongAccount, 
							*firstRatio.ShortAccount, *firstRatio.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestTopTraderLongShortPositionRatio tests querying top trader Long/Short Position Ratio
func TestTopTraderLongShortPositionRatio(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TopTraderLongShortPositionRatio", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					
					req := client.FuturesAPI.GetFuturesDataTopLongShortPositionRatio(ctx).
						Pair(pair).
						Period("5m")
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TopTraderLongShortPositionRatio") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "FuturesAnalyticsOperation")
						t.Fatalf("Top trader long short position ratio failed: %v", err)
					}
					
					t.Logf("Top trader long short position ratio for %s: count=%d", pair, len(resp))
					
					// Check structure of first ratio entry if any exist
					if len(resp) > 0 {
						firstRatio := resp[0]
						if firstRatio.Pair == nil {
							t.Fatal("First ratio has nil Pair")
						}
						
						if firstRatio.LongShortRatio == nil {
							t.Fatal("First ratio has nil LongShortRatio")
						}
						
						if firstRatio.LongPosition == nil {
							t.Fatal("First ratio has nil LongPosition")
						}
						
						if firstRatio.ShortPosition == nil {
							t.Fatal("First ratio has nil ShortPosition")
						}
						
						if firstRatio.Timestamp == nil {
							t.Fatal("First ratio has nil Timestamp")
						}
						
						t.Logf("First ratio: pair=%s, longShortRatio=%s, longPosition=%s, shortPosition=%s, timestamp=%d", 
							*firstRatio.Pair, *firstRatio.LongShortRatio, *firstRatio.LongPosition, 
							*firstRatio.ShortPosition, *firstRatio.Timestamp)
					}
				})
			})
			break
		}
	}
}

// TestFuturesAnalyticsComprehensive tests multiple futures analytics endpoints together
func TestFuturesAnalyticsComprehensive(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "FuturesAnalyticsComprehensive", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					pair := "BTCUSD"
					contractType := "PERPETUAL"
					period := "5m"
					
					// Test all analytics endpoints with the same parameters
					endpoints := []struct {
						name string
						test func() error
					}{
						{
							name: "Basis",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataBasis(ctx).
									Pair(pair).
									ContractType(contractType).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
						{
							name: "Global Long Short Account Ratio",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataGlobalLongShortAccountRatio(ctx).
									Pair(pair).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
						{
							name: "Open Interest History",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataOpenInterestHist(ctx).
									Pair(pair).
									ContractType(contractType).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
						{
							name: "Taker Buy Sell Volume",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataTakerBuySellVol(ctx).
									Pair(pair).
									ContractType(contractType).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
						{
							name: "Top Trader Long Short Account Ratio",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataTopLongShortAccountRatio(ctx).
									Symbol(pair).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
						{
							name: "Top Trader Long Short Position Ratio",
							test: func() error {
								req := client.FuturesAPI.GetFuturesDataTopLongShortPositionRatio(ctx).
									Pair(pair).
									Period(period)
								_, _, err := req.Execute()
								return err
							},
						},
					}
					
					var successCount int
					var failureCount int
					
					for _, endpoint := range endpoints {
						err := endpoint.test()
						if err != nil {
							t.Logf("❌ %s failed: %v", endpoint.name, err)
							failureCount++
						} else {
							t.Logf("✅ %s succeeded", endpoint.name)
							successCount++
						}
						
						// Small delay between requests to respect rate limits
						time.Sleep(100 * time.Millisecond)
					}
					
					t.Logf("Comprehensive analytics test completed: %d succeeded, %d failed", successCount, failureCount)
					
					// The test passes if at least half of the endpoints work
					if successCount >= len(endpoints)/2 {
						t.Logf("Comprehensive test passed (%d/%d endpoints working)", successCount, len(endpoints))
					} else {
						t.Errorf("Comprehensive test failed - too many endpoints failed (%d/%d working)", successCount, len(endpoints))
					}
				})
			})
			break
		}
	}
}