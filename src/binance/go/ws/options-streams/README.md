# Binance Options WebSocket Streams Integration Tests

This directory contains integration tests for the Binance Options WebSocket Streams SDK. The tests exercise request/response flows and all event handlers for each Options WS channel.

## Overview

- Targets the SDK at `github.com/openxapi/binance-go/ws/options-streams` (WS) and `github.com/openxapi/binance-go/rest/options` (REST for symbol discovery/validation).
- Uses Binance mainnet WS/REST by default; no credentials required for public market data. User Data streams require API keys.
- Each channel has its own IntegrationTestSuite that validates request methods and event handler payloads.

## Streams Covered

| Stream | Pattern | Description | Event |
|-------|---------|-------------|-------|
| Index Price | `{pair}@index` | Underlying index price (e.g., `BTCUSDT@index`) | `IndexPriceEvent` |
| Kline | `{symbol}@kline_{interval}` | OHLCV for an option contract | `KlineEvent` |
| Mark Price | `{underlying}@markPrice` | Mark prices for all options on underlying | `MarkPriceEvent` |
| New Symbol Info | `option_pair` | New option listings stream | `NewSymbolInfoEvent` |
| Open Interest | `{underlying}@openInterest@{expiration}` | Open interest snapshots | `OpenInterestEvent` |
| Partial Depth | `{symbol}@depth{levels}[@{speed}]` | Level 2 orderbook | `PartialDepthEvent` |
| Ticker (by symbol) | `{symbol}@ticker` | 24h stats for one option | `TickerEvent` |
| Ticker (by underlying) | `{underlying}@ticker@{expiration}` | Aggregated by expiry | `TickerByUnderlyingEvent` |
| Trade | `{symbol}@trade` or `{underlying}@trade` | Trade stream | `TradeEvent` |

## Environment Setup

1. Copy and edit env:
   ```bash
   cp env.example env.local
   # Edit values as needed, then
   source env.local
   ```

2. Notable variables (see `env.example`):
   - `BINANCE_API_KEY`, `BINANCE_SECRET_KEY` – only required for User Data streams.
   - `BINANCE_OPTIONS_REST_SERVER` – override REST base (defaults to `https://eapi.binance.com`).
   - `ENABLE_REST_VALIDATION` – `1/true` enables REST cross-checks in assertions.
   - `PREFERRED_UNDERLYING` – comma list to bias underlying selection (e.g., `BTC,ETH`).
   - `DEFAULT_INTERVAL` – kline interval (e.g., `1m`).
   - `EVENT_WAIT_SECS` – base event wait (default 20s; some suites use longer waits).
   - Optional order test knobs (User Data): `ENABLE_USERDATA_ORDER_TESTS=1`, `TEST_ORDER_QTY`, `TEST_ORDER_PRICE`, `TEST_ORDER_SIDE`, `TEST_ORDER_TIF`.

## Running Tests

Quick start:
```bash
go test -v
```

Per‑channel suites:
```bash
go test -v -run TestFullIntegrationSuite_Market
go test -v -run TestFullIntegrationSuite_Combined
go test -v -run TestFullIntegrationSuite_UserData   # requires API keys
```

Helpful flags/env:
```bash
go test -v -timeout 20m
EVENT_WAIT_SECS=30 ENABLE_REST_VALIDATION=1 go test -v -run TestFullIntegrationSuite_Market
go test -short  # skips long-running integration suites
```

## Test Structure

Core files in this module:
- `main_test.go` – entry and printed summary, per‑suite runners.
- `integration_test.go` – shared/dedicated client helpers and utilities.
- `market_channel_test.go` – MarketStreamsChannel suite (requests + events).
- `combined_channel_test.go` – CombinedMarketStreamsChannel suite (requests + events).
- `user_data_channel_test.go` – UserDataStreamsChannel suite (events; needs keys).
- Helpers: `assert_helpers_test.go`, `log_helpers_test.go`, `rest_helpers_test.go`, `symbol_helper.go`, `test_timing_helpers_test.go`.

Architecture highlights:
- Shared and dedicated client helpers minimize reconnect churn.
- Stream path builders from the SDK are used to construct correct subscription strings.
- Each request method is validated via per‑request callbacks (id/fields).
- Event handlers are registered globally in each suite and validated field‑by‑field.
- Graceful timeouts are used due to lower Options market activity.

## Expected Behavior

Options markets can be quiet; timeouts with “acceptable” log notes are not failures. Tests succeed when connections establish, responses are ACKed where applicable, and at least one well‑formed event is observed (when market activity exists). Some endpoints (e.g., `SET_PROPERTY`/`GET_PROPERTY`) may time out on certain servers; tests treat this as informational when appropriate.

## SDK Status

- Event handler names follow `HandleXxxEvent` (updated from older `OnXxxEvent`).
- Tests also fail fast if the SDK logs any “unhandled message:” lines during a suite.

## Stream Path Examples

```
# Pair‑based index
BTCUSDT@index

# Kline on an option contract
BTC-250328-50000-C@kline_1m

# Underlying‑scoped streams
ETH@markPrice
BTC@openInterest@250328
```

## API Coverage

See `API_COVERAGE.md` for a checklist of covered exported methods and next steps.

## References

- Binance Options API: https://binance-docs.github.io/apidocs/voptions/en/
- SDK WS README: ../../../../../../binance-go/ws/options-streams/README.md
- WebSocket Streams Guide: https://binance-docs.github.io/apidocs/voptions/en/#websocket-streams

