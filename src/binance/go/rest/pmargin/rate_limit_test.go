package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestRateLimitOrder tests querying user rate limit information
func TestRateLimitOrder(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Rate Limit Order", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.GetRateLimitOrderV1(ctx).
						Timestamp(generateTimestamp())
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Rate Limit Order") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Rate Limit Order") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetRateLimitOrderV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Rate limit info retrieved successfully")
					if len(resp) > 0 {
						for i, limit := range resp {
							if i >= 3 { // Limit output
								t.Logf("... and %d more rate limits", len(resp)-i)
								break
							}
							if limit.RateLimitType != nil && limit.Interval != nil {
								t.Logf("Rate limit type: %s, Interval: %s", *limit.RateLimitType, *limit.Interval)
								if limit.IntervalNum != nil && limit.Limit != nil {
									t.Logf("  Interval num: %d, Limit: %d", *limit.IntervalNum, *limit.Limit)
								}
								// Note: Count field is not available in GetRateLimitOrderV1RespItem struct
								// This endpoint shows rate limit configuration, not current usage counts
							}
						}
					}
				})
			})
		}
	}
}