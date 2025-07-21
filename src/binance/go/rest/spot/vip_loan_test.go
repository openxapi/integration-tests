package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestVipLoanInfo tests VIP loan information endpoints
func TestVipLoanInfo(t *testing.T) {
	// Skip by default as VIP loans require special account status
	if os.Getenv("BINANCE_TEST_VIP_LOAN") != "true" {
		t.Skip("Set BINANCE_TEST_VIP_LOAN=true to test VIP loan endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetVipLoanableData", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipLoanableDataV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP loanable data") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if VIP loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001 || code == -13000) {
						t.Skip("VIP loan not available for this account")
					}
				}
			}
			t.Fatalf("Failed to get VIP loanable data: %v", err)
		}

		t.Logf("VIP loanable data: %+v", resp)
	})

	t.Run("GetVipCollateralData", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipCollateralDataV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP collateral data") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if VIP loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001 || code == -13000) {
						t.Skip("VIP loan not available for this account")
					}
				}
			}
			t.Fatalf("Failed to get VIP collateral data: %v", err)
		}

		t.Logf("VIP collateral data: %+v", resp)
	})

	t.Run("GetVipRequestInterestRate", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipRequestInterestRateV1(ctx).
			LoanCoin("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP request interest rate") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if VIP loan not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -4001 || code == -13000) {
						t.Skip("VIP loan not available for this account")
					}
				}
			}
			t.Fatalf("Failed to get VIP request interest rate: %v", err)
		}

		t.Logf("VIP request interest rate: %+v", resp)
	})

	t.Run("GetVipRequestData", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipRequestDataV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP request data") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP request data: %v", err)
		}

		t.Logf("VIP request data: %+v", resp)
	})

	t.Run("GetVipInterestRateHistory", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipInterestRateHistoryV1(ctx).
			Coin("USDT").
			RecvWindow(5000).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP interest rate history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP interest rate history: %v", err)
		}

		t.Logf("VIP interest rate history: %+v", resp)
	})
}

// TestVipLoanAccount tests VIP loan account endpoints
func TestVipLoanAccount(t *testing.T) {
	// Skip by default as VIP loans require special account status
	if os.Getenv("BINANCE_TEST_VIP_LOAN") != "true" {
		t.Skip("Set BINANCE_TEST_VIP_LOAN=true to test VIP loan endpoints")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetVipCollateralAccount", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipCollateralAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP collateral account") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP collateral account: %v", err)
		}

		t.Logf("VIP collateral account: %+v", resp)
	})

	t.Run("GetVipOngoingOrders", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipOngoingOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP ongoing orders") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP ongoing orders: %v", err)
		}

		t.Logf("VIP ongoing orders: %+v", resp)
	})

	t.Run("GetVipRepayHistory", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipRepayHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP repay history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP repay history: %v", err)
		}

		t.Logf("VIP repay history: %+v", resp)
	})

	t.Run("GetVipAccruedInterest", func(t *testing.T) {
		resp, httpResp, err := client.VipLoanAPI.GetLoanVipAccruedInterestV1(ctx).
			RecvWindow(5000).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "VIP accrued interest") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get VIP accrued interest: %v", err)
		}

		t.Logf("VIP accrued interest: %+v", resp)
	})
}

// TestVipLoanOperations tests VIP loan operations (use with extreme caution)
func TestVipLoanOperations(t *testing.T) {
	// Skip by default - VIP loans require special account status and large amounts
	if os.Getenv("BINANCE_TEST_VIP_LOAN_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_VIP_LOAN_OPERATIONS=true to test VIP loan operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateVipLoan", func(t *testing.T) {
		// Skip - SDK missing request body parameters for VIP loan borrow
		t.Skip("VIP loan borrow API missing required parameters in SDK")
		return
		
		/*
		resp, _, err := client.VipLoanAPI.CreateLoanVipBorrowV1(ctx).
			LoanCoin("USDT").
			LoanAmount("10000"). // VIP loans typically have high minimums
			CollateralCoin("BTC").
			CollateralAmount("0.5").
			LoanTerm(30). // 30 days
			Timestamp(timestamp).
			Execute()
		*/

		/*
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if not VIP or insufficient collateral
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -13000 || code == -3045) {
						t.Skip("Not VIP account or insufficient collateral")
					}
				}
			}
			t.Fatalf("Failed to create VIP loan: %v", err)
		}

		t.Logf("VIP loan created: %+v", resp)
		*/
	})

	t.Run("RenewVipLoan", func(t *testing.T) {
		orderId := os.Getenv("BINANCE_TEST_VIP_LOAN_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_VIP_LOAN_ORDER_ID not set")
		}

		orderIdInt, convErr := strconv.ParseInt(orderId, 10, 64)
		if convErr != nil {
			t.Fatalf("Failed to convert order ID to int64: %v", convErr)
		}

		resp, _, err := client.VipLoanAPI.CreateLoanVipRenewV1(ctx).
			OrderId(orderIdInt).
			LoanTerm(30). // Renew for another 30 days
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found or not eligible for renewal
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -13001 || code == -13002) {
						t.Skip("VIP loan order not found or not eligible for renewal")
					}
				}
			}
			t.Fatalf("Failed to renew VIP loan: %v", err)
		}

		t.Logf("VIP loan renewed: %+v", resp)
	})

	t.Run("RepayVipLoan", func(t *testing.T) {
		orderId := os.Getenv("BINANCE_TEST_VIP_LOAN_ORDER_ID")
		if orderId == "" {
			t.Skip("BINANCE_TEST_VIP_LOAN_ORDER_ID not set")
		}

		orderIdInt, convErr := strconv.ParseInt(orderId, 10, 64)
		if convErr != nil {
			t.Fatalf("Failed to convert order ID to int64: %v", convErr)
		}

		resp, _, err := client.VipLoanAPI.CreateLoanVipRepayV1(ctx).
			OrderId(orderIdInt).
			Amount("1000"). // Repay 1000 USDT
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if order not found or insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -13001 || code == -3020) {
						t.Skip("VIP loan order not found or insufficient balance")
					}
				}
			}
			t.Fatalf("Failed to repay VIP loan: %v", err)
		}

		t.Logf("VIP loan repaid: %+v", resp)
	})
}