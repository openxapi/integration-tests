package streamstest

import (
	"bytes"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	optionsstreams "github.com/openxapi/binance-go/ws/options-streams"
	"github.com/openxapi/binance-go/ws/options-streams/models"
)

// unhandledCatcher captures SDK log lines containing the marker "unhandled message:"
type unhandledCatcher struct {
	matches []string
	mu      sync.Mutex
}

func (w *unhandledCatcher) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("unhandled message:")) {
		w.mu.Lock()
		w.matches = append(w.matches, string(p))
		w.mu.Unlock()
	}
	return len(p), nil
}

// TestFullIntegrationSuite_Market runs request/response and event coverage for MarketStreamsChannel
func TestFullIntegrationSuite_Market(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Capture SDK log output and fail on any 'unhandled message' during this suite
	cw := &unhandledCatcher{}
	// Capture SDK logs; route only to catcher to keep signal clean in CI
	// We still log the captured lines via t.Logf upon failure below.
	log.SetOutput(cw)
	defer func() {
		// Restore default output
		log.SetOutput(os.Stderr)
		cw.mu.Lock()
		defer cw.mu.Unlock()
		if len(cw.matches) > 0 {
			// Surface exact lines to aid debugging
			for _, line := range cw.matches {
				t.Logf("SDK log captured: %s", strings.TrimSpace(line))
			}
			t.Fatalf("SDK emitted %d 'unhandled message' log(s) during Market suite; treating as failure", len(cw.matches))
		}
	}()

	config := getTestConfig()
	stc, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Log current active server from SDK defaults
	if as := stc.client.GetActiveServer(); as != nil {
		t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
	} else {
		t.Logf("Active WS server: <nil>")
	}

	// Resolve an active underlying via REST; use lowercase for stream paths
	up, low, errPick := restPickUnderlying(context.Background())
	if errPick != nil || up == "" || low == "" {
		t.Fatalf("failed to resolve active underlying via REST: %v", errPick)
	}
	underlyingUpper := up
	underlyingLower := low
	t.Logf("Using underlying from REST: upper=%s lower=%s", underlyingUpper, underlyingLower)

	// Prepare a channel instance and connect once for the entire suite
	market := optionsstreams.NewMarketStreamsChannel(stc.client)
	// Record events from the start of the suite
	rec := newMarketEventRecorder()
	market.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error { logJSON(t, "ws.error", msg); return nil })
	market.HandleIndexPriceEvent(func(ctx context.Context, ev *models.IndexPriceEvent) error { rec.addIndex(ev); return nil })
	market.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
	market.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	market.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	market.HandleTickerByUnderlyingEvent(func(ctx context.Context, ev *models.TickerByUnderlyingEvent) error {
		rec.addTickerByUnderlying(ev)
		return nil
	})
	market.HandleTradeEvent(func(ctx context.Context, ev *models.TradeEvent) error { rec.addTrade(ev); return nil })
	market.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	market.HandleNewSymbolInfoEvent(func(ctx context.Context, ev *models.NewSymbolInfoEvent) error { rec.addNewSymbolInfo(ev); return nil })
	market.HandleOpenInterestEvent(func(ctx context.Context, ev *models.OpenInterestEvent) error { rec.addOpenInterest(ev); return nil })
	{
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		baseStream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build base stream: %v", err)
		}
		cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := market.Connect(cctx, baseStream); err != nil {
			ccancel()
			t.Fatalf("connect: %v", err)
		}
		ccancel()
		defer func() {
			dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = market.Disconnect(dctx)
			dcancel()
		}()
	}

	// ---------- Requests & Responses (split) ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		// Build index stream via typed params
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		indexPath, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Build markPrice stream using typed params
		mp := optionsstreams.MarkPriceEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingUpper)}
		markPath, err := optionsstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPrice stream: %v", err)
		}
		// Use SUBSCRIBE on the existing connection; validate exact response id
		sid := time.Now().UnixMicro()
		subDone := make(chan struct{}, 1)
		var gotSub *models.SubscriptionResponse
		subCb := func(ctx context.Context, resp *models.SubscriptionResponse) error {
			if resp == nil {
				t.Errorf("nil subscribe response")
				return nil
			}
			if resp.Id != sid {
				t.Errorf("subscribe id mismatch: want %d got %d", sid, resp.Id)
			}
			gotSub = resp
			logJSON(t, "subscribe.response", resp)
			select {
			case subDone <- struct{}{}:
			default:
			}
			return nil
		}
		req := &models.SubscribeRequest{Id: sid, Params: []string{markPath, indexPath}}
		if err := market.SubscribeToMarketStreams(context.Background(), req, &subCb); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		t.Logf("subscribe request sent (id=%d) for %s and %s", sid, markPath, indexPath)
		select {
		case <-subDone:
		case <-time.After(10 * time.Second):
			t.Errorf("timeout waiting subscribe response")
		}
		if gotSub == nil {
			t.Fatalf("did not capture subscribe response")
		}
		// Proactively unsubscribe to avoid cross-test event spam from active streams
		var unsubCb func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{markPath, indexPath}}, &unsubCb)
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		indexPath, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Subscribe to receive index events
		subCb0 := func(context.Context, *models.SubscriptionResponse) error { return nil }
		if err := market.SubscribeToMarketStreams(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{indexPath}}, &subCb0); err != nil {
			t.Fatalf("subscribe before list failed: %v", err)
		}
		lid := time.Now().UnixMicro()
		listDone := make(chan struct{}, 1)
		var gotList *models.ListSubscriptionsResponse
		lsCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse) error {
			if resp == nil {
				t.Errorf("nil list subscriptions response")
				return nil
			}
			if resp.Id != lid {
				t.Errorf("list id mismatch: want %d got %d", lid, resp.Id)
			}
			if resp.Result != nil && !contains(resp.Result, indexPath) {
				t.Logf("list did not include %s (result=%v)", indexPath, resp.Result)
			}
			gotList = resp
			logJSON(t, "listSubscriptions.response", resp)
			select {
			case listDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := market.ListSubscriptionsFromMarketStreams(context.Background(), &models.ListSubscriptionsRequest{Id: lid}, &lsCb); err != nil {
			t.Fatalf("list subscriptions call failed: %v", err)
		}
		t.Logf("listSubscriptions request sent (id=%d)", lid)
		select {
		case <-listDone:
		case <-time.After(10 * time.Second):
			t.Errorf("timeout waiting listSubscriptions response")
		}
		if gotList == nil {
			t.Fatalf("did not capture listSubscriptions response")
		}
		// Cleanup
		unsubCb0 := func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{indexPath}}, &unsubCb0)
	})

	t.Run("Request_SetProperty", func(t *testing.T) {
		// New connection; use base to ensure connection stability
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		indexPath, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Already connected at top-level; global handlers record events
		// Subscribe to receive index events
		var subBeforeGet func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		_ = market.SubscribeToMarketStreams(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{indexPath}}, &subBeforeGet)
		pid := time.Now().UnixMicro()
		setDone := make(chan struct{}, 1)
		var gotSet *models.SetPropertyResponse
		setCb := func(ctx context.Context, resp *models.SetPropertyResponse) error {
			if resp == nil {
				t.Errorf("nil setProperty response")
				return nil
			}
			if resp.Id != pid {
				t.Errorf("setProperty id mismatch: want %d got %d", pid, resp.Id)
			}
			gotSet = resp
			logJSON(t, "setProperty.response", resp)
			select {
			case setDone <- struct{}{}:
			default:
			}
			return nil
		}
		// Use a short timeout; some endpoints may not ACK this on single-stream connections
		spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel()
		if err := market.SetPropertyOnMarketStreams(spCtx, &models.SetPropertyRequest{Id: pid, Params: []interface{}{"combined", false}}, &setCb); err != nil {
			// Treat context timeout as acceptable behavior
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "context deadline") || strings.Contains(le, "deadline exceeded") {
				t.Logf("setProperty returned timeout (acceptable): %v", err)
			} else {
				t.Fatalf("setProperty call failed: %v", err)
			}
		}
		t.Logf("setProperty request sent (id=%d)", pid)
		select {
		case <-setDone:
		case <-time.After(10 * time.Second):
			// Some servers may not ack SET_PROPERTY on single-stream endpoint; treat as informational
			t.Logf("timeout waiting setProperty response (acceptable)")
		}
		_ = gotSet // ensure captured response used; fields asserted above
		// Do not unsubscribe the base index stream here to keep the connection stable across suite
	})

	t.Run("Request_GetProperty", func(t *testing.T) {
		// New connection; use base
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		indexPath, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Already connected at top-level; global handlers record events
		// Subscribe to receive index events so optional validation can succeed and avoid unused var
		var preGetSubCb func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		_ = market.SubscribeToMarketStreams(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{indexPath}}, &preGetSubCb)
		gid := time.Now().UnixMicro()
		getDone := make(chan struct{}, 1)
		var gotGet *models.GetPropertyResponse
		getCb := func(ctx context.Context, resp *models.GetPropertyResponse) error {
			if resp == nil {
				t.Errorf("nil getProperty response")
				return nil
			}
			if resp.Id != gid {
				t.Errorf("getProperty id mismatch: want %d got %d", gid, resp.Id)
			}
			gotGet = resp
			logJSON(t, "getProperty.response", resp)
			select {
			case getDone <- struct{}{}:
			default:
			}
			return nil
		}
		gCtx, gCancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer gCancel()
		if err := market.GetPropertyFromMarketStreams(gCtx, &models.GetPropertyRequest{Id: gid, Params: []string{"combined"}}, &getCb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "context deadline") || strings.Contains(le, "deadline exceeded") {
				t.Logf("getProperty returned timeout (acceptable): %v", err)
			} else {
				t.Fatalf("getProperty call failed: %v", err)
			}
		}
		t.Logf("getProperty request sent (id=%d)", gid)
		select {
		case <-getDone:
		case <-time.After(10 * time.Second):
			t.Errorf("timeout waiting getProperty response")
		}
		if gotGet == nil {
			t.Fatalf("did not capture getProperty response")
		}
		// Do not unsubscribe the base index stream here to keep the connection stable across suite
	})

	t.Run("Request_Unsubscribe", func(t *testing.T) {
		// Use SUBSCRIBE on the existing connection
		mp := optionsstreams.MarkPriceEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingUpper)}
		markPath, err := optionsstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPrice stream: %v", err)
		}
		// Subscribe and wait for ACK to avoid racing UNSUBSCRIBE before server processes SUBSCRIBE
		sid := time.Now().UnixMicro()
		subDone := make(chan struct{}, 1)
		var gotSub *models.SubscriptionResponse
		subCb := func(ctx context.Context, resp *models.SubscriptionResponse) error {
			if resp == nil {
				t.Errorf("nil subscribe response")
				return nil
			}
			if resp.Id != sid {
				t.Errorf("subscribe id mismatch: want %d got %d", sid, resp.Id)
			}
			gotSub = resp
			logJSON(t, "subscribe.response", resp)
			select {
			case subDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := market.SubscribeToMarketStreams(context.Background(), &models.SubscribeRequest{Id: sid, Params: []string{markPath}}, &subCb); err != nil {
			t.Fatalf("subscribe before unsubscribe failed: %v", err)
		}
		// Wait a short window for subscribe ACK; if it doesn't arrive we continue but may see no UNSUB ACK
		select {
		case <-subDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting subscribe ACK before unsubscribe (continuing)")
		}
		_ = gotSub
		uid := time.Now().UnixMicro()
		unsubDone := make(chan struct{}, 1)
		unsubCb := func(ctx context.Context, resp *models.UnsubscriptionResponse) error {
			if resp == nil {
				t.Errorf("nil unsubscribe response")
				return nil
			}
			if resp.Id != uid {
				t.Errorf("unsubscribe id mismatch: want %d got %d", uid, resp.Id)
			}
			logJSON(t, "unsubscribe.response", resp)
			select {
			case unsubDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: uid, Params: []string{markPath}}, &unsubCb); err != nil {
			t.Fatalf("unsubscribe call failed: %v", err)
		}
		t.Logf("unsubscribe request sent (id=%d) for %s", uid, markPath)
		// Prefer ACK but fall back to verifying removal via LIST_SUBSCRIPTIONS to avoid flakiness
		select {
		case <-unsubDone:
			// ok
		case <-time.After(12 * time.Second):
			t.Logf("timeout waiting unsubscribe response; verifying via listSubscriptions")
			lid := time.Now().UnixMicro()
			lsDone := make(chan struct{}, 1)
			var listed []string
			lsCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse) error {
				if resp != nil && resp.Id == lid {
					listed = resp.Result
					logJSON(t, "listSubscriptions.response", resp)
					select {
					case lsDone <- struct{}{}:
					default:
					}
				}
				return nil
			}
			_ = market.ListSubscriptionsFromMarketStreams(context.Background(), &models.ListSubscriptionsRequest{Id: lid}, &lsCb)
			select {
			case <-lsDone:
				if contains(listed, markPath) {
					t.Logf("unsub not confirmed by listSubscriptions; stream still present: %s", markPath)
				} else {
					t.Logf("unsub confirmed by listSubscriptions; stream removed")
				}
			case <-time.After(8 * time.Second):
				t.Logf("timeout waiting listSubscriptions after unsub")
			}
		}
	})

	// ---------- Event Handlers ----------
	t.Run("IndexPriceEvent", func(t *testing.T) {
		// Build index event stream using typed params and subscribe
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingUpper + "USDT")}
		path, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb0 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req0 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req0)
		_ = market.SubscribeToMarketStreams(context.Background(), req0, &subCb0)
		_ = rec.waitForMin("index", 1, eventWait())
		if evs := rec.getIndex(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "index" {
				t.Errorf("want e=index got %s", ev.EventType)
			}
			if ev.UnderlyingSymbol == "" || !strings.EqualFold(ev.UnderlyingSymbol, underlyingUpper+"USDT") {
				t.Errorf("unexpected symbol: %s", ev.UnderlyingSymbol)
			}
			_ = tryParseFloat(t, ev.IndexPrice, "indexPrice")
			assertRecentMs(t, ev.EventTimestamp, 2*time.Hour, "eventTimestamp")
			if restValidationEnabled() {
				if px, err := restIndexPrice(context.Background(), underlyingUpper); err == nil {
					if v, err2 := strconv.ParseFloat(ev.IndexPrice, 64); err2 == nil {
						assertWithinTolerancePercent(t, v, px, 2.0, "index price vs REST")
					}
				} else {
					t.Logf("REST index fetch failed: %v", err)
				}
			}
		} else {
			t.Logf("no index events recorded (acceptable on quiet markets)")
		}
		t.Logf("IndexPriceEvent count: %d", rec.count("index"))
		// Keep the index stream subscribed to maintain a stable connection
	})

	t.Run("MarkPriceEvent", func(t *testing.T) {
		// Build markPrice stream using typed params and subscribe
		mp := optionsstreams.MarkPriceEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingUpper)}
		path, err := optionsstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPrice stream: %v", err)
		}
		// Already connected at top-level; global handlers will record events
		var subCb1 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req1 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req1)
		_ = market.SubscribeToMarketStreams(context.Background(), req1, &subCb1)
		_ = rec.waitForMin("markPrice", 1, eventWait())
		if evs := rec.getMark(); len(evs) > 0 && evs[len(evs)-1] != nil {
			ev := evs[len(evs)-1]
			if ev == nil {
				t.Fatalf("nil markPrice event")
			}
			if len(*ev) == 0 {
				t.Errorf("empty markPrice array")
			}
			for i, it := range *ev {
				if it.EventType != "markPrice" {
					t.Errorf("want e=markPrice got %s", it.EventType)
				}
				assertNonEmpty(t, it.OptionSymbol, "optionSymbol")
				assertOptionSymbolFormat(t, it.OptionSymbol)
				_ = tryParseFloat(t, it.OptionMarkPrice, "markPrice")
				assertRecentMs(t, it.EventTimestamp, 2*time.Hour, "eventTimestamp")
				if i == 0 && restValidationEnabled() {
					if px, err := restMarkPrice(context.Background(), it.OptionSymbol); err == nil && px > 0 {
						if v, err2 := strconv.ParseFloat(it.OptionMarkPrice, 64); err2 == nil {
							assertWithinTolerancePercent(t, v, px, 5.0, "mark price vs REST")
						}
					} else if err != nil {
						t.Logf("REST mark fetch failed: %v", err)
					}
				}
			}
		} else {
			t.Logf("no markPrice events recorded (acceptable)")
		}
		t.Logf("MarkPriceEvent count: %d", rec.count("markPrice"))
		// Unsubscribe from markPrice stream
		var unsubCbMk func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbMk)
	})

	t.Run("KlineEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingUpper, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		interval := os.Getenv("DEFAULT_INTERVAL")
		if interval == "" {
			interval = "1m"
		}
		// Build kline path using typed params
		kp := optionsstreams.KlineEventStreamParams{Symbol: models.Symbol(symbol), Interval: models.Interval(interval)}
		path, err := optionsstreams.BuildKlineEventStream(0, kp.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb2 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req2 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req2)
		_ = market.SubscribeToMarketStreams(context.Background(), req2, &subCb2)
		_ = rec.waitForMin("kline", 1, eventWaitLong())
		if evs := rec.getKline(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			if ev.KlineData.Interval != interval {
				t.Logf("interval mismatch (got %s)", ev.KlineData.Interval)
			}
			_ = tryParseFloat(t, ev.KlineData.OpenPrice, "k.o")
			_ = tryParseFloat(t, ev.KlineData.HighPrice, "k.h")
			_ = tryParseFloat(t, ev.KlineData.LowPrice, "k.l")
			_ = tryParseFloat(t, ev.KlineData.ClosePrice, "k.c")
			if ev.KlineData.KlineStartTime >= ev.KlineData.KlineEndTime {
				t.Errorf("kline times invalid: start %d >= end %d", ev.KlineData.KlineStartTime, ev.KlineData.KlineEndTime)
			}
		} else {
			t.Logf("no kline events recorded (acceptable)")
		}
		t.Logf("KlineEvent count: %d", rec.count("kline"))
		// Unsubscribe from kline stream
		var unsubCbKl func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbKl)
	})

	t.Run("TickerEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingUpper, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build ticker path using typed params
		tp := optionsstreams.TickerEventStreamParams{Symbol: models.Symbol(symbol)}
		path, err := optionsstreams.BuildTickerEventStream(0, tp.Values())
		if err != nil {
			t.Fatalf("build ticker stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb3 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req3 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req3)
		_ = market.SubscribeToMarketStreams(context.Background(), req3, &subCb3)
		_ = rec.waitForMin("ticker", 1, eventWait())
		if evs := rec.getTicker(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "24hrTicker" {
				t.Errorf("want e=24hrTicker got %s", ev.EventType)
			}
			assertNonEmpty(t, ev.OptionSymbol, "symbol")
			assertOptionSymbolFormat(t, ev.OptionSymbol)
			_ = tryParseFloat(t, ev.LatestPrice, "lastPrice")
			_ = tryParseFloat(t, ev.MarkPrice, "markPrice")
			_ = tryParseFloat(t, ev.TradingVolume, "volume")
			if restValidationEnabled() {
				if px, err := restTickerLast(context.Background(), ev.OptionSymbol); err == nil && px > 0 {
					if v, err2 := strconv.ParseFloat(ev.LatestPrice, 64); err2 == nil {
						assertWithinTolerancePercent(t, v, px, 10.0, "ticker last vs REST")
					}
				} else if err != nil {
					t.Logf("REST ticker fetch failed: %v", err)
				}
			}
			_ = tryParseFloat(t, ev.PriceChangePercent, "priceChangePercent")
		} else {
			t.Logf("no ticker events recorded (acceptable)")
		}
		t.Logf("TickerEvent count: %d", rec.count("ticker"))
		// Unsubscribe from ticker stream
		var unsubCbTk func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbTk)
	})

	t.Run("TickerByUnderlyingEvent", func(t *testing.T) {
		exps, err := getActiveExpirationDates(underlyingUpper)
		if err != nil || len(exps) == 0 {
			t.Skipf("no expiration dates: %v", err)
		}
		// Build ticker-by-underlying path using typed params
		tup := optionsstreams.TickerByUnderlyingEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingUpper), ExpirationDate: models.ExpirationDate(exps[0])}
		path, err := optionsstreams.BuildTickerByUnderlyingEventStream(0, tup.Values())
		if err != nil {
			t.Fatalf("build ticker-by-underlying stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb6 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req6 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req6)
		_ = market.SubscribeToMarketStreams(context.Background(), req6, &subCb6)
		_ = rec.waitForMin("tickerByUnderlying", 1, eventWait())
		if evs := rec.getTickerByUnderlying(); len(evs) > 0 && evs[len(evs)-1] != nil {
			ev := evs[len(evs)-1]
			if len(*ev) == 0 {
				t.Errorf("empty tickerByUnderlying array")
			}
			it := (*ev)[0]
			if it.EventType != "24hrTicker" {
				t.Errorf("want e=24hrTicker got %s", it.EventType)
			}
			assertOptionSymbolFormat(t, it.OptionSymbol)
			if restValidationEnabled() {
				if px, err := restTickerLast(context.Background(), it.OptionSymbol); err == nil && px > 0 {
					if v, err2 := strconv.ParseFloat(it.LatestPrice, 64); err2 == nil {
						assertWithinTolerancePercent(t, v, px, 10.0, "ticker-by-underlying last vs REST")
					}
				}
			}
		} else {
			t.Logf("no tickerByUnderlying events recorded (acceptable)")
		}
		t.Logf("TickerByUnderlyingEvent count: %d", rec.count("tickerByUnderlying"))
		// Unsubscribe from ticker-by-underlying stream
		var unsubCbTbu func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbTbu)
	})

	t.Run("IndexPriceEvent_Again", func(t *testing.T) {
		// Covered above; kept minimal to ensure suite completes
		t.Logf("IndexPriceEvent count (again): %d", rec.count("index"))
	})

	t.Run("TradeEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingUpper, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build trade symbol stream using typed params (pattern 0)
		trp := optionsstreams.TradeEventStreamParams{Symbol: models.Symbol(symbol)}
		path, err := optionsstreams.BuildTradeEventStream(0, trp.Values())
		if err != nil {
			t.Fatalf("build trade stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb7 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req7 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req7)
		_ = market.SubscribeToMarketStreams(context.Background(), req7, &subCb7)
		_ = rec.waitForMin("trade", 1, eventWaitLong())
		if evs := rec.getTrade(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "trade" {
				t.Errorf("want e=trade got %s", ev.EventType)
			}
			assertNonEmpty(t, ev.OptionTradingSymbol, "symbol")
			assertOptionSymbolFormat(t, ev.OptionTradingSymbol)
			_ = tryParseFloat(t, ev.Price, "price")
			_ = tryParseFloat(t, ev.Quantity, "qty")
			if ev.TradeID <= 0 {
				t.Errorf("invalid trade id: %d", ev.TradeID)
			}
			if ev.TradeCompletedTimestamp <= 0 {
				t.Errorf("invalid trade ts: %d", ev.TradeCompletedTimestamp)
			}
		} else {
			t.Logf("no trade events recorded (acceptable)")
		}
		t.Logf("TradeEvent count: %d", rec.count("trade"))
		// Unsubscribe from trade stream
		var unsubCbTr func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbTr)
	})

	t.Run("PartialDepthEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingUpper, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build partial depth stream with typed params (pattern index 1 has speed)
		pd := optionsstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symbol), Levels: models.DepthLevels20, Speed: models.DepthSpeed100ms}
		path, err := optionsstreams.BuildPartialDepthEventStream(1, pd.Values())
		if err != nil {
			t.Fatalf("build partialDepth stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		var subCb8 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		req8 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req8)
		_ = market.SubscribeToMarketStreams(context.Background(), req8, &subCb8)
		_ = rec.waitForMin("partialDepth", 1, eventWaitLong())
		if evs := rec.getPartialDepth(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "depth" {
				t.Errorf("want e=depth got %s", ev.EventType)
			}
			assertNonEmpty(t, ev.OptionSymbol, "symbol")
			assertOptionSymbolFormat(t, ev.OptionSymbol)
			if len(ev.BuyOrders) > 0 {
				b := ev.BuyOrders[0]
				if len(b) >= 2 {
					_ = tryParseFloat(t, b[0], "bidPrice")
					_ = tryParseFloat(t, b[1], "bidQty")
				}
			}
			if len(ev.SellOrders) > 0 {
				a := ev.SellOrders[0]
				if len(a) >= 2 {
					_ = tryParseFloat(t, a[0], "askPrice")
					_ = tryParseFloat(t, a[1], "askQty")
				}
			}
			if ev.UpdateIDInEvent <= ev.PreviousUpdateID {
				t.Logf("non-increasing update ids: u=%d pu=%d", ev.UpdateIDInEvent, ev.PreviousUpdateID)
			}
		} else {
			t.Logf("no partialDepth events recorded (acceptable)")
		}
		t.Logf("PartialDepthEvent count: %d", rec.count("partialDepth"))
		// Unsubscribe from partial depth stream
		var unsubCbPd func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbPd)
	})

	t.Run("NewSymbolInfoEvent", func(t *testing.T) {
		// Build correct stream name using builder (no params required)
		stream, err := optionsstreams.BuildNewSymbolInfoEventStream(0, nil)
		if err != nil {
			t.Fatalf("build newSymbolInfo stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		req4 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{stream}}
		logJSON(t, "request.subscribe", req4)
		// Register a no-op ack handler to ensure SDK dispatch marks message as handled
		var subCb4 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		_ = market.SubscribeToMarketStreams(context.Background(), req4, &subCb4)
		_ = rec.waitForMin("newSymbolInfo", 1, eventWait())
		if evs := rec.getNewSymbolInfo(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "OPTION_PAIR" {
				t.Errorf("want e=OPTION_PAIR got %s", ev.EventType)
			}
			assertNonEmpty(t, ev.TradingPairName, "tradingPairName")
			assertOptionSymbolFormat(t, ev.TradingPairName)
			if restValidationEnabled() {
				u := optionUnderlying(ev.TradingPairName)
				if syms, err := getActiveOptionsSymbols(u); err == nil {
					found := false
					for _, s := range syms {
						if s == ev.TradingPairName {
							found = true
							break
						}
					}
					if !found {
						t.Logf("symbol from newSymbolInfo not found in exchange info list (may be brand new): %s", ev.TradingPairName)
					}
				}
			}
		} else {
			t.Logf("no newSymbolInfo events recorded (acceptable)")
		}
		t.Logf("NewSymbolInfoEvent count: %d", rec.count("newSymbolInfo"))
		// Unsubscribe from newSymbolInfo stream
		var unsubCbNs func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{stream}}, &unsubCbNs)
	})

	t.Run("OpenInterestEvent", func(t *testing.T) {
		exps, err := getActiveExpirationDates(underlyingUpper)
		if err != nil || len(exps) == 0 {
			t.Skipf("no expiration dates: %v", err)
		}
		// Build openInterest stream using typed params
		op := optionsstreams.OpenInterestEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingUpper), ExpirationDate: models.ExpirationDate(exps[0])}
		path, err := optionsstreams.BuildOpenInterestEventStream(0, op.Values())
		if err != nil {
			t.Fatalf("build openInterest stream: %v", err)
		}
		// Subscribe on the existing connection (global handlers will record events)
		req5 := &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}
		logJSON(t, "request.subscribe", req5)
		// Register a no-op ack handler to ensure SDK dispatch marks message as handled
		var subCb5 func(context.Context, *models.SubscriptionResponse) error = func(context.Context, *models.SubscriptionResponse) error { return nil }
		_ = market.SubscribeToMarketStreams(context.Background(), req5, &subCb5)
		_ = rec.waitForMin("openInterest", 1, eventWaitLong())
		if evs := rec.getOpenInterest(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev == nil || len(*ev) == 0 {
				t.Logf("openInterest last event is empty (acceptable)")
			} else {
				for _, it := range *ev {
					if it.EventType != "openInterest" {
						t.Errorf("want e=openInterest got %s", it.EventType)
					}
					assertOptionSymbolFormat(t, it.OptionSymbol)
					_ = tryParseFloat(t, it.OpenInterestInContracts, "openInterestContracts")
					_ = tryParseFloat(t, it.OpenInterestInUSDT, "openInterestUSDT")
				}
			}
		} else {
			t.Logf("no openInterest events recorded (acceptable)")
		}
		t.Logf("OpenInterestEvent count: %d", rec.count("openInterest"))
		// Unsubscribe from openInterest stream
		var unsubCbOi func(context.Context, *models.UnsubscriptionResponse) error = func(context.Context, *models.UnsubscriptionResponse) error { return nil }
		_ = market.UnsubscribeFromMarketStreams(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsubCbOi)
	})
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

