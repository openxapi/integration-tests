package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
)


// TestWalletTransferOperations tests wallet transfer endpoints
func TestWalletTransferOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetAssetTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAssetTransferV1(ctx).
			Type_("SPOT_TO_MARGIN").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Asset transfer history") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get asset transfer history: %v", err)
		}

		t.Logf("Asset transfer history: %+v", resp)
	})

	t.Run("GetFundingAsset", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.CreateAssetGetFundingAssetV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Funding asset") {
			return
		}
		if err != nil {
			logResponseBody(t, httpResp, "Get funding asset")
			t.Fatalf("Failed to get funding asset: %v", err)
		}

		t.Logf("Funding assets: %+v", resp)
	})

	t.Run("GetUserAsset", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.CreateAssetGetUserAssetV3(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "User assets") {
			return
		}
		if err != nil {
			logResponseBody(t, httpResp, "Get user assets")
			t.Fatalf("Failed to get user assets: %v", err)
		}

		t.Logf("User assets: %+v", resp)
	})

	t.Run("GetWalletBalance", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAssetWalletBalanceV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Wallet balance") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get wallet balance: %v", err)
		}

		t.Logf("Wallet balance: %+v", resp)
	})
}

// TestWalletDustOperations tests dust conversion endpoints
func TestWalletDustOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetDustLog", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAssetDribbletV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Dust log") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get dust log: %v", err)
		}

		t.Logf("Dust log: %+v", resp)
	})

	t.Run("ConvertDustToBNB", func(t *testing.T) {
		// Skip by default to avoid converting actual dust
		if os.Getenv("BINANCE_TEST_DUST_CONVERSION") != "true" {
			t.Skip("Set BINANCE_TEST_DUST_CONVERSION=true to test dust conversion")
		}

		// First get eligible dust assets
		dustResp, httpResp, err := client.WalletAPI.CreateAssetDustBtcV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Get dust info") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if no dust available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1001 {
						t.Skip("No dust available for conversion")
					}
				}
			}
			t.Fatalf("Failed to get dust info: %v", err)
		}

		// If we have dust, convert it
		if len(dustResp.Details) > 0 {
			var assets []string
			for _, detail := range dustResp.Details {
				if detail.Asset != nil {
					assets = append(assets, *detail.Asset)
				}
			}

			if len(assets) > 0 {
				convertResp, httpResp, err := client.WalletAPI.CreateAssetDustV1(ctx).
					Asset(assets).
					Timestamp(timestamp).
					Execute()

				if handleTestnetError(t, err, httpResp, "Convert dust") {
					return
				}
				if err != nil {
					t.Fatalf("Failed to convert dust: %v", err)
				}

				t.Logf("Dust converted: %+v", convertResp)
			}
		}
	})
}

// TestWalletDepositWithdrawOperations tests deposit/withdraw operations
func TestWalletDepositWithdrawOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetDepositAddressList", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetCapitalDepositAddressListV1(ctx).
			Coin("BTC").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Deposit address list") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			t.Fatalf("Failed to get deposit address list: %v", err)
		}

		t.Logf("Deposit address list: %+v", resp)
	})

	t.Run("GetWithdrawAddressList", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetCapitalWithdrawAddressListV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Withdraw address list") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get withdraw address list: %v", err)
		}

		t.Logf("Withdraw address list: %+v", resp)
	})

	t.Run("CreateWithdrawApplication", func(t *testing.T) {
		// Skip by default to avoid actual withdrawals
		if os.Getenv("BINANCE_TEST_WITHDRAWALS") != "true" {
			t.Skip("Set BINANCE_TEST_WITHDRAWALS=true to test withdrawals")
		}

		resp, httpResp, err := client.WalletAPI.CreateCapitalWithdrawApplyV1(ctx).
			Coin("USDT").
			Network("TRX").
			Address("TN8RmacHDQj27bWqWotCddajbrzfKTYJjh"). // Example address
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Create withdraw application") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -6023 {
						t.Skip("Insufficient balance for withdrawal")
					}
				}
			}
			logResponseBody(t, httpResp, "Create withdraw application")
			t.Fatalf("Failed to create withdraw application: %v", err)
		}

		t.Logf("Withdraw application created: %+v", resp)
	})
}

