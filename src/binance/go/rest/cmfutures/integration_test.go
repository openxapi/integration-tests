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
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
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
			Description: "Binance CM Futures Testnet",
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
			Description: "Binance CM Futures Testnet",
		},
	}
	publicClient := openapi.NewAPIClient(publicCfg)
	publicCtx := context.Background()
	
	req := publicClient.FuturesAPI.GetTickerPriceV1(publicCtx).Symbol(symbol)
	resp, _, err := req.Execute()
	if err != nil {
		return 0, err
	}

	if len(resp) == 0 {
		return 0, errors.New("no price data returned")
	}

	firstPrice := resp[0]
	if firstPrice.Price == nil {
		return 0, errors.New("no price returned")
	}

	return strconv.ParseFloat(*firstPrice.Price, 64)
}

// checkAPIError checks if an error is an API error and logs it
// NEVER skips 400 Bad Request errors - these need investigation
func checkAPIError(t *testing.T, err error, httpResp *http.Response, testName string) {
	if err == nil {
		return
	}

	// Always log response body for debugging
	if httpResp != nil {
		logResponseBody(t, httpResp, testName)
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

	// NEVER skip 400 errors - they indicate real API issues
	if httpResp != nil && httpResp.StatusCode == 400 {
		t.Logf("%s: 400 Bad Request - Response text: %s", testName, err.Error())
		t.Fatalf("%s: 400 Bad Request error requires investigation: %v", testName, err)
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

	fmt.Printf("\n=== Running Binance CM Futures REST API Integration Test Suite ===\n")
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
		// General/System API Tests
		{Name: "Ping", Function: TestPing, AuthRequired: AuthTypeNONE, Category: "General"},
		{Name: "Server Time", Function: TestServerTime, AuthRequired: AuthTypeNONE, Category: "General"},
		{Name: "Exchange Info", Function: TestExchangeInfo, AuthRequired: AuthTypeNONE, Category: "General"},
		
		// Market Data API Tests
		{Name: "Order Book Depth", Function: TestOrderBookDepth, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Aggregate Trades", Function: TestAggTrades, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Recent Trades", Function: TestRecentTrades, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Historical Trades", Function: TestHistoricalTrades, AuthRequired: AuthTypeUSER_DATA, Category: "MarketData"},
		{Name: "Klines", Function: TestKlines, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Continuous Klines", Function: TestContinuousKlines, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Index Price Klines", Function: TestIndexPriceKlines, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Mark Price Klines", Function: TestMarkPriceKlines, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Premium Index Klines", Function: TestPremiumIndexKlines, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "24hr Ticker", Function: Test24hrTicker, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Ticker Price", Function: TestTickerPrice, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Ticker Book", Function: TestTickerBook, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Premium Index", Function: TestPremiumIndex, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Funding Rate", Function: TestFundingRate, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Funding Info", Function: TestFundingInfo, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Open Interest", Function: TestOpenInterest, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Index Constituents", Function: TestIndexConstituents, AuthRequired: AuthTypeNONE, Category: "MarketData"},
		{Name: "Force Orders", Function: TestForceOrders, AuthRequired: AuthTypeUSER_DATA, Category: "MarketData"},
		
		// Trading API Tests
		{Name: "Create Order", Function: TestCreateOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Get Order", Function: TestGetOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Cancel Order", Function: TestCancelOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Update Order", Function: TestUpdateOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "All Orders", Function: TestAllOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Open Order", Function: TestOpenOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Open Orders", Function: TestOpenOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Cancel All Orders", Function: TestCancelAllOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Batch Orders", Function: TestBatchOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Batch Update Orders", Function: TestBatchUpdateOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Batch Cancel Orders", Function: TestBatchCancelOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Countdown Cancel All", Function: TestCountdownCancelAll, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Order Amendment", Function: TestOrderAmendment, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "User Trades", Function: TestUserTrades, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Commission Rate", Function: TestCommissionRate, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		
		// Account/Position Management Tests
		{Name: "Account Info", Function: TestAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Account Balance", Function: TestAccountBalance, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Position Risk", Function: TestPositionRisk, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Change Leverage", Function: TestChangeLeverage, AuthRequired: AuthTypeTRADE, Category: "Account"},
		{Name: "Leverage Bracket", Function: TestLeverageBracket, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Leverage Bracket V2", Function: TestLeverageBracketV2, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Change Margin Type", Function: TestChangeMarginType, AuthRequired: AuthTypeTRADE, Category: "Account"},
		{Name: "Position Margin", Function: TestPositionMargin, AuthRequired: AuthTypeTRADE, Category: "Account"},
		{Name: "Position Margin History", Function: TestPositionMarginHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Position Side Dual", Function: TestPositionSideDual, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Change Position Side Dual", Function: TestChangePositionSideDual, AuthRequired: AuthTypeTRADE, Category: "Account"},
		{Name: "ADL Quantile", Function: TestADLQuantile, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "PM Account Info", Function: TestPMAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		
		// Income/History Tests
		{Name: "Income History", Function: TestIncomeHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Income Async", Function: TestIncomeAsync, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Income Async Download", Function: TestIncomeAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Order Async", Function: TestOrderAsync, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Order Async Download", Function: TestOrderAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Trade Async", Function: TestTradeAsync, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		{Name: "Trade Async Download", Function: TestTradeAsyncDownload, AuthRequired: AuthTypeUSER_DATA, Category: "Income"},
		
		// User Data Stream Tests
		{Name: "Create Listen Key", Function: TestCreateListenKey, AuthRequired: AuthTypeUSER_DATA, Category: "Stream"},
		{Name: "Update Listen Key", Function: TestUpdateListenKey, AuthRequired: AuthTypeUSER_DATA, Category: "Stream"},
		{Name: "Delete Listen Key", Function: TestDeleteListenKey, AuthRequired: AuthTypeUSER_DATA, Category: "Stream"},
		
		// Futures Data Analytics Tests
		{Name: "Futures Data Basis", Function: TestFuturesDataBasis, AuthRequired: AuthTypeNONE, Category: "Analytics"},
		{Name: "Global Long Short Ratio", Function: TestGlobalLongShortRatio, AuthRequired: AuthTypeNONE, Category: "Analytics"},
		{Name: "Open Interest History", Function: TestOpenInterestHistory, AuthRequired: AuthTypeNONE, Category: "Analytics"},
		{Name: "Taker Buy Sell Volume", Function: TestTakerBuySellVolume, AuthRequired: AuthTypeNONE, Category: "Analytics"},
		{Name: "Top Trader Long Short Account Ratio", Function: TestTopTraderLongShortAccountRatio, AuthRequired: AuthTypeNONE, Category: "Analytics"},
		{Name: "Top Trader Long Short Position Ratio", Function: TestTopTraderLongShortPositionRatio, AuthRequired: AuthTypeNONE, Category: "Analytics"},
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
	
	// Report SDK issues found during testing
	sdkIssues := getSDKIssues()
	if len(sdkIssues) > 0 {
		fmt.Printf("\n🚨 SDK ISSUES DETECTED (%d total):\n", len(sdkIssues))
		fmt.Printf("These issues should be reported to the SDK maintainers:\n")
		for _, issue := range sdkIssues {
			fmt.Printf("  - %s\n", issue)
		}
		fmt.Printf("\nSDK Repository: ../binance-go/rest/cmfutures\n")
	}
	
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

