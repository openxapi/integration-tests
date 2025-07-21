package main

import (
	"context"
	"os"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// TestGiftCardInfo tests gift card information endpoints
func TestGiftCardInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetGiftCardRSAPublicKey", func(t *testing.T) {
		resp, httpResp, err := client.GiftCardAPI.GetGiftcardCryptographyRsaPublicKeyV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Gift card RSA public key") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if gift card not available on testnet
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Gift card API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get gift card RSA public key: %v", err)
		}

		t.Logf("Gift card RSA public key: %+v", resp)
	})

	t.Run("GetGiftCardBuyCodeTokenLimit", func(t *testing.T) {
		resp, httpResp, err := client.GiftCardAPI.GetGiftcardBuyCodeTokenLimitV1(ctx).
			BaseToken("USDT").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Gift card buy code token limit") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if gift card not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Gift card API not available on testnet")
					}
				}
			}
			t.Fatalf("Failed to get gift card buy code token limit: %v", err)
		}

		t.Logf("Gift card buy code token limit: %+v", resp)
	})

	t.Run("VerifyGiftCard", func(t *testing.T) {
		// This requires a gift card reference number
		referenceNo := os.Getenv("BINANCE_TEST_GIFT_CARD_REF")
		if referenceNo == "" {
			t.Skip("BINANCE_TEST_GIFT_CARD_REF not set")
		}

		resp, _, err := client.GiftCardAPI.GetGiftcardVerifyV1(ctx).
			ReferenceNo(referenceNo).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if gift card not found
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -5003 {
						t.Skip("Gift card not found")
					}
				}
			}
			t.Fatalf("Failed to verify gift card: %v", err)
		}

		t.Logf("Gift card verification result: %+v", resp)
	})
}

// TestGiftCardOperations tests gift card operations (use with caution)
func TestGiftCardOperations(t *testing.T) {
	// Skip by default to avoid creating actual gift cards
	if os.Getenv("BINANCE_TEST_GIFT_CARD_OPERATIONS") != "true" {
		t.Skip("Set BINANCE_TEST_GIFT_CARD_OPERATIONS=true to test gift card operations")
	}

	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	var createdGiftCardCode string

	t.Run("CreateGiftCard", func(t *testing.T) {
		resp, httpResp, err := client.GiftCardAPI.CreateGiftcardCreateCodeV1(ctx).
			Token("USDT").
			Amount(1.0).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create gift card") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or gift card not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -5002 || code == -1121) {
						t.Skip("Insufficient balance or gift card not available")
					}
				}
			}
			t.Fatalf("Failed to create gift card: %v", err)
		}

		t.Logf("Gift card created: %+v", resp)
		
		if resp.Code != nil {
			createdGiftCardCode = *resp.Code
		}
	})

	t.Run("BuyGiftCard", func(t *testing.T) {
		// This requires a product ID
		productId := os.Getenv("BINANCE_TEST_GIFT_CARD_PRODUCT_ID")
		if productId == "" {
			t.Skip("BINANCE_TEST_GIFT_CARD_PRODUCT_ID not set")
		}

		resp, _, err := client.GiftCardAPI.CreateGiftcardBuyCodeV1(ctx).
			BaseToken("USDT").
			FaceToken("USDT").
			BaseTokenAmount(10).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or product not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -5002 || code == -5004) {
						t.Skip("Insufficient balance or product not available")
					}
				}
			}
			t.Fatalf("Failed to buy gift card: %v", err)
		}

		t.Logf("Gift card bought: %+v", resp)
	})

	t.Run("RedeemGiftCard", func(t *testing.T) {
		// Use the created gift card code or a test code
		codeToRedeem := createdGiftCardCode
		if codeToRedeem == "" {
			codeToRedeem = os.Getenv("BINANCE_TEST_GIFT_CARD_CODE")
			if codeToRedeem == "" {
				t.Skip("No gift card code available to redeem")
			}
		}

		resp, _, err := client.GiftCardAPI.CreateGiftcardRedeemCodeV1(ctx).
			Code(codeToRedeem).
			Timestamp(timestamp).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if gift card already redeemed or invalid
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -5003 || code == -5005) {
						t.Skip("Gift card already redeemed or invalid")
					}
				}
			}
			t.Fatalf("Failed to redeem gift card: %v", err)
		}

		t.Logf("Gift card redeemed: %+v", resp)
	})
}