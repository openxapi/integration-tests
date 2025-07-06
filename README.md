# OpenXAPI Integration Tests

This repository contains integration tests for OpenXAPI-generated SDK clients across different exchanges and API types.

## Repository Structure

```
integration-tests/
├── binance/              # Binance exchange tests
│   ├── asyncapi/        # WebSocket API tests
│   │   ├── spot/       # Spot trading WebSocket tests
│   │   └── umfutures/  # USD-M Futures WebSocket tests
│   └── openapi/        # REST API tests (future)
└── okx/                 # OKX exchange tests (future)
```

## Running Integration Tests

### Prerequisites

1. Set up environment variables in `.env` files:
   ```bash
   cp binance/asyncapi/spot/env.example binance/asyncapi/spot/.env
   cp binance/asyncapi/umfutures/env.example binance/asyncapi/umfutures/.env
   ```

2. Configure your API credentials in the `.env` files.

### Running Tests

#### Binance Spot WebSocket Tests
```bash
cd binance/asyncapi/spot
go test -v ./...
```

#### Binance USD-M Futures WebSocket Tests
```bash
cd binance/asyncapi/umfutures
go test -v ./...
```

## Test Categories

### Public Tests
Tests for public endpoints that don't require authentication:
- Market data subscriptions
- Ticker information
- Order book depth

### Trading Tests
Tests for authenticated trading operations:
- Order placement and cancellation
- Account balance queries
- Position management

### Session Tests
Tests for WebSocket session management:
- Session authentication
- Connection handling
- Heartbeat/ping-pong

### User Data Tests
Tests for user data stream operations:
- User data stream lifecycle
- Account updates
- Order updates

## Environment Variables

### Common Variables
- `BINANCE_API_KEY`: Your Binance API key
- `BINANCE_SECRET_KEY`: Your Binance secret key
- `BINANCE_WS_URL`: WebSocket endpoint URL

### Key Type Support
The tests support multiple key types:
- `HMAC`: HMAC-SHA256 (default)
- `RSA`: RSA signature
- `ED25519`: Ed25519 signature

## Notes

- Integration tests interact with real exchange endpoints
- Use testnet environments when available to avoid affecting real trading
- Some tests may require specific account permissions or balances
- Rate limits apply - tests include appropriate delays

## Contributing

When adding new integration tests:
1. Follow the existing directory structure
2. Include appropriate environment variable handling
3. Add comprehensive error checking
4. Document any special requirements
5. Update this README with new test information