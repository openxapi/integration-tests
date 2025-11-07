package wstest

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if testing.Verbose() {
		printHarnessSummary()
	}
	os.Exit(code)
}

func printHarnessSummary() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 78))
	fmt.Println("Binance Spot WS Integration Harness")
	fmt.Println(strings.Repeat("=", 78))

	creds := getCreds()
	fmt.Println("Credential bundles:")
	printConfigStatus("Public-NoAuth", creds.Public != nil)
	printConfigStatus("HMAC", creds.HMAC != nil)
	printConfigStatus("RSA", creds.RSA != nil)
	printConfigStatus("Ed25519", creds.Ed25519 != nil)

	fmt.Println("\nSuites:")
	fmt.Println("  - SpotChannel          → TestFullIntegrationSuite_Spot")

	fmt.Println("\nCommon flags:")
	fmt.Println("  WS_THROTTLE_MS          request pacing (default 300)")
	fmt.Println("  EVENT_WAIT_SECS         event wait duration (default 20)")
	fmt.Println("  BINANCE_SPOT_WS_SERVER  optional WS override (wss://...)")
	fmt.Println("  BINANCE_SPOT_REST_SERVER optional REST override (https://...)")

	fmt.Println(strings.Repeat("=", 78))
}

func printConfigStatus(name string, available bool) {
	if available {
		fmt.Printf("  [x] %s\n", name)
	} else {
		fmt.Printf("  [ ] %s (not configured)\n", name)
	}
}
