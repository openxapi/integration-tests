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

4. (Optional) Pre-fetch module dependencies if your environment has network access:
   ```bash
   go mod download
   ```

## Running Tests

### Run all tests
```bash
go test -v ./...
```

### Run the full channel integration suite only
```bash
go test -v -run TestFullIntegrationSuite_CmFutures ./...
```

- The suite discovers a tradable symbol via the REST client and executes account, trading, and session RPC flows.
- Ed25519 credentials are required to exercise `session.logon`; the subtest is skipped when keys are absent.

## Test Structure

- `integration_test.go` — shared harness for loading credentials, selecting the WS server, and wiring channel connections
- `cmfutures_channel_test.go` — end-to-end coverage of account, trading, and session RPCs (mirrors the USD-M futures harness)
- `rest_helpers_test.go` — REST helpers for exchange info, ticker discovery, and compliant limit order parameters
- `assert_helpers_test.go`, `log_helpers_test.go`, `signing_helpers_test.go`, `timing_helpers_test.go` — utility functions shared across tests

## API Coverage

See [API_COVERAGE.md](./API_COVERAGE.md) for detailed API coverage information.

## Important Notes

1. **Testnet Environment**: All tests target Binance Coin-M Futures testnet endpoints by default (override via `BINANCE_CMFUTURES_WS_SERVER` / `BINANCE_CMFUTURES_REST_SERVER`).
2. **Rate Limits**: The harness throttles websocket calls (`WS_THROTTLE_MS`) and uses REST metadata to keep orders within exchange constraints.
3. **CMFUTURES Specific**: The channel offers account, trading, and session RPCs only; no public market-data streams are exposed by the current spec.
4. **Credentials**: HMAC keys are required for account/trading tests; Ed25519 keys are optional and only needed for `session.logon`.

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
