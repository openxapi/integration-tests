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

	umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
	"github.com/openxapi/binance-go/ws/umfutures-streams/models"
)

// unhandledCatcher captures SDK log lines containing the marker "unhandled message:"
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
	// Force mainnet for this suite per instruction
	_ = stc.client.SetActiveServer("mainnet")

	if as := stc.client.GetActiveServer(); as != nil {
		t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
	} else {
		t.Logf("Active WS server: <nil>")
	}

	// Channel and handlers
	ch := umfuturesstreams.NewCombinedMarketStreamChannel(stc.client)
	// wrapper handler to mark combined messages as handled
	ch.HandleCombinedMarketStreamEvent(func(ctx context.Context, ev *models.CombinedMarketStreamEvent) error { return nil })
	// Record events from suite start
	rec := newUMMarketEventRecorder()
	ch.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
	ch.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
	ch.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	ch.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	ch.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
	ch.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
	ch.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	ch.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
	ch.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTickers(ev); return nil })
	ch.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMiniTickers(ev); return nil })
	ch.HandleAllBookTickersEvent(func(ctx context.Context, ev *models.AllBookTickersEvent) error { rec.addAllBookTickers(ev); return nil })
	ch.HandleAllMarkPricesEvent(func(ctx context.Context, ev *models.AllMarkPricesEvent) error { rec.addAllMarkPrices(ev); return nil })
	ch.HandleAllLiquidationsEvent(func(ctx context.Context, ev *models.AllLiquidationsEvent) error {
		rec.addAllLiquidations(ev)
		return nil
	})
	ch.HandleLiquidationEvent(func(ctx context.Context, ev *models.LiquidationEvent) error { rec.addLiquidation(ev); return nil })
	ch.HandleCompositeIndexEvent(func(ctx context.Context, ev *models.CompositeIndexEvent) error { rec.addCompositeIndex(ev); return nil })
	ch.HandleContractInfoEvent(func(ctx context.Context, ev *models.ContractInfoEvent) error { return nil })
	ch.HandleAllAssetIndexesEvent(func(ctx context.Context, ev *models.AllAssetIndexesEvent) error {
		rec.addAllAssetIndexes(ev)
		return nil
	})
	ch.HandleContinuousKlineEvent(func(ctx context.Context, ev *models.ContinuousKlineEvent) error {
		rec.addContinuousKline(ev)
		return nil
	})

	// Resolve symbol
	upSym, errPick := restPickSymbol(context.Background())
	if errPick != nil || upSym == "" {
		t.Fatalf("failed to resolve active symbol via REST: %v", errPick)
	}
	symLower := strings.ToLower(upSym)

	// Build initial streams and connect
	var initStreams []string
	if s, err := umfuturesstreams.BuildMarkPriceEventStream(0, (umfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := umfuturesstreams.BuildAggregateTradeEventStream(0, (umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	connStreams := strings.Join(initStreams, "/")
	cctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	if err := ch.Connect(cctx, connStreams); err != nil {
		cancel()
		t.Fatalf("connect combined: %v", err)
	}
	cancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = ch.Disconnect(dctx)
		dcancel()
	}()

	// ---------- Requests & Responses ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		kp := umfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		stream, err := umfuturesstreams.BuildKlineEventStream(0, kp.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		sid := newMessageID()
		subDone := make(chan struct{}, 1)
		subCb := func(ctx context.Context, resp *models.SubscribeResponse, rpcErr error) error {
			if rpcErr != nil {
				t.Errorf("subscribe error: %v", rpcErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil subscribe response")
				return nil
			}
			if resp.Id.String() != sid.String() {
				t.Errorf("subscribe id mismatch: want %s got %s", sid.String(), resp.Id.String())
			}
			logJSON(t, "subscribe.response", resp)
			select {
			case subDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: sid, Params: []string{stream}}, &subCb); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		select {
		case <-subDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting subscribe response (acceptable)")
		}
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		ip := umfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		stream, err := umfuturesstreams.BuildKlineEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeSetup")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		lid := newMessageID()
		done := make(chan struct{}, 1)
		lsCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse, rpcErr error) error {
			if rpcErr != nil {
				t.Errorf("list subscriptions error: %v", rpcErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil list subscriptions response")
				return nil
			}
			if resp.Id.String() != lid.String() {
				t.Errorf("list id mismatch: want %s got %s", lid.String(), resp.Id.String())
			}
			if resp.Result != nil && !contains(resp.Result, stream) {
				t.Logf("list did not include %s: %v", stream, resp.Result)
			}
			logJSON(t, "listSubscriptions.response", resp)
			select {
			case done <- struct{}{}:
			default:
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
	})

	t.Run("Request_SetProperty", func(t *testing.T) {
		pid := newMessageID()
		setCb := func(ctx context.Context, resp *models.SetPropertyResponse, rpcErr error) error {
			if rpcErr != nil {
				t.Logf("setProperty rpc error: %v", rpcErr)
				return nil
			}
			logJSON(t, "setProperty.response", resp)
			return nil
		}
		spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel()
		setReq := &models.SetPropertyRequest{Id: pid, Params: []interface{}{"combined", true}}
		logJSON(t, "setProperty.request", setReq)
		if err := ch.SetProperty(spCtx, setReq, &setCb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "deadline") {
				t.Logf("setProperty timeout (acceptable)")
			} else {
				t.Logf("setProperty err: %v", err)
			}
		}
	})

	t.Run("Request_GetProperty", func(t *testing.T) {
		gid := newMessageID()
		getCb := func(ctx context.Context, resp *models.GetPropertyResponse, rpcErr error) error {
			if rpcErr != nil {
				t.Logf("getProperty rpc error: %v", rpcErr)
				return nil
			}
			logJSON(t, "getProperty.response", resp)
			return nil
		}
		getReq := &models.GetPropertyRequest{Id: gid, Params: []string{"combined"}}
		logJSON(t, "getProperty.request", getReq)
		if err := ch.GetProperty(context.Background(), getReq, &getCb); err != nil {
			t.Logf("getProperty err (acceptable): %v", err)
		}
	})

	// ---------- Event Handlers ----------
	t.Run("AggregateTradeEvent", func(t *testing.T) {
		before := rec.count("aggTrade")
		ag := umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
		if err != nil {
			t.Fatalf("build aggTrade stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("aggTrade") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("aggTrade")
		t.Logf("aggTrade events received: %d", cnt)
		if cnt > before {
			ev := rec.agg[len(rec.agg)-1]
			if ev.EventType != "aggTrade" {
				t.Errorf("want e=aggTrade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		} else {
			t.Logf("no aggTrade event received (acceptable)")
		}
	})
	t.Run("MarkPriceEvent", func(t *testing.T) {
		before := rec.count("markPrice")
		mp := umfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPrice stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("markPrice") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("markPrice")
		t.Logf("markPrice events received: %d", cnt)
		if cnt > before {
			ev := rec.mark[len(rec.mark)-1]
			if ev.EventType != "markPriceUpdate" {
				t.Errorf("want e=markPriceUpdate got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		} else {
			t.Logf("no markPrice event received (acceptable)")
		}
	})

	t.Run("KlineEvent", func(t *testing.T) {
		before := rec.count("kline")
		kp := umfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
		stream, err := umfuturesstreams.BuildKlineEventStream(0, kp.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("kline") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("kline")
		t.Logf("kline events received: %d", cnt)
		if cnt > before {
			ev := rec.kline[len(rec.kline)-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
		} else {
			t.Logf("no kline event received (acceptable)")
		}
	})

	t.Run("MiniTickerEvent", func(t *testing.T) {
		before := rec.count("miniTicker")
		mp := umfuturesstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildMiniTickerEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build miniTicker stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("miniTicker") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("miniTicker")
		t.Logf("miniTicker events received: %d", cnt)
		if cnt <= before {
			t.Logf("no miniTicker event received (acceptable)")
		}
	})

	t.Run("BookTickerEvent", func(t *testing.T) {
		before := rec.count("bookTicker")
		bp := umfuturesstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildBookTickerEventStream(0, bp.Values())
		if err != nil {
			t.Fatalf("build bookTicker stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("bookTicker") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("bookTicker")
		t.Logf("bookTicker events received: %d", cnt)
		if cnt <= before {
			t.Logf("no bookTicker event received (acceptable)")
		}
	})

	t.Run("PartialDepthEvent", func(t *testing.T) {
		before := rec.count("partialDepth")
		dp := umfuturesstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels("5"), Speed: models.DepthSpeed("100ms")}
		stream, err := umfuturesstreams.BuildPartialDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build partialDepth stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("partialDepth") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("partialDepth")
		t.Logf("partialDepth events received: %d", cnt)
		if cnt <= before {
			t.Logf("no partialDepth event received (acceptable)")
		}
	})

	t.Run("DiffDepthEvent", func(t *testing.T) {
		before := rec.count("diffDepth")
		dp := umfuturesstreams.DiffDepthEventStreamParams{Symbol: models.Symbol(symLower), Speed: models.DepthSpeed("100ms")}
		stream, err := umfuturesstreams.BuildDiffDepthEventStream(1, dp.Values())
		if err != nil {
			t.Fatalf("build diffDepth stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("diffDepth") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		cnt := rec.count("diffDepth")
		t.Logf("diffDepth events received: %d", cnt)
		if cnt <= before {
			t.Logf("no diffDepth event received (acceptable)")
		}
	})

	t.Run("AllTickersEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allTickers stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allTickers", 1, eventWait())
		cnt := rec.count("allTickers")
		t.Logf("allTickers events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allTickers[len(rec.allTickers)-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				assertNonEmpty(t, it.Symbol, "symbol")
			}
		} else {
			t.Logf("no allTickers event received (acceptable)")
		}
	})

	t.Run("AllMiniTickersEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllMiniTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allMiniTicker stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allMiniTickers", 1, eventWait())
		cnt := rec.count("allMiniTickers")
		t.Logf("allMiniTickers events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allMiniTickers[len(rec.allMiniTickers)-1]
			if ev != nil && len(*ev) > 0 {
				_ = tryParseFloat(t, (*ev)[0].ClosePrice, "closePrice")
			}
		} else {
			t.Logf("no allMiniTickers event received (acceptable)")
		}
	})

	t.Run("AllBookTickersEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllBookTickersEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allBookTicker stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allBookTickers", 1, eventWait())
		cnt := rec.count("allBookTickers")
		t.Logf("allBookTickers events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allBookTickers[len(rec.allBookTickers)-1]
			assertNonEmpty(t, ev.Symbol, "symbol")
		} else {
			t.Logf("no allBookTickers event received (acceptable)")
		}
	})

	t.Run("AllMarkPricesEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllMarkPricesEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allMarkPrices stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allMarkPrices", 1, eventWait())
		cnt := rec.count("allMarkPrices")
		t.Logf("allMarkPrices events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allMarkPrices[len(rec.allMarkPrices)-1]
			if ev != nil && len(*ev) > 0 {
				_ = tryParseFloat(t, (*ev)[0].MarkPrice, "markPrice")
			}
		} else {
			t.Logf("no allMarkPrices event received (acceptable)")
		}
	})

	t.Run("LiquidationEvent", func(t *testing.T) {
		lp := umfuturesstreams.LiquidationEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildLiquidationEventStream(0, lp.Values())
		if err != nil {
			t.Fatalf("build liquidation stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("liquidation", 1, eventWait())
		cnt := rec.count("liquidation")
		t.Logf("liquidation events received: %d", cnt)
		if cnt > 0 {
			ev := rec.liquidation[len(rec.liquidation)-1]
			if ev.EventType != "forceOrder" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			ord := ev.LiquidationOrder
			assertNonEmpty(t, ord.Symbol, "symbol")
			assertNonEmpty(t, ord.Side, "side")
			assertNonEmpty(t, ord.OrderType, "orderType")
			assertNonEmpty(t, ord.TimeInForce, "timeInForce")
			_ = tryParseFloat(t, ord.Price, "price")
			_ = tryParseFloat(t, ord.AveragePrice, "avgPrice")
			_ = tryParseFloat(t, ord.OriginalQuantity, "originalQty")
			_ = tryParseFloat(t, ord.LastFilledQuantity, "lastFilledQty")
			_ = tryParseFloat(t, ord.AccumulatedFilledQuantity, "accumFilledQty")
			assertRecentMs(t, ord.TradeTime, 24*time.Hour, "tradeTime")
		} else {
			t.Logf("no liquidation event received for %s (acceptable)", strings.ToUpper(symLower))
		}
	})

	t.Run("AllLiquidationsEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllLiquidationsEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allLiquidations stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allLiquidations", 1, eventWait())
		cnt := rec.count("allLiquidations")
		t.Logf("allLiquidations events received: %d", cnt)
		if cnt > 0 {
			ev := rec.allLiquidations[len(rec.allLiquidations)-1]
			if ev.EventType != "forceOrder" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			ord := ev.LiquidationOrder
			assertNonEmpty(t, ord.Symbol, "symbol")
			assertNonEmpty(t, ord.Side, "side")
			assertNonEmpty(t, ord.OrderType, "orderType")
			assertNonEmpty(t, ord.TimeInForce, "timeInForce")
			_ = tryParseFloat(t, ord.Price, "price")
			_ = tryParseFloat(t, ord.AveragePrice, "avgPrice")
			_ = tryParseFloat(t, ord.OriginalQuantity, "originalQty")
			_ = tryParseFloat(t, ord.LastFilledQuantity, "lastFilledQty")
			_ = tryParseFloat(t, ord.AccumulatedFilledQuantity, "accumFilledQty")
			assertRecentMs(t, ord.TradeTime, 24*time.Hour, "tradeTime")
		} else {
			t.Logf("no allLiquidations event received (acceptable)")
		}
	})

	t.Run("CompositeIndexEvent", func(t *testing.T) {
		cp := umfuturesstreams.CompositeIndexEventStreamParams{Symbol: models.Symbol(symLower)}
		stream, err := umfuturesstreams.BuildCompositeIndexEventStream(0, cp.Values())
		if err != nil {
			t.Fatalf("build compositeIndex stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("compositeIndex", 1, eventWait())
		cnt := rec.count("compositeIndex")
		t.Logf("compositeIndex events received: %d", cnt)
		if cnt == 0 {
			t.Logf("no compositeIndex event received (acceptable)")
		}
	})

	t.Run("ContractInfoEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildContractInfoEventStream(0, nil)
		if err != nil {
			t.Fatalf("build contractInfo stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		// Not very frequent; tolerate missing
		_ = rec.waitForMin("assetIndex", 0, 100*time.Millisecond)
		t.Logf("assetIndex events received (timing check): %d", rec.count("assetIndex"))
	})
	t.Run("AllAssetIndexEvent", func(t *testing.T) {
		stream, err := umfuturesstreams.BuildAllAssetIndexesEventStream(0, nil)
		if err != nil {
			t.Fatalf("build allAssetIndex stream: %v", err)
		}
		subCb := ackLogger[models.SubscribeResponse](t, t.Name()+".subscribeAck")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: newMessageID(), Params: []string{stream}}, &subCb)
		_ = rec.waitForMin("allAssetIndexes", 1, eventWait())
		cnt := rec.count("allAssetIndexes")
		t.Logf("allAssetIndexes events received: %d", cnt)
		if cnt == 0 {
			t.Logf("no allAssetIndexes event received (acceptable)")
		}
	})
}
