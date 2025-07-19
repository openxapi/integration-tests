package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/umfutures"
)

// AuthType represents the type of authentication
type AuthType int

const (
	AuthTypeNONE AuthType = iota
	AuthTypeUSER_DATA
	AuthTypeTRADE
)

// TestConfig holds configuration for a test
type TestConfig struct {
	Name         string
	APIKey       string
	SecretKey    string
	SignType     string
	AuthType     AuthType
	TestFunction func(t *testing.T, client *openapi.APIClient, config TestConfig)
}

// TestInfo holds information about a test
type TestInfo struct {
	Name         string
	Function     func(t *testing.T)
	AuthRequired AuthType
	Category     string
}

// TestSuite manages test execution and results
type TestSuite struct {
	Tests       []TestInfo
	Results     map[string]TestResult
	StartTime   time.Time
	EndTime     time.Time
	TotalTests  int
	PassedTests int
	FailedTests int
}

// TestResult holds the result of a single test
type TestResult struct {
	Passed   bool
	Duration time.Duration
	Error    error
}

// getTestConfigs returns test configurations - by default uses Ed25519 for authenticated tests
func getTestConfigs() []TestConfig {
	var configs []TestConfig

	// Check if we should test all auth types (for comprehensive testing)
	testAllAuth := os.Getenv("TEST_ALL_AUTH_TYPES") == "true"

	if testAllAuth {
		// HMAC configuration
		if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
			if secretKey := os.Getenv("BINANCE_SECRET_KEY"); secretKey != "" {
				configs = append(configs, TestConfig{
					Name:      "HMAC Authentication",
					APIKey:    apiKey,
					SecretKey: secretKey,
					SignType:  "HMAC",
					AuthType:  AuthTypeTRADE,
				})
			}
		}

		// RSA configuration
		if apiKey := os.Getenv("BINANCE_RSA_API_KEY"); apiKey != "" {
			if keyPath := os.Getenv("BINANCE_RSA_PRIVATE_KEY_PATH"); keyPath != "" {
				configs = append(configs, TestConfig{
					Name:     "RSA Authentication",
					APIKey:   apiKey,
					SignType: "RSA",
					AuthType: AuthTypeTRADE,
				})
			}
		}
	}

	// Ed25519 configuration (default for authenticated tests)
	if apiKey := os.Getenv("BINANCE_ED25519_API_KEY"); apiKey != "" {
		if keyPath := os.Getenv("BINANCE_ED25519_PRIVATE_KEY_PATH"); keyPath != "" {
			configs = append(configs, TestConfig{
				Name:     "Ed25519 Authentication",
				APIKey:   apiKey,
				SignType: "Ed25519",
				AuthType: AuthTypeTRADE,
			})
		}
	} else if !testAllAuth {
		// Fallback to HMAC if Ed25519 not available and not testing all auth types
		if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
			if secretKey := os.Getenv("BINANCE_SECRET_KEY"); secretKey != "" {
				configs = append(configs, TestConfig{
					Name:      "HMAC Authentication",
					APIKey:    apiKey,
					SecretKey: secretKey,
					SignType:  "HMAC",
					AuthType:  AuthTypeTRADE,
				})
			}
		}
	}

	// Add a config for public endpoints (no auth)
	configs = append(configs, TestConfig{
		Name:     "Public Endpoints",
		AuthType: AuthTypeNONE,
	})

	return configs
}

// setupClient creates and configures a REST API client
func setupClient(config TestConfig) (*openapi.APIClient, context.Context) {
	cfg := openapi.NewConfiguration()
	
	// Use testnet by default
	cfg.Servers = openapi.ServerConfigurations{
		{
			URL:         "https://testnet.binancefuture.com",
			Description: "Binance USD-M Futures Testnet",
		},
	}

	// Create client
	client := openapi.NewAPIClient(cfg)
	ctx := context.Background()

	// Add authentication if needed
	if config.AuthType != AuthTypeNONE && config.APIKey != "" {
		auth := &openapi.Auth{
			APIKey: config.APIKey,
		}

		switch config.SignType {
		case "HMAC":
			auth.SetSecretKey(config.SecretKey)
		case "RSA":
			if keyPath := os.Getenv("BINANCE_RSA_PRIVATE_KEY_PATH"); keyPath != "" {
				auth.PrivateKeyPath = keyPath
			}
		case "Ed25519":
			if keyPath := os.Getenv("BINANCE_ED25519_PRIVATE_KEY_PATH"); keyPath != "" {
				auth.PrivateKeyPath = keyPath
			}
		}

		// Use ContextWithValue to properly initialize the authentication
		authCtx, err := auth.ContextWithValue(ctx)
		if err != nil {
			// If authentication setup fails, fall back to the old method
			// This ensures tests still run even if key files are missing
			ctx = context.WithValue(ctx, openapi.ContextBinanceAuth, *auth)
		} else {
			ctx = authCtx
		}
	}

	return client, ctx
}

