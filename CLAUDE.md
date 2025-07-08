# OpenXAPI Integration Tests - Development Guide

This repository contains integration tests for OpenXAPI-generated SDKs to ensure they work correctly with each release.

## Project Overview

See @README.md for general project information.

## Directory Structure

The repository follows a strict hierarchical structure:

```
src/{exchange}/{language}/{protocol}/{module}/
```

Where:
- `{exchange}`: The cryptocurrency exchange (e.g., `binance`, `okx`, `bybit`)
- `{language}`: The programming language (e.g., `go`, `python`, `javascript`)
- `{protocol}`: The protocol type (`ws` for WebSocket, `rest` for REST API)
- `{module}`: The specific trading module (e.g., `spot`, `umfutures`, `cmfutures`)

### Examples:
- `@src/binance/go/ws/spot/` - Integration tests for Binance spot trading WebSocket SDK in Go
- `@src/binance/go/ws/umfutures/` - Integration tests for Binance USD-M futures WebSocket SDK in Go
- `@src/okx/python/rest/spot/` - Would contain integration tests for OKX spot trading REST API SDK in Python

## Important Guidelines

1. **Isolation**: Each folder is dedicated to a specific module and is completely separated from other folders. Do not modify files outside your dedicated integration folder.

2. **Adding New Tests**: When adding integration tests for a new exchange, language, protocol, or module:
   - Maintain the directory structure exactly
   - Create your folder following the pattern above
   - Only update files within your dedicated folder

3. **SDK Location**: The actual SDK code being tested is located relative to the integration test folder:
   - For `@src/binance/go/ws/spot`, the SDK is at `../binance-go/ws/spot`
   - For `@src/binance/go/ws/umfutures`, the SDK is at `../binance-go/ws/umfutures`
   - **Never modify SDK files** - only create and update integration test files

4. **Test Implementation**: When implementing tests:
   - Read and understand the SDK code structure first
   - Create comprehensive tests covering all SDK functionality
   - Use the exchange's testnet when available
   - Include proper error handling and rate limiting

## Best Practices

- Always use environment variables for API credentials
- Include clear documentation in your test folder's README
- Follow the testing patterns established in existing modules
- Ensure tests are idempotent and can be run repeatedly
- Add appropriate delays to respect rate limits

## Integration Test Development Workflow

### When Creating New Integration Tests

1. **Check Coverage**: Always check the `API_COVERAGE.md` file in the test directory to see what endpoints are already tested
2. **Update Coverage**: After implementing new tests, update the `API_COVERAGE.md` file with:
   - The endpoints you've tested
   - The test file where they're located
   - Update the overall coverage percentage
3. **Track Progress**: Maintain a clear record of tested vs untested endpoints

### REST API Integration Test Progress

#### Binance Go REST SDK (@src/binance/go/rest/spot/)
- **Current Coverage**: 410/470+ endpoints (87.2%)
- **Files Created**: 
  - ✅ public_test.go (15 endpoints)
  - ✅ account_test.go (5 endpoints)
  - ✅ trading_test.go (11 endpoints)
  - ✅ oco_trading_test.go (7 endpoints)
  - ✅ sor_trading_test.go (3 endpoints)
  - ✅ wallet_test.go (11 endpoints)
  - ✅ wallet_advanced_test.go (29 endpoints)
  - ✅ margin_trading_test.go (10 endpoints)
  - ✅ margin_advanced_test.go (52 endpoints)
  - ✅ subaccount_test.go (47 endpoints)
  - ✅ simple_earn_test.go (24 endpoints)
  - ✅ staking_test.go (24 endpoints)
  - ✅ algo_trading_test.go (11 endpoints)
  - ✅ convert_test.go (9 endpoints)
  - ✅ crypto_loan_test.go (16 endpoints)
  - ✅ vip_loan_test.go (12 endpoints)
  - ✅ mining_test.go (13 endpoints)
  - ✅ portfolio_margin_test.go (19 endpoints)
  - ✅ binance_link_test.go (46 endpoints)
  - ✅ giftcard_test.go (6 endpoints)
  - ✅ dual_investment_test.go (5 endpoints)
  - ✅ small_apis_test.go (10 endpoints)

- **Remaining Work** (~60 endpoints):
  - Some sub-account endpoints requiring special permissions
  - Managed sub-account operations (require investor account)
  - BLVT operations
  - Options trading endpoints
  - Some specialized broker/VIP endpoints that require specific account types

### Next Steps for Continuing Work

When resuming work on REST API integration tests:

1. **Read API_COVERAGE.md** at `@src/binance/go/rest/spot/API_COVERAGE.md` to check current status
2. **Choose an untested service** from the list above
3. **Create a new test file** following the naming pattern: `{service}_test.go`
4. **Implement comprehensive tests** for all endpoints in that service
5. **Update API_COVERAGE.md** with the newly tested endpoints
6. **Update integration_test.go** to include the new test functions in `initializeTests()`
7. **Update this section** in CLAUDE.md with the new coverage numbers

### Test Implementation Guidelines

- Each service should have its own test file
- Test both success and error cases
- Handle endpoints that might not be available on testnet gracefully (use `t.Skip()`)
- Include proper rate limiting between API calls
- Test all authentication methods when applicable (HMAC, RSA, Ed25519)
- Verify response structures thoroughly