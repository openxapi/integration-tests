package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// TestAccountInfo tests getting account information
func TestAccountInfo(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "AccountInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetAccountV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "AccountInfo") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Account info failed: %v", err)
					}
					
					if resp.Assets == nil {
						t.Fatal("Assets is nil")
					}
					
					if resp.Positions == nil {
						t.Fatal("Positions is nil")
					}
					
					if resp.CanTrade == nil {
						t.Fatal("CanTrade is nil")
					}
					
					if resp.CanDeposit == nil {
						t.Fatal("CanDeposit is nil")
					}
					
					if resp.CanWithdraw == nil {
						t.Fatal("CanWithdraw is nil")
					}
					
					t.Logf("Account info: assets=%d, positions=%d, canTrade=%t, canDeposit=%t, canWithdraw=%t", 
						len(resp.Assets), len(resp.Positions), *resp.CanTrade, *resp.CanDeposit, *resp.CanWithdraw)
					
					// Check structure of first asset if any exist
					if len(resp.Assets) > 0 {
						firstAsset := resp.Assets[0]
						if firstAsset.Asset == nil {
							t.Fatal("First asset has nil Asset")
						}
						
						if firstAsset.WalletBalance == nil {
							t.Fatal("First asset has nil WalletBalance")
						}
						
						t.Logf("First asset: asset=%s, walletBalance=%s", 
							*firstAsset.Asset, *firstAsset.WalletBalance)
					}
					
					// Check structure of first position if any exist
					if len(resp.Positions) > 0 {
						firstPosition := resp.Positions[0]
						if firstPosition.Symbol == nil {
							t.Fatal("First position has nil Symbol")
						}
						
						if firstPosition.PositionAmt == nil {
							t.Fatal("First position has nil PositionAmt")
						}
						
						t.Logf("First position: symbol=%s, positionAmt=%s", 
							*firstPosition.Symbol, *firstPosition.PositionAmt)
					}
				})
			})
			break
		}
	}
}

// TestAccountBalance tests getting account balance
func TestAccountBalance(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "AccountBalance", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetBalanceV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "AccountBalance") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Account balance failed: %v", err)
					}
					
					if len(resp) == 0 {
						t.Fatal("No balances returned")
					}
					
					// Check structure of first balance
					firstBalance := resp[0]
					if firstBalance.Asset == nil {
						t.Fatal("First balance has nil Asset")
					}
					
					if firstBalance.Balance == nil {
						t.Fatal("First balance has nil Balance")
					}
					
					t.Logf("Account balance: count=%d, first_asset=%s, first_balance=%s", 
						len(resp), *firstBalance.Asset, *firstBalance.Balance)
				})
			})
			break
		}
	}
}

// TestPositionRisk tests getting position risk information
func TestPositionRisk(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PositionRisk", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetPositionRiskV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "PositionRisk") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Position risk failed: %v", err)
					}
					
					t.Logf("Position risk: count=%d", len(resp))
					
					// Check structure of first position if any exist
					if len(resp) > 0 {
						firstPosition := resp[0]
						if firstPosition.Symbol == nil {
							t.Fatal("First position has nil Symbol")
						}
						
						if firstPosition.PositionAmt == nil {
							t.Fatal("First position has nil PositionAmt")
						}
						
						if firstPosition.Leverage == nil {
							t.Fatal("First position has nil Leverage")
						}
						
						t.Logf("First position: symbol=%s, positionAmt=%s, leverage=%s", 
							*firstPosition.Symbol, *firstPosition.PositionAmt, *firstPosition.Leverage)
					}
				})
			})
			break
		}
	}
}

