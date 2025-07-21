# Binance Go REST Portfolio Margin SDK - API Coverage

This document tracks the integration test coverage for the Binance Go REST Portfolio Margin SDK.

## Overview

- **SDK Package**: `github.com/openxapi/binance-go/rest/pmargin`
- **Total APIs**: 98 functions
- **Base URL**: `https://papi.binance.com`
- **Authentication**: HMAC-SHA256, RSA, ED25519 supported
- **Current Coverage**: 15/98 (15.3%)

## API List by Category

### 1. General & System APIs (1/1)
- [x] `GetPingV1()` - Test connectivity to the REST API *(general_test.go)*

### 2. Account Management APIs (2/2)
- [x] `GetAccountV1()` - Query account information (USER_DATA) *(account_test.go)*
- [x] `GetBalanceV1()` - Query account balance (USER_DATA) *(account_test.go)*

### 3. Asset Collection & Transfer APIs (3/3)
- [x] `CreateAssetCollectionV1()` - Fund Collection by Asset (TRADE) *(asset_collection_test.go)*
- [x] `CreateAutoCollectionV1()` - Fund Auto-collection (TRADE) *(asset_collection_test.go)*
- [x] `CreateBnbTransferV1()` - BNB transfer (TRADE) *(asset_collection_test.go)*

### 4. Repay & Negative Balance APIs (3/3)
- [x] `CreateRepayFuturesNegativeBalanceV1()` - Repay futures negative balance (USER_DATA) *(repay_test.go)*
- [x] `CreateRepayFuturesSwitchV1()` - Change auto-repay-futures status (TRADE) *(repay_test.go)*
- [x] `GetRepayFuturesSwitchV1()` - Query auto-repay-futures status (USER_DATA) *(repay_test.go)*

