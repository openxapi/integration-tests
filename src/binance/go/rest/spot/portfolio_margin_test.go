package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestPortfolioMarginAccount tests portfolio margin account endpoints
func TestPortfolioMarginAccount(t *testing.T) {
	// Skip by default as portfolio margin requires special account setup
	if os.Getenv("BINANCE_TEST_PORTFOLIO_MARGIN") != "true" {
		t.Skip("Set BINANCE_TEST_PORTFOLIO_MARGIN=true to test portfolio margin endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetPortfolioAccount", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if portfolio margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -3000) {
						t.Skip("Portfolio margin not enabled for this account")
					}
				}
			}
			t.Fatalf("Failed to get portfolio account: %v", err)
		}

		t.Logf("Portfolio account: %+v", resp)
	})

	t.Run("GetPortfolioAccountV2", func(t *testing.T) {
		resp, err := client.PortfolioMarginProAPI.GetPortfolioAccountV2(ctx).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			// Check if this is a testnet limitation error
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				
				// Check for 404 status code (endpoint not available on testnet)
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
					t.Skip("Portfolio account v2 endpoint not available on testnet")
				}
				
				// Skip if portfolio margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -3000) {
						t.Skip("Portfolio margin not enabled for this account")
					}
				}
			}
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio account v2: %v", err)
		}

		t.Logf("Portfolio account v2: %+v", resp)
	})

	t.Run("GetPortfolioBalance", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioBalanceV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio balance") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio balance: %v", err)
		}

		t.Logf("Portfolio balance: %+v", resp)
	})

	t.Run("GetPortfolioCollateralRate", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioCollateralRateV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio collateral rate") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio collateral rate: %v", err)
		}

		t.Logf("Portfolio collateral rate: %+v", resp)
	})

	t.Run("GetPortfolioCollateralRateV2", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioCollateralRateV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio collateral rate v2") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio collateral rate v2: %v", err)
		}

		t.Logf("Portfolio collateral rate v2: %+v", resp)
	})

	t.Run("GetPortfolioMarginAssetLeverage", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioMarginAssetLeverageV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio margin asset leverage") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio margin asset leverage: %v", err)
		}

		t.Logf("Portfolio margin asset leverage: %+v", resp)
	})

	t.Run("GetPortfolioAssetIndexPrice", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioAssetIndexPriceV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio asset index price") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio asset index price: %v", err)
		}

		t.Logf("Portfolio asset index price: %+v", resp)
	})
}

// TestPortfolioMarginLoan tests portfolio margin loan endpoints
func TestPortfolioMarginLoan(t *testing.T) {
	// Skip by default as portfolio margin requires special account setup
	if os.Getenv("BINANCE_TEST_PORTFOLIO_MARGIN") != "true" {
		t.Skip("Set BINANCE_TEST_PORTFOLIO_MARGIN=true to test portfolio margin endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetPortfolioPmLoan", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioPmLoanV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio PM loan") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio pm loan: %v", err)
		}

		t.Logf("Portfolio pm loan: %+v", resp)
	})

	t.Run("GetPortfolioPmLoanHistory", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioPmLoanHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio PM loan history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio pm loan history: %v", err)
		}

		t.Logf("Portfolio pm loan history: %+v", resp)
	})

	t.Run("GetPortfolioInterestHistory", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioInterestHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio interest history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio interest history: %v", err)
		}

		t.Logf("Portfolio interest history: %+v", resp)
	})

	t.Run("GetPortfolioRepayFuturesSwitch", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.GetPortfolioRepayFuturesSwitchV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Portfolio repay futures switch") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get portfolio repay futures switch: %v", err)
		}

		t.Logf("Portfolio repay futures switch: %+v", resp)
	})
}

// TestPortfolioMarginOperations tests portfolio margin operations (use with caution)
func TestPortfolioMarginOperations(t *testing.T) {
	// Skip by default to avoid actual operations
	if os.Getenv("BINANCE_TEST_PORTFOLIO_MARGIN_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_PORTFOLIO_MARGIN_OPERATIONS=true to test portfolio margin operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("PortfolioBNBTransfer", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioBnbTransferV1(ctx).
			Amount("0.1").
			TransferSide("TO_UM"). // Transfer to UM account
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Transfer BNB") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3020 {
						t.Skip("Insufficient BNB balance")
					}
				}
			}
			t.Fatalf("Failed to transfer BNB: %v", err)
		}

		t.Logf("BNB transferred: %+v", resp)
	})

	t.Run("SetAutoCollection", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioAutoCollectionV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Set auto collection") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to set auto collection: %v", err)
		}

		t.Logf("Auto collection set: %+v", resp)
	})

	t.Run("AssetCollection", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioAssetCollectionV1(ctx).
			Asset("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Collect asset") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no balance to collect
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3021 {
						t.Skip("No balance to collect")
					}
				}
			}
			t.Fatalf("Failed to collect asset: %v", err)
		}

		t.Logf("Asset collected: %+v", resp)
	})

	t.Run("SetRepayFuturesSwitch", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioRepayFuturesSwitchV1(ctx).
			AutoRepay("true").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Set repay futures switch") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to set repay futures switch: %v", err)
		}

		t.Logf("Repay futures switch set: %+v", resp)
	})

	t.Run("RepayFuturesNegativeBalance", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioRepayFuturesNegativeBalanceV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Repay futures negative balance") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no negative balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3022 {
						t.Skip("No negative balance to repay")
					}
				}
			}
			t.Fatalf("Failed to repay futures negative balance: %v", err)
		}

		t.Logf("Futures negative balance repaid: %+v", resp)
	})

	t.Run("MintPortfolio", func(t *testing.T) {
		// This requires active positions
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioMintV1(ctx).
			FromAsset("USDT").
			TargetAsset("USDM").
			Amount("100").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Mint portfolio") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient collateral
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3045 {
						t.Skip("Insufficient collateral for minting")
					}
				}
			}
			t.Fatalf("Failed to mint portfolio: %v", err)
		}

		t.Logf("Portfolio minted: %+v", resp)
	})

	t.Run("RedeemPortfolio", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioRedeemV1(ctx).
			FromAsset("USDM").
			TargetAsset("USDT").
			Amount("100").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Redeem portfolio") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3020 {
						t.Skip("Insufficient balance to redeem")
					}
				}
			}
			t.Fatalf("Failed to redeem portfolio: %v", err)
		}

		t.Logf("Portfolio redeemed: %+v", resp)
	})

	t.Run("RepayPortfolio", func(t *testing.T) {
		resp, httpResp, err := client.PortfolioMarginProAPI.CreatePortfolioRepayV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Repay portfolio") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no loan to repay
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -3044 {
						t.Skip("No loan to repay")
					}
				}
			}
			t.Fatalf("Failed to repay portfolio: %v", err)
		}

		t.Logf("Portfolio repaid: %+v", resp)
	})
}