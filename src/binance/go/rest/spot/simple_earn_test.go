package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestSimpleEarnFlexibleProducts tests flexible Simple Earn product endpoints
func TestSimpleEarnFlexibleProducts(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetFlexibleProductList", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible product list") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get flexible product list: %v", err)
		}

		t.Logf("Flexible products retrieved: %+v", resp)
		
		// Store first asset for other tests
		if len(resp.Rows) > 0 {
			rows := resp.Rows
			if rows[0].Asset != nil {
				testAsset := *rows[0].Asset
				t.Logf("Using asset %s for further tests", testAsset)
				ctx = context.WithValue(ctx, "testAsset", testAsset)
			}
		}
	})

	t.Run("GetFlexiblePosition", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexiblePositionV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible positions") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get flexible positions: %v", err)
		}

		t.Logf("Flexible positions: %+v", resp)
	})

	t.Run("GetFlexiblePersonalLeftQuota", func(t *testing.T) {
		// Skip if no product ID available
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexiblePersonalLeftQuotaV1(ctx).
			ProductId("USDT001").  // Example product ID
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible personal left quota") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if product not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Product not found or invalid")
					}
				}
			}
			t.Fatalf("Failed to get personal left quota: %v", err)
		}

		t.Logf("Personal left quota: %+v", resp)
	})

	t.Run("GetFlexibleSubscriptionPreview", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleSubscriptionPreviewV1(ctx).
			ProductId("USDT001").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible subscription preview") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or product not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6030) {
						t.Skip("Product not found or insufficient balance")
					}
				}
			}
			t.Fatalf("Failed to get subscription preview: %v", err)
		}

		t.Logf("Subscription preview: %+v", resp)
	})

	t.Run("GetFlexibleRateHistory", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleHistoryRateHistoryV1(ctx).
			ProductId("USDT001").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible rate history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if product not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Product not found")
					}
				}
			}
			t.Fatalf("Failed to get rate history: %v", err)
		}

		t.Logf("Rate history: %+v", resp)
	})
}

// TestSimpleEarnFlexibleHistory tests flexible Simple Earn history endpoints
func TestSimpleEarnFlexibleHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetFlexibleSubscriptionRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleHistorySubscriptionRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible subscription record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get subscription record: %v", err)
		}

		t.Logf("Subscription records: %+v", resp)
	})

	t.Run("GetFlexibleRedemptionRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleHistoryRedemptionRecordV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible redemption record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get redemption record: %v", err)
		}

		t.Logf("Redemption records: %+v", resp)
	})

	t.Run("GetFlexibleRewardsRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleHistoryRewardsRecordV1(ctx).
			Type_("BONUS").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible rewards record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get rewards record: %v", err)
		}

		t.Logf("Rewards records: %+v", resp)
	})

	t.Run("GetFlexibleCollateralRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnFlexibleHistoryCollateralRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible collateral record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get collateral record: %v", err)
		}

		t.Logf("Collateral records: %+v", resp)
	})
}

// TestSimpleEarnLockedProducts tests locked Simple Earn product endpoints
func TestSimpleEarnLockedProducts(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetLockedProductList", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked product list") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get locked product list: %v", err)
		}

		t.Logf("Locked products retrieved: %+v", resp)
	})

	t.Run("GetLockedPosition", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedPositionV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked positions") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get locked positions: %v", err)
		}

		t.Logf("Locked positions: %+v", resp)
	})

	t.Run("GetLockedPersonalLeftQuota", func(t *testing.T) {
		// Skip if no project ID available
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedPersonalLeftQuotaV1(ctx).
			ProjectId("BTC001").  // Example project ID
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked personal left quota") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if project not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Project not found or invalid")
					}
				}
			}
			t.Fatalf("Failed to get personal left quota: %v", err)
		}

		t.Logf("Personal left quota: %+v", resp)
	})

	t.Run("GetLockedSubscriptionPreview", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedSubscriptionPreviewV1(ctx).
			ProjectId("BTC001").
			Amount("0.001").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked subscription preview") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or project not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6030) {
						t.Skip("Project not found or insufficient balance")
					}
				}
			}
			t.Fatalf("Failed to get subscription preview: %v", err)
		}

		t.Logf("Subscription preview: %+v", resp)
	})
}

