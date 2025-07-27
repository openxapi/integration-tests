# Binance Options WebSocket Streams Integration Tests

This directory contains comprehensive integration tests for the Binance Options WebSocket Streams SDK.

## Overview

The integration tests verify that the Binance Options WebSocket Streams SDK functions correctly against the live Binance Options WebSocket API. These tests cover all available options stream types, connection management, error handling, and advanced features.

## Features Tested

### Stream Types Covered

| Stream Type | Pattern | Description | Event Model |
|-------------|---------|-------------|-------------|
| **Index Price** | `{symbol}@index` | Index price for underlying assets | `IndexPriceEvent` |
| **Kline/Candlestick** | `{symbol}@kline_{interval}` | OHLCV data for options contracts | `KlineEvent` |
| **Mark Price** | `{underlyingAsset}@markPrice` | Mark prices for all options on underlying | `MarkPriceEvent` |
| **New Symbol Info** | `option_pair` | New option symbol listings | `NewSymbolInfoEvent` |
| **Open Interest** | `{underlyingAsset}@openInterest@{expirationDate}` | Open interest data | `OpenInterestEvent` |
| **Partial Depth** | `{symbol}@depth{levels}[@{speed}]` | Order book depth data | `PartialDepthEvent` |
| **Individual Ticker** | `{symbol}@ticker` | 24h statistics for specific options | `TickerEvent` |
| **Ticker by Underlying** | `{underlyingAsset}@ticker@{expirationDate}` | Aggregated ticker data | `TickerByUnderlyingEvent` |
| **Trade** | `{symbol}@trade` or `{underlyingAsset}@trade` | Real-time trade data | `TradeEvent` |

### Connection Features

- **Server Management**: Add, remove, update, and switch between servers
- **Connection Resilience**: Error handling and recovery mechanisms
- **Multiple Connections**: Concurrent stream handling
- **Connection Health**: Status monitoring and verification

### Advanced Features

- **Combined Streams**: Multi-stream event handling
- **Event Processing**: Type-safe event handlers for all stream types
- **Error Handling**: Comprehensive error detection and reporting
- **Performance Testing**: High-volume and concurrent stream testing

## Options Market Data Features

The tests verify proper handling of options-specific data including:

- **Greeks**: Delta, Theta, Gamma, Vega (in ticker streams)
- **Implied Volatility**: IV calculations and updates
- **Strike Prices**: Option contract specifications
- **Expiration Dates**: Contract expiry information
- **Option Types**: Call/Put option identification
- **Risk Metrics**: Mark prices and estimated exercise prices

## Test Configuration

### Environment Setup

1. Copy the environment template:
   ```bash
   cp env.example env.local
   ```

2. Configure test parameters in `env.local`:
   ```bash
   # Note: Most options streams are public and don't require authentication
   DEFAULT_SYMBOL=BTC-240329-50000-C
   DEFAULT_UNDERLYING=BTC
   DEFAULT_EXPIRATION=240329
   DEFAULT_INTERVAL=1m
   ```

### Authentication

Most options streams are **public** and don't require API credentials. Authentication is only needed for:
- Subscribe/Unsubscribe operations (if supported)
- Private user data streams (if any)

## Running Tests

### Quick Start

```bash
# Run all tests
go test -v

# Run with extended timeout for thorough testing
go test -v -timeout 20m
```

### Specific Test Categories

```bash
# Test connection functionality
go test -v -run TestConnection

# Test individual stream types
go test -v -run TestIndexPriceStream
go test -v -run TestKlineStream
go test -v -run TestMarkPriceStream
go test -v -run TestTickerStream

# Test advanced features
go test -v -run TestCombinedStreamEventHandler
go test -v -run TestConcurrentStreams

# Run the complete integration suite
go test -v -run TestFullIntegrationSuite
```

### Performance Testing

