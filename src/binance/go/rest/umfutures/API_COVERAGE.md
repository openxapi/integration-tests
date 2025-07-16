# Binance USD-M Futures REST API Test Coverage

This document tracks the test coverage for all endpoints in the Binance USD-M Futures REST API SDK.

## Overall Coverage Summary

- **Total Endpoints**: 103
- **Tested**: 22 (21.4%)
- **Passing**: 21 (20.4%)
- **Skipped (API Issues)**: 1 (1.0%)
- **Failed**: 0 (0%)
- **Untested**: 81 (78.6%)

## Test Coverage by Service

### FuturesAPIService (89 endpoints) - 24.7% Coverage

#### Public Endpoints (39 endpoints) - 56.4% Coverage

| Endpoint | Method | Description | Test File | Status |
|----------|--------|-------------|-----------|--------|
| GetPingV1 | GET | Test Connectivity | public_test.go | ✅ |
| GetTimeV1 | GET | Check Server Time | public_test.go | ✅ |
| GetExchangeInfoV1 | GET | Exchange Information | public_test.go | ✅ |
| GetDepthV1 | GET | Order Book | public_test.go | ✅ |
| GetTradesV1 | GET | Recent Trades List | public_test.go | ✅ |
| GetHistoricalTradesV1 | GET | Old Trades Lookup | public_test.go | ✅ |
| GetAggTradesV1 | GET | Compressed/Aggregate Trades List | public_test.go | ✅ |
| GetKlinesV1 | GET | Kline/Candlestick Data | public_test.go | ✅ |
| GetContinuousKlinesV1 | GET | Continuous Contract Kline/Candlestick Data | public_test.go | ✅ |
| GetIndexPriceKlinesV1 | GET | Index Price Kline/Candlestick Data | public_test.go | ✅ |
| GetMarkPriceKlinesV1 | GET | Mark Price Kline/Candlestick Data | public_test.go | ✅ |
| GetPremiumIndexKlinesV1 | GET | Premium index Kline Data | public_test.go | ✅ |
| GetTicker24hrV1 | GET | 24hr Ticker Price Change Statistics | public_test.go | ✅ |
| GetTickerPriceV1 | GET | Symbol Price Ticker | public_test.go | ✅ |
| GetTickerPriceV2 | GET | Symbol Price Ticker V2 | - | ❌ |
| GetTickerBookTickerV1 | GET | Symbol Order Book Ticker | public_test.go | ✅ |
| GetOpenInterestV1 | GET | Open Interest | public_test.go | ✅ |
| GetPremiumIndexV1 | GET | Mark Price | public_test.go | ✅ |
| GetFundingRateV1 | GET | Get Funding Rate History | public_test.go | ✅ |
| GetFundingInfoV1 | GET | Get Funding Rate Info | public_test.go | ⚠️ (API Not Available) |
| GetIndexInfoV1 | GET | Composite Index Symbol Information | public_test.go | ✅ |
| GetConstituentsV1 | GET | Query Index Price Constituents | public_test.go | ✅ |
| GetAssetIndexV1 | GET | Multi-Assets Mode Asset Index | public_test.go | ✅ |
| GetConvertExchangeInfoV1 | GET | List All Convert Pairs | - | ❌ |
| GetFuturesDataBasis | GET | Basis | - | ❌ |
| GetFuturesDataDeliveryPrice | GET | Quarterly Contract Settlement Price | - | ❌ |
| GetFuturesDataGlobalLongShortAccountRatio | GET | Long/Short Ratio | - | ❌ |
| GetFuturesDataOpenInterestHist | GET | Open Interest Statistics | - | ❌ |
| GetFuturesDataTakerlongshortRatio | GET | Taker Buy/Sell Volume | - | ❌ |
| GetFuturesDataTopLongShortAccountRatio | GET | Top Trader Long/Short Ratio (Accounts) | - | ❌ |
| GetFuturesDataTopLongShortPositionRatio | GET | Top Trader Long/Short Ratio (Positions) | - | ❌ |

#### User Data Endpoints (30 endpoints) - 0% Coverage

