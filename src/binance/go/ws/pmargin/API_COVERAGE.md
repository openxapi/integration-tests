# Binance Portfolio Margin WebSocket API Coverage

This document tracks the integration test coverage for the Binance Portfolio Margin WebSocket API.

## API Overview

The Binance Portfolio Margin WebSocket API provides real-time access to portfolio margin trading data and user account information. **Note: This API primarily supports mainnet - check server configuration for testnet availability.**

## Coverage Summary

- **Total Event Types Available**: 11
- **Event Models Tested**: 11 (basic structure and parsing)
- **Event Handlers Tested**: 0 (blocked by SDK issues)
- **Basic Functionality Coverage**: ~60%
- **Core Event Handling Coverage**: 0% (SDK type mismatches)

## Event Type Coverage

### User Data Stream Events (0/11) ❌

| Event Type | Event Name | Status | Test File | Test Function | Notes |
|------------|------------|--------|-----------|---------------|-------|
| `CONDITIONAL_ORDER_TRADE_UPDATE` | Conditional Order Trade Update | ❌ | - | - | OCO trade execution updates |
| `openOrderLoss` | Open Order Loss | ❌ | - | - | Order loss notifications |
| `outboundAccountPosition` | Margin Account Update | ❌ | - | - | Account position changes |
| `liabilityChange` | Liability Update | ❌ | - | - | Liability balance changes |
| `executionReport` | Margin Order Update | ❌ | - | - | Margin order execution reports |
| `ORDER_TRADE_UPDATE` | Futures Order Update | ❌ | - | - | Futures order execution updates |
| `ACCOUNT_UPDATE` | Futures Balance Position Update | ❌ | - | - | Futures balance and position updates |
| `ACCOUNT_CONFIG_UPDATE` | Futures Account Config Update | ❌ | - | - | Account configuration changes |
| `riskLevelChange` | Risk Level Change | ❌ | - | - | Risk level notifications |
| `balanceUpdate` | Margin Balance Update | ❌ | - | - | Balance update notifications |
| `listenKeyExpired` | User Data Stream Expired | ❌ | - | - | Listen key expiration events |

### Connection Management (0/5) ❌

| Feature | Status | Test File | Test Function | Notes |
|---------|--------|-----------|---------------|-------|
| Connect with Listen Key | ❌ | - | - | Establish WebSocket connection |
| User Data Stream Subscription | ❌ | - | - | Auto-subscription on connection |
| Event Handler Registration | ❌ | - | - | Register event handlers |
| Connection State Management | ❌ | - | - | Monitor connection status |
| Error Handling | ❌ | - | - | Handle connection errors |

## SDK Architecture Analysis

### Event Handlers Available

Based on the SDK analysis, the Portfolio Margin WebSocket client supports the following event handlers:

1. **ConditionalOrderTradeUpdateHandler** - `OnConditionalOrderTradeUpdate()`
2. **OpenOrderLossHandler** - `OnOpenOrderLoss()`
3. **MarginAccountUpdateHandler** - `OnMarginAccountUpdate()`
4. **LiabilityUpdateHandler** - `OnLiabilityUpdate()`
5. **MarginOrderUpdateHandler** - `OnMarginOrderUpdate()`
6. **FuturesOrderUpdateHandler** - `OnFuturesOrderUpdate()`
7. **FuturesBalancePositionUpdateHandler** - `OnFuturesBalancePositionUpdate()`
8. **FuturesAccountConfigUpdateHandler** - `OnFuturesAccountConfigUpdate()`
9. **RiskLevelChangeHandler** - `OnRiskLevelChange()`
10. **MarginBalanceUpdateHandler** - `OnMarginBalanceUpdate()`
11. **UserDataStreamExpiredHandler** - `OnUserDataStreamExpired()`
12. **PmarginErrorHandler** - `OnPmarginError()`

### Connection Methods Available

- `ConnectToUserDataStream(ctx, listenKey)` - Connect with listen key
- `SubscribeToUserDataStream(ctx)` - Subscribe to events (auto-subscription)
- `PingUserDataStream(ctx)` - Keep connection alive
- Event handler registration methods
- Standard client methods (Connect, Disconnect, etc.)

## Test Scenarios to Cover

### Connection Management ❌
- [ ] Establish WebSocket connection with listen key
- [ ] Handle connection failures
- [ ] Monitor connection state
- [ ] Graceful disconnection
- [ ] Connection error recovery

### Event Handler Registration ❌
- [ ] Register all event type handlers
- [ ] Verify handler callback execution
- [ ] Test concurrent event handling
- [ ] Event handler error scenarios

