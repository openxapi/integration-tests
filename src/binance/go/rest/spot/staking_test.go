package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestETHStakingAccount tests ETH staking account endpoints
func TestETHStakingAccount(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetETHStakingAccount", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingAccountV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH staking account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if ETH staking not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("ETH staking not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get ETH staking account: %v", err)
		}

		t.Logf("ETH staking account: %+v", resp)
	})

	t.Run("GetETHQuota", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthQuotaV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH quota") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if ETH staking not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("ETH staking not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get ETH quota: %v", err)
		}

		t.Logf("ETH quota: %+v", resp)
	})
}

// TestETHStakingHistory tests ETH staking history endpoints
func TestETHStakingHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetETHRateHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthHistoryRateHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH rate history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("ETH staking history not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get ETH rate history: %v", err)
		}

		t.Logf("ETH rate history: %+v", resp)
	})

	t.Run("GetETHStakingHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthHistoryStakingHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH staking history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get ETH staking history: %v", err)
		}

		t.Logf("ETH staking history: %+v", resp)
	})

	t.Run("GetETHRedemptionHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthHistoryRedemptionHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH redemption history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get ETH redemption history: %v", err)
		}

		t.Logf("ETH redemption history: %+v", resp)
	})

	t.Run("GetETHRewardsHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthHistoryRewardsHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "ETH rewards history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get ETH rewards history: %v", err)
		}

		t.Logf("ETH rewards history: %+v", resp)
	})

	t.Run("GetWBETHRewardsHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingEthHistoryWbethRewardsHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "WBETH rewards history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get WBETH rewards history: %v", err)
		}

		t.Logf("WBETH rewards history: %+v", resp)
	})

	t.Run("GetWBETHWrapHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingWbethHistoryWrapHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "WBETH wrap history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get WBETH wrap history: %v", err)
		}

		t.Logf("WBETH wrap history: %+v", resp)
	})

	t.Run("GetWBETHUnwrapHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetEthStakingWbethHistoryUnwrapHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "WBETH unwrap history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get WBETH unwrap history: %v", err)
		}

		t.Logf("WBETH unwrap history: %+v", resp)
	})
}

// TestSOLStakingAccount tests SOL staking account endpoints
func TestSOLStakingAccount(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSOLStakingAccount", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL staking account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if SOL staking not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("SOL staking not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get SOL staking account: %v", err)
		}

		t.Logf("SOL staking account: %+v", resp)
	})

	t.Run("GetSOLQuota", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolQuotaV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL quota") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if SOL staking not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("SOL staking not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get SOL quota: %v", err)
		}

		t.Logf("SOL quota: %+v", resp)
	})

	t.Run("GetSOLUnclaimedRewards", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryUnclaimedRewardsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL unclaimed rewards") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get SOL unclaimed rewards: %v", err)
		}

		t.Logf("SOL unclaimed rewards: %+v", resp)
	})
}

// TestSOLStakingHistory tests SOL staking history endpoints
func TestSOLStakingHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetSOLRateHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryRateHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL rate history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -6009) {
						t.Skip("SOL staking history not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get SOL rate history: %v", err)
		}

		t.Logf("SOL rate history: %+v", resp)
	})

	t.Run("GetSOLStakingHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryStakingHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL staking history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get SOL staking history: %v", err)
		}

		t.Logf("SOL staking history: %+v", resp)
	})

	t.Run("GetSOLRedemptionHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryRedemptionHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "SOL redemption history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get SOL redemption history: %v", err)
		}

		t.Logf("SOL redemption history: %+v", resp)
	})

	t.Run("GetBNSOLRewardsHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryBnsolRewardsHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "BNSOL rewards history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get BNSOL rewards history: %v", err)
		}

		t.Logf("BNSOL rewards history: %+v", resp)
	})

	t.Run("GetBoostRewardsHistory", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryBoostRewardsHistoryV1(ctx).
			Type_("BOOST").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Boost rewards history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get boost rewards history: %v", err)
		}

		t.Logf("Boost rewards history: %+v", resp)
	})
}

// TestStakingOperations tests staking operations (use with caution)
func TestStakingOperations(t *testing.T) {
	// Skip by default to avoid actual staking
	if os.Getenv("BINANCE_TEST_STAKING_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_STAKING_OPERATIONS=true to test staking operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("StakeETH", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.CreateEthStakingEthStakeV2(ctx).
			Amount("0.01").  // Minimum stake amount
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Stake ETH") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or minimum not met
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -6030 || code == -6031) {
						t.Skip("Insufficient balance or minimum stake amount not met")
					}
				}
			}
			t.Fatalf("Failed to stake ETH: %v", err)
		}

		t.Logf("ETH staked: %+v", resp)
	})

	t.Run("RedeemETH", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.CreateEthStakingEthRedeemV1(ctx).
			Amount("0.01").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Redeem ETH") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no staked position
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6024 {
						t.Skip("No staked position to redeem")
					}
				}
			}
			t.Fatalf("Failed to redeem ETH: %v", err)
		}

		t.Logf("ETH redeemed: %+v", resp)
	})

	t.Run("WrapETHToWBETH", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.CreateEthStakingWbethWrapV1(ctx).
			Amount("0.01").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Wrap ETH to WBETH") {
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
						t.Skip("Insufficient balance for wrap")
					}
				}
			}
			t.Fatalf("Failed to wrap ETH to WBETH: %v", err)
		}

		t.Logf("WBETH wrapped: %+v", resp)
	})

	t.Run("StakeSOL", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.CreateSolStakingSolStakeV1(ctx).
			Amount("0.1").  // Minimum stake amount
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Stake SOL") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or minimum not met
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -6030 || code == -6031) {
						t.Skip("Insufficient balance or minimum stake amount not met")
					}
				}
			}
			t.Fatalf("Failed to stake SOL: %v", err)
		}

		t.Logf("SOL staked: %+v", resp)
	})

	t.Run("RedeemSOL", func(t *testing.T) {
		resp, httpResp, err := client.StakingAPI.CreateSolStakingSolRedeemV1(ctx).
			Amount("0.1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Redeem SOL") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no staked position
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6024 {
						t.Skip("No staked position to redeem")
					}
				}
			}
			t.Fatalf("Failed to redeem SOL: %v", err)
		}

		t.Logf("SOL redeemed: %+v", resp)
	})

	t.Run("ClaimSOLRewards", func(t *testing.T) {
		// First check unclaimed rewards
		rewardsResp, httpResp, err := client.StakingAPI.GetSolStakingSolHistoryUnclaimedRewardsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			t.Skip("Failed to check unclaimed rewards")
		}

		// Skip if no rewards to claim
		if len(rewardsResp) == 0 {
			t.Skip("No rewards to claim")
		}

		// Claim rewards
		resp, httpResp, err := client.StakingAPI.CreateSolStakingSolClaimV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Claim SOL rewards") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no rewards
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6025 {
						t.Skip("No rewards to claim")
					}
				}
			}
			t.Fatalf("Failed to claim SOL rewards: %v", err)
		}

		t.Logf("SOL rewards claimed: %+v", resp)
	})
}