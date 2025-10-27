# Binance Spot WebSocket Streams Integration Tests

This directory contains channel-focused integration tests for the Binance Spot WebSocket Streams SDK, refactored to follow the same structure and patterns as `../umfutures-streams`.

## Overview

- SDK: `../../../../../../binance-go/ws/spot-streams`
- Server: Uses Binance Spot Testnet by default (`wss://stream.testnet.binance.vision/`)
- Structure: One IntegrationTestSuite per channel (Market, Combined)

## Test Files

Support helpers
- `integration_test.go` – shared client manager and helpers (selects `testnet` by default)
- `assert_helpers_test.go`, `log_helpers_test.go`, `test_timing_helpers_test.go`, `rest_helpers_test.go`

Channel suites
- `market_channel_test.go` – MarketStreamChannel requests (Subscribe/List/Get/Set) and all event handlers
- `combined_channel_test.go` – CombinedMarketStreamChannel requests and all event handlers (incl. combined wrapper)

## Running Tests

```bash
# Run all tests
cd src/binance/go/ws/spot-streams
go test -v

# Run channel suites
go test -v -run TestFullIntegrationSuite_Market
go test -v -run TestFullIntegrationSuite_Combined
# Quick focus examples
go test -v -run TestFullIntegrationSuite_Market/TradeEvent
go test -v -run TestFullIntegrationSuite_Market/TickerEvent
go test -v -run TestFullIntegrationSuite_Combined/Request_Subscribe

# Short mode
go test -v -short
```

## Configuration

Most streams are public and don't require authentication.

```bash
cp env.example env.local
# Edit only if overriding defaults; API keys optional for public streams
source env.local
```

REST + Symbols
- REST is used to pick an active symbol (prefers `PREFERRED_SYMBOL`, else `BTCUSDT`).
- Override REST server via `BINANCE_SPOT_REST_SERVER` if needed.

## Coverage Notes

- MarketStreamChannel: Subscribe, Unsubscribe, ListSubscriptions, SetProperty, GetProperty, and all public event handlers are exercised.
- CombinedMarketStreamChannel: Request methods and both combined wrapper + unwrapped event handlers are covered.

Suites validate field presence and recency when events are observed; some events may not appear within timeouts depending on market activity.

## Server Selection

- Default WebSocket server: `testnet`
- Override WS server: set `BINANCE_SPOT_WS_SERVER` to a full WS URL; tests will create an `override` server and activate it.