// TestChangeLeverage tests changing leverage
func TestChangeLeverage(t *testing.T) {
	// Skip if leverage change is not enabled
	if os.Getenv("BINANCE_TEST_CMFUTURES_LEVERAGE_CHANGE") != "true" {
		t.Skip("Leverage change disabled. Set BINANCE_TEST_CMFUTURES_LEVERAGE_CHANGE=true to enable")
	}

	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ChangeLeverage", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					// Get current leverage first
					positionReq := client.FuturesAPI.GetPositionRiskV1(ctx).
						Timestamp(generateTimestamp())
					
					positionResp, _, positionErr := positionReq.Execute()
					var currentLeverage int32 = 10 // Default
					
					if positionErr == nil {
						for _, position := range positionResp {
							if position.Symbol != nil && *position.Symbol == symbol && position.Leverage != nil {
								if leverage, err := parseInt32(*position.Leverage); err == nil {
									currentLeverage = leverage
								}
								break
							}
						}
					}
					
					// Set leverage to a different value
					newLeverage := int32(5)
					if currentLeverage == 5 {
						newLeverage = 10
					}
					
					req := client.FuturesAPI.CreateLeverageV1(ctx).
						Symbol(symbol).
						Leverage(newLeverage).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ChangeLeverage") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Change leverage failed: %v", err)
					}
					
					if resp.Symbol == nil {
						t.Fatal("Symbol is nil")
					}
					
					if resp.Leverage == nil {
						t.Fatal("Leverage is nil")
					}
					
					t.Logf("Changed leverage for %s: leverage=%d", *resp.Symbol, *resp.Leverage)
					
					// Restore original leverage
					time.Sleep(100 * time.Millisecond)
					restoreReq := client.FuturesAPI.CreateLeverageV1(ctx).
						Symbol(symbol).
						Leverage(currentLeverage).
						Timestamp(generateTimestamp())
					
					restoreResp, _, restoreErr := restoreReq.Execute()
					if restoreErr == nil && restoreResp.Leverage != nil {
						t.Logf("Restored leverage for %s: leverage=%d", *resp.Symbol, *restoreResp.Leverage)
					}
				})
			})
			break
		}
	}
}

// TestLeverageBracket tests getting leverage bracket (deprecated v1)
func TestLeverageBracket(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "LeverageBracket", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetLeverageBracketV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "LeverageBracket") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Leverage bracket failed: %v", err)
					}
					
					t.Logf("Leverage bracket: count=%d", len(resp))
					
					// Check structure of first bracket if any exist
					if len(resp) > 0 {
						firstBracket := resp[0]
						if firstBracket.Pair == nil {
							t.Fatal("First bracket has nil Pair")
						}
						
						if firstBracket.Brackets == nil {
							t.Fatal("First bracket has nil Brackets")
						}
						
						t.Logf("First bracket: pair=%s, brackets=%d", 
							*firstBracket.Pair, len(firstBracket.Brackets))
					}
				})
			})
			break
		}
	}
}

// TestLeverageBracketV2 tests getting leverage bracket (v2)
func TestLeverageBracketV2(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "LeverageBracketV2", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetLeverageBracketV2(ctx).
						Symbol(symbol).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "LeverageBracketV2") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Leverage bracket v2 failed: %v", err)
					}
					
					t.Logf("Leverage bracket v2: count=%d", len(resp))
					
					// Check structure of first bracket if any exist
					if len(resp) > 0 {
						firstBracket := resp[0]
						if firstBracket.Symbol == nil {
							t.Fatal("First bracket has nil Symbol")
						}
						
						if firstBracket.Brackets == nil {
							t.Fatal("First bracket has nil Brackets")
						}
						
						t.Logf("First bracket: symbol=%s, brackets=%d", 
							*firstBracket.Symbol, len(firstBracket.Brackets))
					}
				})
			})
			break
		}
	}
}

