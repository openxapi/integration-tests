package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/options"
)

// testCreateListenKey tests the create listen key endpoint
func testCreateListenKey(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.CreateListenKeyV1(ctx).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "CreateListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("CreateListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	if resp.ListenKey == nil {
		t.Fatal("ListenKey should not be nil")
	}
	
	listenKey := *resp.ListenKey
	t.Logf("Created listen key: %s", listenKey)
	
	// Test update listen key
	testUpdateListenKey(t, client, ctx, listenKey)
	
	// Test delete listen key
	testDeleteListenKey(t, client, ctx, listenKey)
}

// testUpdateListenKey tests the update listen key endpoint
func testUpdateListenKey(t *testing.T, client *openapi.APIClient, ctx interface{}, listenKey string) {
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.UpdateListenKeyV1(ctx.(context.Context)).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "UpdateListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("UpdateListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Updated listen key successfully")
}

// testDeleteListenKey tests the delete listen key endpoint
func testDeleteListenKey(t *testing.T, client *openapi.APIClient, ctx interface{}, listenKey string) {
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.DeleteListenKeyV1(ctx.(context.Context)).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "DeleteListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("DeleteListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Deleted listen key successfully")
}

// testUserDataStreamLifecycle tests the complete user data stream lifecycle
func testUserDataStreamLifecycle(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	// Create listen key
	rateLimiter.WaitForRateLimit()
	
	createResp, httpResp, err := client.OptionsAPI.CreateListenKeyV1(ctx).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "CreateListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("CreateListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if createResp == nil || createResp.ListenKey == nil {
		t.Fatal("Failed to create listen key")
	}
	
	listenKey := *createResp.ListenKey
	t.Logf("Created listen key: %s", listenKey)
	
	// Update listen key (keepalive)
	rateLimiter.WaitForRateLimit()
	
	updateResp, httpResp, err := client.OptionsAPI.UpdateListenKeyV1(ctx).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "UpdateListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("UpdateListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if updateResp == nil {
		t.Fatal("Update response should not be nil")
	}
	
	t.Logf("Updated listen key successfully")
	
	// Delete listen key
	rateLimiter.WaitForRateLimit()
	
	deleteResp, httpResp, err := client.OptionsAPI.DeleteListenKeyV1(ctx).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "DeleteListenKeyV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("DeleteListenKeyV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if deleteResp == nil {
		t.Fatal("Delete response should not be nil")
	}
	
	t.Logf("Deleted listen key successfully")
	t.Logf("User data stream lifecycle test completed successfully")
}