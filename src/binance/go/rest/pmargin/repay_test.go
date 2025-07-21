package main

import (
	"context"
	"os"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestRepayFuturesNegativeBalance tests repaying futures negative balance
func TestRepayFuturesNegativeBalance(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_REPAY_FUTURES") != "true" {
		t.Skip("Repay futures test disabled - enable with BINANCE_TEST_PMARGIN_REPAY_FUTURES=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Repay Futures Negative Balance", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.CreateRepayFuturesNegativeBalanceV1(ctx).
						Timestamp(generateTimestamp())
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Repay Futures Negative Balance") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Repay Futures Negative Balance") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("CreateRepayFuturesNegativeBalanceV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Repay futures negative balance completed successfully")
					if resp.Msg != nil {
						t.Logf("Message: %s", *resp.Msg)
					}
				})
			})
		}
	}
}

// TestRepayFuturesSwitch tests changing auto-repay futures status
func TestRepayFuturesSwitch(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_REPAY_SWITCH") != "true" {
		t.Skip("Repay switch test disabled - enable with BINANCE_TEST_PMARGIN_REPAY_SWITCH=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Repay Futures Switch", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// First get current auto-repay status
					getReq := client.PortfolioMarginAPI.GetRepayFuturesSwitchV1(ctx).
						Timestamp(generateTimestamp())
					getResp, httpResp, err := getReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Repay Futures Switch") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Repay Futures Switch") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetRepayFuturesSwitchV1 failed: %v", err)
					}
					
					if getResp == nil {
						t.Fatal("Get response is nil")
					}
					
					t.Logf("Current auto-repay status retrieved successfully")
					if getResp.AutoRepay != nil {
						t.Logf("Auto-repay enabled: %t", *getResp.AutoRepay)
						
						// Test changing the switch (toggle current state)
						newStatus := !(*getResp.AutoRepay)
						// Convert bool to string as API expects string parameter
						newStatusStr := "false"
						if newStatus {
							newStatusStr = "true"
						}
						setReq := client.PortfolioMarginAPI.CreateRepayFuturesSwitchV1(ctx).
							AutoRepay(newStatusStr).
							Timestamp(generateTimestamp())
						
						setResp, httpResp, err := setReq.Execute()
						
						if handleTestnetError(t, err, httpResp, "Set Repay Futures Switch") {
							return
						}
						
						if err != nil {
							checkAPIError(t, err, httpResp)
							t.Fatalf("CreateRepayFuturesSwitchV1 failed: %v", err)
						} else {
							t.Logf("Auto-repay switch changed successfully")
							if setResp.Msg != nil {
								t.Logf("Message: %s", *setResp.Msg)
							}
							
							// Restore original state
							// Convert bool to string as API expects string parameter
							originalStatusStr := "false"
							if *getResp.AutoRepay {
								originalStatusStr = "true"
							}
							restoreReq := client.PortfolioMarginAPI.CreateRepayFuturesSwitchV1(ctx).
								AutoRepay(originalStatusStr).
								Timestamp(generateTimestamp())
							
							restoreResp, _, err := restoreReq.Execute()
							if err != nil {
								t.Logf("Failed to restore original auto-repay state: %v", err)
							} else {
								t.Logf("Original auto-repay state restored")
								if restoreResp.Msg != nil {
									t.Logf("Restore message: %s", *restoreResp.Msg)
								}
							}
						}
					}
				})
			})
		}
	}
}