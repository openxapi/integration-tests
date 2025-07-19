# Binance USD-M Futures REST API Integration Tests

This directory contains comprehensive integration tests for the Binance USD-M Futures REST API SDK.

## Overview

The integration tests validate the functionality of the OpenXAPI-generated SDK against the actual Binance USD-M Futures testnet endpoints. These tests ensure that the SDK works correctly with real API responses and handles various edge cases.

## Test Coverage

Current test coverage: **22/103 endpoints (21.4%)**

- **Passing**: 21 endpoints (20.4%)
- **Skipped (API Issues)**: 1 endpoint (1.0%)
- **Failed**: 0 endpoints (0%)
- **Untested**: 81 endpoints (78.6%)

For detailed coverage information, see [API_COVERAGE.md](./API_COVERAGE.md).

## Test Structure

### Files

- `integration_test.go` - Main test runner and common utilities
- `main_test.go` - Test entry point and summary
- `public_test.go` - Public endpoint tests (no authentication required)
- `API_COVERAGE.md` - Detailed API coverage tracking
- `SDK_ISSUES_REPORT.md` - Known SDK issues and bugs
- `env.example` - Environment variable template

### Test Categories

1. **Public API Tests** - Market data, ticker information, order book data
2. **Account API Tests** - Account information, balances, positions (TODO)
3. **Trading API Tests** - Order management, position management (TODO)
4. **User Stream Tests** - WebSocket user data streams (TODO)
5. **BinanceLink API Tests** - Referral and affiliate management (TODO)

## Running Tests

### Prerequisites

1. Copy environment configuration:
   ```bash
   cp env.example env.local
   ```

2. For authenticated tests (future), add your API credentials to `env.local`:
   ```bash
   export BINANCE_API_KEY="your_api_key"
   export BINANCE_SECRET_KEY="your_secret_key"
   ```

### Running Tests

```bash
# Run all tests
go test -v

# Run specific test categories
go test -v -run TestPing
go test -v -run TestServerTime
go test -v -run TestExchangeInfo

# Run public API tests
go test -v -run TestPublic

# Run full integration suite
go test -v -run TestFullIntegrationSuite
```

## Test Results

### Working Endpoints ✅

- **GetPingV1** - Basic connectivity test
- **GetTimeV1** - Server time synchronization
- **GetExchangeInfoV1** - Exchange information (with field type handling)
- **GetDepthV1** - Order book data
- **GetTradesV1** - Recent trades data
- **GetHistoricalTradesV1** - Historical trades (with authentication)
- **GetAggTradesV1** - Aggregate trades data
- **GetKlinesV1** - Kline/candlestick data
- **GetContinuousKlinesV1** - Continuous contract klines
- **GetIndexPriceKlinesV1** - Index price klines
- **GetMarkPriceKlinesV1** - Mark price klines
- **GetPremiumIndexKlinesV1** - Premium index klines
- **GetTicker24hrV1** - 24-hour ticker statistics
- **GetTickerPriceV1** - Symbol price tickers
- **GetTickerBookTickerV1** - Order book ticker
- **GetOpenInterestV1** - Open interest data
- **GetPremiumIndexV1** - Mark price and premium index
- **GetFundingRateV1** - Funding rate history
- **GetIndexInfoV1** - Composite index information
- **GetConstituentsV1** - Index price constituents
- **GetAssetIndexV1** - Multi-assets mode asset index

### Skipped Endpoints ⚠️

1. **GetFundingInfoV1** - API endpoint not available on Binance (404 Not Found)

## Latest Test Run Results

**Date**: July 16, 2025  
**Duration**: 42.73s  
**Tests**: 22 total  
**Status**: ✅ All tests passing

### Issues Resolved

1. **Exchange Info Test** - Fixed handling of `deliveryDate` field type mismatch (int32 vs int64)
2. **Historical Trades Test** - Fixed authentication requirement (endpoint needs API key)
3. **Funding Info Test** - Proper handling of 404 response (endpoint not available on Binance API)

### Current Status

All public API endpoints are now working correctly with the test suite. The tests properly handle:
- Field type mismatches with appropriate error reporting
- Authentication requirements
- API endpoint availability issues

## SDK Issues

1. **deliveryDate Field Type**: The SDK defines `deliveryDate` as int32, but Binance returns int64 values that exceed int32 range. This is handled in tests but should be reported as an SDK issue.

For detailed issue tracking, see [SDK_ISSUES_REPORT.md](./SDK_ISSUES_REPORT.md).

## Development Notes

### Authentication

The tests support multiple authentication methods:
- **HMAC** (default) - API key + secret
- **RSA** - API key + RSA private key
- **Ed25519** - API key + Ed25519 private key

### Rate Limiting

Tests include built-in rate limiting with a 2-second minimum interval between requests to respect testnet limitations.

### Error Handling

Tests properly handle:
- API errors with structured error responses
- Network timeouts and connection issues
- SDK-specific parsing errors
- Authentication failures

## Contributing

When adding new tests:

1. Follow the existing test pattern in `public_test.go`
2. Update `API_COVERAGE.md` with new coverage
3. Document any new SDK issues in `SDK_ISSUES_REPORT.md`
4. Add appropriate error handling and rate limiting
5. Include both positive and negative test cases

## Future Work

1. Implement account information tests
2. Add trading operation tests
3. Create user data stream tests
4. Add BinanceLink API tests
5. Implement comprehensive error scenario testing
6. Add performance and load testing