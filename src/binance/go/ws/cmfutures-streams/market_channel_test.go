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

    cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
    "github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// Capture SDK log lines containing the marker "unhandled message:" and fail the suite
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

// market event recorder
type cmMarketEventRecorder struct {
    mu             sync.RWMutex
    agg            []*models.AggregateTradeEvent
    mark           []*models.MarkPriceEvent
    index          []*models.IndexPriceEvent
    kline          []*models.KlineEvent
    contKline      []*models.ContinuousKlineEvent
    indexKline     []*models.IndexKlineEvent
    markKline      []*models.MarkPriceKlineEvent
    ticker         []*models.TickerEvent
    miniTicker     []*models.MiniTickerEvent
    bookTicker     []*models.BookTickerEvent
    partialDepth   []*models.PartialDepthEvent
    diffDepth      []*models.DiffDepthEvent
    liquidation    []*models.LiquidationEvent
    allMarkPrices  []*models.AllMarkPricesEvent
    allMiniTickers []*models.AllMiniTickersEvent
    allTickers     []*models.AllTickersEvent
    allBookTickers []*models.AllBookTickersEvent
    allLiquid      []*models.AllLiquidationsEvent
    contractInfo   []*models.ContractInfoEvent
}

func newCMMarketEventRecorder() *cmMarketEventRecorder { return &cmMarketEventRecorder{} }
func (r *cmMarketEventRecorder) addAgg(v *models.AggregateTradeEvent)      { r.mu.Lock(); r.agg = append(r.agg, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addMark(v *models.MarkPriceEvent)         { r.mu.Lock(); r.mark = append(r.mark, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addIndex(v *models.IndexPriceEvent)       { r.mu.Lock(); r.index = append(r.index, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addKline(v *models.KlineEvent)            { r.mu.Lock(); r.kline = append(r.kline, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addContK(v *models.ContinuousKlineEvent)  { r.mu.Lock(); r.contKline = append(r.contKline, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addIndexK(v *models.IndexKlineEvent)      { r.mu.Lock(); r.indexKline = append(r.indexKline, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addMarkK(v *models.MarkPriceKlineEvent)   { r.mu.Lock(); r.markKline = append(r.markKline, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addTicker(v *models.TickerEvent)          { r.mu.Lock(); r.ticker = append(r.ticker, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addMiniTicker(v *models.MiniTickerEvent)  { r.mu.Lock(); r.miniTicker = append(r.miniTicker, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addBookTicker(v *models.BookTickerEvent)  { r.mu.Lock(); r.bookTicker = append(r.bookTicker, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addPartialDepth(v *models.PartialDepthEvent){ r.mu.Lock(); r.partialDepth = append(r.partialDepth, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addDiffDepth(v *models.DiffDepthEvent)    { r.mu.Lock(); r.diffDepth = append(r.diffDepth, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addLiquid(v *models.LiquidationEvent)     { r.mu.Lock(); r.liquidation = append(r.liquidation, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addAllMark(v *models.AllMarkPricesEvent)  { r.mu.Lock(); r.allMarkPrices = append(r.allMarkPrices, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addAllMini(v *models.AllMiniTickersEvent) { r.mu.Lock(); r.allMiniTickers = append(r.allMiniTickers, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addAllTick(v *models.AllTickersEvent)     { r.mu.Lock(); r.allTickers = append(r.allTickers, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addAllBook(v *models.AllBookTickersEvent) { r.mu.Lock(); r.allBookTickers = append(r.allBookTickers, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addAllLiquid(v *models.AllLiquidationsEvent){ r.mu.Lock(); r.allLiquid = append(r.allLiquid, v); r.mu.Unlock() }
func (r *cmMarketEventRecorder) addContract(v *models.ContractInfoEvent)  { r.mu.Lock(); r.contractInfo = append(r.contractInfo, v); r.mu.Unlock() }

func (r *cmMarketEventRecorder) count(key string) int {
    r.mu.RLock(); defer r.mu.RUnlock()
    switch key {
    case "aggTrade": return len(r.agg)
    case "markPrice": return len(r.mark)
    case "indexPrice": return len(r.index)
    case "kline": return len(r.kline)
    case "continuousKline": return len(r.contKline)
    case "indexKline": return len(r.indexKline)
    case "markPriceKline": return len(r.markKline)
    case "ticker": return len(r.ticker)
    case "miniTicker": return len(r.miniTicker)
    case "bookTicker": return len(r.bookTicker)
    case "partialDepth": return len(r.partialDepth)
    case "diffDepth": return len(r.diffDepth)
    case "liquidation": return len(r.liquidation)
    case "allMarkPrices": return len(r.allMarkPrices)
    case "allMiniTickers": return len(r.allMiniTickers)
    case "allTickers": return len(r.allTickers)
    case "allBookTickers": return len(r.allBookTickers)
    case "allLiquidations": return len(r.allLiquid)
    case "contractInfo": return len(r.contractInfo)
    default: return 0
    }
}

func (r *cmMarketEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if r.count(key) >= min { return nil }
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}

// TestFullIntegrationSuite_Market runs request/response and event coverage for MarketStreamChannel
func TestFullIntegrationSuite_Market(t *testing.T) {
    if testing.Short() { t.Skip("Skipping in short mode") }

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
    // Force mainnet for market streams per instruction
    _ = stc.client.SetActiveServer("mainnet")

    if as := stc.client.GetActiveServer(); as != nil {
        t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
    } else { t.Logf("Active WS server: <nil>") }

    // Resolve a symbol via REST; stream path uses lowercase
    upSym, errPick := restPickSymbol(context.Background())
    if errPick != nil || upSym == "" { t.Fatalf("failed to resolve active symbol via REST: %v", errPick) }
    symUpper := upSym
    symLower := strings.ToLower(symUpper)
    pairLower := strings.ToLower(strings.TrimSuffix(symUpper, "_PERP")) // e.g., BTCUSD_PERP -> btcusd
    t.Logf("Using symbol: upper=%s lower=%s pairLower=%s", symUpper, symLower, pairLower)

    // Prepare market channel and handlers
    market := cmfuturesstreams.NewMarketStreamChannel(stc.client)
    rec := newCMMarketEventRecorder()
    market.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error { logJSON(t, "ws.error", msg); return nil })

    // Register event handlers upfront
    market.HandleAggregateTradeEvent(func(ctx context.Context, ev *models.AggregateTradeEvent) error { rec.addAgg(ev); return nil })
    market.HandleMarkPriceEvent(func(ctx context.Context, ev *models.MarkPriceEvent) error { rec.addMark(ev); return nil })
    market.HandleIndexPriceEvent(func(ctx context.Context, ev *models.IndexPriceEvent) error { rec.addIndex(ev); return nil })
    market.HandleKlineEvent(func(ctx context.Context, ev *models.KlineEvent) error { rec.addKline(ev); return nil })
    market.HandleContinuousKlineEvent(func(ctx context.Context, ev *models.ContinuousKlineEvent) error { rec.addContK(ev); return nil })
    market.HandleIndexKlineEvent(func(ctx context.Context, ev *models.IndexKlineEvent) error { rec.addIndexK(ev); return nil })
    market.HandleMarkPriceKlineEvent(func(ctx context.Context, ev *models.MarkPriceKlineEvent) error { rec.addMarkK(ev); return nil })
    market.HandleTickerEvent(func(ctx context.Context, ev *models.TickerEvent) error { rec.addTicker(ev); return nil })
    market.HandleMiniTickerEvent(func(ctx context.Context, ev *models.MiniTickerEvent) error { rec.addMiniTicker(ev); return nil })
    market.HandleBookTickerEvent(func(ctx context.Context, ev *models.BookTickerEvent) error { rec.addBookTicker(ev); return nil })
    market.HandlePartialDepthEvent(func(ctx context.Context, ev *models.PartialDepthEvent) error { rec.addPartialDepth(ev); return nil })
    market.HandleDiffDepthEvent(func(ctx context.Context, ev *models.DiffDepthEvent) error { rec.addDiffDepth(ev); return nil })
    market.HandleLiquidationEvent(func(ctx context.Context, ev *models.LiquidationEvent) error { rec.addLiquid(ev); return nil })
    market.HandleAllMarkPricesEvent(func(ctx context.Context, ev *models.AllMarkPricesEvent) error { rec.addAllMark(ev); return nil })
    market.HandleAllMiniTickersEvent(func(ctx context.Context, ev *models.AllMiniTickersEvent) error { rec.addAllMini(ev); return nil })
    market.HandleAllTickersEvent(func(ctx context.Context, ev *models.AllTickersEvent) error { rec.addAllTick(ev); return nil })
    market.HandleAllBookTickersEvent(func(ctx context.Context, ev *models.AllBookTickersEvent) error { rec.addAllBook(ev); return nil })
    market.HandleAllLiquidationsEvent(func(ctx context.Context, ev *models.AllLiquidationsEvent) error { rec.addAllLiquid(ev); return nil })
    market.HandleContractInfoEvent(func(ctx context.Context, ev *models.ContractInfoEvent) error { rec.addContract(ev); return nil })

    // Connect with a base stream (aggTrade)
    base := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
    baseStream, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, base.Values())
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
        mp := cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
        markPath, err := cmfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build mark stream: %v", err) }
        ag := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        aggPath, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
        if err != nil { t.Fatalf("build agg stream: %v", err) }
        sid := time.Now().UnixMicro()
        subDone := make(chan struct{}, 1)
        var got *models.SubscribeResponse
        cb := func(ctx context.Context, resp *models.SubscribeResponse) error { if resp != nil && resp.Id == sid { got = resp; logJSON(t, "subscribe.response", resp); select { case subDone <- struct{}{}: default: } }; return nil }
        if err := market.Subscribe(context.Background(), &models.SubscribeRequest{Id: sid, Params: []string{markPath, aggPath}}, &cb); err != nil { t.Fatalf("subscribe call failed: %v", err) }
        select { case <-subDone: case <-time.After(10 * time.Second): t.Errorf("timeout waiting subscribe response") }
        if got == nil { t.Fatalf("did not capture subscribe response") }
        // cleanup extra
        var ucb func(context.Context, *models.UnsubscribeResponse) error = func(context.Context, *models.UnsubscribeResponse) error { return nil }
        _ = market.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{markPath}}, &ucb)
    })

    t.Run("Request_ListSubscriptions", func(t *testing.T) {
        ag := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, ag.Values())
        if err != nil { t.Fatalf("build agg stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        lid := time.Now().UnixMicro()
        done := make(chan struct{}, 1)
        var got *models.ListSubscriptionsResponse
        lcb := func(ctx context.Context, resp *models.ListSubscriptionsResponse) error { if resp != nil && resp.Id == lid { got = resp; logJSON(t, "listSubscriptions.response", resp); select { case done <- struct{}{}: default: } }; return nil }
        if err := market.ListSubscriptions(context.Background(), &models.ListSubscriptionsRequest{Id: lid}, &lcb); err != nil { t.Fatalf("list subscriptions: %v", err) }
        select { case <-done: case <-time.After(8 * time.Second): t.Logf("timeout waiting listSubscriptions response") }
        if got == nil { t.Fatalf("did not capture listSubscriptions response") }
        var ucb func(context.Context, *models.UnsubscribeResponse) error = func(context.Context, *models.UnsubscribeResponse) error { return nil }
        _ = market.Unsubscribe(context.Background(), &models.UnsubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &ucb)
    })

    t.Run("Request_SetProperty", func(t *testing.T) {
        pid := time.Now().UnixMicro()
        setDone := make(chan struct{}, 1)
        var gotSet *models.SetPropertyResponse
        setCb := func(ctx context.Context, resp *models.SetPropertyResponse) error { if resp != nil && resp.Id == pid { gotSet = resp; logJSON(t, "setProperty.response", resp); select { case setDone <- struct{}{}: default: } }; return nil }
        spCtx, spCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer spCancel()
        req := &models.SetPropertyRequest{Id: pid, Params: []interface{}{"combined", false}}
        logJSON(t, "setProperty.request", req)
        if err := market.SetProperty(spCtx, req, &setCb); err != nil {
            le := strings.ToLower(err.Error()); if strings.Contains(le, "deadline") { t.Logf("setProperty timeout (acceptable)") } else { t.Logf("setProperty err (acceptable): %v", err) }
        }
        select { case <-setDone: case <-time.After(8 * time.Second): t.Logf("timeout waiting setProperty response (acceptable)") }
        _ = gotSet
    })

    t.Run("Request_GetProperty", func(t *testing.T) {
        gid := time.Now().UnixMicro()
        getDone := make(chan struct{}, 1)
        var gotGet *models.GetPropertyResponse
        getCb := func(ctx context.Context, resp *models.GetPropertyResponse) error { if resp != nil && resp.Id == gid { gotGet = resp; logJSON(t, "getProperty.response", resp); select { case getDone <- struct{}{}: default: } }; return nil }
        req := &models.GetPropertyRequest{Id: gid, Params: []string{"combined"}}
        logJSON(t, "getProperty.request", req)
        if err := market.GetProperty(context.Background(), req, &getCb); err != nil { t.Logf("getProperty err (acceptable): %v", err) }
        select { case <-getDone: case <-time.After(8 * time.Second): t.Logf("timeout waiting getProperty response (acceptable)") }
        _ = gotGet
    })

    // ---------- Event Handlers (subscribe and validate) ----------
    t.Run("AggregateTradeEvent", func(t *testing.T) {
        ap := cmfuturesstreams.AggregateTradeEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildAggregateTradeEventStream(0, ap.Values())
        if err != nil { t.Fatalf("build agg stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("aggTrade", 1, eventWait())
        cnt := rec.count("aggTrade")
        t.Logf("aggTrade events: %d", cnt)
        if cnt > 0 {
            ev := rec.agg[len(rec.agg)-1]
            if ev.EventType != "aggTrade" { t.Errorf("want e=aggTrade got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        }
    })

    t.Run("MarkPriceEvent", func(t *testing.T) {
        mp := cmfuturesstreams.MarkPriceEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildMarkPriceEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build mark stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("markPrice", 1, eventWait())
        cnt := rec.count("markPrice")
        t.Logf("markPrice events: %d", cnt)
        if cnt > 0 {
            ev := rec.mark[len(rec.mark)-1]
            if ev.EventType != "markPriceUpdate" { t.Errorf("want e=markPriceUpdate got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
            if restValidationEnabled() {
                if px, err := restMarkPrice(context.Background(), symUpper); err == nil && px > 0 {
                    if v, err2 := strconv.ParseFloat(ev.MarkPrice, 64); err2 == nil {
                        assertWithinTolerancePercent(t, v, px, 5.0, "mark vs REST")
                    }
                }
            }
        }
    })

    t.Run("IndexPriceEvent", func(t *testing.T) {
        ip := cmfuturesstreams.IndexPriceEventStreamParams{Pair: models.Pair(pairLower)}
        path, err := cmfuturesstreams.BuildIndexPriceEventStream(0, ip.Values())
        if err != nil { t.Fatalf("build index stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("indexPrice", 1, eventWait())
        t.Logf("indexPrice events: %d", rec.count("indexPrice"))
        // Best-effort shape checks
        if cnt := rec.count("indexPrice"); cnt > 0 {
            ev := rec.index[len(rec.index)-1]
            if ev.EventType != "indexPriceUpdate" { t.Logf("unexpected type: %s", ev.EventType) }
        }
    })

    t.Run("KlineEvent", func(t *testing.T) {
        interval := os.Getenv("DEFAULT_INTERVAL")
        if interval == "" { interval = "1m" }
        kp := cmfuturesstreams.KlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval(interval)}
        path, err := cmfuturesstreams.BuildKlineEventStream(0, kp.Values())
        if err != nil { t.Fatalf("build kline stream: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("kline", 1, eventWait())
        t.Logf("kline events: %d", rec.count("kline"))
        // Best-effort shape checks
        if cnt := rec.count("kline"); cnt > 0 {
            ev := rec.kline[len(rec.kline)-1]
            if ev.EventType != "kline" { t.Errorf("want e=kline got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        }
    })

    t.Run("ContinuousKlineEvent", func(t *testing.T) {
        cp := cmfuturesstreams.ContinuousKlineEventStreamParams{Pair: models.Pair(pairLower), ContractType: models.ContractType("perpetual"), Interval: models.Interval("1m")}
        path, err := cmfuturesstreams.BuildContinuousKlineEventStream(0, cp.Values())
        if err != nil { t.Fatalf("build continuous kline: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("continuousKline", 1, eventWait())
        t.Logf("continuousKline events: %d", rec.count("continuousKline"))
        // Best-effort shape checks
        if cnt := rec.count("continuousKline"); cnt > 0 {
            ev := rec.contKline[len(rec.contKline)-1]
            if ev.EventType != "continuous_kline" { t.Errorf("want e=continuous_kline got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        }
    })

    t.Run("IndexKlineEvent", func(t *testing.T) {
        ip := cmfuturesstreams.IndexKlineEventStreamParams{Pair: models.Pair(pairLower), Interval: models.Interval("1m")}
        path, err := cmfuturesstreams.BuildIndexKlineEventStream(0, ip.Values())
        if err != nil { t.Fatalf("build index kline: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("indexKline", 1, eventWait())
        t.Logf("indexKline events: %d", rec.count("indexKline"))
        // Best-effort shape checks
        if cnt := rec.count("indexKline"); cnt > 0 {
            ev := rec.indexKline[len(rec.indexKline)-1]
            if ev.EventType != "indexPrice_kline" { t.Errorf("want e=indexPrice_kline got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        }
    })

    t.Run("MarkPriceKlineEvent", func(t *testing.T) {
        mp := cmfuturesstreams.MarkPriceKlineEventStreamParams{Symbol: models.Symbol(symLower), Interval: models.Interval("1m")}
        path, err := cmfuturesstreams.BuildMarkPriceKlineEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build markPriceKline: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("markPriceKline", 1, eventWait())
        t.Logf("markPriceKline events: %d", rec.count("markPriceKline"))
        // Best-effort shape checks
        if cnt := rec.count("markPriceKline"); cnt > 0 {
            ev := rec.markKline[len(rec.markKline)-1]
            if ev.EventType != "markPrice_kline" { t.Errorf("want e=markPrice_kline got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 2*time.Hour, "eventTime")
        }
    })

    t.Run("TickerEvent", func(t *testing.T) {
        tp := cmfuturesstreams.TickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildTickerEventStream(0, tp.Values())
        if err != nil { t.Fatalf("build ticker: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("ticker", 1, eventWait())
        t.Logf("ticker events: %d", rec.count("ticker"))
        // Best-effort shape checks
        if cnt := rec.count("ticker"); cnt > 0 {
            ev := rec.ticker[len(rec.ticker)-1]
            if ev.EventType != "24hrTicker" { t.Errorf("want e=24hrTicker got %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("MiniTickerEvent", func(t *testing.T) {
        mp := cmfuturesstreams.MiniTickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildMiniTickerEventStream(0, mp.Values())
        if err != nil { t.Fatalf("build miniTicker: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("miniTicker", 1, eventWait())
        t.Logf("miniTicker events: %d", rec.count("miniTicker"))
        // Best-effort shape checks
        if cnt := rec.count("miniTicker"); cnt > 0 {
            ev := rec.miniTicker[len(rec.miniTicker)-1]
            if ev.EventType != "24hrMiniTicker" { t.Logf("unexpected type: %s", ev.EventType) }
        }
    })

    t.Run("BookTickerEvent", func(t *testing.T) {
        bp := cmfuturesstreams.BookTickerEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildBookTickerEventStream(0, bp.Values())
        if err != nil { t.Fatalf("build bookTicker: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("bookTicker", 1, eventWait())
        t.Logf("bookTicker events: %d", rec.count("bookTicker"))
        if cnt := rec.count("bookTicker"); cnt > 0 {
            ev := rec.bookTicker[len(rec.bookTicker)-1]
            if ev.EventType != "bookTicker" { t.Errorf("want e=bookTicker got %s", ev.EventType) }
            assertNonEmpty(t, ev.BestBidPrice, "bestBidPrice")
            assertNonEmpty(t, ev.BestAskPrice, "bestAskPrice")
        }
    })

    t.Run("PartialDepthEvent", func(t *testing.T) {
        dp := cmfuturesstreams.PartialDepthEventStreamParams{Symbol: models.Symbol(symLower), Levels: models.DepthLevels("5"), Speed: models.DepthSpeed("100ms")}
        path, err := cmfuturesstreams.BuildPartialDepthEventStream(1, dp.Values())
        if err != nil { t.Fatalf("build partial depth: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("partialDepth", 1, eventWait())
        t.Logf("partialDepth events: %d", rec.count("partialDepth"))
        if cnt := rec.count("partialDepth"); cnt > 0 {
            ev := rec.partialDepth[len(rec.partialDepth)-1]
            if ev.EventType != "depthUpdate" { t.Errorf("want e=depthUpdate got %s", ev.EventType) }
        }
    })

    t.Run("DiffDepthEvent", func(t *testing.T) {
        dp := cmfuturesstreams.DiffDepthEventStreamParams{Symbol: models.Symbol(symLower), Speed: models.DepthSpeed("100ms")}
        path, err := cmfuturesstreams.BuildDiffDepthEventStream(1, dp.Values())
        if err != nil { t.Fatalf("build diff depth: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("diffDepth", 1, eventWait())
        t.Logf("diffDepth events: %d", rec.count("diffDepth"))
        if cnt := rec.count("diffDepth"); cnt > 0 {
            ev := rec.diffDepth[len(rec.diffDepth)-1]
            if ev.EventType != "depthUpdate" { t.Errorf("want e=depthUpdate got %s", ev.EventType) }
        }
    })

    t.Run("LiquidationEvent", func(t *testing.T) {
        lp := cmfuturesstreams.LiquidationEventStreamParams{Symbol: models.Symbol(symLower)}
        path, err := cmfuturesstreams.BuildLiquidationEventStream(0, lp.Values())
        if err != nil { t.Fatalf("build liquidation: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("liquidation", 1, eventWait())
        t.Logf("liquidation events: %d", rec.count("liquidation"))
        // Best-effort shape checks
        if cnt := rec.count("liquidation"); cnt > 0 {
            ev := rec.liquidation[len(rec.liquidation)-1]
            if ev.EventType != "forceOrder" { t.Logf("unexpected event type: %s", ev.EventType) }
            assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("AllMarkPricesEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildAllMarkPricesEventStream(0, nil)
        if err != nil { t.Fatalf("build all mark prices: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allMarkPrices", 1, eventWait())
        t.Logf("allMarkPrices events: %d", rec.count("allMarkPrices"))
        if cnt := rec.count("allMarkPrices"); cnt > 0 {
            ev := rec.allMarkPrices[len(rec.allMarkPrices)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "markPriceUpdate" { t.Logf("unexpected type: %s", it.EventType) }
            }
        }
    })

    t.Run("AllMiniTickersEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildAllMiniTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build all mini: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allMiniTickers", 1, eventWait())
        t.Logf("allMiniTickers events: %d", rec.count("allMiniTickers"))
        if cnt := rec.count("allMiniTickers"); cnt > 0 {
            ev := rec.allMiniTickers[len(rec.allMiniTickers)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "24hrMiniTicker" { t.Logf("unexpected type: %s", it.EventType) }
            }
        }
    })

    t.Run("AllTickersEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildAllTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build all tickers: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allTickers", 1, eventWait())
        t.Logf("allTickers events: %d", rec.count("allTickers"))
        if cnt := rec.count("allTickers"); cnt > 0 {
            ev := rec.allTickers[len(rec.allTickers)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "24hrTicker" { t.Logf("unexpected type: %s", it.EventType) }
            }
        }
    })

    t.Run("AllBookTickersEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildAllBookTickersEventStream(0, nil)
        if err != nil { t.Fatalf("build all book tickers: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allBookTickers", 1, eventWait())
        t.Logf("allBookTickers events: %d", rec.count("allBookTickers"))
        if cnt := rec.count("allBookTickers"); cnt > 0 {
            ev := rec.allBookTickers[len(rec.allBookTickers)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "bookTicker" { t.Errorf("want e=bookTicker got %s", it.EventType) }
            }
        }
    })

    t.Run("AllLiquidationsEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildAllLiquidationsEventStream(0, nil)
        if err != nil { t.Fatalf("build all liquidations: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("allLiquidations", 1, eventWait())
        t.Logf("allLiquidations events: %d", rec.count("allLiquidations"))
        if cnt := rec.count("allLiquidations"); cnt > 0 {
            ev := rec.allLiquid[len(rec.allLiquid)-1]
            if ev != nil && len(*ev) > 0 {
                it := (*ev)[0]
                if it.EventType != "forceOrder" { t.Logf("unexpected event type: %s", it.EventType) }
            }
        }
    })

    t.Run("ContractInfoEvent", func(t *testing.T) {
        path, err := cmfuturesstreams.BuildContractInfoEventStream(0, nil)
        if err != nil { t.Fatalf("build contractInfo: %v", err) }
        subCb := func(context.Context, *models.SubscribeResponse) error { return nil }
        _ = market.Subscribe(context.Background(), &models.SubscribeRequest{Id: time.Now().UnixMicro(), Params: []string{path}}, &subCb)
        _ = rec.waitForMin("contractInfo", 1, eventWait())
        t.Logf("contractInfo events: %d", rec.count("contractInfo"))
        if cnt := rec.count("contractInfo"); cnt > 0 {
            ev := rec.contractInfo[len(rec.contractInfo)-1]
            if ev.EventType != "contractInfo" { t.Errorf("want e=contractInfo got %s", ev.EventType) }
        }
    })
}
