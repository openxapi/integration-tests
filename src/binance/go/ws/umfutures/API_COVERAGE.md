# Binance USD-M Futures WebSocket API — Coverage Tracker

This document records how the integration suite exercises the regenerated SDK under
`github.com/openxapi/binance-go/ws/umfutures`.

## Snapshot (2027‑02‑07)
- **Channel requests:** 16/16 covered (public, account, trading, session) + 1 negative path
- **Session logon:** ✅ when Ed25519 credentials are supplied (skipped on testnet endpoints or when keys missing)
- **Event handlers:** not generated for this channel
- **Client/server helpers:** exercised indirectly by the test harness
- **Signing paths:** HMAC always; Ed25519 only when keys are present; RSA currently unused on testnet

## UmfuturesChannel Request Coverage

| Method | Description | Auth | Coverage | Notes |
| --- | --- | --- | --- | --- |
| `Connect` / `Disconnect` | Open/close websocket connection to `/` | None | ✅ | Harness reuses a single connection per credential bundle |
| `TickerPrice` | Symbol price ticker | None | ✅ | Response fields validated (symbol, price, timestamp) |
| `TickerBook` | Best bid/ask | None | ✅ | Bid/ask prices & quantities checked |
| `Depth` | Order book snapshot | None | ✅ | Bids/asks decoded and numeric fields parsed |
| `Depth` (invalid symbol) | Error handling for bad symbols | None | ✅ | `newErrorResponseHandler` asserts status/code/msg for `ErrorMessage` |
| `AccountBalance` | Account balances (v1) | USER_DATA | ✅ | HMAC signing, array entries validated |
| `AccountPosition` | Position snapshot (v1) | USER_DATA | ✅ | Handles empty/flat accounts gracefully |
| `AccountStatus` | Account + position summary | USER_DATA | ✅ | Asset and position arrays inspected |
| `V2AccountBalance` | Account balances (v2) | USER_DATA | ✅ | Same assertions as v1 |
| `V2AccountPosition` | Position snapshot (v2) | USER_DATA | ✅ | Mirrors v1 coverage |
| `V2AccountStatus` | Account status (v2) | USER_DATA | ✅ | Asset totals verified |
| `OrderPlace` | Place limit order | TRADE | ✅ | Uses REST helper to generate compliant price/qty |
| `OrderStatus` | Query order | TRADE | ✅ | Confirms ID & symbol match placed order |
| `OrderModify` | Amend existing order | TRADE | ✅ | Adjusts price while maintaining position |
| `OrderCancel` | Cancel order | TRADE | ✅ | Validates cancellation payload |
| `SessionStatus` | Session capability probe | None | ✅ | Confirms server time & user data flag |
| `SessionLogout` | Terminate session | None | ✅ | Run after status to ensure clean shutdown |
| `SessionLogon` | Authenticate session | SIGNED | ✅* | Requires Ed25519 credentials; subtest skips when credentials missing or server is testnet |

\* When Ed25519 keys are absent the subtest is skipped and the API is marked "pending".

## Client / Server Manager Surface

| Component | Methods | Status | Notes |
| --- | --- | --- | --- |
| `Client` | `NewClient`, `NewClientWithOptions`, `SetAuth`, `AddOrUpdateServer`, `SetActiveServer`, `RegisterHandlers`, `StopReadLoop`, `Wait` | ✅ (harness) | Harness exercises these indirectly when establishing channels |
| `ServerManager` | `AddServer`, `AddOrUpdateServer`, `RemoveServer`, `UpdateServer`, `SetActiveServer`, `GetActiveServer`, `ListServers` | ⚠️ Partial | `ensureDefaultServer` uses add/update + set active; removal paths not yet covered |
| `Auth` | `NewAuth`, `SetSecretKey`, `SetPrivateKeyPath`, `SetPassphrase`, `ContextWithValue` | ✅ | Covered via request signing helpers |
| `RequestSigner` | `EnsureInitialized`, `SignRequest` | ✅ | Used for HMAC (always) and Ed25519 (when available) |
| `GetAuthTypeFromMessageName`, `RequiresSignature` | — | ⚠️ Pending | Not relied upon by current tests |

## Models & Utilities

- `models.MessageID` helpers (`NewMessageIDInt64`, `String`, marshal/unmarshal) are exercised through every request/response exchange.
- Dynamic response helpers (`NewResponseRegistry`, `ParseDynamicMessage`, etc.) remain untested; the channel uses strongly typed models only.

## Follow-up Opportunities

1. Add explicit tests for `ServerManager.RemoveServer` / `UpdateServer` error paths.
2. Introduce dedicated unit tests for `GetAuthTypeFromMessageName` and `RequiresSignature`.
3. Consider validating failure scenarios (e.g., signature mismatch, invalid quantity) once reliable error fixtures are available.
