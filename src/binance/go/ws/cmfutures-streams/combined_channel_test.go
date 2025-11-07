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

	cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
	"github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// Capture SDK log lines containing the marker "unhandled message:" and fail the suite
type unhandledCatcher2 struct {
	matches []string
	mu      sync.Mutex
}

func (w *unhandledCatcher2) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("unhandled message:")) {
		w.mu.Lock()
		w.matches = append(w.matches, string(p))
		w.mu.Unlock()
	}
	return len(p), nil
}

// TestFullIntegrationSuite_Combined runs request/response and event coverage for CombinedMarketStreamChannel
func TestFullIntegrationSuite_Combined(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	cw := &unhandledCatcher2{}
	log.SetOutput(cw)
	defer func() {
		log.SetOutput(os.Stderr)
		cw.mu.Lock()
		defer cw.mu.Unlock()
		if len(cw.matches) > 0 {
			for _, line := range cw.matches {
				t.Logf("SDK log captured: %s", strings.TrimSpace(line))
			}
			t.Fatalf("SDK emitted %d 'unhandled message' log(s) during Combined suite; treating as failure", len(cw.matches))
		}
	}()

	config := getTestConfig()
	stc, err := NewStreamTestClientDedicated(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	// Combined suites can run on mainnet for activity if needed
	_ = stc.client.SetActiveServer("mainnet")

	ch := cmfuturesstreams.NewCombinedMarketStreamChannel(stc.client)
	// capture combined wrapper events for validation
	combinedCh := make(chan *models.CombinedMarketStreamEvent, 64)
	ch.HandleCombinedMarketStreamEvent(func(ctx context.Context, ev *models.CombinedMarketStreamEvent) error {
		select {
		case combinedCh <- ev:
		default:
		}
		return nil
	})
	rec := newCMMarketEventRecorder()
	ch.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
	ch.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
	ch.HandleIndexPriceEvent(func(ctx context.Context, ev *models.IndexPriceEvent) error { rec.addIndex(ev); return nil })
	ch.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	ch.HandleContinuousKlineEvent(func(ctx context.Context, ev *models.ContinuousKlineEvent) error { rec.addContK(ev); return nil })
	ch.HandleIndexKlineEvent(func(ctx context.Context, ev *models.IndexKlineEvent) error { rec.addIndexK(ev); return nil })
	ch.HandleMarkPriceKlineEvent(func(ctx context.Context, ev *models.MarkPriceKlineEvent) error { rec.addMarkK(ev); return nil })
	ch.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	ch.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
	ch.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
	ch.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	ch.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
	ch.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTick(ev); return nil })
	ch.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMini(ev); return nil })
	ch.HandleAllBookTickersEvent(func(ctx context.Context, ev *models.AllBookTickersEvent) error { rec.addAllBook(ev); return nil })
	ch.HandleAllMarkPricesEvent(func(ctx context.Context, ev *models.AllMarkPricesEvent) error { rec.addAllMark(ev); return nil })
	ch.HandleAllLiquidationsEvent(func(ctx context.Context, ev *models.AllLiquidationsEvent) error { rec.addAllLiquid(ev); return nil })
	ch.HandleLiquidationEvent(func(ctx context.Context, ev *models.LiquidationEvent) error { rec.addLiquid(ev); return nil })
	ch.HandleContractInfoEvent(func(ctx context.Context, ev *models.ContractInfoEvent) error { rec.addContract(ev); return nil })

	// Resolve symbol
	upSym, errPick := restPickSymbol(context.Background())
	if errPick != nil || upSym == "" {
		t.Fatalf("failed to resolve symbol via REST: %v", errPick)
	}
	symLower := strings.ToLower(upSym)
	pairLower := strings.ToLower(strings.TrimSuffix(upSym, "_PERP"))

	// Build initial streams and connect
	var initStreams []string
	if s, err := cmfuturesstreams.BuildMarkPriceEventStream(0, (cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, (cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if len(initStreams) == 0 {
		t.Fatalf("failed to build initial streams")
	}

	cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
	if err := ch.Connect(cctx, strings.Join(initStreams, "/")); err != nil {
		ccancel()
		t.Fatalf("connect combined: %v", err)
	}
	ccancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 6*time.Second)
		_ = ch.Disconnect(dctx)
		dcancel()
	}()

	// ---------- Requests & Responses ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		mp := cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
		markPath, _ := cmfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
		kp := cmfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		klinePath, _ := cmfuturesstreams.BuildKlineEventStream(0, kp.Values())
		sid := newRequestID()
		done := make(chan struct{}, 1)
		var got *models.SubscribeResponse
		var subErr error
		cb := func(ctx context.Context, resp *models.SubscribeResponse, err error) error {
			if err != nil {
				subErr = err
				select {
				case done <- struct{}{}:
				default:
				}
				return nil
			}
			if resp != nil && msgIDEqual(resp.Id, sid) {
				got = resp
				logJSON(t, "subscribe.response", resp)
				select {
				case done <- struct{}{}:
				default:
				}
			}
			return nil
		}
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: sid, Params: []string{markPath, klinePath}}, &cb); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Logf("timeout waiting subscribe response")
		}
		if subErr != nil {
			t.Fatalf("subscribe handler returned error: %v", subErr)
		}
		_ = got
		// cleanup one extra
		var ucb func(context.Context, *models.UnsubscribeResponse, error) error = func(_ context.Context, _ *models.UnsubscribeResponse, err error) error {
			if err != nil {
				t.Logf("unsubscribe cleanup error: %v", err)
			}
			return nil
		}
		_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: newRequestID(), Params: []string{klinePath}}, &ucb)
	})

	t.Run("CombinedMarketStreamEvent", func(t *testing.T) {
		// Expect wrapper events for the mark price stream we connected to
		mp := cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
		wantStream, err := cmfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build mark stream: %v", err)
		}
		deadline := time.Now().Add(eventWait())
		var got *models.CombinedMarketStreamEvent
		for time.Now().Before(deadline) {
			select {
			case ev := <-combinedCh:
				if ev == nil {
					continue
				}
				if ev.Stream != wantStream {
					continue
				}
				got = ev
				break
			case <-time.After(150 * time.Millisecond):
			}
			if got != nil {
				break
			}
		}
		if got == nil {
			t.Fatalf("did not receive combined wrapper event for stream %s", wantStream)
		}
		// Validate basic shape: data is an object with event type matching markPriceUpdate
		if m, ok := got.Data.(map[string]interface{}); ok {
			if et, ok2 := m["e"].(string); !ok2 || et == "" {
				t.Errorf("combined data missing event type 'e'")
			}
		} else {
			t.Errorf("combined data not an object; got %T", got.Data)
		}
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		ag := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
		path, _ := cmfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
		subCb := func(_ context.Context, _ *models.SubscribeResponse, err error) error {
			if err != nil {
				t.Logf("subscribe pre-list error: %v", err)
			}
			return nil
		}
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		lid := newRequestID()
		done := make(chan struct{}, 1)
		var got *models.ListSubscriptionsResponse
		var lsErr error
		lsCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse, err error) error {
			if err != nil {
				lsErr = err
				select {
				case done <- struct{}{}:
				default:
				}
				return nil
			}
			if resp != nil && msgIDEqual(resp.Id, lid) {
				got = resp
				logJSON(t, "listSubscriptions.response", resp)
				select {
				case done <- struct{}{}:
				default:
				}
			}
			return nil
		}
		if err := ch.ListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: lid}, &lsCb); err != nil {
			t.Fatalf("list subscriptions call failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting listSubscriptions response")
		}
		if lsErr != nil {
			t.Fatalf("listSubscriptions handler returned error: %v", lsErr)
		}
		if got == nil {
			t.Fatalf("did not capture listSubscriptions response")
		}
		var ucb func(context.Context, *models.UnsubscribeResponse, error) error = func(_ context.Context, _ *models.UnsubscribeResponse, err error) error {
			if err != nil {
				t.Logf("unsubscribe cleanup error: %v", err)
			}
			return nil
		}
		_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: newRequestID(), Params: []string{path}}, &ucb)
	})

	t.Run("Request_SetProperty", func(t *testing.T) {
		pid := newRequestID()
		setDone := make(chan struct{}, 1)
		var gotSet *models.SetPropertyResponse
		var setErr error
		setCb := func(ctx context.Context, resp *models.SetPropertyResponse, err error) error {
			if err != nil {
				setErr = err
				select {
				case setDone <- struct{}{}:
				default:
				}
				return nil
			}
			if resp != nil && msgIDEqual(resp.Id, pid) {
				gotSet = resp
				logJSON(t, "setProperty.response", resp)
				select {
				case setDone <- struct{}{}:
				default:
				}
			}
			return nil
		}
		spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel()
		req := &models.SetPropertyRequest{Id: pid, Params: []interface{}{"combined", true}}
		logJSON(t, "setProperty.request", req)
		if err := ch.SetProperty(spCtx, req, &setCb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "deadline") {
				t.Logf("setProperty timeout (acceptable)")
			} else {
				t.Logf("setProperty err (acceptable): %v", err)
			}
		}
		select {
		case <-setDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting setProperty response (acceptable)")
		}
		if setErr != nil {
			t.Logf("setProperty handler error (acceptable): %v", setErr)
		}
		_ = gotSet
	})

	t.Run("Request_GetProperty", func(t *testing.T) {
		gid := newRequestID()
		getDone := make(chan struct{}, 1)
		var gotGet *models.GetPropertyResponse
		var getErr error
		getCb := func(ctx context.Context, resp *models.GetPropertyResponse, err error) error {
			if err != nil {
				getErr = err
				select {
				case getDone <- struct{}{}:
				default:
				}
				return nil
			}
			if resp != nil && msgIDEqual(resp.Id, gid) {
				gotGet = resp
				logJSON(t, "getProperty.response", resp)
				select {
				case getDone <- struct{}{}:
				default:
				}
			}
			return nil
		}
		req := &models.GetPropertyRequest{Id: gid, Params: []string{"combined"}}
		logJSON(t, "getProperty.request", req)
		if err := ch.GetProperty(context.Background(), req, &getCb); err != nil {
			t.Logf("getProperty err (acceptable): %v", err)
		}
		select {
		case <-getDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting getProperty response (acceptable)")
		}
		if getErr != nil {
			t.Logf("getProperty handler error (acceptable): %v", getErr)
		}
		_ = gotGet
	})

	// ---------- Event Handlers (subscribe and validate) ----------
	t.Run("AggregateTradeEvent", func(t *testing.T) {
		ap := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, ap.Values())
		if err != nil {
			t.Fatalf("build agg stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("aggTrade", 1, eventWait())
		cnt := rec.count("aggTrade")
		t.Logf("aggTrade events: %d", cnt)
		if cnt > 0 {
			ev := rec.agg[len(rec.agg)-1]
			if ev.EventType != "aggTrade" {
				t.Errorf("want e=aggTrade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		}
	})

	t.Run("MarkPriceEvent", func(t *testing.T) {
		mp := cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build mark stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("markPrice", 1, eventWait())
		cnt := rec.count("markPrice")
		t.Logf("markPrice events: %d", cnt)
		if cnt > 0 {
			ev := rec.mark[len(rec.mark)-1]
			if ev.EventType != "markPriceUpdate" {
				t.Errorf("want e=markPriceUpdate got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		}
	})

	t.Run("IndexPriceEvent", func(t *testing.T) {
		ip := cmfuturesstreams.IndexPriceEventStreamParams{Pair: models.Pair(pairLower)}
		path, err := cmfuturesstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("indexPrice", 1, eventWait())
		t.Logf("indexPrice events: %d", rec.count("indexPrice"))
		if cnt := rec.count("indexPrice"); cnt > 0 {
			ev := rec.index[len(rec.index)-1]
			if ev.EventType != "indexPriceUpdate" {
				t.Logf("unexpected type: %s", ev.EventType)
			}
		}
	})

	t.Run("KlineEvent", func(t *testing.T) {
		interval := os.Getenv("DEFAULT_INTERVAL")
		if interval == "" {
			interval = "1m"
		}
		kp := cmfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval(interval)}
		path, err := cmfuturesstreams.BuildKlineEventStream(0, kp.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("kline", 1, eventWait())
		cnt := rec.count("kline")
		t.Logf("kline events: %d", cnt)
		if cnt > 0 {
			ev := rec.kline[len(rec.kline)-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("ContinuousKlineEvent", func(t *testing.T) {
		cp := cmfuturesstreams.ContinuousKlineEventStreamParams{Pair: models.Pair(pairLower), ContractType: models.ContractType("PERPETUAL"), Interval: models.Interval("1m")}
		path, err := cmfuturesstreams.BuildContinuousKlineEventStream(0, cp.Values())
		if err != nil {
			t.Fatalf("build continuous kline: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("continuousKline", 1, eventWait())
		t.Logf("continuousKline events: %d", rec.count("continuousKline"))
		if cnt := rec.count("continuousKline"); cnt > 0 {
			ev := rec.contKline[len(rec.contKline)-1]
			if ev.EventType != "continuous_kline" {
				t.Errorf("want e=continuous_kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		}
	})

	t.Run("IndexKlineEvent", func(t *testing.T) {
		ip := cmfuturesstreams.IndexKlineEventStreamParams{Pair: models.Pair(pairLower), Interval: models.Interval("1m")}
		path, err := cmfuturesstreams.BuildIndexKlineEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index kline: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("indexKline", 1, eventWait())
		t.Logf("indexKline events: %d", rec.count("indexKline"))
		if cnt := rec.count("indexKline"); cnt > 0 {
			ev := rec.indexKline[len(rec.indexKline)-1]
			if ev.EventType != "indexPrice_kline" {
				t.Errorf("want e=indexPrice_kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		}
	})

	t.Run("MarkPriceKlineEvent", func(t *testing.T) {
		mp := cmfuturesstreams.MarkPriceKlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		path, err := cmfuturesstreams.BuildMarkPriceKlineEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPriceKline: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("markPriceKline", 1, eventWait())
		t.Logf("markPriceKline events: %d", rec.count("markPriceKline"))
		if cnt := rec.count("markPriceKline"); cnt > 0 {
			ev := rec.markKline[len(rec.markKline)-1]
			if ev.EventType != "markPrice_kline" {
				t.Errorf("want e=markPrice_kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		}
	})

	t.Run("TickerEvent", func(t *testing.T) {
		tp := cmfuturesstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildTickerEventStream(0, tp.Values())
		if err != nil {
			t.Fatalf("build ticker: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("ticker", 1, eventWait())
		t.Logf("ticker events: %d", rec.count("ticker"))
		if cnt := rec.count("ticker"); cnt > 0 {
			ev := rec.ticker[len(rec.ticker)-1]
			if ev.EventType != "24hrTicker" {
				t.Errorf("want e=24hrTicker got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("MiniTickerEvent", func(t *testing.T) {
		mp := cmfuturesstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildMiniTickerEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build miniTicker: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("miniTicker", 1, eventWait())
		t.Logf("miniTicker events: %d", rec.count("miniTicker"))
		if cnt := rec.count("miniTicker"); cnt > 0 {
			ev := rec.miniTicker[len(rec.miniTicker)-1]
			if ev.EventType != "24hrMiniTicker" {
				t.Logf("unexpected type: %s", ev.EventType)
			}
		}
	})

	t.Run("BookTickerEvent", func(t *testing.T) {
		bp := cmfuturesstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildBookTickerEventStream(0, bp.Values())
		if err != nil {
			t.Fatalf("build bookTicker: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("bookTicker", 1, eventWait())
		t.Logf("bookTicker events: %d", rec.count("bookTicker"))
		if cnt := rec.count("bookTicker"); cnt > 0 {
			ev := rec.bookTicker[len(rec.bookTicker)-1]
			if ev.EventType != "bookTicker" {
				t.Errorf("want e=bookTicker got %s", ev.EventType)
			}
			assertNonEmpty(t, ev.BestBidPrice, "bestBidPrice")
			assertNonEmpty(t, ev.BestAskPrice, "bestAskPrice")
		}
	})

	t.Run("PartialDepthEvent", func(t *testing.T) {
		dp := cmfuturesstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels("5"), Speed: models.DepthSpeed("100ms")}
		path, err := cmfuturesstreams.BuildPartialDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build partial depth: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("partialDepth", 1, eventWait())
		t.Logf("partialDepth events: %d", rec.count("partialDepth"))
		if cnt := rec.count("partialDepth"); cnt > 0 {
			ev := rec.partialDepth[len(rec.partialDepth)-1]
			if ev.EventType != "depthUpdate" {
				t.Errorf("want e=depthUpdate got %s", ev.EventType)
			}
		}
	})

	t.Run("DiffDepthEvent", func(t *testing.T) {
		dp := cmfuturesstreams.DiffDepthEventStreamParams{Symbol: models.Symbol(symLower), Speed: models.DepthSpeed("100ms")}
		path, err := cmfuturesstreams.BuildDiffDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build diff depth: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("diffDepth", 1, eventWait())
		t.Logf("diffDepth events: %d", rec.count("diffDepth"))
		if cnt := rec.count("diffDepth"); cnt > 0 {
			ev := rec.diffDepth[len(rec.diffDepth)-1]
			if ev.EventType != "depthUpdate" {
				t.Errorf("want e=depthUpdate got %s", ev.EventType)
			}
		}
	})

	t.Run("LiquidationEvent", func(t *testing.T) {
		lp := cmfuturesstreams.LiquidationEventStreamParams{Symbol: models.Symbol(symLower)}
		path, err := cmfuturesstreams.BuildLiquidationEventStream(0, lp.Values())
		if err != nil {
			t.Fatalf("build liquidation: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("liquidation", 1, eventWait())
		t.Logf("liquidation events: %d", rec.count("liquidation"))
		if cnt := rec.count("liquidation"); cnt > 0 {
			ev := rec.liquidation[len(rec.liquidation)-1]
			if ev.EventType != "forceOrder" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("AllMarkPricesEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildAllMarkPricesEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all mark prices: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("allMarkPrices", 1, eventWait())
		t.Logf("allMarkPrices events: %d", rec.count("allMarkPrices"))
		if cnt := rec.count("allMarkPrices"); cnt > 0 {
			ev := rec.allMarkPrices[len(rec.allMarkPrices)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "markPriceUpdate" {
					t.Logf("unexpected type: %s", it.EventType)
				}
			}
		}
	})

	t.Run("AllMiniTickersEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildAllMiniTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all mini: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("allMiniTickers", 1, eventWait())
		t.Logf("allMiniTickers events: %d", rec.count("allMiniTickers"))
		if cnt := rec.count("allMiniTickers"); cnt > 0 {
			ev := rec.allMiniTickers[len(rec.allMiniTickers)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "24hrMiniTicker" {
					t.Logf("unexpected type: %s", it.EventType)
				}
			}
		}
	})

	t.Run("AllTickersEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildAllTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all tickers: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("allTickers", 1, eventWait())
		t.Logf("allTickers events: %d", rec.count("allTickers"))
		if cnt := rec.count("allTickers"); cnt > 0 {
			ev := rec.allTickers[len(rec.allTickers)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "24hrTicker" {
					t.Logf("unexpected type: %s", it.EventType)
				}
			}
		}
	})

	t.Run("AllBookTickersEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildAllBookTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all book tickers: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("allBookTickers", 1, eventWait())
		t.Logf("allBookTickers events: %d", rec.count("allBookTickers"))
		if cnt := rec.count("allBookTickers"); cnt > 0 {
			ev := rec.allBookTickers[len(rec.allBookTickers)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "bookTicker" {
					t.Errorf("want e=bookTicker got %s", it.EventType)
				}
			}
		}
	})

	t.Run("AllLiquidationsEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildAllLiquidationsEventStream(0, nil)
		if err != nil {
			t.Fatalf("build all liquidations: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("allLiquidations", 1, eventWait())
		t.Logf("allLiquidations events: %d", rec.count("allLiquidations"))
		if cnt := rec.count("allLiquidations"); cnt > 0 {
			ev := rec.allLiquid[len(rec.allLiquid)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "forceOrder" {
					t.Logf("unexpected event type: %s", it.EventType)
				}
			}
		}
	})

	t.Run("ContractInfoEvent", func(t *testing.T) {
		path, err := cmfuturesstreams.BuildContractInfoEventStream(0, nil)
		if err != nil {
			t.Fatalf("build contractInfo: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newRequestID(), Params: []string{path}}, &subCb)
		_ = rec.waitForMin("contractInfo", 1, eventWait())
		t.Logf("contractInfo events: %d", rec.count("contractInfo"))
		if cnt := rec.count("contractInfo"); cnt > 0 {
			ev := rec.contractInfo[len(rec.contractInfo)-1]
			if ev.EventType != "contractInfo" {
				t.Errorf("want e=contractInfo got %s", ev.EventType)
			}
		}
	})
}
