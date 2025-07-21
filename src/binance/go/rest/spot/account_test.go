package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestAccountInfo tests the account information endpoint
func TestAccountInfo(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetAccountV3(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					t.Fatalf("Failed to get account info: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.MakerCommission == nil {
					t.Error("Expected maker commission in response")
				}
				
				if resp.TakerCommission == nil {
					t.Error("Expected taker commission in response")
				}
				
				if resp.BuyerCommission == nil {
					t.Error("Expected buyer commission in response")
				}
				
				if resp.SellerCommission == nil {
					t.Error("Expected seller commission in response")
				}
				
				if resp.CanTrade == nil {
					t.Error("Expected canTrade flag in response")
				}
				
				if resp.CanWithdraw == nil {
					t.Error("Expected canWithdraw flag in response")
				}
				
				if resp.CanDeposit == nil {
					t.Error("Expected canDeposit flag in response")
				}
				
				if resp.UpdateTime == nil {
					t.Error("Expected updateTime in response")
				}
				
				if resp.AccountType == nil || *resp.AccountType == "" {
					t.Error("Expected account type in response")
				}
				
				if resp.Balances == nil {
					t.Error("Expected balances in response")
				}
				
				// Check if we have any balances
				hasBalance := false
				for _, balance := range resp.Balances {
					if balance.Asset != nil && balance.Free != nil && balance.Locked != nil {
						hasBalance = true
						t.Logf("Asset: %s, Free: %s, Locked: %s", 
							*balance.Asset, *balance.Free, *balance.Locked)
						
						// Check USDT balance specifically (common test asset)
						if *balance.Asset == "USDT" {
							t.Logf("USDT Balance - Free: %s, Locked: %s", 
								*balance.Free, *balance.Locked)
						}
					}
				}
				
				if !hasBalance {
					t.Log("Warning: No balances found in account")
				}
				
				// Log permissions
				t.Logf("Account permissions - Trade: %v, Withdraw: %v, Deposit: %v",
					*resp.CanTrade, *resp.CanWithdraw, *resp.CanDeposit)
			})
		})
	}
}

// TestAccountCommission tests the account commission endpoint
func TestAccountCommission(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountCommission", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.SpotTradingAPI.GetAccountCommissionV3(ctx).
					Symbol("BTCUSDT").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					// Never skip 400 - these are bad requests that need fixing
					// This will log response body for 400 errors automatically
					checkAPIErrorWithResponse(t, err, httpResp, "Get account commission")
					t.Fatalf("Failed to get account commission: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Symbol == nil || *resp.Symbol != "BTCUSDT" {
					t.Errorf("Expected symbol BTCUSDT, got %v", resp.Symbol)
				}
				
				if resp.StandardCommission == nil {
					t.Error("Expected standard commission in response")
				} else {
					if resp.StandardCommission.Maker == nil || *resp.StandardCommission.Maker == "" {
						t.Error("Expected maker commission rate")
					}
					if resp.StandardCommission.Taker == nil || *resp.StandardCommission.Taker == "" {
						t.Error("Expected taker commission rate")
					}
					
					t.Logf("Standard commission - Maker: %s, Taker: %s",
						*resp.StandardCommission.Maker, *resp.StandardCommission.Taker)
				}
				
				if resp.TaxCommission == nil {
					t.Error("Expected tax commission in response")
				} else {
					if resp.TaxCommission.Maker == nil || *resp.TaxCommission.Maker == "" {
						t.Error("Expected tax maker commission rate")
					}
					if resp.TaxCommission.Taker == nil || *resp.TaxCommission.Taker == "" {
						t.Error("Expected tax taker commission rate")
					}
					
					t.Logf("Tax commission - Maker: %s, Taker: %s",
						*resp.TaxCommission.Maker, *resp.TaxCommission.Taker)
				}
				
				if resp.Discount == nil {
					t.Error("Expected discount information in response")
				} else {
					if resp.Discount.EnabledForAccount != nil {
						t.Logf("Discount enabled for account: %v", *resp.Discount.EnabledForAccount)
					}
					if resp.Discount.EnabledForSymbol != nil {
						t.Logf("Discount enabled for symbol: %v", *resp.Discount.EnabledForSymbol)
					}
					if resp.Discount.DiscountAsset != nil && *resp.Discount.DiscountAsset != "" {
						t.Logf("Discount asset: %s", *resp.Discount.DiscountAsset)
					}
				}
			})
		})
	}
}

// TestTradeFee tests the trade fee endpoint from wallet API
func TestTradeFee(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "TradeFee", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAssetTradeFeeV1(ctx).
					Symbol("BTCUSDT").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Trade fee endpoint not available on testnet")
					}
					t.Fatalf("Failed to get trade fee: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected trade fee data in response")
				}
				
				// Check the first fee entry
				if len(resp) > 0 {
					fee := resp[0]
					if fee.Symbol == nil || *fee.Symbol != "BTCUSDT" {
						t.Errorf("Expected symbol BTCUSDT, got %v", fee.Symbol)
					}
					
					if fee.MakerCommission == nil || *fee.MakerCommission == "" {
						t.Error("Expected maker commission")
					}
					
					if fee.TakerCommission == nil || *fee.TakerCommission == "" {
						t.Error("Expected taker commission")
					}
					
					t.Logf("Trade fee for BTCUSDT - Maker: %s, Taker: %s",
						*fee.MakerCommission, *fee.TakerCommission)
				}
			})
		})
	}
}

// TestAPIKeyPermissions tests the API key permissions endpoint
// NOTE: This endpoint might not be available in the SDK
/*
func TestAPIKeyPermissions(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "APIKeyPermissions", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// This endpoint may not exist in the current SDK version
				t.Skip("API key permissions endpoint not available in SDK")
			})
		})
	}
}
*/

// TestAccountStatus tests the account status endpoint
func TestAccountStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "AccountStatus", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAccountStatusV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Account status endpoint not available on testnet")
					}
					t.Fatalf("Failed to get account status: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Data == nil || *resp.Data == "" {
					t.Error("Expected account status data")
				} else {
					t.Logf("Account status: %s", *resp.Data)
				}
			})
		})
	}
}