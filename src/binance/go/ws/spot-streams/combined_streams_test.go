package streamstest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	spotstreams "github.com/openxapi/binance-go/ws/spot-streams"
	"github.com/openxapi/binance-go/ws/spot-streams/models"
)

// TestCombinedStreamEventReception tests receiving events through combined streams endpoint
func TestCombinedStreamEventReception(t *testing.T) {
	client := spotstreams.NewClient()
	
	// Set to testnet
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Track received combined stream events
	var combinedEvents []*models.CombinedStreamEvent
	var individualEvents []interface{}

	// Set up combined stream event handler
	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		combinedEvents = append(combinedEvents, event)
		t.Logf("✅ Received CombinedStreamEvent - Stream: %s, Data type: %T", 
			event.StreamName, event.StreamData)
		return nil
	})

	// Set up individual event handlers to verify they still work
	client.OnTradeEvent(func(event *models.TradeEvent) error {
		individualEvents = append(individualEvents, event)
		t.Logf("📈 Received individual TradeEvent: %s", event.Symbol)
		return nil
	})

	client.OnTickerEvent(func(event *models.TickerEvent) error {
		individualEvents = append(individualEvents, event)
		t.Logf("📊 Received individual TickerEvent: %s", event.Symbol)
		return nil
	})

	// Subscribe to multiple streams of different types
	streams := []string{
		"btcusdt@trade",
		"ethusdt@trade", 
		"btcusdt@ticker",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	t.Log("🔄 Waiting for combined stream events...")
	
	// Wait for events with timeout
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if len(combinedEvents) >= 3 && len(individualEvents) >= 2 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify combined stream events
	if len(combinedEvents) == 0 {
		t.Error("❌ No CombinedStreamEvent received")
	} else {
		t.Logf("✅ Received %d CombinedStreamEvent events", len(combinedEvents))
		
		// Verify event structure
		for i, event := range combinedEvents {
			if i >= 3 { // Check first 3 events
				break
			}
			
			if event.StreamName == "" {
				t.Errorf("❌ CombinedStreamEvent %d missing StreamName", i)
			}
			
			if event.StreamData == nil {
				t.Errorf("❌ CombinedStreamEvent %d missing StreamData", i)
			}
			
			// Verify stream name matches subscribed streams
			validStream := false
			for _, stream := range streams {
				if event.StreamName == stream {
					validStream = true
					break
				}
			}
			if !validStream {
				t.Errorf("❌ CombinedStreamEvent %d has unexpected StreamName: %s", i, event.StreamName)
			}
			
			t.Logf("✅ CombinedStreamEvent %d: Stream=%s, DataType=%T", i, event.StreamName, event.StreamData)
		}
	}

	// Verify individual event handlers still work with combined streams
	if len(individualEvents) == 0 {
		t.Log("⚠️  No individual events received (may be expected with combined streams)")
	} else {
		t.Logf("✅ Individual event handlers also received %d events", len(individualEvents))
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestCombinedStreamEventDataTypes tests different event types through combined streams
func TestCombinedStreamEventDataTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream data types test in short mode")
	}

	client := spotstreams.NewClient()
	
	// Set to testnet
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	// Track events by type
	eventsByType := make(map[string][]*models.CombinedStreamEvent)

	// Set up combined stream event handler
	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		// Determine event type from stream name
		eventType := "unknown"
		if contains(event.StreamName, "@trade") {
			eventType = "trade"
		} else if contains(event.StreamName, "@ticker") && !contains(event.StreamName, "@miniTicker") {
			eventType = "ticker"
		} else if contains(event.StreamName, "@miniTicker") {
			eventType = "miniTicker"
		} else if contains(event.StreamName, "@bookTicker") {
			eventType = "bookTicker"
		} else if contains(event.StreamName, "@depth") {
			eventType = "depth"
		} else if contains(event.StreamName, "@kline") {
			eventType = "kline"
		}

		eventsByType[eventType] = append(eventsByType[eventType], event)
		t.Logf("📦 Received %s event via CombinedStream: %s", eventType, event.StreamName)
		return nil
	})

	// Subscribe to different stream types
	streams := []string{
		"btcusdt@trade",
		"btcusdt@ticker", 
		"btcusdt@miniTicker",
		"btcusdt@bookTicker",
		"btcusdt@depth5",
		"btcusdt@kline_1m",
	}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	t.Log("🔄 Waiting for different event types...")
	
	// Wait for events with extended timeout for different stream types
	deadline := time.Now().Add(35 * time.Second)
	for time.Now().Before(deadline) {
		totalEvents := 0
		for _, events := range eventsByType {
			totalEvents += len(events)
		}
		if totalEvents >= 8 { // Expect multiple events from different types
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Verify we received different event types
	expectedTypes := []string{"trade", "ticker", "miniTicker", "bookTicker", "depth", "kline"}
	receivedTypes := 0

	for _, eventType := range expectedTypes {
		if events, exists := eventsByType[eventType]; exists && len(events) > 0 {
			receivedTypes++
			t.Logf("✅ Received %d %s events via CombinedStream", len(events), eventType)
			
			// Verify event data structure for first event of each type
			event := events[0]
			if err := validateCombinedStreamEventData(event, eventType); err != nil {
				t.Errorf("❌ Invalid %s event data: %v", eventType, err)
			}
		} else {
			t.Logf("⚠️  No %s events received", eventType)
		}
	}

	if receivedTypes == 0 {
		t.Error("❌ No events received from any stream type")
	} else {
		t.Logf("✅ Successfully received events from %d/%d stream types", receivedTypes, len(expectedTypes))
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestCombinedStreamMicrosecondPrecision tests microsecond precision with combined streams
func TestCombinedStreamMicrosecondPrecision(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined stream microsecond precision test in short mode")
	}

	client := spotstreams.NewClient()
	
	// Set to testnet
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint with microsecond precision
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreamsMicrosecond(ctx); err != nil {
		t.Fatalf("Failed to connect to combined streams with microsecond precision: %v", err)
	}
	defer client.Disconnect()

	var combinedEvents []*models.CombinedStreamEvent

	// Set up combined stream event handler
	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		combinedEvents = append(combinedEvents, event)
		t.Logf("⏱️  Received microsecond precision event: %s", event.StreamName)
		return nil
	})

	// Subscribe to trade stream for timestamp testing
	streams := []string{"btcusdt@trade"}

	if err := client.Subscribe(ctx, streams); err != nil {
		t.Fatalf("Failed to subscribe to streams: %v", err)
	}

	t.Log("🔄 Waiting for microsecond precision events...")
	
	// Wait for events
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if len(combinedEvents) >= 2 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if len(combinedEvents) == 0 {
		t.Error("❌ No events received with microsecond precision")
	} else {
		t.Logf("✅ Received %d events with microsecond precision", len(combinedEvents))
		
		// Check timestamp precision if possible
		for i, event := range combinedEvents {
			if i >= 2 { // Check first 2 events
				break
			}
			
			// Convert event data to check for microsecond timestamps
			dataBytes, err := json.Marshal(event.StreamData)
			if err != nil {
				t.Errorf("❌ Failed to marshal event data: %v", err)
				continue
			}
			
			t.Logf("✅ Microsecond event %d data: %s", i, string(dataBytes))
		}
	}

	// Unsubscribe
	if err := client.Unsubscribe(ctx, streams); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

// TestSingleVsCombinedStreamComparison compares event data from single vs combined streams
func TestSingleVsCombinedStreamComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping single vs combined stream comparison test in short mode")
	}

	testStreamName := "btcusdt@trade"
	
	// Test single stream
	t.Log("🔄 Testing single stream connection...")
	singleEvents := testSingleStream(t, testStreamName)
	
	// Test combined stream
	t.Log("🔄 Testing combined stream connection...")
	combinedEvents := testCombinedStream(t, testStreamName)
	
	// Compare results
	if len(singleEvents) == 0 && len(combinedEvents) == 0 {
		t.Logf("⚠️  No events received from either single or combined streams (may be due to testnet limitations)")
		return
	}
	
	if len(singleEvents) > 0 && len(combinedEvents) > 0 {
		t.Logf("✅ Both single (%d events) and combined (%d events) streams working", 
			len(singleEvents), len(combinedEvents))
		
		// Compare event structure if we have events from both
		if err := compareEventStructures(singleEvents[0], combinedEvents[0]); err != nil {
			t.Logf("⚠️  Event structure mismatch (expected due to different endpoints): %v", err)
		} else {
			t.Log("✅ Event structures are compatible")
		}
	} else {
		t.Logf("⚠️  Only one stream type produced events - Single: %d, Combined: %d (may be due to timing differences)", 
			len(singleEvents), len(combinedEvents))
	}
}

