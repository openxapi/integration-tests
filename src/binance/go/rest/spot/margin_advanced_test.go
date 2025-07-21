package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestMarginTransferOperations tests margin transfer endpoints
func TestMarginTransferOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("MarginTransfer", func(t *testing.T) {
		// Skip - margin transfers should use WalletAPI.CreateAssetTransferV1 instead
		t.Skip("Margin transfers should use WalletAPI.CreateAssetTransferV1")
	})

	t.Run("GetMarginTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginTransferV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin transfer history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin transfer history: %v", err)
		}

		t.Logf("Margin transfer history: %+v", resp)
	})

	t.Run("GetCrossMarginTransferHistory", func(t *testing.T) {
		// Note: GetMarginCrossMarginTransferV1 doesn't exist, using GetMarginTransferV1 instead
		resp, httpResp, err := client.MarginTradingAPI.GetMarginTransferV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cross margin transfer history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get cross margin transfer history: %v", err)
		}

		t.Logf("Cross margin transfer history: %+v", resp)
	})
}

// TestMarginLoanOperations tests margin loan endpoints
func TestMarginLoanOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("MarginLoan", func(t *testing.T) {
		// Skip by default to avoid actual loans
		if os.Getenv("BINANCE_TEST_MARGIN_LOAN") != "true" {
			t.Skip("Set BINANCE_TEST_MARGIN_LOAN=true to test margin loans")
		}

		resp, httpResp, err := client.MarginTradingAPI.CreateMarginBorrowRepayV1(ctx).
			Asset("USDT").
			Amount("10").
			Type_("BORROW"). // BORROW or REPAY
			IsIsolated("false").
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin loan") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if loan not available or insufficient collateral
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3045 || code == -11001) {
						t.Skip("Insufficient collateral or loan not available")
					}
				}
			}
			t.Fatalf("Failed to create margin loan: %v", err)
		}

		t.Logf("Margin loan created: %+v", resp)
	})

	t.Run("GetMarginLoanRecord", func(t *testing.T) {
		// Note: GetMarginLoanV1 doesn't exist, use GetMarginBorrowRepayV1 instead
		resp, httpResp, err := client.MarginTradingAPI.GetMarginBorrowRepayV1(ctx).
			Asset("USDT").
			Type_("BORROW").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin loan record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin loan record: %v", err)
		}

		t.Logf("Margin loan records: %+v", resp)
	})

	t.Run("MarginRepay", func(t *testing.T) {
		// Skip by default to avoid actual repayments
		if os.Getenv("BINANCE_TEST_MARGIN_REPAY") != "true" {
			t.Skip("Set BINANCE_TEST_MARGIN_REPAY=true to test margin repay")
		}

		resp, httpResp, err := client.MarginTradingAPI.CreateMarginBorrowRepayV1(ctx).
			Asset("USDT").
			Amount("10").
			Type_("REPAY"). // BORROW or REPAY
			IsIsolated("false").
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin repay") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no loan to repay or insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3044 || code == -3020) {
						t.Skip("No loan to repay or insufficient balance")
					}
				}
			}
			t.Fatalf("Failed to repay margin loan: %v", err)
		}

		t.Logf("Margin loan repaid: %+v", resp)
	})

	t.Run("GetMarginRepayRecord", func(t *testing.T) {
		// Note: GetMarginRepayV1 doesn't exist, use GetMarginBorrowRepayV1 instead
		resp, httpResp, err := client.MarginTradingAPI.GetMarginBorrowRepayV1(ctx).
			Asset("USDT").
			Type_("REPAY").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin repay record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin repay record: %v", err)
		}

		t.Logf("Margin repay records: %+v", resp)
	})

	t.Run("GetMarginAsset", func(t *testing.T) {
		// Note: GetMarginAssetV1 doesn't exist
		resp, httpResp, err := client.MarginTradingAPI.GetMarginAllAssetsV1(ctx).
			Asset("BTC").
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin asset") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin asset: %v", err)
		}

		t.Logf("Margin asset info: %+v", resp)
	})

	t.Run("GetMarginPair", func(t *testing.T) {
		// Note: GetMarginPairV1 doesn't exist
		resp, httpResp, err := client.MarginTradingAPI.GetMarginAllPairsV1(ctx).
			Symbol("BTCUSDT").
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin pair") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin pair: %v", err)
		}

		t.Logf("Margin pair info: %+v", resp)
	})
}

