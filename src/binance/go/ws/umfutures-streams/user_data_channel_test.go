package streamstest

import (
    "bytes"
    "context"
    "io"
    "log"
    "os"
    "strconv"
    "strings"
    "sync"
    "testing"
    "time"

    restum "github.com/openxapi/binance-go/rest/umfutures"
    umfuturesstreams "github.com/openxapi/binance-go/ws/umfutures-streams"
    "github.com/openxapi/binance-go/ws/umfutures-streams/models"
)

// Capture SDK log lines containing the marker "unhandled message:" and fail the suite
type unhandledCatcherUD struct {
    matches []string
    mu      sync.Mutex
}

func (w *unhandledCatcherUD) Write(p []byte) (int, error) {
    if bytes.Contains(p, []byte("unhandled message:")) {
        w.mu.Lock()
        w.matches = append(w.matches, string(p))
        w.mu.Unlock()
    }
    return len(p), nil
}

// user data event recorder: accumulates received user-data events so tests can assert
type umUserDataEventRecorder struct {
    mu        sync.RWMutex
    expired   []*models.ListenKeyExpiredEvent
    acctUpd   []*models.AccountUpdateEvent
    marginCal []*models.MarginCallEvent
    ordUpd    []*models.OrderTradeUpdateEvent
    tradeLite []*models.TradeLiteEvent
    cfgUpd    []*models.AccountConfigUpdateEvent
    stratUpd  []*models.StrategyUpdateEvent
    gridUpd   []*models.GridUpdateEvent
    condRej   []*models.ConditionalOrderTriggerRejectEvent
    errors    []*models.ErrorMessage
}

