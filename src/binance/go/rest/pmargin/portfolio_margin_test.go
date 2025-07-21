package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestPortfolioInterestHistory tests querying portfolio margin interest history
func TestPortfolioInterestHistory(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Portfolio Interest History", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.GetPortfolioInterestHistoryV1(ctx).
						Timestamp(generateTimestamp())
					// Optional: add time range filters
					// req = req.StartTime(time.Now().Add(-24*time.Hour).UnixMilli())
					// req = req.EndTime(time.Now().UnixMilli())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Portfolio Interest History") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Portfolio Interest History") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetPortfolioInterestHistoryV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Portfolio interest history retrieved successfully")
					if len(resp) > 0 {
						for i, interest := range resp {
							if i >= 5 { // Limit output
								t.Logf("... and %d more interest records", len(resp)-i)
								break
							}
							if interest.Asset != nil && interest.Interest != nil {
								t.Logf("Asset: %s, Interest: %s", *interest.Asset, *interest.Interest)
								if interest.InterestAccuredTime != nil {
									t.Logf("  Time: %d", *interest.InterestAccuredTime)
								}
							}
						}
					} else {
						t.Logf("No interest history found")
					}
				})
			})
		}
	}
}

// TestPortfolioNegativeBalanceExchange tests querying negative balance exchange record
func TestPortfolioNegativeBalanceExchange(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Portfolio Negative Balance Exchange", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// This API requires startTime and endTime parameters - get records from last 30 days
					currentTime := generateTimestamp()
					startTime := currentTime - (30 * 24 * 60 * 60 * 1000) // 30 days ago in milliseconds
					
					req := client.PortfolioMarginAPI.GetPortfolioNegativeBalanceExchangeRecordV1(ctx).
						StartTime(startTime).
						EndTime(currentTime).
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Portfolio Negative Balance Exchange") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Portfolio Negative Balance Exchange") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetPortfolioNegativeBalanceExchangeRecordV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Portfolio negative balance exchange record retrieved successfully")
					if resp.Rows != nil && len(resp.Rows) > 0 {
						for i, record := range resp.Rows {
							if i >= 5 { // Limit output
								t.Logf("... and %d more exchange records", len(resp.Rows)-i)
								break
							}
							
							// Log time range for this record
							if record.StartTime != nil && record.EndTime != nil {
								t.Logf("Exchange record %d: Start time: %d, End time: %d", i+1, *record.StartTime, *record.EndTime)
							}
							
							// The actual asset details are in the Details array
							if record.Details != nil && len(record.Details) > 0 {
								t.Logf("  Found %d asset details:", len(record.Details))
								for j, detail := range record.Details {
									if j >= 3 { // Limit detail output
										t.Logf("    ... and %d more asset details", len(record.Details)-j)
										break
									}
									if detail.Asset != nil {
										t.Logf("    Asset: %s", *detail.Asset)
										if detail.NegativeBalance != nil {
											t.Logf("      Negative Balance: %d", *detail.NegativeBalance)
										}
										if detail.NegativeMaxThreshold != nil {
											t.Logf("      Negative Max Threshold: %d", *detail.NegativeMaxThreshold)
										}
									}
								}
							}
						}
					} else {
						t.Logf("No negative balance exchange records found")
					}
				})
			})
		}
	}
}