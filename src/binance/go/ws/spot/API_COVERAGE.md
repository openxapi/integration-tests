# Binance Spot WebSocket API — Coverage Tracker

Refactor status: **in progress**. The regenerated SDK at `../../../../../../binance-go/ws/spot` now exposes a single `SpotChannel` with 49 request operations (plus `Connect` / `Disconnect`) whose callbacks accept `(ctx, resp, error)`, and six typed event handlers. The legacy test suite is being realigned to this surface; this document captures the exported interfaces and tracks which ones already have integration coverage under the new structure.

## Current Snapshot
- Channel exports: 51 request-style methods (49 RPC-style operations + lifecycle connect/disconnect), 6 handler registration helpers, 6 unregister helpers. The prior `HandleErrorMessage` hook was removed.
- Client/server helpers: 12 client methods, 10 server-manager methods, 6 auth mutators, 2 signing helpers.
- Tests aligned to new SDK: **40 / 49 request operations** (public market data, session lifecycle, user-stream control, core user-data reads, primary order management, SOR flows) and **6 / 6 event handlers** (account update, balance update, order update, list status, external lock updates, event stream termination). Remaining gaps: commission/myFilters, open order-list views, and all `orderList*` requests.
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
- [x] `Klines`
- [x] `UiKlines`
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
- [x] `OrderAmendments`
- [x] `OrderAmendKeepPriority`
- [x] `SorOrderTest` *(skips when no SOR symbols available)*
- [x] `SorOrderPlace` *(skips when no SOR symbols available)*
- [x] `OpenOrdersCancelAll`

### User Data (Auth: USER_DATA)
- [ ] `AccountCommission`
- [x] `AccountRateLimitsOrders`
- [x] `AccountStatus`
- [x] `AllOrderLists`
- [x] `AllOrders`
- [ ] `MyFilters`
- [x] `MyTrades`
- [x] `MyAllocations`
- [x] `MyPreventedMatches`
- [x] `OpenOrdersStatus`
- [ ] `OpenOrderListsStatus`

### Session & User Stream
- [x] `SessionLogon` (Auth: SIGNED / Ed25519)
- [x] `SessionStatus`
- [x] `SessionLogout`
- [x] `SessionSubscriptions`
- [x] `UserDataStreamSubscribe`
- [x] `UserDataStreamSubscribeSignature`
- [x] `UserDataStreamUnsubscribe`

## SpotChannel Event Handlers
- [x] `HandleAccountUpdateEvent` / `UnregisterAccountUpdateEvent`
- [x] `HandleBalanceUpdateEvent` / `UnregisterBalanceUpdateEvent`
- [x] `HandleOrderUpdateEvent` / `UnregisterOrderUpdateEvent`
- [x] `HandleListStatusEvent` / `UnregisterListStatusEvent`
- [x] `HandleExternalLockUpdateEvent` / `UnregisterExternalLockUpdateEvent`
- [x] `HandleEventStreamTerminatedEvent` / `UnregisterEventStreamTerminatedEvent`

## Utility Types (Models)
- [ ] `NewResponseRegistry`, `RegisterMessageType`, `ParseMessage`, `ParseDynamicMessage`, `ParseOneOfResult`, `RegisterAllEventTypes`, `ValidateMessage`, `DetectResponseType`
- [ ] `NewMessageIDInt64`, `NewMessageIDString`, `NewMessageIDNull`

## Coverage Plan
1. Build `spot_channel_test.go` with an integration suite that connects once, reuses throttled helper calls, and records responses/events (pattern copied from `../spot-streams/market_channel_test.go`).
2. Add shared helpers (signing, auth fixture loading, REST symbol discovery) in `integration_test.go` akin to `../spot-streams/integration_test.go`.
3. Incrementally fill remaining gaps: trading order-list management requests, `AccountCommission`, `MyFilters`, and read-only order-list views.

_Updated: 2026-11-17 — refreshed for regenerated SDK, response handlers now include error parameter, and external lock events verified._
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

**Last Updated**: November 17, 2026 — synced with regenerated SDK and new handler signatures
**Test Coverage**: 40/49 request operations, 6/6 event handlers (order-list workflows, commission stats, myFilters, and open order-list views pending)
**Latest Test Results**: `go test -v ./...` *(SOR subtests skip automatically when no supported symbols are advertised; network access required for live execution.)*
**Note**: `userDataStream.start` and `userDataStream.stop` remain deprecated in spot trading. Use `session.logon` with Ed25519 authentication and `userDataStream.subscribe` instead.
