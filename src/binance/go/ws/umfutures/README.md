# Binance USDⓈ-M Futures WebSocket Integration Tests

This module exercises the regenerated Go SDK at `github.com/openxapi/binance-go/ws/umfutures`.  
The suite connects to the Binance USDⓈ-M Futures **testnet** WebSocket endpoint, drives every
exported channel method, and validates typed responses alongside signing flows.

## Test Layout

| File | Role |
| --- | --- |
| `main_test.go` | Prints harness summary and wiring for verbose runs |
| `integration_test.go` | Credential loader, client/channel harness, server helpers |
| `umfutures_channel_test.go` | End-to-end suite (`TestFullIntegrationSuite_UmFutures`) with subtests for public, user-data, trading, and session flows |
| `assert_helpers_test.go` | Shared numeric/string assertions |
| `signing_helpers_test.go` | Wrapper that mirrors SDK signing into request structs |
| `timing_helpers_test.go` | Throttling and event wait utilities |
| `log_helpers_test.go` | Debug logging of requests on test failure |
| `rest_helpers_test.go` | Lightweight REST helpers (symbol discovery, order parameter sizing) |

## Executing the Suite

```bash
# from src/binance/go/ws/umfutures
go test -v
```

`TestFullIntegrationSuite_UmFutures` opens one connection per credential bundle and fans out into subtests:

1. **PublicRequests** – `ticker.price`, `ticker.book`, `depth`
2. **UserDataRequests_HMAC** – `account.balance`, `account.position`, `account.status`, `v2` equivalents
3. **TradingRequests_HMAC** – `order.place`, `order.status`, `order.modify`, `order.cancel`
4. **SessionRequests** – `session.status`, `session.logout`, optional `session.logon` (Ed25519)

Each subtest reuses the same channel, applies conservative throttling, and asserts every field returned by the SDK models.

## Credentials & Environment

The harness reads the following environment variables (see `env.example`):

```
BINANCE_API_KEY            # HMAC API key (required for authenticated flows)
BINANCE_SECRET_KEY         # HMAC secret
BINANCE_RSA_API_KEY        # Optional RSA API key
BINANCE_RSA_PRIVATE_KEY_PATH
BINANCE_RSA_PRIVATE_KEY_PASSPHRASE
BINANCE_ED25519_API_KEY
BINANCE_ED25519_PRIVATE_KEY_PATH
BINANCE_ED25519_PRIVATE_KEY_PASSPHRASE

BINANCE_UMFUTURES_WS_SERVER   # Optional WS override (defaults to testnet)
BINANCE_UMFUTURES_REST_SERVER # Optional REST override for helper calls
BINANCE_UMFUTURES_SYMBOL      # Optional preferred symbol (comma separated)
WS_THROTTLE_MS                # Request pacing override (default 300 ms)
EVENT_WAIT_SECS               # Event wait window (default 20 s)
```

Only the public flow runs without credentials. Trading and user-data tests require HMAC keys; `session.logon` additionally requires an Ed25519 key pair.

## REST Assist

`rest_helpers_test.go` touches the REST SDK (`github.com/openxapi/binance-go/rest/umfutures`) to:

- Confirm the chosen symbol exists on testnet
- Fetch current ticker price
- Derive valid price/quantity pairs that satisfy filters before placing orders

All REST calls target the futures **testnet** (`https://testnet.binancefuture.com`) unless `BINANCE_UMFUTURES_REST_SERVER` overrides it.

## Safety Considerations

- The suite only interacts with Binance **testnet** endpoints; no real funds are at risk.
- WebSocket requests are throttled to stay below rate limits.
- Every request uses bounded contexts to prevent hangs.
- On failure the last request payload is logged to simplify triage.

## Known Limitations

- `session.logon` runs only when Ed25519 credentials are available; the suite skips gracefully otherwise.
- The AsyncAPI spec omits explicit `LOT_SIZE` metadata, so minimum quantity defaults to reasonable fallbacks (0.001 for BTC pairs).
- Event handlers are not generated for this channel; only request/response surfaces are covered.

Feel free to extend the suite with specialised subtests—`TestFullIntegrationSuite_UmFutures` is the single entry point for continuous validation.