```bash
# Test high-volume streams
go test -v -run TestHighVolumeStreams

# Test concurrent connections
go test -v -run TestConcurrentStreams
```

## Test Structure

### Core Files

- `integration_test.go` - Main test client and utilities
- `main_test.go` - Test suite orchestration and summary
- `connection_test.go` - Connection and server management tests
- `streams_test.go` - Individual stream type tests
- `enhanced_features_test.go` - Advanced feature and performance tests

### Test Architecture

The tests use a sophisticated architecture featuring:

- **Shared Client Management**: Efficient resource utilization across tests
- **Dedicated Clients**: Isolated testing for specific scenarios
- **Event Recording**: Comprehensive event tracking and verification
- **Graceful Timeouts**: Handling low-activity periods on mainnet
- **Error Recovery**: Automatic reconnection and error handling

## Expected Behavior

### Market Activity Considerations

Options markets typically have lower activity than spot or futures markets, so:

- **Timeouts are Expected**: Many tests use graceful timeout handling
- **Event Frequency**: Some streams may have infrequent updates
- **Market Hours**: Activity varies by time of day and market conditions
- **Contract Popularity**: Some option contracts have very low trading volume

### Success Criteria

Tests are considered successful when they:

1. **Establish Connections**: Successfully connect to WebSocket endpoints
2. **Receive Events**: Get at least one event (if market is active)
3. **Handle Timeouts Gracefully**: Don't fail on expected low activity
4. **Verify Data Structure**: Ensure events match expected models
5. **Test Functionality**: Verify SDK methods work as intended

## Troubleshooting

### Common Issues

1. **Connection Timeouts**
   - Increase timeout values in test configuration
   - Check network connectivity to Binance

2. **No Events Received**
   - This is often expected due to low options trading activity
   - Tests use graceful timeout handling for this scenario

3. **Stream Not Available**
   - Some option symbols may not exist or be delisted
   - Update test symbols to currently active contracts

### Debug Mode

Enable verbose logging to see detailed event information:

```bash
go test -v -run TestFullIntegrationSuite 2>&1 | tee test.log
```

## SDK Status

✅ **SDK Fully Updated**: All event handler naming patterns have been updated!

### 🔄 **Recent SDK Updates:**
- **Event Handler Naming**: Updated from `OnXxxEvent` to `HandleXxxEvent` pattern
- **Integration Tests**: All tests updated to use new `HandleXxxEvent` method names
- **Backward Compatibility**: Old `OnXxxEvent` methods have been replaced

### Event Handler Pattern
```go
// Updated naming pattern for all event handlers
client.HandleCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
    // Process combined stream events
    return nil
})

client.HandleTradeEvent(func(event *models.TradeEvent) error {
    // Process trade events
    return nil
})
```

## Development Notes

### Adding New Tests

When adding tests for new features:

1. Follow the existing naming convention: `Test{FeatureName}`
2. Use the shared client pattern for efficiency
3. Include graceful timeout handling for low-activity scenarios
4. Add comprehensive event verification
5. Update the test suite in `main_test.go`

### Stream Path Format

Options streams use specific path formats:

```
# Individual option symbol
BTC-250328-50000-C@{streamType}

# Underlying asset aggregation  
{underlying}@{streamType}[@{expiration}]

# Examples
BTC-250328-50000-C@ticker          # Individual option ticker
ETH@markPrice                      # All ETH options mark prices
BTC@openInterest@250328           # Open interest for BTC options expiring 250328
```

## API Coverage

See `API_COVERAGE.md` for detailed information about:
- Tested vs untested stream types
- Coverage statistics
- Known limitations
- Future test expansion plans

## Related Documentation

- [Binance Options API Documentation](https://binance-docs.github.io/apidocs/voptions/en/)
- [SDK Documentation](../../../../../../binance-go/ws/options-streams/README.md)
- [WebSocket Streams Guide](https://binance-docs.github.io/apidocs/voptions/en/#websocket-streams)