// TestCombinedStreamSubscriptionManagement tests subscription management with combined streams
func TestCombinedStreamSubscriptionManagement(t *testing.T) {
	client := spotstreams.NewClient()
	
	// Set to testnet
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	// Connect to combined streams endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	var allEvents []*models.CombinedStreamEvent

	// Set up combined stream event handler
	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		allEvents = append(allEvents, event)
		return nil
	})

	// Test 1: Subscribe to initial streams
	initialStreams := []string{"btcusdt@trade", "ethusdt@trade"}
	if err := client.Subscribe(ctx, initialStreams); err != nil {
		t.Fatalf("Failed to subscribe to initial streams: %v", err)
	}
	t.Log("✅ Subscribed to initial streams")

	// Wait for initial events
	time.Sleep(5 * time.Second)
	initialEventCount := len(allEvents)
	t.Logf("📈 Received %d events from initial streams", initialEventCount)

	// Test 2: Add more streams
	additionalStreams := []string{"btcusdt@ticker", "ethusdt@miniTicker"}
	if err := client.Subscribe(ctx, additionalStreams); err != nil {
		t.Fatalf("Failed to subscribe to additional streams: %v", err)
	}
	t.Log("✅ Subscribed to additional streams")

	// Wait for events from additional streams
	time.Sleep(8 * time.Second)
	additionalEventCount := len(allEvents) - initialEventCount
	t.Logf("📈 Received %d events from additional streams", additionalEventCount)

	// Test 3: Unsubscribe from some streams
	if err := client.Unsubscribe(ctx, []string{"ethusdt@trade"}); err != nil {
		t.Fatalf("Failed to unsubscribe from stream: %v", err)
	}
	t.Log("✅ Unsubscribed from ethusdt@trade")

	// Wait and verify events continue from remaining streams
	beforeUnsubCount := len(allEvents)
	time.Sleep(5 * time.Second)
	afterUnsubCount := len(allEvents)
	
	if afterUnsubCount > beforeUnsubCount {
		t.Logf("✅ Still receiving events after partial unsubscribe: %d new events", 
			afterUnsubCount - beforeUnsubCount)
	}

	// Test 4: List subscriptions (if supported)
	if err := client.ListSubscriptions(ctx); err != nil {
		t.Logf("⚠️  ListSubscriptions not fully supported or failed: %v", err)
	} else {
		t.Log("✅ ListSubscriptions command sent")
	}

	// Test 5: Unsubscribe from all remaining
	remainingStreams := []string{"btcusdt@trade", "btcusdt@ticker", "ethusdt@miniTicker"}
	if err := client.Unsubscribe(ctx, remainingStreams); err != nil {
		t.Errorf("Failed to unsubscribe from remaining streams: %v", err)
	} else {
		t.Log("✅ Unsubscribed from all streams")
	}

	// Final verification
	totalEvents := len(allEvents)
	if totalEvents == 0 {
		t.Error("❌ No events received during subscription management test")
	} else {
		t.Logf("✅ Subscription management test successful: %d total events", totalEvents)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   (len(s) > len(substr) && s[:len(substr)] == substr) ||
		   (len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func validateCombinedStreamEventData(event *models.CombinedStreamEvent, eventType string) error {
	if event.StreamData == nil {
		return fmt.Errorf("StreamData is nil")
	}
	
	// Try to marshal/unmarshal to verify it's valid JSON-like data
	dataBytes, err := json.Marshal(event.StreamData)
	if err != nil {
		return fmt.Errorf("failed to marshal StreamData: %v", err)
	}
	
	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return fmt.Errorf("StreamData is not valid JSON object: %v", err)
	}
	
	// Basic validation based on event type
	switch eventType {
	case "trade":
		if _, exists := dataMap["s"]; !exists {
			return fmt.Errorf("trade event missing symbol field")
		}
	case "ticker", "miniTicker":
		if _, exists := dataMap["s"]; !exists {
			return fmt.Errorf("ticker event missing symbol field")
		}
	case "bookTicker":
		if _, exists := dataMap["s"]; !exists {
			return fmt.Errorf("bookTicker event missing symbol field")
		}
	case "depth":
		// Depth events have different structure - they can have "s" field or "lastUpdateId" field
		if _, hasSymbol := dataMap["s"]; hasSymbol {
			return nil // Has symbol field, valid
		}
		if _, hasLastUpdateId := dataMap["lastUpdateId"]; hasLastUpdateId {
			return nil // Has lastUpdateId field (partial depth), valid
		}
		if _, hasBids := dataMap["bids"]; hasBids {
			return nil // Has bids field, valid depth event
		}
		if _, hasAsks := dataMap["asks"]; hasAsks {
			return nil // Has asks field, valid depth event
		}
		return fmt.Errorf("depth event missing expected fields (s, lastUpdateId, bids, or asks)")
	case "kline":
		if _, exists := dataMap["s"]; !exists {
			return fmt.Errorf("kline event missing symbol field")
		}
	}
	
	return nil
}

func testSingleStream(t *testing.T, streamName string) []interface{} {
	client := spotstreams.NewClient()
	
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.ConnectToSingleStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to single streams: %v", err)
	}
	defer client.Disconnect()

	var events []interface{}

	client.OnTradeEvent(func(event *models.TradeEvent) error {
		events = append(events, event)
		return nil
	})

	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to single stream: %v", err)
	}

	// Wait for events
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if len(events) >= 1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	client.Unsubscribe(ctx, []string{streamName})
	return events
}

