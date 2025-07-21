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
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// AuthType represents the type of authentication
type AuthType int

const (
	AuthTypeNONE AuthType = iota
	AuthTypeUSER_DATA
	AuthTypeTRADE
	AuthTypeMARGIN
	AuthTypeUSER_STREAM
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
	
	// Use testnet by default, but portfolio margin may not be available
	// Check if testnet is supported for portfolio margin
	if os.Getenv("BINANCE_PMARGIN_TESTNET_SUPPORTED") == "true" {
		cfg.Servers = openapi.ServerConfigurations{
			{
				URL:         "https://testnet.binance.vision",
				Description: "Binance Testnet",
			},
		}
	} else {
		// Use production server with caution
		cfg.Servers = openapi.ServerConfigurations{
			{
				URL:         "https://papi.binance.com",
				Description: "Binance Portfolio Margin Production",
			},
		}
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

// checkAPIError checks if an error is an API error and logs it
func checkAPIError(t *testing.T, err error, httpResp *http.Response) {
	if err == nil {
		return
	}

	// Special handling for 400 errors - print extra details
	if httpResp != nil && httpResp.StatusCode == 400 {
		t.Logf("🚨 400 BAD REQUEST ERROR DETAILS:")
		t.Logf("Status Code: %d", httpResp.StatusCode)
		t.Logf("Status: %s", httpResp.Status)
		
		// Log headers for 400 errors
		t.Logf("Response Headers:")
		for key, values := range httpResp.Header {
			for _, value := range values {
				t.Logf("  %s: %s", key, value)
			}
		}
	}

	// Always log the response body for any error
	logResponseBody(t, httpResp, "API Error")

	if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
		body := string(apiErr.Body())
		
		// Extra emphasis for 400 errors
		if httpResp != nil && httpResp.StatusCode == 400 {
			t.Logf("🚨 400 ERROR RESPONSE BODY: %s", body)
		} else {
			t.Logf("API Error Response: %s", body)
		}
		
		if model := apiErr.Model(); model != nil {
			if apiError, ok := model.(*openapi.APIError); ok {
				if apiError.Code != nil && apiError.Msg != nil {
					if httpResp != nil && httpResp.StatusCode == 400 {
						t.Logf("🚨 400 ERROR CODE: %d, MESSAGE: %s", *apiError.Code, *apiError.Msg)
					} else {
						t.Logf("Error Code: %d, Message: %s", *apiError.Code, *apiError.Msg)
					}
				}
			}
		}
	} else {
		if httpResp != nil && httpResp.StatusCode == 400 {
			t.Logf("🚨 400 NON-API ERROR: %v", err)
		} else {
			t.Logf("Non-API Error: %v", err)
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

	fmt.Printf("\n=== Running Binance Portfolio Margin REST API Integration Test Suite ===\n")
	fmt.Printf("Total tests to run: %d\n", len(suite.Tests))
	fmt.Printf("IMPORTANT: Portfolio Margin API requires special account setup\n")
	fmt.Printf("Many tests may be skipped if account is not enabled for portfolio margin\n\n")

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
		// General & System Tests
		{Name: "Ping", Function: TestPing, AuthRequired: AuthTypeNONE, Category: "General"},
		
		// Account Management Tests
		{Name: "Account Info", Function: TestAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Account Balance", Function: TestAccountBalance, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		
		// User Data Stream Tests
		{Name: "Listen Key Management", Function: TestListenKeyManagement, AuthRequired: AuthTypeUSER_STREAM, Category: "UserDataStream"},
		
		// Rate Limit Tests
		{Name: "Rate Limit Order", Function: TestRateLimitOrder, AuthRequired: AuthTypeUSER_DATA, Category: "RateLimit"},
		
		// Asset Collection & Transfer Tests
		{Name: "Asset Collection", Function: TestAssetCollection, AuthRequired: AuthTypeTRADE, Category: "AssetCollection"},
		{Name: "Auto Collection", Function: TestAutoCollection, AuthRequired: AuthTypeTRADE, Category: "AssetCollection"},
		{Name: "BNB Transfer", Function: TestBNBTransfer, AuthRequired: AuthTypeTRADE, Category: "AssetCollection"},
		
		// Repay & Negative Balance Tests
		{Name: "Repay Futures Negative Balance", Function: TestRepayFuturesNegativeBalance, AuthRequired: AuthTypeUSER_DATA, Category: "Repay"},
		{Name: "Repay Futures Switch", Function: TestRepayFuturesSwitch, AuthRequired: AuthTypeTRADE, Category: "Repay"},
		
		// Margin Trading Tests
		{Name: "Margin Loan", Function: TestMarginLoan, AuthRequired: AuthTypeMARGIN, Category: "MarginTrading"},
		{Name: "Margin Order", Function: TestMarginOrder, AuthRequired: AuthTypeTRADE, Category: "MarginTrading"},
		{Name: "Margin OCO Order", Function: TestMarginOCOOrder, AuthRequired: AuthTypeTRADE, Category: "MarginTrading"},
		{Name: "Margin Repay", Function: TestMarginRepay, AuthRequired: AuthTypeTRADE, Category: "MarginTrading"},
		{Name: "Margin Order Management", Function: TestMarginOrderManagement, AuthRequired: AuthTypeUSER_DATA, Category: "MarginTrading"},
		
		// UM Futures Tests (Note: These need to be implemented)
		// {Name: "UM Order", Function: TestUMOrder, AuthRequired: AuthTypeTRADE, Category: "UMFutures"},
		// {Name: "UM Conditional Order", Function: TestUMConditionalOrder, AuthRequired: AuthTypeTRADE, Category: "UMFutures"},
		// {Name: "UM Leverage", Function: TestUMLeverage, AuthRequired: AuthTypeTRADE, Category: "UMFutures"},
		// {Name: "UM Position", Function: TestUMPosition, AuthRequired: AuthTypeTRADE, Category: "UMFutures"},
		// {Name: "UM Account", Function: TestUMAccount, AuthRequired: AuthTypeUSER_DATA, Category: "UMFutures"},
		
		// CM Futures Tests (Note: These need to be implemented)
		// {Name: "CM Order", Function: TestCMOrder, AuthRequired: AuthTypeTRADE, Category: "CMFutures"},
		// {Name: "CM Conditional Order", Function: TestCMConditionalOrder, AuthRequired: AuthTypeTRADE, Category: "CMFutures"},
		// {Name: "CM Leverage", Function: TestCMLeverage, AuthRequired: AuthTypeTRADE, Category: "CMFutures"},
		// {Name: "CM Position", Function: TestCMPosition, AuthRequired: AuthTypeTRADE, Category: "CMFutures"},
		// {Name: "CM Account", Function: TestCMAccount, AuthRequired: AuthTypeUSER_DATA, Category: "CMFutures"},
		
		// Portfolio Margin Specific Tests
		{Name: "Portfolio Interest History", Function: TestPortfolioInterestHistory, AuthRequired: AuthTypeUSER_DATA, Category: "PortfolioMargin"},
		{Name: "Portfolio Negative Balance Exchange", Function: TestPortfolioNegativeBalanceExchange, AuthRequired: AuthTypeUSER_DATA, Category: "PortfolioMargin"},
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

// Note: logResponseBody is defined in testnet_helpers.go to avoid redeclaration

// getTestSymbol returns appropriate test symbol for the given market
func getTestSymbol(market string) string {
	switch strings.ToLower(market) {
	case "spot":
		if symbol := os.Getenv("BINANCE_TEST_SYMBOL_SPOT"); symbol != "" {
			return symbol
		}
		return "BTCUSDT"
	case "um", "umfutures":
		if symbol := os.Getenv("BINANCE_TEST_SYMBOL_UM"); symbol != "" {
			return symbol
		}
		return "BTCUSDT"
	case "cm", "cmfutures":
		if symbol := os.Getenv("BINANCE_TEST_SYMBOL_CM"); symbol != "" {
			return symbol
		}
		return "BTCUSD_PERP"
	default:
		return "BTCUSDT"
	}
}

// getTestOrderQuantity returns test order quantity
func getTestOrderQuantity() string {
	if qty := os.Getenv("BINANCE_TEST_ORDER_QUANTITY"); qty != "" {
		return qty
	}
	return "0.001"
}

// getTestOrderPrice returns test order price
func getTestOrderPrice() string {
	if price := os.Getenv("BINANCE_TEST_ORDER_PRICE"); price != "" {
		return price
	}
	return "30000"
}