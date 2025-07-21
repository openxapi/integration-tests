package main

import (
	"context"
	"os"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestMarginLoan tests margin loan operations
func TestMarginLoan(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_MARGIN_LOAN") != "true" {
		t.Skip("Margin loan test disabled - enable with BINANCE_TEST_PMARGIN_MARGIN_LOAN=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeMARGIN {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Margin Loan", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test getting margin loan record first
					getLoanReq := client.PortfolioMarginAPI.GetMarginMarginLoanV1(ctx).
						Asset("USDT").
						Timestamp(generateTimestamp())
					
					getLoanResp, httpResp, err := getLoanReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Loan") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Margin Loan") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginMarginLoanV1 failed: %v", err)
					} else {
						t.Logf("Margin loan record retrieved successfully")
						if getLoanResp != nil && getLoanResp.Rows != nil {
							t.Logf("Total loan records: %d", len(getLoanResp.Rows))
						}
					}
					
					// Test max borrowable amount
					maxBorrowReq := client.PortfolioMarginAPI.GetMarginMaxBorrowableV1(ctx).
						Asset("USDT").
						Timestamp(generateTimestamp())
					
					maxBorrowResp, httpResp, err := maxBorrowReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Max Borrowable") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginMaxBorrowableV1 failed: %v", err)
					} else {
						t.Logf("Max borrowable amount retrieved successfully")
						if maxBorrowResp != nil && maxBorrowResp.Amount != nil {
							t.Logf("Max borrowable USDT: %g", *maxBorrowResp.Amount)
						}
					}
				})
			})
		}
	}
}

// TestMarginOrder tests margin order operations
func TestMarginOrder(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_MARGIN_ORDERS") != "true" {
		t.Skip("Margin order test disabled - enable with BINANCE_TEST_PMARGIN_MARGIN_ORDERS=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Margin Order", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol("spot")
					
					// Test getting margin open orders
					openOrdersReq := client.PortfolioMarginAPI.GetMarginOpenOrdersV1(ctx).
						Symbol(symbol).
						Timestamp(generateTimestamp())
					
					openOrdersResp, httpResp, err := openOrdersReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Open Orders") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Margin Open Orders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginOpenOrdersV1 failed: %v", err)
					} else {
						t.Logf("Margin open orders retrieved successfully")
						if openOrdersResp != nil {
							t.Logf("Number of open orders: %d", len(openOrdersResp))
						}
					}
					
					// Test getting all margin orders
					allOrdersReq := client.PortfolioMarginAPI.GetMarginAllOrdersV1(ctx).
						Symbol(symbol).
						Limit(10).
						Timestamp(generateTimestamp())
					
					allOrdersResp, httpResp, err := allOrdersReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get All Margin Orders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginAllOrdersV1 failed: %v", err)
					} else {
						t.Logf("All margin orders retrieved successfully")
						if allOrdersResp != nil {
							t.Logf("Number of orders: %d", len(allOrdersResp))
							for i, order := range allOrdersResp {
								if i >= 3 { // Limit output
									break
								}
								if order.Symbol != nil && order.Side != nil && order.Type != nil {
									t.Logf("Order: %s %s %s", *order.Symbol, *order.Side, *order.Type)
								}
							}
						}
					}
				})
			})
		}
	}
}

// TestMarginOCOOrder tests margin OCO order operations
func TestMarginOCOOrder(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_MARGIN_OCO") != "true" {
		t.Skip("Margin OCO test disabled - enable with BINANCE_TEST_PMARGIN_MARGIN_OCO=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Margin OCO Order", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test getting margin OCO orders
					// Note: GetMarginAllOrderListV1 doesn't have Symbol parameter - it gets all margin OCO orders for account
					ocoOrdersReq := client.PortfolioMarginAPI.GetMarginAllOrderListV1(ctx).
						Limit(10).
						Timestamp(generateTimestamp())
					
					ocoOrdersResp, httpResp, err := ocoOrdersReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin OCO Orders") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Margin OCO Orders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginAllOrderListV1 failed: %v", err)
					} else {
						t.Logf("Margin OCO orders retrieved successfully")
						if ocoOrdersResp != nil {
							t.Logf("Number of OCO orders: %d", len(ocoOrdersResp))
						}
					}
					
					// Test getting open margin OCO orders
					openOcoReq := client.PortfolioMarginAPI.GetMarginOpenOrderListV1(ctx).
						Timestamp(generateTimestamp())
					openOcoResp, httpResp, err := openOcoReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Open Margin OCO Orders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginOpenOrderListV1 failed: %v", err)
					} else {
						t.Logf("Open margin OCO orders retrieved successfully")
						if openOcoResp != nil {
							t.Logf("Number of open OCO orders: %d", len(openOcoResp))
						}
					}
				})
			})
		}
	}
}

// TestMarginRepay tests margin repay operations
func TestMarginRepay(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_MARGIN_REPAY") != "true" {
		t.Skip("Margin repay test disabled - enable with BINANCE_TEST_PMARGIN_MARGIN_REPAY=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Margin Repay", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test getting margin repay record
					repayReq := client.PortfolioMarginAPI.GetMarginRepayLoanV1(ctx).
						Asset("USDT").
						Timestamp(generateTimestamp())
					
					repayResp, httpResp, err := repayReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Repay") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Margin Repay") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginRepayLoanV1 failed: %v", err)
					} else {
						t.Logf("Margin repay record retrieved successfully")
						if repayResp != nil && repayResp.Rows != nil {
							t.Logf("Total repay records: %d", len(repayResp.Rows))
						}
					}
					
					// Test getting margin interest history
					interestReq := client.PortfolioMarginAPI.GetMarginMarginInterestHistoryV1(ctx).
						Asset("USDT").
						Timestamp(generateTimestamp())
					
					interestResp, httpResp, err := interestReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Interest History") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginMarginInterestHistoryV1 failed: %v", err)
					} else {
						t.Logf("Margin interest history retrieved successfully")
						if interestResp != nil && interestResp.Rows != nil {
							t.Logf("Total interest records: %d", len(interestResp.Rows))
						}
					}
				})
			})
		}
	}
}

// TestMarginOrderManagement tests margin order management operations
func TestMarginOrderManagement(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Margin Order Management", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					symbol := getTestSymbol("spot")
					
					// Test getting margin trades
					tradesReq := client.PortfolioMarginAPI.GetMarginMyTradesV1(ctx).
						Symbol(symbol).
						Limit(10).
						Timestamp(generateTimestamp())
					
					tradesResp, httpResp, err := tradesReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Trades") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Get Margin Trades") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginMyTradesV1 failed: %v", err)
					} else {
						t.Logf("Margin trades retrieved successfully")
						if tradesResp != nil {
							t.Logf("Number of trades: %d", len(tradesResp))
						}
					}
					
					// Test getting margin force orders  
					// Note: GetMarginForceOrdersV1 doesn't have Symbol parameter - it gets all margin force orders for account
					forceOrdersReq := client.PortfolioMarginAPI.GetMarginForceOrdersV1(ctx).
						Size(10).
						Timestamp(generateTimestamp())
					
					forceOrdersResp, httpResp, err := forceOrdersReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Get Margin Force Orders") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetMarginForceOrdersV1 failed: %v", err)
					} else {
						t.Logf("Margin force orders retrieved successfully")
						if forceOrdersResp != nil && forceOrdersResp.Rows != nil {
							t.Logf("Number of force orders: %d", len(forceOrdersResp.Rows))
						}
					}
				})
			})
		}
	}
}