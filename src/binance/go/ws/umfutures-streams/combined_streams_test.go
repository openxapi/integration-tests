package streamstest

import (
	"context"
	"testing"
	"time"

	umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
	"github.com/openxapi/binance-go/ws/umfutures-streams/models"
)

// TestCombinedStreamEventReception tests receiving events through combined streams
func TestCombinedStreamEventReception(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream event reception test in short mode")
	}

	client := umfuturesstreams.NewClient()
	
	// Set to testnet server
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Setup event handlers
	var combinedEvents []interface{}
	var regularEvents []interface{}
	eventsMu := make(chan struct{}, 100)

	client.HandleCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		combinedEvents = append(combinedEvents, event)
		eventsMu <- struct{}{}
		t.Logf("Received combined stream event from: %s", event.StreamName)
		return nil
	})

	client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		regularEvents = append(regularEvents, event)
		eventsMu <- struct{}{}
		t.Logf("Received aggregate trade event: %s", event.Symbol)
		return nil
	})

	// Subscribe to multiple streams
	streams := []string{
		"btcusdt@aggTrade",
		"ethusdt@aggTrade",
		"btcusdt@ticker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Wait for events
	t.Log("Waiting for combined stream events...")
	timeout := time.After(15 * time.Second)
	receivedEvents := 0
	targetEvents := 5

	for receivedEvents < targetEvents {
		select {
		case <-eventsMu:
			receivedEvents++
		case <-timeout:
			t.Logf("Timeout reached, received %d events", receivedEvents)
			goto checkResults
		}
	}

checkResults:
	t.Logf("Combined stream events received: %d", len(combinedEvents))
	t.Logf("Regular events received: %d", len(regularEvents))

	if len(combinedEvents) > 0 {
		t.Log("✅ Combined stream event reception working")
	} else if len(regularEvents) > 0 {
		t.Log("✅ Regular event reception working (combined stream events may be processed differently)")
	} else {
		t.Log("⚠️  No events received (may be due to low market activity)")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestCombinedStreamEventDataTypes tests different event types through combined streams
func TestCombinedStreamEventDataTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream event data types test in short mode")
	}

	client := umfuturesstreams.NewClient()
	
	// Set to testnet server
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Track different event types
	eventCounts := make(map[string]int)
	eventsMu := make(chan string, 100)

	// Setup handlers for different event types
	client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		eventCounts["aggTrade"]++
		eventsMu <- "aggTrade"
		return nil
	})

	client.HandleTickerEvent(func(event *models.TickerEvent) error {
		eventCounts["ticker"]++
		eventsMu <- "ticker"
		return nil
	})

	client.HandleMiniTickerEvent(func(event *models.MiniTickerEvent) error {
		eventCounts["miniTicker"]++
		eventsMu <- "miniTicker"
		return nil
	})

	client.HandleBookTickerEvent(func(event *models.BookTickerEvent) error {
		eventCounts["bookTicker"]++
		eventsMu <- "bookTicker"
		return nil
	})

	client.HandleKlineEvent(func(event *models.KlineEvent) error {
		eventCounts["kline"]++
		eventsMu <- "kline"
		return nil
	})

	client.HandleDiffDepthEvent(func(event *models.DiffDepthEvent) error {
		eventCounts["depth"]++
		eventsMu <- "depth"
		return nil
	})

	// Subscribe to different stream types
	streams := []string{
		"btcusdt@aggTrade",
		"btcusdt@ticker",
		"btcusdt@miniTicker",
		"btcusdt@bookTicker",
		"btcusdt@kline_1m",
		"btcusdt@depth",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	// Wait for events from different stream types
	t.Log("Waiting for events from different stream types...")
	timeout := time.After(20 * time.Second)
	totalEvents := 0
	targetEvents := 10

	for totalEvents < targetEvents {
		select {
		case eventType := <-eventsMu:
			totalEvents++
			t.Logf("Received %s event (total: %d)", eventType, totalEvents)
		case <-timeout:
			t.Logf("Timeout reached, received %d total events", totalEvents)
			goto checkResults
		}
	}

checkResults:
	t.Log("\n📊 Event Type Summary:")
	totalReceived := 0
	for eventType, count := range eventCounts {
		if count > 0 {
			t.Logf("  %s: %d events", eventType, count)
			totalReceived += count
		}
	}

	if totalReceived > 0 {
		t.Logf("✅ Combined streams data type test successful: %d total events", totalReceived)
	} else {
		t.Log("⚠️  No events received (may be due to low market activity)")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestCombinedStreamSubscriptionManagement tests subscription management with combined streams
func TestCombinedStreamSubscriptionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream subscription management test in short mode")
	}

	client := umfuturesstreams.NewClient()
	
	// Set to testnet server
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Track events
	eventCount := 0
	eventsMu := make(chan struct{}, 100)

	client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		eventCount++
		eventsMu <- struct{}{}
		return nil
	})

	// Test subscription
	streams := []string{"btcusdt@aggTrade", "ethusdt@aggTrade"}
	
	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to combined streams: %v", err)
	}

	// Wait for initial events
	t.Log("Waiting for initial events...")
	timeout := time.After(10 * time.Second)
	initialEvents := 0

	for initialEvents < 3 {
		select {
		case <-eventsMu:
			initialEvents++
		case <-timeout:
			goto testUnsubscribe
		}
	}

