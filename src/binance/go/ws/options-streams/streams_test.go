package streamstest

import (
	"testing"
)

// TestIndexPriceStream tests index price stream functionality
func TestIndexPriceStream(t *testing.T) {
	testStreamSubscriptionWithGracefulTimeout(
		t,
		"ETHUSDT@index", 
		"indexPrice",
		1,
		"Index price streams provide underlying asset price data every 1 second",
	)
}

// TestKlineStream tests kline stream functionality
func TestKlineStream(t *testing.T) {
	// Get a dynamic active symbol
	symbol, err := selectNearestExpirySymbol("BTC", "C")
	if err != nil {
		t.Skipf("No active BTC call options available for kline test: %v", err)
		return
	}
	
	streamName := symbol + "@kline_1m"
	t.Logf("Testing kline stream with active symbol: %s", streamName)
	
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"kline",
		1,
		"Kline streams provide candlestick data but may be sparse for individual option symbols",
	)
}

// TestMarkPriceStream tests mark price stream functionality
func TestMarkPriceStream(t *testing.T) {
	testStreamSubscriptionWithGracefulTimeout(
		t,
		"ETH@markPrice",
		"markPrice",
		1,
		"Mark price streams provide pricing data for all options on an underlying but may have low frequency",
	)
}

// TestNewSymbolInfoStream tests new symbol info stream functionality
func TestNewSymbolInfoStream(t *testing.T) {
	testStreamSubscriptionWithGracefulTimeout(
		t,
		"option_pair",
		"newSymbolInfo",
		1,
		"New symbol info streams only send data when new options contracts are listed",
	)
}

// TestOpenInterestStream tests open interest stream functionality
func TestOpenInterestStream(t *testing.T) {
	// Get active expiration dates for ETH
	expirations, err := getActiveExpirationDates("ETH")
	if err != nil || len(expirations) == 0 {
		t.Skipf("No active ETH expiration dates available: %v", err)
		return
	}
	
	// Use the nearest expiration
	streamName := "ETH@openInterest@" + expirations[0]
	t.Logf("Testing open interest stream with active expiration: %s", streamName)
	
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"openInterest",
		1,
		"Open interest streams update every minute and may have sparse data for specific expirations",
	)
}

// TestPartialDepthStream tests partial depth stream functionality
func TestPartialDepthStream(t *testing.T) {
	// Get a dynamic active symbol
	symbol, err := selectATMSymbol("BTC", "C")
	if err != nil {
		t.Skipf("No active BTC call options available for depth test: %v", err)
		return
	}
	
	streamName := symbol + "@depth10"
	t.Logf("Testing partial depth stream with active symbol: %s", streamName)
	
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"partialDepth",
		1,
		"Partial depth streams provide order book data but may be sparse for individual option symbols",
	)
}

// TestTickerStream tests individual ticker stream functionality
func TestTickerStream(t *testing.T) {
	// Get a dynamic active symbol
	symbol, err := selectATMSymbol("BTC", "C")
	if err != nil {
		t.Skipf("No active BTC call options available for ticker test: %v", err)
		return
	}
	
	streamName := symbol + "@ticker"
	t.Logf("Testing ticker stream with active symbol: %s", streamName)
	
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"ticker",
		1,
		"Individual ticker streams provide 24h statistics but may have low update frequency",
	)
}

// TestTickerByUnderlyingStream tests ticker by underlying stream functionality
func TestTickerByUnderlyingStream(t *testing.T) {
	// Get active expiration dates for ETH
	expirations, err := getActiveExpirationDates("ETH")
	if err != nil || len(expirations) == 0 {
		t.Skipf("No active ETH expiration dates available: %v", err)
		return
	}
	
	// Use the nearest expiration
	streamName := "ETH@ticker@" + expirations[0]
	t.Logf("Testing ticker by underlying stream with active expiration: %s", streamName)
	
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"tickerByUnderlying",
		1,
		"Ticker by underlying streams provide aggregated data but may have low frequency updates",
	)
}

// TestTradeStream tests trade stream functionality
func TestTradeStream(t *testing.T) {
	// Get a dynamic active symbol (preferably ATM for more trading activity)
	symbol, err := selectATMSymbol("BTC", "C")
	if err != nil {
		t.Skipf("No active BTC call options available for trade test: %v", err)
		return
	}
	
	streamName := symbol + "@trade"
	t.Logf("Testing trade stream with active symbol: %s", streamName)
	
	// Test individual symbol trade stream
	testStreamSubscriptionWithGracefulTimeout(
		t,
		streamName,
		"trade",
		1,
		"Individual symbol trade streams only send data when that specific option is traded",
	)
}

