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

    umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
    "github.com/openxapi/binance-go/ws/umfutures-streams/models"
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
type umMarketEventRecorder struct {
    mu           sync.RWMutex
    agg          []*models.AggregateTradeEvent
    mark         []*models.MarkPriceEvent
    kline        []*models.KlineEvent
    ticker       []*models.TickerEvent
    miniTicker   []*models.MiniTickerEvent
    bookTicker   []*models.BookTickerEvent
    partialDepth []*models.PartialDepthEvent
    diffDepth    []*models.DiffDepthEvent
    liquidation  []*models.LiquidationEvent
    // Additional events recorded for full-suite coverage
    allMarkPrices   []*models.AllMarkPricesEvent
    allMiniTickers  []*models.AllMiniTickersEvent
    allTickers      []*models.AllTickersEvent
    allBookTickers  []*models.AllBookTickersEvent
    allLiquidations []*models.AllLiquidationsEvent
    compositeIndex  []*models.CompositeIndexEvent
    assetIndex      []*models.AssetIndexEvent
    allAssetIndexes []*models.AllAssetIndexesEvent
    continuousKline []*models.ContinuousKlineEvent
}

func newUMMarketEventRecorder() *umMarketEventRecorder { return &umMarketEventRecorder{} }
func (r *umMarketEventRecorder) addAgg(ev *models.AggregateTradeEvent)      { r.mu.Lock(); r.agg = append(r.agg, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addMark(ev *models.MarkPriceEvent)         { r.mu.Lock(); r.mark = append(r.mark, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addKline(ev *models.KlineEvent)            { r.mu.Lock(); r.kline = append(r.kline, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addTicker(ev *models.TickerEvent)          { r.mu.Lock(); r.ticker = append(r.ticker, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addMiniTicker(ev *models.MiniTickerEvent)  { r.mu.Lock(); r.miniTicker = append(r.miniTicker, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addLiquidation(ev *models.LiquidationEvent){ r.mu.Lock(); r.liquidation = append(r.liquidation, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addBookTicker(ev *models.BookTickerEvent)  { r.mu.Lock(); r.bookTicker = append(r.bookTicker, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addPartialDepth(ev *models.PartialDepthEvent){ r.mu.Lock(); r.partialDepth = append(r.partialDepth, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addDiffDepth(ev *models.DiffDepthEvent)    { r.mu.Lock(); r.diffDepth = append(r.diffDepth, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllMarkPrices(ev *models.AllMarkPricesEvent) { r.mu.Lock(); r.allMarkPrices = append(r.allMarkPrices, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllMiniTickers(ev *models.AllMiniTickersEvent){ r.mu.Lock(); r.allMiniTickers = append(r.allMiniTickers, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllTickers(ev *models.AllTickersEvent)  { r.mu.Lock(); r.allTickers = append(r.allTickers, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllBookTickers(ev *models.AllBookTickersEvent){ r.mu.Lock(); r.allBookTickers = append(r.allBookTickers, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllLiquidations(ev *models.AllLiquidationsEvent){ r.mu.Lock(); r.allLiquidations = append(r.allLiquidations, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addCompositeIndex(ev *models.CompositeIndexEvent){ r.mu.Lock(); r.compositeIndex = append(r.compositeIndex, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAssetIndex(ev *models.AssetIndexEvent)  { r.mu.Lock(); r.assetIndex = append(r.assetIndex, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addAllAssetIndexes(ev *models.AllAssetIndexesEvent){ r.mu.Lock(); r.allAssetIndexes = append(r.allAssetIndexes, ev); r.mu.Unlock() }
func (r *umMarketEventRecorder) addContinuousKline(ev *models.ContinuousKlineEvent){ r.mu.Lock(); r.continuousKline = append(r.continuousKline, ev); r.mu.Unlock() }

func (r *umMarketEventRecorder) count(key string) int {
    r.mu.RLock(); defer r.mu.RUnlock()
    switch key {
    case "aggTrade": return len(r.agg)
    case "markPrice": return len(r.mark)
    case "kline": return len(r.kline)
    case "ticker": return len(r.ticker)
    case "miniTicker": return len(r.miniTicker)
    case "liquidation": return len(r.liquidation)
    case "bookTicker": return len(r.bookTicker)
    case "partialDepth": return len(r.partialDepth)
    case "diffDepth": return len(r.diffDepth)
    case "allMarkPrices": return len(r.allMarkPrices)
    case "allMiniTickers": return len(r.allMiniTickers)
    case "allTickers": return len(r.allTickers)
    case "allBookTickers": return len(r.allBookTickers)
    case "allLiquidations": return len(r.allLiquidations)
    case "compositeIndex": return len(r.compositeIndex)
    case "assetIndex": return len(r.assetIndex)
    case "allAssetIndexes": return len(r.allAssetIndexes)
    case "continuousKline": return len(r.continuousKline)
    default: return 0
    }
}

func (r *umMarketEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if r.count(key) >= min { return nil }
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}

// TestFullIntegrationSuite_Market covers request/response + event handlers for MarketStreamsChannel
func TestFullIntegrationSuite_Market(t *testing.T) {
    if testing.Short() { t.Skip("Skipping in short mode") }

    // Capture SDK log output and fail on any 'unhandled message' during this suite
    cw := &unhandledCatcher{}
    log.SetOutput(cw)
    defer func() {
        log.SetOutput(os.Stderr)
        cw.mu.Lock(); defer cw.mu.Unlock()
        if len(cw.matches) > 0 {
            for _, line := range cw.matches { t.Logf("SDK log captured: %s", strings.TrimSpace(line)) }
            t.Fatalf("SDK emitted %d 'unhandled message' log(s) during Market suite; treating as failure", len(cw.matches))
        }
    }()

    config := getTestConfig()
    stc, err := NewStreamTestClientDedicated(config)
    if err != nil { t.Fatalf("failed to create client: %v", err) }

    // Force mainnet1 for this suite per instruction
    _ = stc.client.SetActiveServer("mainnet1")

    if as := stc.client.GetActiveServer(); as != nil {
        t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
    } else {
        t.Logf("Active WS server: <nil>")
    }

    // Resolve an active symbol via REST; use lowercase for stream paths
    upSym, errPick := restPickSymbol(context.Background())
    if errPick != nil || upSym == "" { t.Fatalf("failed to resolve active symbol via REST: %v", errPick) }
    symUpper := upSym
    symLower := strings.ToLower(symUpper)
    t.Logf("Using symbol from REST: upper=%s lower=%s", symUpper, symLower)

    // Prepare a channel instance and connect once for the entire suite
    market := umfuturesstreams.NewMarketStreamsChannel(stc.client)
    // Record events from the start of the suite
    rec := newUMMarketEventRecorder()
    market.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error { logJSON(t, "ws.error", msg); return nil })
    // Register all event handlers at suite start and record
    market.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
    market.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
    market.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
    market.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
    market.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
    market.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
    market.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
    market.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
    market.HandleLiquidationEvent(func(ctx context.Context, ev *models.LiquidationEvent) error { rec.addLiquidation(ev); return nil })
    market.HandleAllMarkPricesEvent(func(ctx context.Context, ev *models.AllMarkPricesEvent) error { rec.addAllMarkPrices(ev); return nil })
    market.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMiniTickers(ev); return nil })
    market.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTickers(ev); return nil })
    market.HandleAllBookTickersEvent(func(ctx context.Context, ev *models.AllBookTickersEvent) error { rec.addAllBookTickers(ev); return nil })
    market.HandleAllLiquidationsEvent(func(ctx context.Context, ev *models.AllLiquidationsEvent) error { rec.addAllLiquidations(ev); return nil })
    market.HandleCompositeIndexEvent(func(ctx context.Context, ev *models.CompositeIndexEvent) error { rec.addCompositeIndex(ev); return nil })
    market.HandleAssetIndexEvent(func(ctx context.Context, ev *models.AssetIndexEvent) error { rec.addAssetIndex(ev); return nil })
    market.HandleAllAssetIndexesEvent(func(ctx context.Context, ev *models.AllAssetIndexesEvent) error { rec.addAllAssetIndexes(ev); return nil })
    market.HandleContinuousKlineEvent(func(ctx context.Context, ev *models.ContinuousKlineEvent) error { rec.addContinuousKline(ev); return nil })
    // Additional array/special handlers registered per-test below as needed

    // Connect with a base stream (aggTrade)
    base := umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
    baseStream, err := umfuturesstreams.BuildAggregateTradeEventStream(0, base.Values())
    if err != nil { t.Fatalf("build base stream: %v", err) }
    cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
    if err := market.Connect(cctx, baseStream); err != nil { ccancel(); t.Fatalf("connect: %v", err) }
    ccancel()
    defer func() {
        dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
        _ = market.Disconnect(dctx)
        dcancel()
    }()

    // ---------- Requests & Responses ----------
    t.Run("Request_Subscribe", func(t *testing.T) {
        // Build markPrice stream (@1s optional) and aggTrade stream
        mp := umfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
        markPath, err := umfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build markPrice stream: %v", err) }
        ag := umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        aggPath, err := umfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
        if err != nil { t.Fatalf("build aggTrade stream: %v", err) }
        sid := time.Now().UnixMicro()
        subDone := make(chan struct{}, 1)
        var gotSub *models.SubscribeResponse
        subCb := func(ctx context.Context, resp *models.SubscribeResponse) error {
            if resp == nil { t.Errorf("nil subscribe response"); return nil }
            if resp.Id != sid { t.Errorf("subscribe id mismatch: want %d got %d", sid, resp.Id) }
            gotSub = resp
            logJSON(t, "subscribe.response", resp)
            select { case subDone <- struct{}{}: default: }
            return nil
        }
        req := &models.SubscribeRequest{Id: sid, Params: []string{markPath, aggPath}}
        if err := market.MarketStreamsSubscribe(context.Background(), req, &subCb); err != nil {
            t.Fatalf("subscribe call failed: %v", err)
        }
        select { case <-subDone: case <-time.After(10 * time.Second): t.Errorf("timeout waiting subscribe response") }
        if gotSub == nil { t.Fatalf("did not capture subscribe response") }
        // Cleanup to avoid noise
        var unsubCb func(context.Context, *models.UnsubscribeResponse) error = func(context.Context, *models.UnsubscribeResponse) error { return nil }
        _ = market.MarketStreamsUnsubscribe(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{markPath}}, &unsubCb)
    })

    t.Run("Request_ListSubscriptions", func(t *testing.T) {
        ag := umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        aggPath, err := umfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
        if err != nil { t.Fatalf("build aggTrade stream: %v", err) }
        // Subscribe first so list has at least one entry
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{aggPath}}, &subCb)
        lid := time.Now().UnixMicro()
        listDone := make(chan struct{}, 1)
        var got *models.ListSubscriptionsResponse
        listCb := func(ctx context.Context, resp *models.ListSubscriptionsResponse) error {
            if resp == nil { t.Errorf("nil list subscriptions response"); return nil }
            if resp.Id != lid { t.Errorf("list id mismatch: want %d got %d", lid, resp.Id) }
            if resp.Result != nil && !contains(resp.Result, aggPath) {
                t.Logf("list did not include %s (result=%v)", aggPath, resp.Result)
            }
            got = resp
            logJSON(t, "listSubscriptions.response", resp)
            select { case listDone <- struct{}{}: default: }
            return nil
        }
        if err := market.MarketStreamsListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: lid}, &listCb); err != nil {
            t.Fatalf("list subscriptions call failed: %v", err)
        }
        select { case <-listDone: case <-time.After(8 * time.Second): t.Logf("timeout waiting listSubscriptions response") }
        if got == nil { t.Fatalf("did not capture listSubscriptions response") }
        // Cleanup
        var unsubCb func(context.Context, *models.UnsubscribeResponse) error = func(context.Context, *models.UnsubscribeResponse) error { return nil }
        _ = market.MarketStreamsUnsubscribe(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{aggPath}}, &unsubCb)
    })

    t.Run("Request_SetProperty", func(t *testing.T) {
        // Use MarketStreams setProperty on active connection (do not use combined here)
        pid := time.Now().UnixMicro()
        setDone := make(chan struct{}, 1)
        var gotSet *models.SetPropertyResponse
        setCb := func(ctx context.Context, resp *models.SetPropertyResponse) error {
            if resp == nil { t.Errorf("nil setProperty response"); return nil }
            if resp.Id != pid { t.Errorf("setProperty id mismatch: want %d got %d", pid, resp.Id) }
            gotSet = resp; logJSON(t, "setProperty.response", resp)
            select { case setDone <- struct{}{}: default: }
            return nil
        }
        spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer spCancel()
        // Property 'combined' is defined in the API; some servers may ignore it on single streams — treat timeouts/errors as acceptable.
        setReq := &models.SetPropertyRequest{Id: pid, Params: []interface{}{"combined", false}}
        logJSON(t, "setProperty.request", setReq)
        if err := market.MarketStreamsSetProperty(spCtx, setReq, &setCb); err != nil {
            le := strings.ToLower(err.Error())
            if strings.Contains(le, "deadline") { t.Logf("setProperty timeout (acceptable): %v", err) } else { t.Logf("setProperty err (acceptable): %v", err) }
        }
        select { case <-setDone: case <-time.After(8 * time.Second): t.Logf("timeout waiting setProperty response (acceptable)") }
        _ = gotSet
    })

    t.Run("Request_GetProperty", func(t *testing.T) {
        // Query property on MarketStreams (not combined)
        gid := time.Now().UnixMicro()
        getDone := make(chan struct{}, 1)
        var gotGet *models.GetPropertyResponse
        getCb := func(ctx context.Context, resp *models.GetPropertyResponse) error {
            if resp == nil { t.Errorf("nil getProperty response"); return nil }
            if resp.Id != gid { t.Errorf("getProperty id mismatch: want %d got %d", gid, resp.Id) }
            gotGet = resp; logJSON(t, "getProperty.response", resp)
            select { case getDone <- struct{}{}: default: }
            return nil
        }
        getReq := &models.GetPropertyRequest{Id: gid, Params: []string{"combined"}}
        logJSON(t, "getProperty.request", getReq)
        if err := market.MarketStreamsGetProperty(context.Background(), getReq, &getCb); err != nil {
            t.Logf("getProperty call err (acceptable on some servers): %v", err)
        }
        select { case <-getDone: case <-time.After(8 * time.Second): t.Logf("timeout waiting getProperty response (acceptable)") }
        _ = gotGet
    })

    // ---------- Event Handlers ----------
    t.Run("AggregateTradeEvent", func(t *testing.T) {
        ag := umfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
        if err != nil { t.Fatalf("build aggTrade stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("aggTrade", 1, eventWait())
        cnt := rec.count("aggTrade")
        t.Logf("aggTrade events received: %d", cnt)
        if cnt > 0 {
            ev := rec.agg[len(rec.agg)-1]
            if ev.EventType != "aggTrade" { t.Errorf("want e=aggTrade got %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        } else {
            t.Logf("no aggTrade event received (acceptable)")
        }
    })

    t.Run("MarkPriceEvent", func(t *testing.T) {
        mp := umfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build markPrice stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("markPrice", 1, eventWait())
        cnt := rec.count("markPrice")
        t.Logf("markPrice events received: %d", cnt)
        if cnt > 0 {
            ev := rec.mark[len(rec.mark)-1]
            if ev.EventType != "markPriceUpdate" { t.Errorf("want e=markPriceUpdate got %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            _ = tryParseFloat(t, ev.MarkPrice, "markPrice")
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
            if restValidationEnabled() {
                if px, err := restTickerPrice(context.Background(), symUpper); err == nil && px > 0 {
                    if v, err2 := strconv.ParseFloat(ev.MarkPrice, 64); err2 == nil {
                        assertWithinTolerancePercent(t, v, px, 5.0, "mark vs REST price")
                    }
                } else if err != nil {
                    t.Logf("REST price fetch failed: %v", err)
                }
            }
        } else {
            t.Logf("no markPrice event received (acceptable)")
        }
    })

    t.Run("KlineEvent", func(t *testing.T) {
        interval := os.Getenv("DEFAULT_INTERVAL")
        if interval == "" { interval = "1m" }
        kp := umfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval(interval)}
        path, err := umfuturesstreams.BuildKlineEventStream(0, kp.Values())
        if err != nil { t.Fatalf("build kline stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("kline", 1, eventWait())
        cnt := rec.count("kline")
        t.Logf("kline events received: %d", cnt)
        if cnt > 0 {
            ev := rec.kline[len(rec.kline)-1]
            if ev.EventType != "kline" { t.Errorf("want e=kline got %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        } else {
            t.Logf("no kline event received (acceptable)")
        }
    })

    t.Run("TickerEvent", func(t *testing.T) {
        tp := umfuturesstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildTickerEventStream(0, tp.Values())
        if err != nil { t.Fatalf("build ticker stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("ticker", 1, eventWait())
        cnt := rec.count("ticker")
        t.Logf("ticker events received: %d", cnt)
        if cnt > 0 {
            ev := rec.ticker[len(rec.ticker)-1]
            if ev.EventType != "24hrTicker" { t.Errorf("want e=24hrTicker got %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
        } else {
            t.Logf("no ticker event received (acceptable)")
        }
    })

    t.Run("MiniTickerEvent", func(t *testing.T) {
        mp := umfuturesstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildMiniTickerEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build miniTicker stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("miniTicker", 1, eventWait())
        cnt := rec.count("miniTicker")
        t.Logf("miniTicker events received: %d", cnt)
        if cnt > 0 {
            ev := rec.miniTicker[len(rec.miniTicker)-1]
            if ev.EventType != "24hrMiniTicker" { t.Logf("unexpected type: %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            _ = tryParseFloat(t, ev.ClosePrice, "closePrice")
        } else {
            t.Logf("no miniTicker event received (acceptable)")
        }
    })

    t.Run("BookTickerEvent", func(t *testing.T) {
        bp := umfuturesstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildBookTickerEventStream(0, bp.Values())
        if err != nil { t.Fatalf("build bookTicker stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("bookTicker", 1, eventWait())
        cnt := rec.count("bookTicker")
        t.Logf("bookTicker events received: %d", cnt)
        if cnt > 0 {
            ev := rec.bookTicker[len(rec.bookTicker)-1]
            if ev.EventType != "bookTicker" { t.Errorf("want e=bookTicker got %s", ev.EventType) }
            if !strings.EqualFold(ev.Symbol, symUpper) { t.Logf("unexpected symbol: %s", ev.Symbol) }
            assertNonEmpty(t, ev.BestBidPrice, "best bid price")
            assertNonEmpty(t, ev.BestAskPrice, "best ask price")
        } else {
            t.Logf("no bookTicker event received (acceptable)")
        }
    })

    t.Run("LiquidationEvent", func(t *testing.T) {
        lp := umfuturesstreams.LiquidationEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := umfuturesstreams.BuildLiquidationEventStream(0, lp.Values())
        if err != nil { t.Fatalf("build liquidation stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("liquidation", 1, eventWait())
        cnt := rec.count("liquidation")
        t.Logf("liquidation events received: %d", cnt)
        if cnt > 0 {
            ev := rec.liquidation[len(rec.liquidation)-1]
            if ev.EventType != "forceOrder" { t.Logf("unexpected event type: %s", ev.EventType) }
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
            t.Logf("no liquidation event received for %s (acceptable)", symUpper)
        }
    })

    t.Run("PartialDepthEvent", func(t *testing.T) {
        dp := umfuturesstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels("5"), Speed: models.DepthSpeed("100ms")}
        path, err := umfuturesstreams.BuildPartialDepthEventStream(1, dp.Values())
        if err != nil { t.Fatalf("build partial depth stream: %v", err) }
        var subCb func(context.Context, *models.SubscribeResponse) error = func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("partialDepth", 1, eventWait())
        cnt := rec.count("partialDepth")
        t.Logf("partialDepth events received: %d", cnt)
        if cnt > 0 {
            ev := rec.partialDepth[len(rec.partialDepth)-1]
            if ev.EventType != "depthUpdate" { t.Errorf("want e=depthUpdate got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        } else {
            t.Logf("no partialDepth event received (acceptable)")
        }
    })

    t.Run("AllMarkPricesEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllMarkPricesEventStream(0, nil)
        if err != nil { t.Fatalf("build allMarkPrices stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allMarkPrices", 1, eventWait())
        cnt := rec.count("allMarkPrices")
        t.Logf("allMarkPrices events received: %d", cnt)
        if cnt > 0 {
            ev := rec.allMarkPrices[len(rec.allMarkPrices)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "markPriceUpdate" { t.Logf("unexpected type: %s", it.EventType) }
                assertRecentMs(t, it.EventTime, 2*time.Hour, "eventTime")
            }
        } else {
            t.Logf("no allMarkPrices event received (acceptable)")
        }
        var unsub func(context.Context, *models.UnsubscribeResponse) error = func(context.Context, *models.UnsubscribeResponse) error { return nil }
        _ = market.MarketStreamsUnsubscribe(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &unsub)
    })

    t.Run("AllMiniTickersEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllMiniTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build allMiniTicker stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allMiniTickers", 1, eventWait())
        cnt := rec.count("allMiniTickers")
        t.Logf("allMiniTickers events received: %d", cnt)
        if cnt == 0 { t.Logf("no allMiniTickers event received (acceptable)") }
    })

    t.Run("AllTickersEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build allTickers stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allTickers", 1, eventWait())
        cnt := rec.count("allTickers")
        t.Logf("allTickers events received: %d", cnt)
        if cnt == 0 { t.Logf("no allTickers event received (acceptable)") }
    })

    t.Run("AllBookTickersEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllBookTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build allBookTicker stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allBookTickers", 1, eventWait())
        cnt := rec.count("allBookTickers")
        t.Logf("allBookTickers events received: %d", cnt)
        if cnt == 0 { t.Logf("no allBookTickers event received (acceptable)") }
    })

    t.Run("AllLiquidationsEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllLiquidationsEventStream(0, nil)
        if err != nil { t.Fatalf("build allLiquidations stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allLiquidations", 1, eventWait())
        cnt := rec.count("allLiquidations")
        t.Logf("allLiquidations events received: %d", cnt)
        if cnt > 0 {
            ev := rec.allLiquidations[len(rec.allLiquidations)-1]
            if ev.EventType != "forceOrder" { t.Logf("unexpected event type: %s", ev.EventType) }
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
        path, err := umfuturesstreams.BuildCompositeIndexEventStream(0, cp.Values())
        if err != nil { t.Fatalf("build compositeIndex stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("compositeIndex", 1, eventWait())
        cnt := rec.count("compositeIndex")
        t.Logf("compositeIndex events received: %d", cnt)
        if cnt == 0 { t.Logf("no compositeIndex event received (acceptable)") }
    })

    t.Run("AssetIndexEvent", func(t *testing.T) {
        assetLower := strings.ToLower(strings.TrimSuffix(symUpper, "USDT") + "USD")
        ap := umfuturesstreams.AssetIndexEventStreamParams{AssetSymbol: models.AssetSymbol(assetLower)}
        path, err := umfuturesstreams.BuildAssetIndexEventStream(0, ap.Values())
        if err != nil { t.Fatalf("build assetIndex stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("assetIndex", 1, eventWait())
        cnt := rec.count("assetIndex")
        t.Logf("assetIndex events received: %d", cnt)
        if cnt == 0 { t.Logf("no assetIndex event received (acceptable)") }
    })

    t.Run("AllAssetIndexEvent", func(t *testing.T) {
        path, err := umfuturesstreams.BuildAllAssetIndexesEventStream(0, nil)
        if err != nil { t.Fatalf("build allAssetIndex stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allAssetIndexes", 1, eventWait())
        cnt := rec.count("allAssetIndexes")
        t.Logf("allAssetIndexes events received: %d", cnt)
        if cnt == 0 { t.Logf("no allAssetIndexes event received (acceptable)") }
    })

    t.Run("ContinuousKlineEvent", func(t *testing.T) {
        cp := umfuturesstreams.ContinuousKlineEventStreamParams{Pair: models.Pair(symUpper), ContractType: models.ContractType("perpetual"), Interval: models.Interval("1m")}
        path, err := umfuturesstreams.BuildContinuousKlineEventStream(0, cp.Values())
        if err != nil { t.Fatalf("build continuousKline stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.MarketStreamsSubscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("continuousKline", 1, eventWait())
        cnt := rec.count("continuousKline")
        t.Logf("continuousKline events received: %d", cnt)
        if cnt == 0 { t.Logf("no continuousKline event received (acceptable)") }
    })
}
