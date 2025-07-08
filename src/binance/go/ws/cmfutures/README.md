# Binance CMFUTURES WebSocket Integration Tests

This directory contains integration tests for the Binance CMFUTURES (Coin-M Futures) WebSocket SDK.

## Overview

These tests validate the functionality of the Binance CMFUTURES WebSocket API SDK by testing:
- Account information retrieval
- Order placement and management
- User data stream operations

## Prerequisites

1. **Binance Testnet Account**: Create an account at https://testnet.binancefuture.com/
2. **API Credentials**: Generate API keys with appropriate permissions
3. **Go 1.21+**: Required for running tests

## Setup

1. Copy the environment configuration:
   ```bash
   cp env.example env.local
   ```

2. Edit `env.local` with your testnet API credentials:
   ```bash
   export BINANCE_API_KEY="your-testnet-api-key"
   export BINANCE_SECRET_KEY="your-testnet-secret-key"
   ```

3. Source the environment variables:
   ```bash
   source env.local
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

## Running Tests

### Run all tests:
```bash
go test -v ./...
```

### Run specific test suites:
```bash
# Account tests only
go test -v -run TestAccount ./...

# Trading tests only
go test -v -run TestTrading ./...

# User data stream tests only
go test -v -run TestUserDataStream ./...
```

### Run full integration suite:
```bash
go test -v -run TestFullIntegrationSuite ./...
```

## Test Structure

- `main_test.go` - Test setup and initialization
- `account_test.go` - Account API tests (balance, position, status)
- `trading_test.go` - Trading API tests (place, modify, cancel, status)
- `userdata_test.go` - User data stream tests
- `integration_test.go` - Full integration test suite

## API Coverage

See [API_COVERAGE.md](./API_COVERAGE.md) for detailed API coverage information.

## Important Notes

1. **Testnet Environment**: All tests run against the Binance testnet
2. **Rate Limits**: Tests include delays to respect API rate limits
3. **CMFUTURES Specific**: These tests are for Coin-M Futures (contracts settled in crypto)
4. **No Public APIs**: Unlike SPOT/UMFUTURES, CMFUTURES WebSocket has no public data endpoints

## Troubleshooting

1. **Authentication Errors**: Ensure your API keys have the necessary permissions
2. **Connection Errors**: Check if testnet.binancefuture.com is accessible
3. **Symbol Errors**: CMFUTURES uses different symbols (e.g., BTCUSD_PERP instead of BTCUSDT)

## Test Symbols

CMFUTURES (Coin-M) uses USD-based perpetual contract symbols:
- `BTCUSD_PERP` - Bitcoin perpetual contract (settled in BTC)
- `ETHUSD_PERP` - Ethereum perpetual contract (settled in ETH)

Note: CMFUTURES symbols use USD not USDT as they are coin-margined contracts.

## Contributing

When adding new tests:
1. Follow the existing test patterns
2. Update API_COVERAGE.md with newly tested endpoints
3. Include proper error handling
4. Add appropriate delays for rate limiting