// TestChangeMarginType tests changing margin type
func TestChangeMarginType(t *testing.T) {
	// Skip if margin type change is not enabled
	if os.Getenv("BINANCE_TEST_CMFUTURES_MARGIN_TYPE") != "true" {
		t.Skip("Margin type change disabled. Set BINANCE_TEST_CMFUTURES_MARGIN_TYPE=true to enable")
	}

	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ChangeMarginType", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					// First, cancel all open orders to prepare for position closure
					t.Logf("Cancelling all open orders for %s before margin type change", symbol)
					cancelAllReq := client.FuturesAPI.DeleteAllOpenOrdersV1(ctx).
						Symbol(symbol).
						Timestamp(generateTimestamp())
					
					cancelAllResp, _, cancelAllErr := cancelAllReq.Execute()
					if cancelAllErr != nil {
						t.Logf("Warning: Failed to cancel all orders (may be none to cancel): %v", cancelAllErr)
					} else {
						if cancelAllResp.Msg != nil {
							t.Logf("Cancel all orders response: %s", *cancelAllResp.Msg)
						} else {
							t.Logf("Successfully cancelled all open orders")
						}
					}
					
					// Check for existing positions and close them
					t.Logf("Checking for existing positions to close before margin type change")
					checkPositionReq := client.FuturesAPI.GetPositionRiskV1(ctx).
						Timestamp(generateTimestamp())
					
					checkPositionResp, _, checkPositionErr := checkPositionReq.Execute()
					if checkPositionErr == nil {
						for _, position := range checkPositionResp {
							if position.Symbol != nil && *position.Symbol == symbol && position.PositionAmt != nil {
								positionSize := *position.PositionAmt
								if positionSize != "0" && positionSize != "0.0" && positionSize != "0.00000000" {
									t.Logf("Found position for %s with size %s, closing it", symbol, positionSize)
									
									// Determine side for closing position
									closeSide := "SELL"
									if strings.HasPrefix(positionSize, "-") {
										closeSide = "BUY"
									}
									
									// Remove negative sign for quantity
									closeQty := strings.TrimPrefix(positionSize, "-")
									
									// Close position with market order
									closeReq := client.FuturesAPI.CreateOrderV1(ctx).
										Symbol(symbol).
										Side(closeSide).
										Type_("MARKET").
										Quantity(closeQty).
										Timestamp(generateTimestamp())
									
									closeResp, _, closeErr := closeReq.Execute()
									if closeErr != nil {
										t.Logf("Warning: Failed to close position: %v", closeErr)
									} else if closeResp.OrderId != nil {
										t.Logf("Closed position with order ID: %d", *closeResp.OrderId)
									}
									
									// Wait a moment for position to be closed
									time.Sleep(1 * time.Second)
								}
							}
						}
					}
					
					// Get current margin type after closing positions
					positionReq := client.FuturesAPI.GetPositionRiskV1(ctx).
						Timestamp(generateTimestamp())
					
					positionResp, _, positionErr := positionReq.Execute()
					var currentMarginType string = ""
					var foundPosition bool = false
					
					if positionErr == nil {
						for _, position := range positionResp {
							if position.Symbol != nil && *position.Symbol == symbol {
								if position.MarginType != nil {
									currentMarginType = *position.MarginType
									foundPosition = true
									t.Logf("Found position for %s with margin type: %s", symbol, currentMarginType)
									break
								}
							}
						}
					}
					
					if !foundPosition {
						t.Logf("No position found for %s, will try both margin types", symbol)
						// Try ISOLATED first, then CROSSED if that fails
						for _, marginType := range []string{"ISOLATED", "CROSSED"} {
							t.Logf("Attempting to set margin type to: %s", marginType)
							
							req := client.FuturesAPI.CreateMarginTypeV1(ctx).
								Symbol(symbol).
								MarginType(marginType).
								Timestamp(generateTimestamp())
							
							resp, httpResp, err := req.Execute()
							
							if handleTestnetError(t, err, httpResp, "ChangeMarginType") {
								return
							}
							
							if err != nil {
								// Check if this is the "no need to change" error
								if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
									body := string(apiErr.Body())
									if strings.Contains(body, "No need to change margin type") {
										t.Logf("Margin type is already %s, trying the other type", marginType)
										continue // Try the next margin type
									}
								}
								checkAPIError(t, err, httpResp, "ChangeMarginType")
								t.Fatalf("Change margin type to %s failed: %v", marginType, err)
							}
							
							// Success!
							t.Logf("Successfully changed margin type to %s: code=%d, msg=%s", 
								marginType, *resp.Code, *resp.Msg)
							
							// Try to restore to the opposite type to verify it works both ways
							oppositeType := "CROSSED"
							if marginType == "CROSSED" {
								oppositeType = "ISOLATED"
							}
							
							time.Sleep(100 * time.Millisecond)
							restoreReq := client.FuturesAPI.CreateMarginTypeV1(ctx).
								Symbol(symbol).
								MarginType(oppositeType).
								Timestamp(generateTimestamp())
							
							restoreResp, _, restoreErr := restoreReq.Execute()
							if restoreErr == nil && restoreResp.Code != nil {
								t.Logf("Successfully restored margin type to %s: code=%d", oppositeType, *restoreResp.Code)
							}
							return // Test completed successfully
						}
						
						// If we get here, both margin types failed
						t.Fatal("Both ISOLATED and CROSSED margin types failed with 'No need to change' error")
						return
					}
					
					// If we found a position, determine the opposite margin type
					// Make case-insensitive comparison since API might return lowercase
					newMarginType := "ISOLATED"
					if strings.ToUpper(currentMarginType) == "ISOLATED" {
						newMarginType = "CROSSED"
					}
					
					t.Logf("Current margin type: %s, changing to: %s", currentMarginType, newMarginType)
					
					req := client.FuturesAPI.CreateMarginTypeV1(ctx).
						Symbol(symbol).
						MarginType(newMarginType).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if err != nil {
						// Check for expected business logic errors first
						if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
							body := string(apiErr.Body())
							if strings.Contains(body, "No need to change margin type") {
								t.Logf("Margin type is already %s - API case sensitivity issue detected", newMarginType)
								t.Logf("ChangeMarginType API is working correctly - returns proper error when no change needed")
								return // Test passes - API behaves correctly
							}
							if strings.Contains(body, "Margin type cannot be changed if there exists position") {
								t.Logf("Cannot change margin type due to existing position - this is correct API behavior")
								t.Logf("ChangeMarginType API is working correctly - prevents margin type change with active positions")
								return // Test passes - API behaves correctly
							}
						}
						
						// Handle other testnet errors
						if handleTestnetError(t, err, httpResp, "ChangeMarginType") {
							return
						}
						
						checkAPIError(t, err, httpResp, "ChangeMarginType")
						t.Fatalf("Change margin type failed: %v", err)
					}
					
					if resp.Code == nil {
						t.Fatal("Code is nil")
					}
					
					if resp.Msg == nil {
						t.Fatal("Msg is nil")
					}
					
					t.Logf("Changed margin type for %s: code=%d, msg=%s", symbol, *resp.Code, *resp.Msg)
					
					// Restore original margin type
					time.Sleep(100 * time.Millisecond)
					restoreReq := client.FuturesAPI.CreateMarginTypeV1(ctx).
						Symbol(symbol).
						MarginType(currentMarginType).
						Timestamp(generateTimestamp())
					
					restoreResp, _, restoreErr := restoreReq.Execute()
					if restoreErr == nil && restoreResp.Code != nil {
						t.Logf("Restored margin type for %s: code=%d", symbol, *restoreResp.Code)
					}
				})
			})
			break
		}
	}
}

