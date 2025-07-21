package main

import (
	"context"
	"strings"
	"testing"
	"time"

	openapi "github.com/openxapi/binance-go/rest/options"
)

// testMarketDataPing tests the ping endpoint
func testMarketDataPing(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetPingV1(ctx).Execute()
	
	if handleTestnetError(t, err, httpResp, "GetPingV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetPingV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	t.Logf("Ping successful: %+v", resp)
}

// testMarketDataTime tests the server time endpoint
func testMarketDataTime(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetTimeV1(ctx).Execute()
	
	if handleTestnetError(t, err, httpResp, "GetTimeV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetTimeV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp.ServerTime == nil {
		t.Fatal("ServerTime should not be nil")
	}
	
	// Check if server time is reasonable (within 5 minutes of current time)
	now := time.Now().UnixMilli()
	diff := abs(now - *resp.ServerTime)
	if diff > 5*60*1000 { // 5 minutes in milliseconds
		t.Errorf("Server time seems incorrect. Server: %d, Local: %d, Diff: %d ms", 
			*resp.ServerTime, now, diff)
	}
	
	t.Logf("Server time: %d (%s)", *resp.ServerTime, time.UnixMilli(*resp.ServerTime))
}

// testMarketDataExchangeInfo tests the exchange info endpoint
func testMarketDataExchangeInfo(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetExchangeInfoV1(ctx).Execute()
	
	if handleTestnetError(t, err, httpResp, "GetExchangeInfoV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetExchangeInfoV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp.Timezone == nil {
		t.Fatal("Timezone should not be nil")
	}
	
	if resp.ServerTime == nil {
		t.Fatal("ServerTime should not be nil")
	}
	
	if resp.OptionContracts == nil {
		t.Fatal("OptionContracts should not be nil")
	}
	
	t.Logf("Exchange info - Timezone: %s, Option contracts: %d", 
		*resp.Timezone, len(resp.OptionContracts))
	
	// Log first few option contracts if available
	if len(resp.OptionContracts) > 0 {
		for i, contract := range resp.OptionContracts {
			if i >= 3 { // Only log first 3 contracts
				break
			}
			t.Logf("Contract %d: BaseAsset=%s, QuoteAsset=%s, SettleAsset=%s, Underlying=%s", 
				i+1, 
				getStringValue(contract.BaseAsset),
				getStringValue(contract.QuoteAsset),
				getStringValue(contract.SettleAsset),
				getStringValue(contract.Underlying))
		}
	}
}

// testMarketDataDepth tests the order book depth endpoint
func testMarketDataDepth(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	// First get available symbols from exchange info
	symbol := getTestOptionsSymbol(t, client, ctx)
	if symbol == "" {
		t.Skip("No options symbols available for testing")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetDepthV1(ctx).
		Symbol(symbol).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetDepthV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetDepthV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp.Bids == nil || resp.Asks == nil {
		t.Fatal("Bids and Asks should not be nil")
	}
	
	t.Logf("Order book for %s - Bids: %d, Asks: %d", 
		symbol, len(resp.Bids), len(resp.Asks))
	
	// Log top bids and asks
	if len(resp.Bids) > 0 {
		bid := resp.Bids[0]
		t.Logf("Top bid: [%s, %s]", bid[0], bid[1])
	}
	
	if len(resp.Asks) > 0 {
		ask := resp.Asks[0]
		t.Logf("Top ask: [%s, %s]", ask[0], ask[1])
	}
}

// testMarketDataTrades tests the recent trades endpoint
func testMarketDataTrades(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	symbol := getTestOptionsSymbol(t, client, ctx)
	if symbol == "" {
		t.Skip("No options symbols available for testing")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetTradesV1(ctx).
		Symbol(symbol).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetTradesV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetTradesV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Recent trades for %s: %d trades", symbol, len(resp))
	
	// Log first few trades
	if len(resp) > 0 {
		for i, trade := range resp {
			if i >= 3 { // Only log first 3 trades
				break
			}
			t.Logf("Trade %d: ID=%d, Price=%s, Qty=%s, Time=%d", 
				i+1,
				getInt64Value(trade.Id),
				getStringValue(trade.Price),
				getStringValue(trade.Qty),
				getInt64Value(trade.Time))
		}
	}
}

// testMarketDataTicker tests the 24hr ticker endpoint
func testMarketDataTicker(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	symbol := getTestOptionsSymbol(t, client, ctx)
	if symbol == "" {
		t.Skip("No options symbols available for testing")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetTickerV1(ctx).
		Symbol(symbol).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetTickerV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetTickerV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("24hr ticker for %s:", symbol)
	if len(resp) > 0 {
		ticker := resp[0]
		t.Logf("  Symbol: %s", getStringValue(ticker.Symbol))
		t.Logf("  Price Change: %s", getStringValue(ticker.PriceChange))
		t.Logf("  Price Change Percent: %s", getStringValue(ticker.PriceChangePercent))
		t.Logf("  Last Price: %s", getStringValue(ticker.LastPrice))
		t.Logf("  Volume: %s", getStringValue(ticker.Volume))
		t.Logf("  Amount: %s", getStringValue(ticker.Amount))
	}
}

// testMarketDataKlines tests the kline/candlestick endpoint
func testMarketDataKlines(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	symbol := getTestOptionsSymbol(t, client, ctx)
	if symbol == "" {
		t.Skip("No options symbols available for testing")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetKlinesV1(ctx).
		Symbol(symbol).
		Interval("1m").
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetKlinesV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetKlinesV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Klines for %s (1m interval): %d klines", symbol, len(resp))
	
	// Log first few klines
	if len(resp) > 0 {
		for i, kline := range resp {
			if i >= 3 { // Only log first 3 klines
				break
			}
			t.Logf("Kline %d: OpenTime=%d, Open=%s, High=%s, Low=%s, Close=%s, Volume=%s", 
				i+1, 
				getInt64Value(kline.OpenTime),
				getStringValue(kline.Open),
				getStringValue(kline.High),
				getStringValue(kline.Low),
				getStringValue(kline.Close),
				getStringValue(kline.Volume))
		}
	}
}

// testMarketDataMark tests the mark price endpoint
func testMarketDataMark(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	symbol := getTestOptionsSymbol(t, client, ctx)
	if symbol == "" {
		t.Skip("No options symbols available for testing")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetMarkV1(ctx).
		Symbol(symbol).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetMarkV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetMarkV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Mark price for %s:", symbol)
	if len(resp) > 0 {
		mark := resp[0]
		t.Logf("  Symbol: %s", getStringValue(mark.Symbol))
		t.Logf("  Mark Price: %s", getStringValue(mark.MarkPrice))
		t.Logf("  Bid IV: %s", getStringValue(mark.BidIV))
		t.Logf("  Ask IV: %s", getStringValue(mark.AskIV))
		t.Logf("  Delta: %s", getStringValue(mark.Delta))
		t.Logf("  Gamma: %s", getStringValue(mark.Gamma))
		t.Logf("  Vega: %s", getStringValue(mark.Vega))
		t.Logf("  Theta: %s", getStringValue(mark.Theta))
	}
}

// testMarketDataIndex tests the index price endpoint
func testMarketDataIndex(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetIndexV1(ctx).
		Underlying("BTCUSDT").
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetIndexV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetIndexV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Index price for BTC:")
	t.Logf("  Index Price: %s", getStringValue(resp.IndexPrice))
	t.Logf("  Time: %d", getInt64Value(resp.Time))
}

// testMarketDataOpenInterest tests the open interest endpoint
func testMarketDataOpenInterest(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	// Get exchange info to find available expiration dates and underlying assets
	exchangeResp, _, err := client.OptionsAPI.GetExchangeInfoV1(ctx).Execute()
	if err != nil {
		t.Fatalf("Failed to get exchange info for open interest test: %v", err)
	}
	
	if exchangeResp.OptionSymbols == nil || len(exchangeResp.OptionSymbols) == 0 {
		t.Skip("No option symbols available for open interest testing")
	}
	
	// Find the first available symbol and extract underlying asset and expiration
	var underlyingAsset, expiration string
	for _, optionSymbol := range exchangeResp.OptionSymbols {
		if optionSymbol.Symbol != nil && optionSymbol.Underlying != nil {
			symbol := *optionSymbol.Symbol
			// Extract expiration from symbol format: "BTC-250926-110000-C"
			parts := strings.Split(symbol, "-")
			if len(parts) >= 2 {
				underlyingAsset = parts[0] // e.g., "BTC"
				expiration = parts[1]      // e.g., "250926"
				break
			}
		}
	}
	
	if underlyingAsset == "" || expiration == "" {
		t.Skip("Could not extract underlying asset and expiration from available symbols")
	}
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetOpenInterestV1(ctx).
		UnderlyingAsset(underlyingAsset).
		Expiration(expiration).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetOpenInterestV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetOpenInterestV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Open interest data for %s (expiration: %s): %d entries", underlyingAsset, expiration, len(resp))
	if len(resp) > 0 {
		oi := resp[0]
		t.Logf("  Symbol: %s", getStringValue(oi.Symbol))
		t.Logf("  Sum Open Interest: %s", getStringValue(oi.SumOpenInterest))
		t.Logf("  Sum Open Interest USD: %s", getStringValue(oi.SumOpenInterestUsd))
		t.Logf("  Timestamp: %s", getStringValue(oi.Timestamp))
	}
}

// Helper functions

func getTestOptionsSymbol(t *testing.T, client *openapi.APIClient, ctx context.Context) string {
	// Get exchange info to find available symbols
	resp, httpResp, err := client.OptionsAPI.GetExchangeInfoV1(ctx).Execute()
	if err != nil || httpResp.StatusCode != 200 {
		t.Logf("Cannot get exchange info for symbol lookup: %v", err)
		return ""
	}
	
	if resp.OptionSymbols == nil || len(resp.OptionSymbols) == 0 {
		return ""
	}
	
	// Find a tradeable symbol
	for _, symbol := range resp.OptionSymbols {
		if symbol.Symbol != nil {
			return *symbol.Symbol
		}
	}
	
	// Fallback to first symbol
	if len(resp.OptionSymbols) > 0 && resp.OptionSymbols[0].Symbol != nil {
		return *resp.OptionSymbols[0].Symbol
	}
	
	return ""
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func getFloat64Value(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}