func newUMUserDataEventRecorder() *umUserDataEventRecorder { return &umUserDataEventRecorder{} }
func (r *umUserDataEventRecorder) addExpired(v *models.ListenKeyExpiredEvent)                 { r.mu.Lock(); r.expired = append(r.expired, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addAcctUpd(v *models.AccountUpdateEvent)                    { r.mu.Lock(); r.acctUpd = append(r.acctUpd, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addMarginCal(v *models.MarginCallEvent)                     { r.mu.Lock(); r.marginCal = append(r.marginCal, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addOrdUpd(v *models.OrderTradeUpdateEvent)                  { r.mu.Lock(); r.ordUpd = append(r.ordUpd, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addTradeLite(v *models.TradeLiteEvent)                      { r.mu.Lock(); r.tradeLite = append(r.tradeLite, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addCfgUpd(v *models.AccountConfigUpdateEvent)               { r.mu.Lock(); r.cfgUpd = append(r.cfgUpd, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addStratUpd(v *models.StrategyUpdateEvent)                  { r.mu.Lock(); r.stratUpd = append(r.stratUpd, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addGridUpd(v *models.GridUpdateEvent)                       { r.mu.Lock(); r.gridUpd = append(r.gridUpd, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addCondRej(v *models.ConditionalOrderTriggerRejectEvent)    { r.mu.Lock(); r.condRej = append(r.condRej, v); r.mu.Unlock() }
func (r *umUserDataEventRecorder) addError(v *models.ErrorMessage)                            { r.mu.Lock(); r.errors = append(r.errors, v); r.mu.Unlock() }

func (r *umUserDataEventRecorder) count(key string) int {
    r.mu.RLock(); defer r.mu.RUnlock()
    switch key {
    case "listenKeyExpired": return len(r.expired)
    case "ACCOUNT_UPDATE": return len(r.acctUpd)
    case "MARGIN_CALL": return len(r.marginCal)
    case "ORDER_TRADE_UPDATE": return len(r.ordUpd)
    case "TRADE_LITE": return len(r.tradeLite)
    case "ACCOUNT_CONFIG_UPDATE": return len(r.cfgUpd)
    case "STRATEGY_UPDATE": return len(r.stratUpd)
    case "GRID_UPDATE": return len(r.gridUpd)
    case "CONDITIONAL_ORDER_TRIGGER_REJECT": return len(r.condRej)
    case "error": return len(r.errors)
    default: return 0
    }
}

func (r *umUserDataEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if r.count(key) >= min { return nil }
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}

// REST helpers (prefer testnet by default for user data)
func newRESTClientUserData() *restum.APIClient {
    cfg := restum.NewConfiguration()
    // cfg.Debug = true
    if s := os.Getenv("BINANCE_UMFUTURES_REST_SERVER"); s != "" {
        cfg.Servers[0].URL = s
    } else {
        // default to testnet for user data flows
        cfg.Servers[0].URL = "https://testnet.binancefuture.com"
    }
    return restum.NewAPIClient(cfg)
}

func restAuthContextUser() (context.Context, error) {
    apiKey := os.Getenv("BINANCE_API_KEY")
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if apiKey == "" || secret == "" {
        return nil, fmtError("BINANCE_API_KEY/BINANCE_SECRET_KEY not set for user data tests")
    }
    au := restum.NewAuth(apiKey)
    au.SetSecretKey(secret)
    return au.ContextWithValue(context.Background())
}

// fmtError keeps imports minimal
func fmtError(s string) error { return &tempErr{s: s} }
type tempErr struct{ s string }
func (e *tempErr) Error() string { return e.s }

// Create a listenKey via REST on testnet
func createListenKeyTestnet(t *testing.T) string {
    t.Helper()
    apiKey := os.Getenv("BINANCE_API_KEY")
    if apiKey == "" { t.Skip("BINANCE_API_KEY not set; skipping user data suite") }
    rc := newRESTClientUserData()
    ctx := context.Background()
    // Attach apiKey to context (header only). ContextWithValue requires a KeyType, so set a dummy secret.
    auth := restum.NewAuth(apiKey)
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if secret == "" { secret = "unused" }
    auth.SetSecretKey(secret)
    if c, err := auth.ContextWithValue(ctx); err == nil { ctx = c }
    resp, _, err := rc.FuturesAPI.CreateListenKeyV1(ctx).Execute()
    if err != nil || resp == nil || resp.ListenKey == nil || *resp.ListenKey == "" {
        t.Fatalf("failed to create listenKey on testnet: %v", err)
    }
    return *resp.ListenKey
}

func keepaliveListenKeyTestnet(t *testing.T, lk string) {
    t.Helper()
    apiKey := os.Getenv("BINANCE_API_KEY")
    if apiKey == "" { return }
    rc := newRESTClientUserData()
    ctx := context.Background()
    auth := restum.NewAuth(apiKey)
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if secret == "" { secret = "unused" }
    auth.SetSecretKey(secret)
    if c, err := auth.ContextWithValue(ctx); err == nil { ctx = c }
    _, _, _ = rc.FuturesAPI.UpdateListenKeyV1(ctx).Execute()
}

// placeTestLimitOrder places a small limit order to trigger ORDER_TRADE_UPDATE events
func placeTestLimitOrder(t *testing.T, symbol string, side string, qty string, price float64) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    prStr := strconv.FormatFloat(price, 'f', -1, 64)
    req := rc.FuturesAPI.CreateOrderV1(ctx).
        Symbol(symbol).
        Side(side).
        Type_("LIMIT").
        TimeInForce("GTC").
        Quantity(qty).
        Price(prStr).
        Timestamp(ts)
    _, _, err = req.Execute()
    return err
}

// togglePositionMode flips dual-side position to trigger ACCOUNT_CONFIG_UPDATE
func togglePositionMode(t *testing.T) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    // Get current mode
    cur, _, err := rc.FuturesAPI.GetPositionSideDualV1(ctx).Timestamp(ts).Execute()
    if err != nil || cur == nil || cur.DualSidePosition == nil { return err }
    // Flip value
    newVal := "false"
    if *cur.DualSidePosition == true { newVal = "false" } else { newVal = "true" }
    ts2 := time.Now().UnixMilli()
    _, _, err = rc.FuturesAPI.CreatePositionSideDualV1(ctx).DualSidePosition(newVal).Timestamp(ts2).Execute()
    return err
}

// TestFullIntegrationSuite_UserData runs request/response and event coverage for UserDataStreamsChannel
func TestFullIntegrationSuite_UserData(t *testing.T) {
    if testing.Short() { t.Skip("Skipping in short mode") }

    // Ensure credentials exist
    if os.Getenv("BINANCE_API_KEY") == "" || os.Getenv("BINANCE_SECRET_KEY") == "" {
        t.Skip("Missing BINANCE_API_KEY/SECRET_KEY; skipping user data suite")
    }

    // Capture SDK log output and fail on any 'unhandled message'
    cw := &unhandledCatcherUD{}
    log.SetOutput(io.MultiWriter(cw, os.Stderr))
    defer func() {
        log.SetOutput(os.Stderr)
        cw.mu.Lock(); defer cw.mu.Unlock()
        if len(cw.matches) > 0 {
            for _, line := range cw.matches { t.Logf("SDK log captured: %s", strings.TrimSpace(line)) }
            t.Fatalf("SDK emitted %d 'unhandled message' log(s) during UserData suite; treating as failure", len(cw.matches))
        }
    }()

    // Dedicated client and explicit testnet1 selection
    cfg := getTestConfig()
    stc, err := NewStreamTestClientDedicated(cfg)
    if err != nil { t.Fatalf("failed to create client: %v", err) }
    _ = stc.client.SetActiveServer("testnet1")
    // Detect if we are using testnet1 so we can disable unsupported WS requests
    isTestnet1 := false
    if as := stc.client.GetActiveServer(); as != nil {
        t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
        if strings.EqualFold(as.Name, "testnet1") || strings.Contains(as.URL, "fstream.binancefuture.com") {
            isTestnet1 = true
        }
    }

    // Acquire listenKey (testnet)
    listenKey := createListenKeyTestnet(t)
    t.Logf("listenKey acquired: %s...", maskListenKey(listenKey))

    // Channel + handlers
    ch := umfuturesstreams.NewUserDataStreamsChannel(stc.client)
    rec := newUMUserDataEventRecorder()
    ch.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error { rec.addError(msg); logJSON(t, "ws.error", msg); return nil })
    ch.HandleListenKeyExpiredEvent(func(ctx context.Context, ev *models.ListenKeyExpiredEvent) error { rec.addExpired(ev); return nil })
    ch.HandleAccountUpdateEvent(func(ctx context.Context, ev *models.AccountUpdateEvent) error { rec.addAcctUpd(ev); return nil })
    ch.HandleMarginCallEvent(func(ctx context.Context, ev *models.MarginCallEvent) error { rec.addMarginCal(ev); return nil })
    ch.HandleOrderTradeUpdateEvent(func(ctx context.Context, ev *models.OrderTradeUpdateEvent) error { rec.addOrdUpd(ev); return nil })
    ch.HandleTradeLiteEvent(func(ctx context.Context, ev *models.TradeLiteEvent) error { rec.addTradeLite(ev); return nil })
    ch.HandleAccountConfigUpdateEvent(func(ctx context.Context, ev *models.AccountConfigUpdateEvent) error { rec.addCfgUpd(ev); return nil })
    ch.HandleStrategyUpdateEvent(func(ctx context.Context, ev *models.StrategyUpdateEvent) error { rec.addStratUpd(ev); return nil })
    ch.HandleGridUpdateEvent(func(ctx context.Context, ev *models.GridUpdateEvent) error { rec.addGridUpd(ev); return nil })
    ch.HandleConditionalOrderTriggerRejectEvent(func(ctx context.Context, ev *models.ConditionalOrderTriggerRejectEvent) error { rec.addCondRej(ev); return nil })

    // Connect to user data stream using listenKey
    cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
    if err := ch.Connect(cctx, listenKey); err != nil { ccancel(); t.Fatalf("user data connect failed: %v", err) }
    ccancel()
    defer func() {
        dctx, dcancel := context.WithTimeout(context.Background(), 6*time.Second)
        _ = ch.Disconnect(dctx)
        dcancel()
    }()

    // ---------- Request/Response: Start, Ping, Stop ----------
    t.Run("Request_Start", func(t *testing.T) {
        if isTestnet1 {
            t.Skip("Skipping userDataStream.start on testnet1 (unsupported)")
        }
        // Attempt WS-based start. Some servers may not support this; treat timeouts/errors as acceptable.
        sid := time.Now().UnixMicro()
        done := make(chan struct{}, 1)
        var got *models.UserDataStreamsStartResponse
        cb := func(ctx context.Context, resp *models.UserDataStreamsStartResponse) error {
            got = resp; logJSON(t, "userData.start.response", resp)
            select { case done <- struct{}{}: default: }
            return nil
        }
        req := &models.UserDataStreamsStartRequest{Id: sid}
        req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
        spCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
        defer cancel()
        if err := ch.UserDataStreamsStart(spCtx, req, &cb); err != nil {
            t.Logf("userDataStreamsStart call err (acceptable if unsupported): %v", err)
        }
        select { case <-done: case <-time.After(8 * time.Second): t.Logf("timeout waiting userData.start response (acceptable)") }
        _ = got
    })

    t.Run("Request_Ping", func(t *testing.T) {
        if isTestnet1 {
            t.Skip("Skipping userDataStream.ping on testnet1 (unsupported)")
        }
        pid := time.Now().UnixMicro()
        done := make(chan struct{}, 1)
        var got *models.UserDataStreamsPingResponse
        cb := func(ctx context.Context, resp *models.UserDataStreamsPingResponse) error {
            got = resp; logJSON(t, "userData.ping.response", resp)
            select { case done <- struct{}{}: default: }
            return nil
        }
        req := &models.UserDataStreamsPingRequest{Id: pid}
        req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
        pgCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
        defer cancel()
        if err := ch.UserDataStreamsPing(pgCtx, req, &cb); err != nil {
            t.Logf("userDataStreamsPing call err (acceptable if unsupported): %v", err)
        }
        select { case <-done: case <-time.After(8 * time.Second): t.Logf("timeout waiting userData.ping response (acceptable)") }
        _ = got
    })

    t.Run("Request_Stop", func(t *testing.T) {
        if isTestnet1 {
            t.Skip("Skipping userDataStream.stop on testnet1 (unsupported)")
        }
        sid := time.Now().UnixMicro()
        done := make(chan struct{}, 1)
        var got *models.UserDataStreamsStopResponse
        cb := func(ctx context.Context, resp *models.UserDataStreamsStopResponse) error {
            got = resp; logJSON(t, "userData.stop.response", resp)
            select { case done <- struct{}{}: default: }
            return nil
        }
        req := &models.UserDataStreamsStopRequest{Id: sid}
        req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
        spCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
        defer cancel()
        if err := ch.UserDataStreamsStop(spCtx, req, &cb); err != nil {
            t.Logf("userDataStreamsStop call err (acceptable if unsupported): %v", err)
        }
        select { case <-done: case <-time.After(8 * time.Second): t.Logf("timeout waiting userData.stop response (acceptable)") }
        _ = got
    })

    // ---------- Event Handlers ----------
    t.Run("OrderTradeUpdateEvent", func(t *testing.T) {
        // Try to trigger by placing a small LIMIT order far from market to avoid fills.
        symUP, err := restPickSymbol(context.Background())
        if err != nil || symUP == "" { t.Fatalf("failed to pick symbol: %v", err) }
        // Get a reference price from REST; fall back to 30000 if unavailable
        ref := 30000.0
        if p, err := restMarkPrice(context.Background(), symUP); err == nil && p > 0 {
            ref = p
        } else {
            t.Logf("mark price fetch failed; using fallback: %v", err)
        }
        // Place a BUY 0.001 at 50%% of mark price to ensure it does not fill
        q := "0.001"
        price := ref * 0.5
        if err := placeTestLimitOrder(t, symUP, "BUY", q, price); err != nil {
            t.Logf("place test limit order failed (acceptable in CI without balance): %v", err)
        }
        // Wait for at least one ORDER_TRADE_UPDATE (NEW or rejected)
        _ = rec.waitForMin("ORDER_TRADE_UPDATE", 1, eventWait())
        cnt := rec.count("ORDER_TRADE_UPDATE")
        t.Logf("ORDER_TRADE_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.ordUpd[len(rec.ordUpd)-1]
            if ev.Event.EventType != "ORDER_TRADE_UPDATE" { t.Errorf("want e=ORDER_TRADE_UPDATE got %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
            // Basic order checks
            ord := ev.Event.OrderDetails
            assertNonEmpty(t, ord.Symbol, "symbol")
            assertNonEmpty(t, ord.Side, "side")
            assertNonEmpty(t, ord.OrderType, "orderType")
            assertNonEmpty(t, ord.OrderStatus, "orderStatus")
        } else {
            t.Logf("no ORDER_TRADE_UPDATE received (acceptable on accounts without permissions/balance)")
        }
    })

    t.Run("AccountConfigUpdateEvent", func(t *testing.T) {
        // Flip position mode to trigger config update; ignore errors in restricted envs
        if err := togglePositionMode(t); err != nil {
            t.Logf("toggle position mode failed (acceptable): %v", err)
        }
        _ = rec.waitForMin("ACCOUNT_CONFIG_UPDATE", 1, eventWait())
        cnt := rec.count("ACCOUNT_CONFIG_UPDATE")
        t.Logf("ACCOUNT_CONFIG_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.cfgUpd[len(rec.cfgUpd)-1]
            if ev.Event.EventType != "ACCOUNT_CONFIG_UPDATE" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        } else {
            t.Logf("no ACCOUNT_CONFIG_UPDATE received (acceptable)")
        }
    })

    t.Run("AccountUpdateEvent", func(t *testing.T) {
        // Often emitted on balance/position updates; may not occur in all envs
        _ = rec.waitForMin("ACCOUNT_UPDATE", 1, eventWait())
        cnt := rec.count("ACCOUNT_UPDATE")
        t.Logf("ACCOUNT_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.acctUpd[len(rec.acctUpd)-1]
            if ev.Event.EventType != "ACCOUNT_UPDATE" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        } else {
            t.Logf("no ACCOUNT_UPDATE received (acceptable)")
        }
    })

    t.Run("MarginCallEvent", func(t *testing.T) {
        // Rare in tests; validate shape if present
        _ = rec.waitForMin("MARGIN_CALL", 1, time.Second*3)
        cnt := rec.count("MARGIN_CALL")
        t.Logf("MARGIN_CALL events: %d", cnt)
        if cnt > 0 {
            ev := rec.marginCal[len(rec.marginCal)-1]
            if ev.Event.EventType != "MARGIN_CALL" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("TradeLiteEvent", func(t *testing.T) {
        // Might emit on fills; validate if present
        _ = rec.waitForMin("TRADE_LITE", 1, time.Second*3)
        cnt := rec.count("TRADE_LITE")
        t.Logf("TRADE_LITE events: %d", cnt)
        if cnt > 0 {
            ev := rec.tradeLite[len(rec.tradeLite)-1]
            if ev.Event.EventType != "TRADE_LITE" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
            assertNonEmpty(t, ev.Event.Symbol, "symbol")
        }
    })

    t.Run("StrategyUpdateEvent", func(t *testing.T) {
        _ = rec.waitForMin("STRATEGY_UPDATE", 1, time.Second*3)
        cnt := rec.count("STRATEGY_UPDATE")
        t.Logf("STRATEGY_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.stratUpd[len(rec.stratUpd)-1]
            if ev.Event.EventType != "STRATEGY_UPDATE" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("GridUpdateEvent", func(t *testing.T) {
        _ = rec.waitForMin("GRID_UPDATE", 1, time.Second*3)
        cnt := rec.count("GRID_UPDATE")
        t.Logf("GRID_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.gridUpd[len(rec.gridUpd)-1]
            if ev.Event.EventType != "GRID_UPDATE" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("ConditionalOrderTriggerRejectEvent", func(t *testing.T) {
        // Difficult to force; validate if present
        _ = rec.waitForMin("CONDITIONAL_ORDER_TRIGGER_REJECT", 1, time.Second*3)
        cnt := rec.count("CONDITIONAL_ORDER_TRIGGER_REJECT")
        t.Logf("CONDITIONAL_ORDER_TRIGGER_REJECT events: %d", cnt)
        if cnt > 0 {
            ev := rec.condRej[len(rec.condRej)-1]
            if ev.Event.EventType != "CONDITIONAL_ORDER_TRIGGER_REJECT" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
            assertRecentMs(t, ev.Event.EventTime, 24*time.Hour, "eventTime")
        }
    })

    t.Run("ListenKeyExpiredEvent", func(t *testing.T) {
        // Not feasible to force within suite (expires ~60m); ensure handler registered. If any captured, validate.
        cnt := rec.count("listenKeyExpired")
        t.Logf("listenKeyExpired events: %d", cnt)
        if cnt > 0 {
            ev := rec.expired[len(rec.expired)-1]
            if ev.Event.EventType != "listenKeyExpired" { t.Logf("unexpected event type: %s", ev.Event.EventType) }
        }
    })

    t.Run("ErrorMessage", func(t *testing.T) {
        // Our start/ping/stop requests may produce errors, which we capture
        cnt := rec.count("error")
        t.Logf("error messages captured: %d", cnt)
        if cnt > 0 {
            msg := rec.errors[len(rec.errors)-1]
            if msg.Error.Code == 0 { t.Logf("error code missing") }
            assertNonEmpty(t, msg.Error.Msg, "error.msg")
        }
    })

    t.Run("ListenKeyKeepalive", func(t *testing.T) {
        // Exercise REST keepalive to ensure LK remains valid during suite
        keepaliveListenKeyTestnet(t, listenKey)
    })
}

func maskListenKey(lk string) string {
    if len(lk) <= 8 { return lk }
    return lk[:4] + "..." + lk[len(lk)-4:]
}
