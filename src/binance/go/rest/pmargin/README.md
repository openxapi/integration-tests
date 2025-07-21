# Binance Portfolio Margin REST API Integration Tests

This directory contains comprehensive integration tests for the Binance Portfolio Margin REST API SDK.

## Overview

Portfolio Margin is a unified margin account that allows trading across multiple markets (Spot, USD-M Futures, Coin-M Futures) with shared margin and cross-collateral benefits.

- **SDK Package**: `github.com/openxapi/binance-go/rest/pmargin`
- **API Base URL**: `https://papi.binance.com`
- **Total APIs**: 98 functions
- **Authentication**: HMAC-SHA256, RSA, ED25519

## Important Requirements

### Account Requirements
- **Portfolio Margin Account**: Must have a Portfolio Margin account enabled
- **Special Permissions**: Many endpoints require specific account permissions
- **Testnet Limitations**: Portfolio Margin may not be fully supported on testnet

### API Key Requirements
- **API Key**: Required for all authenticated endpoints
- **Secret Key**: Required for HMAC authentication
- **Permissions**: API key must have Portfolio Margin permissions enabled

## Setup Instructions

### 1. Environment Configuration
```bash
# Copy the example environment file
cp env.example env.local

# Edit the configuration file
nano env.local

# Set your API credentials
export BINANCE_API_KEY="your_api_key"
export BINANCE_SECRET_KEY="your_secret_key"

# Or use Ed25519 authentication (recommended)
export BINANCE_ED25519_API_KEY="your_ed25519_api_key"
export BINANCE_ED25519_PRIVATE_KEY_PATH="/path/to/your/ed25519_private_key.pem"

# Load the environment
source env.local
```

### 2. Enable Test Features
Most destructive operations are disabled by default. Enable them carefully:

```bash
# Enable basic read-only tests
export BINANCE_TEST_PMARGIN_PING="true"
export BINANCE_TEST_PMARGIN_ACCOUNT="true"
export BINANCE_TEST_PMARGIN_LISTEN_KEY="true"

# Enable trading operations (use with caution)
export BINANCE_TEST_PMARGIN_MARGIN_ORDERS="true"
export BINANCE_TEST_PMARGIN_UM_ORDERS="true"
export BINANCE_TEST_PMARGIN_CM_ORDERS="true"
```

### 3. Portfolio Margin Account Setup
```bash
# Set this to true if you have a Portfolio Margin account
export BINANCE_PMARGIN_ACCOUNT_ENABLED="true"

# Set this to true if testnet supports Portfolio Margin (rarely)
export BINANCE_PMARGIN_TESTNET_SUPPORTED="false"
```

## Running Tests

### Install Dependencies
```bash
go mod tidy
```

### Run All Tests
```bash
go test -v -run TestFullIntegrationSuite ./...
```

### Run Specific Test Categories
```bash
# General connectivity tests
go test -v -run TestPing ./...

# Account information tests
go test -v -run TestAccount ./...

# User data stream tests
go test -v -run TestUserDataStream ./...

# Margin trading tests
go test -v -run TestMarginTrading ./...

# UM Futures tests
go test -v -run TestUMFutures ./...

# CM Futures tests
go test -v -run TestCMFutures ./...

# Portfolio margin specific tests
go test -v -run TestPortfolioMargin ./...
```

### Run Individual Test Functions
```bash
# Test specific functionality
go test -v -run TestAccountInfo ./...
go test -v -run TestListenKeyManagement ./...
go test -v -run TestMarginLoan ./...
```

## Test Categories

### 1. General & System Tests
- **Ping**: Basic connectivity test
- **Rate Limit**: User rate limit information

### 2. Account Management Tests
- **Account Info**: Portfolio margin account information
- **Account Balance**: Account balance across all markets

### 3. Asset Collection & Transfer Tests
- **Asset Collection**: Fund collection by specific asset
- **Auto Collection**: Automated fund collection
- **BNB Transfer**: BNB transfer operations

### 4. Repay & Negative Balance Tests
- **Repay Futures**: Repay futures negative balance
- **Repay Switch**: Auto-repay configuration

### 5. Margin Trading Tests
- **Margin Loan**: Margin loan operations
- **Margin Order**: Margin order management
- **Margin OCO**: One-Cancels-Other margin orders
- **Margin Repay**: Margin debt repayment

