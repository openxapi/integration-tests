package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestCryptoLoanInfo tests crypto loan information endpoints
func TestCryptoLoanInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetFlexibleLoanableData", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleLoanableDataV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible loanable data") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if crypto loan not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001) {
						t.Skip("Crypto loan not available on testnet or permission denied")
					}
				}
			}
			t.Fatalf("Failed to get flexible loanable data: %v", err)
		}

		t.Logf("Flexible loanable data: %+v", resp)
	})

	t.Run("GetFlexibleCollateralData", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleCollateralDataV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible collateral data") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if crypto loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001) {
						t.Skip("Crypto loan not available on testnet or permission denied")
					}
				}
			}
			t.Fatalf("Failed to get flexible collateral data: %v", err)
		}

		t.Logf("Flexible collateral data: %+v", resp)
	})

	t.Run("GetFlexibleRepayRate", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleRepayRateV2(ctx).
			LoanCoin("USDT").
			CollateralCoin("BTC").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible repay rate") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if crypto loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001) {
						t.Skip("Crypto loan not available on testnet or permission denied")
					}
				}
			}
			t.Fatalf("Failed to get flexible repay rate: %v", err)
		}

		t.Logf("Flexible repay rate: %+v", resp)
	})
}

// TestCryptoLoanHistory tests crypto loan history endpoints
func TestCryptoLoanHistory(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetLoanBorrowHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanBorrowHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Loan borrow history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get loan borrow history: %v", err)
		}

		t.Logf("Loan borrow history: %+v", resp)
	})

	t.Run("GetFlexibleBorrowHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleBorrowHistoryV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible borrow history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get flexible borrow history: %v", err)
		}

		t.Logf("Flexible borrow history: %+v", resp)
	})

	t.Run("GetLoanRepayHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanRepayHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Loan repay history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get loan repay history: %v", err)
		}

		t.Logf("Loan repay history: %+v", resp)
	})

	t.Run("GetFlexibleRepayHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleRepayHistoryV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible repay history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get flexible repay history: %v", err)
		}

		t.Logf("Flexible repay history: %+v", resp)
	})

	t.Run("GetLoanLtvAdjustmentHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanLtvAdjustmentHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Loan LTV adjustment history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get loan LTV adjustment history: %v", err)
		}

		t.Logf("Loan LTV adjustment history: %+v", resp)
	})

	t.Run("GetFlexibleLtvAdjustmentHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleLtvAdjustmentHistoryV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible LTV adjustment history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get flexible LTV adjustment history: %v", err)
		}

		t.Logf("Flexible LTV adjustment history: %+v", resp)
	})

	t.Run("GetFlexibleLiquidationHistory", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleLiquidationHistoryV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible liquidation history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get flexible liquidation history: %v", err)
		}

		t.Logf("Flexible liquidation history: %+v", resp)
	})

	t.Run("GetLoanIncome", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanIncomeV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Loan income") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get loan income: %v", err)
		}

		t.Logf("Loan income: %+v", resp)
	})
}

// TestCryptoLoanOrders tests crypto loan order endpoints
func TestCryptoLoanOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetFlexibleOngoingOrders", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.GetLoanFlexibleOngoingOrdersV2(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Flexible ongoing orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get flexible ongoing orders: %v", err)
		}

		t.Logf("Flexible ongoing orders: %+v", resp)
	})
}

// TestCryptoLoanOperations tests crypto loan operations (use with caution)
func TestCryptoLoanOperations(t *testing.T) {
	// Skip by default to avoid creating actual loans
	if os.Getenv("BINANCE_TEST_CRYPTO_LOAN") != "true" {
		t.Skip("Set BINANCE_TEST_CRYPTO_LOAN=true to test crypto loan operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateFlexibleLoan", func(t *testing.T) {
		resp, httpResp, err := client.CryptoLoanAPI.CreateLoanFlexibleBorrowV2(ctx).
			LoanCoin("USDT").
			CollateralCoin("BTC").
			LoanAmount("10"). // Borrow 10 USDT
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create flexible loan") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient collateral or loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3045 || code == -4026) {
						t.Skip("Insufficient collateral or crypto loan not available")
					}
				}
			}
			t.Fatalf("Failed to create flexible loan: %v", err)
		}

		t.Logf("Flexible loan created: %+v", resp)
	})

	t.Run("AdjustFlexibleLoanLTV", func(t *testing.T) {
		// This requires an active loan
		orderId := os.Getenv("BINANCE_TEST_LOAN_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_LOAN_ORDER_ID not set")
		}

		resp, _, err := client.CryptoLoanAPI.CreateLoanFlexibleAdjustLtvV2(ctx).
			LoanCoin("USDT").
			CollateralCoin("BTC").
			Direction("ADDITIONAL").    // Add more collateral
			AdjustmentAmount("0.0001"). // Add 0.0001 BTC
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -4013 {
						t.Skip("Loan order not found")
					}
				}
			}
			t.Fatalf("Failed to adjust LTV: %v", err)
		}

		t.Logf("LTV adjusted: %+v", resp)
	})

	t.Run("RepayFlexibleLoan", func(t *testing.T) {
		// This requires an active loan
		orderId := os.Getenv("BINANCE_TEST_LOAN_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_LOAN_ORDER_ID not set")
		}

		resp, _, err := client.CryptoLoanAPI.CreateLoanFlexibleRepayV2(ctx).
			LoanCoin("USDT").
			CollateralCoin("BTC").
			RepayAmount("10").
			CollateralReturn(false). // false = repay with loan coin
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found or insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -4013 || code == -3020) {
						t.Skip("Loan order not found or insufficient balance")
					}
				}
			}
			t.Fatalf("Failed to repay flexible loan: %v", err)
		}

		t.Logf("Flexible loan repaid: %+v", resp)
	})

	t.Run("RepayFlexibleLoanWithCollateral", func(t *testing.T) {
		// This requires an active loan
		orderId := os.Getenv("BINANCE_TEST_LOAN_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_LOAN_ORDER_ID not set")
		}

		resp, _, err := client.CryptoLoanAPI.CreateLoanFlexibleRepayCollateralV2(ctx).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -4013 {
						t.Skip("Loan order not found")
					}
				}
			}
			t.Fatalf("Failed to repay with collateral: %v", err)
		}

		t.Logf("Loan repaid with collateral: %+v", resp)
	})
}