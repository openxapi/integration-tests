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

	// Print summary if running all tests
	if testing.Verbose() {
		printTestSummary()
	}

	os.Exit(code)
}

func printTestSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 SPOT STREAMS INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📋 Available Stream Types:\n")
	fmt.Printf("  - Trade Streams: symbol@trade\n")
	fmt.Printf("  - Aggregate Trade Streams: symbol@aggTrade\n")
	fmt.Printf("  - Kline Streams: symbol@kline_interval\n")
	fmt.Printf("  - Mini Ticker Streams: symbol@miniTicker\n")
	fmt.Printf("  - Ticker Streams: symbol@ticker\n")
	fmt.Printf("  - Book Ticker Streams: symbol@bookTicker\n")
	fmt.Printf("  - Depth Streams: symbol@depth\n")
	fmt.Printf("  - Partial Depth Streams: symbol@depth5, symbol@depth10, symbol@depth20\n")
	fmt.Printf("  - Rolling Window Ticker: symbol@ticker_1h, symbol@ticker_4h\n")
	fmt.Printf("  - Average Price: symbol@avgPrice\n")

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")

	fmt.Printf("  # Run the complete integration suite:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite\n\n")

	fmt.Printf("  # Run specific stream tests:\n")
	fmt.Printf("  go test -v -run TestTradeStream\n")
	fmt.Printf("  go test -v -run TestKlineStream\n")
	fmt.Printf("  go test -v -run TestDepthStream\n")
	fmt.Printf("  go test -v -run TestDepthStreamUpdateSpeed\n")
	fmt.Printf("  go test -v -run TestPartialDepthStreamUpdateSpeed\n")
	fmt.Printf("  go test -v -run TestMultipleStreamTypes\n\n")

	fmt.Printf("  # Run connection tests:\n")
	fmt.Printf("  go test -v -run TestConnection\n\n")

	fmt.Printf("  # Run subscription management tests:\n")
	fmt.Printf("  go test -v -run TestSubscription\n\n")

	fmt.Printf("  # Run with timeout:\n")
	fmt.Printf("  go test -v -timeout 10m\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Most spot streams are public and don't require authentication\n")
	fmt.Printf("  - Tests use Binance testnet servers (wss://stream.testnet.binance.vision/ws)\n")
	fmt.Printf("  - Rate limiting: 1 connection per test for stability\n")
	fmt.Printf("  - Tests wait for real market data events\n")
	fmt.Printf("  - Some tests may be skipped in short mode\n")

	fmt.Println(strings.Repeat("=", 80))
}

// Integration test that runs the full integration suite
func TestFullIntegrationSuite(t *testing.T) {
	t.Log("🚀 Running Full Spot Streams Integration Test Suite")
	t.Log("================================================================================")
	t.Log("🌐 Server: Binance Testnet (wss://stream.testnet.binance.vision/ws)")
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
		{"ConnectionTimeout", TestConnectionTimeout, true},
		{"MultipleConnections", TestMultipleConnections, true},
		{"ConnectToSpecificServer", TestConnectToSpecificServer, true},
		{"ConnectionRecovery", TestConnectionRecovery, false},
		{"ConnectToSingleStreams", TestConnectToSingleStreams, true},
		{"ConnectToCombinedStreams", TestConnectToCombinedStreams, true},
		{"ConnectToSingleStreamsMicrosecond", TestConnectToSingleStreamsMicrosecond, false},
		{"ConnectToCombinedStreamsMicrosecond", TestConnectToCombinedStreamsMicrosecond, false},

		// Basic stream tests
		{"TradeStream", TestTradeStream, true},
		{"AggregateTradeStream", TestAggregateTradeStream, true},
		{"KlineStream", TestKlineStream, true},
		{"TickerStream", TestTickerStream, true},
		{"MiniTickerStream", TestMiniTickerStream, true},
		{"BookTickerStream", TestBookTickerStream, true},
		{"MultipleSymbolStreams", TestMultipleSymbolStreams, true},
		{"DifferentKlineIntervals", TestDifferentKlineIntervals, false},
		{"AllTickerStream", TestAllTickerStream, false},
		{"AllMiniTickerStream", TestAllMiniTickerStream, false},
		{"AllBookTickerStream", TestAllBookTickerStream, false},

		// Depth stream tests
		{"DepthStream", TestDepthStream, true},
		{"PartialDepthStream", TestPartialDepthStream, true},
		{"DifferentDepthLevels", TestDifferentDepthLevels, true},
		{"DepthStreamUpdateSpeed", TestDepthStreamUpdateSpeed, true},
		{"PartialDepthStreamUpdateSpeed", TestPartialDepthStreamUpdateSpeed, true},
		{"DepthStreamSpeedComparison", TestDepthStreamSpeedComparison, false},

		// Advanced stream tests
		{"RollingWindowTickerStream", TestRollingWindowTickerStream, false},
		{"AvgPriceStream", TestAvgPriceStream, false},
		{"MultipleStreamTypes", TestMultipleStreamTypes, true},

		// Subscription management tests
		{"SubscriptionManagement", TestSubscriptionManagement, true},
		{"MultipleStreamsSubscription", TestMultipleStreamsSubscription, true},
		{"StreamUnsubscription", TestStreamUnsubscription, true},
		{"ListSubscriptions", TestListSubscriptions, false},
		{"SubscriptionToInvalidStream", TestSubscriptionToInvalidStream, true},
		{"Resubscription", TestResubscription, false},
		{"BatchSubscription", TestBatchSubscription, false},

		// Error handling tests
		{"ErrorHandling", TestErrorHandling, true},
		{"InvalidStreamNames", TestInvalidStreamNames, true},
		{"ConnectionErrors", TestConnectionErrors, true},
		{"UnsubscribeNonExistentStream", TestUnsubscribeNonExistentStream, true},
		{"EmptyStreamList", TestEmptyStreamList, true},
		{"MaxStreamLimits", TestMaxStreamLimits, false},
		{"ReconnectionAfterError", TestReconnectionAfterError, false},
		{"ConcurrentSubscriptionsError", TestConcurrentSubscriptions, false},

		// Combined streams tests
		{"CombinedStreamEventReception", TestCombinedStreamEventReception, true},
		{"CombinedStreamEventDataTypes", TestCombinedStreamEventDataTypes, true},
		{"CombinedStreamMicrosecondPrecision", TestCombinedStreamMicrosecondPrecision, false},
		{"SingleVsCombinedStreamComparison", TestSingleVsCombinedStreamComparison, false},
		{"CombinedStreamSubscriptionManagement", TestCombinedStreamSubscriptionManagement, true},

		// Performance tests
		{"ConcurrentStreams", TestConcurrentStreams, false},
		{"HighVolumeStreams", TestHighVolumeStreams, false},
		{"StreamLatency", TestStreamLatency, false},
		{"MemoryUsage", TestMemoryUsage, false},
		{"RapidSubscriptionChanges", TestRapidSubscriptionChanges, false},
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