### 6. UM Futures Tests
- **UM Order**: USD-M futures order operations
- **UM Conditional**: Conditional order operations
- **UM Leverage**: Leverage management
- **UM Position**: Position management
- **UM Account**: Account information

### 7. CM Futures Tests
- **CM Order**: Coin-M futures order operations
- **CM Conditional**: Conditional order operations
- **CM Leverage**: Leverage management
- **CM Position**: Position management
- **CM Account**: Account information

### 8. Portfolio Margin Specific Tests
- **Portfolio Interest**: Interest history
- **Negative Balance Exchange**: Negative balance exchange records

### 9. User Data Stream Tests
- **Listen Key**: User data stream management

## Authentication Methods

### HMAC-SHA256 (Default)
```bash
export BINANCE_API_KEY="your_api_key"
export BINANCE_SECRET_KEY="your_secret_key"
```

### RSA Signing
```bash
export BINANCE_RSA_API_KEY="your_rsa_api_key"
export BINANCE_RSA_PRIVATE_KEY_PATH="/path/to/rsa_private_key.pem"
```

### Ed25519 Signing (Recommended)
```bash
export BINANCE_ED25519_API_KEY="your_ed25519_api_key"
export BINANCE_ED25519_PRIVATE_KEY_PATH="/path/to/ed25519_private_key.pem"
```

### Test All Authentication Methods
```bash
export TEST_ALL_AUTH_TYPES="true"
```

## Test Configuration

### Test Symbols
```bash
export BINANCE_TEST_SYMBOL_SPOT="BTCUSDT"
export BINANCE_TEST_SYMBOL_UM="BTCUSDT"
export BINANCE_TEST_SYMBOL_CM="BTCUSD_PERP"
```

### Test Order Parameters
```bash
export BINANCE_TEST_ORDER_QUANTITY="0.001"
export BINANCE_TEST_ORDER_PRICE="30000"
```

## Important Notes

### Testnet vs Production
- **Testnet**: Limited Portfolio Margin support
- **Production**: Full functionality but uses real money
- **Default**: Uses production with read-only operations

### Rate Limiting
- **Minimum Interval**: 2 seconds between requests
- **Automatic**: Built-in rate limiting in tests
- **Monitoring**: Request counts are tracked

### Error Handling
- **Testnet Limitations**: Tests skip unavailable endpoints
- **Portfolio Margin Errors**: Specific error handling for PM requirements
- **API Errors**: Detailed error logging and reporting

### Safety Features
- **Feature Toggles**: Destructive operations disabled by default
- **Environment Variables**: Explicit opt-in for dangerous operations
- **Logging**: Comprehensive logging of all operations

## Coverage Tracking

The test suite tracks API coverage in `API_COVERAGE.md`. After running tests, check the coverage file to see which endpoints have been tested and which need attention.

## Troubleshooting

### Common Issues
1. **Portfolio Margin Not Enabled**: Ensure your account has PM enabled
2. **Testnet Limitations**: Many endpoints not available on testnet
3. **API Key Permissions**: Ensure API key has Portfolio Margin permissions
4. **Rate Limits**: Built-in rate limiting should prevent most issues

### Debug Mode
Enable detailed logging:
```bash
go test -v -run TestFullIntegrationSuite ./... > test_output.log 2>&1
```

## Contributing

When adding new tests:
1. Follow the existing pattern in other test files
2. Use environment variables for destructive operations
3. Add comprehensive error handling
4. Update `API_COVERAGE.md` when implementing new endpoints
5. Include both success and error case testing

## Files Structure

```
src/binance/go/rest/pmargin/
├── API_COVERAGE.md              # API coverage tracking
├── README.md                    # This file
├── env.example                  # Environment configuration template
├── go.mod                       # Go module definition
├── main_test.go                 # Test suite entry point
├── integration_test.go          # Test framework and utilities
├── testnet_helpers.go           # Testnet-specific helpers
├── general_test.go              # General & system tests
├── account_test.go              # Account management tests
├── user_data_stream_test.go     # User data stream tests
├── rate_limit_test.go           # Rate limit tests
├── asset_collection_test.go     # Asset collection tests
├── repay_test.go                # Repay operation tests
├── margin_trading_test.go       # Margin trading tests
├── portfolio_margin_test.go     # Portfolio margin specific tests
└── [additional test files]     # UM/CM futures tests (to be added)
```