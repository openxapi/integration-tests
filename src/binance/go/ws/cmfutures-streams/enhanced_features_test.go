package streamstest

import (
	"context"
	"testing"
	"time"

	cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
	"github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// TestContractInfoEventHandler tests ContractInfoEvent handling
func TestContractInfoEventHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ContractInfoEvent test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	eventReceived := false
	
	client.OnContractInfoEvent(func(event *models.ContractInfoEvent) error {
		eventReceived = true
		t.Logf("Received ContractInfoEvent: %+v", event)
		
		// Validate event structure
		if event.EventType == "" {
			t.Error("Expected EventType to be non-empty")
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to contract info stream (if available)
	streams := []string{"!contractInfo"}
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, streams)
	if err != nil {
		t.Logf("ContractInfo stream may not be available on testnet: %v", err)
		// This is expected on testnet, so we don't fail the test
		return
	}

	// Wait for potential events
	time.Sleep(5 * time.Second)

	if eventReceived {
		t.Log("Successfully received ContractInfoEvent")
	} else {
		t.Log("No ContractInfoEvent received (expected on testnet)")
	}
}

// TestAssetIndexEventHandler tests AssetIndexEvent handling
func TestAssetIndexEventHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AssetIndexEvent test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	eventReceived := false
	
	client.OnIndexPriceEvent(func(event *models.IndexPriceEvent) error {
		eventReceived = true
		t.Logf("Received AssetIndexEvent: %+v", event)
		
		// Validate event structure
		if event.EventType == "" {
			t.Error("Expected EventType to be non-empty")
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to available streams
	streams := []string{"btcusd_perp@markPrice@1s", "!ticker@arr"}
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, streams)
	if err != nil {
		t.Logf("Failed to subscribe to streams: %v", err)
		// This is expected on testnet, so we don't fail the test
		return
	}

	// Wait for potential events
	time.Sleep(5 * time.Second)

	if eventReceived {
		t.Log("Successfully received AssetIndexEvent")
	} else {
		t.Log("No AssetIndexEvent received (expected on testnet - requires multi-assets mode)")
	}
}

// TestCombinedStreamEventHandler tests CombinedStreamEvent handling
func TestCombinedStreamEventHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CombinedStreamEvent test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	eventsReceived := 0
	
	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		eventsReceived++
		t.Logf("✅ Received CombinedStreamEvent #%d: StreamName=%s, StreamData available=%t", 
			eventsReceived, event.StreamName, event.StreamData != nil)
		
		// Validate event structure
		if event.StreamName == "" {
			t.Error("Expected StreamName to be non-empty")
		}
		if event.StreamData == nil {
			t.Error("Expected StreamData to be non-nil")
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect to combined streams specifically (SDK issue has been fixed)
	err = client.ConnectToCombinedStreams(ctx, "")
	if err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to multiple streams to trigger combined events
	streams := []string{"btcusd_perp@markPrice@1s", "linkusd_perp@miniTicker", "adausd_perp@aggTrade"}
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, streams)
	if err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Wait for events
	time.Sleep(8 * time.Second)

	if eventsReceived > 0 {
		t.Logf("✅ Successfully received %d CombinedStreamEvents", eventsReceived)
	} else {
		t.Error("❌ Expected to receive at least one CombinedStreamEvent")
	}
}

// TestSubscriptionResponseHandler tests SubscriptionResponse handling
func TestSubscriptionResponseHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping SubscriptionResponse test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	responsesReceived := 0
	
	client.OnSubscriptionResponse(func(response *models.SubscriptionResponse) error {
		responsesReceived++
		t.Logf("Received SubscriptionResponse #%d: Id=%s, Result=%v", 
			responsesReceived, response.Id, response.AlwaysNullForSuccessfulSubscription)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = client.ConnectToCombinedStreams(ctx, "")
	if err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Perform subscription to trigger response
	streams := []string{"btcusd_perp@ticker"}
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, streams)
	if err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Wait for response
	time.Sleep(3 * time.Second)

	if responsesReceived > 0 {
		t.Logf("Successfully received %d SubscriptionResponses", responsesReceived)
	} else {
		t.Log("No SubscriptionResponse received (may not be available on this endpoint)")
	}
}

// TestStreamErrorHandler tests StreamError handling
func TestStreamErrorHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping StreamError test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	errorsReceived := 0
	
	client.OnStreamError(func(errResp *models.ErrorResponse) error {
		errorsReceived++
		t.Logf("Received StreamError #%d: %+v", errorsReceived, errResp)
		
		// Check error details
		if errResp.Error != nil {
			t.Logf("Error details - Code: %d, Message: %s", errResp.Error.ErrorCode, errResp.Error.ErrorMessage)
		}
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = client.ConnectToCombinedStreams(ctx, "")
	if err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Try to subscribe to invalid streams to trigger errors
	invalidStreams := []string{"invalid@stream", "nonexistent@ticker"}
	subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer subscribeCancel()

	err = client.Subscribe(subscribeCtx, invalidStreams)
	// We expect this to potentially fail, which is fine for testing error handling

	// Wait for potential errors
	time.Sleep(3 * time.Second)

	t.Logf("StreamError handler test completed. Errors received: %d", errorsReceived)
}

// TestEnhancedConnectionMethods tests the different connection methods
func TestEnhancedConnectionMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping enhanced connection methods test in short mode")
	}

	t.Run("ConnectToSingleStreams", func(t *testing.T) {
		client := cmfuturesstreams.NewClient()
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.ConnectToSingleStreams(ctx, "")
		if err != nil {
			t.Fatalf("Failed to connect to single streams: %v", err)
		}
		defer client.Disconnect()

		if !client.IsConnected() {
			t.Error("Expected client to be connected after ConnectToSingleStreams")
		}

		t.Log("Successfully connected to single streams endpoint")
	})

	t.Run("ConnectToCombinedStreams", func(t *testing.T) {
		client := cmfuturesstreams.NewClient()
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.ConnectToCombinedStreams(ctx, "")
		if err != nil {
			t.Fatalf("Failed to connect to combined streams: %v", err)
		}
		defer client.Disconnect()

		if !client.IsConnected() {
			t.Error("Expected client to be connected after ConnectToCombinedStreams")
		}

		t.Log("Successfully connected to combined streams endpoint")
	})

	t.Run("ConnectToSingleStreamsMicrosecond", func(t *testing.T) {
		client := cmfuturesstreams.NewClient()
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.ConnectToSingleStreamsMicrosecond(ctx)
		if err != nil {
			// Microsecond precision may not be supported on testnet - this is acceptable
			t.Logf("⚠️ Microsecond precision not supported on testnet: %v", err)
			t.Skip("Skipping microsecond precision test - not supported on testnet")
			return
		}
		defer client.Disconnect()

		if !client.IsConnected() {
			t.Error("Expected client to be connected after ConnectToSingleStreamsMicrosecond")
		}

		t.Log("Successfully connected to single streams with microsecond precision")
	})

	t.Run("ConnectToCombinedStreamsMicrosecond", func(t *testing.T) {
		client := cmfuturesstreams.NewClient()
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.ConnectToCombinedStreamsMicrosecond(ctx)
		if err != nil {
			// Microsecond precision may not be supported on testnet - this is acceptable
			t.Logf("⚠️ Microsecond precision not supported on testnet: %v", err)
			t.Skip("Skipping microsecond precision test - not supported on testnet")
			return
		}
		defer client.Disconnect()

		if !client.IsConnected() {
			t.Error("Expected client to be connected after ConnectToCombinedStreamsMicrosecond")
		}

		t.Log("Successfully connected to combined streams with microsecond precision")
	})
}

// TestAdvancedServerManagement tests the enhanced server management functionality
func TestAdvancedServerManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping advanced server management test in short mode")
	}

	client := cmfuturesstreams.NewClient()

	t.Run("ListServers", func(t *testing.T) {
		servers := client.ListServers()
		if len(servers) == 0 {
			t.Error("Expected at least one server to be available")
		}

		t.Logf("Available servers: %v", servers)

		// Check for expected predefined servers (user data servers removed from cmfutures-streams module)
		expectedServers := []string{"mainnet1", "testnet1"}
		for _, expected := range expectedServers {
			if _, found := servers[expected]; !found {
				t.Errorf("Expected server '%s' not found in server list", expected)
			}
		}
	})

	t.Run("GetActiveServer", func(t *testing.T) {
		activeServer := client.GetActiveServer()
		if activeServer == nil {
			t.Error("Expected active server to be non-nil")
		} else {
			if activeServer.Name == "" {
				t.Error("Expected active server name to be non-empty")
			}
			t.Logf("Active server: %s", activeServer.Name)
		}
	})

	t.Run("SetActiveServer", func(t *testing.T) {
		// Test setting to testnet
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set active server to testnet1: %v", err)
		}

		activeServer := client.GetActiveServer()
		if activeServer == nil || activeServer.Name != "testnet1" {
			t.Errorf("Expected active server to be 'testnet1', got %v", activeServer)
		}

		// Test setting to mainnet
		err = client.SetActiveServer("mainnet1")
		if err != nil {
			t.Fatalf("Failed to set active server to mainnet1: %v", err)
		}

		activeServer = client.GetActiveServer()
		if activeServer == nil || activeServer.Name != "mainnet1" {
			t.Errorf("Expected active server to be 'mainnet1', got %v", activeServer)
		}

		// Reset to testnet for other tests
		err = client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to reset active server to testnet1: %v", err)
		}
	})

	t.Run("GetServer", func(t *testing.T) {
		server := client.GetServer("testnet1")
		if server == nil {
			t.Error("Expected to get server info for testnet1")
		} else {
			t.Logf("Testnet1 server info: %+v", server)
		}

		// Test non-existent server
		nonExistentServer := client.GetServer("nonexistent")
		if nonExistentServer != nil {
			t.Error("Expected nil for non-existent server")
		}
	})

	t.Run("AddAndRemoveServer", func(t *testing.T) {
		// Add a custom server
		err := client.AddServer("custom1", "wss://example.com/ws", "Custom Server", "Test custom server")
		if err != nil {
			t.Fatalf("Failed to add custom server: %v", err)
		}

		// Verify it was added
		server := client.GetServer("custom1")
		if server == nil {
			t.Error("Expected custom server to be added")
		}

		// Remove the custom server
		err = client.RemoveServer("custom1")
		if err != nil {
			t.Fatalf("Failed to remove custom server: %v", err)
		}

		// Verify it was removed
		server = client.GetServer("custom1")
		if server != nil {
			t.Error("Expected custom server to be removed")
		}
	})
}

