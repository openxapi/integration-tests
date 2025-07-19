package main

import (
	"context"
	"io"
	"testing"
	"time"

	cmfutures "github.com/openxapi/binance-go/rest/cmfutures"
)

// TestCreateListenKey tests creating a new user data stream
func TestCreateListenKey(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "CreateListenKey", func(t *testing.T, client *cmfutures.APIClient, ctx context.Context) {
					req := client.FuturesAPI.CreateListenKeyV1(ctx)
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "CreateListenKey") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "UserDataStreamOperation")
						t.Fatalf("Create listen key failed: %v", err)
					}
					
					if resp.ListenKey == nil {
						t.Fatal("ListenKey is nil")
					}
					
					listenKey := *resp.ListenKey
					if listenKey == "" {
						t.Fatal("ListenKey is empty")
					}
					
					t.Logf("Created listen key: %s", listenKey)
					
					// Clean up: delete the listen key
					time.Sleep(100 * time.Millisecond)
					deleteReq := client.FuturesAPI.DeleteListenKeyV1(ctx)
					
					_, _, deleteErr := deleteReq.Execute()
					if deleteErr == nil {
						t.Logf("Deleted listen key successfully")
					} else {
						t.Logf("Failed to delete listen key: %v", deleteErr)
					}
					
					// Verify the response structure (delete response typically empty)
					t.Logf("Delete response received")
				})
			})
			break
		}
	}
}

// TestUpdateListenKey tests keeping alive a user data stream
func TestUpdateListenKey(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "UpdateListenKey", func(t *testing.T, client *cmfutures.APIClient, ctx context.Context) {
					// First create a listen key
					createReq := client.FuturesAPI.CreateListenKeyV1(ctx)
					
					createResp, _, createErr := createReq.Execute()
					if createErr != nil {
						checkAPIError(t, createErr, nil, "CreateListenKeyForTest")
						t.Fatalf("Failed to create listen key for update test: %v", createErr)
					}
					
					if createResp.ListenKey == nil {
						t.Fatal("Created listen key is nil")
					}
					
					listenKey := *createResp.ListenKey
					t.Logf("Created listen key for update: %s", listenKey)
					
					// Update (keepalive) the listen key
					time.Sleep(100 * time.Millisecond)
					req := client.FuturesAPI.UpdateListenKeyV1(ctx)
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "UpdateListenKey") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "UserDataStreamOperation")
						// Clean up even if update fails
						deleteReq := client.FuturesAPI.DeleteListenKeyV1(ctx)
						deleteReq.Execute()
						t.Fatalf("Update listen key failed: %v", err)
					}
					
					t.Logf("Updated listen key successfully")
					
					// Verify the response structure (typically empty for success)
					if resp != nil {
						t.Logf("Update response received")
					}
					
					// Clean up: delete the listen key
					time.Sleep(100 * time.Millisecond)
					deleteReq := client.FuturesAPI.DeleteListenKeyV1(ctx)
					
					_, _, deleteErr := deleteReq.Execute()
					if deleteErr == nil {
						t.Logf("Deleted listen key successfully")
					} else {
						t.Logf("Failed to delete listen key: %v", deleteErr)
					}
				})
			})
			break
		}
	}
}

// TestDeleteListenKey tests closing out a user data stream
func TestDeleteListenKey(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "DeleteListenKey", func(t *testing.T, client *cmfutures.APIClient, ctx context.Context) {
					// First create a listen key
					createReq := client.FuturesAPI.CreateListenKeyV1(ctx)
					
					createResp, _, createErr := createReq.Execute()
					if createErr != nil {
						checkAPIError(t, createErr, nil, "CreateListenKeyForTest")
						t.Fatalf("Failed to create listen key for delete test: %v", createErr)
					}
					
					if createResp.ListenKey == nil {
						t.Fatal("Created listen key is nil")
					}
					
					listenKey := *createResp.ListenKey
					t.Logf("Created listen key for delete: %s", listenKey)
					
					// Delete the listen key
					time.Sleep(100 * time.Millisecond)
					req := client.FuturesAPI.DeleteListenKeyV1(ctx)
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "DeleteListenKey") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp, "UserDataStreamOperation")
						t.Fatalf("Delete listen key failed: %v", err)
					}
					
					t.Logf("Deleted listen key successfully")
					
					// Verify the response structure (typically empty for success)
					if resp != nil {
						t.Logf("Delete response received")
					}
					
					// Try to update the deleted listen key to verify it's gone
					time.Sleep(100 * time.Millisecond)
					updateReq := client.FuturesAPI.UpdateListenKeyV1(ctx)
					
					_, updateResp, updateErr := updateReq.Execute()
					if updateErr != nil {
						t.Logf("Update of deleted listen key failed as expected: %v", updateErr)
						
						// Print detailed error information
						if updateResp != nil {
							t.Logf("Response status: %s", updateResp.Status)
							if updateResp.Body != nil {
								bodyBytes, _ := io.ReadAll(updateResp.Body)
								t.Logf("Response body: %s", string(bodyBytes))
								updateResp.Body.Close()
							}
						}
						
						// Check if it's a GenericOpenAPIError and extract details
						if apiErr, ok := updateErr.(*cmfutures.GenericOpenAPIError); ok {
							t.Logf("API Error body: %s", string(apiErr.Body()))
							t.Logf("API Error message: %s", apiErr.Error())
						}
					} else {
						t.Logf("Update of deleted listen key unexpectedly succeeded")
					}
				})
			})
			break
		}
	}
}