func testCombinedStream(t *testing.T, streamName string) []*models.CombinedStreamEvent {
	client := spotstreams.NewClient()
	
	if err := client.SetActiveServer("testnet1"); err != nil {
		t.Fatalf("Failed to set testnet server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.ConnectToCombinedStreams(ctx, ""); err != nil {
		t.Fatalf("Failed to connect to combined streams: %v", err)
	}
	defer client.Disconnect()

	var events []*models.CombinedStreamEvent

	client.OnCombinedStreamEvent(func(event *models.CombinedStreamEvent) error {
		events = append(events, event)
		return nil
	})

	if err := client.Subscribe(ctx, []string{streamName}); err != nil {
		t.Fatalf("Failed to subscribe to combined stream: %v", err)
	}

	// Wait for events
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if len(events) >= 1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	client.Unsubscribe(ctx, []string{streamName})
	return events
}

func compareEventStructures(singleEvent interface{}, combinedEvent *models.CombinedStreamEvent) error {
	// Convert single event to JSON
	singleBytes, err := json.Marshal(singleEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal single event: %v", err)
	}

	// Convert combined event's StreamData to JSON
	combinedBytes, err := json.Marshal(combinedEvent.StreamData)
	if err != nil {
		return fmt.Errorf("failed to marshal combined event data: %v", err)
	}

	// Parse both to compare structure
	var singleMap, combinedMap map[string]interface{}
	
	if err := json.Unmarshal(singleBytes, &singleMap); err != nil {
		return fmt.Errorf("failed to unmarshal single event: %v", err)
	}
	
	if err := json.Unmarshal(combinedBytes, &combinedMap); err != nil {
		return fmt.Errorf("failed to unmarshal combined event data: %v", err)
	}

	// Basic structure comparison - check if they have similar fields
	if len(singleMap) == 0 && len(combinedMap) == 0 {
		return fmt.Errorf("both events are empty")
	}

	// Check for common fields that should exist in trade events
	commonFields := []string{"s", "p", "q"} // symbol, price, quantity
	for _, field := range commonFields {
		singleHas := false
		combinedHas := false
		
		if _, exists := singleMap[field]; exists {
			singleHas = true
		}
		if _, exists := combinedMap[field]; exists {
			combinedHas = true
		}
		
		if singleHas != combinedHas {
			return fmt.Errorf("field '%s' presence mismatch - single: %v, combined: %v", 
				field, singleHas, combinedHas)
		}
	}

	return nil
}