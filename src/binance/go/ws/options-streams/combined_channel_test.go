package streamstest

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	optionsstreams "github.com/openxapi/binance-go/ws/options-streams"
	"github.com/openxapi/binance-go/ws/options-streams/models"
)

// TestFullIntegrationSuite_Combined runs request/response and event coverage for CombinedMarketStreamChannel
func TestFullIntegrationSuite_Combined(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Capture SDK log output and fail on any 'unhandled message' during this suite
	cw := &unhandledCatcher{}
	log.SetOutput(cw)
	defer func() {
		// Restore default output
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

	// Dedicated client
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

	// Combined channel
	ch := optionsstreams.NewCombinedMarketStreamChannel(stc.client)
	// Suite-level event recorder and global typed handlers
	rec := newMarketEventRecorder()
	// Always register wrapper handler to mark combined messages as handled
	ch.HandleCombinedMarketStreamEvent(func(ctx context.Context, ev *models.CombinedMarketStreamEvent) error { return nil })
	ch.HandleIndexPriceEvent(func(ctx context.Context, ev *models.IndexPriceEvent) error { rec.addIndex(ev); return nil })
	ch.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
	ch.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
	ch.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
	ch.HandleTickerByUnderlyingEvent(func(ctx context.Context, ev *models.TickerByUnderlyingEvent) error {
		rec.addTickerByUnderlying(ev)
		return nil
	})
	ch.HandleTradeEvent(func(ctx context.Context, ev *models.TradeEvent) error { rec.addTrade(ev); return nil })
	ch.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
	ch.HandleNewSymbolInfoEvent(func(ctx context.Context, ev *models.NewSymbolInfoEvent) error { rec.addNewSymbolInfo(ev); return nil })
	ch.HandleOpenInterestEvent(func(ctx context.Context, ev *models.OpenInterestEvent) error { rec.addOpenInterest(ev); return nil })

	// Resolve underlyings before connecting so we can attach multiple streams at connect time
	upA, lowA, upB, lowB, pickErr := restPickTwoUnderlyings(context.Background())
	if pickErr != nil || upA == "" || lowA == "" {
		t.Fatalf("failed to resolve active underlying(s) via REST: %v", pickErr)
	}
	underlyingA := upA
	underlyingALower := lowA
	underlyingB := upB
	underlyingBLower := lowB
	t.Logf("Using underlyings from REST: A=%s/%s B=%s/%s", underlyingA, underlyingALower, underlyingB, underlyingBLower)

	// Build multiple initial streams and connect combined channel with them
	var initStreams []string
	if s, err := optionsstreams.BuildNewSymbolInfoEventStream(0, nil); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := optionsstreams.BuildIndexPriceEventStream(0, (optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingB + "USDT")}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := optionsstreams.BuildMarkPriceEventStream(0, (optionsstreams.MarkPriceEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingA)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if s, err := optionsstreams.BuildTradeEventStream(1, (optionsstreams.TradeEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingA)}).Values()); err == nil {
		initStreams = append(initStreams, s)
	}
	if exps, err := getActiveExpirationDates(underlyingB); err == nil && len(exps) > 0 {
		if s, err := optionsstreams.BuildTickerByUnderlyingEventStream(0, (optionsstreams.TickerByUnderlyingEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingBLower), ExpirationDate: models.ExpirationDate(exps[0])}).Values()); err == nil {
			initStreams = append(initStreams, s)
		}
	}
	connStreams := strings.Join(initStreams, "/")
	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := ch.Connect(cctx, connStreams); err != nil {
		cancel()
		t.Fatalf("connect combined: %v", err)
	}
	cancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = ch.Disconnect(dctx)
		dcancel()
	}()

	// underlyings already resolved above for initial stream connect

	// ---------- Requests & Responses (split) ----------
	t.Run("Request_Subscribe", func(t *testing.T) {
		// Build index stream using full pair symbol per spec (e.g., BTCUSDT@index)
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingA + "USDT")}
		stream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		start := rec.count("index")
		sid := time.Now().UnixMicro()
		sidMsg := models.NewMessageIDInt64(sid)
		subDone := make(chan struct{}, 1)
		subCb := func(ctx context.Context, resp *models.SubscribeResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("subscribe response error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil subscribe response")
				return nil
			}
			assertMessageIDInt64(t, resp.Id, sid, "subscribe response")
			logJSON(t, "subscribe.response", resp)
			select {
			case subDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: sidMsg, Params: []string{stream}}, &subCb); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		t.Logf("subscribe request sent (id=%d) for %s", sid, stream)
		// Wait for subscribe callback to complete before continuing
		select {
		case <-subDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting subscribe response (acceptable)")
		}
		// Wait for at least one new index event recorded
		_ = rec.waitForMin("index", start+1, eventWait())
		if evs := rec.getIndex(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "index" {
				t.Errorf("want e=index got %s", ev.EventType)
			}
		}
		t.Logf("IndexPriceEvent total recorded: %d", rec.count("index"))
	})

	t.Run("Request_ListSubscriptions", func(t *testing.T) {
		// Build index stream and ensure it's subscribed so list includes it
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingA + "USDT")}
		stream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb); err != nil {
			t.Fatalf("subscribe before list failed: %v", err)
		}
		lid := time.Now().UnixMicro()
		lidMsg := models.NewMessageIDInt64(lid)
		lsDone := make(chan struct{}, 1)
		lsCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("listSubscriptions response error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil list subscriptions response")
				return nil
			}
			assertMessageIDInt64(t, resp.Id, lid, "listSubscriptions response")
			if resp.Result != nil && !contains(resp.Result, stream) {
				t.Logf("list did not include %s: %v", stream, resp.Result)
			}
			logJSON(t, "listSubscriptions.response", resp)
			select {
			case lsDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.ListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: lidMsg}, &lsCb); err != nil {
			t.Fatalf("list subscriptions call failed: %v", err)
		}
		t.Logf("listSubscriptions request sent (id=%d)", lid)
		select {
		case <-lsDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting listSubscriptions response (acceptable)")
		}
		// Clean up
		unsubCb := func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb)
	})

	t.Run("Request_SetProperty", func(t *testing.T) {
		pid := time.Now().UnixMicro()
		pidMsg := models.NewMessageIDInt64(pid)
		setDone := make(chan struct{}, 1)
		setCb := func(ctx context.Context, resp *models.SetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("setProperty response error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil setProperty response")
				return nil
			}
			assertMessageIDInt64(t, resp.Id, pid, "setProperty response")
			logJSON(t, "setProperty.response", resp)
			select {
			case setDone <- struct{}{}:
			default:
			}
			return nil
		}
		spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer spCancel()
		if err := ch.SetProperty(spCtx, &models.SetPropertyRequest{Id: pidMsg, Params: []interface{}{"combined", true}}, &setCb); err != nil {
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
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting setProperty response (acceptable)")
		}
	})

	t.Run("Request_GetProperty", func(t *testing.T) {
		gid := time.Now().UnixMicro()
		gidMsg := models.NewMessageIDInt64(gid)
		getDone := make(chan struct{}, 1)
		getCb := func(ctx context.Context, resp *models.GetPropertyResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("getProperty response error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil getProperty response")
				return nil
			}
			assertMessageIDInt64(t, resp.Id, gid, "getProperty response")
			logJSON(t, "getProperty.response", resp)
			select {
			case getDone <- struct{}{}:
			default:
			}
			return nil
		}
		gCtx, gCancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer gCancel()
		if err := ch.GetProperty(gCtx, &models.GetPropertyRequest{Id: gidMsg, Params: []string{"combined"}}, &getCb); err != nil {
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
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting getProperty response (acceptable)")
		}
	})

	t.Run("Request_Unsubscribe", func(t *testing.T) {
		// Build index stream using full pair symbol per spec (e.g., BTCUSDT@index)
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingA + "USDT")}
		stream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		subCb := func(context.Context, *models.SubscribeResponse, error) error { return nil }
		if err := ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb); err != nil {
			t.Fatalf("subscribe before unsubscribe failed: %v", err)
		}
		uid := time.Now().UnixMicro()
		uidMsg := models.NewMessageIDInt64(uid)
		unsubDone := make(chan struct{}, 1)
		unsubCb := func(ctx context.Context, resp *models.UnsubscribeResponse, respErr error) error {
			if respErr != nil {
				t.Errorf("unsubscribe response error: %v", respErr)
				return nil
			}
			if resp == nil {
				t.Errorf("nil unsubscribe response")
				return nil
			}
			assertMessageIDInt64(t, resp.Id, uid, "unsubscribe response")
			logJSON(t, "unsubscribe.response", resp)
			select {
			case unsubDone <- struct{}{}:
			default:
			}
			return nil
		}
		if err := ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: uidMsg, Params: []string{stream}}, &unsubCb); err != nil {
			t.Fatalf("unsubscribe call failed: %v", err)
		}
		t.Logf("unsubscribe request sent (id=%d) for %s", uid, stream)
		select {
		case <-unsubDone:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting unsubscribe response (acceptable)")
		}
	})

	// ---------- Combined Wrapper Event ----------
	t.Run("CombinedWrapperEvent", func(t *testing.T) {
		evCh := make(chan *models.CombinedMarketStreamEvent, 1)
		ch.HandleCombinedMarketStreamEvent(func(ctx context.Context, ev *models.CombinedMarketStreamEvent) error {
			select {
			case evCh <- ev:
			default:
				{
				}
			}
			return nil
		})
		// Build index stream using full pair symbol per spec (e.g., BTCUSDT@index)
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingA + "USDT")}
		stream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		var subCb func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb)
		var unsubCb func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb)
		}()
		cnt := 0
		select {
		case ev := <-evCh:
			cnt++
			if ev.Stream != stream {
				t.Logf("wrapper stream mismatch: want %s got %s", stream, ev.Stream)
			}
			if m, ok := ev.Data.(map[string]interface{}); ok {
				if et, _ := m["e"].(string); et == "" {
					t.Errorf("wrapper data missing event type 'e'")
				}
			} else {
				t.Logf("wrapper data not object (ok): %T", ev.Data)
			}
			logJSON(t, "event.combinedWrapper", ev)
		case <-time.After(10 * time.Second):
			t.Logf("timeout waiting combined wrapper event (acceptable)")
		}
		t.Logf("CombinedWrapperEvent count: %d", cnt)
	})

	// ---------- Event Handlers ----------
	t.Run("IndexPriceEvent", func(t *testing.T) {
		ip := optionsstreams.IndexPriceEventStreamParams{Symbol: models.Symbol(underlyingA + "USDT")}
		stream, err := optionsstreams.BuildIndexPriceEventStream(0, ip.Values())
		if err != nil {
			t.Fatalf("build index stream: %v", err)
		}
		// Directly subscribe (no list gating in new SDK)
		var subCb2 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("index")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb2)
		var unsubCb2 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb2)
		}()
		_ = rec.waitForMin("index", start+1, eventWait())
		if evs := rec.getIndex(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "index" {
				t.Errorf("want e=index got %s", ev.EventType)
			}
			// SDK field `UnderlyingSymbol` maps to JSON `s` which for index events is the full pair (e.g., BTCUSDT)
			expectedPair := underlyingA + "USDT"
			if !strings.EqualFold(ev.UnderlyingSymbol, expectedPair) {
				t.Logf("index event symbol mismatch: want %s got %s", expectedPair, ev.UnderlyingSymbol)
			}
			_ = tryParseFloat(t, ev.IndexPrice, "indexPrice")
			assertRecentMs(t, ev.EventTimestamp, 2*time.Hour, "eventTimestamp")
			if restValidationEnabled() {
				if px, err := restIndexPrice(context.Background(), underlyingA); err == nil {
					if v, err2 := strconv.ParseFloat(ev.IndexPrice, 64); err2 == nil {
						assertWithinTolerancePercent(t, v, px, 2.0, "index price vs REST")
					}
				}
			}
			logJSON(t, "event.indexPrice", ev)
		} else {
			t.Logf("no index events recorded (acceptable)")
		}
		t.Logf("IndexPriceEvent total recorded: %d", rec.count("index"))
	})

	t.Run("MarkPriceEvent", func(t *testing.T) {
		// Build markPrice stream using typed params (underlying asset)
		mp := optionsstreams.MarkPriceEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(strings.ToLower(underlyingA))}
		stream, err := optionsstreams.BuildMarkPriceEventStream(0, mp.Values())
		if err != nil {
			t.Fatalf("build markPrice stream: %v", err)
		}
		var subCb3 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("markPrice")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb3)
		var unsubCb3 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb3)
		}()
		_ = rec.waitForMin("markPrice", start+1, eventWait())
		if evs := rec.getMark(); len(evs) > 0 && evs[len(evs)-1] != nil {
			ev := evs[len(evs)-1]
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
				if i == 0 && restValidationEnabled() {
					if px, err := restMarkPrice(context.Background(), it.OptionSymbol); err == nil && px > 0 {
						if v, err2 := strconv.ParseFloat(it.OptionMarkPrice, 64); err2 == nil {
							assertWithinTolerancePercent(t, v, px, 5.0, "mark price vs REST")
						}
					}
				}
			}
			logJSON(t, "event.markPrice", ev)
		} else {
			t.Logf("no markPrice events recorded (acceptable)")
		}
		t.Logf("MarkPriceEvent total recorded: %d", rec.count("markPrice"))
	})

	t.Run("KlineEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingA, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		interval := os.Getenv("DEFAULT_INTERVAL")
		if interval == "" {
			interval = "1m"
		}
		// Build kline stream using typed params
		kp := optionsstreams.KlineEventStreamParams{Symbol: models.Symbol(symbol), Interval: models.Interval(interval)}
		stream, err := optionsstreams.BuildKlineEventStream(0, kp.Values())
		if err != nil {
			t.Fatalf("build kline stream: %v", err)
		}
		var subCb4 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("kline")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb4)
		var unsubCb4 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb4)
		}()
		_ = rec.waitForMin("kline", start+1, eventWaitLong())
		if evs := rec.getKline(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "kline" {
				t.Errorf("want e=kline got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
			if !strings.EqualFold(ev.OptionSymbol, symbol) {
				t.Logf("kline event symbol mismatch: want %s got %s", symbol, ev.OptionSymbol)
			}
			if ev.KlineData.Interval != interval {
				t.Logf("interval mismatch: %s", ev.KlineData.Interval)
			}
			if ev.KlineData.Symbol != "" && !strings.EqualFold(ev.KlineData.Symbol, symbol) {
				t.Logf("kline data symbol mismatch: want %s got %s", symbol, ev.KlineData.Symbol)
			}
			_ = tryParseInt64(t, ev.KlineData.FirstTradeID, "k.F")
			_ = tryParseInt64(t, ev.KlineData.LastTradeID, "k.L")
			if ev.KlineData.NumberOfTrades < 0 {
				t.Errorf("numberOfTrades < 0: %d", ev.KlineData.NumberOfTrades)
			}
			_ = tryParseFloat(t, ev.KlineData.OpenPrice, "k.o")
			_ = tryParseFloat(t, ev.KlineData.HighPrice, "k.h")
			_ = tryParseFloat(t, ev.KlineData.LowPrice, "k.l")
			_ = tryParseFloat(t, ev.KlineData.ClosePrice, "k.c")
			_ = tryParseFloat(t, ev.KlineData.Volume, "k.v")
			_ = tryParseFloat(t, ev.KlineData.CompletedTradeAmount, "k.q")
			_ = tryParseFloat(t, ev.KlineData.TakerCompletedTradeVolume, "k.V")
			_ = tryParseFloat(t, ev.KlineData.TakerTradeAmount, "k.Q")
			if ev.KlineData.KlineStartTime >= ev.KlineData.KlineEndTime {
				t.Errorf("kline times invalid")
			}
			if restValidationEnabled() {
				if klines, err := restKlines(context.Background(), symbol, interval, 1); err == nil && len(klines) > 0 {
					item := klines[len(klines)-1]
					if item.Close != nil {
						if closeREST, err1 := strconv.ParseFloat(*item.Close, 64); err1 == nil {
							if closeWS, err2 := strconv.ParseFloat(ev.KlineData.ClosePrice, 64); err2 == nil {
								assertWithinTolerancePercent(t, closeWS, closeREST, 5.0, "kline close vs REST")
							}
						}
					}
					if item.High != nil {
						if highREST, err1 := strconv.ParseFloat(*item.High, 64); err1 == nil {
							if highWS, err2 := strconv.ParseFloat(ev.KlineData.HighPrice, 64); err2 == nil {
								assertWithinTolerancePercent(t, highWS, highREST, 5.0, "kline high vs REST")
							}
						}
					}
					if item.Low != nil {
						if lowREST, err1 := strconv.ParseFloat(*item.Low, 64); err1 == nil {
							if lowWS, err2 := strconv.ParseFloat(ev.KlineData.LowPrice, 64); err2 == nil {
								assertWithinTolerancePercent(t, lowWS, lowREST, 5.0, "kline low vs REST")
							}
						}
					}
				} else if err != nil {
					t.Logf("rest kline validation skipped: %v", err)
				}
			}
			logJSON(t, "event.kline", ev)
		} else {
			t.Logf("no kline events recorded (acceptable)")
		}
		t.Logf("KlineEvent total recorded: %d", rec.count("kline"))
	})

	t.Run("TickerEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingA, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build ticker stream using typed params
		tp := optionsstreams.TickerEventStreamParams{Symbol: models.Symbol(symbol)}
		stream, err := optionsstreams.BuildTickerEventStream(0, tp.Values())
		if err != nil {
			t.Fatalf("build ticker stream: %v", err)
		}
		var subCb5 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("ticker")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb5)
		var unsubCb5 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb5)
		}()
		_ = rec.waitForMin("ticker", start+1, eventWait())
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
				}
			}
			logJSON(t, "event.ticker", ev)
		} else {
			t.Logf("no ticker events recorded (acceptable)")
		}
		t.Logf("TickerEvent total recorded: %d", rec.count("ticker"))
	})

	t.Run("TickerByUnderlyingEvent", func(t *testing.T) {
		exps, err := getActiveExpirationDates(underlyingA)
		if err != nil || len(exps) == 0 {
			t.Skipf("no expiration dates: %v", err)
		}
		// Build ticker-by-underlying stream with typed params
		tup := optionsstreams.TickerByUnderlyingEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(strings.ToLower(underlyingA)), ExpirationDate: models.ExpirationDate(exps[0])}
		stream, err := optionsstreams.BuildTickerByUnderlyingEventStream(0, tup.Values())
		if err != nil {
			t.Fatalf("build ticker-by-underlying stream: %v", err)
		}
		params := []string{stream}
		for i := 1; i < len(exps) && i < 3; i++ {
			p := optionsstreams.TickerByUnderlyingEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(strings.ToLower(underlyingA)), ExpirationDate: models.ExpirationDate(exps[i])}
			s, err := optionsstreams.BuildTickerByUnderlyingEventStream(0, p.Values())
			if err == nil {
				params = append(params, s)
			}
		}
		var subCb6 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("tickerByUnderlying")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: params}, &subCb6)
		var unsubCb6 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			if len(params) > 0 {
				_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: params}, &unsubCb6)
			}
		}()
		_ = rec.waitForMin("tickerByUnderlying", start+1, eventWaitLong())
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
			logJSON(t, "event.tickerByUnderlying", ev)
		} else {
			t.Logf("no tickerByUnderlying events recorded (acceptable)")
		}
		t.Logf("TickerByUnderlyingEvent total recorded: %d", rec.count("tickerByUnderlying"))
	})

	t.Run("TradeEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingB, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build trade (symbol variant) using typed params
		trp := optionsstreams.TradeEventStreamParams{Symbol: models.Symbol(symbol)}
		stream, err := optionsstreams.BuildTradeEventStream(0, trp.Values())
		if err != nil {
			t.Fatalf("build trade stream: %v", err)
		}
		var subCb7 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("trade")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb7)
		var unsubCb7 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb7)
		}()
		_ = rec.waitForMin("trade", start+1, eventWaitLong())
		if evs := rec.getTrade(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "trade" {
				t.Errorf("want e=trade got %s", ev.EventType)
			}
			assertOptionSymbolFormat(t, ev.OptionTradingSymbol)
			_ = tryParseFloat(t, ev.Price, "price")
			_ = tryParseFloat(t, ev.Quantity, "qty")
			if ev.TradeID == "" {
				t.Errorf("invalid trade id: %s", ev.TradeID)
			}
			if ev.TradeCompletedTimestamp <= 0 {
				t.Errorf("invalid trade ts: %d", ev.TradeCompletedTimestamp)
			}
			logJSON(t, "event.trade", ev)
		} else {
			t.Logf("no trade events recorded (acceptable)")
		}
	})

	t.Run("PartialDepthEvent", func(t *testing.T) {
		symbol, err := selectATMSymbol(underlyingB, "")
		if err != nil {
			t.Skipf("no active symbol: %v", err)
		}
		// Build partial depth stream using typed params (pattern 1 includes speed)
		pd := optionsstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symbol), Levels: models.DepthLevels20, Speed: models.DepthSpeed100ms}
		stream, err := optionsstreams.BuildPartialDepthEventStream(1, pd.Values())
		if err != nil {
			t.Fatalf("build partialDepth stream: %v", err)
		}
		var subCb8 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("partialDepth")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb8)
		var unsubCb8 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb8)
		}()
		_ = rec.waitForMin("partialDepth", start+1, eventWaitLong())
		if evs := rec.getPartialDepth(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "depth" {
				t.Errorf("want e=depth got %s", ev.EventType)
			}
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
			logJSON(t, "event.partialDepth", ev)
		} else {
			t.Logf("no partialDepth events recorded (acceptable)")
		}
		t.Logf("PartialDepthEvent total recorded: %d", rec.count("partialDepth"))
	})

	t.Run("NewSymbolInfoEvent", func(t *testing.T) {
		// Build correct stream name using builder (no params required)
		stream, err := optionsstreams.BuildNewSymbolInfoEventStream(0, nil)
		if err != nil {
			t.Fatalf("build newSymbolInfo stream: %v", err)
		}
		var subCb10 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		start := rec.count("newSymbolInfo")
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &subCb10)
		var unsubCb10 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: []string{stream}}, &unsubCb10)
		}()
		_ = rec.waitForMin("newSymbolInfo", start+1, eventWait())
		if evs := rec.getNewSymbolInfo(); len(evs) > 0 {
			ev := evs[len(evs)-1]
			if ev.EventType != "option_pair" {
				t.Errorf("want e=option_pair got %s", ev.EventType)
			}
			assertOptionSymbolFormat(t, ev.TradingPairName)
			logJSON(t, "event.newSymbolInfo", ev)
		} else {
			t.Logf("no newSymbolInfo events recorded (acceptable)")
		}
	})

	t.Run("OpenInterestEvent", func(t *testing.T) {
		exps, err := getActiveExpirationDates(underlyingB)
		if err != nil || len(exps) == 0 {
			t.Skipf("no expiration dates: %v", err)
		}
		// Use the global recorder; do not override the handler again
		// Build primary stream using typed params
		op := optionsstreams.OpenInterestEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingBLower), ExpirationDate: models.ExpirationDate(exps[0])}
		stream, err := optionsstreams.BuildOpenInterestEventStream(0, op.Values())
		if err != nil {
			t.Fatalf("build openInterest stream: %v", err)
		}
		params := []string{stream}
		for i := 1; i < len(exps) && i < 3; i++ {
			p := optionsstreams.OpenInterestEventStreamParams{UnderlyingAsset: models.UnderlyingAsset(underlyingBLower), ExpirationDate: models.ExpirationDate(exps[i])}
			if s, err := optionsstreams.BuildOpenInterestEventStream(0, p.Values()); err == nil {
				params = append(params, s)
			}
		}
		var subCb9 func(context.Context, *models.SubscribeResponse, error) error = func(context.Context, *models.SubscribeResponse, error) error { return nil }
		_ = ch.Subscribe(context.Background(), &models.SubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: params}, &subCb9)
		var unsubCb9 func(context.Context, *models.UnsubscribeResponse, error) error = func(context.Context, *models.UnsubscribeResponse, error) error { return nil }
		defer func() {
			if len(params) > 0 {
				_ = ch.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: models.NewMessageIDInt64(time.Now().UnixMicro()), Params: params}, &unsubCb9)
			}
		}()
		// Open interest updates every ~60s; wait longer for the recorder to capture at least one
		_ = rec.waitForMin("openInterest", 1, eventWaitLong())
		if evs := rec.getOpenInterest(); len(evs) > 0 && evs[len(evs)-1] != nil {
			ev := evs[len(evs)-1]
			if ev == nil || len(*ev) == 0 {
				t.Errorf("openInterest event empty")
			} else {
				for _, it := range *ev {
					if it.EventType != "openInterest" {
						t.Errorf("want e=openInterest got %s", it.EventType)
					}
					assertOptionSymbolFormat(t, it.OptionSymbol)
					_ = tryParseFloat(t, it.OpenInterestInContracts, "openInterestContracts")
					_ = tryParseFloat(t, it.OpenInterestInUSDT, "openInterestUSDT")
				}
				logJSON(t, "event.openInterest", ev)
			}
		} else {
			t.Logf("no openInterest events recorded (acceptable)")
		}
		t.Logf("OpenInterestEvent count: %d", rec.count("openInterest"))
	})
}
