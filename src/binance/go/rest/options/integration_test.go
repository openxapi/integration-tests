package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/options"
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

// setupClient creates and configures a REST API client - similar to umfutures pattern
func setupClient(config TestConfig) (*openapi.APIClient, context.Context) {
	cfg := openapi.NewConfiguration()
	
	// Use production server by default (no dedicated Options testnet available)
	cfg.Servers = openapi.ServerConfigurations{
		{
			URL:         "https://eapi.binance.com",
			Description: "Binance Options API (Production)",
		},
	}

	// Override with custom server URL if provided
	if serverURL := os.Getenv("BINANCE_OPTIONS_REST_SERVER"); serverURL != "" {
		cfg.Servers = openapi.ServerConfigurations{
			{
				URL: serverURL,
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


// initializeTests initializes test functions - PRODUCTION SAFE: Only Market Data tests
func initializeTests() []TestInfo {
	var tests []TestInfo

	// Market Data Tests (PUBLIC - Safe for production server)
	tests = append(tests, TestInfo{
		Name:         "Market Data - Ping",
		Function:     testMarketDataPing,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Server Time",
		Function:     testMarketDataTime,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Exchange Info",
		Function:     testMarketDataExchangeInfo,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Order Book Depth",
		Function:     testMarketDataDepth,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Recent Trades",
		Function:     testMarketDataTrades,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - 24hr Ticker",
		Function:     testMarketDataTicker,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Klines",
		Function:     testMarketDataKlines,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Mark Price",
		Function:     testMarketDataMark,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Index Price",
		Function:     testMarketDataIndex,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	tests = append(tests, TestInfo{
		Name:         "Market Data - Open Interest",
		Function:     testMarketDataOpenInterest,
		AuthRequired: AuthTypeNONE,
		Category:     "Market Data",
	})

	// DISABLED FOR PRODUCTION SAFETY: Account and trading tests excluded
	// Uncomment below to enable account/trading tests (requires production credentials)
	
	/*
	// Account Tests (requires authentication) - DISABLED FOR SAFETY
	tests = append(tests, TestInfo{
		Name:         "Account - Options Account Info",
		Function:     testAccountInfo,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Position Info",
		Function:     testPositionInfo,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Margin Account Info",
		Function:     testMarginAccountInfo,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Bill/Funding Flow",
		Function:     testAccountBill,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - User Trades",
		Function:     testUserTrades,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Block User Trades",
		Function:     testBlockUserTrades,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Exercise Record",
		Function:     testExerciseRecord,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	tests = append(tests, TestInfo{
		Name:         "Account - Income Async",
		Function:     testIncomeAsync,
		AuthRequired: AuthTypeTRADE,
		Category:     "Account",
	})

	// User Data Stream Tests - DISABLED FOR SAFETY
	tests = append(tests, TestInfo{
		Name:         "User Data Stream - Create Listen Key",
		Function:     testCreateListenKey,
		AuthRequired: AuthTypeTRADE,
		Category:     "User Data Stream",
	})

	tests = append(tests, TestInfo{
		Name:         "User Data Stream - Lifecycle",
		Function:     testUserDataStreamLifecycle,
		AuthRequired: AuthTypeTRADE,
		Category:     "User Data Stream",
	})
	*/

	return tests
}

// TestFullIntegrationSuite runs all integration tests with emoji output - Public endpoints only
func TestFullIntegrationSuite(t *testing.T) {
	tests := initializeTests()
	
	fmt.Printf("\n=== Running Binance Options REST API Integration Test Suite ===\n")
	fmt.Printf("Total tests to run: %d (Market Data only - Production Safe)\n\n", len(tests))
	
	var totalTests, passedTests, failedTests, skippedTests int
	
	// Only run public endpoints configuration for market data tests
	config := TestConfig{
		Name:     "Public Endpoints",
		AuthType: AuthTypeNONE,
	}
	
	t.Run(config.Name, func(t *testing.T) {
		for _, test := range tests {
			totalTests++
			
			// Check if we can run this test
			if test.AuthRequired > config.AuthType {
				skippedTests++
				fmt.Printf("⚠️  SKIP %s - Authentication required but not provided\n", test.Name)
				continue
			}

			testName := test.Name
			testFunction := test.Function
			
			success := t.Run(test.Name, func(subT *testing.T) {
				testStart := time.Now()
				fmt.Printf("▶️  Running %s...", testName)
				
				defer func() {
					duration := time.Since(testStart)
					if subT.Failed() {
						failedTests++
						fmt.Printf(" ❌ FAILED (%.2fs)\n", duration.Seconds())
					} else {
						passedTests++
						fmt.Printf(" ✅ PASSED (%.2fs)\n", duration.Seconds())
					}
				}()
				
				defer func() {
					if r := recover(); r != nil {
						subT.Errorf("Test panicked: %v", r)
					}
				}()

				// Rate limiting
				rateLimiter.WaitForRateLimit()

				testFunction(subT)
			})
			
			if !success {
				failedTests++
			}
		}
	})
	
	// Print final summary
	fmt.Printf("\n=== Test Suite Summary ===\n")
	fmt.Printf("📊 Total Tests: %d\n", totalTests)
	fmt.Printf("✅ Passed: %d\n", passedTests)
	fmt.Printf("❌ Failed: %d\n", failedTests)
	fmt.Printf("⚠️  Skipped: %d\n", skippedTests)
	fmt.Printf("📞 Total API Requests: %d\n", rateLimiter.GetRequestCount())
	
	if failedTests > 0 {
		fmt.Printf("\n❌ Some tests failed - see details above\n")
	} else if passedTests > 0 {
		fmt.Printf("\n🎉 All tests passed successfully!\n")
	}
	fmt.Printf("\n")
}

// Helper function to get client and context from test environment
func getTestClientAndContext(t *testing.T) (*openapi.APIClient, context.Context) {
	// Get the first available authentication config, preferring HMAC/Ed25519
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType == AuthTypeTRADE {
			client, ctx := setupClient(config)
			return client, ctx
		}
	}
	
	// Fallback to public endpoints if no auth available
	if len(configs) > 0 {
		client, ctx := setupClient(configs[0])
		return client, ctx
	}
	
	t.Fatal("No test configuration available")
	return nil, nil
}

// prettyPrintJSON formats JSON for better readability
func prettyPrintJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(b)
}

// parseIntOrDefault parses a string to int with default value
func parseIntOrDefault(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