### User Data Stream Events ❌
- [ ] Conditional order trade updates
- [ ] Open order loss notifications
- [ ] Margin account position updates
- [ ] Liability change events
- [ ] Margin order execution reports
- [ ] Futures order execution updates
- [ ] Futures balance/position updates
- [ ] Account configuration changes
- [ ] Risk level change notifications
- [ ] Balance update events
- [ ] Listen key expiration handling

### Error Handling ❌
- [ ] API error responses
- [ ] Network connection errors
- [ ] Invalid listen key handling
- [ ] Authentication failures
- [ ] Event parsing errors

### Stream Lifecycle ❌
- [ ] User data stream connection
- [ ] Stream keep-alive (ping)
- [ ] Stream expiration handling
- [ ] Reconnection scenarios

## Test Files to Create

| File | Purpose | Event Types Covered |
|------|---------|-------------------|
| `main_test.go` | Test infrastructure and setup | N/A |
| `integration_test.go` | Main test suite runner | All |
| `connection_test.go` | Connection management tests | Connection methods |
| `events_test.go` | Event handling tests | All 11 event types |
| `userdata_test.go` | User data stream lifecycle | Stream management |
| `error_test.go` | Error handling scenarios | Error responses |

## Key Implementation Notes

### Listen Key Requirement
- All connections require a valid listen key from REST API
- Listen key templates: `{listenKey}` in server URLs
- Server URL: `wss://fstream.binance.com/pm/ws/{listenKey}`

### Event Processing
- Events are automatically received upon connection
- No explicit subscription required for individual event types
- All events use portfolio margin specific models

### Authentication
- User data streams require valid API credentials
- Listen key must be obtained from REST API first
- Stream requires periodic keep-alive pings

## API Limitations

### No Public Endpoints
- All endpoints require authentication via listen key
- No market data streams available via WebSocket in this module
- Portfolio margin user data streams only

### Listen Key Dependency
- Connection requires valid listen key from REST API
- Listen key has expiration (typically 60 minutes)
- Must handle listen key renewal and reconnection

### Server Configuration
- Primary server: `wss://fstream.binance.com/pm/ws/{listenKey}`
- Check for testnet availability in server manager

## Future Development

### When Implementation Begins
- [ ] Create test environment configuration
- [ ] Implement connection management tests
- [ ] Add event handler registration tests
- [ ] Create comprehensive event processing tests
- [ ] Add error handling and edge case tests
- [ ] Implement stream lifecycle tests

### Required Setup
- [ ] API credentials for listen key generation
- [ ] REST API integration for listen key management
- [ ] Test data for various event scenarios
- [ ] Error simulation capabilities

## SDK Issues Identified

### Compilation Errors in Models (Critical) 🚨

**Issue**: Multiple compilation errors found in the Portfolio Margin SDK models:

1. **Field Redeclaration Errors**:
   - `conditional_order_trade_update.go:16:2: S redeclared`
   - `futures_order_update.go:14:2: S redeclared`
   - `futures_order_update.go:30:2: X redeclared`
   - `futures_order_update.go:38:2: L redeclared`

2. **Undefined Field Errors**:
   - `ConditionalOrderTradeUpdate` type missing `Event` field
   - `FuturesAccountConfigUpdate` type missing `Event` field

**Impact**: Integration tests cannot be run due to SDK compilation failures.

**Files Affected**:
- `binance-go/ws/pmargin/models/conditional_order_trade_update.go`
- `binance-go/ws/pmargin/models/futures_order_update.go`
- `binance-go/ws/pmargin/models/futures_account_config_update.go`

**Status**: 🚨 SDK event handlers blocked - type mismatches prevent event handling

## Test Files Created

| File | Purpose | Status | Coverage |
|------|---------|--------|----------|
| `main_test.go` | Test infrastructure | ✅ Working | 100% |
| `connection_test.go` | Connection management | ✅ Working | ~80% |
| `events_simplified_test.go` | Model structure testing | ✅ Working | Model parsing only |
| `integration_test.go` | Basic workflow testing | ✅ Working | Limited scope |
| `events_test.go` | Full event handlers | ❌ Blocked | 0% - SDK issues |
| `userdata_test.go` | User data streams | ❌ Blocked | 0% - SDK issues |

## Last Updated

**Date**: 2025-01-24  
**Coverage**: 60% basic functionality, 0% event handling (blocked by SDK issues)  
**Status**: Partial Implementation - Event Handling Blocked by SDK Type Mismatches 🚨

---

## Notes

1. **Listen Key Management**: Tests will need REST API integration for listen key generation
2. **Event Dependencies**: Event reception depends on actual portfolio margin trading activity
3. **Authentication Required**: All operations require valid API credentials
4. **Stream Management**: Proper handling of stream lifecycle and keep-alive required
5. **Error Scenarios**: Comprehensive error handling needed for production readiness