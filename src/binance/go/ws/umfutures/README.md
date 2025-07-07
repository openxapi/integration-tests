# Binance USDⓈ-M Futures WebSocket Integration Tests

This directory contains integration tests for the Binance USDⓈ-M Futures WebSocket API client generated from AsyncAPI specifications.

## Overview

These tests validate the generated Go WebSocket client for Binance USD-M Futures trading against the official Binance testnet. The tests cover:

- **Public Endpoints**: Market data that doesn't require authentication
- **User Data Endpoints**: Account information requiring USER_DATA permission
- **Trading Endpoints**: Order management requiring TRADE permission

## Test Structure

- `main_test.go` - Main test suite orchestration and configuration
- `integration_test.go` - Core test infrastructure and utilities
- `public_test.go` - Public endpoint tests (ticker, depth, etc.)
- `userdata_test.go` - User data endpoint tests (account, positions, status)
- `trading_test.go` - Trading endpoint tests (orders, user data streams)

## Available Endpoints

### Public Endpoints (No Authentication)
- `ticker.price` - Get symbol price ticker
- `ticker.book` - Get best bid/ask prices
- `depth` - Get order book depth

### User Data Endpoints (USER_DATA Authentication)
- `account.balance` - Get account balance
- `account.position` - Get position information
- `account.status` - Get account status

### Trading Endpoints (TRADE Authentication)
- `userDataStream.start` - Start user data stream
- `userDataStream.ping` - Keep user data stream alive
- `userDataStream.stop` - Stop user data stream
- `order.place` - Place new orders
- `order.status` - Check order status
- `order.cancel` - Cancel orders
- `order.modify` - Modify existing orders

## Authentication Types Supported

- **HMAC-SHA256**: Traditional API key + secret
- **RSA**: RSA private key signing
- **Ed25519**: Modern elliptic curve signing

## Setup Instructions

1. **Copy Environment File**:
   ```bash
   cp env.example env.local
   ```

2. **Configure Testnet Credentials**:
   Edit `env.local` with your Binance Futures testnet API credentials:
   - Get testnet keys from: https://testnet.binancefuture.com/
   
   ```bash
   # HMAC Authentication
   export BINANCE_API_KEY=your_testnet_hmac_api_key_here
   export BINANCE_SECRET_KEY=your_testnet_secret_key_here
   
   # RSA Authentication (optional)
   export BINANCE_RSA_API_KEY=your_testnet_rsa_api_key_here
   export BINANCE_RSA_PRIVATE_KEY_PATH=/path/to/testnet_rsa_private_key.pem
   
   # Ed25519 Authentication (optional)
   export BINANCE_ED25519_API_KEY=your_testnet_ed25519_api_key_here
   export BINANCE_ED25519_PRIVATE_KEY_PATH=/path/to/testnet_ed25519_private_key.pem
   ```

3. **Source Environment**:
   ```bash
   source env.local
   ```

## Running Tests

### All Tests
```bash
go test -v
```

### Public Endpoints Only
```bash
go test -v -run TestTickerPrice
go test -v -run TestBookTicker
go test -v -run TestDepth
```

### User Data Endpoints
```bash
go test -v -run TestAccountBalance
go test -v -run TestAccountPosition
go test -v -run TestAccountStatus
```

### Trading Endpoints
```bash
go test -v -run TestOrderPlace
go test -v -run TestUserDataStream
```

### Specific Authentication Types
```bash
# HMAC authentication tests
go test -v -run "HMAC"

# Ed25519 authentication tests  
go test -v -run "Ed25519"

# RSA authentication tests
go test -v -run "RSA"
```

### With Extended Timeout
```bash
go test -v -timeout 10m
```

## Test Configuration

The test suite automatically detects available authentication methods based on environment variables and runs appropriate test combinations:

- **Public-NoAuth**: Tests public endpoints (no credentials needed)
- **HMAC-UserData**: Tests USER_DATA endpoints with HMAC auth
- **HMAC-Trade**: Tests TRADE endpoints with HMAC auth
- **Ed25519-UserData**: Tests USER_DATA endpoints with Ed25519 auth
- **Ed25519-Trade**: Tests TRADE endpoints with Ed25519 auth
- **RSA-UserData**: Tests USER_DATA endpoints with RSA auth
- **RSA-Trade**: Tests TRADE endpoints with RSA auth

## Safety Features

- **Testnet Only**: All tests use Binance Futures testnet (no real money at risk)
- **Rate Limiting**: Built-in 2-second delays between requests to prevent IP banning
- **Timeout Protection**: All requests have timeouts to prevent hanging
- **Error Handling**: Comprehensive error reporting and handling

## Test Server

- **Endpoint**: `wss://testnet.binancefuture.com/ws-fapi/v1`
- **Environment**: Binance Futures Testnet
- **Risk**: No real money - safe for testing

## Common Issues

### Authentication Errors
- Verify testnet API keys are correct
- Ensure private key files have proper permissions: `chmod 600 /path/to/key.pem`
- Check that testnet keys are from https://testnet.binancefuture.com/

### Connection Issues
- Verify internet connectivity
- Check if testnet is accessible
- Ensure no firewall blocking WebSocket connections

### Rate Limiting
- Tests include built-in rate limiting
- If you see rate limit errors, increase delays between tests

## Debugging

Enable verbose logging to see detailed test execution:

```bash
go test -v -run TestOrderPlace 2>&1 | tee test.log
```

## Notes

- These tests are based on the current umfutures AsyncAPI specification
- The specification includes basic futures trading functionality
- Tests validate both request/response handling and authentication
- All operations are performed on testnet for safety