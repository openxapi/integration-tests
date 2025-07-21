package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestBrokerInfo tests broker information endpoints
func TestBrokerInfo(t *testing.T) {
	// Skip by default as broker features require special account status
	if os.Getenv("BINANCE_TEST_BROKER") != "true" {
		t.Skip("Set BINANCE_TEST_BROKER=true to test broker endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetBrokerInfo", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerInfoV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker info") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if not a broker account
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2008 || code == -2014) {
						t.Skip("Not a broker account")
					}
				}
			}
			t.Fatalf("Failed to get broker info: %v", err)
		}

		t.Logf("Broker info: %+v", resp)
	})

	t.Run("GetBrokerRebateRecentRecord", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerRebateRecentRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker rebate recent record") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker rebate recent record: %v", err)
		}

		t.Logf("Broker rebate recent record: %+v", resp)
	})

	t.Run("GetBrokerRebateFuturesRecentRecord", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerRebateFuturesRecentRecordV1(ctx).
			FuturesType(1).
			StartTime(time.Now().Add(-30*24*time.Hour).UnixMilli()).
			EndTime(time.Now().UnixMilli()).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker rebate futures recent record") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker rebate futures recent record: %v", err)
		}

		t.Logf("Broker rebate futures recent record: %+v", resp)
	})
}

// TestBrokerSubAccount tests broker sub-account endpoints
func TestBrokerSubAccount(t *testing.T) {
	// Skip by default as broker features require special account status
	if os.Getenv("BINANCE_TEST_BROKER") != "true" {
		t.Skip("Set BINANCE_TEST_BROKER=true to test broker endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateBrokerSubAccount", func(t *testing.T) {
		// Skip actual creation unless explicitly enabled
		if os.Getenv("BINANCE_CREATE_BROKER_SUB_ACCOUNT") != "true" {
			t.Skip("Set BINANCE_CREATE_BROKER_SUB_ACCOUNT=true to test sub-account creation")
		}

		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to create broker sub-account: %v", err)
		}

		t.Logf("Broker sub-account created: %+v", resp)
	})

	t.Run("GetBrokerSubAccountAPI", func(t *testing.T) {
		subAccountId := os.Getenv("BINANCE_BROKER_SUB_ACCOUNT_ID")
		if subAccountId == "" {
			t.Skip("BINANCE_BROKER_SUB_ACCOUNT_ID not set")
		}

		resp, _, err := client.BinanceLinkAPI.GetBrokerSubAccountApiV1(ctx).
			SubAccountId(subAccountId).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker sub-account API: %v", err)
		}

		t.Logf("Broker sub-account API: %+v", resp)
	})

	t.Run("GetBrokerSubAccountDepositHistory", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerSubAccountDepositHistV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker sub-account deposit history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker sub-account deposit history: %v", err)
		}

		t.Logf("Broker sub-account deposit history: %+v", resp)
	})

	t.Run("GetBrokerSubAccountTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerTransferV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker sub-account transfer history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker transfer history: %v", err)
		}

		t.Logf("Broker transfer history: %+v", resp)
	})

	t.Run("GetBrokerSubAccountTransferFuturesHistory", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetBrokerTransferFuturesV1(ctx).
			FuturesType(1). // 1: USDT-M, 2: COIN-M
			SubAccountId("test123").
			StartTime(time.Now().Add(-30*24*time.Hour).UnixMilli()).
			EndTime(time.Now().UnixMilli()).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Broker sub-account transfer futures history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			logResponseBody(t, httpResp, "Get broker transfer futures history")
			t.Fatalf("Failed to get broker transfer futures history: %v", err)
		}

		t.Logf("Broker transfer futures history: %+v", resp)
	})
}

// TestBrokerCommission tests broker commission endpoints
func TestBrokerCommission(t *testing.T) {
	// Skip by default as broker features require special account status
	if os.Getenv("BINANCE_TEST_BROKER") != "true" {
		t.Skip("Set BINANCE_TEST_BROKER=true to test broker endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	subAccountId := os.Getenv("BINANCE_BROKER_SUB_ACCOUNT_ID")
	if subAccountId == "" {
		t.Skip("BINANCE_BROKER_SUB_ACCOUNT_ID not set")
	}

	t.Run("CreateBrokerSubAccountCommission", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountApiCommissionV1(ctx).
			SubAccountId(subAccountId).
			MakerCommission(100).  // 100 = 0.1%
			TakerCommission(100).  // 100 = 0.1%
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to create broker sub-account commission: %v", err)
		}

		t.Logf("Broker sub-account commission created: %+v", resp)
	})

	t.Run("GetBrokerSubAccountCommission", func(t *testing.T) {
		// Note: The SDK only has futures commission getters, not spot
		t.Skip("Spot commission getter not available in SDK")
	})

	t.Run("CreateBrokerSubAccountFuturesCommission", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountApiCommissionFuturesV1(ctx).
			SubAccountId(subAccountId).
			MakerAdjustment(10). // 10 = 0.01%
			TakerAdjustment(10). // 10 = 0.01%
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to create broker sub-account futures commission: %v", err)
		}

		t.Logf("Broker sub-account futures commission created: %+v", resp)
	})

	t.Run("GetBrokerSubAccountFuturesCommission", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.GetBrokerSubAccountApiCommissionFuturesV1(ctx).
			SubAccountId(subAccountId).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get broker sub-account futures commission: %v", err)
		}

		t.Logf("Broker sub-account futures commission: %+v", resp)
	})
}

