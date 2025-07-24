# Binance Options WebSocket API Integration Tests

This directory contains comprehensive integration tests for the Binance Options WebSocket API SDK.

## Overview

The Binance Options WebSocket API provides real-time access to options trading data and user account information through WebSocket connections. **Important: This API uses the Futures WebSocket API infrastructure (ws-fapi.binance.com) and only supports mainnet - there is no testnet environment available.**

## API Coverage

This test suite covers **100% of available APIs** (3/3):

### User Data Stream Management
- ✅ `userDataStream.start` - Start user data stream
- ✅ `userDataStream.ping` - Keep user data stream alive  
- ✅ `userDataStream.stop` - Stop user data stream

### Event Handling
- ✅ `AccountUpdate` events - Real-time account balance updates
- ✅ `OrderTradeUpdate` events - Order execution and status updates
- ✅ `RiskLevelChange` events - Risk level change notifications

## Key Features Tested

- **Authentication**: All operations require authentication (no public endpoints)
- **User Data Streams**: Complete lifecycle management (start/ping/stop)
- **Event Handlers**: Real-time event processing with concurrent handling
- **Error Handling**: Comprehensive error scenarios and API error codes
- **Multiple Streams**: Managing multiple concurrent user data streams
- **Connection Management**: WebSocket connection lifecycle

## Prerequisites

### 1. API Credentials
You need valid Binance API credentials with options trading permissions:
- API Key
- Secret Key

### 2. Environment Setup
```bash
# Copy environment template
cp env.example env.local

# Edit env.local with your credentials
# BINANCE_API_KEY=your_api_key_here
# BINANCE_SECRET_KEY=your_secret_key_here

# Load environment variables
source env.local
```

### 3. Go Dependencies
```bash
# Install dependencies
go mod download
```

## Running Tests

### Run All Tests
```bash
# Run the complete integration test suite
go test -v -run TestFullIntegrationSuite ./...
```

### Run Specific Test Suites
```bash
# User Data Stream tests only
go test -v -run TestUserDataStreamSuite ./...

# Event handling tests only  
go test -v -run TestEventsSuite ./...

# Individual test suites
go test -v -run TestUserDataStreamLifecycle ./...
go test -v -run TestEventHandlersRegistration ./...
```

### Verbose Testing
```bash
# Enable verbose logging
export TEST_VERBOSE=true
go test -v -run TestFullIntegrationSuite ./...
```

## Test Structure

### Files
- `main_test.go` - Common test infrastructure and utilities
- `userdata_test.go` - User data stream lifecycle tests
- `events_test.go` - Event handler registration and processing tests
- `integration_test.go` - Comprehensive end-to-end workflow tests

### Test Suites
1. **UserDataTestSuite** - Tests user data stream operations
2. **EventsTestSuite** - Tests event handling capabilities
3. **FullIntegrationTestSuite** - Comprehensive workflow testing

## Important Notes

### Mainnet Only
- **No testnet support** - All tests run against live mainnet environment
- Use minimal API key permissions for safety
- Consider read-only API keys where possible

### Rate Limits
- Tests include proper rate limiting delays (100ms between requests)
- Comprehensive rate limit error handling
- Respects Binance API rate limit guidelines

### Authentication
- All API operations require authentication
- Tests skip gracefully when credentials are not provided
- Supports both client-level and per-request authentication

### Event Testing
- Event handlers are tested for registration and error handling
- Real-time events depend on actual trading activity
- Tests wait for potential events but don't require them

## Sample Output

```
🚀 === Starting Binance Options WebSocket Integration Tests ===
🌐 Server: mainnet1 (ws-fapi.binance.com/ws-fapi/v1)
Note: Options WebSocket API only supports mainnet - no testnet available
=============================================================

📋 --- Running User Data Stream Tests ---
📡 User data stream started with listen key: abc123...
🏓 User data stream ping successful
🛑 User data stream stopped successfully

📋 --- Running Event Handling Tests ---
📊 Account Update Event handler registered
📈 Order Trade Update Event handler registered
⚠️ Risk Level Change Event handler registered

🎯 --- Running Comprehensive Integration Test ---
✅ All operations completed successfully

📊 === Test Summary ===
✅ All Options WebSocket integration tests completed successfully!
🎉 100% API coverage achieved (3/3 APIs tested)
📝 APIs tested: userDataStream.start, userDataStream.ping, userDataStream.stop
🎭 Event handlers tested: AccountUpdate, OrderTradeUpdate, RiskLevelChange
======================
```

## Security Considerations

1. **API Key Safety**
   - Never commit API keys to version control
   - Use environment variables only
   - Consider IP restrictions on API keys

2. **Minimal Permissions**
   - Use API keys with only required permissions
   - Avoid spot trading permissions if only testing streams

3. **Mainnet Awareness**
   - All operations occur on live environment
   - No sandbox/testnet protection available

## Troubleshooting

### Authentication Errors
```
Error: authentication required for USER_STREAM request
Solution: Ensure BINANCE_API_KEY and BINANCE_SECRET_KEY are set
```

### Connection Errors
```
Error: failed to connect to WebSocket
Solution: Check network connectivity and API key permissions
```

### Rate Limit Errors
```
Error: Status=429, Code=-1003, Message=Too many requests
Solution: Reduce test frequency or wait before retrying
```

## API Reference

For detailed API documentation, refer to:
- [Binance Options API Documentation](https://binance-docs.github.io/apidocs/voptions/en/)
- [WebSocket API Specification](https://binance-docs.github.io/apidocs/voptions/en/#websocket-api)

## Contributing

When adding new tests:
1. Follow the existing test patterns
2. Include proper error handling
3. Add rate limiting between requests
4. Update this README with new test coverage
5. Ensure tests work with minimal API permissions