// TestWalletAccountRestrictions tests account restriction endpoints
func TestWalletAccountRestrictions(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetAccountAPIRestrictions", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAccountApiRestrictionsV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Account API restrictions") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get account API restrictions: %v", err)
		}

		t.Logf("Account API restrictions: %+v", resp)
	})

	t.Run("EnableFastWithdrawSwitch", func(t *testing.T) {
		// Skip by default as this changes account settings
		if os.Getenv("BINANCE_TEST_ACCOUNT_SETTINGS") != "true" {
			t.Skip("Set BINANCE_TEST_ACCOUNT_SETTINGS=true to test account settings")
		}

		resp, httpResp, err := client.WalletAPI.CreateAccountEnableFastWithdrawSwitchV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Enable fast withdraw switch") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			logResponseBody(t, httpResp, "Enable fast withdraw switch")
			t.Fatalf("Failed to enable fast withdraw switch: %v", err)
		}

		t.Logf("Fast withdraw switch enabled: %+v", resp)
	})

	t.Run("ToggleBNBBurn", func(t *testing.T) {
		// Skip by default as this changes account settings
		if os.Getenv("BINANCE_TEST_ACCOUNT_SETTINGS") != "true" {
			t.Skip("Set BINANCE_TEST_ACCOUNT_SETTINGS=true to test account settings")
		}

		resp, httpResp, err := client.WalletAPI.CreateBnbBurnV1(ctx).
			SpotBNBBurn("true").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Toggle BNB burn") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
			}
			logResponseBody(t, httpResp, "Toggle BNB burn")
			t.Fatalf("Failed to toggle BNB burn: %v", err)
		}

		t.Logf("BNB burn toggled: %+v", resp)
	})
}

// TestWalletAssetOperations tests asset-related endpoints
func TestWalletAssetOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("AssetTransfer", func(t *testing.T) {
		// Skip by default to avoid actual transfers
		if os.Getenv("BINANCE_TEST_ASSET_TRANSFER") != "true" {
			t.Skip("Set BINANCE_TEST_ASSET_TRANSFER=true to test asset transfers")
		}

		resp, httpResp, err := client.WalletAPI.CreateAssetTransferV1(ctx).
			Type_("SPOT_TO_MARGIN").
			Asset("USDT").
			Amount("1").
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Transfer asset") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if insufficient balance or margin not enabled
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -3020 || code == -11002) {
						t.Skip("Insufficient balance or margin not enabled")
					}
				}
			}
			logResponseBody(t, httpResp, "Transfer asset")
			t.Fatalf("Failed to transfer asset: %v", err)
		}

		t.Logf("Asset transferred: %+v", resp)
	})

	t.Run("GetAssetCustodyTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAssetCustodyTransferHistoryV1(ctx).
			Email("test@example.com").
			StartTime(time.Now().Add(-30*24*time.Hour).UnixMilli()).
			EndTime(time.Now().UnixMilli()).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Custody transfer history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if custody features not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Custody features not available")
					}
				}
			}
			t.Fatalf("Failed to get custody transfer history: %v", err)
		}

		t.Logf("Custody transfer history: %+v", resp)
	})

	t.Run("GetCloudMiningTransferHistory", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetAssetLedgerTransferCloudMiningQueryByPageV1(ctx).
			StartTime(time.Now().Add(-30*24*time.Hour).UnixMilli()).
			EndTime(time.Now().UnixMilli()).
			Execute()

		if handleTestnetError(t, err, httpResp, "Cloud mining transfer history") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if cloud mining not available
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && code == -1121 {
						t.Skip("Cloud mining features not available")
					}
				}
			}
			t.Fatalf("Failed to get cloud mining transfer history: %v", err)
		}

		t.Logf("Cloud mining transfer history: %+v", resp)
	})
}

// TestWalletSpotInfo tests spot trading related wallet endpoints
func TestWalletSpotInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("GetDelistSchedule", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetSpotDelistScheduleV1(ctx).
			Timestamp(timestamp).
			Execute()

		if handleTestnetError(t, err, httpResp, "Delist schedule") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get delist schedule: %v", err)
		}

		t.Logf("Delist schedule: %+v", resp)
	})

	t.Run("GetOpenSymbolList", func(t *testing.T) {
		resp, httpResp, err := client.WalletAPI.GetSpotOpenSymbolListV1(ctx).
			Execute()

		if handleTestnetError(t, err, httpResp, "Open symbol list") {
			return
		}
		if err != nil {
			t.Fatalf("Failed to get open symbol list: %v", err)
		}

		t.Logf("Open symbol list has %d symbols", len(resp))
		if len(resp) > 0 && len(resp) <= 5 {
			// Show first few symbols if not too many
			t.Logf("Symbols: %+v", resp)
		}
	})
}

