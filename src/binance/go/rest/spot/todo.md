# Binance Go REST API Integration Tests - TODO

## Current Status
- **Total Coverage**: 410/470+ endpoints (87.2%)
- **Remaining**: ~60 endpoints (12.8%)

## Completed Services ✅

### 1. SpotTradingAPI (42/42 endpoints) - 100% ✅
- All core trading, market data, and order management endpoints tested
- Files: `public_test.go`, `trading_test.go`, `account_test.go`, `oco_trading_test.go`, `sor_trading_test.go`

### 2. WalletAPI (40/40 endpoints) - 100% ✅  
- All wallet operations, deposits, withdrawals, transfers tested
- Files: `wallet_test.go`, `wallet_advanced_test.go`

### 3. MarginTradingAPI (62/62 endpoints) - 100% ✅
- All margin trading operations (cross/isolated) tested
- Files: `margin_trading_test.go`, `margin_advanced_test.go`

### 4. AlgoTradingAPI (11/11 endpoints) - 100% ✅
- All algorithmic trading endpoints tested
- File: `algo_trading_test.go`

### 5. SimpleEarnAPI (24/25 endpoints) - 96% ✅
- Nearly all flexible/locked savings endpoints tested
- File: `simple_earn_test.go`

### 6. StakingAPI (24/24 endpoints) - 100% ✅
- All ETH and SOL staking endpoints tested
- File: `staking_test.go`

### 7. CryptoLoanAPI (16/17 endpoints) - 94% ✅
- Most crypto loan endpoints tested
- File: `crypto_loan_test.go`

### 8. ConvertAPI (9/9 endpoints) - 100% ✅
- All conversion endpoints tested
- File: `convert_test.go`

### 9. PortfolioMarginProAPI (19/19 endpoints) - 100% ✅
- All portfolio margin endpoints tested
- File: `portfolio_margin_test.go`

### 10. VipLoanAPI (12/13 endpoints) - 92% ✅
- Most VIP loan endpoints tested
- File: `vip_loan_test.go`

### 11. MiningAPI (13/13 endpoints) - 100% ✅
- All mining endpoints tested
- File: `mining_test.go`

### 12. BinanceLinkAPI (46/46 endpoints) - 100% ✅
- All broker/affiliate endpoints tested
- File: `binance_link_test.go`

### 13. GiftCardAPI (6/6 endpoints) - 100% ✅
- All gift card endpoints tested
- File: `giftcard_test.go`

### 14. DualInvestmentAPI (5/5 endpoints) - 100% ✅
- All dual investment endpoints tested
- File: `dual_investment_test.go`

### 15. Small APIs (10/10 endpoints) - 100% ✅
- NFT, Fiat, C2C, Pay, CopyTrading, FuturesData, Rebate
- File: `small_apis_test.go`

## Remaining Work 🔄

### SubAccountAPI (47/58 endpoints) - 81% Coverage
**File**: `subaccount_test.go`
**Remaining**: ~11 endpoints

#### ❌ Still Need Testing:
1. **Managed Sub-Account Operations** (require investor account):
   - Managed sub-account deposit operations
   - Managed sub-account asset management
   - Managed sub-account query operations

2. **Special Permission Sub-Account Operations**:
   - Some advanced sub-account trading features
   - Broker-specific sub-account operations
   - VIP sub-account management

3. **BLVT Operations** (if available in sub-account context):
   - BLVT subscription/redemption for sub-accounts

4. **Options Trading Sub-Account Operations**:
   - Options account management for sub-accounts
   - Options trading permissions

#### 📝 Implementation Strategy:
- Create `subaccount_advanced_test.go` for remaining endpoints
- Handle graceful failures for endpoints requiring special permissions
- Use conditional testing based on account capabilities
- Add comprehensive error handling for permission-denied scenarios

### Summary of Remaining Tasks

| Service | Remaining Endpoints | Estimated Effort | Priority |
|---------|-------------------|------------------|----------|
| SubAccountAPI | ~11 endpoints | Medium | High |
| Various APIs | ~49 endpoints | Low-Medium | Medium |

### Notes on Remaining Endpoints

1. **Account Type Dependencies**: Many remaining endpoints require:
   - VIP account status
   - Broker account permissions  
   - Investor account type (for managed sub-accounts)
   - Special feature enablement

2. **Testnet Limitations**: Some endpoints may not be available on testnet:
   - VIP loan operations
   - Managed sub-account features
   - Some broker operations

3. **Testing Strategy**: 
   - Use conditional testing with `t.Skip()` for unavailable features
   - Test error handling for permission-denied scenarios
   - Verify response structure even for failed requests

## Next Steps for Implementation

### Phase 1: Complete Sub-Account Testing
1. **Create** `subaccount_advanced_test.go`
2. **Implement** remaining 11 sub-account endpoints
3. **Test** graceful handling of permission errors
4. **Update** API_COVERAGE.md with results

### Phase 2: Comprehensive Review
1. **Audit** existing test files for any missed edge cases
2. **Enhance** error handling across all tests
3. **Optimize** test execution time and reliability
4. **Document** any permanent limitations (VIP-only features, etc.)

## Maintenance Notes

- **Update Frequency**: Check for new endpoints when SDK is updated
- **Coverage Goal**: Maintain >95% coverage of publicly available endpoints
- **Documentation**: Keep API_COVERAGE.md in sync with actual test coverage
- **Continuous Integration**: Ensure all tests pass on testnet environment

## Success Criteria

✅ **Achieved**: 87.2% overall coverage with comprehensive test suite
🎯 **Target**: 95% coverage of testnet-available endpoints
🔄 **Ongoing**: Maintain test reliability and update with new SDK releases