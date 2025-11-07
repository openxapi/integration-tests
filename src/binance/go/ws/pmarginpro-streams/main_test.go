package streamstest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestMain coordinates setup/teardown for the pmargin pro stream integration tests
func TestMain(m *testing.M) {
	flag.Parse()

	exitCode := m.Run()

	// Ensure cached clients are released
	disconnectAllSharedClients()

	if testing.Verbose() {
		printTestSummary()
	}

	os.Exit(exitCode)
}

func printTestSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 BINANCE PORTFOLIO MARGIN PRO STREAMS INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	configs := getTestConfigs()
	fmt.Println("📋 Available Test Configurations:")
	for _, cfg := range configs {
		fmt.Printf("  - %s: %s\n", cfg.Name, cfg.Description)
	}

	fmt.Printf("\n📡 Available Stream Channels:\n")
	fmt.Printf("  - User Data Stream: listen-key scoped account risk notifications\n")

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")
	fmt.Printf("  # Run the user data integration suite only:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite_UserData\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Portfolio Margin Pro WebSocket data requires BINANCE_API_KEY/SECRET_KEY\n")
	fmt.Printf("  - Tests obtain a listen key via REST; override with BINANCE_PM_REST_SERVER or BINANCE_REST_SERVER_URL if needed\n")
	fmt.Printf("  - Events depend on live PM Pro activity; timeouts are reported as flaky but acceptable\n")
	fmt.Printf("  - SDK defaults dial wss://fstream.binance.com/pm-classic/{listenKey}\n")
	fmt.Println(strings.Repeat("=", 80))
}
