# Binance Portfolio Margin WebSocket – Integration Test Coverage

This document tracks the current coverage of the Binance Portfolio Margin WebSocket SDK exercised by the integration tests under `src/binance/go/ws/pmargin-streams`.

## Test Suites Added

| File | Purpose |
|------|---------|
| `client_management_test.go` | Covers `Client`, `Auth`, and `ServerManager` exported APIs (server lifecycle, auth context, handler registration safety, stop/wait idempotency). |
| `user_data_channel_test.go` | Implements `TestFullIntegrationSuite_UserData`, validating listen-key lifecycle, connection/keep-alive/disconnect flows, and registering every user-data event handler. |

Supporting helpers live in `integration_test.go`, `env_helpers_test.go`, `rest_helpers_test.go`, `assert_helpers_test.go`, and friends, mirroring the structure used by the options streams module.

## Coverage Summary

| Area | Status | Notes |
|------|--------|-------|
| Client / Server management APIs | ✅ Covered in `client_management_test.go` (add/update/remove/list servers, active switching, URL resolution, new client variants, auth context, stop/wait). |
| Listen key lifecycle (REST) | ✅ `user_data_channel_test.go` creates, renews, and tears down listen keys (best-effort – failures are logged but do not halt the suite). |
| WebSocket connect / disconnect | ✅ `TestFullIntegrationSuite_UserData` performs connect, double-disconnect (idempotency), and captures SDK log anomalies. |
| Event handler registration | ✅ All 11 user-data handlers are registered, validated, and unregistered while connected. |
| Live event assertions | ⚠️ Best-effort. Each handler records the first event received and performs field sanity checks (timestamps, numeric parsing). Timeouts are tolerated because production traffic may be sparse. |
| Error / recovery scenarios | ⏳ Not yet implemented. Future work should inject negative cases (invalid listen key, forced disconnects, handler errors). |

## User Data Stream Event Matrix

| Event Type | Handler Method | Test | Status | Notes |
|------------|----------------|------|--------|-------|
| `CONDITIONAL_ORDER_TRADE_UPDATE` | `HandleConditionalOrderTradeUpdateEvent` | `user_data_channel_test.go :: ConditionalOrderTradeUpdateEvent` | ⚠️ | Records payload, validates timestamp/quantity when present. |
| `openOrderLoss` | `HandleOpenOrderLossEvent` | `user_data_channel_test.go :: OpenOrderLossEvent` | ⚠️ | Checks event type, numeric fields if populated. |
| `outboundAccountPosition` | `HandleMarginAccountUpdateEvent` | `user_data_channel_test.go :: MarginAccountUpdateEvent` | ⚠️ | Iterates positions, validates numeric strings. |
| `liabilityChange` | `HandleLiabilityUpdateEvent` | `user_data_channel_test.go :: LiabilityUpdateEvent` | ⚠️ | Confirms asset + timestamp. |
| `executionReport` | `HandleMarginOrderUpdateEvent` | `user_data_channel_test.go :: MarginOrderUpdateEvent` | ⚠️ | Validates price/quantity fields when provided. |
| `ORDER_TRADE_UPDATE` | `HandleFuturesOrderUpdateEvent` | `user_data_channel_test.go :: FuturesOrderUpdateEvent` | ⚠️ | Confirms order data shape. |
| `ACCOUNT_UPDATE` | `HandleFuturesBalancePositionUpdateEvent` | `user_data_channel_test.go :: FuturesBalancePositionUpdateEvent` | ⚠️ | Parses balances/positions if available. |
| `ACCOUNT_CONFIG_UPDATE` | `HandleFuturesAccountConfigUpdateEvent` | `user_data_channel_test.go :: FuturesAccountConfigUpdateEvent` | ⚠️ | Validates event envelope, logs payload. |
| `riskLevelChange` | `HandleRiskLevelChangeEvent` | `user_data_channel_test.go :: RiskLevelChangeEvent` | ⚠️ | Ensures risk level string present. |
| `balanceUpdate` | `HandleMarginBalanceUpdateEvent` | `user_data_channel_test.go :: MarginBalanceUpdateEvent` | ⚠️ | Parses balance delta. |
| `listenKeyExpired` | `HandleUserDataStreamExpiredEvent` | `user_data_channel_test.go :: UserDataStreamExpiredEvent` | ⚠️ | Captures expiry notifications when emitted. |

Status legend: ✅ covered deterministically, ⚠️ exercised but depends on live market activity, ⏳ pending.

## Open Follow-Ups

1. **Negative scenarios:** add tests for invalid/expired listen keys, handler error propagation, and forced disconnect recovery.
2. **Metrics for event arrival:** capture counts to detect prolonged inactivity and surface warnings in CI.
3. **Optional REST validation:** add toggles to confirm WebSocket payloads against REST endpoints (mirroring the options suite pattern).
4. **Docs:** keep `README.md` and this file in sync when additional suites are added (e.g., error injection once SDK supports it).
