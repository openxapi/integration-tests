# Binance COIN‑M Futures WebSocket Streams — Integration Tests

Integration tests for the Binance COIN‑M Futures WebSocket Streams SDK.

## Overview

- SDK: `../../../../../../binance-go/ws/cmfutures-streams`
- Suites: MarketStreamChannel, CombinedMarketStreamChannel, UserDataStreamChannel
- Servers: testnet by default; Market/Combined force mainnet for richer traffic
- Auth: only required for user‑data suite (listen key and trading actions)

## Test Files

- `main_test.go` — Suite entry and summary
- `integration_test.go` — Client setup and shared helpers
- `server_selection_test.go` — Server defaults and overrides
- `assert_helpers_test.go`, `log_helpers_test.go`, `test_timing_helpers_test.go`, `rest_helpers_test.go` — utilities
- `market_channel_test.go` — MarketStreamChannel: requests + all event handlers
- `combined_channel_test.go` — CombinedMarketStreamChannel: requests + all event handlers + wrapper
- `user_data_channel_test.go` — UserDataStreamChannel: listen key, requests (skipped on testnet), event handlers

## Running

```bash
# From this directory
go test -v

# Run a specific suite
go test -v -run TestFullIntegrationSuite_Market
go test -v -run TestFullIntegrationSuite_Combined
go test -v -run TestFullIntegrationSuite_UserData

# Increase wait windows for low‑activity periods
EVENT_WAIT_SECS=40 go test -v -run TestFullIntegrationSuite_Market
```

## Configuration

- `BINANCE_CMFUTURES_WS_SERVER`: optional WS endpoint override
- `BINANCE_CMFUTURES_REST_SERVER`: optional REST endpoint override (defaults to testnet)
- `PREFERRED_SYMBOL`: optional comma‑separated list to bias symbol choice (e.g. `BTCUSD_PERP,ETHUSD_PERP`)
- `DEFAULT_INTERVAL`: kline interval (default `1m`)
- `EVENT_WAIT_SECS`: per‑test wait window for events (default `20`)
- `ENABLE_REST_VALIDATION`: when `1`, enable extra cross‑checks via REST

Auth for user‑data suite:

```bash
export BINANCE_API_KEY=...
export BINANCE_SECRET_KEY=...
```

## Notes on Testnet

- User‑data RPCs `userDataStream.start/ping/stop` are not supported on testnet. The suite detects testnet and skips these requests to avoid server‑initiated closes.
- User‑data events are best‑effort to induce:
  - `ORDER_TRADE_UPDATE`: placed via small limit/market order
  - `ACCOUNT_CONFIG_UPDATE`: induced via REST `CreateLeverageV1`
  - `ACCOUNT_UPDATE`: opportunistic (e.g., fills)
  - Rare events (margin call, strategy/grid updates, listenKeyExpired) are registered but not forced

## Coverage Snapshot

- MarketStreamChannel
  - Requests: Subscribe, Unsubscribe, ListSubscriptions, SetProperty, GetProperty
  - Events: aggTrade, markPrice, indexPrice, kline, continuousKline, indexKline, markPriceKline,
    ticker, miniTicker, bookTicker, partialDepth, diffDepth, liquidation, allMarkPrices, allMiniTickers,
    allTickers, allBookTickers, allLiquidations, contractInfo
- CombinedMarketStreamChannel
  - Requests: Subscribe, Unsubscribe, ListSubscriptions, SetProperty, GetProperty
  - Events: same as Market + Combined wrapper (`stream`/`data`)
- UserDataStreamChannel
  - Requests: Start, Ping, Stop (skipped on testnet; validated via error handler when exercised)
  - Events: error, listenKeyExpired, ACCOUNT_UPDATE, MARGIN_CALL, ORDER_TRADE_UPDATE,
    ACCOUNT_CONFIG_UPDATE, STRATEGY_UPDATE, GRID_UPDATE

See API_COVERAGE.md for a detailed tracker and next steps.

## Troubleshooting

- “unhandled message” logs: the suite captures these and fails; please report SDK decode/routing issues if seen.
- No events observed: raise `EVENT_WAIT_SECS`, switch to mainnet for public streams, or set `PREFERRED_SYMBOL`.
- User‑data actions fail on testnet: ensure keys are testnet‑enabled and that leverage/position mode changes are permitted for the account state.

## Next Steps

- Extend field‑level assertions where induction becomes reliable
- Add more cross‑validation to REST when `ENABLE_REST_VALIDATION=1`
