package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestPing tests the connectivity to the Portfolio Margin API
func TestPing(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType == AuthTypeNONE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Ping", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.GetPingV1(ctx)
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Ping") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("GetPingV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Ping successful: %+v", resp)
				})
			})
			break // Only need to test ping once
		}
	}
}