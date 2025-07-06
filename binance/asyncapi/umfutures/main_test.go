//go:build ignore

package wstest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	umfuturesws "github.com/openxapi/binance-go/ws/umfutures"
)

// TestMain controls test execution and can run the full integration suite if needed
func TestMain(m *testing.M) {
	flag.Parse()

	// Run the tests
	code := m.Run()

	// Print summary if running all tests
	if testing.Verbose() {
		printTestSummary()
	}

	os.Exit(code)
}

func printTestSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 UMFUTURES INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	configs := getTestConfigs()

	fmt.Printf("📋 Available Test Configurations:\n")
	for _, config := range configs {
		fmt.Printf("  - %s: %s (%s auth)\n", config.Name, config.Description, config.AuthType)
	}

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")

	fmt.Printf("  # Run only public endpoint tests:\n")
	fmt.Printf("  go test -v -run TestPing\n")
	fmt.Printf("  go test -v -run 'Test.*Public.*'\n\n")

	fmt.Printf("  # Run tests for specific auth type:\n")
	fmt.Printf("  go test -v -run 'Test.*HMAC.*'\n")
	fmt.Printf("  go test -v -run 'Test.*Ed25519.*'\n\n")

	fmt.Printf("  # Run specific endpoint tests:\n")
	fmt.Printf("  go test -v -run TestSessionLogon\n")
	fmt.Printf("  go test -v -run TestNewOrder\n")
	fmt.Printf("  go test -v -run TestUserDataStream\n\n")

	fmt.Printf("  # Run trading tests only:\n")
	fmt.Printf("  go test -v trading_test.go integration_test.go\n\n")

	fmt.Printf("  # Run with timeout:\n")
	fmt.Printf("  go test -v -timeout 10m\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Set environment variables for authentication:\n")
	fmt.Printf("    BINANCE_API_KEY & BINANCE_SECRET_KEY (HMAC)\n")
	fmt.Printf("    BINANCE_RSA_API_KEY & BINANCE_RSA_PRIVATE_KEY_PATH (RSA)\n")
	fmt.Printf("    BINANCE_ED25519_API_KEY & BINANCE_ED25519_PRIVATE_KEY_PATH (Ed25519)\n")
	fmt.Printf("  - Tests use Binance Futures testnet for safety\n")
	fmt.Printf("  - Some tests require specific auth types (Ed25519 for SessionLogon)\n")
	fmt.Printf("  - Rate limiting: 200ms between requests\n")

	fmt.Println(strings.Repeat("=", 80))
}

