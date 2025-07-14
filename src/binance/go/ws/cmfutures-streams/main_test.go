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
	fmt.Println("📊 COIN-M FUTURES STREAMS INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	configs := getTestConfigs()

	fmt.Printf("📋 Available Test Configurations:\n")
	for _, config := range configs {
		fmt.Printf("  - %s: %s\n", config.Name, config.Description)
	}

	fmt.Printf("\n📋 Available Stream Types:\n")
	fmt.Printf("  - Aggregate Trade Streams: symbol@aggTrade\n")
	fmt.Printf("  - Mark Price Streams: symbol@markPrice\n")
	fmt.Printf("  - Kline Streams: symbol@kline_interval\n")
	fmt.Printf("  - Continuous Kline Streams: pair_contractType@continuousKline_interval\n")
	fmt.Printf("  - Mini Ticker Streams: symbol@miniTicker\n")
	fmt.Printf("  - Ticker Streams: symbol@ticker\n")
	fmt.Printf("  - Book Ticker Streams: symbol@bookTicker\n")
	fmt.Printf("  - Liquidation Order Streams: symbol@forceOrder\n")
	fmt.Printf("  - Partial Depth Streams: symbol@depth5, symbol@depth10, symbol@depth20\n")
	fmt.Printf("  - Diff Depth Streams: symbol@depth\n")
	fmt.Printf("  - BLVT Info Streams: symbol@tokenInfo\n")
	fmt.Printf("  - BLVT Kline Streams: symbol@tokenKline_interval\n")

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")

	fmt.Printf("  # Run the complete integration suite:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite\n\n")

	fmt.Printf("  # Run specific stream tests:\n")
	fmt.Printf("  go test -v -run TestAggregateTradeStream\n")
	fmt.Printf("  go test -v -run TestMarkPriceStream\n")
	fmt.Printf("  go test -v -run TestKlineStream\n")
	fmt.Printf("  go test -v -run TestContinuousKlineStream\n")
	fmt.Printf("  go test -v -run TestLiquidationOrderStream\n")
	fmt.Printf("  go test -v -run TestPartialDepthStream\n")
	fmt.Printf("  go test -v -run TestDiffDepthStream\n")
	fmt.Printf("  go test -v -run TestMultipleStreamTypes\n\n")

	fmt.Printf("  # Run connection tests:\n")
	fmt.Printf("  go test -v -run TestConnection\n\n")

	fmt.Printf("  # Run subscription management tests:\n")
	fmt.Printf("  go test -v -run TestSubscription\n\n")

	fmt.Printf("  # Run comprehensive integration suites:\n")
	fmt.Printf("  go test -v -run TestMarketStreamsIntegration\n\n")

	fmt.Printf("  # Run with timeout:\n")
	fmt.Printf("  go test -v -timeout 20m\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Most Coin-M futures streams are public and don't require authentication\n")
	fmt.Printf("  - Tests use Binance testnet servers (wss://dstream.binancefuture.com/ws)\n")
	fmt.Printf("  - Rate limiting: 1 connection per test for stability\n")
	fmt.Printf("  - Tests wait for real market data events\n")
	fmt.Printf("  - Some tests may be skipped in short mode\n")

	fmt.Println(strings.Repeat("=", 80))
}

// Integration test that runs the full integration suite
func TestFullIntegrationSuite(t *testing.T) {
	t.Log("🚀 Running Full Coin-M Futures Streams Integration Test Suite")
	t.Log("================================================================================")
	t.Log("🌐 Server: Binance Testnet (wss://dstream.binancefuture.com/ws)")
	t.Log("💡 Public streams - no authentication required")
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
		{"AdvancedServerManagement", TestAdvancedServerManagement, true},

		// Enhanced connection methods
		{"EnhancedConnectionMethods", TestEnhancedConnectionMethods, true},

		// Basic stream tests
		{"AggregateTradeStream", TestAggregateTradeStream, true},
		{"MarkPriceStream", TestMarkPriceStream, true},
		{"KlineStream", TestKlineStream, true},
		{"ContinuousKlineStream", TestContinuousKlineStream, true},
		{"MiniTickerStream", TestMiniTickerStream, true},
		{"TickerStream", TestTickerStream, true},
		{"BookTickerStream", TestBookTickerStream, true},
		{"LiquidationOrderStream", TestLiquidationOrderStream, true},

		// Array streams (@arr) tests
		{"AllArrayStreams", TestAllArrayStreams, false},

		// Depth stream tests
		{"PartialDepthStream", TestPartialDepthStream, true},
		{"DiffDepthStream", TestDiffDepthStream, true},
		{"DifferentDepthLevels", TestDifferentDepthLevels, true},
		{"DiffDepthStreamUpdateSpeed", TestDiffDepthStreamUpdateSpeed, true},
		{"PartialDepthStreamUpdateSpeed", TestPartialDepthStreamUpdateSpeed, true},

		// Special stream tests (Coin-M specific streams only)
		{"MultipleStreamTypes", TestMultipleStreamTypes, true},

		// New enhanced event handlers
		{"ContractInfoEventHandler", TestContractInfoEventHandler, false},
		{"AssetIndexEventHandler", TestAssetIndexEventHandler, false},
		{"CombinedStreamEventHandler", TestCombinedStreamEventHandler, true},
		{"SubscriptionResponseHandler", TestSubscriptionResponseHandler, true},
		{"StreamErrorHandler", TestStreamErrorHandler, true},

		// Missing stream type tests (for 100% coverage)
		{"IndexPriceKlineStream", TestIndexPriceKlineStream, false},
		{"MarkPriceKlineStream", TestMarkPriceKlineStream, false},
		{"ContractInfoStream", TestContractInfoStream, false},
		{"IndividualIndexPriceStream", TestIndividualIndexPriceStream, false},


		// Subscription management tests
		{"SubscriptionManagement", TestSubscriptionManagement, true},
		{"MultipleStreamsSubscription", TestMultipleStreamsSubscription, true},
		{"StreamUnsubscription", TestStreamUnsubscription, true},

		// Error handling tests
		{"ErrorHandling", TestErrorHandling, true},
		{"InvalidStreamNames", TestInvalidStreamNames, true},
		{"ComprehensiveErrorHandling", TestComprehensiveErrorHandling, true},

		// Advanced feature tests
		{"AdvancedPropertyManagement", TestAdvancedPropertyManagement, true},
		{"RateLimitingBehavior", TestRateLimitingBehavior, false},

		// Combined streams tests
		{"CombinedStreamEventReception", TestCombinedStreamEventReception, true},
		{"CombinedStreamEventDataTypes", TestCombinedStreamEventDataTypes, true},
		{"CombinedStreamSubscriptionManagement", TestCombinedStreamSubscriptionManagement, true},

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