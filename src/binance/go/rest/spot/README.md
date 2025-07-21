# Binance Spot REST API Integration Tests (Go)

This directory contains integration tests for the Binance Spot REST API SDK in Go.

## Setup

1. **Configure Environment Variables**
   ```bash
   cp env.example env.local
   # Edit env.local with your testnet API credentials
   source env.local
   ```

2. **Get Testnet Credentials**
   - Visit https://testnet.binance.vision/
   - Create an account and generate API keys
   - Fund your testnet account with test tokens

## Running Tests

### Run All Tests
```bash
go test -v ./...
```

### Run Full Integration Suite
```bash
go test -v -run TestFullIntegrationSuite ./...
```

### Run Specific Test Categories

**Public API Tests (No Auth Required):**
```bash
go test -v -run TestPublic ./...
go test -v -run TestExchangeInfo ./...
go test -v -run TestServerTime ./...
go test -v -run TestMarketDepth ./...
go test -v -run TestKlines ./...
```

**Account Tests (Auth Required):**
```bash
go test -v -run TestAccount ./...
go test -v -run TestAccountInfo ./...
go test -v -run TestAccountCommission ./...
```

**Trading Tests (Auth Required):**
```bash
go test -v -run TestTrading ./...
go test -v -run TestCreateOrder ./...
go test -v -run TestCancelOrder ./...
go test -v -run TestMyTrades ./...
go test -v -run TestOrderCancelReplace ./...
```

**OCO/OTO Trading Tests (Auth Required):**
```bash
go test -v -run TestCreateOrderOco ./...
go test -v -run TestCreateOrderListOco ./...
go test -v -run TestGetOrderList ./...
```

**SOR Trading Tests (Auth Required):**
```bash
go test -v -run TestCreateSorOrder ./...
go test -v -run TestGetMyAllocations ./...
```

**Wallet Tests (Auth Required):**
```bash
go test -v -run TestGetCapitalConfig ./...
go test -v -run TestGetDepositHistory ./...
go test -v -run TestGetWithdrawHistory ./...
```

**Margin Trading Tests (Auth Required):**
```bash
go test -v -run TestGetMarginAccount ./...
go test -v -run TestCreateMarginOrder ./...
go test -v -run TestGetMarginInterestHistory ./...
```

## Test Structure

### Core Files
- `main_test.go` - Test orchestration and summary
- `integration_test.go` - Core test infrastructure and utilities
- `API_COVERAGE.md` - Comprehensive API coverage tracking

### Test Categories
- `public_test.go` - Tests for public endpoints (market data, tickers, klines)
- `account_test.go` - Tests for account-related endpoints
- `trading_test.go` - Tests for basic trading operations
- `oco_trading_test.go` - Tests for OCO/OTO/OTOCO order types
- `sor_trading_test.go` - Tests for Smart Order Routing
- `wallet_test.go` - Tests for wallet operations
- `margin_trading_test.go` - Tests for margin trading operations

## Authentication Methods

The tests support three authentication methods:

1. **HMAC (Default)**
   - Set `BINANCE_API_KEY` and `BINANCE_SECRET_KEY`

2. **RSA**
   - Set `BINANCE_RSA_API_KEY` and `BINANCE_RSA_PRIVATE_KEY_PATH`

3. **Ed25519**
   - Set `BINANCE_ED25519_API_KEY` and `BINANCE_ED25519_PRIVATE_KEY_PATH`

## Rate Limiting

Tests include automatic rate limiting (2 seconds between requests) to avoid hitting testnet limits.

## Notes

- All tests use the Binance testnet by default for safety
- Some wallet endpoints may not be available on testnet
- Tests create real orders (on testnet) but cancel them immediately
- Ensure your testnet account has some USDT balance for trading tests