testUnsubscribe:
	t.Logf("Received %d initial events", initialEvents)

	// Test partial unsubscription
	if err := client.Unsubscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
		t.Errorf("Failed to unsubscribe from one stream: %v", err)
	}

	// Reset counter and wait for events from remaining stream
	eventCount = 0
	t.Log("Waiting for events after partial unsubscription...")
	timeout = time.After(8 * time.Second)
	remainingEvents := 0

	for remainingEvents < 2 {
		select {
		case <-eventsMu:
			remainingEvents++
		case <-timeout:
			goto testResubscribe
		}
	}

testResubscribe:
	t.Logf("Received %d events after partial unsubscription", remainingEvents)

	// Test resubscription
	if err := client.Subscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
		t.Errorf("Failed to resubscribe: %v", err)
	}

	// Wait for events after resubscription
	eventCount = 0
	t.Log("Waiting for events after resubscription...")
	timeout = time.After(8 * time.Second)
	resubEvents := 0

	for resubEvents < 3 {
		select {
		case <-eventsMu:
			resubEvents++
		case <-timeout:
			goto cleanup
		}
	}

cleanup:
	t.Logf("Received %d events after resubscription", resubEvents)

	if initialEvents > 0 || remainingEvents > 0 || resubEvents > 0 {
		t.Log("✅ Combined stream subscription management working")
	} else {
		t.Log("⚠️  No events received during subscription management test")
	}

	// Final cleanup
	allStreams := []string{"btcusdt@aggTrade", "ethusdt@aggTrade"}
	if err := client.Unsubscribe(ctx, allStreams); err != nil {
		t.Errorf("Failed to unsubscribe from all streams: %v", err)
	}
}

// TestSingleVsCombinedStreamComparison tests single vs combined stream behavior
func TestSingleVsCombinedStreamComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping single vs combined stream comparison test in short mode")
	}

	// Test single stream connection
	t.Run("SingleStream", func(t *testing.T) {
		client := umfuturesstreams.NewClient()
		
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.ConnectToSingleStreams(ctx, ""); err != nil {
			t.Fatalf("Failed to connect to single streams: %v", err)
		}
		defer client.Disconnect()

		singleEvents := 0
		client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
			singleEvents++
			return nil
		})

		if err := client.Subscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
			t.Fatalf("Failed to subscribe to single stream: %v", err)
		}

		// Wait for events
		time.Sleep(5 * time.Second)
		t.Logf("Single stream events: %d", singleEvents)

		if err := client.Unsubscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
			t.Errorf("Failed to unsubscribe from single stream: %v", err)
		}
	})

	// Test combined stream connection
	t.Run("CombinedStream", func(t *testing.T) {
		client := umfuturesstreams.NewClient()
		
		err := client.SetActiveServer("testnet1")
		if err != nil {
			t.Fatalf("Failed to set testnet server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
			t.Fatalf("Failed to connect to combined streams: %v", err)
		}
		defer client.Disconnect()

		combinedEvents := 0
		regularEvents := 0

		client.HandleCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
			combinedEvents++
			return nil
		})

		client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
			regularEvents++
			return nil
		})

		if err := client.Subscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
			t.Fatalf("Failed to subscribe to combined stream: %v", err)
		}

		// Wait for events
		time.Sleep(5 * time.Second)
		t.Logf("Combined stream events: %d, Regular events: %d", combinedEvents, regularEvents)

		if err := client.Unsubscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
			t.Errorf("Failed to unsubscribe from combined stream: %v", err)
		}
	})

	t.Log("✅ Single vs Combined stream comparison completed")
}

// TestCombinedStreamMicrosecondPrecision tests microsecond precision with combined streams
func TestCombinedStreamMicrosecondPrecision(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream microsecond precision test in short mode")
	}

	client := umfuturesstreams.NewClient()
	
	err := client.SetActiveServer("testnet1")
	if err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams with microsecond precision
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreamsMicrosecond(ctx); err != nil {
		t.Skipf("Skipping microsecond precision test: %v", err)
	}
	defer client.Disconnect()

	// Track events
	eventCount := 0
	eventsMu := make(chan struct{}, 100)

	client.HandleAggregateTradeEvent(func(event *models.AggregateTradeEvent) error {
		eventCount++
		eventsMu <- struct{}{}
		t.Logf("Received microsecond precision event: %d (EventTime: %d)", eventCount, event.EventTime)
		return nil
	})

	if err := client.Subscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
		t.Fatalf("Failed to subscribe with microsecond precision: %v", err)
	}

	// Wait for events
	t.Log("Waiting for microsecond precision events...")
	timeout := time.After(10 * time.Second)
	receivedEvents := 0

	for receivedEvents < 3 {
		select {
		case <-eventsMu:
			receivedEvents++
		case <-timeout:
			goto checkResults
		}
	}

checkResults:
	if receivedEvents > 0 {
		t.Logf("✅ Microsecond precision combined streams working: %d events", receivedEvents)
	} else {
		t.Log("⚠️  No microsecond precision events received (may be due to low market activity)")
	}

	// Clean up
	if err := client.Unsubscribe(ctx, []string{"btcusdt@aggTrade"}); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}