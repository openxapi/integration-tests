# Integration Test Fix - Market Streams Event Handler Issue

## Problem Summary

The `TestMarketStreamsIntegration` tests are failing because they use raw SDK clients without proper event handler setup, while the working integration tests use a `StreamTestClient` wrapper.

## Root Cause

The market integration tests in `market_streams_integration_test.go` create raw SDK clients and set up event handlers directly:

```go
// BROKEN PATTERN (current)
client := cmfuturesstreams.NewClient()
client.OnAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
    // This handler will NEVER be called!
    return nil
})
```

But they need to use the same pattern as the working tests:

```go
// WORKING PATTERN 
client, _ := NewStreamTestClientDedicated(TestConfig{Name: "test"})
client.SetupEventHandlers()
```

## Recommended Fix

Update the market integration tests to use the `StreamTestClient` infrastructure instead of raw SDK clients.

### Example Fix for AggregateTradeStreamIntegration:

**Before (Broken)**:
```go
func testAggregateTradeStreamIntegration(t *testing.T) {
    client := cmfuturesstreams.NewClient()
    err := client.SetActiveServer("testnet1")
    // ... setup raw client handlers
}
```

**After (Fixed)**:
```go
func testAggregateTradeStreamIntegration(t *testing.T) {
    client, err := NewStreamTestClientDedicated(TestConfig{
        Name: "MarketStreamTest", 
        Description: "Market stream integration test"
    })
    if err != nil {
        t.Fatalf("Failed to create test client: %v", err)
    }
    defer client.Disconnect()
    
    // This sets up all the event handler infrastructure
    client.SetupEventHandlers()
    
    // Now use the working test pattern
    err = client.Subscribe(context.Background(), []string{"BTCUSD_PERP@aggTrade"})
    if err != nil {
        t.Fatalf("Failed to subscribe: %v", err)
    }
    
    // Use the working event waiting pattern
    err = client.WaitForEventsByType("aggTrade", 1, 10*time.Second)
    if err != nil {
        t.Error("Expected to receive aggregate trade events")
    } else {
        events := client.GetEventsByType("aggTrade")
        t.Logf("Successfully received %d aggTrade events", len(events))
    }
}
```

## Benefits of This Fix

1. **Uses proven working pattern** - Same infrastructure as successful tests
2. **Proper event handler setup** - All event handlers are initialized correctly  
3. **Event capturing** - Can wait for and count specific event types
4. **Consistent testing** - Same patterns across all integration tests
5. **Better error handling** - Graceful timeout handling already implemented

## Implementation Steps

1. Update each failing test function in `market_streams_integration_test.go`
2. Replace raw SDK client creation with `NewStreamTestClientDedicated`
3. Replace manual event counting with `WaitForEventsByType` and `GetEventsByType`
4. Update test expectations to use the working patterns

This fix addresses the immediate test failures while highlighting the underlying SDK event handler initialization issue that needs to be resolved in the SDK itself.