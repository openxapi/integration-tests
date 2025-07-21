package main

import (
	"context"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestListenKeyManagement tests user data stream listen key management
func TestListenKeyManagement(t *testing.T) {
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeUSER_DATA {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Listen Key Management", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test creating listen key
					createReq := client.PortfolioMarginAPI.CreateListenKeyV1(ctx)
					createResp, httpResp, err := createReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Create Listen Key") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Create Listen Key") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("CreateListenKeyV1 failed: %v", err)
					}
					
					if createResp == nil {
						t.Fatal("Create response is nil")
					}
					
					if createResp.ListenKey == nil {
						t.Fatal("ListenKey is nil")
					}
					
					listenKey := *createResp.ListenKey
					t.Logf("Listen key created: %s", listenKey[:10]+"...")
					
					// Test updating listen key
					// Note: UpdateListenKeyV1 doesn't take listenKey parameter - it uses the existing one
					updateReq := client.PortfolioMarginAPI.UpdateListenKeyV1(ctx)
					updateResp, httpResp, err := updateReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Update Listen Key") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Logf("UpdateListenKeyV1 failed (may not be supported): %v", err)
					} else {
						t.Logf("Listen key updated successfully: %+v", updateResp)
					}
					
					// Test deleting listen key
					// Note: DeleteListenKeyV1 doesn't take listenKey parameter - it deletes the current user's listen key
					deleteReq := client.PortfolioMarginAPI.DeleteListenKeyV1(ctx)
					deleteResp, httpResp, err := deleteReq.Execute()
					
					if handleTestnetError(t, err, httpResp, "Delete Listen Key") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Logf("DeleteListenKeyV1 failed (may not be supported): %v", err)
					} else {
						t.Logf("Listen key deleted successfully: %+v", deleteResp)
					}
				})
			})
		}
	}
}