// loadRSAPrivateKey loads an RSA private key from file
func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

// loadEd25519PrivateKey loads an Ed25519 private key from file
func loadEd25519PrivateKey(path string) (ed25519.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try PEM format first
	block, _ := pem.Decode(keyData)
	if block != nil {
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			if ed25519Key, ok := privateKey.(ed25519.PrivateKey); ok {
				return ed25519Key, nil
			}
		}
	}

	// Try raw hex format
	keyStr := strings.TrimSpace(string(keyData))
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(keyBytes), nil
	}

	return nil, errors.New("invalid Ed25519 private key format")
}

// testEndpoint is a helper to test an endpoint with proper setup and teardown
func testEndpoint(t *testing.T, config TestConfig, testName string, testFunc func(*testing.T, *openapi.APIClient, context.Context)) {
	rateLimiter.WaitForRateLimit()

	client, ctx := setupClient(config)
	
	// Create a context with timeout for the HTTP requests
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Run test function directly - t.Fatal will properly fail the test immediately
	testFunc(t, client, timeoutCtx)
}

// getCurrentPrice fetches the current price for a symbol using a public (non-authenticated) client
func getCurrentPrice(client *openapi.APIClient, ctx context.Context, symbol string) (float64, error) {
	// Create a new public client without authentication for this public endpoint
	publicCfg := openapi.NewConfiguration()
	publicCfg.Servers = openapi.ServerConfigurations{
		{
			URL:         "https://testnet.binancefuture.com",
			Description: "Binance USD-M Futures Testnet",
		},
	}
	publicClient := openapi.NewAPIClient(publicCfg)
	publicCtx := context.Background()
	
	req := publicClient.FuturesAPI.GetTickerPriceV1(publicCtx)
	if symbol != "" {
		req = req.Symbol(symbol)
	}
	resp, _, err := req.Execute()
	if err != nil {
		return 0, err
	}

	// Handle both single item and array response
	if resp.UmfuturesGetTickerPriceV1RespItem != nil {
		item := resp.UmfuturesGetTickerPriceV1RespItem
		if item.Symbol != nil && *item.Symbol == symbol && item.Price != nil {
			return strconv.ParseFloat(*item.Price, 64)
		}
	}

	if resp.ArrayOfUmfuturesGetTickerPriceV1RespItem != nil {
		for _, item := range *resp.ArrayOfUmfuturesGetTickerPriceV1RespItem {
			if item.Symbol != nil && *item.Symbol == symbol && item.Price != nil {
				return strconv.ParseFloat(*item.Price, 64)
			}
		}
	}

	return 0, errors.New("no price returned")
}

// checkAPIError checks if an error is an API error and logs it
func checkAPIError(t *testing.T, err error) {
	if err == nil {
		return
	}

	if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
		t.Logf("API Error: %s", string(apiErr.Body()))
		
		if model := apiErr.Model(); model != nil {
			if apiError, ok := model.(*openapi.APIError); ok {
				if apiError.Code != nil && apiError.Msg != nil {
					t.Logf("Error Code: %d, Message: %s", *apiError.Code, *apiError.Msg)
				}
			}
		}
	}
}

// generateTimestamp generates a timestamp for API requests
func generateTimestamp() int64 {
	return time.Now().UnixMilli()
}

