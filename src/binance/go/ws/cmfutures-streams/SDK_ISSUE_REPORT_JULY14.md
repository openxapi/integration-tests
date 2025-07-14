# SDK Issue Report - July 14, 2025

## 🚨 **Critical SDK Issues Found**

After analyzing the latest test run (`out.log`), I've identified **2 critical SDK issues** and **symbol format inconsistencies** that need immediate attention.

## **Issue 1: CombinedStreamEventHandler Not Working**

### **Problem**
```
TestFullIntegrationSuite/CombinedStreamEventHandler FAILED
enhanced_features_test.go:179: Expected to receive at least one CombinedStreamEvent
```

### **Analysis**
- The `OnCombinedStreamEvent` handler is not receiving any events
- Individual event handlers (like `OnMarkPriceEvent`) work correctly in combined streams
- This suggests the combined stream wrapper event is not being properly dispatched

### **Expected Behavior**
When subscribing to combined streams, the `OnCombinedStreamEvent` handler should receive wrapped events containing:
- Stream name
- Raw event data
- Event metadata

### **Current Behavior**
- Individual event handlers receive events correctly ✅
- Combined stream event handler receives 0 events ❌

## **Issue 2: Individual Index Price Streams Not Working**

### **Problem**
```
TestFullIntegrationSuite/IndividualIndexPriceStream FAILED
integration_test.go:764: Failed to receive indexPriceUpdate events: timeout waiting for indexPriceUpdate events: expected 3, got 0
```

### **Analysis**
- Stream: `BTCUSD@indexPrice@1s`
- Expected: 3 `indexPriceUpdate` events
- Received: 0 events
- This may be a testnet limitation or SDK parsing issue

### **Possible Causes**
1. **Testnet Limitation**: Index price streams may not be available on testnet
2. **Symbol Format**: Mixed case `BTCUSD` vs lowercase `btcusd`
3. **Event Type Mapping**: `indexPriceUpdate` events not properly mapped to handlers

## **Issue 3: Symbol Format Inconsistencies**

### **Working Patterns (Lowercase)**
✅ **These work and receive events:**
```
btcusd_perp@markPrice@1s     → Received 3+ markPrice events
btcusd_perp@ticker           → Received 1 ticker event  
btcusd_perp@depth            → Received 1 depthUpdate event
btcusd_perp@kline_1m         → Received 1 kline event
btcusd_perp@bookTicker       → Received 1 bookTicker event
```

### **Problematic Patterns (Uppercase/Mixed)**
❌ **These don't receive events:**
```
BTCUSD_PERP@aggTrade         → Received 0 events
BTCUSD@indexPriceKline_1m    → Received 0 events
BTCUSD_PERP@markPriceKline_1m → Received 0 events
BTCUSD@indexPrice@1s         → Received 0 events (FAILED TEST)
```

### **Pattern Analysis**
The SDK appears to be **case-sensitive** for symbol names, with **lowercase being required** for reliable event reception.

## **Rate Limiting Issue Observed**

### **Problem**
```
2025/07/14 10:59:56 Error reading message: websocket: close 1008 (policy violation): Too many requests
```

### **Analysis**
- Rapid subscription test triggered Binance rate limiting
- Connection closed with policy violation error
- Subsequent operations failed with "websocket not connected"

### **Recommendation**
The SDK should implement **better rate limiting protection** and **automatic reconnection** when rate limited.

## **Recommendations for SDK Team**

### **High Priority Fixes**

1. **Fix CombinedStreamEvent Dispatching**
   - Ensure `OnCombinedStreamEvent` handlers receive wrapped events
   - Verify combined stream event processing pipeline
   - Add comprehensive tests for combined stream event handlers

2. **Investigate Index Price Stream Support**
   - Verify if index price streams work on mainnet
   - Add proper error handling for testnet limitations
   - Document which streams are/aren't available on testnet

3. **Standardize Symbol Format Handling**
   - Implement automatic case conversion to lowercase
   - Add validation for symbol formats
   - Document required symbol format conventions

### **Medium Priority Improvements**

4. **Enhanced Rate Limiting Protection**
   - Implement client-side rate limiting
   - Add automatic backoff and retry logic
   - Graceful handling of rate limit violations

5. **Better Error Messaging**
   - Distinguish between testnet limitations and actual errors
   - Provide clearer error messages for symbol format issues
   - Add debugging information for stream subscription failures

## **Integration Test Updates Needed**

Based on these findings, the following integration test updates are required:

1. **Fix symbol formats** to use consistent lowercase
2. **Update CombinedStreamEventHandler test** to work with fixed SDK
3. **Add testnet-aware handling** for index price streams
4. **Implement rate limiting protection** in rapid subscription tests

## **Summary**

- **Total Test Cases**: 41
- **Passed**: 39 (95.1%)
- **Failed**: 2 (4.9%)
- **Key Issues**: CombinedStreamEvent handler, Index price streams, Symbol format consistency

The SDK is **95.1% functional** but needs fixes for combined stream event handling and index price stream support to achieve 100% reliability.