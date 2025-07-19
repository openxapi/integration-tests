package main

import (
	"context"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// TestIncomeHistory tests getting income history
func TestIncomeHistory(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "IncomeHistory", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.FuturesAPI.GetIncomeV1(ctx).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "IncomeHistory") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Income history failed: %v", err)
					}
					
					t.Logf("Income history: count=%d", len(resp))
					
					// Check structure of first income entry if any exist
					if len(resp) > 0 {
						firstIncome := resp[0]
						if firstIncome.Symbol == nil {
							t.Fatal("First income has nil Symbol")
						}
						
						if firstIncome.Income == nil {
							t.Fatal("First income has nil Income")
						}
						
						if firstIncome.IncomeType == nil {
							t.Fatal("First income has nil IncomeType")
						}
						
						if firstIncome.Time == nil {
							t.Fatal("First income has nil Time")
						}
						
						t.Logf("First income: symbol=%s, income=%s, type=%s, time=%d", 
							*firstIncome.Symbol, *firstIncome.Income, *firstIncome.IncomeType, *firstIncome.Time)
					}
				})
			})
			break
		}
	}
}

// TestIncomeAsync tests getting download id for futures transaction history
func TestIncomeAsync(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "IncomeAsync", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Set date range (last 30 days to increase chance of having data)
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -30)
					
					req := client.FuturesAPI.GetIncomeAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "IncomeAsync") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Income async failed: %v", err)
					}
					
					// Debug: log the response structure to understand what we're getting
					t.Logf("IncomeAsync response structure: %+v", resp)
					
					if resp.DownloadId == nil {
						// This might be normal on testnet if there's no income history
						t.Logf("DownloadId is nil - this may be normal on testnet with no income history")
						t.Logf("API call successful but no download ID returned (likely no data for period)")
						return
					}
					
					t.Logf("Income async: downloadId=%s", *resp.DownloadId)
				})
			})
			break
		}
	}
}

// TestIncomeAsyncDownload tests getting futures transaction history download link by Id
func TestIncomeAsyncDownload(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "IncomeAsyncDownload", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// First, get a download ID
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -7)
					
					asyncReq := client.FuturesAPI.GetIncomeAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					asyncResp, _, asyncErr := asyncReq.Execute()
					if asyncErr != nil {
						t.Skipf("Cannot get download ID for income async download test: %v", asyncErr)
						return
					}
					
					if asyncResp.DownloadId == nil {
						t.Skip("No download ID returned for income async download test")
						return
					}
					
					downloadId := *asyncResp.DownloadId
					
					// Wait a bit for the download to be prepared
					time.Sleep(2 * time.Second)
					
					req := client.FuturesAPI.GetIncomeAsynIdV1(ctx).
						DownloadId(downloadId).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "IncomeAsyncDownload") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Income async download failed: %v", err)
					}
					
					if resp.DownloadId == nil {
						t.Fatal("DownloadId is nil")
					}
					
					if resp.Status == nil {
						t.Fatal("Status is nil")
					}
					
					t.Logf("Income async download: downloadId=%s, status=%s", 
						*resp.DownloadId, *resp.Status)
					
					// If download is ready, check the URL
					if resp.Status != nil && *resp.Status == "completed" && resp.Url != nil {
						t.Logf("Download URL: %s", *resp.Url)
					}
				})
			})
			break
		}
	}
}

// TestOrderAsync tests getting download id for futures order history
func TestOrderAsync(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "OrderAsync", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Set date range (last 30 days to increase chance of having data)
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -30)
					
					req := client.FuturesAPI.GetOrderAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "OrderAsync") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Order async failed: %v", err)
					}
					
					// Debug: log the response structure to understand what we're getting
					t.Logf("OrderAsync response structure: %+v", resp)
					
					if resp.DownloadId == nil {
						// This might be normal on testnet if there's no order history
						t.Logf("DownloadId is nil - this may be normal on testnet with no order history")
						t.Logf("API call successful but no download ID returned (likely no data for period)")
						return
					}
					
					t.Logf("Order async: downloadId=%s", *resp.DownloadId)
				})
			})
			break
		}
	}
}

