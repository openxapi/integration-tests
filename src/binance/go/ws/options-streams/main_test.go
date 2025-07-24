package streamstest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain controls test execution and can run the full integration suite if needed
func TestMain(m *testing.M) {
	flag.Parse()

	// Run the tests
	code := m.Run()

	// Clean up all shared clients
	disconnectAllSharedClients()

	// Print summary if running all tests
	if testing.Verbose() {
		printTestSummary()
	}

	os.Exit(code)
}

func printTestSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 BINANCE OPTIONS STREAMS INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	configs := getTestConfigs()

	fmt.Printf("📋 Available Test Configurations:\n")
	for _, config := range configs {
		fmt.Printf("  - %s: %s\n", config.Name, config.Description)
	}

	fmt.Printf("\n📋 Available Options Stream Types:\n")
	fmt.Printf("  - Index Price Streams: symbol@index\n")
	fmt.Printf("  - Kline Streams: symbol@kline_interval\n")
	fmt.Printf("  - Mark Price Streams: underlyingAsset@markPrice\n")
	fmt.Printf("  - New Symbol Info Stream: option_pair\n")
	fmt.Printf("  - Open Interest Streams: underlyingAsset@openInterest@expirationDate\n")
	fmt.Printf("  - Partial Depth Streams: symbol@depth{levels}[@{speed}]\n")
	fmt.Printf("  - Individual Ticker Streams: symbol@ticker\n")
	fmt.Printf("  - Ticker by Underlying Streams: underlyingAsset@ticker@expirationDate\n")
	fmt.Printf("  - Trade Streams: symbol@trade or underlyingAsset@trade\n")

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")

	fmt.Printf("  # Run the complete integration suite:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite\n\n")

	fmt.Printf("  # Run specific stream tests:\n")
	fmt.Printf("  go test -v -run TestIndexPriceStream\n")
	fmt.Printf("  go test -v -run TestKlineStream\n")
	fmt.Printf("  go test -v -run TestMarkPriceStream\n")
	fmt.Printf("  go test -v -run TestNewSymbolInfoStream\n")
	fmt.Printf("  go test -v -run TestOpenInterestStream\n")
	fmt.Printf("  go test -v -run TestPartialDepthStream\n")
	fmt.Printf("  go test -v -run TestTickerStream\n")
	fmt.Printf("  go test -v -run TestTickerByUnderlyingStream\n")
	fmt.Printf("  go test -v -run TestTradeStream\n\n")

	fmt.Printf("  # Run connection tests:\n")
	fmt.Printf("  go test -v -run TestConnection\n\n")

	fmt.Printf("  # Run with timeout:\n")
	fmt.Printf("  go test -v -timeout 20m\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - All options streams are public and don't require authentication\n")
	fmt.Printf("  - Tests use Binance mainnet servers (wss://nbstream.binance.com/eoptions/ws)\n")
	fmt.Printf("  - Rate limiting: 1 connection per test for stability\n")
	fmt.Printf("  - Tests wait for real market data events\n")
	fmt.Printf("  - Some tests may timeout due to low options trading activity\n")
	fmt.Printf("  - Options data includes Greeks, IV, strike prices, and expiration dates\n")

	fmt.Println(strings.Repeat("=", 80))
}

// Integration test that runs the full integration suite
func TestFullIntegrationSuite(t *testing.T) {
	t.Log("🚀 Running Full Binance Options Streams Integration Test Suite")
	t.Log("================================================================================")
	t.Log("🌐 Server: Binance Mainnet (wss://nbstream.binance.com/eoptions/ws)")
	t.Log("💡 Public streams - no authentication required")
	t.Log("📊 Options data includes Greeks, implied volatility, and risk metrics")
	t.Log("================================================================================")

	var totalTests, passedTests int
	var failedTests []string

	startTime := time.Now()

	// Test functions for different stream types
	testFunctions := []struct {
		name     string
		fn       func(*testing.T)
		required bool
	}{
		// Connection tests
		{"Connection", TestConnection, true},
		{"ServerManagement", TestServerManagement, true},

		// Basic stream tests - all options-specific streams
		{"IndexPriceStream", TestIndexPriceStream, true},
		{"KlineStream", TestKlineStream, true},
		{"MarkPriceStream", TestMarkPriceStream, true},
		{"NewSymbolInfoStream", TestNewSymbolInfoStream, true},
		{"OpenInterestStream", TestOpenInterestStream, true},
		{"PartialDepthStream", TestPartialDepthStream, true},
		{"TickerStream", TestTickerStream, true},
		{"TickerByUnderlyingStream", TestTickerByUnderlyingStream, true},
		{"TradeStream", TestTradeStream, true},

		// Advanced feature tests
		{"MultipleStreamTypes", TestMultipleStreamTypes, true},
		{"CombinedStreamEventHandler", TestCombinedStreamEventHandler, true},
		{"StreamErrorHandler", TestStreamErrorHandler, true},

		// Performance tests
		{"ConcurrentStreams", TestConcurrentStreams, false},
		{"HighVolumeStreams", TestHighVolumeStreams, false},
	}

	for _, testFunc := range testFunctions {
		totalTests++

		t.Logf("\n🧪 Running %s...", testFunc.name)
		start := time.Now()

		// Run test in a sub-test to capture failures
		success := t.Run(testFunc.name, testFunc.fn)
		duration := time.Since(start)

		if success {
			t.Logf("   ✅ %s passed (%v)", testFunc.name, duration)
			passedTests++
		} else {
			t.Logf("   ❌ %s failed (%v)", testFunc.name, duration)
			failedTests = append(failedTests, testFunc.name)
		}
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