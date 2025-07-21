package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// Note: getTestClient and parseJSON are defined in integration_test.go

// TestSubAccountManagement tests sub-account management endpoints
func TestSubAccountManagement(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("ListSubAccounts", func(t *testing.T) {
		// Note: This endpoint requires master account permissions
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied (code -2015 or -4001)
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to list sub-accounts: %v", err)
		}

		t.Logf("Sub-accounts list retrieved: %+v", resp)
	})

	t.Run("GetSubAccountStatusV2", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountStatusV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account status") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account status: %v", err)
		}

		t.Logf("Sub-account status: %+v", resp)
	})

	t.Run("GetSubAccountApiIpRestrictionV1", func(t *testing.T) {
		// This requires a sub-account email
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		if subEmail == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetSubAccountSubAccountApiIpRestrictionV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account API IP restriction") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account API IP restriction: %v", err)
		}

		t.Logf("Sub-account API IP restriction: %+v", resp)
	})

	t.Run("DeleteSubAccountApiIpRestriction", func(t *testing.T) {
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		subApiKey := os.Getenv("BINANCE_SUB_ACCOUNT_API_KEY")
		if subEmail == "" || subApiKey == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL or BINANCE_SUB_ACCOUNT_API_KEY not set")
		}

		resp, httpResp, err := client.SubAccountAPI.DeleteSubAccountSubAccountApiIpRestrictionIpListV1(ctx).
			Email(subEmail).
			SubAccountApiKey(subApiKey).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Delete sub-account API IP restriction") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or no IP restriction to delete
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -2023) {
						t.Skip("Sub-account endpoints require master account permissions or no IP restriction exists")
					}
				}
			}
			t.Fatalf("Failed to delete sub-account API IP restriction: %v", err)
		}

		t.Logf("Sub-account API IP restriction deleted: %+v", resp)
	})
}

// TestSubAccountAssets tests sub-account asset endpoints
func TestSubAccountAssets(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSubAccountAssets", func(t *testing.T) {
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		if subEmail == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetSubAccountAssetsV3(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account assets") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account assets: %v", err)
		}

		t.Logf("Sub-account assets: %+v", resp)
	})

	t.Run("GetSubAccountSpotAssetSummary", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountSpotSummaryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account spot summary") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account spot summary: %v", err)
		}

		t.Logf("Sub-account spot summary: %+v", resp)
	})

	t.Run("GetManagedSubAccountAssetDetails", func(t *testing.T) {
		managedEmail := os.Getenv("BINANCE_MANAGED_SUB_EMAIL")
		if managedEmail == "" {
			t.Skip("BINANCE_MANAGED_SUB_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetManagedSubaccountAssetV1(ctx).
			Email(managedEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Managed sub-account assets") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -2008) {
						t.Skip("Managed sub-account endpoints require investor account permissions")
					}
				}
			}
			t.Fatalf("Failed to get managed sub-account assets: %v", err)
		}

		t.Logf("Managed sub-account assets: %+v", resp)
	})
}

// TestSubAccountTransferHistory tests sub-account transfer history endpoints
func TestSubAccountTransferHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSubAccountTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountTransferSubUserHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account transfer history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account transfer history: %v", err)
		}

		t.Logf("Sub-account transfer history: %+v", resp)
	})

	t.Run("GetSubAccountUniversalTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountUniversalTransferV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account universal transfer history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account universal transfer history: %v", err)
		}

		t.Logf("Sub-account universal transfer history: %+v", resp)
	})

	t.Run("GetManagedSubAccountSnapshot", func(t *testing.T) {
		managedEmail := os.Getenv("BINANCE_MANAGED_SUB_EMAIL")
		if managedEmail == "" {
			t.Skip("BINANCE_MANAGED_SUB_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetManagedSubaccountAccountSnapshotV1(ctx).
			Email(managedEmail).
			Type_("SPOT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Managed sub-account snapshot") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -2008) {
						t.Skip("Managed sub-account endpoints require investor account permissions")
					}
				}
			}
			t.Fatalf("Failed to get managed sub-account snapshot: %v", err)
		}

		t.Logf("Managed sub-account snapshot: %+v", resp)
	})
}

