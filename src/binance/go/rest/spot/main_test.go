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
	fmt.Println("=== Binance Spot REST API Integration Tests ===")
	fmt.Println("Running integration tests for Binance Spot REST API SDK")
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
	fmt.Println("1. Public API Tests - No authentication required")
	fmt.Println("2. Trading API Tests - Requires API key and secret")
	fmt.Println("3. Account API Tests - Account information and balances")
	fmt.Println("4. Wallet API Tests - Deposits, withdrawals, and asset info")
	fmt.Println("\nTo run specific test categories:")
	fmt.Println("  go test -v -run TestPublic ./...")
	fmt.Println("  go test -v -run TestTrading ./...")
	fmt.Println("  go test -v -run TestAccount ./...")
	fmt.Println("  go test -v -run TestWallet ./...")
	fmt.Println("\nTo run the full integration suite:")
	fmt.Println("  go test -v -run TestFullIntegrationSuite ./...")
	fmt.Println("\nAuthentication methods supported:")
	fmt.Println("  - HMAC (default)")
	fmt.Println("  - RSA")
	fmt.Println("  - Ed25519")
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