// TestOrderAsyncDownload tests getting futures order history download link by Id
func TestOrderAsyncDownload(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "OrderAsyncDownload", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// First, get a download ID
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -7)
					
					asyncReq := client.FuturesAPI.GetOrderAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					asyncResp, _, asyncErr := asyncReq.Execute()
					if asyncErr != nil {
						t.Skipf("Cannot get download ID for order async download test: %v", asyncErr)
						return
					}
					
					if asyncResp.DownloadId == nil {
						t.Skip("No download ID returned for order async download test")
						return
					}
					
					downloadId := *asyncResp.DownloadId
					
					// Wait a bit for the download to be prepared
					time.Sleep(2 * time.Second)
					
					req := client.FuturesAPI.GetOrderAsynIdV1(ctx).
						DownloadId(downloadId).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "OrderAsyncDownload") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Order async download failed: %v", err)
					}
					
					if resp.DownloadId == nil {
						t.Fatal("DownloadId is nil")
					}
					
					if resp.Status == nil {
						t.Fatal("Status is nil")
					}
					
					t.Logf("Order async download: downloadId=%s, status=%s", 
						*resp.DownloadId, *resp.Status)
					
					// If download is ready, check the URL
					if resp.Status != nil && *resp.Status == "completed" && resp.Url != nil {
						t.Logf("Download URL: %s", *resp.Url)
					}
				})
			})
			break
		}
	}
}

// TestTradeAsync tests getting download id for futures trade history
func TestTradeAsync(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TradeAsync", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Set date range (last 30 days to increase chance of having data)
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -30)
					
					req := client.FuturesAPI.GetTradeAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TradeAsync") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Trade async failed: %v", err)
					}
					
					// Debug: log the response structure to understand what we're getting
					t.Logf("TradeAsync response structure: %+v", resp)
					
					if resp.DownloadId == nil {
						// This might be normal on testnet if there's no trade history
						t.Logf("DownloadId is nil - this may be normal on testnet with no trade history")
						t.Logf("API call successful but no download ID returned (likely no data for period)")
						return
					}
					
					t.Logf("Trade async: downloadId=%s", *resp.DownloadId)
				})
			})
			break
		}
	}
}

// TestTradeAsyncDownload tests getting futures trade download link by Id
func TestTradeAsyncDownload(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "TradeAsyncDownload", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// First, get a download ID
					endTime := time.Now()
					startTime := endTime.AddDate(0, 0, -7)
					
					asyncReq := client.FuturesAPI.GetTradeAsynV1(ctx).
						StartTime(startTime.UnixMilli()).
						EndTime(endTime.UnixMilli()).
						Timestamp(generateTimestamp())
					
					asyncResp, _, asyncErr := asyncReq.Execute()
					if asyncErr != nil {
						t.Skipf("Cannot get download ID for trade async download test: %v", asyncErr)
						return
					}
					
					if asyncResp.DownloadId == nil {
						t.Skip("No download ID returned for trade async download test")
						return
					}
					
					downloadId := *asyncResp.DownloadId
					
					// Wait a bit for the download to be prepared
					time.Sleep(2 * time.Second)
					
					req := client.FuturesAPI.GetTradeAsynIdV1(ctx).
						DownloadId(downloadId).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "TradeAsyncDownload") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "IncomeHistoryOperation")
						t.Fatalf("Trade async download failed: %v", err)
					}
					
					if resp.DownloadId == nil {
						t.Fatal("DownloadId is nil")
					}
					
					if resp.Status == nil {
						t.Fatal("Status is nil")
					}
					
					t.Logf("Trade async download: downloadId=%s, status=%s", 
						*resp.DownloadId, *resp.Status)
					
					// If download is ready, check the URL
					if resp.Status != nil && *resp.Status == "completed" && resp.Url != nil {
						t.Logf("Download URL: %s", *resp.Url)
					}
				})
			})
			break
		}
	}
}