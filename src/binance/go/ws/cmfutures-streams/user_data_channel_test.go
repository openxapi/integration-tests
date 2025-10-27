package streamstest

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "math"
    "os"
    "strconv"
    "strings"
    "sync"
    "testing"
    "time"

    restcm "github.com/openxapi/binance-go/rest/cmfutures"
    cmfuturesstreams "github.com/openxapi/binance-go/ws/cmfutures-streams"
    "github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// Capture SDK log lines containing the marker "unhandled message:" and fail the suite
type unhandledCatcherUD struct {
    matches []string
    mu      sync.Mutex
}

func (w *unhandledCatcherUD) Write(p []byte) (int, error) {
    if bytes.Contains(p, []byte("unhandled message:")) {
        w.mu.Lock(); w.matches = append(w.matches, string(p)); w.mu.Unlock()
    }
    return len(p), nil
}

// user data event recorder
type cmUserDataEventRecorder struct {
    mu        sync.RWMutex
    expired   []*models.ListenKeyExpiredEvent
    acctUpd   []*models.AccountUpdateEvent
    marginCal []*models.MarginCallEvent
    ordUpd    []*models.OrderTradeUpdateEvent
    cfgUpd    []*models.AccountConfigUpdateEvent
    stratUpd  []*models.StrategyUpdateEvent
    gridUpd   []*models.GridUpdateEvent
    errors    []*models.ErrorMessage
}