// TestSubAccountMarginFutures tests sub-account margin and futures endpoints
func TestSubAccountMarginFutures(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSubAccountMarginAccountDetail", func(t *testing.T) {
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		if subEmail == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetSubAccountMarginAccountV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account margin account detail") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -11002) {
						t.Skip("Sub-account endpoints require master account permissions or margin not enabled")
					}
				}
			}
			t.Fatalf("Failed to get sub-account margin account detail: %v", err)
		}

		t.Logf("Sub-account margin account detail: %+v", resp)
	})

	t.Run("GetSubAccountMarginAccountSummary", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountMarginAccountSummaryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account margin account summary") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account margin account summary: %v", err)
		}

		t.Logf("Sub-account margin account summary: %+v", resp)
	})

	t.Run("GetSubAccountFuturesAccountDetail", func(t *testing.T) {
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		if subEmail == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetSubAccountFuturesAccountV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account futures account detail") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -11002) {
						t.Skip("Sub-account endpoints require master account permissions or futures not enabled")
					}
				}
			}
			t.Fatalf("Failed to get sub-account futures account detail: %v", err)
		}

		t.Logf("Sub-account futures account detail: %+v", resp)
	})

	t.Run("GetSubAccountFuturesAccountSummary", func(t *testing.T) {
		resp, httpResp, err := client.SubAccountAPI.GetSubAccountFuturesAccountSummaryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account futures account summary") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account endpoints require master account permissions")
					}
				}
			}
			t.Fatalf("Failed to get sub-account futures account summary: %v", err)
		}

		t.Logf("Sub-account futures account summary: %+v", resp)
	})

	t.Run("GetSubAccountFuturesPositionRisk", func(t *testing.T) {
		subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
		if subEmail == "" {
			t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
		}

		resp, httpResp, err := client.SubAccountAPI.GetSubAccountFuturesPositionRiskV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Sub-account futures position risk") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or futures not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -11002) {
						t.Skip("Sub-account endpoints require master account permissions or futures not enabled")
					}
				}
			}
			t.Fatalf("Failed to get sub-account futures position risk: %v", err)
		}

		t.Logf("Sub-account futures position risk: %+v", resp)
	})
}

// TestSubAccountCreate tests sub-account creation (use with caution)
func TestSubAccountCreate(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateVirtualSubAccount", func(t *testing.T) {
		// Skip by default to avoid creating test accounts
		if os.Getenv("BINANCE_CREATE_TEST_SUB_ACCOUNT") != "true" {
			t.Skip("Set BINANCE_CREATE_TEST_SUB_ACCOUNT=true to test sub-account creation")
		}

		subAccountString := fmt.Sprintf("test%d", time.Now().Unix())
		resp, httpResp, err := client.SubAccountAPI.CreateSubAccountVirtualSubAccountV1(ctx).
			SubAccountString(subAccountString).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create virtual sub-account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001) {
						t.Skip("Sub-account creation requires master account permissions")
					}
				}
			}
			t.Fatalf("Failed to create virtual sub-account: %v", err)
		}

		t.Logf("Virtual sub-account created: %+v", resp)
	})
}

// TestSubAccountEnableFeatures tests enabling features for sub-accounts
func TestSubAccountEnableFeatures(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	subEmail := os.Getenv("BINANCE_SUB_ACCOUNT_EMAIL")
	if subEmail == "" {
		t.Skip("BINANCE_SUB_ACCOUNT_EMAIL not set")
	}

	t.Run("EnableMarginForSubAccount", func(t *testing.T) {
		// Skip if already enabled
		resp, httpResp, err := client.SubAccountAPI.CreateSubAccountMarginEnableV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Enable margin for sub-account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or already enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -11000) {
						t.Skip("Sub-account endpoints require master account permissions or margin already enabled")
					}
				}
			}
			t.Fatalf("Failed to enable margin for sub-account: %v", err)
		}

		t.Logf("Margin enabled for sub-account: %+v", resp)
	})

	t.Run("EnableFuturesForSubAccount", func(t *testing.T) {
		// Skip if already enabled
		resp, httpResp, err := client.SubAccountAPI.CreateSubAccountFuturesEnableV1(ctx).
			Email(subEmail).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Enable futures for sub-account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied or already enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2015 || code == -4001 || code == -11000) {
						t.Skip("Sub-account endpoints require master account permissions or futures already enabled")
					}
				}
			}
			t.Fatalf("Failed to enable futures for sub-account: %v", err)
		}

		t.Logf("Futures enabled for sub-account: %+v", resp)
	})
}