### 5. Margin Trading APIs (9/22)
- [ ] `CreateMarginLoanV1()` - Apply for margin loan (MARGIN)
- [ ] `CreateMarginOrderV1()` - Place new margin order (TRADE)
- [ ] `CreateMarginOrderOcoV1()` - Place margin OCO order (TRADE)
- [ ] `CreateMarginRepayDebtV1()` - Repay margin debt (TRADE)
- [ ] `CreateRepayLoanV1()` - Repay margin loan (MARGIN)
- [ ] `DeleteMarginOrderV1()` - Cancel margin order (TRADE)
- [ ] `DeleteMarginOrderListV1()` - Cancel margin OCO order (TRADE)
- [ ] `DeleteMarginAllOpenOrdersV1()` - Cancel all margin open orders (TRADE)
- [ ] `GetMarginOrderV1()` - Query margin order (USER_DATA)
- [ ] `GetMarginOrderListV1()` - Query margin OCO order (USER_DATA)
- [x] `GetMarginOpenOrdersV1()` - Query current margin open orders (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginOpenOrderListV1()` - Query current margin open OCO orders (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginAllOrdersV1()` - Query all margin orders (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginAllOrderListV1()` - Query all margin OCO orders (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginMyTradesV1()` - Query margin trade list (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginForceOrdersV1()` - Query margin force orders (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginMarginLoanV1()` - Query margin loan record (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginRepayLoanV1()` - Query margin repay record (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginMarginInterestHistoryV1()` - Query margin interest history (USER_DATA) *(margin_trading_test.go)*
- [x] `GetMarginMaxBorrowableV1()` - Query margin max borrowable (USER_DATA) *(margin_trading_test.go)*
- [ ] `GetMarginMaxWithdrawV1()` - Query margin max withdraw (USER_DATA)

### 6. UM Futures (USD-M) APIs (47/47)
- [ ] `CreateUmOrderV1()` - Place new UM order (TRADE)
- [ ] `CreateUmConditionalOrderV1()` - Place new UM conditional order (TRADE)
- [ ] `CreateUmLeverageV1()` - Change UM initial leverage (TRADE)
- [ ] `CreateUmPositionSideDualV1()` - Change UM position mode (TRADE)
- [ ] `CreateUmFeeBurnV1()` - Toggle BNB burn on UM futures (TRADE)
- [ ] `DeleteUmOrderV1()` - Cancel UM order (TRADE)
- [ ] `DeleteUmConditionalOrderV1()` - Cancel UM conditional order (TRADE)
- [ ] `DeleteUmAllOpenOrdersV1()` - Cancel all UM open orders (TRADE)
- [ ] `DeleteUmConditionalAllOpenOrdersV1()` - Cancel all UM conditional orders (TRADE)
- [ ] `UpdateUmOrderV1()` - Modify UM order (TRADE)
- [ ] `GetUmOrderV1()` - Query UM order (USER_DATA)
- [ ] `GetUmConditionalOpenOrderV1()` - Query UM conditional open order (USER_DATA)
- [ ] `GetUmOpenOrderV1()` - Query current UM open order (USER_DATA)
- [ ] `GetUmOpenOrdersV1()` - Query all current UM open orders (USER_DATA)
- [ ] `GetUmConditionalOpenOrdersV1()` - Query all UM conditional open orders (USER_DATA)
- [ ] `GetUmAllOrdersV1()` - Query all UM orders (USER_DATA)
- [ ] `GetUmConditionalAllOrdersV1()` - Query all UM conditional orders (USER_DATA)
- [ ] `GetUmConditionalOrderHistoryV1()` - Query UM conditional order history (USER_DATA)
- [ ] `GetUmOrderAmendmentV1()` - Query UM order modification history (TRADE)
- [ ] `GetUmUserTradesV1()` - Query UM trade list (USER_DATA)
- [ ] `GetUmAccountV1()` - Query UM account detail (USER_DATA)
- [ ] `GetUmAccountV2()` - Query UM account detail V2 (USER_DATA)
- [ ] `GetUmAccountConfigV1()` - Query UM account configuration (USER_DATA)
- [ ] `GetUmPositionRiskV1()` - Query UM position information (USER_DATA)
- [ ] `GetUmPositionSideDualV1()` - Query UM position mode (USER_DATA)
- [ ] `GetUmAdlQuantileV1()` - Query UM ADL quantile (USER_DATA)
- [ ] `GetUmCommissionRateV1()` - Query UM commission rate (USER_DATA)
- [ ] `GetUmForceOrdersV1()` - Query UM force orders (USER_DATA)
- [ ] `GetUmIncomeV1()` - Query UM income history (USER_DATA)
- [ ] `GetUmIncomeAsynV1()` - Query UM income async (USER_DATA)
- [ ] `GetUmIncomeAsynIdV1()` - Query UM income async by ID (USER_DATA)
- [ ] `GetUmLeverageBracketV1()` - Query UM leverage bracket (USER_DATA)
- [ ] `GetUmApiTradingStatusV1()` - Query UM API trading status (USER_DATA)
- [ ] `GetUmFeeBurnV1()` - Query UM fee burn status (USER_DATA)
- [ ] `GetUmSymbolConfigV1()` - Query UM symbol configuration (USER_DATA)
- [ ] `GetUmOrderAsynV1()` - Query UM order async (USER_DATA)
- [ ] `GetUmOrderAsynIdV1()` - Query UM order async by ID (USER_DATA)
- [ ] `GetUmTradeAsynV1()` - Query UM trade async (USER_DATA)
- [ ] `GetUmTradeAsynIdV1()` - Query UM trade async by ID (USER_DATA)

### 7. CM Futures (Coin-M) APIs (26/26)
- [ ] `CreateCmOrderV1()` - Place new CM order (TRADE)
- [ ] `CreateCmConditionalOrderV1()` - Place new CM conditional order (TRADE)
- [ ] `CreateCmLeverageV1()` - Change CM initial leverage (TRADE)
- [ ] `CreateCmPositionSideDualV1()` - Change CM position mode (TRADE)
- [ ] `DeleteCmOrderV1()` - Cancel CM order (TRADE)
- [ ] `DeleteCmConditionalOrderV1()` - Cancel CM conditional order (TRADE)
- [ ] `DeleteCmAllOpenOrdersV1()` - Cancel all CM open orders (TRADE)
- [ ] `DeleteCmConditionalAllOpenOrdersV1()` - Cancel all CM conditional orders (TRADE)
- [ ] `UpdateCmOrderV1()` - Modify CM order (TRADE)
- [ ] `GetCmOrderV1()` - Query CM order (USER_DATA)
- [ ] `GetCmConditionalOpenOrderV1()` - Query CM conditional open order (USER_DATA)
- [ ] `GetCmOpenOrderV1()` - Query current CM open order (USER_DATA)
- [ ] `GetCmOpenOrdersV1()` - Query all current CM open orders (USER_DATA)
- [ ] `GetCmConditionalOpenOrdersV1()` - Query all CM conditional open orders (USER_DATA)
- [ ] `GetCmAllOrdersV1()` - Query all CM orders (USER_DATA)
- [ ] `GetCmConditionalAllOrdersV1()` - Query all CM conditional orders (USER_DATA)
- [ ] `GetCmConditionalOrderHistoryV1()` - Query CM conditional order history (USER_DATA)
- [ ] `GetCmOrderAmendmentV1()` - Query CM order modification history (TRADE)
- [ ] `GetCmUserTradesV1()` - Query CM trade list (USER_DATA)
- [ ] `GetCmAccountV1()` - Query CM account detail (USER_DATA)
- [ ] `GetCmPositionRiskV1()` - Query CM position information (USER_DATA)
- [ ] `GetCmPositionSideDualV1()` - Query CM position mode (USER_DATA)
- [ ] `GetCmAdlQuantileV1()` - Query CM ADL quantile (USER_DATA)
- [ ] `GetCmCommissionRateV1()` - Query CM commission rate (USER_DATA)
- [ ] `GetCmForceOrdersV1()` - Query CM force orders (USER_DATA)
- [ ] `GetCmIncomeV1()` - Query CM income history (USER_DATA)
- [ ] `GetCmLeverageBracketV1()` - Query CM leverage bracket (USER_DATA)

### 8. Portfolio Margin Specific APIs (2/2)
- [x] `GetPortfolioInterestHistoryV1()` - Query portfolio margin interest history (USER_DATA) *(portfolio_margin_test.go)*
- [x] `GetPortfolioNegativeBalanceExchangeRecordV1()` - Query negative balance exchange record (USER_DATA) *(portfolio_margin_test.go)*

### 9. User Data Stream APIs (3/3)
- [x] `CreateListenKeyV1()` - Start user data stream (USER_STREAM) *(user_data_stream_test.go)*
- [x] `UpdateListenKeyV1()` - Extend user data stream (USER_STREAM) *(user_data_stream_test.go)*
- [x] `DeleteListenKeyV1()` - Close user data stream (USER_STREAM) *(user_data_stream_test.go)*

### 10. Rate Limit API (1/1)
- [x] `GetRateLimitOrderV1()` - Query user rate limit (USER_DATA) *(rate_limit_test.go)*

## Implementation Status

### Completed Test Files
- [x] `general_test.go` - General & system API tests (1 API)
- [x] `account_test.go` - Account management API tests (2 APIs)
- [x] `asset_collection_test.go` - Asset collection & transfer API tests (3 APIs)
- [x] `repay_test.go` - Repay & negative balance API tests (3 APIs)
- [x] `margin_trading_test.go` - Margin trading API tests (9 APIs)
- [x] `portfolio_margin_test.go` - Portfolio margin specific API tests (2 APIs)
- [x] `user_data_stream_test.go` - User data stream API tests (3 APIs)
- [x] `rate_limit_test.go` - Rate limit API tests (1 API)

### In Progress
- [ ] UM futures API tests (0/47 APIs)
- [ ] CM futures API tests (0/26 APIs)
- [ ] Complete margin trading API tests (13/22 APIs remaining)

### To Do
- [ ] Implement remaining margin trading API tests
- [ ] Implement UM futures API tests
- [ ] Implement CM futures API tests
- [ ] Add more comprehensive error handling tests
- [ ] Add performance benchmarking tests

## Notes

- Portfolio Margin APIs require special account type and permissions
- Some APIs may not be available on testnet (will be handled with t.Skip())
- Testing will focus on successful API calls and response structure validation
- Error handling tests will be included where appropriate
- Rate limiting will be implemented to respect API limits

## Authentication Requirements

- **HMAC-SHA256**: Most common authentication method
- **RSA**: Alternative authentication method
- **ED25519**: Alternative authentication method
- All methods require API key and secret configuration

## Test Environment

- **Primary**: Binance testnet where available
- **Fallback**: Production environment with minimal operations (read-only where possible)
- **Credentials**: Environment variables for API key and secret