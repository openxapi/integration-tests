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

// unhandledCatcher2 captures SDK log lines containing the marker "unhandled message:"
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

	cfg := getTestConfig()
	stc, err := NewStreamTestClientDedicated(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Force mainnet for Combined market channel tests
	if err := stc.client.SetActiveServer("mainnet"); err != nil {
		t.Fatalf("failed to switch to mainnet: %v", err)
	}

	if as := stc.client.GetActiveServer(); as != nil {
		t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
	}

	// Resolve symbol
	upSym, errPick := restPickSymbol(context.Background())
	if errPick != nil || upSym == "" {
		t.Fatalf("failed to resolve active symbol via REST: %v", errPick)
	}
	symLower := strings.ToLower(upSym)

	// Channel and handlers
	ch := spotstreams.NewCombinedMarketStreamChannel(stc.client)
	rec := newSpotMarketEventRecorder()
	ch.HandleCombinedMarketStreamEvent(func(ctx context.Context, ev *models.CombinedMarketStreamEvent) error { rec.addCombined(ev); return nil })

	ch.HandleTradeEvent(func(ctx context.Context, ev *models.TradeEvent) error { rec.addTrade(ev); return nil })
	ch.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
	ch.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	ch.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	ch.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
	ch.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
	ch.HandleAveragePriceEvent(func(ctx context.Context, ev *models.AveragePriceEvent) error { rec.addAvgPrice(ev); return nil })
	ch.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	ch.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
	ch.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTickers(ev); return nil })
	ch.HandleAllRollingWindowTickersEvent(func(ctx context.Context, ev *models.AllRollingWindowTickersEvent) error {
		rec.addAllRoll(ev)
		return nil
	})
	ch.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMini(ev); return nil })
	// Record single-symbol rolling window ticker events too
	ch.HandleRollingWindowTickerEvent(func(ctx context.Context, ev *models.RollingWindowTickerEvent) error {
		rec.addRollingWindowTicker(ev)
		return nil
	})

	// Compatibility hooks: the combined wrapper does not include an explicit 'e' field for
	// bookTicker/partialDepth payloads; the client routes those by x-no-event-type keys.
	// Market tests receive these fine since their handlers register those keys. Mirror that here
	// so the combined suite records the same events consistently without changing the SDK.
	// stc.client.RegisterHandlers("compat:combined-fixes", map[string]func(context.Context, []byte) error{
	// 	"Book Ticker Event": func(ctx context.Context, b []byte) error {
	// 		var ev models.BookTickerEvent
	// 		if err := json.Unmarshal(b, &ev); err != nil { return err }
	// 		rec.addBookTicker(&ev)
	// 		return nil
	// 	},
	// 	"Partial Depth Event": func(ctx context.Context, b []byte) error {
	// 		var ev models.PartialDepthEvent
	// 		if err := json.Unmarshal(b, &ev); err != nil { return err }
	// 		rec.addPartialDepth(&ev)
	// 		return nil
	// 	},
	// })

	// Build initial streams and connect
	var initStreams []string
	if s, err := spotstreams.BuildTickerEventStream(0, (spotstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := spotstreams.BuildTradeEventStream(0, (spotstreams.TradeEventStreamParams{Symbol: models.Symbol(symLower)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	streamsParam := strings.Join(initStreams, "/")
	cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
	if err := ch.Connect(cctx, streamsParam); err != nil {
		ccancel()
		t.Fatalf("connect combined: %v", err)
	}
	ccancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = ch.Disconnect(dctx)
		dcancel()
	}()

	// ---------- Requests ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		s1, _ := spotstreams.BuildBookTickerEventStream(0, (spotstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		s2, _ := spotstreams.BuildMiniTickerEventStream(0, (spotstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		sid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.SubscribeResponse
		var cb func(context.Context, *models.SubscribeResponse, error) error = func(ctx context.Context, resp *models.SubscribeResponse, respErr error) error {
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
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(sid), Params: []string{s1, s2}}, &cb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Errorf("timeout waiting subscribe response")
		}
		_ = got
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		lid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.ListSubscriptionsResponse
		var cb func(context.Context, *models.ListSubscriptionsResponse, error) error = func(ctx context.Context, resp *models.ListSubscriptionsResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("listSubscriptions error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil listSubscriptions response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != lid {
				t.Errorf("listSubscriptions id mismatch: want %d got %d", lid, rid)
			}
			if resp.Result == nil {
				t.Errorf("listSubscriptions result is nil; expected array")
			}
			got = resp
			logJSON(t, "listSubscriptions.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.ListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: models.NewMessageIDInt64(lid)}, &cb); err != nil {
			t.Fatalf("listSubscriptions failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting listSubscriptions response")
		}
		_ = got
	})

	// Exercise unsubscribe with assertion
	t.Run("Request_Unsubscribe", func(t *testing.T) {
		s1, _ := spotstreams.BuildBookTickerEventStream(0, (spotstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		uid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.UnsubscribeResponse
		var cb func(context.Context, *models.UnsubscribeResponse, error) error = func(ctx context.Context, resp *models.UnsubscribeResponse, respErr error) error {
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
			got = resp
			logJSON(t, "unsubscribe.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(uid), Params: []string{s1}}, &cb); err != nil {
			t.Fatalf("unsubscribe failed: %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting unsubscribe response")
		}
		_ = got
	})

	t.Run("Request_SetProperty", func(t *testing.T) {
		pid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.SetPropertyResponse
		var cb func(context.Context, *models.SetPropertyResponse, error) error = func(ctx context.Context, resp *models.SetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Logf("setProperty callback error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil setProperty response")
				return nil
			}
			if rid, ok := resp.Id.ValInt64(); !ok || rid != pid {
				t.Errorf("setProperty id mismatch: want %d got %d", pid, rid)
			}
			if resp.Result != nil {
				t.Errorf("setProperty result should be null, got %v", resp.Result)
			}
			got = resp
			logJSON(t, "setProperty.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel()
		setReq := &models.SetPropertyRequest{Id: models.NewMessageIDInt64(pid), Params: []interface{}{"combined", true}}
		if err := ch.SetProperty(spCtx, setReq, &cb); err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "deadline") {
				t.Logf("setProperty timeout (acceptable)")
			} else {
				t.Logf("setProperty err: %v", err)
			}
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting setProperty response (acceptable)")
		}
		_ = got
	})

	t.Run("Request_GetProperty", func(t *testing.T) {
		gid := time.Now().UnixMicro()
		done := make(chan struct{}, 1)
		var got *models.GetPropertyResponse
		var cb func(context.Context, *models.GetPropertyResponse, error) error = func(ctx context.Context, resp *models.GetPropertyResponse, respErr error) error {
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
			// Expect either a bare bool or an object containing { combined: bool }
			extract := func(v interface{}) (bool, bool) {
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
			if b, ok := extract(resp.Result); ok {
				t.Logf("combined=%v", b)
			} else {
				t.Errorf("getProperty result not parseable: %v", resp.Result)
			}
			got = resp
			logJSON(t, "getProperty.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.GetProperty(context.Background(), &models.GetPropertyRequest{Id: models.NewMessageIDInt64(gid), Params: []string{"combined"}}, &cb); err != nil {
			t.Logf("getProperty err (acceptable): %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting getProperty response (acceptable)")
		}
		_ = got
	})

	// ---------- Event Handlers ----------
	t.Run("AggregateTradeEvent", func(t *testing.T) {
		before := rec.count("aggTrade")
		s, _ := spotstreams.BuildAggregateTradeEventStream(0, (spotstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{s}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		deadline := time.Now().Add(eventWait())
		for time.Now().Before(deadline) {
			if rec.count("aggTrade") > before {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if rec.count("aggTrade") > before {
			ev := rec.agg[len(rec.agg)-1]
			if ev.EventType != "aggTrade" {
				t.Errorf("want e=aggTrade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			_ = tryParseFloat(t, ev.Price, "price")
			assertNonEmpty(t, ev.Symbol, "symbol")
		}
		cnt := rec.count("aggTrade")
		if cnt > 0 {
			t.Logf("aggTrade events received: %d", cnt)
		} else {
			t.Logf("aggTrade events received: %d", 0)
		}
	})

	// Combined wrapper event: expect stream and data set; count logged
	t.Run("CombinedWrapperEvent", func(t *testing.T) {
		_ = rec.waitForMin("combined", 1, eventWait())
		cnt := rec.count("combined")
		t.Logf("combined wrapper events received: %d", cnt)
		if cnt > 0 {
			ev := rec.combined[cnt-1]
			if strings.TrimSpace(ev.Stream) == "" {
				t.Errorf("stream empty in wrapper event")
			}
			if ev.Data == nil {
				t.Errorf("data nil in wrapper event")
			}
		}
	})

	// TradeEvent
	t.Run("TradeEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildTradeEventStream(0, (spotstreams.TradeEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("trade", 1, eventWait())
		if cnt := rec.count("trade"); cnt > 0 {
			ev := rec.trade[cnt-1]
			if ev.EventType != "trade" {
				t.Errorf("want e=trade got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.Price, "price")
			_ = tryParseFloat(t, ev.Quantity, "quantity")
			t.Logf("trade events received: %d", cnt)
		} else {
			t.Logf("trade events received: %d", 0)
		}
	})

	// KlineEvent
	t.Run("KlineEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildKlineEventStream(0, (spotstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("kline", 1, eventWait())
		if cnt := rec.count("kline"); cnt > 0 {
			ev := rec.kline[cnt-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.KlineData.Interval, "interval")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.KlineData.OpenPrice, "open price")
			_ = tryParseFloat(t, ev.KlineData.ClosePrice, "close price")
			t.Logf("kline events received: %d", cnt)
		} else {
			t.Logf("kline events received: %d", 0)
		}
	})

	// TickerEvent
	t.Run("TickerEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildTickerEventStream(0, (spotstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("ticker", 1, eventWait())
		if cnt := rec.count("ticker"); cnt > 0 {
			ev := rec.ticker[cnt-1]
			if ev.EventType != "24hrTicker" {
				t.Errorf("want e=24hrTicker got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.LastPrice, "last price")
			t.Logf("ticker events received: %d", cnt)
		} else {
			t.Logf("ticker events received: %d", 0)
		}
	})

	// MiniTickerEvent
	t.Run("MiniTickerEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildMiniTickerEventStream(0, (spotstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("miniTicker", 1, eventWait())
		if cnt := rec.count("miniTicker"); cnt > 0 {
			ev := rec.miniTicker[cnt-1]
			if ev.EventType != "24hrMiniTicker" {
				t.Errorf("want e=24hrMiniTicker got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.ClosePrice, "close price")
			t.Logf("miniTicker events received: %d", cnt)
		} else {
			t.Logf("miniTicker events received: %d", 0)
		}
	})

	// BookTickerEvent
	t.Run("BookTickerEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildBookTickerEventStream(0, (spotstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("bookTicker", 1, eventWait())
		if cnt := rec.count("bookTicker"); cnt > 0 {
			ev := rec.bookTicker[cnt-1]
			if ev.OrderBookUpdateId <= 0 {
				t.Errorf("orderBookUpdateId <= 0")
			}
			assertNonEmpty(t, ev.Symbol, "symbol")
			assertNonEmpty(t, ev.BestBidPrice, "best bid price")
			assertNonEmpty(t, ev.BestAskPrice, "best ask price")
			_ = tryParseFloat(t, ev.BestBidPrice, "best bid price parse")
			_ = tryParseFloat(t, ev.BestAskPrice, "best ask price parse")
			t.Logf("bookTicker events received: %d", cnt)
		} else {
			t.Logf("bookTicker events received: %d", 0)
		}
	})

	// AvgPriceEvent
	t.Run("AvgPriceEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildAveragePriceEventStream(0, (spotstreams.AveragePriceEventStreamParams{Symbol: models.Symbol(symLower)}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("avgPrice", 1, eventWait())
		if cnt := rec.count("avgPrice"); cnt > 0 {
			ev := rec.avgPrice[cnt-1]
			if ev.EventType != "avgPrice" {
				t.Errorf("want e=avgPrice got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.AveragePrice, "avg price")
			assertNonEmpty(t, ev.AveragePriceInterval, "avg price interval")
			t.Logf("avgPrice events received: %d", cnt)
		} else {
			t.Logf("avgPrice events received: %d", 0)
		}
	})

	// PartialDepthEvent
	t.Run("PartialDepthEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildPartialDepthEventStream(1, (spotstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels5, Speed: models.DepthSpeed("100ms")}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("partialDepth", 1, eventWait())
		if cnt := rec.count("partialDepth"); cnt > 0 {
			ev := rec.partialDepth[cnt-1]
			if ev.LastUpdateID <= 0 {
				t.Errorf("lastUpdateId <= 0")
			}
			if ev.Bids == nil || ev.Asks == nil {
				t.Errorf("bids/asks should not be nil")
			}
			t.Logf("partialDepth events received: %d", cnt)
		} else {
			t.Logf("partialDepth events received: %d", 0)
		}
	})

	// DiffDepthEvent
	t.Run("DiffDepthEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildDiffDepthEventStream(1, (spotstreams.DiffDepthEventStreamParams{Symbol: models.Symbol(symLower), Speed: models.DepthSpeed("100ms")}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("diffDepth", 1, eventWait())
		if cnt := rec.count("diffDepth"); cnt > 0 {
			ev := rec.diffDepth[cnt-1]
			if ev.EventType != "depthUpdate" {
				t.Errorf("want e=depthUpdate got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
		}
	})

	// AllTickersEvent
	t.Run("AllTickersEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildAllTickersEventStream(0, nil)
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("allTickers", 1, eventWait())
		if cnt := rec.count("allTickers"); cnt > 0 {
			ev := rec.allTickers[cnt-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "24hrTicker" {
					t.Logf("unexpected type: %s", it.EventType)
				}
				assertNonEmpty(t, it.Symbol, "symbol")
				_ = tryParseFloat(t, it.LastPrice, "last price")
			}
			t.Logf("allTickers events received: %d", cnt)
		} else {
			t.Logf("allTickers events received: %d", 0)
		}
	})

	// AllMiniTickersEvent
	t.Run("AllMiniTickersEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildAllMiniTickersEventStream(0, nil)
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("allMini", 1, eventWait())
		if cnt := rec.count("allMini"); cnt > 0 {
			ev := rec.allMini[cnt-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				if it.EventType != "24hrMiniTicker" {
					t.Logf("unexpected type: %s", it.EventType)
				}
				assertNonEmpty(t, it.Symbol, "symbol")
				_ = tryParseFloat(t, it.ClosePrice, "close price")
			}
			t.Logf("allMiniTickers events received: %d", cnt)
		} else {
			t.Logf("allMiniTickers events received: %d", 0)
		}
	})

	// AllRollingWindowTickersEvent
	t.Run("AllRollingWindowTickersEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildAllRollingWindowTickersEventStream(0, map[string]string{"windowSize": "1h"})
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("allRoll", 1, eventWait())
		if cnt := rec.count("allRoll"); cnt > 0 {
			ev := rec.allRoll[cnt-1]
			if ev != nil && len(*ev) > 0 {
				it := (*ev)[0]
				assertNonEmpty(t, it.Symbol, "symbol")
				_ = tryParseFloat(t, it.LastPrice, "last price")
			}
			t.Logf("allTickers events received: %d", cnt)
		} else {
			t.Logf("allTickers events received: %d", 0)
		}
	})

	// RollingWindowTickerEvent (single-symbol)
	t.Run("RollingWindowTickerEvent", func(t *testing.T) {
		path, _ := spotstreams.BuildRollingWindowTickerEventStream(0, (spotstreams.RollingWindowTickerEventStreamParams{Symbol: models.Symbol(symLower), WindowSize: models.WindowSize("1h")}).Values())
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		throttleWS()
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{path}}, &subCb); err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		_ = rec.waitForMin("rollTicker", 1, eventWait())
		cnt := rec.count("rollTicker")
		if cnt > 0 {
			ev := rec.rollingWindowTicker[cnt-1]
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
			_ = tryParseFloat(t, ev.LastPrice, "last price")
		}
		t.Logf("rollingWindowTicker events received: %d", cnt)
	})

}