// TestFullIntegrationSuite runs all integration tests
func TestFullIntegrationSuite(t *testing.T) {
	suite := &TestSuite{
		Results:   make(map[string]TestResult),
		StartTime: time.Now(),
	}

	// Initialize all tests
	suite.initializeTests()

	fmt.Printf("\n=== Running Binance USD-M Futures REST API Integration Test Suite ===\n")
	fmt.Printf("Total tests to run: %d\n\n", len(suite.Tests))

	// Run all tests using proper t.Run subtests
	for _, test := range suite.Tests {
		suite.TotalTests++
		
		// Check if we have necessary auth for this test
		configs := getTestConfigs()
		canRun := false
		
		if test.AuthRequired == AuthTypeNONE {
			canRun = true
		} else {
			for _, config := range configs {
				if config.AuthType >= test.AuthRequired {
					canRun = true
					break
				}
			}
		}

		if !canRun {
			fmt.Printf("⚠️  SKIP %s - No authentication configured\n", test.Name)
			suite.Results[test.Name] = TestResult{
				Passed:   false,
				Duration: 0,
				Error:    errors.New("skipped - no authentication"),
			}
			continue
		}

		// Use proper subtest
		testName := test.Name
		testFunction := test.Function
		
		success := t.Run(testName, func(subT *testing.T) {
			testStart := time.Now()
			fmt.Printf("▶️  Running %s...", testName)
			
			defer func() {
				duration := time.Since(testStart)
				if subT.Failed() {
					suite.FailedTests++
					fmt.Printf(" ❌ FAILED (%.2fs)\n", duration.Seconds())
					suite.Results[testName] = TestResult{
						Passed:   false,
						Duration: duration,
						Error:    errors.New("test failed"),
					}
				} else {
					suite.PassedTests++
					fmt.Printf(" ✅ PASSED (%.2fs)\n", duration.Seconds())
					suite.Results[testName] = TestResult{
						Passed:   true,
						Duration: duration,
						Error:    nil,
					}
				}
			}()
			
			testFunction(subT)
		})
		
		if !success {
			suite.FailedTests++
		}
	}

	suite.EndTime = time.Now()
	suite.printSummary()
}