// TestSimpleEarnLockedHistory tests locked Simple Earn history endpoints
func TestSimpleEarnLockedHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetLockedSubscriptionRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedHistorySubscriptionRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked subscription record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get subscription record: %v", err)
		}

		t.Logf("Subscription records: %+v", resp)
	})

	t.Run("GetLockedRedemptionRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedHistoryRedemptionRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked redemption record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get redemption record: %v", err)
		}

		t.Logf("Redemption records: %+v", resp)
	})

	t.Run("GetLockedRewardsRecord", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnLockedHistoryRewardsRecordV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Locked rewards record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get rewards record: %v", err)
		}

		t.Logf("Rewards records: %+v", resp)
	})
}

// TestSimpleEarnAccount tests Simple Earn account endpoint
func TestSimpleEarnAccount(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSimpleEarnAccount", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.GetSimpleEarnAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Simple Earn account") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get Simple Earn account: %v", err)
		}

		t.Logf("Simple Earn account: %+v", resp)
	})
}

// TestSimpleEarnSubscriptionOperations tests subscription operations (use with caution)
func TestSimpleEarnSubscriptionOperations(t *testing.T) {
	// Skip by default to avoid actual subscriptions
	if os.Getenv("BINANCE_TEST_SIMPLE_EARN_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_SIMPLE_EARN_OPERATIONS=true to test subscription operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateFlexibleSubscription", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnFlexibleSubscribeV1(ctx).
			ProductId("USDT001").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create flexible subscription") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6030 {
						t.Skip("Insufficient balance for subscription")
					}
				}
			}
			t.Fatalf("Failed to create flexible subscription: %v", err)
		}

		t.Logf("Flexible subscription created: %+v", resp)
	})

	t.Run("SetFlexibleAutoSubscribe", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnFlexibleSetAutoSubscribeV1(ctx).
			ProductId("USDT001").
			AutoSubscribe(true).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Set flexible auto subscribe") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to set auto subscribe: %v", err)
		}

		t.Logf("Auto subscribe set: %+v", resp)
	})

	t.Run("CreateFlexibleRedemption", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnFlexibleRedeemV1(ctx).
			ProductId("USDT001").
			RedeemAll(true).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create flexible redemption") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no position to redeem
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6024 {
						t.Skip("No position to redeem")
					}
				}
			}
			t.Fatalf("Failed to create flexible redemption: %v", err)
		}

		t.Logf("Flexible redemption created: %+v", resp)
	})

	t.Run("CreateLockedSubscription", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnLockedSubscribeV1(ctx).
			ProjectId("BTC001").
			Amount("0.001").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create locked subscription") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or project not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -6030 || code == -1121) {
						t.Skip("Insufficient balance or project not available")
					}
				}
			}
			t.Fatalf("Failed to create locked subscription: %v", err)
		}

		t.Logf("Locked subscription created: %+v", resp)
	})

	t.Run("SetLockedAutoSubscribe", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnLockedSetAutoSubscribeV1(ctx).
			PositionId(12345).
			AutoSubscribe(true).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Set locked auto subscribe") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if position not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6024 {
						t.Skip("Position not found")
					}
				}
			}
			t.Fatalf("Failed to set auto subscribe: %v", err)
		}

		t.Logf("Auto subscribe set: %+v", resp)
	})

	t.Run("SetLockedRedeemOption", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnLockedSetRedeemOptionV1(ctx).
			PositionId("12345").
			RedeemTo("SPOT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Set locked redeem option") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if position not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6024 {
						t.Skip("Position not found")
					}
				}
			}
			t.Fatalf("Failed to set redeem option: %v", err)
		}

		t.Logf("Redeem option set: %+v", resp)
	})

	t.Run("CreateLockedRedemption", func(t *testing.T) {
		resp, httpResp, err := client.SimpleEarnAPI.CreateSimpleEarnLockedRedeemV1(ctx).
			PositionId(12345).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create locked redemption") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if position not found or not redeemable
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -6024 || code == -6028) {
						t.Skip("Position not found or not redeemable")
					}
				}
			}
			t.Fatalf("Failed to create locked redemption: %v", err)
		}

		t.Logf("Locked redemption created: %+v", resp)
	})
}