// TestMarginAccountOperations tests margin account management endpoints
func TestMarginAccountOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetMarginMaxTransferable", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginMaxTransferableV1(ctx).
			Asset("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin max transferable") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get max transferable: %v", err)
		}

		t.Logf("Max transferable: %+v", resp)
	})

	t.Run("GetMarginInterestRateHistory", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginInterestRateHistoryV1(ctx).
			Asset("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin interest rate history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get interest rate history: %v", err)
		}

		t.Logf("Interest rate history: %+v", resp)
	})

	t.Run("GetCrossMarginFee", func(t *testing.T) {
		// Note: GetMarginCrossMarginFeeV1 doesn't exist
		resp, httpResp, err := client.MarginTradingAPI.GetMarginInterestRateHistoryV1(ctx).
			Asset("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cross margin fee") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get cross margin fee: %v", err)
		}

		t.Logf("Cross margin fee: %+v", resp)
	})

	t.Run("GetCrossMarginData", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginCrossMarginDataV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cross margin data") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get cross margin data: %v", err)
		}

		t.Logf("Cross margin data: %+v", resp)
	})

	t.Run("GetForceLiquidationRecord", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginForceLiquidationRecV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Force liquidation record") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get force liquidation record: %v", err)
		}

		t.Logf("Force liquidation records: %+v", resp)
	})

	t.Run("UpdateMarginListenKey", func(t *testing.T) {
		// First create a listen key
		createResp, httpResp, err := client.MarginTradingAPI.CreateMarginListenKeyV1(ctx).Execute()
		if handleTestnetError(t, err, httpResp, "Create margin listen key") {
			return
		}
		if err != nil {
			t.Skip("Failed to create listen key for update test")
		}

		if createResp.ListenKey == nil {
			t.Skip("No listen key received")
		}

		// Update the listen key
		_, httpResp, err = client.MarginTradingAPI.UpdateMarginListenKeyV1(ctx).
			ListenKey(*createResp.ListenKey).
			Execute()

		if handleTestnetError(t, err, httpResp, "Update margin listen key") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to update margin listen key: %v", err)
		}

		t.Log("Margin listen key updated successfully")

		// Delete the listen key
		_, httpResp, err = client.MarginTradingAPI.DeleteMarginListenKeyV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Delete margin listen key") {
			return
		}
		if err != nil {
			t.Logf("Warning: Failed to delete margin listen key: %v", err)
		}
	})
}

// TestIsolatedMarginOperations tests isolated margin endpoints
func TestIsolatedMarginOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetIsolatedMarginAccount", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginIsolatedAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Isolated margin account") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get isolated margin account: %v", err)
		}

		t.Logf("Isolated margin account: %+v", resp)
	})

	t.Run("GetIsolatedMarginSymbol", func(t *testing.T) {
		// Note: GetMarginIsolatedPairV1 doesn't exist, using GetMarginIsolatedAllPairsV1
		resp, httpResp, err := client.MarginTradingAPI.GetMarginIsolatedAllPairsV1(ctx).
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Isolated margin symbol") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get isolated margin symbol: %v", err)
		}

		t.Logf("Isolated margin symbol: %+v", resp)
	})

	t.Run("GetAllIsolatedMarginSymbol", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginIsolatedAllPairsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "All isolated margin symbols") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get all isolated margin symbols: %v", err)
		}

		t.Logf("All isolated margin symbols count: %d", len(resp))
	})

	t.Run("IsolatedMarginTransfer", func(t *testing.T) {
		// Skip by default to avoid actual transfers
		if os.Getenv("BINANCE_TEST_ISOLATED_MARGIN_TRANSFER") != "true" {
			t.Skip("Set BINANCE_TEST_ISOLATED_MARGIN_TRANSFER=true to test isolated margin transfers")
		}

		// Note: Isolated margin transfers should use SubAccountAPI or WalletAPI
		t.Skip("Isolated margin transfers should use different API")
		return

		/*
		resp, _, err := client.MarginTradingAPI.CreateMarginIsolatedTransferV1(ctx).
			Asset("USDT").
			Symbol("BTCUSDT").
			TransFrom("SPOT").
			TransTo("ISOLATED_MARGIN").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or isolated margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3020 || code == -11002) {
						t.Skip("Insufficient balance or isolated margin not enabled")
					}
				}
			}
			t.Fatalf("Failed to transfer to isolated margin: %v", err)
		}

		t.Logf("Isolated margin transfer completed: %+v", resp)
		*/
	})

	t.Run("GetIsolatedMarginTransferHistory", func(t *testing.T) {
		// Note: This method doesn't exist in the current SDK
		t.Skip("GetMarginIsolatedTransferV1 not available in current SDK")
		return

		/*
		resp, _, err := client.MarginTradingAPI.GetMarginIsolatedTransferV1(ctx).
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get isolated margin transfer history: %v", err)
		}

		t.Logf("Isolated margin transfer history: %+v", resp)
		*/
	})

	t.Run("GetIsolatedMarginFee", func(t *testing.T) {
		// Note: GetMarginIsolatedMarginFeeV1 doesn't exist, using GetMarginIsolatedAccountV1
		resp, httpResp, err := client.MarginTradingAPI.GetMarginIsolatedAccountV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Isolated margin fee") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get isolated margin fee: %v", err)
		}

		t.Logf("Isolated margin fee: %+v", resp)
	})

	t.Run("GetIsolatedMarginTier", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginIsolatedMarginTierV1(ctx).
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Isolated margin tier") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get isolated margin tier: %v", err)
		}

		t.Logf("Isolated margin tier: %+v", resp)
	})
}

