package main

import (
	"context"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestGetSystemStatus tests the system status endpoint
func TestGetSystemStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeNONE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetSystemStatus", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetSystemStatusV1(ctx)
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("System status endpoint not available on testnet")
					}
					t.Fatalf("Failed to get system status: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Status == nil {
					t.Error("Expected status in response")
				} else {
					t.Logf("System status: %d", *resp.Status)
					if resp.Msg != nil {
						t.Logf("System message: %s", *resp.Msg)
					}
				}
			})
		})
	}
}

// TestGetCapitalConfigGetall tests getting all coins' information
func TestGetCapitalConfigGetall(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetCapitalConfigGetall", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetCapitalConfigGetallV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Capital config endpoint not available on testnet")
					}
					t.Fatalf("Failed to get capital config: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if len(resp) == 0 {
					t.Error("Expected coin information in response")
				}
				
				// Check first few coins
				foundBTC := false
				foundUSDT := false
				
				for _, coin := range resp {
					if coin.Coin != nil {
						if *coin.Coin == "BTC" {
							foundBTC = true
							if coin.Free == nil {
								t.Error("Expected free balance for BTC")
							}
							if coin.Locked == nil {
								t.Error("Expected locked balance for BTC")
							}
							if coin.NetworkList == nil || len(coin.NetworkList) == 0 {
								t.Error("Expected network list for BTC")
							}
						} else if *coin.Coin == "USDT" {
							foundUSDT = true
						}
					}
				}
				
				if !foundBTC {
					t.Error("Expected to find BTC in coin list")
				}
				if !foundUSDT {
					t.Error("Expected to find USDT in coin list")
				}
				
				t.Logf("Found %d coins in capital config", len(resp))
			})
		})
	}
}

// TestWalletAccountInfo tests the account information endpoint (wallet version)
func TestWalletAccountInfo(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAccountInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAccountInfoV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Account info endpoint not available on testnet")
					}
					t.Fatalf("Failed to get account info: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.VipLevel == nil {
					t.Error("Expected VIP level in response")
				} else {
					t.Logf("VIP Level: %d", *resp.VipLevel)
				}
			})
		})
	}
}

// TestGetAssetDetail tests getting asset detail
func TestGetAssetDetail(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAssetDetail", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAssetAssetDetailV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Asset detail endpoint not available on testnet")
					}
					t.Fatalf("Failed to get asset detail: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response has asset info
				if resp == nil || len(*resp) == 0 {
					t.Error("Expected asset information in response")
				} else {
					// Check if any asset info is returned
					for asset, details := range *resp {
						t.Logf("Asset: %s, Details: %+v", asset, details)
						if details.MinWithdrawAmount != nil && *details.MinWithdrawAmount != "" {
							t.Logf("Minimum withdrawal amount for %s: %s", asset, *details.MinWithdrawAmount)
						}
						break // Just check the first asset for validation
					}
				}
			})
		})
	}
}

// TestGetDepositHistory tests getting deposit history
func TestGetDepositHistory(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetDepositHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetCapitalDepositHisrecV1(ctx).
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Deposit history endpoint not available on testnet")
					}
					t.Fatalf("Failed to get deposit history: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty if no deposits
				t.Logf("Found %d deposits", len(resp))
				
				// If there are deposits, verify structure
				if len(resp) > 0 {
					deposit := resp[0]
					if deposit.Coin == nil || *deposit.Coin == "" {
						t.Error("Expected coin in deposit")
					}
					if deposit.Amount == nil || *deposit.Amount == "" {
						t.Error("Expected amount in deposit")
					}
					if deposit.Status == nil {
						t.Error("Expected status in deposit")
					}
					if deposit.InsertTime == nil {
						t.Error("Expected insert time in deposit")
					}
				}
			})
		})
	}
}

// TestGetWithdrawHistory tests getting withdrawal history
func TestGetWithdrawHistory(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetWithdrawHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetCapitalWithdrawHistoryV1(ctx).
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Withdraw history endpoint not available on testnet")
					}
					t.Fatalf("Failed to get withdraw history: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty if no withdrawals
				t.Logf("Found %d withdrawals", len(resp))
				
				// If there are withdrawals, verify structure
				if len(resp) > 0 {
					withdrawal := resp[0]
					if withdrawal.Coin == nil || *withdrawal.Coin == "" {
						t.Error("Expected coin in withdrawal")
					}
					if withdrawal.Amount == nil || *withdrawal.Amount == "" {
						t.Error("Expected amount in withdrawal")
					}
					if withdrawal.Status == nil {
						t.Error("Expected status in withdrawal")
					}
					if withdrawal.ApplyTime == nil || *withdrawal.ApplyTime == "" {
						t.Error("Expected apply time in withdrawal")
					}
				}
			})
		})
	}
}

// TestGetDepositAddress tests getting deposit address
func TestGetDepositAddress(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetDepositAddress", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetCapitalDepositAddressV1(ctx).
					Coin("USDT").
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Deposit address endpoint not available on testnet")
					}
					t.Fatalf("Failed to get deposit address: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Coin == nil || *resp.Coin != "USDT" {
					t.Errorf("Expected coin USDT, got %v", resp.Coin)
				}
				
				if resp.Address == nil || *resp.Address == "" {
					t.Error("Expected deposit address")
				} else {
					t.Logf("USDT deposit address: %s", *resp.Address)
				}
				
				if resp.Url != nil && *resp.Url != "" {
					t.Logf("Deposit URL: %s", *resp.Url)
				}
			})
		})
	}
}