| Endpoint | Method | Description | Test File | Status |
|----------|--------|-------------|-----------|--------|
| GetAccountV2 | GET | Account Information V2 | - | ❌ |
| GetAccountV3 | GET | Account Information V3 | - | ❌ |
| GetBalanceV2 | GET | Futures Account Balance V2 | - | ❌ |
| GetBalanceV3 | GET | Futures Account Balance V3 | - | ❌ |
| GetAccountConfigV1 | GET | Futures Account Configuration | - | ❌ |
| GetPositionRiskV2 | GET | Position Information V2 | - | ❌ |
| GetPositionRiskV3 | GET | Position Information V3 | - | ❌ |
| GetUserTradesV1 | GET | Account Trade List | - | ❌ |
| GetAllOrdersV1 | GET | All Orders | - | ❌ |
| GetOpenOrdersV1 | GET | Current All Open Orders | - | ❌ |
| GetOpenOrderV1 | GET | Query Current Open Order | - | ❌ |
| GetOrderV1 | GET | Query Order | - | ❌ |
| GetIncomeV1 | GET | Get Income History | - | ❌ |
| GetForceOrdersV1 | GET | User's Force Orders | - | ❌ |
| GetAdlQuantileV1 | GET | Position ADL Quantile Estimation | - | ❌ |
| GetCommissionRateV1 | GET | User Commission Rate | - | ❌ |
| GetApiTradingStatusV1 | GET | Futures Trading Quantitative Rules Indicators | - | ❌ |
| GetSymbolConfigV1 | GET | Symbol Configuration | - | ❌ |
| GetLeverageBracketV1 | GET | Notional and Leverage Brackets | - | ❌ |
| GetPositionSideDualV1 | GET | Get Current Position Mode | - | ❌ |
| GetMultiAssetsMarginV1 | GET | Get Current Multi-Assets Mode | - | ❌ |
| GetFeeBurnV1 | GET | Get BNB Burn Status | - | ❌ |
| GetPositionMarginHistoryV1 | GET | Get Position Margin Change History | - | ❌ |
| GetOrderAmendmentV1 | GET | Get Order Modify History | - | ❌ |
| GetRateLimitOrderV1 | GET | Query User Rate Limit | - | ❌ |
| GetPmAccountInfoV1 | GET | Classic Portfolio Margin Account Information | - | ❌ |
| GetConvertOrderStatusV1 | GET | Order status | - | ❌ |
| GetIncomeAsynV1 | GET | Get Download Id For Futures Transaction History | - | ❌ |
| GetIncomeAsynIdV1 | GET | Get Futures Transaction History Download Link by Id | - | ❌ |
| GetOrderAsynV1 | GET | Get Download Id For Futures Order History | - | ❌ |
| GetOrderAsynIdV1 | GET | Get Futures Order History Download Link by Id | - | ❌ |
| GetTradeAsynV1 | GET | Get Download Id For Futures Trade History | - | ❌ |
| GetTradeAsynIdV1 | GET | Get Futures Trade Download Link by Id | - | ❌ |

#### Trading Endpoints (16 endpoints) - 0% Coverage

| Endpoint | Method | Description | Test File | Status |
|----------|--------|-------------|-----------|--------|
| CreateOrderV1 | POST | New Order | - | ❌ |
| CreateOrderTestV1 | POST | Test Order | - | ❌ |
| DeleteOrderV1 | DELETE | Cancel Order | - | ❌ |
| DeleteAllOpenOrdersV1 | DELETE | Cancel All Open Orders | - | ❌ |
| UpdateOrderV1 | PUT | Modify Order | - | ❌ |
| CreateBatchOrdersV1 | POST | Place Multiple Orders | - | ❌ |
| UpdateBatchOrdersV1 | PUT | Modify Multiple Orders | - | ❌ |
| DeleteBatchOrdersV1 | DELETE | Cancel Multiple Orders | - | ❌ |
| CreateLeverageV1 | POST | Change Initial Leverage | - | ❌ |
| CreateMarginTypeV1 | POST | Change Margin Type | - | ❌ |
| CreatePositionMarginV1 | POST | Modify Isolated Position Margin | - | ❌ |
| CreatePositionSideDualV1 | POST | Change Position Mode | - | ❌ |
| CreateMultiAssetsMarginV1 | POST | Change Multi-Assets Mode | - | ❌ |
| CreateFeeBurnV1 | POST | Toggle BNB Burn On Futures Trade | - | ❌ |
| CreateCountdownCancelAllV1 | POST | Auto-Cancel All Open Orders | - | ❌ |
| CreateConvertAcceptQuoteV1 | POST | Accept the offered quote | - | ❌ |
| CreateConvertGetQuoteV1 | POST | Send Quote Request | - | ❌ |

