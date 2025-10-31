# Binance Spot WebSocket API — Coverage Tracker

Refactor status: **in progress**. The SDK at `../../../../../../binance-go/ws/spot` was regenerated and exposes a single `SpotChannel` with 50 request-style methods plus 9 typed event handlers. The legacy test suite no longer matches the SDK surface; this document captures the exported interfaces and tracks which ones already have integration coverage under the new structure.

## Current Snapshot
- Channel exports: 50 request methods, 9 handler registration helpers, 9 unregister helpers.
- Client/server helpers: 12 client methods, 10 server-manager methods, 6 auth mutators, 2 signing helpers.
- Tests aligned to new SDK: **24 request methods / helpers** (public market data, session lifecycle, user-stream control, core user-data reads), **2 / 9 event handlers** (kline / UI kline responses); trading flows and user-data event emissions still pending.
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
- [ ] `OrderTest`
- [ ] `OrderPlace`
- [ ] `OrderStatus`
- [ ] `OrderCancel`
- [ ] `OrderCancelReplace`
- [ ] `OrderListPlace`
- [ ] `OrderListPlaceOco`
- [ ] `OrderListPlaceOto`
- [ ] `OrderListPlaceOtoco`
- [ ] `OrderListStatus`
- [ ] `OrderListCancel`
- [ ] `OrderAmendments`
- [ ] `OrderAmendKeepPriority`
- [ ] `SorOrderTest`
- [ ] `SorOrderPlace`
- [ ] `OpenOrdersCancelAll`

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
- [ ] `HandleErrorMessage` / `UnregisterErrorMessage`
- [ ] `HandleKlinesResponse` / `UnregisterKlinesResponse`
- [ ] `HandleUiKlinesResponse` / `UnregisterUiKlinesResponse`
- [ ] `HandleOutboundAccountPositionEvent` / `UnregisterOutboundAccountPositionEvent`
- [ ] `HandleBalanceUpdateEvent` / `UnregisterBalanceUpdateEvent`
- [ ] `HandleExecutionReportEvent` / `UnregisterExecutionReportEvent`
- [ ] `HandleListStatusEvent` / `UnregisterListStatusEvent`
- [ ] `HandleExternalLockUpdateEvent` / `UnregisterExternalLockUpdateEvent`
- [ ] `HandleEventStreamTerminatedEvent` / `UnregisterEventStreamTerminatedEvent`

## Utility Types (Models)
- [ ] `NewResponseRegistry`, `RegisterMessageType`, `ParseMessage`, `ParseDynamicMessage`, `ParseOneOfResult`, `RegisterAllEventTypes`, `ValidateMessage`, `DetectResponseType`
- [ ] `NewMessageIDInt64`, `NewMessageIDString`, `NewMessageIDNull`

## Coverage Plan
1. Build `spot_channel_test.go` with an integration suite that connects once, reuses throttled helper calls, and records responses/events (pattern copied from `../spot-streams/market_channel_test.go`).
2. Add shared helpers (signing, auth fixture loading, REST symbol discovery) in `integration_test.go` akin to `../spot-streams/integration_test.go`.
3. Incrementally fill remaining gaps: trading (order placement/cancel flows), order-list management, SOR operations, and user-data event handlers (execution report, balance/account updates, list status, stream termination).

_Updated: 2025-10-27 — public, session, and user-stream suites implemented; trading + advanced user-data/event coverage pending._
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

**Last Updated**: September 2025
**Test Coverage**: 100% (39/39 endpoints, excluding 2 deprecated)
**Latest Test Results**: All tests passed (100% success rate)
**Note**: `userDataStream.start` and `userDataStream.stop` are deprecated in spot trading. Use `session.logon` with Ed25519 authentication and `userDataStream.subscribe` instead.
