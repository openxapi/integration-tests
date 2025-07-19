package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// RateLimiter helps manage API request rate limits
type RateLimiter struct {
	mu           sync.Mutex
	requests     int
	lastRequest  time.Time
	minInterval  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		minInterval: time.Second / time.Duration(requestsPerSecond),
	}
}

// WaitForRateLimit waits if necessary to respect rate limits
func (rl *RateLimiter) WaitForRateLimit() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	if timeSinceLastRequest := now.Sub(rl.lastRequest); timeSinceLastRequest < rl.minInterval {
		time.Sleep(rl.minInterval - timeSinceLastRequest)
	}
	
	rl.requests++
	rl.lastRequest = time.Now()
}

// GetRequestCount returns the total number of requests made
func (rl *RateLimiter) GetRequestCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.requests
}

// Global rate limiter instance
var rateLimiter = NewRateLimiter(10) // 10 requests per second

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	fmt.Println("=== Binance CM Futures REST API Integration Tests ===")
	fmt.Println("Setting up test environment...")
	
	// Set up any global test configuration here
	// For example, check for required environment variables
	
	// Run the tests
	exitCode := m.Run()
	
	fmt.Println("Test suite completed.")
	fmt.Printf("Total API requests made: %d\n", rateLimiter.GetRequestCount())
	
	// Report any SDK issues found during testing
	sdkIssues := getSDKIssues()
	if len(sdkIssues) > 0 {
		fmt.Printf("\n🚨 FINAL SDK ISSUES REPORT (%d issues detected):\n", len(sdkIssues))
		fmt.Println("The following SDK issues should be reported to the maintainers:")
		for i, issue := range sdkIssues {
			fmt.Printf("%d. %s\n", i+1, issue)
		}
		fmt.Println("\nSDK Repository: ../binance-go/rest/cmfutures")
		fmt.Println("These endpoints have incorrect URLs causing 404 HTML responses instead of API responses.")
	}
	
	// Exit with the test result code
	fmt.Printf("Exit code: %d\n", exitCode)
}