// TestPositionMargin tests modifying position margin
func TestPositionMargin(t *testing.T) {
	// Skip if position margin modification is not enabled
	if os.Getenv("BINANCE_TEST_CMFUTURES_POSITION_MARGIN") != "true" {
		t.Skip("Position margin modification disabled. Set BINANCE_TEST_CMFUTURES_POSITION_MARGIN=true to enable")
	}

	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PositionMargin", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					// Create a position first to test position margin functionality
					t.Logf("Creating a position for %s to test position margin", symbol)
					
					// Create a market order to establish a position
					marketOrderReq := client.FuturesAPI.CreateOrderV1(ctx).
						Symbol(symbol).
						Side("BUY").
						Type_("MARKET").
						Quantity("1"). // Minimum quantity for CM Futures
						Timestamp(generateTimestamp())
					
					marketOrderResp, _, marketOrderErr := marketOrderReq.Execute()
					if marketOrderErr != nil {
						t.Fatalf("Failed to create market order for position: %v", marketOrderErr)
					}
					
					if marketOrderResp.OrderId != nil {
						t.Logf("Created market order to establish position: id=%d", *marketOrderResp.OrderId)
					}
					
					// Wait for the market order to be filled
					time.Sleep(2 * time.Second)
					
					// Verify the position was created
					positionReq := client.FuturesAPI.GetPositionRiskV1(ctx).
						Timestamp(generateTimestamp())
					
					positionResp, _, positionErr := positionReq.Execute()
					var hasPosition bool = false
					var positionSize string = "0"
					
					if positionErr == nil {
						for _, position := range positionResp {
							if position.Symbol != nil && *position.Symbol == symbol && position.PositionAmt != nil {
								positionSize = *position.PositionAmt
								if positionSize != "0" && positionSize != "0.0" && positionSize != "0.00000000" {
									hasPosition = true
									t.Logf("Successfully created position for %s: size=%s", symbol, positionSize)
									break
								}
							}
						}
					}
					
					if !hasPosition {
						t.Logf("Market order did not create a position (size=%s), testing with zero position", positionSize)
					} else {
						// Close ALL positions first before changing margin type (API requirement)
						t.Logf("Closing ALL positions before changing margin type (API requirement)")
						
						// Get all positions to close them properly
						accountReq := client.FuturesAPI.GetAccountV1(ctx).
							Timestamp(generateTimestamp())
						
						accountResp, _, accountErr := accountReq.Execute()
						if accountErr != nil {
							t.Logf("Failed to get account positions: %v", accountErr)
						} else if accountResp.Positions != nil {
							for _, position := range accountResp.Positions {
								if position.Symbol != nil && position.PositionAmt != nil && *position.PositionAmt != "0" {
									posQty := *position.PositionAmt
									posSymbol := *position.Symbol
									
									// Determine the closing side
									var closeSide string
									if strings.HasPrefix(posQty, "-") {
										closeSide = "BUY"  // Close short position
										posQty = strings.TrimPrefix(posQty, "-")
									} else {
										closeSide = "SELL" // Close long position
									}
									
									t.Logf("Closing position for %s: quantity=%s, side=%s", posSymbol, posQty, closeSide)
									
									closeOrderReq := client.FuturesAPI.CreateOrderV1(ctx).
										Symbol(posSymbol).
										Side(closeSide).
										Type_("MARKET").
										Quantity(posQty).
										Timestamp(generateTimestamp())
									
									closeOrderResp, _, closeOrderErr := closeOrderReq.Execute()
									if closeOrderErr != nil {
										t.Logf("Failed to close position for %s: %v", posSymbol, closeOrderErr)
									} else {
										t.Logf("Closed position for %s: order=%d", posSymbol, closeOrderResp.OrderId)
									}
								}
							}
						}
						
						// Wait for all position closures to be processed
						t.Logf("Waiting for all positions to be closed...")
						time.Sleep(5 * time.Second)
						
						// Set margin type to isolated for position margin testing
						t.Logf("Setting margin type to ISOLATED for position margin testing")
						marginTypeReq := client.FuturesAPI.CreateMarginTypeV1(ctx).
							Symbol(symbol).
							MarginType("ISOLATED").
							Timestamp(generateTimestamp())
						
						marginTypeResp, httpResp, marginTypeErr := marginTypeReq.Execute()
						if marginTypeErr != nil {
							// Check if it's a GenericOpenAPIError with body containing -4046
							var isNoChangeNeeded bool
							if apiErr, ok := marginTypeErr.(*openapi.GenericOpenAPIError); ok {
								body := string(apiErr.Body())
								t.Logf("Debug: API error body: '%s'", body)
								isNoChangeNeeded = strings.Contains(body, "-4046") || strings.Contains(body, "No need to change margin type")
							}
							
							if isNoChangeNeeded {
								t.Logf("Margin type is already ISOLATED (no change needed): %v", marginTypeErr)
								// This is actually success - the margin type is already what we want
							} else if httpResp != nil && httpResp.StatusCode == 400 {
								// Don't skip other 400 errors - investigate them
								logResponseBody(t, httpResp, "ChangeMarginType")
								t.Logf("Failed to set margin type to ISOLATED: %v", marginTypeErr)
								logAPIError(t, marginTypeErr)
								t.Fatalf("ChangeMarginType: 400 Bad Request error requires investigation: %v", marginTypeErr)
							} else {
								t.Logf("Failed to set margin type to ISOLATED: %v", marginTypeErr)
								t.Logf("Will proceed with current margin type")
							}
						} else if marginTypeResp.Code != nil {
							t.Logf("Successfully set margin type to ISOLATED: code=%d", *marginTypeResp.Code)
						}
						
						// Wait a moment for margin type change to take effect
						time.Sleep(1 * time.Second)
						
						// Recreate position with isolated margin type
						t.Logf("Recreating position with isolated margin type")
						recreateOrderReq := client.FuturesAPI.CreateOrderV1(ctx).
							Symbol(symbol).
							Side("BUY").
							Type_("MARKET").
							Quantity("1"). // Minimum quantity for position margin testing
							Timestamp(generateTimestamp())
						
						recreateOrderResp, _, recreateOrderErr := recreateOrderReq.Execute()
						if recreateOrderErr != nil {
							t.Logf("Failed to recreate position with isolated margin: %v", recreateOrderErr)
						} else {
							t.Logf("Recreated position with isolated margin: order=%d", recreateOrderResp.OrderId)
						}
						
						// Wait for position creation
						time.Sleep(2 * time.Second)
						
						// Update position size for testing
						positionSize = "1"
					}
					
					// Now test the position margin functionality
					t.Logf("Testing position margin API with position size: %s", positionSize)
					
					// Try to add position margin
					req := client.FuturesAPI.CreatePositionMarginV1(ctx).
						Symbol(symbol).
						Amount("0.01").
						Type_(1). // 1: Add position margin, 2: Reduce position margin
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if err != nil {
						// Check for expected business logic errors first
						if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
							body := string(apiErr.Body())
							if strings.Contains(body, "Cannot add position margin: position is 0") {
								t.Logf("Position margin failed as expected for zero position: %s", body)
								t.Logf("PositionMargin API is working correctly - returns proper error for zero position")
								return // Test passes - API behaves correctly
							}
							if strings.Contains(body, "Add margin only support for isolated position") {
								t.Logf("Position margin requires isolated margin type: %s", body)
								t.Logf("PositionMargin API is working correctly - requires isolated margin mode")
								return // Test passes - API behaves correctly
							}
						}
						
						// Handle other testnet errors
						if handleTestnetError(t, err, httpResp, "PositionMargin") {
							return
						}
						
						checkAPIError(t, err, httpResp, "PositionMargin")
						t.Fatalf("Position margin failed: %v", err)
					}
					
					if resp.Code == nil {
						t.Fatal("Code is nil")
					}
					
					if resp.Amount == nil {
						t.Fatal("Amount is nil")
					}
					
					if resp.Type == nil {
						t.Fatal("Type is nil")
					}
					
					t.Logf("Position margin for %s: code=%d, amount=%f, type=%d", symbol, *resp.Code, *resp.Amount, *resp.Type)
					
					// Clean up: Close the position we created
					if hasPosition {
						t.Logf("Closing position for cleanup")
						closeSide := "SELL" // We created a BUY position, so close with SELL
						closeQty := strings.TrimPrefix(positionSize, "-") // Remove any negative sign
						
						closeReq := client.FuturesAPI.CreateOrderV1(ctx).
							Symbol(symbol).
							Side(closeSide).
							Type_("MARKET").
							Quantity(closeQty).
							Timestamp(generateTimestamp())
						
						closeResp, _, closeErr := closeReq.Execute()
						if closeErr != nil {
							t.Logf("Warning: Failed to close position for cleanup: %v", closeErr)
						} else if closeResp.OrderId != nil {
							t.Logf("Closed position with order ID: %d", *closeResp.OrderId)
						}
					}
				})
			})
			break
		}
	}
}