// TestListenKeyLifecycle tests the full lifecycle of a listen key
func TestListenKeyLifecycle(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "ListenKeyLifecycle", func(t *testing.T, client *cmfutures.APIClient, ctx context.Context) {
					// Step 1: Create a listen key
					createReq := client.FuturesAPI.CreateListenKeyV1(ctx)
					
					createResp, _, createErr := createReq.Execute()
					if createErr != nil {
						checkAPIError(t, createErr, nil, "CreateListenKeyForTest")
						t.Fatalf("Failed to create listen key: %v", createErr)
					}
					
					if createResp.ListenKey == nil {
						t.Fatal("Created listen key is nil")
					}
					
					listenKey := *createResp.ListenKey
					t.Logf("Step 1: Created listen key: %s", listenKey)
					
					// Step 2: Update (keepalive) the listen key multiple times
					for i := 0; i < 3; i++ {
						time.Sleep(100 * time.Millisecond)
						updateReq := client.FuturesAPI.UpdateListenKeyV1(ctx)
						
						_, _, updateErr := updateReq.Execute()
						if updateErr != nil {
							checkAPIError(t, updateErr, nil, "UpdateListenKeyInTest")
							t.Logf("Step 2.%d: Update failed: %v", i+1, updateErr)
						} else {
							t.Logf("Step 2.%d: Updated listen key successfully", i+1)
						}
					}
					
					// Step 3: Delete the listen key
					time.Sleep(100 * time.Millisecond)
					deleteReq := client.FuturesAPI.DeleteListenKeyV1(ctx)
					
					_, _, deleteErr := deleteReq.Execute()
					if deleteErr != nil {
						checkAPIError(t, deleteErr, nil, "DeleteListenKeyInTest")
						t.Logf("Step 3: Delete failed: %v", deleteErr)
					} else {
						t.Logf("Step 3: Deleted listen key successfully")
					}
					
					// Step 4: Verify the listen key is gone by trying to update it
					time.Sleep(100 * time.Millisecond)
					verifyReq := client.FuturesAPI.UpdateListenKeyV1(ctx)
					
					_, _, verifyErr := verifyReq.Execute()
					if verifyErr != nil {
						t.Logf("Step 4: Verification - listen key properly deleted (update failed as expected)")
					} else {
						t.Logf("Step 4: Verification - listen key may still exist (update succeeded)")
					}
					
					t.Logf("Listen key lifecycle test completed")
				})
			})
			break
		}
	}
}

// TestMultipleListenKeys tests creating and managing multiple listen keys
func TestMultipleListenKeys(t *testing.T) {
	configs := getTestConfigs()
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "MultipleListenKeys", func(t *testing.T, client *cmfutures.APIClient, ctx context.Context) {
					var listenKeys []string
					
					// Create multiple listen keys
					for i := 0; i < 3; i++ {
						createReq := client.FuturesAPI.CreateListenKeyV1(ctx)
						
						createResp, _, createErr := createReq.Execute()
						if createErr != nil {
							checkAPIError(t, createErr, nil, "CreateListenKeyForTest")
							t.Logf("Failed to create listen key %d: %v", i+1, createErr)
							continue
						}
						
						if createResp.ListenKey == nil {
							t.Logf("Created listen key %d is nil", i+1)
							continue
						}
						
						listenKey := *createResp.ListenKey
						listenKeys = append(listenKeys, listenKey)
						t.Logf("Created listen key %d: %s", i+1, listenKey)
						
						time.Sleep(100 * time.Millisecond)
					}
					
					if len(listenKeys) == 0 {
						t.Skip("No listen keys created successfully")
						return
					}
					
					t.Logf("Created %d listen keys", len(listenKeys))
					
					// Update all listen keys
					for i, _ := range listenKeys {
						updateReq := client.FuturesAPI.UpdateListenKeyV1(ctx)
						
						_, _, updateErr := updateReq.Execute()
						if updateErr != nil {
							checkAPIError(t, updateErr, nil, "UpdateListenKeyInTest")
							t.Logf("Failed to update listen key %d: %v", i+1, updateErr)
						} else {
							t.Logf("Updated listen key %d successfully", i+1)
						}
						
						time.Sleep(100 * time.Millisecond)
					}
					
					// Clean up: delete all listen keys
					for i, _ := range listenKeys {
						deleteReq := client.FuturesAPI.DeleteListenKeyV1(ctx)
						
						_, _, deleteErr := deleteReq.Execute()
						if deleteErr != nil {
							checkAPIError(t, deleteErr, nil, "DeleteListenKeyInTest")
							t.Logf("Failed to delete listen key %d: %v", i+1, deleteErr)
						} else {
							t.Logf("Deleted listen key %d successfully", i+1)
						}
						
						time.Sleep(100 * time.Millisecond)
					}
					
					t.Logf("Multiple listen keys test completed")
				})
			})
			break
		}
	}
}