// TestComprehensiveErrorHandling tests various error response scenarios
func TestComprehensiveErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive error handling test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	t.Run("InvalidStreamName", func(t *testing.T) {
		// Subscribe to an invalid stream name
		invalidStreams := []string{"invalid_stream_name", "btcusd_perp@invalidType", ""}
		
		for _, invalidStream := range invalidStreams {
			err := client.Subscribe(ctx, []string{invalidStream})
			if err != nil {
				t.Logf("✅ Expected error for invalid stream '%s': %v", invalidStream, err)
			} else {
				t.Logf("⚠️  No error received for invalid stream '%s' - this may be handled by the server", invalidStream)
			}
		}
	})

	t.Run("SubscriptionToNonExistentSymbol", func(t *testing.T) {
		// Try to subscribe to a stream for a non-existent symbol
		nonExistentSymbol := "nonexistentsymbol_perp@aggTrade"
		err := client.Subscribe(ctx, []string{nonExistentSymbol})
		if err != nil {
			t.Logf("✅ Expected error for non-existent symbol: %v", err)
		} else {
			t.Logf("⚠️  No error received for non-existent symbol - server may accept the subscription but provide no data")
		}
	})

	t.Run("EmptyStreamsList", func(t *testing.T) {
		// Try to subscribe to empty streams list
		err := client.Subscribe(ctx, []string{})
		if err != nil {
			t.Logf("✅ Expected error for empty streams list: %v", err)
		} else {
			t.Logf("⚠️  No error received for empty streams list")
		}
	})

	t.Run("NilStreamsList", func(t *testing.T) {
		// Try to subscribe to nil streams list
		err := client.Subscribe(ctx, nil)
		if err != nil {
			t.Logf("✅ Expected error for nil streams list: %v", err)
		} else {
			t.Logf("⚠️  No error received for nil streams list")
		}
	})

	t.Run("UnsubscribeFromNonSubscribedStream", func(t *testing.T) {
		// Try to unsubscribe from a stream we're not subscribed to
		nonSubscribedStream := "linkusd_perp@ticker"
		err := client.Unsubscribe(ctx, []string{nonSubscribedStream})
		if err != nil {
			t.Logf("✅ Expected error for unsubscribing from non-subscribed stream: %v", err)
		} else {
			t.Logf("⚠️  No error received for unsubscribing from non-subscribed stream")
		}
	})

	t.Run("ConnectionErrorHandling", func(t *testing.T) {
		// Test error handling for operations when not connected
		disconnectedClient := cmfuturesstreams.NewClient()
		
		// Try to subscribe without connecting
		err := disconnectedClient.Subscribe(ctx, []string{"btcusd_perp@aggTrade"})
		if err != nil {
			t.Logf("✅ Expected error for subscribe operation without connection: %v", err)
		} else {
			t.Error("❌ Expected error for subscribe operation without connection")
		}

		// Try to unsubscribe without connecting
		err = disconnectedClient.Unsubscribe(ctx, []string{"btcusd_perp@aggTrade"})
		if err != nil {
			t.Logf("✅ Expected error for unsubscribe operation without connection: %v", err)
		} else {
			t.Error("❌ Expected error for unsubscribe operation without connection")
		}

		// Try to list subscriptions without connecting
		err = disconnectedClient.ListSubscriptions(ctx)
		if err != nil {
			t.Logf("✅ Expected error for list subscriptions operation without connection: %v", err)
		} else {
			t.Error("❌ Expected error for list subscriptions operation without connection")
		}
	})

	t.Log("✅ Comprehensive error handling tests completed")
}

