# Binance Portfolio Margin Pro WebSocket – Integration Test Coverage

This document tracks the current coverage of the Binance Portfolio Margin **Pro** WebSocket SDK exercised by the integration tests under `src/binance/go/ws/pmarginpro-streams`.

## Test Suites Added

| File | Purpose |
|------|---------|
| `client_management_test.go` | Covers `Client`, `Auth`, and `ServerManager` exported APIs (server lifecycle, auth context, handler registration safety, stop/wait idempotency). |
| `user_data_channel_test.go` | Implements `TestFullIntegrationSuite_UserData`, validating listen-key lifecycle, connection/keep-alive/disconnect flows, and the available user-data event handler. |

Supporting helpers live beside these files (`integration_test.go`, `env_helpers_test.go`, `rest_helpers_test.go`, `assert_helpers_test.go`, etc.), mirroring the layout shared with the portfolio margin module.

## Coverage Summary

| Area | Status | Notes |
|------|--------|-------|
| Client / Server management APIs | ✅ Covered in `client_management_test.go` (add/update/remove/list servers, active switching, URL resolution, new client variants, auth context, stop/wait). |
| Listen key lifecycle (REST) | ✅ `user_data_channel_test.go` creates, renews, and tears down listen keys when credentials are available (best-effort – failures log and skip as flaky). |
| WebSocket connect / disconnect | ✅ `TestFullIntegrationSuite_UserData` performs connect, double-disconnect (idempotency), and captures SDK log anomalies. |
| Event handler registration | ✅ `HandleRiskLevelChangeEvent` is registered/unregistered while connected. |
| Live event assertions | ⚠️ Best-effort: the suite records the first `riskLevelChange` payload received and performs field sanity checks (timestamp, numeric parsing). Timeouts are tolerated because production traffic may be sparse. |
| Error / recovery scenarios | ⏳ Not yet implemented. Future work should inject negative cases (invalid listen key, forced disconnect recovery, handler errors). |

## User Data Stream Event Matrix

| Event Type | Handler Method | Test | Status | Notes |
|------------|----------------|------|--------|-------|
| `riskLevelChange` | `HandleRiskLevelChangeEvent` | `user_data_channel_test.go :: RiskLevelChangeEvent` | ⚠️ | Validates event type, timestamp recency, and numeric fields when present. |

Status legend: ✅ covered deterministically, ⚠️ exercised but depends on live market activity, ⏳ pending.

## Open Follow-Ups

1. **Negative scenarios:** add tests for invalid/expired listen keys, handler error propagation, and forced disconnect recovery once hooks exist.
2. **Metrics for event arrival:** capture counts to detect prolonged inactivity and surface warnings in CI.
3. **REST verification:** optionally cross-check risk metrics against REST endpoints for additional confidence.
4. **Docs:** keep `README.md` and this file in sync when additional suites or channels are added.