// initializeTests sets up all test cases
func (suite *TestSuite) initializeTests() {
	suite.Tests = []TestInfo{
		// Public API Tests
		{Name: "Exchange Info", Function: TestExchangeInfo, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Server Time", Function: TestServerTime, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ping", Function: TestPing, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Order Book", Function: TestOrderBook, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Recent Trades", Function: TestRecentTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Agg Trades", Function: TestAggTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Klines", Function: TestKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "24hr Ticker", Function: Test24hrTicker, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Price Ticker", Function: TestPriceTicker, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Book Ticker", Function: TestBookTicker, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Open Interest", Function: TestOpenInterest, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Premium Index", Function: TestPremiumIndex, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Funding Rate", Function: TestFundingRate, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Funding Info", Function: TestFundingInfo, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Index Info", Function: TestIndexInfo, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Constituents", Function: TestConstituents, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Asset Index", Function: TestAssetIndex, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Continuous Klines", Function: TestContinuousKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Index Price Klines", Function: TestIndexPriceKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Mark Price Klines", Function: TestMarkPriceKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Premium Index Klines", Function: TestPremiumIndexKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Historical Trades", Function: TestHistoricalTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		
		// Futures Data API Tests (TODO: Implement these tests)
		// {Name: "Futures Data Basis", Function: TestFuturesDataBasis, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Delivery Price", Function: TestFuturesDataDeliveryPrice, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Open Interest", Function: TestFuturesDataOpenInterest, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Long Short Ratio", Function: TestFuturesDataLongShortRatio, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Taker Volume", Function: TestFuturesDataTakerVolume, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Top Trader Account Ratio", Function: TestFuturesDataTopTraderAccountRatio, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		// {Name: "Futures Data Top Trader Position Ratio", Function: TestFuturesDataTopTraderPositionRatio, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		
		// TODO: Implement remaining tests
		// Convert API Tests
		// {Name: "Convert Exchange Info", Function: TestConvertExchangeInfo, AuthRequired: AuthTypeNONE, Category: "Convert"},
		// {Name: "Convert Get Quote", Function: TestConvertGetQuote, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		// {Name: "Convert Accept Quote", Function: TestConvertAcceptQuote, AuthRequired: AuthTypeTRADE, Category: "Convert"},
		// {Name: "Convert Order Status", Function: TestConvertOrderStatus, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		
		// Account API Tests
		// {Name: "Account Info V2", Function: TestAccountInfoV2, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Account Info V3", Function: TestAccountInfoV3, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Account Balance V2", Function: TestAccountBalanceV2, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Account Balance V3", Function: TestAccountBalanceV3, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Account Config", Function: TestAccountConfig, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Position Risk V2", Function: TestPositionRiskV2, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Position Risk V3", Function: TestPositionRiskV3, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "User Trades", Function: TestUserTrades, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "All Orders", Function: TestAllOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Open Orders", Function: TestOpenOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Income History", Function: TestIncomeHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Force Orders", Function: TestForceOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "ADL Quantile", Function: TestADLQuantile, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Commission Rate", Function: TestCommissionRate, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "API Trading Status", Function: TestAPITradingStatus, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Symbol Config", Function: TestSymbolConfig, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Leverage Bracket", Function: TestLeverageBracket, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Position Side Dual", Function: TestPositionSideDual, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Multi Assets Margin", Function: TestMultiAssetsMargin, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Fee Burn", Function: TestFeeBurn, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Position Margin History", Function: TestPositionMarginHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Order Amendment", Function: TestOrderAmendment, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "Rate Limit Order", Function: TestRateLimitOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "PM Account Info", Function: TestPMAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		
		// Trading API Tests
		// {Name: "Create Order", Function: TestCreateOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Create Order Test", Function: TestCreateOrderTest, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Query Order", Function: TestQueryOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		// {Name: "Cancel Order", Function: TestCancelOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Cancel All Orders", Function: TestCancelAllOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Modify Order", Function: TestModifyOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Batch Orders", Function: TestBatchOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Change Leverage", Function: TestChangeLeverage, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Change Margin Type", Function: TestChangeMarginType, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Position Margin", Function: TestPositionMargin, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Change Position Mode", Function: TestChangePositionMode, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Change Multi Assets Margin", Function: TestChangeMultiAssetsMargin, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Change Fee Burn", Function: TestChangeFeeBurn, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		// {Name: "Countdown Cancel All", Function: TestCountdownCancelAll, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		
		// User Data Stream Tests
		// {Name: "User Data Stream", Function: TestUserDataStream, AuthRequired: AuthTypeUSER_DATA, Category: "Stream"},
		
		// BinanceLink API Tests
		// {Name: "API Referral Overview", Function: TestAPIReferralOverview, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral If New User", Function: TestAPIReferralIfNewUser, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral Customization", Function: TestAPIReferralCustomization, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral User Customization", Function: TestAPIReferralUserCustomization, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral Rebate Volume", Function: TestAPIReferralRebateVolume, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral Trade Volume", Function: TestAPIReferralTradeVolume, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral Trader Number", Function: TestAPIReferralTraderNumber, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		// {Name: "API Referral Trader Summary", Function: TestAPIReferralTraderSummary, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		
		// Async Download Tests
		// {Name: "Income Async Download", Function: TestIncomeAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Async"},
		// {Name: "Order Async Download", Function: TestOrderAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Async"},
		// {Name: "Trade Async Download", Function: TestTradeAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Async"},
	}
}

// printSummary prints the test suite summary
func (suite *TestSuite) printSummary() {
	fmt.Printf("\n=== Test Suite Summary ===\n")
	fmt.Printf("Total Duration: %.2fs\n", suite.EndTime.Sub(suite.StartTime).Seconds())
	fmt.Printf("Total Tests: %d\n", suite.TotalTests)
	fmt.Printf("Passed: %d\n", suite.PassedTests)
	fmt.Printf("Failed: %d\n", suite.FailedTests)
	fmt.Printf("Total API Requests: %d\n", rateLimiter.GetRequestCount())
	
	if suite.FailedTests > 0 {
		fmt.Printf("\n❌ Failed Tests:\n")
		for name, result := range suite.Results {
			if !result.Passed {
				fmt.Printf("  - %s", name)
				if result.Error != nil {
					fmt.Printf(" (Error: %v)", result.Error)
				}
				fmt.Println()
			}
		}
	}
	
	fmt.Printf("\n")
}

// getTestClient returns a configured test client for individual test files
func getTestClient(t *testing.T) *openapi.APIClient {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			client, _ := setupClient(config)
			return client
		}
	}
	t.Fatal("No authenticated client available")
	return nil
}

// parseJSON is a helper to parse JSON responses
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}