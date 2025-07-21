package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	fmt.Println("=== Binance Portfolio Margin REST API Integration Tests ===")
	fmt.Println("Running integration tests for Binance Portfolio Margin REST API SDK")
	fmt.Println("Using testnet server by default")
	fmt.Println()

	// Run tests
	code := m.Run()

	// Print summary based on test results
	if code == 0 {
		printTestSummary()
	}

	os.Exit(code)
}

// printTestSummary provides a comprehensive summary of available tests
func printTestSummary() {
	fmt.Println("\n=== Test Summary ===")
	fmt.Println("All tests completed successfully!")
	fmt.Println("\nAvailable test categories:")
	fmt.Println("1. General & System Tests - Basic connectivity and ping")
	fmt.Println("2. Account Tests - Account information and balance queries")
	fmt.Println("3. Asset Collection Tests - Asset collection and transfer operations")
	fmt.Println("4. Repay Tests - Negative balance and auto-repay operations")
	fmt.Println("5. Margin Trading Tests - Margin loan and trading operations")
	fmt.Println("6. UM Futures Tests - USD-M futures trading operations")
	fmt.Println("7. CM Futures Tests - Coin-M futures trading operations")
	fmt.Println("8. Portfolio Margin Tests - Portfolio margin specific operations")
	fmt.Println("9. User Data Stream Tests - Listen key management")
	fmt.Println("10. Rate Limit Tests - Rate limit information")
	fmt.Println("\nTo run specific test categories:")
	fmt.Println("  go test -v -run TestGeneral ./...")
	fmt.Println("  go test -v -run TestAccount ./...")
	fmt.Println("  go test -v -run TestAssetCollection ./...")
	fmt.Println("  go test -v -run TestRepay ./...")
	fmt.Println("  go test -v -run TestMarginTrading ./...")
	fmt.Println("  go test -v -run TestUMFutures ./...")
	fmt.Println("  go test -v -run TestCMFutures ./...")
	fmt.Println("  go test -v -run TestPortfolioMargin ./...")
	fmt.Println("  go test -v -run TestUserDataStream ./...")
	fmt.Println("  go test -v -run TestRateLimit ./...")
	fmt.Println("\nTo run the full integration suite:")
	fmt.Println("  go test -v -run TestFullIntegrationSuite ./...")
	fmt.Println("\nAuthentication methods supported:")
	fmt.Println("  - HMAC (default)")
	fmt.Println("  - RSA")
	fmt.Println("  - Ed25519")
	fmt.Println("\nIMPORTANT: Portfolio Margin API requires special account type and permissions")
	fmt.Println("Most endpoints may not be available on testnet - use production with caution")
}

// RateLimitManager handles rate limiting across all tests
type RateLimitManager struct {
	mu             sync.Mutex
	lastCallTime   time.Time
	minInterval    time.Duration
	requestCounter int
}

// Global rate limit manager
var rateLimiter = &RateLimitManager{
	minInterval: 2 * time.Second, // Minimum 2 seconds between API calls for testnet
}

// WaitForRateLimit ensures we don't exceed rate limits
func (r *RateLimitManager) WaitForRateLimit() {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastCallTime)
	if elapsed < r.minInterval {
		waitTime := r.minInterval - elapsed
		time.Sleep(waitTime)
	}

	r.lastCallTime = time.Now()
	r.requestCounter++
}

// GetRequestCount returns the total number of requests made
func (r *RateLimitManager) GetRequestCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.requestCounter
}

// ResetCounter resets the request counter
func (r *RateLimitManager) ResetCounter() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requestCounter = 0
}