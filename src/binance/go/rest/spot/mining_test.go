package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestMiningPublicInfo tests mining public information endpoints
func TestMiningPublicInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	t.Run("GetMiningAlgoList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningPubAlgoListV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining algo list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if mining not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -12000) {
						t.Skip("Mining API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get mining algo list: %v", err)
		}

		t.Logf("Mining algo list: %+v", resp)
	})

	t.Run("GetMiningCoinList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningPubCoinListV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining coin list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if mining not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -12000) {
						t.Skip("Mining API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get mining coin list: %v", err)
		}

		t.Logf("Mining coin list: %+v", resp)
	})
}

// TestMiningUserData tests mining user data endpoints
func TestMiningUserData(t *testing.T) {
	// Skip by default as mining requires special setup
	if os.Getenv("BINANCE_TEST_MINING") != "true" {
		t.Skip("Set BINANCE_TEST_MINING=true to test mining endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetMiningUserStatus", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningStatisticsUserStatusV1(ctx).
			Algo("sha256").
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining user status") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if mining not enabled for account
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -12001 || code == -12002) {
						t.Skip("Mining not enabled for this account")
					}
				}
			}
			t.Fatalf("Failed to get mining user status: %v", err)
		}

		t.Logf("Mining user status: %+v", resp)
	})

	t.Run("GetMiningUserList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningStatisticsUserListV1(ctx).
			Algo("sha256").
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining user list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if mining not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -12001 || code == -12002) {
						t.Skip("Mining not enabled for this account")
					}
				}
			}
			t.Fatalf("Failed to get mining user list: %v", err)
		}

		t.Logf("Mining user list: %+v", resp)
	})

	t.Run("GetMiningWorkerList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningWorkerListV1(ctx).
			Algo("sha256").
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining worker list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get mining worker list: %v", err)
		}

		t.Logf("Mining worker list: %+v", resp)
	})

	t.Run("GetMiningWorkerDetail", func(t *testing.T) {
		// This requires a worker name
		workerName := os.Getenv("BINANCE_MINING_WORKER_NAME")
		if workerName == "" {
			t.Skip("BINANCE_MINING_WORKER_NAME not set")
		}

		resp, _, err := client.MiningAPI.GetMiningWorkerDetailV1(ctx).
			Algo("sha256").
			WorkerName(workerName).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get mining worker detail: %v", err)
		}

		t.Logf("Mining worker detail: %+v", resp)
	})
}

// TestMiningPayments tests mining payment endpoints
func TestMiningPayments(t *testing.T) {
	// Skip by default as mining requires special setup
	if os.Getenv("BINANCE_TEST_MINING") != "true" {
		t.Skip("Set BINANCE_TEST_MINING=true to test mining endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetMiningPaymentList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningPaymentListV1(ctx).
			Algo("sha256").
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining payment list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get mining payment list: %v", err)
		}

		t.Logf("Mining payment list: %+v", resp)
	})

	t.Run("GetMiningPaymentOther", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningPaymentOtherV1(ctx).
			Algo("sha256").
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining payment other") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get mining payment other: %v", err)
		}

		t.Logf("Mining payment other: %+v", resp)
	})

	t.Run("GetMiningPaymentUid", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningPaymentUidV1(ctx).
			Algo("sha256").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mining payment uid") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get mining payment uid: %v", err)
		}

		t.Logf("Mining payment uid: %+v", resp)
	})
}

// TestMiningHashTransfer tests mining hash transfer endpoints
func TestMiningHashTransfer(t *testing.T) {
	// Skip by default as mining requires special setup
	if os.Getenv("BINANCE_TEST_MINING_HASH_TRANSFER") != "true" {
		t.Skip("Set BINANCE_TEST_MINING_HASH_TRANSFER=true to test hash transfer endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetHashTransferConfigList", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningHashTransferConfigDetailsListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Hash transfer config list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get hash transfer config list: %v", err)
		}

		t.Logf("Hash transfer config list: %+v", resp)
	})

	t.Run("GetHashTransferProfitDetails", func(t *testing.T) {
		resp, httpResp, err := client.MiningAPI.GetMiningHashTransferProfitDetailsV1(ctx).
			ConfigId(123).
			UserName("testuser").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Hash transfer profit details") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get hash transfer profit details: %v", err)
		}

		t.Logf("Hash transfer profit details: %+v", resp)
	})

	t.Run("ConfigureHashTransfer", func(t *testing.T) {
		// This requires specific mining setup
		toUser := os.Getenv("BINANCE_MINING_TRANSFER_USER")
		if toUser == "" {
			t.Skip("BINANCE_MINING_TRANSFER_USER not set")
		}

		resp, _, err := client.MiningAPI.CreateMiningHashTransferConfigV1(ctx).
			UserName(toUser).
			Algo("sha256").
			ToPoolUser(toUser).
			HashRate(100000000). // 100 MH/s
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient hash rate or invalid user
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -12003 || code == -12004) {
						t.Skip("Insufficient hash rate or invalid user")
					}
				}
			}
			t.Fatalf("Failed to configure hash transfer: %v", err)
		}

		t.Logf("Hash transfer configured: %+v", resp)

		// Cancel the configuration if created
		if resp.Data != nil {
			cancelResp, _, err := client.MiningAPI.CreateMiningHashTransferConfigCancelV1(ctx).
				ConfigId(int32(*resp.Data)).
				Timestamp(timestamp).
				Execute()

			if err != nil {
				t.Logf("Warning: Failed to cancel hash transfer config: %v", err)
			} else {
				t.Logf("Hash transfer config cancelled: %+v", cancelResp)
			}
		}
	})
}