// TestTradeStreamByUnderlying tests trade stream by underlying functionality
func TestTradeStreamByUnderlying(t *testing.T) {
	// Test underlying asset trade stream
	testStreamSubscriptionWithGracefulTimeout(
		t,
		"ETH@trade",
		"trade",
		1,
		"Underlying trade streams aggregate all options trades but may have sparse activity",
	)
}

// TestPartialDepthStreamWithSpeed tests partial depth stream with different speeds
func TestPartialDepthStreamWithSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partial depth speed tests in short mode")
	}

	// Test different speed options
	speeds := []string{"100ms", "500ms", "1000ms"}
	
	for _, speed := range speeds {
		t.Run("Speed_"+speed, func(t *testing.T) {
			// Get a dynamic active symbol
			symbol, err := selectATMSymbol("BTC", "C")
			if err != nil {
				t.Skipf("No active BTC call options available for depth speed test: %v", err)
				return
			}
			
			streamName := symbol + "@depth10@" + speed
			t.Logf("Testing depth stream with speed %s and active symbol: %s", speed, streamName)
			
			testStreamSubscriptionWithGracefulTimeout(
				t,
				streamName,
				"partialDepth",
				1,
				"Depth streams with specific speeds may have varying update frequencies based on market activity",
			)
		})
	}
}

// TestPartialDepthStreamWithLevels tests partial depth stream with different levels
func TestPartialDepthStreamWithLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partial depth levels tests in short mode")
	}

	// Test different depth levels
	levels := []string{"10", "20", "50", "100"}
	
	for _, level := range levels {
		t.Run("Level_"+level, func(t *testing.T) {
			// Get a dynamic active symbol
			symbol, err := selectATMSymbol("BTC", "C")
			if err != nil {
				t.Skipf("No active BTC call options available for depth level test: %v", err)
				return
			}
			
			streamName := symbol + "@depth" + level
			t.Logf("Testing depth stream with level %s and active symbol: %s", level, streamName)
			
			testStreamSubscriptionWithGracefulTimeout(
				t,
				streamName,
				"partialDepth",
				1,
				"Depth streams with different levels provide varying amounts of order book data",
			)
		})
	}
}

// TestKlineStreamWithIntervals tests kline streams with different intervals
func TestKlineStreamWithIntervals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping kline interval tests in short mode")
	}

	// Test different kline intervals
	intervals := []string{"1m", "5m", "15m", "1h", "1d"}
	
	for _, interval := range intervals {
		t.Run("Interval_"+interval, func(t *testing.T) {
			// Get a dynamic active symbol
			symbol, err := selectNearestExpirySymbol("BTC", "C")
			if err != nil {
				t.Skipf("No active BTC call options available for kline interval test: %v", err)
				return
			}
			
			streamName := symbol + "@kline_" + interval
			t.Logf("Testing kline stream with interval %s and active symbol: %s", interval, streamName)
			
			testStreamSubscriptionWithGracefulTimeout(
				t,
				streamName,
				"kline",
				1,
				"Kline streams with different intervals may have varying data availability",
			)
		})
	}
}

// TestMultipleStreamTypes tests connecting to multiple different stream types sequentially
func TestMultipleStreamTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple streams tests in short mode")
	}

	t.Log("Testing multiple stream types for comprehensive options data coverage")

	// Test connecting to various stream types sequentially
	streamTests := []struct {
		name       string
		streamPath string
		description string
	}{
		{
			name:       "IndexPrice",
			streamPath: "ETHUSDT@index",
			description: "Index price for underlying asset",
		},
		{
			name:       "MarkPrice",
			streamPath: "ETH@markPrice",
			description: "Mark prices for all ETH options",
		},
		{
			name:       "Ticker",
			streamPath: "", // Will be set dynamically
			description: "24h ticker for specific option",
		},
	}

	for _, streamTest := range streamTests {
		t.Run(streamTest.name, func(t *testing.T) {
			// Use appropriate event type for each stream
			var eventType string
			streamPath := streamTest.streamPath
			
			switch streamTest.name {
			case "IndexPrice":
				eventType = "indexPrice"
			case "MarkPrice":
				eventType = "markPrice"
			case "Ticker":
				eventType = "ticker"
				// Get dynamic symbol for ticker test
				symbol, err := selectATMSymbol("BTC", "C")
				if err != nil {
					t.Skipf("No active BTC call options available for ticker test: %v", err)
					return
				}
				streamPath = symbol + "@ticker"
				t.Logf("Testing ticker stream with active symbol: %s", streamPath)
			default:
				eventType = "unknown"
			}
			
			testStreamSubscriptionWithGracefulTimeout(
				t,
				streamPath,
				eventType,
				1,
				streamTest.description,
			)
		})
	}

	t.Log("✅ Multiple stream types testing completed")
}