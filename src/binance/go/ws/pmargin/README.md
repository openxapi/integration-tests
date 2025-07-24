# Binance Portfolio Margin WebSocket Integration Tests

This directory contains comprehensive integration tests for the Binance Portfolio Margin WebSocket SDK.

## Overview

The Portfolio Margin WebSocket API provides real-time access to portfolio margin trading data and user account information through user data streams. Unlike market data streams, these require authentication via listen keys obtained from the REST API.

## Architecture

### SDK Structure
- **Client**: Portfolio Margin WebSocket client with connection management
- **Events**: 11 different event types for portfolio margin operations
- **Models**: Strongly-typed event models for all portfolio margin events
- **Authentication**: Listen key-based authentication via REST API

### Event Types Supported
1. **CONDITIONAL_ORDER_TRADE_UPDATE** - Conditional order trade execution updates
2. **openOrderLoss** - Open order loss notifications
3. **outboundAccountPosition** - Margin account position updates
4. **liabilityChange** - Liability balance changes
5. **executionReport** - Margin order execution reports
6. **ORDER_TRADE_UPDATE** - Futures order execution updates
7. **ACCOUNT_UPDATE** - Futures balance and position updates
8. **ACCOUNT_CONFIG_UPDATE** - Account configuration changes
9. **riskLevelChange** - Risk level change notifications
10. **balanceUpdate** - Balance update notifications
11. **listenKeyExpired** - Listen key expiration events

## Setup Instructions

### 1. Environment Configuration

Copy the environment template:
```bash
cp env.example env.local
```

Edit `env.local` with your API credentials:
```bash
# Required
BINANCE_API_KEY=your_api_key_here
BINANCE_SECRET_KEY=your_secret_key_here

# Required for user data stream tests
BINANCE_LISTEN_KEY=your_listen_key_here

# Optional
BINANCE_WS_SERVER_URL=wss://fstream.binance.com/pm/ws/{listenKey}
```

### 2. Obtain Listen Key

Portfolio Margin WebSocket requires a valid listen key from the REST API:

```bash
# Get listen key for Portfolio Margin
curl -X POST 'https://fapi.binance.com/fapi/v1/listenKey' \
     -H 'X-MBX-APIKEY: your_api_key_here'
```

Add the returned listen key to `env.local` as `BINANCE_LISTEN_KEY`.

### 3. Load Environment

```bash
source env.local
```

## Running Tests

### Run All Tests
```bash
go test -v -run TestFullIntegrationSuite ./...
```

### Run Specific Test Suites

**Connection Management Tests:**
```bash
go test -v -run TestConnectionSuite ./...
```

**Event Handling Tests:**
```bash
go test -v -run TestEventsSuite ./...
```

**User Data Stream Tests:**
```bash
go test -v -run TestUserDataSuite ./...
```

### Test Individual Components

**Connection tests only:**
```bash
go test -v ./... -run "TestConnection"
```

**Event handler tests only:**
```bash
go test -v ./... -run "TestEvent"
```

**User data stream tests only:**
```bash
go test -v ./... -run "TestUserData"
```

## Test Coverage

### Current Coverage: 0% (Implementation Ready)

See [API_COVERAGE.md](./API_COVERAGE.md) for detailed coverage information.

**Test Files:**
- `main_test.go` - Test infrastructure and setup
- `connection_test.go` - Connection management tests
- `events_test.go` - Event handling tests
- `userdata_test.go` - User data stream lifecycle tests
- `integration_test.go` - Comprehensive workflow tests

## Test Categories

### 1. Connection Management
- ✅ Client creation and configuration
- ✅ Server management and switching
- ✅ Connection with listen key resolution
- ✅ Connection state management
- ✅ Error handling and edge cases

### 2. Event Handling
- ✅ Event handler registration for all 11 event types
- ✅ Concurrent event processing
- ✅ Event parsing and validation
- ✅ Error handling in event processors

### 3. User Data Stream Operations
- ✅ Stream connection with listen key
- ✅ Stream subscription (auto-subscription)
- ✅ Stream keep-alive (ping)
- ✅ Stream lifecycle management
- ✅ Reconnection scenarios

### 4. Error Scenarios
- ✅ Invalid listen key handling
- ✅ Connection failures
- ✅ Authentication errors
- ✅ Network interruptions
- ✅ API error responses

## Important Notes

### Authentication Requirements
- **Listen Key Required**: All connections require a valid listen key from REST API
- **No Public Endpoints**: Portfolio Margin WebSocket only supports user data streams
- **Key Expiration**: Listen keys expire after 60 minutes and need renewal

### Event Dependencies
- **Trading Activity**: Events depend on actual portfolio margin trading activity
- **Account Type**: Some events require specific account configurations
- **Real-time Data**: Events are real-time and cannot be simulated in tests

### Connection Behavior
- **Auto-subscription**: Events are automatically received upon connection
- **Keep-alive Required**: Connections need periodic keep-alive pings
- **Single Stream**: Only one user data stream per listen key

## API Limitations

### Portfolio Margin Specific
- Only user data streams available (no market data streams)
- Requires portfolio margin account enabled
- Some events depend on specific trading configurations

### WebSocket Constraints  
- Listen key dependency for all connections
- No testnet environment available
- Real credentials required for all tests

## Troubleshooting

### Common Issues

**"Listen key required" error:**
```bash
# Obtain fresh listen key
curl -X POST 'https://fapi.binance.com/fapi/v1/listenKey' \
     -H 'X-MBX-APIKEY: your_api_key'
```

**Connection failures:**
- Check internet connectivity
- Verify API key permissions
- Ensure listen key is fresh (< 60 minutes old)

**No events received:**
- Normal if no portfolio margin trading activity
- Events are real-time and depend on account activity
- Some events require specific account configurations

**Tests skipped:**
- `BINANCE_LISTEN_KEY` not set - user data stream tests skipped
- `BINANCE_API_KEY`/`BINANCE_SECRET_KEY` not set - auth tests skipped

### Debug Mode

Enable verbose logging:
```bash
LOG_LEVEL=DEBUG go test -v ./...
```

## Security Considerations

- **Never commit** `env.local` or actual API credentials
- Use **minimal permissions** on API keys
- Consider **IP restrictions** on your API keys
- **Regularly rotate** API credentials
- **Monitor usage** for unauthorized access

## Integration with CI/CD

For automated testing in CI/CD:

```yaml
# Example GitHub Actions workflow
env:
  BINANCE_API_KEY: ${{ secrets.BINANCE_API_KEY }}
  BINANCE_SECRET_KEY: ${{ secrets.BINANCE_SECRET_KEY }}
  # Note: Listen key must be obtained dynamically via REST API
```

## Related Documentation

- [Portfolio Margin WebSocket API Documentation](https://binance-docs.github.io/apidocs/futures/en/#websocket-market-streams)
- [Listen Key Management](https://binance-docs.github.io/apidocs/futures/en/#start-user-data-stream-user_stream)
- [API Coverage](./API_COVERAGE.md)
- [Main Integration Tests README](../../../../../../README.md)