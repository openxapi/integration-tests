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