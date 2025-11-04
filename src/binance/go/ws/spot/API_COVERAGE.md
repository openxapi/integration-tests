# Binance Spot WebSocket API — Coverage Tracker

Refactor status: **in progress**. The SDK at `../../../../../../binance-go/ws/spot` was regenerated and exposes a single `SpotChannel` with 50 request-style methods plus 9 typed event handlers. The legacy test suite no longer matches the SDK surface; this document captures the exported interfaces and tracks which ones already have integration coverage under the new structure.

## Current Snapshot
- Channel exports: 50 request methods, 9 handler registration helpers, 9 unregister helpers.
- Client/server helpers: 12 client methods, 10 server-manager methods, 6 auth mutators, 2 signing helpers.
- Tests aligned to new SDK: **36 request methods / helpers** (public market data, session lifecycle, user-stream control, core user-data reads, spot trading order flows, SOR requests with graceful skips when unsupported), **6 / 9 event handlers** (channel-level error routing, event stream termination, order update, account update, balance update, list status); kline response handlers and external lock updates still pending.
- Target suite layout: channel-focused tests mirroring `../spot-streams` (single `TestFullIntegrationSuite_Spot` with subtests per request + event).

## Client Entry Points
- [ ] `NewClient` / `NewClientWithOptions` / `NewClientWithAuth`
- [ ] `Client.SetAuth`
- [ ] `Client.AddServer`, `AddOrUpdateServer`, `UpdateServer`, `RemoveServer`
- [ ] `Client.SetActiveServer`, `GetActiveServer`, `GetServer`, `ListServers`, `GetCurrentURL`, `GetURL`
- [ ] `Client.RegisterHandlers`, `StopReadLoop`, `Wait`

## Server Manager API
- [ ] `NewServerManager`
- [ ] `ServerManager.AddServer`, `AddOrUpdateServer`, `UpdateServer`, `UpdateServerPathname`, `RemoveServer`
- [ ] `ServerManager.ResolveServerURL`
- [ ] `ServerManager.SetActiveServer`, `GetActiveServer`, `GetServer`, `ListServers`, `GetActiveServerURL`

## Authentication & Signing
- [ ] `NewAuth`
- [ ] `Auth.SetSecretKey`, `SetPrivateKey`, `SetPrivateKeyPath`, `SetPrivateKeyReader`, `SetPassphrase`, `ContextWithValue`
- [ ] `NewRequestSigner`, `RequestSigner.EnsureInitialized`, `RequestSigner.SignRequest`
- [ ] `GetAuthTypeFromMessageName`, `RequiresSignature`

## SpotChannel Lifecycle
- [ ] `NewSpotChannel`
- [ ] `SpotChannel.Connect`
- [ ] `SpotChannel.Disconnect`

## SpotChannel Request APIs

### Public Market Data (Auth: NONE unless noted)
- [x] `Ping`
- [x] `Time`
- [x] `ExchangeInfo`
- [x] `AvgPrice`
- [x] `Depth`
- [x] `Klines` (`SendKlines` uses `HandleKlinesResponse`)
- [x] `UiKlines` (`SendUiKlines` uses `HandleUiKlinesResponse`)
- [x] `Ticker`
- [x] `Ticker24hr`
- [x] `TickerPrice`
- [x] `TickerBook`
- [x] `TickerTradingDay`
- [x] `TradesAggregate`
- [x] `TradesHistorical`
- [x] `TradesRecent`

### Trading (Auth: TRADE)
- [x] `OrderTest`
- [x] `OrderPlace`
- [x] `OrderStatus`
- [x] `OrderCancel`
- [x] `OrderCancelReplace`
- [ ] `OrderListPlace`
- [ ] `OrderListPlaceOco`
- [ ] `OrderListPlaceOto`
- [ ] `OrderListPlaceOtoco`
- [ ] `OrderListStatus`
- [ ] `OrderListCancel`
- [ ] `OrderAmendments`
- [ ] `OrderAmendKeepPriority`
- [x] `SorOrderTest` *(skips when no SOR symbols available)*
- [x] `SorOrderPlace` *(skips when no SOR symbols available)*
- [x] `OpenOrdersCancelAll`

