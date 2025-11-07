# Binance Spot WebSocket Streams — Coverage

This document tracks integration test coverage for the Spot WebSocket SDK using channel‑focused suites (pattern aligned with `../umfutures-streams`).

- SDK: `../../../../../../binance-go/ws/spot-streams`
- Tests: `src/binance/go/ws/spot-streams`
- Default WS Server: `testnet`
- REST override: `BINANCE_SPOT_REST_SERVER`

## Channel Suites

- MarketStreamChannel — `market_channel_test.go`
  - Requests: Subscribe, Unsubscribe, ListSubscriptions, SetProperty, GetProperty
  - Events: Trade, AggregateTrade, Kline, MiniTicker, Ticker, BookTicker, AveragePrice, PartialDepth, DiffDepth, AllTickers, AllMiniTickers, AllRollingWindowTickers

- CombinedMarketStreamChannel — `combined_channel_test.go`
  - Requests: Subscribe, Unsubscribe, ListSubscriptions, SetProperty, GetProperty
  - Events: Combined wrapper + unwrapped handlers for all public market events (Trade, AggregateTrade, Kline, MiniTicker, Ticker, BookTicker, AveragePrice, PartialDepth, DiffDepth, AllTickers, AllMiniTickers, RollingWindowTicker, AllRollingWindowTickers)

## Request Coverage

- Subscribe: Market, Combined — covered; assert id matches and result=null
- Unsubscribe: Market, Combined — covered; assert id matches and result=null
- ListSubscriptions: Market, Combined — covered; assert id matches and result is array
- SetProperty: Market, Combined — covered (combined flag; timeouts accepted); assert id matches and result=null when present
- GetProperty: Market, Combined — covered; assert id matches and result parseable as bool (or object containing combined)

## Event Coverage (Market)

- Trade: validate type, time, symbol, price
- AggregateTrade: validate type, time, price
- Kline: validate type, time, interval
- MiniTicker: validate type, time
- Ticker: validate type, time
- BookTicker: validate best bid/ask presence
- AveragePrice: validate type, time
- PartialDepth: validate lastUpdateId, bids/asks presence
- DiffDepth: validate type, time
- AllTickers / AllMiniTickers / AllRollingWindowTickers: presence, basic shape

## Event Coverage (Combined)

- Combined wrapper: stream non-empty; data non-nil
- Trade: validate type, time, symbol, price/qty
- AggregateTrade: validate type, time, symbol, price/qty
- Kline: validate type, time, interval, open/close parseable
- MiniTicker: validate type, time, symbol, close price
- Ticker: validate type, time, symbol, last price
- BookTicker: validate update id, symbol, bid/ask presence and parseable
- AveragePrice: validate type, time, symbol, avg price and interval
- PartialDepth: validate lastUpdateId, bids/asks presence
- DiffDepth: validate type, time, symbol
- AllTickers: validate array item type, symbol, last price
- AllMiniTickers: validate array item type, symbol, close price
- RollingWindowTicker: validate time, symbol, last price (1h)
- AllRollingWindowTickers: validate array item symbol and last price (1h)
## Notes

- Timeouts on property requests and low‑activity symbols are treated as acceptable.
- Symbol selection uses REST; set `PREFERRED_SYMBOL` (comma‑separated) to force a symbol when active.
- Suites aim to exercise all exported channel methods and handlers; remaining gaps will be filled incrementally.
- User data stream coverage is not included in this module because Spot user data events are exposed via `ws/spot`, not `ws/spot-streams`.

Last updated: 2025‑10‑27 — removed UserDataStreamChannel suite (only available in ws/spot SDK) and refreshed coverage notes.
