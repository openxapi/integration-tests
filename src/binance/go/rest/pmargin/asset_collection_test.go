package main

import (
	"context"
	"os"
	"testing"

	openapi "github.com/openxapi/binance-go/rest/pmargin"
)

// TestAssetCollection tests fund collection by asset
func TestAssetCollection(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_ASSET_COLLECTION") != "true" {
		t.Skip("Asset collection test disabled - enable with BINANCE_TEST_PMARGIN_ASSET_COLLECTION=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Asset Collection", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test asset collection - using USDT as example
					req := client.PortfolioMarginAPI.CreateAssetCollectionV1(ctx).
						Asset("USDT").
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Asset Collection") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Asset Collection") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("CreateAssetCollectionV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Asset collection completed successfully")
					if resp.Msg != nil {
						t.Logf("Message: %s", *resp.Msg)
					}
				})
			})
		}
	}
}

// TestAutoCollection tests fund auto-collection
func TestAutoCollection(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_AUTO_COLLECTION") != "true" {
		t.Skip("Auto collection test disabled - enable with BINANCE_TEST_PMARGIN_AUTO_COLLECTION=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "Auto Collection", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					req := client.PortfolioMarginAPI.CreateAutoCollectionV1(ctx).
						Timestamp(generateTimestamp())
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "Auto Collection") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "Auto Collection") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("CreateAutoCollectionV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("Auto collection completed successfully")
					if resp.Msg != nil {
						t.Logf("Message: %s", *resp.Msg)
					}
				})
			})
		}
	}
}

// TestBNBTransfer tests BNB transfer
func TestBNBTransfer(t *testing.T) {
	if os.Getenv("BINANCE_TEST_PMARGIN_BNB_TRANSFER") != "true" {
		t.Skip("BNB transfer test disabled - enable with BINANCE_TEST_PMARGIN_BNB_TRANSFER=true")
	}
	
	configs := getTestConfigs()
	
	for _, config := range configs {
		if config.AuthType >= AuthTypeTRADE {
			t.Run(config.Name, func(t *testing.T) {
				testEndpoint(t, config, "BNB Transfer", func(t *testing.T, client *openapi.APIClient, ctx context.Context) {
					// Test BNB transfer with small amount
					req := client.PortfolioMarginAPI.CreateBnbTransferV1(ctx).
						Amount("0.001").
						TransferSide("TO_UM"). // Transfer to UM futures
						Timestamp(generateTimestamp())
					
					resp, httpResp, err := req.Execute()
					
					if handleTestnetError(t, err, httpResp, "BNB Transfer") {
						return
					}
					
					if handlePortfolioMarginError(t, err, "BNB Transfer") {
						return
					}
					
					if err != nil {
						checkAPIError(t, err, httpResp)
						t.Fatalf("CreateBnbTransferV1 failed: %v", err)
					}
					
					if resp == nil {
						t.Fatal("Response is nil")
					}
					
					t.Logf("BNB transfer completed successfully")
					if resp.TranId != nil {
						t.Logf("Transaction ID: %d", *resp.TranId)
					}
				})
			})
		}
	}
}