### User Data (Auth: USER_DATA)
- [ ] `AccountCommission`
- [x] `AccountRateLimitsOrders`
- [x] `AccountStatus`
- [x] `AllOrderLists`
- [x] `AllOrders`
- [x] `MyTrades`
- [x] `MyAllocations`
- [x] `MyPreventedMatches`
- [x] `OpenOrdersStatus`
- [ ] `OpenOrderListsStatus`
- [ ] `OrderAmendments` (also listed under Trading for TRADE)
- [ ] `TradesRecent` (overlaps with market data but requires signature when `apiKey` present)

### Session & User Stream
- [x] `SessionLogon` (Auth: SIGNED / Ed25519)
- [x] `SessionStatus`
- [x] `SessionLogout`
- [x] `SessionSubscriptions`
- [x] `UserDataStreamSubscribe`
- [x] `UserDataStreamSubscribeSignature`
- [x] `UserDataStreamUnsubscribe`

### Miscellaneous Helpers
- [x] `SpotChannel.SendKlines` (requires `HandleKlinesResponse`)
- [x] `SpotChannel.SendUiKlines` (requires `HandleUiKlinesResponse`)

## SpotChannel Event Handlers
- [x] `HandleErrorMessage` / `UnregisterErrorMessage`
- [ ] `HandleKlinesResponse` / `UnregisterKlinesResponse`
- [ ] `HandleUiKlinesResponse` / `UnregisterUiKlinesResponse`
- [x] `HandleAccountUpdateEvent` / `UnregisterAccountUpdateEvent`
- [x] `HandleBalanceUpdateEvent` / `UnregisterBalanceUpdateEvent`
- [x] `HandleOrderUpdateEvent` / `UnregisterOrderUpdateEvent`
- [x] `HandleListStatusEvent` / `UnregisterListStatusEvent`
- [ ] `HandleExternalLockUpdateEvent` / `UnregisterExternalLockUpdateEvent`
- [x] `HandleEventStreamTerminatedEvent` / `UnregisterEventStreamTerminatedEvent`

## Utility Types (Models)
- [ ] `NewResponseRegistry`, `RegisterMessageType`, `ParseMessage`, `ParseDynamicMessage`, `ParseOneOfResult`, `RegisterAllEventTypes`, `ValidateMessage`, `DetectResponseType`
- [ ] `NewMessageIDInt64`, `NewMessageIDString`, `NewMessageIDNull`

## Coverage Plan
1. Build `spot_channel_test.go` with an integration suite that connects once, reuses throttled helper calls, and records responses/events (pattern copied from `../spot-streams/market_channel_test.go`).
2. Add shared helpers (signing, auth fixture loading, REST symbol discovery) in `integration_test.go` akin to `../spot-streams/integration_test.go`.
3. Incrementally fill remaining gaps: trading order-list management, additional SOR permutations, and the remaining user-data event handlers (external lock) plus response handlers for klines.

_Updated: 2025-11-03 — public, session, user-stream, and core trading suites implemented; SOR flows now covered with graceful skips when unavailable; order-list management & additional user-data events still pending._
- Authentication credentials must be configured via environment variables
- Rate limiting is properly handled with appropriate delays
- Tests include comprehensive error scenario coverage
- Response validation ensures API contract compliance

## Maintenance

This coverage document should be updated when:
- New APIs are added to the Binance Spot WebSocket interface
- New test cases are implemented
- API endpoints are deprecated or modified
- Authentication methods are added or changed

**Last Updated**: November 3, 2025 — added SOR discovery flow and ListStatus event validation
**Test Coverage**: 36/50 request methods, 6/9 event handlers (excludes order-list management variants, kline responses, and external lock events)
**Latest Test Results**: `go test -v ./...` *(SOR subtests skip automatically when no supported symbols are advertised; network access required for live execution.)*
**Note**: `userDataStream.start` and `userDataStream.stop` remain deprecated in spot trading. Use `session.logon` with Ed25519 authentication and `userDataStream.subscribe` instead.