// TestGetAccountSnapshot tests getting account snapshots
func TestGetAccountSnapshot(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAccountSnapshot", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAccountSnapshotV1(ctx).
					Type_("SPOT").
					Limit(7).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Account snapshot endpoint not available on testnet")
					}
					t.Fatalf("Failed to get account snapshot: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Code == nil || *resp.Code != 200 {
					t.Errorf("Expected code 200, got %v", resp.Code)
				}
				
				if resp.SnapshotVos == nil {
					t.Error("Expected snapshot data")
				} else {
					t.Logf("Found %d snapshots", len(resp.SnapshotVos))
					
					if len(resp.SnapshotVos) > 0 {
						snapshot := resp.SnapshotVos[0]
						if snapshot.UpdateTime == nil {
							t.Error("Expected update time in snapshot")
						}
						if snapshot.Type == nil || *snapshot.Type != "spot" {
							t.Errorf("Expected type 'spot', got %v", snapshot.Type)
						}
					}
				}
			})
		})
	}
}

// TestGetAssetDividend tests getting asset dividend record
func TestGetAssetDividend(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAssetDividend", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAssetAssetDividendV1(ctx).
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("Asset dividend endpoint not available on testnet")
					}
					t.Fatalf("Failed to get asset dividend: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Total == nil {
					t.Error("Expected total in response")
				} else {
					t.Logf("Total dividend records: %d", *resp.Total)
				}
				
				if resp.Rows != nil && len(resp.Rows) > 0 {
					dividend := resp.Rows[0]
					if dividend.Asset == nil || *dividend.Asset == "" {
						t.Error("Expected asset in dividend")
					}
					if dividend.Amount == nil || *dividend.Amount == "" {
						t.Error("Expected amount in dividend")
					}
					if dividend.DivTime == nil {
						t.Error("Expected dividend time")
					}
				}
			})
		})
	}
}

// TestDisableFastWithdraw tests disabling fast withdraw
func TestDisableFastWithdraw(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "DisableFastWithdraw", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.CreateAccountDisableFastWithdrawSwitchV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				
				if err != nil {
					checkAPIError(t, err)
					
					// Handle specific error cases
					if httpResp != nil {
						if httpResp.StatusCode == 404 {
							t.Skip("Fast withdraw switch endpoint not available on testnet")
						}
						
						// Check if testnet returned HTML error response
						contentType := httpResp.Header.Get("Content-Type")
						if strings.Contains(contentType, "text/html") {
							t.Skip("Fast withdraw switch endpoint returns HTML error on testnet - endpoint not supported")
						}
					}
					
					// Check for SDK-specific "undefined response type" error (likely from HTML response)
					if strings.Contains(err.Error(), "undefined response type") {
						t.Skip("Fast withdraw switch endpoint not properly supported on testnet (HTML response received)")
					}
					
					t.Fatalf("Failed to disable fast withdraw: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response is usually empty on success
				_ = resp
				t.Log("Fast withdraw disabled successfully")
				
				// Re-enable fast withdraw to restore original state
				time.Sleep(1 * time.Second)
				rateLimiter.WaitForRateLimit()
				
				enableReq := client.WalletAPI.CreateAccountEnableFastWithdrawSwitchV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				_, _, enableErr := enableReq.Execute()
				if enableErr != nil {
					t.Logf("Warning: Failed to re-enable fast withdraw: %v", enableErr)
				}
			})
		})
	}
}

// TestGetAPITradingStatus tests getting API trading status
func TestGetAPITradingStatus(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetAPITradingStatus", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				req := client.WalletAPI.GetAccountApiTradingStatusV1(ctx).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					// This endpoint might not be available on testnet
					if httpResp != nil && httpResp.StatusCode == 404 {
						t.Skip("API trading status endpoint not available on testnet")
					}
					t.Fatalf("Failed to get API trading status: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.Data == nil {
					t.Error("Expected data in response")
				} else if resp.Data.IsLocked == nil {
					t.Error("Expected isLocked status")
				} else {
					t.Logf("API trading locked: %v", *resp.Data.IsLocked)
				}
				
				if resp.Data.PlannedRecoverTime != nil && *resp.Data.PlannedRecoverTime > 0 {
					t.Logf("Planned recover time: %d", *resp.Data.PlannedRecoverTime)
				}
				
				if resp.Data.TriggerCondition != nil {
					t.Logf("Trigger condition - GCR: %d, IFER: %d, UFR: %d", 
						resp.Data.TriggerCondition.GCR, 
						resp.Data.TriggerCondition.IFER, 
						resp.Data.TriggerCondition.UFR)
				}
				
				if resp.Data.UpdateTime != nil && *resp.Data.UpdateTime > 0 {
					updateTime := time.Unix(*resp.Data.UpdateTime/1000, 0)
					t.Logf("Last update: %v", updateTime)
				}
			})
		})
	}
}