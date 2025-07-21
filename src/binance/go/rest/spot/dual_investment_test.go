package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestDualInvestmentInfo tests dual investment information endpoints
func TestDualInvestmentInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetDualInvestmentProductList", func(t *testing.T) {
		resp, httpResp, err := client.DualInvestmentAPI.GetDciProductListV1(ctx).
			OptionType("CALL"). // CALL or PUT
			ExercisedCoin("BTC").
			InvestCoin("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Dual investment product list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if dual investment not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -14000) {
						t.Skip("Dual investment not available on testnet")
					}
				}
			}
			logResponseBody(t, httpResp, "Get dual investment product list")
			t.Fatalf("Failed to get dual investment product list: %v", err)
		}

		t.Logf("Dual investment product list: %+v", resp)
	})

	t.Run("GetDualInvestmentAccounts", func(t *testing.T) {
		resp, httpResp, err := client.DualInvestmentAPI.GetDciProductAccountsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Dual investment accounts") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get dual investment accounts: %v", err)
		}

		t.Logf("Dual investment accounts: %+v", resp)
	})

	t.Run("GetDualInvestmentPositions", func(t *testing.T) {
		resp, httpResp, err := client.DualInvestmentAPI.GetDciProductPositionsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Dual investment positions") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get dual investment positions: %v", err)
		}

		t.Logf("Dual investment positions: %+v", resp)
	})
}

// TestDualInvestmentOperations tests dual investment operations (use with caution)
func TestDualInvestmentOperations(t *testing.T) {
	// Skip by default to avoid creating actual dual investment positions
	if os.Getenv("BINANCE_TEST_DUAL_INVESTMENT_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_DUAL_INVESTMENT_OPERATIONS=true to test dual investment operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("SubscribeDualInvestmentProduct", func(t *testing.T) {
		// First get available products
		listResp, _, err := client.DualInvestmentAPI.GetDciProductListV1(ctx).
			OptionType("CALL").
			Timestamp(timestamp).
			Execute()

		if err != nil || len(listResp.List) == 0 {
			t.Skip("No dual investment products available")
		}

		// Use the first available product
		firstProduct := listResp.List[0]
		
		var productId string
		if firstProduct.Id != nil {
			productId = *firstProduct.Id
		} else {
			t.Skip("No product ID available")
		}

		resp, _, err := client.DualInvestmentAPI.CreateDciProductSubscribeV1(ctx).
			Id(productId).
			DepositAmount("10"). // Subscribe with 10 USDT
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or product not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -14001 || code == -14002) {
						t.Skip("Insufficient balance or product not available for subscription")
					}
				}
			}
			t.Fatalf("Failed to subscribe to dual investment product: %v", err)
		}

		t.Logf("Dual investment subscription created: %+v", resp)
	})

	t.Run("EditAutoCompoundStatus", func(t *testing.T) {
		// This requires an active position
		positionId := os.Getenv("BINANCE_TEST_DUAL_INVESTMENT_POSITION_ID")
		if positionId == "" {
			t.Skip("BINANCE_TEST_DUAL_INVESTMENT_POSITION_ID not set")
		}

		resp, _, err := client.DualInvestmentAPI.CreateDciProductAutoCompoundEditStatusV1(ctx).
			PositionId(positionId).
			AutoCompoundPlan("STANDARD"). // Enable auto-compound
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if position not found or not eligible
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -14003 || code == -14004) {
						t.Skip("Position not found or not eligible for auto-compound")
					}
				}
			}
			t.Fatalf("Failed to edit auto-compound status: %v", err)
		}

		t.Logf("Auto-compound status updated: %+v", resp)
	})
}