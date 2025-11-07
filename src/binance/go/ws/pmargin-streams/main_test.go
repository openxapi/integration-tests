package streamstest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestMain coordinates setup/teardown for the pmargin stream integration tests
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
	fmt.Println("📊 BINANCE PORTFOLIO MARGIN STREAMS INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	configs := getTestConfigs()
	fmt.Println("📋 Available Test Configurations:")
	for _, cfg := range configs {
		fmt.Printf("  - %s: %s\n", cfg.Name, cfg.Description)
	}

	fmt.Printf("\n📡 Available Stream Channels:\n")
	fmt.Printf("  - User Data Stream: listen-key scoped account and trading events\n")

	fmt.Printf("\n💡 Usage Examples:\n")
	fmt.Printf("  # Run all tests:\n")
	fmt.Printf("  go test -v\n\n")
	fmt.Printf("  # Run the user data integration suite only:\n")
	fmt.Printf("  go test -v -run TestFullIntegrationSuite_UserData\n\n")

	fmt.Printf("⚠️  Notes:\n")
	fmt.Printf("  - Portfolio margin WebSocket data requires BINANCE_API_KEY/SECRET_KEY\n")
	fmt.Printf("  - Tests obtain a fresh listen key via REST; set BINANCE_PM_REST_SERVER to override endpoint\n")
	fmt.Printf("  - Events depend on live portfolio margin activity; timeouts are reported as flaky but acceptable\n")
	fmt.Printf("  - SDK defaults dial wss://fstream.binance.com/pm/ws/{listenKey}\n")
	fmt.Println(strings.Repeat("=", 80))
}