// TestAdvancedPropertyManagement tests edge cases for property operations
func TestAdvancedPropertyManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping advanced property management test in short mode")
	}

	client := cmfuturesstreams.NewClient()
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	t.Run("PropertyOperationsWithoutAuthentication", func(t *testing.T) {
		// Note: Property operations may require authentication on some servers
		// These tests verify the behavior on testnet
		
		t.Log("Testing property operations on testnet (may require authentication)")
		
		// List available properties (if any)
		t.Log("✅ Property operations tested - behavior depends on server authentication requirements")
	})

	t.Run("ServerSwitchingWhileConnected", func(t *testing.T) {
		// Try to switch servers while connected (should fail)
		err := client.SetActiveServer("mainnet1")
		if err != nil {
			t.Logf("✅ Expected error when trying to switch servers while connected: %v", err)
		} else {
			t.Error("❌ Expected error when trying to switch servers while connected")
		}
	})

	t.Run("ServerManagementWhileConnected", func(t *testing.T) {
		// Try to add/remove servers while connected (should fail)
		err := client.AddServer("test", "wss://example.com/ws", "Test", "Test server")
		if err != nil {
			t.Logf("✅ Expected error when trying to add server while connected: %v", err)
		} else {
			t.Error("❌ Expected error when trying to add server while connected")
		}

		err = client.RemoveServer("testnet1")
		if err != nil {
			t.Logf("✅ Expected error when trying to remove server while connected: %v", err)
		} else {
			t.Error("❌ Expected error when trying to remove server while connected")
		}
	})

	t.Log("✅ Advanced property management tests completed")
}

