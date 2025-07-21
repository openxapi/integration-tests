package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestAccountInfo tests querying portfolio margin account information
func TestAccountInfo(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Account Info", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.GetAccountV1(ctx).
						Timestamp(generateTimestamp())
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Account Info") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Account Info") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetAccountV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Account info retrieved successfully")
					if resp.UniMMR != nil {
						t.Logf("UniMMR: %s", *resp.UniMMR)
					}
					if resp.UpdateTime != nil {
						t.Logf("UpdateTime: %d", *resp.UpdateTime)
					}
				})
			})
		}
	}
}

// TestAccountBalance tests querying portfolio margin account balance
func TestAccountBalance(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Account Balance", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.GetBalanceV1(ctx).
						Timestamp(generateTimestamp())
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Account Balance") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Account Balance") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetBalanceV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Account balance retrieved successfully")
					
					// Handle the oneOf response type - it can be either a single item or an array
					actualInstance := resp.GetActualInstance()
					if actualInstance != nil {
						switch v := actualInstance.(type) {
						case *openapi.PmarginGetBalanceV1RespItem:
							// Single balance item
							if v.Asset != nil && v.TotalWalletBalance != nil {
								t.Logf("Single Balance - Asset: %s, Total Wallet Balance: %s", *v.Asset, *v.TotalWalletBalance)
							}
						case *[]openapi.PmarginGetBalanceV1RespItem:
							// Array of balance items
							balances := *v
							if len(balances) > 0 {
								for i, balance := range balances {
									if i >= 5 { // Limit output
										t.Logf("... and %d more balances", len(balances)-i)
										break
									}
									if balance.Asset != nil && balance.TotalWalletBalance != nil {
										t.Logf("Asset: %s, Total Wallet Balance: %s", *balance.Asset, *balance.TotalWalletBalance)
									}
								}
							}
						default:
							t.Logf("Unknown balance response type: %T", actualInstance)
						}
					} else {
						t.Logf("No balance data available")
					}
				})
			})
		}
	}
}