package main

import (
	"context"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// TestPing tests the ping endpoint
func TestPing(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Ping", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetPingV1(ctx)
					_, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Ping") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "Ping")
						t.Fatalf("Ping failed: %v", err)
					}
					
					t.Logf("Ping successful")
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ServerTime", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetTimeV1(ctx)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ServerTime") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "ServerTime")
						t.Fatalf("Server time failed: %v", err)
					}
					
					if resp.ServerTime == nil {
						t.Fatal("Server time response is nil")
					}
					
					serverTime := *resp.ServerTime
					currentTime := time.Now().UnixMilli()
					
					// Check if server time is reasonable (within 5 minutes of current time)
					timeDiff := currentTime - serverTime
					if timeDiff < 0 {
						timeDiff = -timeDiff
					}
					
					if timeDiff > 5*60*1000 { // 5 minutes in milliseconds
						t.Errorf("Server time seems incorrect. Server: %d, Local: %d, Diff: %d ms", 
							serverTime, currentTime, timeDiff)
					}
					
					t.Logf("Server time: %d (diff: %d ms)", serverTime, timeDiff)
				})
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
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ExchangeInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetExchangeInfoV1(ctx)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ExchangeInfo") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "ExchangeInfo")
						t.Fatalf("Exchange info failed: %v", err)
					}
					
					if resp.Timezone == nil {
						t.Fatal("Timezone is nil")
					}
					
					if resp.ServerTime == nil {
						t.Fatal("ServerTime is nil")
					}
					
					if resp.Symbols == nil {
						t.Fatal("Symbols is nil")
					}
					
					if len(resp.Symbols) == 0 {
						t.Fatal("No symbols returned")
					}
					
					// Check first symbol has required fields
					firstSymbol := resp.Symbols[0]
					if firstSymbol.Symbol == nil {
						t.Fatal("First symbol has nil symbol")
					}
					
					if firstSymbol.ContractStatus == nil {
						t.Fatal("First symbol has nil contract status")
					}
					
					if firstSymbol.BaseAsset == nil {
						t.Fatal("First symbol has nil base asset")
					}
					
					if firstSymbol.QuoteAsset == nil {
						t.Fatal("First symbol has nil quote asset")
					}
					
					t.Logf("Exchange info: timezone=%s, symbols=%d, first_symbol=%s", 
						*resp.Timezone, len(resp.Symbols), *firstSymbol.Symbol)
				})
			})
			break
		}
	}
}