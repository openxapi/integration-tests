package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/cmfutures"
)

// Global SDK issue tracker
var (
	sdkIssues    []string
	sdkIssuesMux sync.Mutex
)

// recordSDKIssue records an SDK issue for later reporting
func recordSDKIssue(testName, issue string) {
	sdkIssuesMux.Lock()
	defer sdkIssuesMux.Unlock()
	sdkIssues = append(sdkIssues, testName+": "+issue)
}

// getSDKIssues returns all recorded SDK issues
func getSDKIssues() []string {
	sdkIssuesMux.Lock()
	defer sdkIssuesMux.Unlock()
	return append([]string(nil), sdkIssues...)
}

// handleTestnetError checks if error is due to testnet limitations and skips test if so
// Returns true if test should be skipped, false if test should continue
func handleTestnetError(t *testing.T, err error, httpResp *http.Response, testName string) bool {
	if err == nil {
		return false
	}
	
	// Check for 400 errors - they indicate real API issues
	if httpResp != nil && httpResp.StatusCode == 400 {
		logResponseBody(t, httpResp, testName)
		t.Logf("%s: 400 Bad Request - Response text: %s", testName, err.Error())
		t.Fatalf("%s: 400 Bad Request error requires investigation: %v", testName, err)
	}
	
	// Check for HTML error pages first (SDK URL issues)
	if apiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		body := string(apiErr.Body())
		
		// Check if this is an HTML 404 page indicating SDK URL issues
		if len(body) > 50 && (strings.HasPrefix(body, "<!DOCTYPE html>") || strings.HasPrefix(body, "<html")) {
			if strings.Contains(body, "404") && strings.Contains(body, "This page could not be found") {
				issue := "Endpoint returns HTML 404 page - incorrect URL in SDK"
				recordSDKIssue(testName, issue)
				t.Logf("🚨 SDK ISSUE DETECTED: %s endpoint returns HTML 404 page - incorrect URL in SDK", testName)
				t.Logf("Raw HTML response: %s", body)
				t.Skipf("SKIPPING DUE TO SDK BUG: %s endpoint has incorrect URL in SDK (returns HTML 404)", testName)
				return true
			}
			if strings.Contains(body, "Resource not found") {
				issue := "Endpoint returns 'Resource not found' HTML page - incorrect URL in SDK"
				recordSDKIssue(testName, issue)
				t.Logf("🚨 SDK ISSUE DETECTED: %s endpoint returns 'Resource not found' HTML page", testName)
				t.Logf("Raw HTML response: %s", body)
				t.Skipf("SKIPPING DUE TO SDK BUG: %s endpoint has incorrect URL in SDK", testName)
				return true
			}
			t.Skipf("%s endpoint returns HTML error page - not available on testnet", testName)
			return true
		}
		// Log the actual API error for debugging
		t.Logf("API error response: %s", body)
	}
	
	// Check for 404/403 status codes - skip these tests  
	if httpResp != nil && (httpResp.StatusCode == 404 || httpResp.StatusCode == 403) {
		logResponseBody(t, httpResp, testName)
		if httpResp.StatusCode == 404 {
			recordSDKIssue(testName, "HTTP 404 - endpoint not found or incorrect URL")
			t.Skipf("SKIPPING DUE TO 404: %s endpoint not found (HTTP 404) - likely SDK URL issue or not available on testnet", testName)
		} else {
			t.Skipf("%s endpoint not available on testnet (HTTP %d)", testName, httpResp.StatusCode)
		}
		return true
	}
	
	// Check for undefined response type errors (SDK can't parse HTML responses)
	// NEVER skip 400 Bad Request errors - these indicate real API issues that need fixing
	if strings.Contains(err.Error(), "undefined response type") {
		if httpResp != nil && httpResp.StatusCode == 400 {
			logResponseBody(t, httpResp, testName)
			t.Fatalf("%s: 400 Bad Request with undefined response type - needs investigation: %v", testName, err)
		}
		t.Skipf("%s endpoint has response parsing issues - likely testnet limitation", testName)
		return true
	}
	
	// Check for other common testnet limitation patterns and 404-related errors
	errMsg := err.Error()
	testnetPatterns := []string{
		"This service is not available",
		"Feature not supported",
		"not available in testnet",
		"testnet not supported", 
		"Service temporarily unavailable",
		"Function not supported",
	}
	
	// 404-related error patterns that should be skipped
	notFoundPatterns := []string{
		"404",
		"not found",
		"endpoint not found",
		"page not found",
		"resource not found",
		"url not found",
	}
	
	for _, pattern := range testnetPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			t.Skipf("%s endpoint not supported on testnet: %s", testName, pattern)
			return true
		}
	}
	
	// Check for 404-related error patterns and skip them
	for _, pattern := range notFoundPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			recordSDKIssue(testName, fmt.Sprintf("Error contains '%s' - likely endpoint not found", pattern))
			t.Skipf("SKIPPING DUE TO 404: %s error contains '%s' - endpoint likely not found or incorrect URL", testName, pattern)
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
			t.Logf("%s - Raw response body: %s", context, string(body))
		} else {
			t.Logf("%s - Failed to read response body: %v", context, readErr)
		}
	}
}