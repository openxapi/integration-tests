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
	"net/http"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/spot"
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
			URL:         "https://testnet.binance.vision",
			Description: "Binance Testnet",
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
			URL:         "https://testnet.binance.vision",
			Description: "Binance Testnet",
		},
	}
	publicClient := openapi.NewAPIClient(publicCfg)
	publicCtx := context.Background()
	
	req := publicClient.SpotTradingAPI.GetAvgPriceV3(publicCtx).Symbol(symbol)
	resp, _, err := req.Execute()
	if err != nil {
		return 0, err
	}

	if resp.Price == nil {
		return 0, errors.New("no price returned")
	}

	return strconv.ParseFloat(*resp.Price, 64)
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

	fmt.Printf("\n=== Running Binance Spot REST API Integration Test Suite ===\n")
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
		{Name: "Market Depth", Function: TestMarketDepth, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Recent Trades", Function: TestRecentTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Klines", Function: TestKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "24hr Ticker", Function: Test24hrTicker, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Average Price", Function: TestAveragePrice, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ping", Function: TestPing, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Agg Trades", Function: TestAggTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Historical Trades", Function: TestHistoricalTrades, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ticker 24hr", Function: TestTicker24hr, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ticker Price", Function: TestTickerPrice, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ticker Book", Function: TestTickerBookTicker, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "Ticker Trading Day", Function: TestTickerTradingDay, AuthRequired: AuthTypeNONE, Category: "Public"},
		{Name: "UI Klines", Function: TestUiKlines, AuthRequired: AuthTypeNONE, Category: "Public"},
		
		// Account API Tests
		{Name: "Account Info", Function: TestAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Account Commission", Function: TestAccountCommission, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Trade Fee", Function: TestTradeFee, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		// {Name: "API Key Permissions", Function: TestAPIKeyPermissions, AuthRequired: AuthTypeUSER_DATA, Category: "Account"}, // Commented out in account_test.go
		{Name: "Account Status", Function: TestAccountStatus, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		{Name: "Rate Limit Order", Function: TestRateLimitOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Account"},
		
		// Trading API Tests
		{Name: "Create Order", Function: TestCreateOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Query Order", Function: TestQueryOrder, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Cancel Order", Function: TestCancelOrder, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "All Orders", Function: TestAllOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "My Trades", Function: TestMyTrades, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Create Order Test", Function: TestCreateOrderTest, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "Open Orders", Function: TestOpenOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Delete Open Orders", Function: TestDeleteOpenOrders, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		{Name: "My Prevented Matches", Function: TestMyPreventedMatches, AuthRequired: AuthTypeUSER_DATA, Category: "Trading"},
		{Name: "Order Cancel Replace", Function: TestOrderCancelReplace, AuthRequired: AuthTypeTRADE, Category: "Trading"},
		
		// OCO Trading Tests
		{Name: "Create Order OCO", Function: TestCreateOrderOco, AuthRequired: AuthTypeTRADE, Category: "OCO"},
		{Name: "Create Order List OCO", Function: TestCreateOrderListOco, AuthRequired: AuthTypeTRADE, Category: "OCO"},
		{Name: "Get Order List", Function: TestGetOrderList, AuthRequired: AuthTypeUSER_DATA, Category: "OCO"},
		{Name: "Get Open Order List", Function: TestGetOpenOrderList, AuthRequired: AuthTypeUSER_DATA, Category: "OCO"},
		{Name: "Get All Order List", Function: TestGetAllOrderList, AuthRequired: AuthTypeUSER_DATA, Category: "OCO"},
		{Name: "Create Order List OTO", Function: TestCreateOrderListOto, AuthRequired: AuthTypeTRADE, Category: "OCO"},
		{Name: "Create Order List OTOCO", Function: TestCreateOrderListOtoco, AuthRequired: AuthTypeTRADE, Category: "OCO"},
		
		// SOR Trading Tests
		{Name: "Create SOR Order", Function: TestCreateSorOrder, AuthRequired: AuthTypeTRADE, Category: "SOR"},
		{Name: "Create SOR Order Test", Function: TestCreateSorOrderTest, AuthRequired: AuthTypeTRADE, Category: "SOR"},
		{Name: "Get My Allocations", Function: TestGetMyAllocations, AuthRequired: AuthTypeUSER_DATA, Category: "SOR"},
		
		// Wallet API Tests
		{Name: "System Status", Function: TestGetSystemStatus, AuthRequired: AuthTypeNONE, Category: "Wallet"},
		{Name: "Capital Config", Function: TestGetCapitalConfigGetall, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Account Info", Function: TestWalletAccountInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Asset Detail", Function: TestGetAssetDetail, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Deposit History", Function: TestGetDepositHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Withdraw History", Function: TestGetWithdrawHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Deposit Address", Function: TestGetDepositAddress, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Account Snapshot", Function: TestGetAccountSnapshot, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Asset Dividend", Function: TestGetAssetDividend, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Disable Fast Withdraw", Function: TestDisableFastWithdraw, AuthRequired: AuthTypeTRADE, Category: "Wallet"},
		{Name: "API Trading Status", Function: TestGetAPITradingStatus, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Transfer Operations", Function: TestWalletTransferOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Dust Operations", Function: TestWalletDustOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Deposit/Withdraw Operations", Function: TestWalletDepositWithdrawOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Account Restrictions", Function: TestWalletAccountRestrictions, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Asset Operations", Function: TestWalletAssetOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Spot Info", Function: TestWalletSpotInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Wallet"},
		{Name: "Wallet Deposit Credit Operations", Function: TestWalletDepositCreditOperations, AuthRequired: AuthTypeTRADE, Category: "Wallet"},
		{Name: "Travel Rule Withdraw", Function: TestTravelRuleWithdraw, AuthRequired: AuthTypeTRADE, Category: "Wallet"},
		
		// User Data Stream Tests
		{Name: "User Data Stream", Function: TestUserDataStream, AuthRequired: AuthTypeUSER_DATA, Category: "Stream"},
		
		// Margin Trading Tests
		{Name: "Margin Account", Function: TestGetMarginAccount, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin All Assets", Function: TestGetMarginAllAssets, AuthRequired: AuthTypeNONE, Category: "Margin"},
		{Name: "Margin All Pairs", Function: TestGetMarginAllPairs, AuthRequired: AuthTypeNONE, Category: "Margin"},
		{Name: "Margin Price Index", Function: TestGetMarginPriceIndex, AuthRequired: AuthTypeNONE, Category: "Margin"},
		{Name: "Create Margin Order", Function: TestCreateMarginOrder, AuthRequired: AuthTypeTRADE, Category: "Margin"},
		{Name: "Margin All Orders", Function: TestGetMarginAllOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin My Trades", Function: TestGetMarginMyTrades, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Max Borrowable", Function: TestGetMarginMaxBorrowable, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Interest History", Function: TestGetMarginInterestHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Listen Key", Function: TestCreateMarginListenKey, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Transfer Operations", Function: TestMarginTransferOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Loan Operations", Function: TestMarginLoanOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Account Operations", Function: TestMarginAccountOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Isolated Margin Operations", Function: TestIsolatedMarginOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin OCO Orders", Function: TestMarginOCOOrders, AuthRequired: AuthTypeTRADE, Category: "Margin"},
		{Name: "Margin Order Operations", Function: TestMarginOrderOperations, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin BNB Burn", Function: TestMarginBNBBurn, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Trade Fee", Function: TestMarginTradeFee, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		{Name: "Margin Collateral", Function: TestMarginCollateral, AuthRequired: AuthTypeUSER_DATA, Category: "Margin"},
		
		// Sub-Account Tests
		{Name: "SubAccount Management", Function: TestSubAccountManagement, AuthRequired: AuthTypeUSER_DATA, Category: "SubAccount"},
		{Name: "SubAccount Assets", Function: TestSubAccountAssets, AuthRequired: AuthTypeUSER_DATA, Category: "SubAccount"},
		{Name: "SubAccount Transfer History", Function: TestSubAccountTransferHistory, AuthRequired: AuthTypeUSER_DATA, Category: "SubAccount"},
		{Name: "SubAccount Margin/Futures", Function: TestSubAccountMarginFutures, AuthRequired: AuthTypeUSER_DATA, Category: "SubAccount"},
		{Name: "SubAccount Create", Function: TestSubAccountCreate, AuthRequired: AuthTypeTRADE, Category: "SubAccount"},
		{Name: "SubAccount Enable Features", Function: TestSubAccountEnableFeatures, AuthRequired: AuthTypeTRADE, Category: "SubAccount"},
		
		// Simple Earn Tests
		{Name: "Simple Earn Flexible Products", Function: TestSimpleEarnFlexibleProducts, AuthRequired: AuthTypeUSER_DATA, Category: "SimpleEarn"},
		{Name: "Simple Earn Flexible History", Function: TestSimpleEarnFlexibleHistory, AuthRequired: AuthTypeUSER_DATA, Category: "SimpleEarn"},
		{Name: "Simple Earn Locked Products", Function: TestSimpleEarnLockedProducts, AuthRequired: AuthTypeUSER_DATA, Category: "SimpleEarn"},
		{Name: "Simple Earn Locked History", Function: TestSimpleEarnLockedHistory, AuthRequired: AuthTypeUSER_DATA, Category: "SimpleEarn"},
		{Name: "Simple Earn Account", Function: TestSimpleEarnAccount, AuthRequired: AuthTypeUSER_DATA, Category: "SimpleEarn"},
		{Name: "Simple Earn Subscription Operations", Function: TestSimpleEarnSubscriptionOperations, AuthRequired: AuthTypeTRADE, Category: "SimpleEarn"},
		
		// Staking Tests
		{Name: "ETH Staking Account", Function: TestETHStakingAccount, AuthRequired: AuthTypeUSER_DATA, Category: "Staking"},
		{Name: "ETH Staking History", Function: TestETHStakingHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Staking"},
		{Name: "SOL Staking Account", Function: TestSOLStakingAccount, AuthRequired: AuthTypeUSER_DATA, Category: "Staking"},
		{Name: "SOL Staking History", Function: TestSOLStakingHistory, AuthRequired: AuthTypeUSER_DATA, Category: "Staking"},
		{Name: "Staking Operations", Function: TestStakingOperations, AuthRequired: AuthTypeTRADE, Category: "Staking"},
		
		// Algo Trading Tests
		{Name: "Algo Spot Orders", Function: TestAlgoSpotOrders, AuthRequired: AuthTypeUSER_DATA, Category: "AlgoTrading"},
		{Name: "Algo Futures Orders", Function: TestAlgoFuturesOrders, AuthRequired: AuthTypeUSER_DATA, Category: "AlgoTrading"},
		{Name: "Algo Spot TWAP Order", Function: TestAlgoSpotTWAPOrder, AuthRequired: AuthTypeTRADE, Category: "AlgoTrading"},
		{Name: "Algo Futures TWAP Order", Function: TestAlgoFuturesTWAPOrder, AuthRequired: AuthTypeTRADE, Category: "AlgoTrading"},
		{Name: "Algo Futures VP Order", Function: TestAlgoFuturesVPOrder, AuthRequired: AuthTypeTRADE, Category: "AlgoTrading"},
		
		// Convert Tests
		{Name: "Convert Info", Function: TestConvertInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		{Name: "Convert Quote", Function: TestConvertQuote, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		{Name: "Convert Orders", Function: TestConvertOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		{Name: "Convert Limit Orders", Function: TestConvertLimitOrders, AuthRequired: AuthTypeUSER_DATA, Category: "Convert"},
		{Name: "Convert Operations", Function: TestConvertOperations, AuthRequired: AuthTypeTRADE, Category: "Convert"},
		
		// Crypto Loan Tests
		{Name: "Crypto Loan Info", Function: TestCryptoLoanInfo, AuthRequired: AuthTypeUSER_DATA, Category: "CryptoLoan"},
		{Name: "Crypto Loan History", Function: TestCryptoLoanHistory, AuthRequired: AuthTypeUSER_DATA, Category: "CryptoLoan"},
		{Name: "Crypto Loan Orders", Function: TestCryptoLoanOrders, AuthRequired: AuthTypeUSER_DATA, Category: "CryptoLoan"},
		{Name: "Crypto Loan Operations", Function: TestCryptoLoanOperations, AuthRequired: AuthTypeTRADE, Category: "CryptoLoan"},
		
		// VIP Loan Tests
		{Name: "VIP Loan Info", Function: TestVipLoanInfo, AuthRequired: AuthTypeUSER_DATA, Category: "VipLoan"},
		{Name: "VIP Loan Account", Function: TestVipLoanAccount, AuthRequired: AuthTypeUSER_DATA, Category: "VipLoan"},
		{Name: "VIP Loan Operations", Function: TestVipLoanOperations, AuthRequired: AuthTypeTRADE, Category: "VipLoan"},
		
		// Mining Tests
		{Name: "Mining Public Info", Function: TestMiningPublicInfo, AuthRequired: AuthTypeUSER_DATA, Category: "Mining"},
		{Name: "Mining User Data", Function: TestMiningUserData, AuthRequired: AuthTypeUSER_DATA, Category: "Mining"},
		{Name: "Mining Payments", Function: TestMiningPayments, AuthRequired: AuthTypeUSER_DATA, Category: "Mining"},
		{Name: "Mining Hash Transfer", Function: TestMiningHashTransfer, AuthRequired: AuthTypeUSER_DATA, Category: "Mining"},
		
		// Portfolio Margin Tests
		{Name: "Portfolio Margin Account", Function: TestPortfolioMarginAccount, AuthRequired: AuthTypeUSER_DATA, Category: "PortfolioMargin"},
		{Name: "Portfolio Margin Loan", Function: TestPortfolioMarginLoan, AuthRequired: AuthTypeUSER_DATA, Category: "PortfolioMargin"},
		{Name: "Portfolio Margin Operations", Function: TestPortfolioMarginOperations, AuthRequired: AuthTypeTRADE, Category: "PortfolioMargin"},
		
		// Binance Link Tests
		{Name: "Broker Info", Function: TestBrokerInfo, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		{Name: "Broker SubAccount", Function: TestBrokerSubAccount, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		{Name: "Broker Commission", Function: TestBrokerCommission, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		{Name: "Referral Operations", Function: TestReferralOperations, AuthRequired: AuthTypeUSER_DATA, Category: "BinanceLink"},
		{Name: "Broker Operations", Function: TestBrokerOperations, AuthRequired: AuthTypeTRADE, Category: "BinanceLink"},
		
		// Gift Card Tests
		{Name: "Gift Card Info", Function: TestGiftCardInfo, AuthRequired: AuthTypeUSER_DATA, Category: "GiftCard"},
		{Name: "Gift Card Operations", Function: TestGiftCardOperations, AuthRequired: AuthTypeTRADE, Category: "GiftCard"},
		
		// Dual Investment Tests
		{Name: "Dual Investment Info", Function: TestDualInvestmentInfo, AuthRequired: AuthTypeUSER_DATA, Category: "DualInvestment"},
		{Name: "Dual Investment Operations", Function: TestDualInvestmentOperations, AuthRequired: AuthTypeTRADE, Category: "DualInvestment"},
		
		// Small APIs Tests
		{Name: "NFT API", Function: TestNFTAPI, AuthRequired: AuthTypeUSER_DATA, Category: "NFT"},
		{Name: "Fiat API", Function: TestFiatAPI, AuthRequired: AuthTypeUSER_DATA, Category: "Fiat"},
		{Name: "C2C API", Function: TestC2CAPI, AuthRequired: AuthTypeUSER_DATA, Category: "C2C"},
		{Name: "Binance Pay History API", Function: TestBinancePayHistoryAPI, AuthRequired: AuthTypeUSER_DATA, Category: "Pay"},
		{Name: "Copy Trading API", Function: TestCopyTradingAPI, AuthRequired: AuthTypeUSER_DATA, Category: "CopyTrading"},
		{Name: "Futures Data API", Function: TestFuturesDataAPI, AuthRequired: AuthTypeNONE, Category: "FuturesData"},
		{Name: "Rebate API", Function: TestRebateAPI, AuthRequired: AuthTypeUSER_DATA, Category: "Rebate"},
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

// logResponseBody logs the raw response body for debugging purposes
func logResponseBody(t *testing.T, httpResp *http.Response, context string) {
	if httpResp != nil && httpResp.Body != nil {
		// Try to read the body
		body, readErr := io.ReadAll(httpResp.Body)
		if readErr == nil {
			t.Logf("%s - Raw response body: %s", context, string(body))
		} else {
			t.Logf("%s - Failed to read response body: %v", context, readErr)
		}
	}
}

// checkAPIErrorWithResponse checks API errors and logs response body for 400 errors
func checkAPIErrorWithResponse(t *testing.T, err error, httpResp *http.Response, context string) {
	if err == nil {
		return
	}

	// Always call the original checkAPIError function
	checkAPIError(t, err)

	// Additionally log response body for 400 Bad Request errors
	if httpResp != nil && httpResp.StatusCode == 400 {
		logResponseBody(t, httpResp, context)
	}
}