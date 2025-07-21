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
	fmt.Println("🚀 === Binance Options REST API Integration Tests ===")
	fmt.Println("📡 Running integration tests for Binance Options REST API SDK")
	fmt.Println("⚠️  WARNING: Using PRODUCTION server (https://eapi.binance.com) - no testnet available")
	fmt.Println("🔴 IMPORTANT: These tests will interact with real Options API endpoints")
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
	fmt.Println("Market Data tests completed successfully!")
	fmt.Println("\n🔒 PRODUCTION-SAFE MODE: Only Market Data tests enabled")
	fmt.Println("✅ Available test categories (ENABLED):")
	fmt.Println("1. Market Data Tests - Public endpoints, no authentication required")
	fmt.Println("\n⚠️ DISABLED test categories (for production safety):")
	fmt.Println("2. Account & Position Tests - Requires production credentials")
	fmt.Println("3. User Data Stream Tests - Requires production credentials")  
	fmt.Println("4. Options Trading Tests - Risk of real money operations")
	fmt.Println("5. Block Trading Tests - Risk of real money operations")
	fmt.Println("6. MMP & Kill Switch Tests - Risk of affecting live trading")
	fmt.Println("\nTo run the market data tests:")
	fmt.Println("  go test -v -run TestFullIntegrationSuite ./...")
	fmt.Println("\n💡 To enable account/trading tests:")
	fmt.Println("  Uncomment tests in integration_test.go initializeTests() function")
	fmt.Println("  Set up production API credentials with Options permissions")
	fmt.Println("\n⚠️  CAUTION: Account/trading tests use PRODUCTION server")
	fmt.Println("  https://eapi.binance.com - No testnet available for Options API")
}

// RateLimitManager handles rate limiting across all tests
type RateLimitManager struct {
	mu             sync.Mutex
	lastCallTime   time.Time
	minInterval    time.Duration
	requestCounter int
}

// Global rate limit manager - more conservative for options API
var rateLimiter = &RateLimitManager{
	minInterval: 3 * time.Second, // Minimum 3 seconds between API calls for options
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