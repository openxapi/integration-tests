package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// getSorSupportedSymbol gets a SOR-supported symbol from exchange info, or returns empty string if none available
func getSorSupportedSymbol(client *openapi.APIClient, ctx context.Context) (string, error) {
	// Create a new unauthenticated client for the public exchange info endpoint
	// GetExchangeInfoV3 is a public endpoint and doesn't accept authentication parameters
	publicConfig := openapi.NewConfiguration()
	publicConfig.Servers[0].URL = client.GetConfig().Servers[0].URL // Use same base URL
	publicClient := openapi.NewAPIClient(publicConfig)
	
	// Use a fresh context without authentication parameters
	publicCtx := context.Background()
	req := publicClient.SpotTradingAPI.GetExchangeInfoV3(publicCtx)
	resp, httpResp, err := req.Execute()
	if err != nil {
		// Log detailed error information for debugging
		if httpResp != nil {
			if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
				body := string(apiErr.Body())
				return "", fmt.Errorf("failed to get exchange info (status %d): %v - Response: %s", httpResp.StatusCode, err, body)
			}
			return "", fmt.Errorf("failed to get exchange info (status %d): %v", httpResp.StatusCode, err)
		}
		return "", fmt.Errorf("failed to get exchange info: %v", err)
	}
	
	// Check if SOR is available
	if !resp.HasSors() || len(resp.GetSors()) == 0 {
		return "", fmt.Errorf("no SOR configurations available")
	}
	
	// Get the first available SOR symbol
	sors := resp.GetSors()
	for _, sor := range sors {
		if sor.HasSymbols() && len(sor.GetSymbols()) > 0 {
			return sor.GetSymbols()[0], nil
		}
	}
	
	return "", fmt.Errorf("no symbols found in SOR configurations")
}

// TestCreateSorOrder tests creating a SOR (Smart Order Routing) order
func TestCreateSorOrder(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateSorOrder", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get a SOR-supported symbol
				sorSymbol, err := getSorSupportedSymbol(client, ctx)
				if err != nil {
					// Check if this is a real API error (400/500) or just no SOR support
					if strings.Contains(err.Error(), "status 400") || strings.Contains(err.Error(), "status 500") {
						t.Fatalf("Failed to get exchange info: %v", err)
					}
					// Skip only if SOR is genuinely not configured
					t.Skipf("No SOR-supported symbols available: %v", err)
				}
				
				t.Logf("Using SOR-supported symbol: %s", sorSymbol)
				
				// Get current price
				price, err := getCurrentPrice(client, ctx, sorSymbol)
				if err != nil {
					t.Fatalf("Failed to get current price for %s: %v", sorSymbol, err)
				}
				
				// Set limit price slightly below market for buy order
				limitPrice := price * 0.99
				limitPriceStr := fmt.Sprintf("%.2f", limitPrice)
				
				req := client.SpotTradingAPI.CreateSorOrderV3(ctx).
					Symbol(sorSymbol).
					Side("BUY").
					Type_("LIMIT").
					TimeInForce("GTC").
					Quantity("0.001").
					Price(limitPriceStr).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					// This will log response body for 400 errors automatically
					checkAPIErrorWithResponse(t, err, httpResp, "Create SOR order")
					
					// Check for specific SOR-related errors that should be skipped
					if httpResp != nil {
						if httpResp.StatusCode == 404 {
							t.Skip("SOR orders not available on this account/testnet")
						}
						// Handle -1013 error: "This symbol has no SOR"
						if httpResp.StatusCode == 400 {
							if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
								body := string(apiErr.Body())
								if strings.Contains(body, "This symbol has no SOR") || strings.Contains(body, "-1013") {
									t.Skip("Symbol does not support SOR - this is expected for symbols not in SOR configuration")
								}
							}
						}
					}
					t.Fatalf("Failed to create SOR order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Verify response
				if resp.OrderId == nil || *resp.OrderId == 0 {
					t.Error("Expected order ID in response")
				}
				
				if resp.Symbol == nil || *resp.Symbol != sorSymbol {
					t.Errorf("Expected symbol %s, got %v", sorSymbol, resp.Symbol)
				}
				
				if resp.Status == nil {
					t.Error("Expected order status in response")
				}
				
				if resp.WorkingFloor != nil {
					t.Logf("SOR order routed to: %s", *resp.WorkingFloor)
				}
				
				// Cancel the order to clean up
				if resp.OrderId != nil {
					rateLimiter.WaitForRateLimit()
					cancelReq := client.SpotTradingAPI.DeleteOrderV3(ctx).
						Symbol(sorSymbol).
						OrderId(*resp.OrderId).
						Timestamp(generateTimestamp()).
						RecvWindow(5000)
					
					_, _, cancelErr := cancelReq.Execute()
					if cancelErr != nil {
						t.Logf("Warning: Failed to cancel SOR order: %v", cancelErr)
					}
				}
			})
		})
	}
}

