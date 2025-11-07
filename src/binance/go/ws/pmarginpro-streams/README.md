## Binance Portfolio Margin Pro WebSocket – Integration Tests

This module exercises the Binance Portfolio Margin **Pro** WebSocket SDK using live integration tests. The structure mirrors the neighbouring `pmargin-streams` harness so the exchange suites stay consistent.

### Directory Structure

| Path | Purpose |
|------|---------|
| `integration_test.go` | Shared test configuration and client factory helpers. |
| `main_test.go` | Custom `TestMain` for summary output and shared client cleanup. |
| `client_management_test.go` | Coverage for `Client`, `Auth`, and `ServerManager` exported APIs. |
| `user_data_channel_test.go` | `TestFullIntegrationSuite_UserData` validating listen-key lifecycle, connection flows, and user data event handlers. |
| `*_helpers_test.go` | Assertion, logging, timing, REST, event, and env helpers reused across suites. |
| `API_COVERAGE.md` | Running checklist that tracks which SDK entry points are already covered. |

### Prerequisites

- Go 1.21+ (module pins 1.24 to match upstream SDK).
- Binance API key/secret with **Portfolio Margin Pro** WebSocket permissions.
- Network access to Binance REST + WebSocket endpoints.

### Environment Setup

1. Copy the template and provide credentials:
   ```bash
   cp env.example env.local
   # edit env.local and set:
   #   BINANCE_API_KEY=...
   #   BINANCE_SECRET_KEY=...
   ```
2. (Optional) Overrides:
   - `BINANCE_PM_REST_SERVER` – alternate REST base URL (papi).
   - `BINANCE_REST_SERVER_URL` – fallback REST override.
   - `BINANCE_WS_SERVER_URL` – override WebSocket server template if needed.
3. Load the environment before running tests:
   ```bash
   source env.local
   ```

Provide `BINANCE_LISTEN_KEY` to reuse an existing listen key; otherwise the suite will create, renew, and close one via REST (best effort).

### Running the Tests

```bash
# Run everything
go test -v ./...

# Run the user-data integration suite only
go test -v -run TestFullIntegrationSuite_UserData

# Run client/server management checks
go test -v -run TestClient
```

The tests assume live mainnet data. Event handlers log timeouts (rather than fail) when the Portfolio Margin Pro stream is idle.

### What Gets Exercised

- `Client` server management: add/update/remove, active switching, URL resolution, `StopReadLoop`, `Wait`, `SetAuth`, and constructor variants.
- REST listen-key lifecycle: create → renew → delete (if credentials supplied; failures mark the suite flaky rather than fatal).
- User data channel:
  - `Connect` and double `Disconnect` to confirm idempotency.
  - `HandleRiskLevelChangeEvent` registration/unregistration and payload field sanity checks (timestamps, numeric parsing).
  - SDK log watcher fails the suite on “unhandled message” output.

See [API_COVERAGE.md](./API_COVERAGE.md) for the detailed checklist and remaining follow-ups (error injection, reconnection exercises, etc.).

### Known Gaps / Next Steps

- Add negative scenarios (invalid listen key, forced disconnect recovery) once the SDK exposes suitable hooks.
- Capture metrics on event arrival frequency to spot prolonged inactivity in CI.
- Optional REST cross-checks (e.g., confirm risk metrics against REST endpoints) can be added behind a feature flag.