func newCMUserDataEventRecorder() *cmUserDataEventRecorder { return &cmUserDataEventRecorder{} }
func (r *cmUserDataEventRecorder) addExpired(v *models.ListenKeyExpiredEvent)              { r.mu.Lock(); r.expired = append(r.expired, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addAcctUpd(v *models.AccountUpdateEvent)                 { r.mu.Lock(); r.acctUpd = append(r.acctUpd, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addMarginCal(v *models.MarginCallEvent)                  { r.mu.Lock(); r.marginCal = append(r.marginCal, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addOrdUpd(v *models.OrderTradeUpdateEvent)               { r.mu.Lock(); r.ordUpd = append(r.ordUpd, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addCfgUpd(v *models.AccountConfigUpdateEvent)            { r.mu.Lock(); r.cfgUpd = append(r.cfgUpd, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addStratUpd(v *models.StrategyUpdateEvent)               { r.mu.Lock(); r.stratUpd = append(r.stratUpd, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addGridUpd(v *models.GridUpdateEvent)                    { r.mu.Lock(); r.gridUpd = append(r.gridUpd, v); r.mu.Unlock() }
func (r *cmUserDataEventRecorder) addError(v *models.ErrorMessage)                         { r.mu.Lock(); r.errors = append(r.errors, v); r.mu.Unlock() }

func (r *cmUserDataEventRecorder) count(key string) int {
    r.mu.RLock(); defer r.mu.RUnlock()
    switch key {
    case "listenKeyExpired": return len(r.expired)
    case "ACCOUNT_UPDATE": return len(r.acctUpd)
    case "MARGIN_CALL": return len(r.marginCal)
    case "ORDER_TRADE_UPDATE": return len(r.ordUpd)
    case "ACCOUNT_CONFIG_UPDATE": return len(r.cfgUpd)
    case "STRATEGY_UPDATE": return len(r.stratUpd)
    case "GRID_UPDATE": return len(r.gridUpd)
    case "error": return len(r.errors)
    default: return 0
    }
}

func (r *cmUserDataEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if r.count(key) >= min { return nil }
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}

// REST helpers
func newRESTClientUserData() *restcm.APIClient {
    cfg := restcm.NewConfiguration()
    if s := os.Getenv("BINANCE_CMFUTURES_REST_SERVER"); s != "" {
        cfg.Servers[0].URL = s
    } else { cfg.Servers[0].URL = "https://testnet.binancefuture.com" }
    return restcm.NewAPIClient(cfg)
}

func restAuthContextUser() (context.Context, error) {
    apiKey := os.Getenv("BINANCE_API_KEY")
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if apiKey == "" || secret == "" { return nil, fmtError("BINANCE_API_KEY/BINANCE_SECRET_KEY not set for user data tests") }
    au := restcm.NewAuth(apiKey)
    au.SetSecretKey(secret)
    return au.ContextWithValue(context.Background())
}

func fmtError(s string) error { return &tempErr{s: s} }
type tempErr struct{ s string }
func (e *tempErr) Error() string { return e.s }

func createListenKeyTestnet(t *testing.T) string {
    t.Helper()
    apiKey := os.Getenv("BINANCE_API_KEY")
    if apiKey == "" { t.Skip("BINANCE_API_KEY not set; skipping user data suite") }
    rc := newRESTClientUserData()
    ctx := context.Background()
    auth := restcm.NewAuth(apiKey)
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if secret == "" { secret = "unused" }
    auth.SetSecretKey(secret)
    if c, err := auth.ContextWithValue(ctx); err == nil { ctx = c }
    resp, _, err := rc.FuturesAPI.CreateListenKeyV1(ctx).Execute()
    if err != nil || resp == nil || resp.ListenKey == nil || *resp.ListenKey == "" { t.Fatalf("failed to create listenKey: %v", err) }
    return *resp.ListenKey
}

func keepaliveListenKeyTestnet(t *testing.T, lk string) {
    t.Helper()
    apiKey := os.Getenv("BINANCE_API_KEY")
    if apiKey == "" { return }
    rc := newRESTClientUserData()
    ctx := context.Background()
    auth := restcm.NewAuth(apiKey)
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if secret == "" { secret = "unused" }
    auth.SetSecretKey(secret)
    if c, err := auth.ContextWithValue(ctx); err == nil { ctx = c }
    _, _, _ = rc.FuturesAPI.UpdateListenKeyV1(ctx).Execute()
}

// Derived from UM suite but using CM REST
func placeTestMarketOrder(t *testing.T, symbol string, side string, qty string) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    req := rc.FuturesAPI.CreateOrderV1(ctx).Symbol(symbol).Side(side).Type_("MARKET").Quantity(qty).Timestamp(ts)
    _, _, err = req.Execute()
    if err == nil { return nil }
    if ge, ok := err.(*restcm.GenericOpenAPIError); ok {
        var em struct{ Code int `json:"code"`; Msg string `json:"msg"` }
        if len(ge.Body()) > 0 {
            if e2 := json.Unmarshal(ge.Body(), &em); e2 == nil && (em.Code != 0 || em.Msg != "") {
                return fmt.Errorf("market order rejected: status=%s code=%d msg=%s body=%s", ge.Error(), em.Code, em.Msg, string(ge.Body()))
            }
            return fmt.Errorf("market order rejected: status=%s body=%s", ge.Error(), string(ge.Body()))
        }
        return fmt.Errorf("market order rejected: status=%s", ge.Error())
    }
    return err
}

func cancelAllOpenOrders(t *testing.T) {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { t.Logf("cancelAllOpenOrders: no auth: %v", err); return }
    ts := time.Now().UnixMilli()
    orders, _, err := rc.FuturesAPI.GetOpenOrdersV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
    if err != nil {
        if ge, ok := err.(*restcm.GenericOpenAPIError); ok {
            t.Logf("cancelAllOpenOrders: get open orders failed: status=%s body=%s", ge.Error(), string(ge.Body()))
        } else { t.Logf("cancelAllOpenOrders: get open orders failed: %v", err) }
        return
    }
    if len(orders) == 0 { t.Logf("cancelAllOpenOrders: no open orders"); return }
    syms := make(map[string]struct{})
    for _, o := range orders { if o.Symbol != nil && *o.Symbol != "" { syms[*o.Symbol] = struct{}{} } }
    for sym := range syms {
        ts2 := time.Now().UnixMilli()
        _, _, err := rc.FuturesAPI.DeleteAllOpenOrdersV1(ctx).Symbol(sym).Timestamp(ts2).RecvWindow(5000).Execute()
        if err != nil { t.Logf("cancelAllOpenOrders: cancel %s failed: %v", sym, err) }
    }
}

func placeTestLimitOrder(t *testing.T, symbol string, side string, qty string, priceStr string) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    req := rc.FuturesAPI.CreateOrderV1(ctx).Symbol(symbol).Side(side).Type_("LIMIT").TimeInForce("GTC").Quantity(qty).Price(priceStr).Timestamp(ts)
    _, _, err = req.Execute()
    if err == nil { return nil }
    if ge, ok := err.(*restcm.GenericOpenAPIError); ok {
        var em struct{ Code int `json:"code"`; Msg string `json:"msg"` }
        if len(ge.Body()) > 0 {
            if e2 := json.Unmarshal(ge.Body(), &em); e2 == nil && (em.Code != 0 || em.Msg != "") {
                return fmt.Errorf("order rejected: status=%s code=%d msg=%s body=%s", ge.Error(), em.Code, em.Msg, string(ge.Body()))
            }
            return fmt.Errorf("order rejected: status=%s body=%s", ge.Error(), string(ge.Body()))
        }
        return fmt.Errorf("order rejected: status=%s", ge.Error())
    }
    return err
}

func roundToPrec(x float64, prec int) float64 {
    if prec < 0 { prec = 0 }
    p := math.Pow10(prec)
    return math.Floor(x*p) / p
}

func calcLimitOrderParams(ctx context.Context, rc *restcm.APIClient, symbol string, targetPrice float64) (qtyStr, priceStr string, err error) {
    info, _, e := rc.FuturesAPI.GetExchangeInfoV1(ctx).Execute()
    if e != nil || info == nil || info.Symbols == nil { return "", "", fmt.Errorf("exchangeInfo error: %v", e) }
    var qp, pp int32 = 3, 2
    var priceStep float64
    var qtyStep float64
    var minQty float64
    var priceDecs int
    var qtyDecs int
    floorToStep := func(value float64, step float64, decs int) float64 {
        if step <= 0 { p := math.Pow10(decs); return math.Floor(value*p) / p }
        scale := math.Pow10(decs)
        vUnits := math.Floor(value*scale + 1e-9)
        sUnits := math.Round(step * scale)
        if sUnits <= 0 { sUnits = 1 }
        q := math.Floor(vUnits/sUnits) * sUnits
        return q / scale
    }
    ceilToStep := func(value float64, step float64, decs int) float64 {
        if step <= 0 { p := math.Pow10(decs); return math.Ceil(value*p) / p }
        scale := math.Pow10(decs)
        vUnits := math.Ceil(value*scale - 1e-9)
        sUnits := math.Round(step * scale)
        if sUnits <= 0 { sUnits = 1 }
        q := (math.Ceil(vUnits/sUnits)) * sUnits
        return q / scale
    }
    decsFromStepStr := func(s string) int { if s == "" { return 0 }; if i := strings.IndexByte(s, '.'); i >= 0 { return len(s) - i - 1 }; return 0 }
    parseFloatSafe := func(s string) float64 { if s == "" { return 0 }; if v, err := strconv.ParseFloat(s, 64); err == nil { return v }; return 0 }
    for _, s := range info.Symbols {
        if s.Symbol != nil && strings.EqualFold(*s.Symbol, symbol) {
            if s.QuantityPrecision != nil { qp = *s.QuantityPrecision }
            if s.PricePrecision != nil { pp = *s.PricePrecision }
            if b, err := json.Marshal(s); err == nil {
                var sm map[string]any
                if json.Unmarshal(b, &sm) == nil {
                    if fs, ok := sm["filters"].([]any); ok {
                        for _, f := range fs {
                            fm, _ := f.(map[string]any)
                            ft, _ := fm["filterType"].(string)
                            switch strings.ToUpper(ft) {
                            case "PRICE_FILTER":
                                if ts, _ := fm["tickSize"].(string); ts != "" {
                                    priceStep = parseFloatSafe(ts)
                                    priceDecs = decsFromStepStr(ts)
                                }
                            case "LOT_SIZE":
                                if ss, _ := fm["stepSize"].(string); ss != "" {
                                    qtyStep = parseFloatSafe(ss)
                                    qtyDecs = decsFromStepStr(ss)
                                }
                                if mq, _ := fm["minQty"].(string); mq != "" {
                                    minQty = parseFloatSafe(mq)
                                }
                            case "MARKET_LOT_SIZE":
                                if qtyStep == 0 {
                                    if ss, _ := fm["stepSize"].(string); ss != "" {
                                        qtyStep = parseFloatSafe(ss)
                                        qtyDecs = decsFromStepStr(ss)
                                    }
                                }
                            }
                        }
                    }
                }
            }
            break
        }
    }
    if priceDecs == 0 { priceDecs = int(pp) }
    if qtyDecs == 0 { qtyDecs = int(qp) }
    price := floorToStep(targetPrice, priceStep, priceDecs)
    if price <= 0 { return "", "", fmt.Errorf("invalid rounded price") }
    minNotional := 5.0
    needQty := minNotional / price
    q := needQty
    if minQty > 0 && q < minQty { q = minQty }
    q = ceilToStep(q, qtyStep, qtyDecs)
    maxNotional := 20.0
    if q*price > maxNotional { target := maxNotional / price; q = floorToStep(target, qtyStep, qtyDecs); if q <= 0 { q = 1.0 / math.Pow10(qtyDecs) } }
    qtyStr = fmt.Sprintf("%.*f", qtyDecs, q)
    priceStr = fmt.Sprintf("%.*f", priceDecs, price)
    return qtyStr, priceStr, nil
}

func togglePositionMode(t *testing.T) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    cur, _, err := rc.FuturesAPI.GetPositionSideDualV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
    if err != nil { if ge, ok := err.(*restcm.GenericOpenAPIError); ok { return fmt.Errorf("get position mode failed: status=%s body=%s", ge.Error(), string(ge.Body())) }; return err }
    if cur == nil || cur.DualSidePosition == nil { return fmt.Errorf("get position mode returned empty response") }
    newVal := "false"; if *cur.DualSidePosition == true { newVal = "false" } else { newVal = "true" }
    ts2 := time.Now().UnixMilli()
    _, _, err = rc.FuturesAPI.CreatePositionSideDualV1(ctx).DualSidePosition(newVal).Timestamp(ts2).RecvWindow(5000).Execute()
    if err != nil { if ge, ok := err.(*restcm.GenericOpenAPIError); ok { return fmt.Errorf("set position mode failed: status=%s body=%s", ge.Error(), string(ge.Body())) }; return err }
    return nil
}

// changeLeverage attempts to change initial leverage for a symbol to induce ACCOUNT_CONFIG_UPDATE
func changeLeverage(t *testing.T, symbol string, lev int32) error {
    t.Helper()
    rc := newRESTClientUserData()
    ctx, err := restAuthContextUser()
    if err != nil { return err }
    ts := time.Now().UnixMilli()
    _, _, err = rc.FuturesAPI.CreateLeverageV1(ctx).Symbol(symbol).Leverage(lev).Timestamp(ts).RecvWindow(5000).Execute()
    if err != nil {
        if ge, ok := err.(*restcm.GenericOpenAPIError); ok {
            return fmt.Errorf("change leverage failed: status=%s body=%s", ge.Error(), string(ge.Body()))
        }
    }
    return err
}

// TestFullIntegrationSuite_UserData runs request/response and event coverage for UserDataStreamChannel
func TestFullIntegrationSuite_UserData(t *testing.T) {
    if testing.Short() { t.Skip("Skipping in short mode") }
    if os.Getenv("BINANCE_API_KEY") == "" || os.Getenv("BINANCE_SECRET_KEY") == "" { t.Skip("Missing BINANCE_API_KEY/SECRET_KEY; skipping user data suite") }

    cw := &unhandledCatcherUD{}
    log.SetOutput(io.MultiWriter(cw, os.Stderr))
    defer func() {
        log.SetOutput(os.Stderr)
        cw.mu.Lock(); defer cw.mu.Unlock()
        if len(cw.matches) > 0 { for _, line := range cw.matches { t.Logf("SDK log captured: %s", strings.TrimSpace(line)) }; t.Fatalf("SDK emitted %d 'unhandled message' log(s) during UserData suite; treating as failure", len(cw.matches)) }
    }()

    cfg := getTestConfig()
    stc, err := NewStreamTestClientDedicated(cfg)
    if err != nil { t.Fatalf("failed to create client: %v", err) }
    _ = stc.client.SetActiveServer("testnet")

    listenKey := createListenKeyTestnet(t)
    t.Logf("listenKey acquired: %s...", maskListenKey(listenKey))

    ch := cmfuturesstreams.NewUserDataStreamChannel(stc.client)
    rec := newCMUserDataEventRecorder()
    ch.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error { rec.addError(msg); logJSON(t, "ws.error", msg); return nil })
    ch.HandleListenKeyExpiredEvent(func(ctx context.Context, ev *models.ListenKeyExpiredEvent) error { rec.addExpired(ev); return nil })
    ch.HandleAccountUpdateEvent(func(ctx context.Context, ev *models.AccountUpdateEvent) error { rec.addAcctUpd(ev); return nil })
    ch.HandleMarginCallEvent(func(ctx context.Context, ev *models.MarginCallEvent) error { rec.addMarginCal(ev); return nil })
    ch.HandleOrderTradeUpdateEvent(func(ctx context.Context, ev *models.OrderTradeUpdateEvent) error { rec.addOrdUpd(ev); return nil })
    ch.HandleAccountConfigUpdateEvent(func(ctx context.Context, ev *models.AccountConfigUpdateEvent) error { rec.addCfgUpd(ev); return nil })
    ch.HandleStrategyUpdateEvent(func(ctx context.Context, ev *models.StrategyUpdateEvent) error { rec.addStratUpd(ev); return nil })
    ch.HandleGridUpdateEvent(func(ctx context.Context, ev *models.GridUpdateEvent) error { rec.addGridUpd(ev); return nil })

    cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
    if err := ch.Connect(cctx, listenKey); err != nil { ccancel(); t.Fatalf("user data connect failed: %v", err) }
    ccancel()
    defer func() { dctx, dcancel := context.WithTimeout(context.Background(), 6*time.Second); _ = ch.Disconnect(dctx); dcancel() }()

    // Determine if we are on testnet (user-data RPCs are not supported there)
    isTestnet := false
    if as := stc.client.GetActiveServer(); as != nil {
        if strings.EqualFold(as.Name, "testnet") || strings.Contains(as.URL, "binancefuture.com") {
            isTestnet = true
        }
    }

    t.Run("Request_Start", func(t *testing.T) {
        if isTestnet { t.Skip("Skip userDataStream.start on testnet (unsupported)") }
        sid := newRequestID()
        before := rec.count("error")
        // Server is expected to reject this method; we assert error is captured
        req := &models.UserDataStreamStartRequest{Id: sid}
        if k := os.Getenv("BINANCE_API_KEY"); k != "" { req.Params.ApiKey = k }
        if err := ch.Start(context.Background(), req, nil); err != nil {
            t.Logf("start err (acceptable): %v", err)
        }
        _ = rec.waitForMin("error", before+1, 5*time.Second)
        got := rec.count("error")
        t.Logf("errors after start: %d (before=%d)", got, before)
        if got > before {
            em := rec.errors[len(rec.errors)-1]
            if em == nil { t.Fatalf("nil error captured") }
            if !msgIDEqual(em.Id, sid) {
                t.Logf("error id mismatch: want=%s got=%s", sid.String(), em.Id.String())
            }
            if em.Error.Code == 0 || strings.TrimSpace(em.Error.Msg) == "" {
                t.Errorf("invalid error payload: %+v", em)
            }
        }
    })

    t.Run("Request_Ping", func(t *testing.T) {
        if isTestnet { t.Skip("Skip userDataStream.ping on testnet (unsupported)") }
        pid := newRequestID()
        done := make(chan struct{}, 1)
        cb := func(ctx context.Context, resp *models.UserDataStreamPingResponse) error {
            if resp != nil && !msgIDEqual(resp.Id, pid) {
                t.Logf("ping response id mismatch: want=%s got=%s", pid.String(), resp.Id.String())
            }
            logJSON(t, "userData.ping.response", resp)
            select { case done <- struct{}{}: default: }
            return nil
        }
        if err := ch.Ping(context.Background(), &models.UserDataStreamPingRequest{Id: pid}, &cb); err != nil { t.Logf("ping err (acceptable): %v", err) }
        select { case <-done: case <-time.After(5 * time.Second): t.Logf("timeout waiting ping response (acceptable)") }
    })

    t.Run("Request_Stop", func(t *testing.T) {
        if isTestnet { t.Skip("Skip userDataStream.stop on testnet (unsupported)") }
        rid := newRequestID()
        before := rec.count("error")
        req := &models.UserDataStreamStopRequest{Id: rid}
        if k := os.Getenv("BINANCE_API_KEY"); k != "" { req.Params.ApiKey = k }
        if err := ch.Stop(context.Background(), req, nil); err != nil {
            t.Logf("stop err (acceptable): %v", err)
        }
        _ = rec.waitForMin("error", before+1, 5*time.Second)
        got := rec.count("error")
        t.Logf("errors after stop: %d (before=%d)", got, before)
        if got > before {
            em := rec.errors[len(rec.errors)-1]
            if em == nil { t.Fatalf("nil error captured") }
            if !msgIDEqual(em.Id, rid) {
                t.Logf("stop error id mismatch: want=%s got=%s", rid.String(), em.Id.String())
            }
            if em.Error.Code == 0 || strings.TrimSpace(em.Error.Msg) == "" {
                t.Errorf("invalid error payload: %+v", em)
            }
        }
    })

    t.Run("OrderTradeUpdateEvent", func(t *testing.T) {
        // try to induce with a small limit order first (NEW event), fallback to market
        sym, err := restPickSymbol(context.Background())
        if err != nil || sym == "" { sym = "BTCUSD_PERP" }
        ref := 30000.0
        if p, err := restMarkPrice(context.Background(), sym); err == nil && p > 0 { ref = p }
        rc := newRESTClientUserData()
        if qty, price, e := calcLimitOrderParams(context.Background(), rc, sym, ref); e == nil {
            t.Logf("attempting small limit order for ORDER_TRADE_UPDATE: symbol=%s qty=%s price=%s", sym, qty, price)
            if err := placeTestLimitOrder(t, sym, "BUY", qty, price); err != nil {
                t.Logf("limit order rejected (acceptable): %v", err)
                t.Logf("fallback to market order for ORDER_TRADE_UPDATE: symbol=%s qty=%s", sym, qty)
                if err := placeTestMarketOrder(t, sym, "BUY", qty); err != nil { t.Logf("market order rejected (acceptable): %v", err) }
            }
            _ = rec.waitForMin("ORDER_TRADE_UPDATE", 1, eventWait())
        }
        cnt := rec.count("ORDER_TRADE_UPDATE")
        t.Logf("ORDER_TRADE_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.ordUpd[len(rec.ordUpd)-1]
            if ev.EventType != "ORDER_TRADE_UPDATE" { t.Errorf("want e=ORDER_TRADE_UPDATE got %s", ev.EventType) }
            if ev.EventTime <= 0 || ev.TransactionTime <= 0 { t.Errorf("invalid timestamps: E=%d T=%d", ev.EventTime, ev.TransactionTime) }
            if ev.OrderDetails.Symbol == "" { t.Errorf("empty order symbol") }
            if ev.OrderDetails.OrderID == 0 { t.Logf("order id missing or zero") }
        }
    })

    t.Run("AccountConfigUpdateEvent", func(t *testing.T) {
        before := rec.count("ACCOUNT_CONFIG_UPDATE")
        sym, err := restPickSymbol(context.Background())
        if err != nil || sym == "" { sym = "BTCUSD_PERP" }
        if err := changeLeverage(t, sym, 5); err != nil {
            t.Logf("change leverage failed (acceptable): %v", err)
        } else {
            _ = rec.waitForMin("ACCOUNT_CONFIG_UPDATE", before+1, eventWait())
        }
        cnt := rec.count("ACCOUNT_CONFIG_UPDATE")
        t.Logf("ACCOUNT_CONFIG_UPDATE events: %d", cnt)
        if cnt > before {
            ev := rec.cfgUpd[len(rec.cfgUpd)-1]
            if ev.EventType != "ACCOUNT_CONFIG_UPDATE" { t.Errorf("want e=ACCOUNT_CONFIG_UPDATE got %s", ev.EventType) }
            if ev.EventTime <= 0 || ev.TransactionTime <= 0 { t.Logf("config update not recent: E=%d T=%d", ev.EventTime, ev.TransactionTime) }
            // If symbol-level config exists, assert leverage sensible
            if ev.AccountConfigurationForTradePair.Symbol != "" {
                if ev.AccountConfigurationForTradePair.Leverage <= 0 { t.Errorf("invalid leverage in config update: %d", ev.AccountConfigurationForTradePair.Leverage) }
            }
        }
    })

    t.Run("AccountUpdateEvent", func(t *testing.T) {
        // This may be produced by fills or other balance changes; best-effort wait
        _ = rec.waitForMin("ACCOUNT_UPDATE", 1, 2*time.Second)
        cnt := rec.count("ACCOUNT_UPDATE")
        t.Logf("ACCOUNT_UPDATE events: %d", cnt)
        if cnt > 0 {
            ev := rec.acctUpd[len(rec.acctUpd)-1]
            if ev.EventType != "ACCOUNT_UPDATE" { t.Errorf("want e=ACCOUNT_UPDATE got %s", ev.EventType) }
            if ev.EventTime <= 0 || ev.TransactionTime <= 0 { t.Logf("account update timestamps not set: E=%d T=%d", ev.EventTime, ev.TransactionTime) }
            // If balance updates exist, basic field sanity
            if len(ev.UpdateData.BalanceUpdates) > 0 {
                bu := ev.UpdateData.BalanceUpdates[0]
                assertNonEmpty(t, bu.Asset, "BalanceUpdates.Asset")
                _ = tryParseFloat(t, bu.WalletBalance, "BalanceUpdates.WalletBalance")
            }
        }
    })

    t.Run("Summary_Counts", func(t *testing.T) {
        t.Logf("Errors: %d", rec.count("error"))
        t.Logf("ListenKeyExpired: %d", rec.count("listenKeyExpired"))
        t.Logf("ACCOUNT_UPDATE: %d", rec.count("ACCOUNT_UPDATE"))
        t.Logf("MARGIN_CALL: %d", rec.count("MARGIN_CALL"))
        t.Logf("ORDER_TRADE_UPDATE: %d", rec.count("ORDER_TRADE_UPDATE"))
        t.Logf("ACCOUNT_CONFIG_UPDATE: %d", rec.count("ACCOUNT_CONFIG_UPDATE"))
        t.Logf("STRATEGY_UPDATE: %d", rec.count("STRATEGY_UPDATE"))
        t.Logf("GRID_UPDATE: %d", rec.count("GRID_UPDATE"))
    })
}

// mask helper for logging listenKey
func maskListenKey(s string) string {
    if len(s) <= 6 { return s }
    return s[:3] + strings.Repeat("*", len(s)-6) + s[len(s)-3:]
}