// TestCreateSorOrderTest tests the SOR test order endpoint
func TestCreateSorOrderTest(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType != AuthTypeTRADE {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "CreateSorOrderTest", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get a SOR-supported symbol
				sorSymbol, err := getSorSupportedSymbol(client, ctx)
				if err != nil {
					// Check if this is a real API error (400/500) or just no SOR support
					if strings.Contains(err.Error(), "status 400") || strings.Contains(err.Error(), "status 500") {
						t.Fatalf("Failed to get exchange info: %v", err)
					}
					// Skip only if SOR is genuinely not configured
					t.Skipf("No SOR-supported symbols available: %v", err)
				}
				
				t.Logf("Using SOR-supported symbol: %s", sorSymbol)
				
				// Get current price
				price, err := getCurrentPrice(client, ctx, sorSymbol)
				if err != nil {
					t.Fatalf("Failed to get current price for %s: %v", sorSymbol, err)
				}
				
				limitPrice := price * 0.99
				limitPriceStr := fmt.Sprintf("%.2f", limitPrice)
				
				req := client.SpotTradingAPI.CreateSorOrderTestV3(ctx).
					Symbol(sorSymbol).
					Side("BUY").
					Type_("LIMIT").
					TimeInForce("GTC").
					Quantity("0.001").
					Price(limitPriceStr).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					// This will log response body for 400 errors automatically
					checkAPIErrorWithResponse(t, err, httpResp, "Create SOR test order")
					
					// Check for specific SOR-related errors that should be skipped
					if httpResp != nil {
						if httpResp.StatusCode == 404 {
							t.Skip("SOR test orders not available on this account/testnet")
						}
						// Handle -1013 error: "This symbol has no SOR"
						if httpResp.StatusCode == 400 {
							if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
								body := string(apiErr.Body())
								if strings.Contains(body, "This symbol has no SOR") || strings.Contains(body, "-1013") {
									t.Skip("Symbol does not support SOR - this is expected for symbols not in SOR configuration")
								}
							}
						}
					}
					t.Fatalf("Failed to create SOR test order: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Test order typically returns empty response or basic validation result
				_ = resp
				t.Log("SOR test order validated successfully")
			})
		})
	}
}

// TestGetMyAllocations tests retrieving SOR allocations
func TestGetMyAllocations(t *testing.T) {
	for _, config := range getTestConfigs() {
		if config.AuthType < AuthTypeUSER_DATA {
			continue
		}
		
		t.Run(config.Name, func(t *testing.T) {
			testEndpoint(t, config, "GetMyAllocations", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
				// Get a SOR-supported symbol
				sorSymbol, err := getSorSupportedSymbol(client, ctx)
				if err != nil {
					// Check if this is a real API error (400/500) or just no SOR support
					if strings.Contains(err.Error(), "status 400") || strings.Contains(err.Error(), "status 500") {
						t.Fatalf("Failed to get exchange info: %v", err)
					}
					// Skip only if SOR is genuinely not configured
					t.Skipf("No SOR-supported symbols available: %v", err)
				}
				
				t.Logf("Using SOR-supported symbol: %s", sorSymbol)
				
				req := client.SpotTradingAPI.GetMyAllocationsV3(ctx).
					Symbol(sorSymbol).
					Limit(10).
					Timestamp(generateTimestamp()).
					RecvWindow(5000)
				
				resp, httpResp, err := req.Execute()
				if err != nil {
					checkAPIError(t, err)
					
					// Check for specific SOR-related errors that should be skipped
					if httpResp != nil {
						if httpResp.StatusCode == 404 {
							t.Skip("SOR allocations not available on this account")
						}
						// Handle -1013 error: "This symbol has no SOR"
						if httpResp.StatusCode == 400 {
							if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
								body := string(apiErr.Body())
								if strings.Contains(body, "This symbol has no SOR") || strings.Contains(body, "-1013") {
									t.Skip("Symbol does not support SOR - this is expected for symbols not in SOR configuration")
								}
							}
						}
					}
					t.Fatalf("Failed to get allocations: %v", err)
				}
				
				if httpResp.StatusCode != 200 {
					t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
				}
				
				// Response can be empty if no SOR allocations
				t.Logf("Found %d SOR allocations", len(resp))
				
				// If there are allocations, verify structure
				if len(resp) > 0 {
					allocation := resp[0]
					if allocation.Symbol == nil || *allocation.Symbol != sorSymbol {
						t.Errorf("Expected symbol %s, got %v", sorSymbol, allocation.Symbol)
					}
					if allocation.AllocationId == nil || *allocation.AllocationId == 0 {
						t.Error("Expected allocation ID")
					}
					if allocation.OrderId == nil || *allocation.OrderId == 0 {
						t.Error("Expected order ID")
					}
					if allocation.Qty == nil || *allocation.Qty == "" {
						t.Error("Expected allocation quantity")
					}
					if allocation.AllocationType == nil || *allocation.AllocationType == "" {
						t.Error("Expected allocation type")
					}
				}
			})
		})
	}
}