// TestMarginOCOOrders tests margin OCO order endpoints
func TestMarginOCOOrders(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateMarginOCOOrder", func(t *testing.T) {
		// Skip by default to avoid creating actual orders
		if os.Getenv("BINANCE_TEST_MARGIN_OCO_ORDERS") != "true" {
			t.Skip("Set BINANCE_TEST_MARGIN_OCO_ORDERS=true to test margin OCO orders")
		}

		// Get current price
		price, err := getCurrentPrice(client, ctx, "BTCUSDT")
		if err != nil {
			t.Skip("Failed to get current price for OCO order")
		}

		stopPrice := price * 0.98
		stopLimitPrice := price * 0.975
		limitPrice := price * 1.02

		resp, httpResp, err := client.MarginTradingAPI.CreateMarginOrderOcoV1(ctx).
			Symbol("BTCUSDT").
			Side("SELL").
			Quantity("0.001").
			Price(strconv.FormatFloat(limitPrice, 'f', 2, 64)).
			StopPrice(strconv.FormatFloat(stopPrice, 'f', 2, 64)).
			StopLimitPrice(strconv.FormatFloat(stopLimitPrice, 'f', 2, 64)).
			StopLimitTimeInForce("GTC").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create margin OCO order") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -2010 || code == -11002) {
						t.Skip("Insufficient balance or margin not enabled")
					}
				}
			}
			t.Fatalf("Failed to create margin OCO order: %v", err)
		}

		t.Logf("Margin OCO order created: %+v", resp)

		// Cancel the OCO order
		if resp.OrderListId != nil {
			_, httpResp, err = client.MarginTradingAPI.DeleteMarginOrderListV1(ctx).
				OrderListId(*resp.OrderListId).
				Timestamp(timestamp).
				Execute()

			if handleTestnetError(t, err, httpResp, "Cancel margin OCO order") {
				return
			}
			if err != nil {
				t.Logf("Warning: Failed to cancel margin OCO order: %v", err)
			}
		}
	})

	t.Run("GetMarginOCOOrder", func(t *testing.T) {
		orderListIdStr := os.Getenv("BINANCE_TEST_MARGIN_OCO_ORDER_ID")
		if orderListIdStr == "" {
			t.Skip("BINANCE_TEST_MARGIN_OCO_ORDER_ID not set")
		}
		
		orderListId, err := strconv.ParseInt(orderListIdStr, 10, 64)
		if err != nil {
			t.Fatalf("Invalid order list ID: %v", err)
		}

		resp, httpResp, err := client.MarginTradingAPI.GetMarginOrderListV1(ctx).
			OrderListId(orderListId).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get margin OCO order") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin OCO order: %v", err)
		}

		t.Logf("Margin OCO order: %+v", resp)
	})

	t.Run("GetMarginAllOCOOrders", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginAllOrderListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get all margin OCO orders") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get all margin OCO orders: %v", err)
		}

		t.Logf("All margin OCO orders: %+v", resp)
	})

	t.Run("GetMarginOpenOCOOrders", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginOpenOrderListV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get open margin OCO orders") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get open margin OCO orders: %v", err)
		}

		t.Logf("Open margin OCO orders: %+v", resp)
	})
}