// TestWalletDepositCreditOperations tests deposit credit endpoints (for VIP users)
func TestWalletDepositCreditOperations(t *testing.T) {
	// Skip by default as these are VIP features
	if os.Getenv("BINANCE_TEST_VIP_FEATURES") != "true" {
		t.Skip("Set BINANCE_TEST_VIP_FEATURES=true to test VIP features")
	}

	client := getTestClient(t)
	ctx := context.Background()

	t.Run("ApplyDepositCredit", func(t *testing.T) {
		// Skip this test as it requires a valid deposit ID
		t.Skip("Deposit credit apply requires valid deposit ID - skipping test")
		
		resp, _, err := client.WalletAPI.CreateCapitalDepositCreditApplyV1(ctx).
			Execute()

		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Skip if not eligible for deposit credit
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok && (code == -1121 || code == -2015) {
						t.Skip("Deposit credit not available for this account")
					}
				}
			}
			t.Fatalf("Failed to apply deposit credit: %v", err)
		}

		t.Logf("Deposit credit applied: %+v", resp)
	})
}

// TestTravelRuleWithdraw tests the Travel Rule withdraw endpoint
func TestTravelRuleWithdraw(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	timestamp := time.Now().UnixMilli()

	t.Run("CreateTravelRuleWithdraw", func(t *testing.T) {
		// Skip by default as this creates a real withdrawal request
		if os.Getenv("BINANCE_TEST_TRAVEL_RULE_WITHDRAW") != "true" {
			t.Skip("Set BINANCE_TEST_TRAVEL_RULE_WITHDRAW=true to test Travel Rule withdraw (CAUTION: Creates real withdrawal)")
		}

		// Example questionnaire - format varies by local entity
		questionnaire := `{"questions":[{"question":"What is the purpose of this withdrawal?","answer":"Test withdrawal"}]}`
		
		resp, httpResp, err := client.WalletAPI.CreateLocalentityWithdrawApplyV1(ctx).
			Coin("USDT").
			Address("TExampleAddress123456789"). // Use a valid address for testing
			Amount("10.0"). // Small test amount
			Questionnaire(questionnaire).
			Timestamp(timestamp).
			RecvWindow(5000).
			Execute()

		if handleTestnetError(t, err, httpResp, "Travel Rule withdraw") {
			return
		}
		if err != nil {
			apiErr, ok := err.(*openapi.GenericOpenAPIError)
			if ok {
				t.Logf("API error response: %s", string(apiErr.Body()))
				// Check for common error codes
				var errResp map[string]interface{}
				if jsonErr := parseJSON(apiErr.Body(), &errResp); jsonErr == nil {
					if code, ok := errResp["code"].(float64); ok {
						switch code {
						case -1000:
							t.Skip("Travel Rule withdraw endpoint unavailable on testnet")
						case -4001:
							t.Skip("Invalid address for Travel Rule withdraw")
						case -4026:
							t.Skip("Insufficient balance for Travel Rule withdraw")
						case -100001:
							t.Skip("Travel Rule not applicable for this withdrawal")
						case -100002:
							t.Skip("Invalid questionnaire format")
						default:
							t.Logf("Travel Rule withdraw error code: %v", code)
						}
					}
				}
			}
			t.Fatalf("Failed to create Travel Rule withdraw: %v", err)
		}

		// Verify response
		if resp.TrId != nil {
			t.Logf("Travel Rule withdraw ID: %d", *resp.TrId)
		}
		if resp.Accepted != nil {  // Note: Field has a typo in the SDK
			t.Logf("Travel Rule withdraw accepted: %v", *resp.Accepted)
		}
		if resp.Info != nil {
			t.Logf("Travel Rule withdraw info: %s", *resp.Info)
		}
		
		t.Logf("Travel Rule withdraw response: %+v", resp)
	})

	t.Run("VerifyQuestionnaireNotURLEncoded", func(t *testing.T) {
		// This test verifies that the questionnaire field is NOT URL-encoded
		// It should be sent as: questionnaire={"isAddressOwner":1,"bnfType":0...}
		// NOT as: questionnaire=%7B%22isAddressOwner%22%3A1%2C%22bnfType%22...
		
		// Create a custom HTTP client with request interceptor
		var capturedBody string
		interceptClient := &http.Client{
			Transport: &requestInterceptor{
				RoundTripper: http.DefaultTransport,
				interceptFunc: func(req *http.Request) {
					if req.Body != nil {
						bodyBytes, _ := io.ReadAll(req.Body)
						capturedBody = string(bodyBytes)
						// Restore the body for the actual request
						req.Body = io.NopCloser(strings.NewReader(capturedBody))
					}
				},
			},
		}
		
		// Create configuration with custom client
		cfg := openapi.NewConfiguration()
		cfg.HTTPClient = interceptClient
		if os.Getenv("BINANCE_TESTNET") == "true" {
			cfg.Servers = openapi.ServerConfigurations{
				{URL: "https://testnet.binance.vision"},
			}
		}
		
		// Set auth
		if apiKey := os.Getenv("BINANCE_ED25519_API_KEY"); apiKey != "" {
			cfg.AddDefaultHeader("X-MBX-APIKEY", apiKey)
		} else if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
			cfg.AddDefaultHeader("X-MBX-APIKEY", apiKey)
		}
		
		customClient := openapi.NewAPIClient(cfg)
		
		// Complex questionnaire to test encoding
		questionnaire := `{"isAddressOwner":1,"bnfType":0,"sendTo":2,"vasp":"VASP","vaspName":"VASP","declaration":true}`
		
		_, _, _ = customClient.WalletAPI.CreateLocalentityWithdrawApplyV1(ctx).
			Coin("testCoin").
			Address("testaddr").
			Amount("10").
			Network("BTC").
			Name("testlabel").
			WithdrawOrderId("testID").
			Questionnaire(questionnaire).
			Timestamp(timestamp).
			Execute()
		
		// Verify the questionnaire is NOT URL-encoded in the request body
		if capturedBody != "" {
			t.Logf("Captured request body: %s", capturedBody)
			t.Logf("Questionnaire value to send: %s", questionnaire)
			
			// Parse the captured body to check questionnaire value
			params, _ := url.ParseQuery(capturedBody)
			actualQuestionnaire := params.Get("questionnaire")
			t.Logf("Parsed questionnaire value: '%s'", actualQuestionnaire)
			
			// Check that the JSON is NOT URL-encoded
			if strings.Contains(capturedBody, "questionnaire=%7B") {
				t.Error("Questionnaire is URL-encoded but should NOT be!")
				t.Logf("Expected format: questionnaire={\"isAddressOwner\":1,\"bnfType\":0...}")
				t.Logf("Got URL-encoded: questionnaire=%%7B%%22isAddressOwner%%22...")
			}
			
			// Verify the questionnaire contains raw JSON characters
			if !strings.Contains(capturedBody, `questionnaire={"isAddressOwner":1`) {
				t.Error("Questionnaire should contain raw JSON without URL encoding")
				t.Logf("Expected to find: questionnaire={\"isAddressOwner\":1")
				t.Logf("Actual body: %s", capturedBody)
			}
			
			// Additional check for specific characters that should NOT be encoded
			encodedChars := map[string]string{
				"%7B": "{",
				"%7D": "}",
				"%22": "\"",
				"%3A": ":",
				"%2C": ",",
			}
			
			for encoded, decoded := range encodedChars {
				if strings.Contains(capturedBody, "questionnaire=") {
					questionnaireStart := strings.Index(capturedBody, "questionnaire=")
					questionnaireEnd := strings.Index(capturedBody[questionnaireStart:], "&")
					if questionnaireEnd == -1 {
						questionnaireEnd = len(capturedBody) - questionnaireStart
					}
					questionnaireValue := capturedBody[questionnaireStart : questionnaireStart+questionnaireEnd]
					
					if strings.Contains(questionnaireValue, encoded) {
						t.Errorf("Found URL-encoded character %s (for '%s') in questionnaire, but it should be raw", encoded, decoded)
					}
				}
			}
			
			t.Log("✓ Questionnaire format validation complete")
		} else {
			t.Log("Note: Could not capture request body (request may have been skipped)")
		}
	})
}

// requestInterceptor is a custom RoundTripper that intercepts HTTP requests
type requestInterceptor struct {
	http.RoundTripper
	interceptFunc func(*http.Request)
}

func (ri *requestInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	if ri.interceptFunc != nil {
		ri.interceptFunc(req)
	}
	// For testing purposes, we can return a mock response to avoid actual API calls
	if strings.Contains(req.URL.Path, "/sapi/v1/localentity/withdraw/apply") {
		// Return a mock response for testing
		return &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(strings.NewReader(`{"code":-1000,"msg":"Test mode - request intercepted"}`)),
			Header:     make(http.Header),
		}, nil
	}
	return ri.RoundTripper.RoundTrip(req)
}