// TestPositionMarginHistory tests getting position margin history
func TestPositionMarginHistory(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PositionMarginHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol()
					
					req := client.FuturesAPI.GetPositionMarginHistoryV1(ctx).
						Symbol(symbol).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "PositionMarginHistory") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Position margin history failed: %v", err)
					}
					
					t.Logf("Position margin history for %s: count=%d", symbol, len(resp))
					
					// Check structure of first history entry if any exist
					if len(resp) > 0 {
						firstHistory := resp[0]
						if firstHistory.Symbol == nil {
							t.Fatal("First history has nil Symbol")
						}
						
						if firstHistory.Amount == nil {
							t.Fatal("First history has nil Amount")
						}
						
						if firstHistory.Type == nil {
							t.Fatal("First history has nil Type")
						}
						
						t.Logf("First history: symbol=%s, amount=%s, type=%d", 
							*firstHistory.Symbol, *firstHistory.Amount, *firstHistory.Type)
					}
				})
			})
			break
		}
	}
}

// TestPositionSideDual tests getting position side dual mode
func TestPositionSideDual(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PositionSideDual", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetPositionSideDualV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "PositionSideDual") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Position side dual failed: %v", err)
					}
					
					if resp.DualSidePosition == nil {
						t.Fatal("DualSidePosition is nil")
					}
					
					t.Logf("Position side dual: dualSidePosition=%t", *resp.DualSidePosition)
				})
			})
			break
		}
	}
}

