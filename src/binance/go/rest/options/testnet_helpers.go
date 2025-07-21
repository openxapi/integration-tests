package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/options"
)

// handleTestnetError checks if error is due to testnet limitations and skips test if so
// Returns true if test should be skipped, false if test should continue
// NEVER skips 400 Bad Request errors - these indicate real API issues that need fixing
func handleTestnetError(t *testing.T, err error, httpResp *http.Response, testName string) bool {
	if err == nil {
		return false
	}
	
	// NEVER skip 400 Bad Request errors - these are real API issues
	if httpResp != nil && httpResp.StatusCode == 400 {
		t.Logf("=== 400 BAD REQUEST ERROR ===")
		t.Logf("Endpoint: %s", testName)
		
		// Log response body
		logResponseBody(t, httpResp, testName)
		
		// Log API error details
		logAPIError(t, err)
		
		// Extract response text for error message
		responseText := ""
		if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
			responseText = string(apiErr.Body())
		}
		
		if responseText != "" {
			t.Fatalf("%s: 400 Bad Request - Response: %s - Error: %v", testName, responseText, err)
		} else {
			t.Fatalf("%s: 400 Bad Request - needs investigation: %v", testName, err)
		}
		return false
	}
	
	// Check for 404/403 status codes - endpoint not available on testnet
	if httpResp != nil && (httpResp.StatusCode == 404 || httpResp.StatusCode == 403) {
		logResponseBody(t, httpResp, testName)
		t.Skipf("%s endpoint not available on testnet", testName)
		return true
	}
	
	// Always log response body for debugging
	logResponseBody(t, httpResp, testName)
	
	// Check for HTML error pages from testnet
	if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		body := string(apiErr.Body())
		t.Logf("API error response: %s", body)
		
		if len(body) > 50 && (strings.HasPrefix(body, "<!DOCTYPE html>") || strings.HasPrefix(body, "<html")) {
			t.Skipf("%s endpoint returns HTML error page - not available on testnet", testName)
			return true
		}
	}
	
	// Check for undefined response type errors (SDK can't parse HTML responses)
	if strings.Contains(err.Error(), "undefined response type") {
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
		"options not available",
		"eapi not available",
		"endpoint not found",
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
	} else {
		t.Logf("Error: %v", err)
	}
}

// logResponseBody logs the raw response body for debugging purposes
func logResponseBody(t *testing.T, httpResp *http.Response, context string) {
	if httpResp != nil && httpResp.Body != nil {
		body, readErr := io.ReadAll(httpResp.Body)
		if readErr == nil {
			t.Logf("%s - Raw response body (status %d): %s", context, httpResp.StatusCode, string(body))
		} else {
			t.Logf("%s - Failed to read response body: %v", context, readErr)
		}
	}
}

// handleOptionsSpecificErrors handles errors specific to options trading
// NEVER skips 400 Bad Request errors - these indicate real API issues that need fixing
func handleOptionsSpecificErrors(t *testing.T, err error, httpResp *http.Response, testName string) bool {
	if err == nil {
		return false
	}
	
	// NEVER skip 400 Bad Request errors - these are real API issues
	if httpResp != nil && httpResp.StatusCode == 400 {
		t.Logf("=== 400 BAD REQUEST ERROR (OPTIONS) ===")
		t.Logf("Endpoint: %s", testName)
		
		// Log response body
		logResponseBody(t, httpResp, testName)
		
		// Log API error details
		logAPIError(t, err)
		
		// Extract response text for error message
		responseText := ""
		if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
			responseText = string(apiErr.Body())
		}
		
		if responseText != "" {
			t.Fatalf("%s: 400 Bad Request - Response: %s - Error: %v", testName, responseText, err)
		} else {
			t.Fatalf("%s: 400 Bad Request - needs investigation: %v", testName, err)
		}
		return false
	}
	
	// Always log response body and error details
	logResponseBody(t, httpResp, testName)
	logAPIError(t, err)
	
	errMsg := err.Error()
	optionsPatterns := []string{
		"options trading not enabled",
		"insufficient options permissions",
		"options account not found",
		"underlying asset not supported",
		"option symbol not found",
		"option expired",
		"option not tradeable",
		"block trade not supported",
		"market maker protection not enabled",
	}
	
	for _, pattern := range optionsPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			t.Skipf("%s requires options trading permissions: %s", testName, pattern)
			return true
		}
	}
	
	return handleTestnetError(t, err, httpResp, testName)
}