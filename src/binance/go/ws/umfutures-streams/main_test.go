package streamstest

import (
    "flag"
    "fmt"
    "os"
    "strings"
    "testing"
)

// TestMain controls test execution and prints a concise summary
func TestMain(m *testing.M) {
    flag.Parse()
    code := m.Run()
    disconnectAllSharedClients()
    if testing.Verbose() { printTestSummary() }
    os.Exit(code)
}

func printTestSummary() {
    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("📊 UM FUTURES STREAMS — INTEGRATION TEST SUMMARY")
    fmt.Println(strings.Repeat("=", 80))

    cfgs := getTestConfigs()
    fmt.Printf("📋 Configurations:\n")
    for _, c := range cfgs { fmt.Printf("  - %s: %s\n", c.Name, c.Description) }

    fmt.Printf("\n📋 Channels with suites:\n")
    fmt.Printf("  - MarketStreamsChannel  → TestFullIntegrationSuite_Market\n")
    fmt.Printf("  - CombinedMarketStreams → TestFullIntegrationSuite_Combined\n")

    fmt.Printf("\n💡 Run examples:\n")
    fmt.Printf("  go test -v\n")
    fmt.Printf("  go test -v -run TestFullIntegrationSuite_Market\n")
    fmt.Printf("  go test -v -run TestFullIntegrationSuite_Combined\n")

    fmt.Printf("\n⚠️  Notes:\n")
    fmt.Printf("  - Tests select WS server 'testnet1' by default\n")
    fmt.Printf("  - Public streams only; user-data suite pending\n")
    fmt.Printf("  - REST validation can be enabled via ENABLE_REST_VALIDATION=1\n")

    fmt.Println(strings.Repeat("=", 80))
}
