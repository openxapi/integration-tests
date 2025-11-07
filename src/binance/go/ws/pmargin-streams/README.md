## Binance Portfolio Margin WebSocket – Integration Tests

This module exercises the Binance Portfolio Margin WebSocket SDK using live integration tests. The layout mirrors the options-streams test harness so that suites run consistently across exchanges.

### Directory Structure

| Path | Purpose |
|------|---------|
| `integration_test.go` | Shared test config and client factory helpers. |
| `main_test.go` | Custom `TestMain` (summary + shared client cleanup). |
| `client_management_test.go` | Unit/integration coverage for `Client`, `Auth`, and `ServerManager`. |
| `user_data_channel_test.go` | `TestFullIntegrationSuite_UserData` covering listen-key lifecycle, connection flows, and all 11 user-data event handlers. |
| `*_helpers_test.go` | Assertion, logging, timing, event, REST, and env helpers. |
| `API_COVERAGE.md` | Live mapping between SDK surface area and test coverage. |

Legacy suites (`connection_test.go`, `events_test.go`, `userdata_test.go`, etc.) have been replaced by the above structure.

### Prerequisites

- Go 1.21+ (module declares 1.24 to match upstream SDK).
- Binance API key/secret with **Portfolio Margin** WebSocket permissions.
- Network access to Binance REST + WebSocket endpoints.

### Environment Setup

1. Copy the template and provide credentials:
   ```bash
   cp env.example env.local
   # edit env.local and set:
   #   BINANCE_API_KEY=...
   #   BINANCE_SECRET_KEY=...
   ```
2. (Optional) Override endpoints:
   - `BINANCE_PM_REST_SERVER` – alternate REST base URL.
   - `BINANCE_WS_SERVER_URL` – override WebSocket server template if needed.
3. Load the environment before running tests:
   ```bash
   source env.local
   ```

The suite creates, renews, and closes listen keys automatically via REST; providing `BINANCE_LISTEN_KEY` is no longer required.

### Running the Tests

Run everything (recommended):
```bash
go test -v ./...
```

Run the user-data integration suite only:
```bash
go test -v -run TestFullIntegrationSuite_UserData
```

Run client/server management checks:
```bash
go test -v -run TestClient
```

The tests assume live mainnet data. Event handlers will log timeouts (rather than fail) when the portfolio margin stream is idle.

### What Gets Exercised

- `Client` server management: add/update/remove, active switching, URL resolution, `StopReadLoop`, `Wait`, `SetAuth`, and constructor variants.
- REST listen-key lifecycle: create → renew → delete (best effort, failures are logged and treated as flaky).
- User data channel:
  - `Connect` and double `Disconnect` (idempotency check).
  - All 11 event handlers are registered while connected; first payload is inspected for type/timestamp/number sanity.
  - SDK log watcher fails the suite on “unhandled message” output.

See [API_COVERAGE.md](./API_COVERAGE.md) for the detailed checklist and remaining follow-ups (error injection, reconnection exercises, etc.).

### Known Gaps / Next Steps

- Add negative tests (invalid/expired listen key, forced disconnect) when the SDK exposes suitable hooks.
- Capture metrics on event arrival frequency to detect noteworthy inactivity in CI runs.
- Optional REST cross-checks can be added behind a feature flag (e.g., compare WS balances against REST snapshots).