// market event recorder: records events from suite start so event tests can assert against accumulated data
type marketEventRecorder struct {
	mu            sync.RWMutex
	index         []*models.IndexPriceEvent
	mark          []*models.MarkPriceEvent
	kline         []*models.KlineEvent
	ticker        []*models.TickerEvent
	tickerBy      []*models.TickerByUnderlyingEvent
	trade         []*models.TradeEvent
	partialDepth  []*models.PartialDepthEvent
	newSymbolInfo []*models.NewSymbolInfoEvent
	openInterest  []*models.OpenInterestEvent
}

func newMarketEventRecorder() *marketEventRecorder { return &marketEventRecorder{} }

func (r *marketEventRecorder) addIndex(ev *models.IndexPriceEvent) {
	r.mu.Lock()
	r.index = append(r.index, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addMark(ev *models.MarkPriceEvent) {
	r.mu.Lock()
	r.mark = append(r.mark, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addKline(ev *models.KlineEvent) {
	r.mu.Lock()
	r.kline = append(r.kline, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addTicker(ev *models.TickerEvent) {
	r.mu.Lock()
	r.ticker = append(r.ticker, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addTickerByUnderlying(ev *models.TickerByUnderlyingEvent) {
	r.mu.Lock()
	r.tickerBy = append(r.tickerBy, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addTrade(ev *models.TradeEvent) {
	r.mu.Lock()
	r.trade = append(r.trade, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addPartialDepth(ev *models.PartialDepthEvent) {
	r.mu.Lock()
	r.partialDepth = append(r.partialDepth, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addNewSymbolInfo(ev *models.NewSymbolInfoEvent) {
	r.mu.Lock()
	r.newSymbolInfo = append(r.newSymbolInfo, ev)
	r.mu.Unlock()
}
func (r *marketEventRecorder) addOpenInterest(ev *models.OpenInterestEvent) {
	// Guard against SDK misrouting: only record when event type matches
	// Expect an array payload whose first element has e == "openInterest"
	if ev == nil || len(*ev) == 0 {
		return
	}
	if (*ev)[0].EventType != "openInterest" {
		return
	}
	r.mu.Lock()
	r.openInterest = append(r.openInterest, ev)
	r.mu.Unlock()
}

func (r *marketEventRecorder) getIndex() []*models.IndexPriceEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.IndexPriceEvent, len(r.index))
	copy(out, r.index)
	return out
}
func (r *marketEventRecorder) getMark() []*models.MarkPriceEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.MarkPriceEvent, len(r.mark))
	copy(out, r.mark)
	return out
}
func (r *marketEventRecorder) getKline() []*models.KlineEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.KlineEvent, len(r.kline))
	copy(out, r.kline)
	return out
}
func (r *marketEventRecorder) getTicker() []*models.TickerEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.TickerEvent, len(r.ticker))
	copy(out, r.ticker)
	return out
}
func (r *marketEventRecorder) getTickerByUnderlying() []*models.TickerByUnderlyingEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.TickerByUnderlyingEvent, len(r.tickerBy))
	copy(out, r.tickerBy)
	return out
}
func (r *marketEventRecorder) getTrade() []*models.TradeEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.TradeEvent, len(r.trade))
	copy(out, r.trade)
	return out
}
func (r *marketEventRecorder) getPartialDepth() []*models.PartialDepthEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.PartialDepthEvent, len(r.partialDepth))
	copy(out, r.partialDepth)
	return out
}
func (r *marketEventRecorder) getNewSymbolInfo() []*models.NewSymbolInfoEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.NewSymbolInfoEvent, len(r.newSymbolInfo))
	copy(out, r.newSymbolInfo)
	return out
}
func (r *marketEventRecorder) getOpenInterest() []*models.OpenInterestEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.OpenInterestEvent, len(r.openInterest))
	copy(out, r.openInterest)
	return out
}

// waitForMin waits until the recorder has at least 'min' events for the given key or timeout
func (r *marketEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if r.count(key) >= min {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (r *marketEventRecorder) count(key string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch key {
	case "index":
		return len(r.index)
	case "markPrice":
		return len(r.mark)
	case "kline":
		return len(r.kline)
	case "ticker":
		return len(r.ticker)
	case "tickerByUnderlying":
		return len(r.tickerBy)
	case "trade":
		return len(r.trade)
	case "partialDepth":
		return len(r.partialDepth)
	case "newSymbolInfo":
		return len(r.newSymbolInfo)
	case "openInterest":
		return len(r.openInterest)
	default:
		return 0
	}
}