#### User Stream Endpoints (3 endpoints) - 0% Coverage

| Endpoint | Method | Description | Test File | Status |
|----------|--------|-------------|-----------|--------|
| CreateListenKeyV1 | POST | Start User Data Stream | - | ❌ |
| UpdateListenKeyV1 | PUT | Keepalive User Data Stream | - | ❌ |
| DeleteListenKeyV1 | DELETE | Close User Data Stream | - | ❌ |

### BinanceLinkAPIService (14 endpoints) - 0% Coverage

#### Referral Management (14 endpoints) - 0% Coverage

| Endpoint | Method | Description | Test File | Status |
|----------|--------|-------------|-----------|--------|
| GetApiReferralOverviewV1 | GET | Get Rebate Data Overview | - | ❌ |
| GetApiReferralIfNewUserV1 | GET | Query Client If The New User | - | ❌ |
| GetApiReferralIfNewUserPAPIV1 | GET | Query Client If The New User (PAPI) | - | ❌ |
| GetApiReferralCustomizationV1 | GET | Get Client Email Customized Id | - | ❌ |
| CreateApiReferralCustomizationV1 | POST | Customize Id For Client (For Partner) | - | ❌ |
| GetApiReferralUserCustomizationV1 | GET | Get User's Customize Id | - | ❌ |
| GetApiReferralUserCustomizationPAPIV1 | GET | Get User's Customize Id (PAPI) | - | ❌ |
| CreateApiReferralUserCustomizationV1 | POST | Customize Id For Client (For client) | - | ❌ |
| CreateApiReferralUserCustomizationPAPIV1 | POST | Customize Id For Client (For client)(PAPI) | - | ❌ |
| GetApiReferralRebateVolV1 | GET | Get Rebate Volume | - | ❌ |
| GetApiReferralTradeVolV1 | GET | Get User Trade Volume | - | ❌ |
| GetApiReferralTraderNumV1 | GET | Get Trader Number | - | ❌ |
| GetApiReferralTraderSummaryV1 | GET | Get Trader Detail | - | ❌ |

## Test Files Structure

### Planned Test Files:
- `public_test.go` - Public market data endpoints (39 endpoints)
- `account_test.go` - Account information and balance endpoints (30 endpoints)
- `trading_test.go` - Trading operations and order management (16 endpoints)
- `user_stream_test.go` - User data stream management (3 endpoints)
- `binance_link_test.go` - Referral and affiliate management (14 endpoints)
- `async_download_test.go` - Async download operations (6 endpoints)

## Recent Test Results (Latest Run)

**Date**: July 16, 2025  
**Total Duration**: 42.73s  
**Tests Run**: 22  
**Passed**: 21  
**Skipped**: 1 (GetFundingInfoV1 - API endpoint not available)  
**Failed**: 0  

### Issues Resolved

1. **Exchange Info Test** - Fixed deliveryDate field type mismatch handling
2. **Historical Trades Test** - Fixed authentication requirement (now uses AUTH required config)
3. **Funding Info Test** - Properly handles 404 response (endpoint not available on Binance API)

### Current Status

All public API tests are now passing successfully. The test suite runs reliably against Binance testnet with proper error handling for API limitations.

## Notes

1. Tests should be run against testnet environment
2. Some endpoints may require special permissions or account configurations
3. Rate limiting must be respected across all tests
4. Tests should handle graceful failures for unavailable endpoints
5. Each test file should be self-contained and runnable independently
6. **GetFundingInfoV1** endpoint is not supported by Binance API despite being in the SDK

## Next Steps

1. Implement public API tests first (no authentication required)
2. Add account information tests
3. Implement trading operation tests
4. Add user stream management tests
5. Complete with BinanceLink API tests
6. Implement async download tests