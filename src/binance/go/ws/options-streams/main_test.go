package streamstest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
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

	fmt.Printf("  # Run per-channel full integration suites:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite_Combined\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite_Market\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite_UserData\n\n")

	fmt.Printf("  # Run with timeout:\n")
	fmt.Printf("  go test -v -timeout 20m\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Market and combined streams are public (no auth)\n")
	fmt.Printf("  - User data streams require BINANCE_API_KEY/SECRET_KEY + listenKey\n")
	fmt.Printf("  - Tests use Binance mainnet servers (wss://nbstream.binance.com/eoptions/ws)\n")
	fmt.Printf("  - Rate limiting: 1 connection per test for stability\n")
	fmt.Printf("  - Tests wait for real market data events\n")
	fmt.Printf("  - Some tests may timeout due to low options trading activity\n")
	fmt.Printf("  - Options data includes Greeks, IV, strike prices, and expiration dates\n")
	fmt.Printf("  - REST validation can be toggled via ENABLE_REST_VALIDATION=1\n")

	fmt.Println(strings.Repeat("=", 80))
}

// No monolithic suite here; each channel has its own TestFullIntegrationSuite_* entry.