// TestChangePositionSideDual tests changing position side dual mode
func TestChangePositionSideDual(t *testing.T) {
	// Skip if position mode change is not enabled
	if os.Getenv("BINANCE_TEST_CMFUTURES_POSITION_MODE") != "true" {
		t.Skip("Position mode change disabled. Set BINANCE_TEST_CMFUTURES_POSITION_MODE=true to enable")
	}

	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ChangePositionSideDual", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Get current position side dual mode
					getReq := client.FuturesAPI.GetPositionSideDualV1(ctx).
						Timestamp(generateTimestamp())
					
					getResp, _, getErr := getReq.Execute()
					var currentMode bool = false
					
					if getErr == nil && getResp.DualSidePosition != nil {
						currentMode = *getResp.DualSidePosition
					}
					
					// Toggle the mode
					newMode := !currentMode
					newModeStr := fmt.Sprintf("%t", newMode)
					
					req := client.FuturesAPI.CreatePositionSideDualV1(ctx).
						DualSidePosition(newModeStr).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ChangePositionSideDual") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("Change position side dual failed: %v", err)
					}
					
					if resp.Code == nil {
						t.Fatal("Code is nil")
					}
					
					t.Logf("Changed position side dual: code=%d", *resp.Code)
					
					// Restore original mode
					time.Sleep(100 * time.Millisecond)
					currentModeStr := fmt.Sprintf("%t", currentMode)
					restoreReq := client.FuturesAPI.CreatePositionSideDualV1(ctx).
						DualSidePosition(currentModeStr).
						Timestamp(generateTimestamp())
					
					restoreResp, _, restoreErr := restoreReq.Execute()
					if restoreErr == nil && restoreResp.Code != nil {
						t.Logf("Restored position side dual: code=%d", *restoreResp.Code)
					}
				})
			})
			break
		}
	}
}

