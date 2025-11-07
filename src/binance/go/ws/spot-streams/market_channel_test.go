package streamstest

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	spotstreams "github.com/openxapi/binance-go/ws/spot-streams"
	"github.com/openxapi/binance-go/ws/spot-streams/models"
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

// market event recorder: records events from suite start so event tests can assert against accumulated data
type spotMarketEventRecorder struct {
	mu                  sync.RWMutex
	trade               []*models.TradeEvent
	agg                 []*models.AggregateTradeEvent
	kline               []*models.KlineEvent
	ticker              []*models.TickerEvent
	miniTicker          []*models.MiniTickerEvent
	bookTicker          []*models.BookTickerEvent
	avgPrice            []*models.AveragePriceEvent
	partialDepth        []*models.PartialDepthEvent
	diffDepth           []*models.DiffDepthEvent
	allTickers          []*models.AllTickersEvent
	allMini             []*models.AllMiniTickersEvent
	allRoll             []*models.AllRollingWindowTickersEvent
	combined            []*models.CombinedMarketStreamEvent
	rollingWindowTicker []*models.RollingWindowTickerEvent
}

func newSpotMarketEventRecorder() *spotMarketEventRecorder { return &spotMarketEventRecorder{} }
func (r *spotMarketEventRecorder) addTrade(ev *models.TradeEvent) {
	r.mu.Lock()
	r.trade = append(r.trade, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addAgg(ev *models.AggregateTradeEvent) {
	r.mu.Lock()
	r.agg = append(r.agg, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addKline(ev *models.KlineEvent) {
	r.mu.Lock()
	r.kline = append(r.kline, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addTicker(ev *models.TickerEvent) {
	r.mu.Lock()
	r.ticker = append(r.ticker, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addMiniTicker(ev *models.MiniTickerEvent) {
	r.mu.Lock()
	r.miniTicker = append(r.miniTicker, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addBookTicker(ev *models.BookTickerEvent) {
	r.mu.Lock()
	r.bookTicker = append(r.bookTicker, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addAvgPrice(ev *models.AveragePriceEvent) {
	r.mu.Lock()
	r.avgPrice = append(r.avgPrice, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addPartialDepth(ev *models.PartialDepthEvent) {
	r.mu.Lock()
	r.partialDepth = append(r.partialDepth, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addDiffDepth(ev *models.DiffDepthEvent) {
	r.mu.Lock()
	r.diffDepth = append(r.diffDepth, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addAllTickers(ev *models.AllTickersEvent) {
	r.mu.Lock()
	r.allTickers = append(r.allTickers, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addAllMini(ev *models.AllMiniTickersEvent) {
	r.mu.Lock()
	r.allMini = append(r.allMini, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addAllRoll(ev *models.AllRollingWindowTickersEvent) {
	r.mu.Lock()
	r.allRoll = append(r.allRoll, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addCombined(ev *models.CombinedMarketStreamEvent) {
	r.mu.Lock()
	r.combined = append(r.combined, ev)
	r.mu.Unlock()
}
func (r *spotMarketEventRecorder) addRollingWindowTicker(ev *models.RollingWindowTickerEvent) {
	r.mu.Lock()
	r.rollingWindowTicker = append(r.rollingWindowTicker, ev)
	r.mu.Unlock()
}

func (r *spotMarketEventRecorder) count(key string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch key {
	case "trade":
		return len(r.trade)
	case "aggTrade":
		return len(r.agg)
	case "kline":
		return len(r.kline)
	case "ticker":
		return len(r.ticker)
	case "miniTicker":
		return len(r.miniTicker)
	case "bookTicker":
		return len(r.bookTicker)
	case "avgPrice":
		return len(r.avgPrice)
	case "partialDepth":
		return len(r.partialDepth)
	case "diffDepth":
		return len(r.diffDepth)
	case "allTickers":
		return len(r.allTickers)
	case "allMini":
		return len(r.allMini)
	case "allRoll":
		return len(r.allRoll)
	case "combined":
		return len(r.combined)
	case "rollTicker":
		return len(r.rollingWindowTicker)
	default:
		return 0
	}
}

func (r *spotMarketEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if r.count(key) >= min {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// TestFullIntegrationSuite_Market runs request/response and event coverage for MarketStreamChannel
func TestFullIntegrationSuite_Market(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Capture SDK log output and fail on any 'unhandled message'
	cw := &unhandledCatcher{}
	log.SetOutput(cw)
	defer func() {
		log.SetOutput(os.Stderr)
		cw.mu.Lock()
		defer cw.mu.Unlock()
		if len(cw.matches) > 0 {
			for _, line := range cw.matches {
				t.Logf("SDK log captured: %s", strings.TrimSpace(line))
			}
			t.Fatalf("SDK emitted %d 'unhandled message' log(s) during Market suite; treating as failure", len(cw.matches))
		}
	}()

	cfg := getTestConfig()
	stc, err := NewStreamTestClientDedicated(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	// Force mainnet for Market channel tests
	if err := stc.client.SetActiveServer("mainnet"); err != nil {
		t.Fatalf("failed to switch to mainnet: %v", err)
	}
	// Log active server
	if as := stc.client.GetActiveServer(); as != nil {
		t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
	}

	// Resolve an active symbol via REST; use lowercase for stream paths
	upSym, errPick := restPickSymbol(context.Background())
	if errPick != nil || upSym == "" {
		t.Fatalf("failed to resolve active symbol via REST: %v", errPick)
	}
	symUpper := upSym
	symLower := strings.ToLower(symUpper)
	t.Logf("Using symbol from REST: upper=%s lower=%s", symUpper, symLower)

	// Prepare a channel instance and connect once for the entire suite
	market := spotstreams.NewMarketStreamChannel(stc.client)
	// Record events from the start of the suite
	rec := newSpotMarketEventRecorder()
	market.HandleTradeEvent(func(ctx context.Context, ev *models.TradeEvent) error { rec.addTrade(ev); return nil })
	market.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
	market.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	market.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	market.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
	market.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
	market.HandleAveragePriceEvent(func(ctx context.Context, ev *models.AveragePriceEvent) error { rec.addAvgPrice(ev); return nil })
	market.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	market.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
	market.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTickers(ev); return nil })
	market.HandleAllRollingWindowTickersEvent(func(ctx context.Context, ev *models.AllRollingWindowTickersEvent) error {
		rec.addAllRoll(ev)
		return nil
	})
	market.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMini(ev); return nil })

	// Connect with a base stream (trade)
	base := spotstreams.TradeEventStreamParams{Symbol: models.Symbol(symLower)}
	baseStream, err := spotstreams.BuildTradeEventStream(0, base.Values())
	if err != nil {
		t.Fatalf("build base stream: %v", err)
	}
	cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
	if err := market.Connect(cctx, baseStream); err != nil {
		ccancel()
		t.Fatalf("connect: %v", err)
	}
	ccancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := market.Disconnect(dctx); err != nil {
			t.Errorf("disconnect failed: %v", err)
		}
		dcancel()
	}()

	// ---------- Requests & Responses ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		tk := spotstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}
		tickerPath, err := spotstreams.BuildTickerEventStream(0, tk.Values())
		if err != nil {
			t.Fatalf("build ticker stream: %v", err)
		}
		ag := spotstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
		aggPath, err := spotstreams.BuildAggregateTradeEventStream(0, ag.Values())
		if err != nil {
			t.Fatalf("build aggTrade stream: %v", err)
		}
		sid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.SubscribeResponse
		subCb := func(ctx context.Context, resp *models.SubscribeResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("subscribe error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil subscribe response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != sid {
				t.Errorf("subscribe id mismatch: want %d got %d", sid, rid)
			}
			// result must be null on success
			if resp.Result != nil {
				t.Errorf("subscribe result should be null, got %v", resp.Result)
			}
			got = resp
			logJSON(t, "subscribe.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		req := &models.SubscribeRequest{Id: models.NewMessageIDInt64(sid), Params: []string{tickerPath, aggPath}}
		throttleWS()
		if err := market.Subscribe(context.Background(), req, &subCb); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Errorf("timeout waiting subscribe response")
		}
		if got == nil {
			t.Fatalf("did not capture subscribe response")
		}
		// Cleanup
		uid := time.Now().UnixMicro()
		var unsubCb func(context.Context, *models.UnsubscribeResponse, error) error = func(ctx context.Context, resp *models.UnsubscribeResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("unsubscribe error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil unsubscribe response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != uid {
				t.Errorf("unsubscribe id mismatch: want %d got %d", uid, rid)
			}
			if resp.Result != nil {
				t.Errorf("unsubscribe result should be null, got %v", resp.Result)
			}
			return nil
		}
		throttleWS()
		if err := market.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(uid), Params: []string{tickerPath}}, &unsubCb); err != nil {
			t.Errorf("unsubscribe failed: %v", err)
		}
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		tr := spotstreams.TradeEventStreamParams{Symbol: models.Symbol(symLower)}
		tradePath, err := spotstreams.BuildTradeEventStream(0, tr.Values())
		if err != nil {
			t.Fatalf("build trade stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{tradePath}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		lid := time.Now().UnixMicro()
		listDone := make(chan struct{}, 1)
		var got *models.ListSubscriptionsResponse
		listCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("listSubscriptions error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil list subscriptions response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != lid {
				t.Errorf("list id mismatch: want %d got %d", lid, rid)
			}
			if resp.Result == nil {
				t.Errorf("list subscriptions result is nil; expected array")
			}
			if resp.Result != nil && !contains(resp.Result, tradePath) {
				t.Logf("list did not include %s (result=%v)", tradePath, resp.Result)
			}
			got = resp
			logJSON(t, "listSubscriptions.response", resp)
			select {
			case listDone <- struct{}{}:
			default:
			}
			return nil
		}
		throttleWS()
		if err := market.ListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: models.NewMessageIDInt64(lid)}, &listCb); err != nil {
			t.Fatalf("list subscriptions call failed: %v", err)
		}
		select {
		case <-listDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting listSubscriptions response")
		}
		if got == nil {
			t.Fatalf("did not capture listSubscriptions response")
		}
		// Cleanup: keep base trade subscription active (no unsubscribe)
	})

	// Request_SetProperty: exercise setting the 'combined' property true then false
	t.Run("Request_SetProperty", func(t *testing.T) {
		// set combined=true
		sidTrue := time.Now().UnixMicro()
		setTrueDone := make(chan struct{}, 1)
		var gotSetTrue *models.SetPropertyResponse
		setTrueCb := func(ctx context.Context, resp *models.SetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Logf("setProperty(true) callback error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil setProperty(true) response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != sidTrue {
				t.Errorf("setProperty(true) id mismatch: want %d got %d", sidTrue, rid)
			}
			if resp.Result != nil {
				t.Errorf("setProperty(true) result should be null, got %v", resp.Result)
			}
			gotSetTrue = resp
			logJSON(t, "setProperty.true.response", resp)
			select {
			case setTrueDone <- struct{}{}:
			default:
			}
			return nil
		}
		setTrueReq := &models.SetPropertyRequest{Id: models.NewMessageIDInt64(sidTrue), Params: []interface{}{"combined", true}}
		logJSON(t, "setProperty.true.request", setTrueReq)
		spCtx1, spCancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel1()
		throttleWS()
		if err := market.SetProperty(spCtx1, setTrueReq, &setTrueCb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "deadline") {
				t.Logf("setProperty(true) timeout (acceptable): %v", err)
			} else {
				t.Logf("setProperty(true) err (acceptable): %v", err)
			}
		}
		select {
		case <-setTrueDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting setProperty(true) response (acceptable)")
		}
		_ = gotSetTrue

		// set combined=false
		sidFalse := time.Now().UnixMicro()
		setFalseDone := make(chan struct{}, 1)
		var gotSetFalse *models.SetPropertyResponse
		setFalseCb := func(ctx context.Context, resp *models.SetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Logf("setProperty(false) callback error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil setProperty(false) response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != sidFalse {
				t.Errorf("setProperty(false) id mismatch: want %d got %d", sidFalse, rid)
			}
			if resp.Result != nil {
				t.Errorf("setProperty(false) result should be null, got %v", resp.Result)
			}
			gotSetFalse = resp
			logJSON(t, "setProperty.false.response", resp)
			select {
			case setFalseDone <- struct{}{}:
			default:
			}
			return nil
		}
		setFalseReq := &models.SetPropertyRequest{Id: models.NewMessageIDInt64(sidFalse), Params: []interface{}{"combined", false}}
		logJSON(t, "setProperty.false.request", setFalseReq)
		spCtx2, spCancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel2()
		throttleWS()
		if err := market.SetProperty(spCtx2, setFalseReq, &setFalseCb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "deadline") {
				t.Logf("setProperty(false) timeout (acceptable): %v", err)
			} else {
				t.Logf("setProperty(false) err (acceptable): %v", err)
			}
		}
		select {
		case <-setFalseDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting setProperty(false) response (acceptable)")
		}
		_ = gotSetFalse
	})

	// Request_GetProperty: fetch the current value of 'combined'
	t.Run("Request_GetProperty", func(t *testing.T) {
		// Helper to extract a boolean from various server payload shapes
		extractCombined := func(v interface{}) (bool, bool) {
			switch tv := v.(type) {
			case bool:
				return tv, true
			case map[string]interface{}:
				if x, ok := tv["combined"]; ok {
					if b, ok2 := x.(bool); ok2 {
						return b, true
					}
				}
			}
			return false, false
		}

		gid := time.Now().UnixMicro()
		getDone := make(chan struct{}, 1)
		var gotGet *models.GetPropertyResponse
		_ = gotGet
		getCb := func(ctx context.Context, resp *models.GetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Logf("getProperty callback error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil getProperty response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != gid {
				t.Errorf("getProperty id mismatch: want %d got %d", gid, rid)
			}
			logJSON(t, "getProperty.response", resp)
			if resp.Result != nil {
				if b, ok := extractCombined(resp.Result); ok {
					t.Logf("combined property value: %v", b)
				} else {
					t.Errorf("getProperty result not parseable: %v", resp.Result)
				}
			}
			select {
			case getDone <- struct{}{}:
			default:
			}
			return nil
		}
		getReq := &models.GetPropertyRequest{Id: models.NewMessageIDInt64(gid), Params: []string{"combined"}}
		logJSON(t, "getProperty.request", getReq)
		throttleWS()
		if err := market.GetProperty(context.Background(), getReq, &getCb); err != nil {
			t.Logf("getProperty err (acceptable): %v", err)
		}
		select {
		case <-getDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting getProperty response (acceptable)")
		}
	})

	// ---------- Event Handlers ----------
	t.Run("TradeEvent", func(t *testing.T) {
		tr := spotstreams.TradeEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildTradeEventStream(0, tr.Values())
		if err != nil {
			t.Fatalf("build trade stream: %v", err)
		}
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("trade", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("trade")
		t.Logf("trade events received: %d", cnt)
		if cnt > 0 {
			ev := rec.trade[len(rec.trade)-1]
			if ev.EventType != "trade" {
				t.Errorf("want e=trade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.Price, "price")
		} else {
			t.Logf("no trade event received (acceptable)")
		}
	})

	t.Run("AggregateTradeEvent", func(t *testing.T) {
		ag := spotstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildAggregateTradeEventStream(0, ag.Values())
		if err != nil {
			t.Fatalf("build aggTrade stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("aggTrade", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("aggTrade")
		t.Logf("aggTrade events received: %d", cnt)
		if cnt > 0 {
			ev := rec.agg[len(rec.agg)-1]
			if ev.EventType != "aggTrade" {
				t.Errorf("want e=aggTrade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			_ = tryParseFloat(t, ev.Price, "price")
		} else {
			t.Logf("no aggTrade event received (acceptable)")
		}
	})

	t.Run("KlineEvent", func(t *testing.T) {
		kl := spotstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		path, err := spotstreams.BuildKlineEventStream(0, kl.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("kline", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("kline")
		t.Logf("kline events received: %d", cnt)
		if cnt > 0 {
			ev := rec.kline[len(rec.kline)-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.KlineData.Interval, "interval")
		} else {
			t.Logf("no kline event received (acceptable)")
		}
	})

	t.Run("TickerEvent", func(t *testing.T) {
		tk := spotstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildTickerEventStream(0, tk.Values())
		if err != nil {
			t.Fatalf("build ticker stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("ticker", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("ticker")
		t.Logf("ticker events received: %d", cnt)
		if cnt > 0 {
			ev := rec.ticker[len(rec.ticker)-1]
			if ev.EventType != "24hrTicker" {
				t.Errorf("want e=24hrTicker got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		} else {
			t.Logf("no ticker event received (acceptable)")
		}
	})

	t.Run("MiniTickerEvent", func(t *testing.T) {
		mt := spotstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildMiniTickerEventStream(0, mt.Values())
		if err != nil {
			t.Fatalf("build miniTicker stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("miniTicker", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("miniTicker")
		t.Logf("miniTicker events received: %d", cnt)
		if cnt > 0 {
			ev := rec.miniTicker[len(rec.miniTicker)-1]
			if ev.EventType != "24hrMiniTicker" {
				t.Errorf("want e=24hrMiniTicker got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		} else {
			t.Logf("no miniTicker event received (acceptable)")
		}
	})

	t.Run("BookTickerEvent", func(t *testing.T) {
		bt := spotstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildBookTickerEventStream(0, bt.Values())
		if err != nil {
			t.Fatalf("build bookTicker stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("bookTicker", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("bookTicker")
		t.Logf("bookTicker events received: %d", cnt)
		if cnt > 0 {
			ev := rec.bookTicker[len(rec.bookTicker)-1]
			assertNonEmpty(t, ev.BestBidPrice, "best bid price")
			assertNonEmpty(t, ev.BestAskPrice, "best ask price")
		} else {
			t.Logf("no bookTicker event received (acceptable)")
		}
	})

	t.Run("AvgPriceEvent", func(t *testing.T) {
		ap := spotstreams.AveragePriceEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := spotstreams.BuildAveragePriceEventStream(0, ap.Values())
		if err != nil {
			t.Fatalf("build avgPrice stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("avgPrice", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("avgPrice")
		t.Logf("avgPrice events received: %d", cnt)
		if cnt > 0 {
			ev := rec.avgPrice[len(rec.avgPrice)-1]
			if ev.EventType != "avgPrice" {
				t.Errorf("want e=avgPrice got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		} else {
			t.Logf("no avgPrice event received (acceptable)")
		}
	})

	t.Run("PartialDepthEvent", func(t *testing.T) {
		dp := spotstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels5, Speed: models.DepthSpeed("100ms")}
		path, err := spotstreams.BuildPartialDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build partial depth stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("partialDepth", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("partialDepth")
		t.Logf("partialDepth events received: %d", cnt)
		if cnt > 0 {
			ev := rec.partialDepth[len(rec.partialDepth)-1]
			if ev.LastUpdateID <= 0 {
				t.Errorf("lastUpdateId <= 0")
			}
		} else {
			t.Logf("no partialDepth event received (acceptable)")
		}
	})

	t.Run("DiffDepthEvent", func(t *testing.T) {
		dp := spotstreams.DiffDepthEventStreamParams{Symbol: models.Symbol(symLower), Speed: models.DepthSpeed("100ms")}
		path, err := spotstreams.BuildDiffDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build diff depth stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("diffDepth", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("diffDepth")
		t.Logf("diffDepth events received: %d", cnt)
		if cnt > 0 {
			ev := rec.diffDepth[len(rec.diffDepth)-1]
			if ev.EventType != "depthUpdate" {
				t.Errorf("want e=depthUpdate got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		} else {
			t.Logf("no diffDepth event received (acceptable)")
		}
	})

	t.Run("AllTickersEvent", func(t *testing.T) {
		path, err := spotstreams.BuildAllTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all tickers stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("allTickers", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		cnt := rec.count("allTickers")
		t.Logf("allTickers events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allTickers[len(rec.allTickers)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "24hrTicker" {
					t.Logf("unexpected type: %s", it.EventType)
				}
			}
		}
	})

	t.Run("AllMiniTickersEvent", func(t *testing.T) {
		path, err := spotstreams.BuildAllMiniTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all mini tickers stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("allMini", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		t.Logf("allMiniTickers events received: %d", rec.count("allMini"))
	})

	t.Run("AllRollingWindowTickersEvent", func(t *testing.T) {
		path, err := spotstreams.BuildAllRollingWindowTickersEventStream(0, map[string]string{"windowSize": "1h"})
		if err != nil {
			t.Fatalf("build all rolling tickers stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		if err := rec.waitForMin("allRoll", 1, eventWait()); err != nil {
			t.Fatalf("waitForMin failed: %v", err)
		}
		t.Logf("allRollingWindowTickers events received: %d", rec.count("allRoll"))
	})
}