// TestReferralOperations tests referral endpoints
func TestReferralOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetReferralIfNewUser", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetApiReferralIfNewUserV1(ctx).
			ApiAgentCode("test123").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Referral if new user") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if referral not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Referral API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to check if new user: %v", err)
		}

		t.Logf("Is new user: %+v", resp)
	})

	t.Run("GetReferralCustomization", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetApiReferralCustomizationV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Referral customization") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get referral customization: %v", err)
		}

		t.Logf("Referral customization: %+v", resp)
	})

	t.Run("GetReferralUserCustomization", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetApiReferralUserCustomizationV1(ctx).
			ApiAgentCode("test123").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Referral user customization") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get referral user customization: %v", err)
		}

		t.Logf("Referral user customization: %+v", resp)
	})

	t.Run("GetReferralRebateRecentRecord", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetApiReferralRebateRecentRecordV1(ctx).
			StartTime(time.Now().Add(-30*24*time.Hour).UnixMilli()).
			EndTime(time.Now().UnixMilli()).
			Limit(100).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Referral rebate recent record") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get referral rebate recent record: %v", err)
		}

		t.Logf("Referral rebate recent record: %+v", resp)
	})

	t.Run("GetReferralKickbackRecentRecord", func(t *testing.T) {
		resp, httpResp, err := client.BinanceLinkAPI.GetApiReferralKickbackRecentRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Referral kickback recent record") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get referral kickback recent record: %v", err)
		}

		t.Logf("Referral kickback recent record: %+v", resp)
	})
}

// TestBrokerOperations tests broker operations (use with caution)
func TestBrokerOperations(t *testing.T) {
	// Skip by default as broker operations require special account status
	if os.Getenv("BINANCE_TEST_BROKER_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_BROKER_OPERATIONS=true to test broker operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	subAccountId := os.Getenv("BINANCE_BROKER_SUB_ACCOUNT_ID")
	if subAccountId == "" {
		t.Skip("BINANCE_BROKER_SUB_ACCOUNT_ID not set")
	}

	t.Run("CreateBrokerTransfer", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerTransferV1(ctx).
			ToId(subAccountId).
			Asset("USDT").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3020 {
						t.Skip("Insufficient balance for transfer")
					}
				}
			}
			t.Fatalf("Failed to create broker transfer: %v", err)
		}

		t.Logf("Broker transfer created: %+v", resp)
	})

	t.Run("CreateBrokerUniversalTransfer", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerUniversalTransferV1(ctx).
			ToId(subAccountId).
			FromAccountType("SPOT").
			ToAccountType("MARGIN").
			Asset("USDT").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or account type not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3020 || code == -11002) {
						t.Skip("Insufficient balance or account type not enabled")
					}
				}
			}
			t.Fatalf("Failed to create broker universal transfer: %v", err)
		}

		t.Logf("Broker universal transfer created: %+v", resp)
	})

	t.Run("CreateBrokerSubAccountAPI", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountApiV1(ctx).
			SubAccountId(subAccountId).
			CanTrade("1").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to create broker sub-account API: %v", err)
		}

		t.Logf("Broker sub-account API created: %+v", resp)
	})

	t.Run("EnableBrokerSubAccountFutures", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountFuturesV1(ctx).
			SubAccountId(subAccountId).
			Futures("1"). // "1" to enable
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if already enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -2047 {
						t.Skip("Futures already enabled for sub-account")
					}
				}
			}
			t.Fatalf("Failed to enable broker sub-account futures: %v", err)
		}

		t.Logf("Broker sub-account futures enabled: %+v", resp)
	})

	t.Run("SetBrokerSubAccountBNBBurn", func(t *testing.T) {
		resp, _, err := client.BinanceLinkAPI.CreateBrokerSubAccountBnbBurnSpotV1(ctx).
			SubAccountId(subAccountId).
			SpotBNBBurn("true").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to set broker sub-account BNB burn: %v", err)
		}

		t.Logf("Broker sub-account BNB burn set: %+v", resp)
	})
}