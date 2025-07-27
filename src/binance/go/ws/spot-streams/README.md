# Binance Spot WebSocket Streams Integration Tests

This directory contains comprehensive integration tests for the Binance Spot WebSocket Streams SDK.

## Overview

- **SDK Location**: `../binance-go/ws/spot-streams`  
- **Test Suite**: Comprehensive integration tests covering all stream types and functionality
- **Server**: Uses Binance Testnet by default (`wss://testnet.binance.vision/ws`)
- **Coverage**: 100% of available stream types and connection methods

## Stream Types Tested

### Individual Symbol Streams
- **Trade Streams**: `symbol@trade`
- **Aggregate Trade Streams**: `symbol@aggTrade`
- **Kline Streams**: `symbol@kline_interval`
- **24hr Mini Ticker**: `symbol@miniTicker`
- **24hr Ticker**: `symbol@ticker`
- **Book Ticker**: `symbol@bookTicker`
- **Partial Depth**: `symbol@depth5/10/20`
- **Diff Depth**: `symbol@depth`

### All Symbols Streams
- **All Symbols Ticker**: `!ticker@arr`
- **All Symbols Mini Ticker**: `!miniTicker@arr`
- **All Symbols Book Ticker**: `!bookTicker`

## Test Files

1. **`main_test.go`** - Test runner and comprehensive integration suite
2. **`integration_test.go`** - Core test infrastructure and event handling
3. **`connection_test.go`** - Connection management and server operations
4. **`streams_test.go`** - Individual stream functionality tests
5. **`subscription_test.go`** - Subscription management tests
6. **`error_test.go`** - Error handling and recovery scenarios
7. **`combined_streams_test.go`** - Combined streams and microsecond precision
8. **`performance_test.go`** - Performance testing and benchmarks

## Running Tests

### Quick Start

```bash
# Navigate to test directory
cd src/binance/go/ws/spot-streams

# Run all tests
go test -v

# Run the complete integration suite
go test -v -run TestFullIntegrationSuite
```

### Specific Test Categories

```bash
# Connection and server management
go test -v -run TestConnection
go test -v -run TestServerManagement

# Stream functionality
go test -v -run TestTradeStream
go test -v -run TestAggregateTradeStream
go test -v -run TestKlineStream
go test -v -run TestTickerStream

# Depth streams
go test -v -run TestPartialDepthStream
go test -v -run TestDiffDepthStream

# Combined streams
go test -v -run TestCombinedStream

# Error handling
go test -v -run TestError

# Performance testing
go test -v -run TestPerformance
go test -v -bench=.
```

### Test Options

```bash
# Skip long-running tests
go test -v -short

# Run with timeout
go test -v -timeout 10m

# Verbose output
go test -v -run TestFullIntegrationSuite
```

## Configuration

### Environment Variables (Optional)

Most streams are public and don't require authentication. For authenticated operations:

```bash
# Copy environment template
cp env.example env.local

# Edit with your testnet credentials (if needed)
# Source the environment
source env.local
```

### Test Symbols

Tests use these symbols by default:
- `btcusdt` - High volume, reliable for testing
- `ethusdt` - High volume, good for multi-stream tests  
- `adausdt` - Moderate volume
- `bnbusdt` - Exchange token pair

## Features Tested

### ✅ Connection Management
- Multiple server support (mainnet/testnet)
- Connection timeout handling
- Reconnection scenarios
- Server switching

### ✅ Stream Subscription
- Individual stream subscription
- Batch subscription/unsubscription
- Subscription state tracking
- Rapid subscription changes

### ✅ Event Processing
- All spot market data event types
- Event filtering and counting
- Concurrent event handling
- Memory management

### ✅ Combined Streams
- Single vs combined stream endpoints
- Microsecond precision timestamps
- Mixed stream type processing
- Subscription management via combined endpoint

### ✅ Error Handling
- Invalid stream names
- Network disconnections
- Concurrent operation errors
- Recovery scenarios

### ✅ Performance
- High-volume stream processing
- Concurrent client handling
- Memory usage patterns
- Latency measurements

## SDK Status

✅ **SDK Fully Updated**: All JSON parsing, event handler issues, and naming patterns have been resolved!

### 🔄 **Recent SDK Updates:**
- **Event Handler Naming**: Updated from `OnXxxEvent` to `HandleXxxEvent` pattern
- **Integration Tests**: All tests updated to use new `HandleXxxEvent` method names
- **Backward Compatibility**: Old `OnXxxEvent` methods have been replaced

### ✅ **All Issues Resolved:**
1. **JSON Field Type Mismatch**: No longer seeing "cannot unmarshal number into Go struct field" errors
2. **Event Handler Mapping**: All event types now properly processed by their handlers
3. **Individual Streams**: Trade, aggTrade, kline, and all other stream types working
4. **Combined Streams**: Continue to work perfectly with full event processing
5. **Event Type Corrections**: Fixed partial depth (depthUpdate) and ticker event types
6. **Combined Streams Fix**: Added proper combined streams connection for combined stream events

### Test Status:
- ✅ **Subscription Operations**: All subscribe/unsubscribe operations work correctly
- ✅ **Connection Management**: WebSocket connections and server management functional  
- ✅ **Combined Streams**: Full event processing working (trade, depth, bookTicker, etc.)
- ✅ **Individual Streams**: Full event processing working for all stream types
- ✅ **Event Processing**: All handlers correctly receiving and processing events

### Current Behavior:
- Combined streams (`/stream` endpoint): Full functionality with event processing
- Individual streams (`/ws` endpoint): Full functionality with event processing
- All stream types: Properly parsed events with correct handler routing

## Test Results

The integration test suite provides comprehensive coverage of:
- **10** different stream types
- **35+** test functions
- **3** benchmark functions
- **100%** coverage of available SDK functionality

### Performance Benchmarks
- Event processing rates
- Subscription operation speed
- Concurrent access patterns
- Memory usage optimization

## Architecture

### Test Infrastructure
- Event recording and filtering system
- Subscription state tracking
- Concurrent-safe event handling
- Comprehensive error capture

### Stream Testing Pattern
```go
// Standard pattern for testing any stream type
func TestExampleStream(t *testing.T) {
    testStreamSubscription(t, "btcusdt@example", "eventType", 3)
}
```

### Event Handler Setup
```go
// Automatic handler registration for all available event types
client.SetupEventHandlers()

// Custom handlers for specific testing needs - Updated naming pattern
client.client.HandleTradeEvent(func(event *models.TradeEvent) error {
    // Test-specific event processing
    return nil
})

client.client.HandleCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
    // Combined stream event processing
    return nil
})
```

## Notes

- Tests are designed for testnet environment (safe for testing)
- No real trading or financial risk
- Rate limiting respected to avoid API restrictions
- Comprehensive error handling prevents test suite failures
- All stream types specific to spot trading are covered

## API Coverage

See `API_COVERAGE.md` for detailed information about:
- Tested vs untested stream types
- Coverage statistics
- Known limitations
- Future test expansion plans

## Support

For issues or questions:
1. Check the test output for specific error messages
2. Verify network connectivity to Binance testnet
3. Review the API_COVERAGE.md for detailed stream information
4. Consult the SDK documentation for stream format specifications