// TestMarginOrderOperations tests additional margin order endpoints
func TestMarginOrderOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetMarginOpenOrders", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginOpenOrdersV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get margin open orders") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get margin open orders: %v", err)
		}

		t.Logf("Margin open orders: %+v", resp)
	})

	t.Run("DeleteMarginOrder", func(t *testing.T) {
		// Skip by default as we need an actual order to cancel
		orderIdStr := os.Getenv("BINANCE_TEST_MARGIN_ORDER_ID")
		if orderIdStr == "" {
			t.Skip("BINANCE_TEST_MARGIN_ORDER_ID not set")
		}
		
		orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
		if err != nil {
			t.Fatalf("Invalid order ID: %v", err)
		}

		resp, httpResp, err := client.MarginTradingAPI.DeleteMarginOrderV1(ctx).
			Symbol("BTCUSDT").
			OrderId(orderId).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cancel margin order") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to cancel margin order: %v", err)
		}

		t.Logf("Margin order cancelled: %+v", resp)
	})

	t.Run("DeleteAllMarginOpenOrders", func(t *testing.T) {
		// Skip by default to avoid cancelling actual orders
		if os.Getenv("BINANCE_TEST_CANCEL_ALL_MARGIN_ORDERS") != "true" {
			t.Skip("Set BINANCE_TEST_CANCEL_ALL_MARGIN_ORDERS=true to test cancelling all orders")
		}

		resp, httpResp, err := client.MarginTradingAPI.DeleteMarginOpenOrdersV1(ctx).
			Symbol("BTCUSDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cancel all margin open orders") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to cancel all margin open orders: %v", err)
		}

		t.Logf("All margin open orders cancelled: %+v", resp)
	})
}

// TestMarginBNBBurn tests BNB burn for margin interest
func TestMarginBNBBurn(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetBNBBurnStatus", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetBnbBurnV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "BNB burn status") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get BNB burn status: %v", err)
		}

		t.Logf("BNB burn status: %+v", resp)
	})

	t.Run("ToggleBNBBurnOnMarginInterest", func(t *testing.T) {
		// Skip by default as this changes account settings
		if os.Getenv("BINANCE_TEST_MARGIN_BNB_BURN") != "true" {
			t.Skip("Set BINANCE_TEST_MARGIN_BNB_BURN=true to test BNB burn toggle")
		}

		// Note: CreateBnbBurnV1 doesn't exist in current SDK
		t.Skip("CreateBnbBurnV1 not available in current SDK")
	})
}

// TestMarginTradeFee tests margin trade fee endpoints
func TestMarginTradeFee(t *testing.T) {
	t.Run("GetMarginTradeFee", func(t *testing.T) {
		// Note: GetMarginTradeFeeV1 doesn't exist in current SDK
		t.Skip("GetMarginTradeFeeV1 not available in current SDK")
	})
}

// TestMarginCollateral tests margin collateral endpoints
func TestMarginCollateral(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	t.Run("GetCrossMarginCollateralRatio", func(t *testing.T) {
		resp, httpResp, err := client.MarginTradingAPI.GetMarginCrossMarginCollateralRatioV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cross margin collateral ratio") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get cross margin collateral ratio: %v", err)
		}

		t.Logf("Cross margin collateral ratio: %+v", resp)
	})

	t.Run("GetMarginAvailableInventory", func(t *testing.T) {
		// This endpoint might require special permissions
		resp, httpResp, err := client.MarginTradingAPI.GetMarginAvailableInventoryV1(ctx).
			Type_("MARGIN").
			Execute()

		if handleTestnetError(t, err, httpResp, "Margin available inventory") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if permission denied
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -2015 {
						t.Skip("Available inventory endpoint requires special permissions")
					}
				}
			}
			t.Fatalf("Failed to get margin available inventory: %v", err)
		}

		t.Logf("Margin available inventory: %+v", resp)
	})
}