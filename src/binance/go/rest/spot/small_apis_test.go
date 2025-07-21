package main

import (
	"context"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestNFTAPI tests NFT endpoints
func TestNFTAPI(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetNFTTransactionHistory", func(t *testing.T) {
		resp, httpResp, err := client.NftAPI.GetNftUserGetAssetV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "NFT transaction history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if NFT not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -18000) {
						t.Skip("NFT API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get NFT transaction history: %v", err)
		}

		t.Logf("NFT transaction history: %+v", resp)
	})

	t.Run("GetNFTDepositHistory", func(t *testing.T) {
		resp, httpResp, err := client.NftAPI.GetNftHistoryDepositV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "NFT deposit history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get NFT deposit history: %v", err)
		}

		t.Logf("NFT deposit history: %+v", resp)
	})

	t.Run("GetNFTWithdrawHistory", func(t *testing.T) {
		resp, httpResp, err := client.NftAPI.GetNftHistoryWithdrawV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "NFT withdraw history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get NFT withdraw history: %v", err)
		}

		t.Logf("NFT withdraw history: %+v", resp)
	})

	t.Run("GetNFTAsset", func(t *testing.T) {
		resp, httpResp, err := client.NftAPI.GetNftUserGetAssetV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "NFT assets") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get NFT assets: %v", err)
		}

		t.Logf("NFT assets: %+v", resp)
	})
}

// TestFiatAPI tests fiat endpoints
func TestFiatAPI(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetFiatOrderHistory", func(t *testing.T) {
		resp, httpResp, err := client.FiatAPI.GetFiatOrdersV1(ctx).
			TransactionType("0"). // 0: deposit, 1: withdraw
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Fiat order history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if fiat not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Fiat API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get fiat order history: %v", err)
		}

		t.Logf("Fiat order history: %+v", resp)
	})

	t.Run("GetFiatPaymentHistory", func(t *testing.T) {
		resp, httpResp, err := client.FiatAPI.GetFiatPaymentsV1(ctx).
			TransactionType("0"). // 0: buy, 1: sell
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Fiat payment history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get fiat payment history: %v", err)
		}

		t.Logf("Fiat payment history: %+v", resp)
	})
}

// TestC2CAPI tests C2C endpoints
func TestC2CAPI(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetC2CTradeHistory", func(t *testing.T) {
		resp, httpResp, err := client.C2cAPI.GetC2cOrderMatchListUserOrderHistoryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "C2C trade history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if C2C not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("C2C API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get C2C trade history: %v", err)
		}

		t.Logf("C2C trade history: %+v", resp)
	})
}

// TestBinancePayHistoryAPI tests Binance Pay history endpoints
func TestBinancePayHistoryAPI(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetPayTradeHistory", func(t *testing.T) {
		resp, httpResp, err := client.BinancePayHistoryAPI.GetPayTransactionsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Pay trade history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if Binance Pay not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Binance Pay API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get pay trade history: %v", err)
		}

		t.Logf("Pay trade history: %+v", resp)
	})
}

// TestCopyTradingAPI tests copy trading endpoints
func TestCopyTradingAPI(t *testing.T) {

	t.Run("GetCopyTradingStatus", func(t *testing.T) {
		// Note: GetPapiAccountCopytradingStatusV1 doesn't exist in current SDK
		t.Skip("GetPapiAccountCopytradingStatusV1 not available")
		return

		/*
		resp, _, err := client.CopyTradingAPI.GetPapiAccountCopytradingStatusV1(ctx).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if copy trading not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -2015) {
						t.Skip("Copy trading not available on testnet or permission denied")
					}
				}
			}
			t.Fatalf("Failed to get copy trading status: %v", err)
		}

		t.Logf("Copy trading status: %+v", resp)
		*/
	})

	t.Run("GetCopyTradingData", func(t *testing.T) {
		// Note: GetPapiAccountCopytradingDataV1 doesn't exist in current SDK
		t.Skip("GetPapiAccountCopytradingDataV1 not available")
		return

		/*
		resp, _, err := client.CopyTradingAPI.GetPapiAccountCopytradingDataV1(ctx).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get copy trading data: %v", err)
		}

		t.Logf("Copy trading data: %+v", resp)
		*/
	})
}

// TestFuturesDataAPI tests futures data endpoints
func TestFuturesDataAPI(t *testing.T) {

	t.Run("GetFuturesTickLevel", func(t *testing.T) {
		// Note: GetFuturesDataTickLevelOrderbookV1 doesn't exist in current SDK
		t.Skip("GetFuturesDataTickLevelOrderbookV1 not available")
		return

		/*
		resp, _, err := client.FuturesDataAPI.GetFuturesDataTickLevelOrderbookV1(ctx).
			Symbol("BTCUSDT").
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if futures data not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Futures data API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get futures tick level: %v", err)
		}

		t.Logf("Futures tick level: %+v", resp)
		*/
	})
}

// TestRebateAPI tests rebate endpoints
func TestRebateAPI(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetRebateSpotHistory", func(t *testing.T) {
		resp, httpResp, err := client.RebateAPI.GetRebateTaxQueryV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Rebate spot history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if rebate not available or no broker account
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -2008) {
						t.Skip("Rebate API not available or not a broker account")
					}
				}
			}
			t.Fatalf("Failed to get rebate spot history: %v", err)
		}

		t.Logf("Rebate spot history: %+v", resp)
	})
}