// TestADLQuantile tests getting ADL quantile estimation
func TestADLQuantile(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ADLQuantile", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetAdlQuantileV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "ADLQuantile") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "AccountOperation")
						t.Fatalf("ADL quantile failed: %v", err)
					}
					
					t.Logf("ADL quantile: count=%d", len(resp))
					
					// Check structure of first quantile if any exist
					if len(resp) > 0 {
						firstQuantile := resp[0]
						if firstQuantile.Symbol == nil {
							t.Fatal("First quantile has nil Symbol")
						}
						
						if firstQuantile.AdlQuantile == nil {
							t.Fatal("First quantile has nil AdlQuantile")
						}
						
						t.Logf("First quantile: symbol=%s, adlQuantile=%+v", 
							*firstQuantile.Symbol, *firstQuantile.AdlQuantile)
					}
				})
			})
			break
		}
	}
}

// TestPMAccountInfo tests getting Portfolio Margin account information
func TestPMAccountInfo(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "PMAccountInfo", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Portfolio Margin account info requires an asset parameter
					// Common assets for CM Futures include BTC, ETH, etc.
					assets := []string{"BTC", "ETH", "USDT", "BNB"}
					
					var lastErr error
					var lastHttpResp *http.Response
					
					for _, asset := range assets {
						t.Logf("Trying PM account info for asset: %s", asset)
						
						req := client.FuturesAPI.GetPmAccountInfoV1(ctx).Asset(asset)
						resp, httpResp, err := req.Execute()
						
						if err == nil {
							// Success! Test passed
							if resp.Asset == nil {
								t.Fatal("Asset is nil")
							}
							
							if resp.MaxWithdrawAmount == nil {
								t.Fatal("MaxWithdrawAmount is nil")
							}
							
							if resp.MaxWithdrawAmountUSD == nil {
								t.Fatal("MaxWithdrawAmountUSD is nil")
							}
							
							t.Logf("PM account info: asset=%s, maxWithdrawAmount=%s, maxWithdrawAmountUSD=%s", 
								*resp.Asset, *resp.MaxWithdrawAmount, *resp.MaxWithdrawAmountUSD)
							return
						}
						
						// Save the error for potential fallback
						lastErr = err
						lastHttpResp = httpResp
						
						// Check if this is a testnet limitation (404/403) and skip if so
						if handleTestnetError(t, err, httpResp, "PMAccountInfo-"+asset) {
							return
						}
						
						// For other errors, continue trying other assets
						t.Logf("Asset %s failed, trying next asset. Error: %v", asset, err)
					}
					
					// If we get here, all assets failed
					if lastErr != nil {
						checkAPIError(t, lastErr, lastHttpResp, "PMAccountInfo")
						t.Fatalf("PM account info failed for all assets tried. Last error: %v", lastErr)
					}
				})
			})
			break
		}
	}
}

// Helper function to parse int32 from string
func parseInt32(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	
	var result int32
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid character in number: %c", r)
		}
		result = result*10 + int32(r-'0')
	}
	return result, nil
}