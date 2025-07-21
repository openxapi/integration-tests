package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// handleTestnetError checks if error is due to testnet limitations and skips test if so
// Returns true if test should be skipped, false if test should continue
func handleTestnetError(t *testing.T, err error, httpResp *http.Response, testName string) bool {
	if err == nil {
		return false
	}
	
	// NEVER skip 400 Bad Request errors - these indicate real API issues that need fixing
	if httpResp != nil && httpResp.StatusCode == 400 {
		t.Logf("⚠️ 400 BAD REQUEST ERROR in %s", testName)
		logResponseBody(t, httpResp, testName)
		if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
			body := string(apiErr.Body())
			t.Logf("400 Error Response Body: %s", body)
		}
		return false // Let the test fail normally to show the 400 error
	}
	
	// Check for 404/403 status codes - endpoint not available on testnet
	if httpResp != nil && (httpResp.StatusCode == 404 || httpResp.StatusCode == 403) {
		logResponseBody(t, httpResp, testName)
		t.Skipf("%s endpoint not available on testnet", testName)
		return true
	}
	
	// For any other HTTP error, log the response body
	if httpResp != nil && httpResp.StatusCode >= 400 {
		logResponseBody(t, httpResp, testName)
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
	if strings.Contains(err.Error(), "undefined response type") {
		if httpResp != nil && httpResp.StatusCode == 400 {
			t.Logf("⚠️ 400 BAD REQUEST with undefined response type in %s", testName)
			logResponseBody(t, httpResp, testName)
			if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
				body := string(apiErr.Body())
				t.Logf("400 Error Response Body: %s", body)
			}
			t.Fatalf("%s: 400 Bad Request with undefined response type - needs investigation: %v", testName, err)
		}
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
		"portfolio margin not enabled",
		"account not enabled for portfolio margin",
		"pmargin not supported",
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

// logResponseBody logs the raw response body for debugging purposes
func logResponseBody(t *testing.T, httpResp *http.Response, context string) {
	if httpResp != nil && httpResp.Body != nil {
		body, readErr := io.ReadAll(httpResp.Body)
		if readErr == nil {
			bodyStr := string(body)
			if httpResp.StatusCode == 400 {
				t.Logf("🚨 %s - 400 ERROR RAW RESPONSE BODY: %s", context, bodyStr)
				// Also try to format as JSON if possible for better readability
				if strings.HasPrefix(strings.TrimSpace(bodyStr), "{") {
					t.Logf("🚨 %s - 400 ERROR (Formatted): %s", context, bodyStr)
				}
			} else {
				t.Logf("%s - Raw response body: %s", context, bodyStr)
			}
		} else {
			if httpResp.StatusCode == 400 {
				t.Logf("🚨 %s - Failed to read 400 error response body: %v", context, readErr)
			} else {
				t.Logf("%s - Failed to read response body: %v", context, readErr)
			}
		}
	} else {
		if httpResp != nil && httpResp.StatusCode == 400 {
			t.Logf("🚨 %s - 400 error but no response body available", context)
		}
	}
}

// handlePortfolioMarginError checks for portfolio margin specific errors
func handlePortfolioMarginError(t *testing.T, err error, testName string) bool {
	if err == nil {
		return false
	}
	
	errMsg := err.Error()
	portfolioMarginPatterns := []string{
		"portfolio margin not enabled",
		"account not enabled for portfolio margin",
		"pmargin account required",
		"insufficient portfolio margin permissions",
		"portfolio margin account not found",
		"PM account required",
		"cross margin not supported",
		"margin account not found",
	}
	
	for _, pattern := range portfolioMarginPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			t.Skipf("%s requires portfolio margin account: %s", testName, pattern)
			return true
		}
	}
	
	return false
}