// Integration test that runs the full original test suite for comparison
func TestFullIntegrationSuite(t *testing.T) {
	t.Log("🚀 Running Full UMFUTURES Integration Test Suite")
	t.Log("================================================================================")
	t.Log("🌐 Server: Binance Futures Testnet (wss://ws-fapi.testnet.binance.vision/ws-fapi/v3)")
	t.Log("💡 Safe for testing - no real money at risk")
	t.Log("================================================================================")

	configs := getTestConfigs()

	if len(configs) <= 1 {
		t.Log("⚠️  Warning: Limited authentication credentials available.")
		t.Log("   Set environment variables for comprehensive testing:")
		t.Log("   - BINANCE_API_KEY & BINANCE_SECRET_KEY (for HMAC)")
		t.Log("   - BINANCE_RSA_API_KEY & BINANCE_RSA_PRIVATE_KEY_PATH (for RSA)")
		t.Log("   - BINANCE_ED25519_API_KEY & BINANCE_ED25519_PRIVATE_KEY_PATH (for Ed25519)")
	}

	var totalTests, passedTests int
	var failedTests []string

	startTime := time.Now()

	for _, config := range configs {
		t.Logf("\n🔧 Testing Configuration: %s", config.Name)
		t.Logf("   Key Type: %s, Auth Type: %s", config.KeyType, config.AuthType)
		t.Logf("   Description: %s", config.Description)

		configStartTime := time.Now()
		configPassed := 0
		configTotal := 0

		// Run all test functions for this config
		testFunctions := []struct {
			name            string
			fn              func(*umfuturesws.Client, TestConfig) error
			authRequired    AuthType
			keyTypeRequired KeyType
		}{
			// Public tests
			{"TickerPrice", testTickerPrice, AuthTypeNONE, KeyTypeHMAC},
			{"BookTicker", testBookTicker, AuthTypeNONE, KeyTypeHMAC},
			{"Depth", testDepth, AuthTypeNONE, KeyTypeHMAC},

			// User data tests for HMAC
			{"AccountBalance", testAccountBalance, AuthTypeUSER_DATA, KeyTypeHMAC},
			{"AccountPosition", testAccountPosition, AuthTypeUSER_DATA, KeyTypeHMAC},
			{"AccountStatus", testAccountStatus, AuthTypeUSER_DATA, KeyTypeHMAC},
			{"V2AccountBalance", testV2AccountBalance, AuthTypeUSER_DATA, KeyTypeHMAC},
			{"V2AccountPosition", testV2AccountPosition, AuthTypeUSER_DATA, KeyTypeHMAC},
			{"V2AccountStatus", testV2AccountStatus, AuthTypeUSER_DATA, KeyTypeHMAC},

			// Trading tests for HMAC
			{"UserDataStreamStart", testUserDataStreamStart, AuthTypeTRADE, KeyTypeHMAC},
			{"UserDataStreamPing", testUserDataStreamPing, AuthTypeTRADE, KeyTypeHMAC},
			{"UserDataStreamStop", testUserDataStreamStop, AuthTypeTRADE, KeyTypeHMAC},
			{"OrderPlace", testOrderPlace, AuthTypeTRADE, KeyTypeHMAC},
			{"OrderStatus", testOrderStatus, AuthTypeTRADE, KeyTypeHMAC},
			{"OrderCancel", testOrderCancel, AuthTypeTRADE, KeyTypeHMAC},
			{"OrderModify", testOrderModify, AuthTypeTRADE, KeyTypeHMAC},

			// User data tests for Ed25519
			{"AccountBalance", testAccountBalance, AuthTypeUSER_DATA, KeyTypeED25519},
			{"AccountPosition", testAccountPosition, AuthTypeUSER_DATA, KeyTypeED25519},
			{"AccountStatus", testAccountStatus, AuthTypeUSER_DATA, KeyTypeED25519},

			// Trading tests for Ed25519
			{"UserDataStreamStart", testUserDataStreamStart, AuthTypeTRADE, KeyTypeED25519},
			{"UserDataStreamPing", testUserDataStreamPing, AuthTypeTRADE, KeyTypeED25519},
			{"UserDataStreamStop", testUserDataStreamStop, AuthTypeTRADE, KeyTypeED25519},
			{"OrderPlace", testOrderPlace, AuthTypeTRADE, KeyTypeED25519},
			{"OrderStatus", testOrderStatus, AuthTypeTRADE, KeyTypeED25519},
			{"OrderCancel", testOrderCancel, AuthTypeTRADE, KeyTypeED25519},
			{"OrderModify", testOrderModify, AuthTypeTRADE, KeyTypeED25519},

			// User data tests for RSA
			{"AccountBalance", testAccountBalance, AuthTypeUSER_DATA, KeyTypeRSA},
			{"AccountPosition", testAccountPosition, AuthTypeUSER_DATA, KeyTypeRSA},
			{"AccountStatus", testAccountStatus, AuthTypeUSER_DATA, KeyTypeRSA},

			// Trading tests for RSA
			{"UserDataStreamStart", testUserDataStreamStart, AuthTypeTRADE, KeyTypeRSA},
			{"UserDataStreamPing", testUserDataStreamPing, AuthTypeTRADE, KeyTypeRSA},
			{"UserDataStreamStop", testUserDataStreamStop, AuthTypeTRADE, KeyTypeRSA},
			{"OrderPlace", testOrderPlace, AuthTypeTRADE, KeyTypeRSA},
			{"OrderStatus", testOrderStatus, AuthTypeTRADE, KeyTypeRSA},
			{"OrderCancel", testOrderCancel, AuthTypeTRADE, KeyTypeRSA},
			{"OrderModify", testOrderModify, AuthTypeTRADE, KeyTypeRSA},
		}

		client, err := setupClient(config)
		if err != nil {
			t.Fatalf("Failed to setup client for %s: %v", config.Name, err)
		}

		for _, testFunc := range testFunctions {
			// Check if test should run for this configuration
			shouldRun := true

			// Check auth requirements
			if testFunc.authRequired != config.AuthType {
				shouldRun = false
			}

			// Check key type requirements
			if testFunc.keyTypeRequired != config.KeyType {
				shouldRun = false
			}

			if !shouldRun {
				continue
			}

			configTotal++
			totalTests++

			testSuite.rateLimit.Wait()

			start := time.Now()
			err := testFunc.fn(client, config)
			duration := time.Since(start)

			if err != nil {
				t.Logf("   🧪 Running %s... ❌ Failed (%v)", testFunc.name, duration)
				t.Logf("      Error: %v", err)
				failedTests = append(failedTests, fmt.Sprintf("%s-%s", config.Name, testFunc.name))
			} else {
				t.Logf("   🧪 Running %s... ✅ Passed (%v)", testFunc.name, duration)
				configPassed++
				passedTests++
			}
		}

		client.Disconnect()

		configDuration := time.Since(configStartTime)
		t.Logf("   📊 Configuration %s: %d/%d passed (%.1f%%) in %v",
			config.Name, configPassed, configTotal,
			float64(configPassed)/float64(configTotal)*100, configDuration)
	}

	totalDuration := time.Since(startTime)

	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("📊 TEST SUMMARY")
	t.Log(strings.Repeat("=", 80))
	t.Logf("Total Tests: %d", totalTests)
	t.Logf("✅ Passed: %d", passedTests)
	t.Logf("❌ Failed: %d", totalTests-passedTests)
	t.Logf("⏱️  Total Duration: %v", totalDuration)
	t.Logf("📈 Success Rate: %.1f%%", float64(passedTests)/float64(totalTests)*100)

	if len(failedTests) > 0 {
		t.Log("\n❌ Failed Tests:")
		for _, failedTest := range failedTests {
			t.Logf("  - %s", failedTest)
		}
	}

	t.Log(strings.Repeat("=", 80))
}