package main

import (
	"net/http"
	"strings"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/spot"
)

// handleTestnetError checks if error is due to testnet limitations and skips test if so
// Returns true if test should be skipped, false if test should continue
func handleTestnetError(t *testing.T, err error, httpResp *http.Response, testName string) bool {
	if err == nil {
		return false
	}
	
	// Check for 404/403 status codes - endpoint not available on testnet
	if httpResp != nil && (httpResp.StatusCode == 404 || httpResp.StatusCode == 403) {
		t.Skipf("%s endpoint not available on testnet", testName)
		return true
	}
	
	// Check for HTML error pages from testnet
	if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		body := string(apiErr.Body())
		if len(body) > 50 && (strings.HasPrefix(body, "<!DOCTYPE html>") || strings.HasPrefix(body, "<html")) {
			t.Skipf("%s endpoint returns HTML error page - not available on testnet", testName)
			return true
		}
		// Log the actual API error for debugging
		t.Logf("API error response: %s", body)
	}
	
	// Check for undefined response type errors (SDK can't parse HTML responses)
	// NEVER skip 400 Bad Request errors - these indicate real API issues that need fixing
	if strings.Contains(err.Error(), "undefined response type") {
		if httpResp != nil && httpResp.StatusCode == 400 {
			// NEVER skip 400 errors - they indicate real API issues that need investigation
			t.Logf("%s: 400 Bad Request with undefined response type - this is a real API error that needs fixing", testName)
			return false
		}
		// Only skip non-400 undefined response type errors (these are usually testnet limitations)
		t.Skipf("%s endpoint has response parsing issues - likely testnet limitation", testName)
		return true
	}
	
	// Check for other common testnet limitation patterns
	errMsg := err.Error()
	testnetPatterns := []string{
		"This service is not available",
		"Feature not supported",
		"not available in testnet",
		"testnet not supported",
	}
	
	for _, pattern := range testnetPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			t.Skipf("%s endpoint not supported on testnet: %s", testName, pattern)
			return true
		}
	}
	
	return false
}

// logAPIError safely logs API error details for debugging
func logAPIError(t *testing.T, err error) {
	if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		t.Logf("API error response: %s", string(apiErr.Body()))
	}
}

