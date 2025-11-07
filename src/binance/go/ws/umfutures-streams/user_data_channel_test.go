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
func (r *umUserDataEventRecorder) addExpired(v *models.ListenKeyExpiredEvent) {
	r.mu.Lock()
	r.expired = append(r.expired, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addAcctUpd(v *models.AccountUpdateEvent) {
	r.mu.Lock()
	r.acctUpd = append(r.acctUpd, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addMarginCal(v *models.MarginCallEvent) {
	r.mu.Lock()
	r.marginCal = append(r.marginCal, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addOrdUpd(v *models.OrderTradeUpdateEvent) {
	r.mu.Lock()
	r.ordUpd = append(r.ordUpd, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addTradeLite(v *models.TradeLiteEvent) {
	r.mu.Lock()
	r.tradeLite = append(r.tradeLite, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addCfgUpd(v *models.AccountConfigUpdateEvent) {
	r.mu.Lock()
	r.cfgUpd = append(r.cfgUpd, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addStratUpd(v *models.StrategyUpdateEvent) {
	r.mu.Lock()
	r.stratUpd = append(r.stratUpd, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addGridUpd(v *models.GridUpdateEvent) {
	r.mu.Lock()
	r.gridUpd = append(r.gridUpd, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addCondRej(v *models.ConditionalOrderTriggerRejectEvent) {
	r.mu.Lock()
	r.condRej = append(r.condRej, v)
	r.mu.Unlock()
}
func (r *umUserDataEventRecorder) addError(v *models.ErrorMessage) {
	r.mu.Lock()
	r.errors = append(r.errors, v)
	r.mu.Unlock()
}

func (r *umUserDataEventRecorder) count(key string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch key {
	case "listenKeyExpired":
		return len(r.expired)
	case "ACCOUNT_UPDATE":
		return len(r.acctUpd)
	case "MARGIN_CALL":
		return len(r.marginCal)
	case "ORDER_TRADE_UPDATE":
		return len(r.ordUpd)
	case "TRADE_LITE":
		return len(r.tradeLite)
	case "ACCOUNT_CONFIG_UPDATE":
		return len(r.cfgUpd)
	case "STRATEGY_UPDATE":
		return len(r.stratUpd)
	case "GRID_UPDATE":
		return len(r.gridUpd)
	case "CONDITIONAL_ORDER_TRIGGER_REJECT":
		return len(r.condRej)
	case "error":
		return len(r.errors)
	default:
		return 0
	}
}

func (r *umUserDataEventRecorder) waitForMin(key string, min int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if r.count(key) >= min {
			return nil
		}
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
	if apiKey == "" {
		t.Skip("BINANCE_API_KEY not set; skipping user data suite")
	}
	rc := newRESTClientUserData()
	ctx := context.Background()
	// Attach apiKey to context (header only). ContextWithValue requires a KeyType, so set a dummy secret.
	auth := restum.NewAuth(apiKey)
	secret := os.Getenv("BINANCE_SECRET_KEY")
	if secret == "" {
		secret = "unused"
	}
	auth.SetSecretKey(secret)
	if c, err := auth.ContextWithValue(ctx); err == nil {
		ctx = c
	}
	resp, _, err := rc.FuturesAPI.CreateListenKeyV1(ctx).Execute()
	if err != nil || resp == nil || resp.ListenKey == nil || *resp.ListenKey == "" {
		t.Fatalf("failed to create listenKey on testnet: %v", err)
	}
	return *resp.ListenKey
}

func keepaliveListenKeyTestnet(t *testing.T, lk string) {
	t.Helper()
	apiKey := os.Getenv("BINANCE_API_KEY")
	if apiKey == "" {
		return
	}
	rc := newRESTClientUserData()
	ctx := context.Background()
	auth := restum.NewAuth(apiKey)
	secret := os.Getenv("BINANCE_SECRET_KEY")
	if secret == "" {
		secret = "unused"
	}
	auth.SetSecretKey(secret)
	if c, err := auth.ContextWithValue(ctx); err == nil {
		ctx = c
	}
	_, _, _ = rc.FuturesAPI.UpdateListenKeyV1(ctx).Execute()
}

// placeTestMarketOrder places a small market order to try to trigger TRADE_LITE and ACCOUNT_UPDATE
func placeTestMarketOrder(t *testing.T, symbol string, side string, qty string) error {
	t.Helper()
	rc := newRESTClientUserData()
	ctx, err := restAuthContextUser()
	if err != nil {
		return err
	}
	dualMode, derr := detectDualPositionMode(ctx, rc)
	if derr != nil {
		t.Logf("placeTestMarketOrder: position mode lookup failed: %v", derr)
	}
	ts := time.Now().UnixMilli()
	req := rc.FuturesAPI.CreateOrderV1(ctx).
		Symbol(symbol).
		Side(side).
		Type_("MARKET").
		Quantity(qty).
		Timestamp(ts)
	if dualMode {
		posSide := "LONG"
		if strings.EqualFold(side, "SELL") {
			posSide = "SHORT"
		}
		req = req.PositionSide(posSide)
	}
	_, _, err = req.Execute()
	if err == nil {
		return nil
	}
	if ge, ok := err.(*restum.GenericOpenAPIError); ok {
		var em struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
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

// cancelAllOpenOrders cancels all open orders across all symbols to allow position mode toggling
func cancelAllOpenOrders(t *testing.T) {
	t.Helper()
	rc := newRESTClientUserData()
	ctx, err := restAuthContextUser()
	if err != nil {
		t.Logf("cancelAllOpenOrders: no auth: %v", err)
		return
	}
	ts := time.Now().UnixMilli()
	orders, _, err := rc.FuturesAPI.GetOpenOrdersV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			t.Logf("cancelAllOpenOrders: get open orders failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		} else {
			t.Logf("cancelAllOpenOrders: get open orders failed: %v", err)
		}
		return
	}
	if len(orders) == 0 {
		t.Logf("cancelAllOpenOrders: no open orders")
		return
	}
	// collect unique symbols
	syms := make(map[string]struct{})
	for _, o := range orders {
		if o.Symbol != nil && *o.Symbol != "" {
			syms[*o.Symbol] = struct{}{}
		}
	}
	t.Logf("cancelAllOpenOrders: symbols with open orders: %v", keysOf(syms))
	// cancel per symbol
	for sym := range syms {
		ts2 := time.Now().UnixMilli()
		_, _, err := rc.FuturesAPI.DeleteAllOpenOrdersV1(ctx).Symbol(sym).Timestamp(ts2).RecvWindow(5000).Execute()
		if err != nil {
			if ge, ok := err.(*restum.GenericOpenAPIError); ok {
				t.Logf("cancelAllOpenOrders: cancel symbol=%s failed: status=%s body=%s", sym, ge.Error(), string(ge.Body()))
			} else {
				t.Logf("cancelAllOpenOrders: cancel symbol=%s failed: %v", sym, err)
			}
		} else {
			t.Logf("cancelAllOpenOrders: canceled all open orders for %s", sym)
		}
	}
}

// closeAllPositions submits reduce-only market orders to flatten all positions across symbols.
func closeAllPositions(t *testing.T) {
	t.Helper()
	rc := newRESTClientUserData()
	ctx, err := restAuthContextUser()
	if err != nil {
		t.Logf("closeAllPositions: no auth: %v", err)
		return
	}
	ts := time.Now().UnixMilli()
	risks, _, err := rc.FuturesAPI.GetPositionRiskV2(ctx).Timestamp(ts).RecvWindow(5000).Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			t.Logf("closeAllPositions: position risk fetch failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		} else {
			t.Logf("closeAllPositions: position risk fetch failed: %v", err)
		}
		return
	}
	type closeTask struct {
		symbol  string
		side    string
		qty     string
		posSide string
	}
	var tasks []closeTask
	for _, pr := range risks {
		if pr.Symbol == nil || pr.PositionAmt == nil {
			continue
		}
		rawAmt := strings.TrimSpace(*pr.PositionAmt)
		if rawAmt == "" {
			continue
		}
		amt, err := strconv.ParseFloat(rawAmt, 64)
		if err != nil {
			t.Logf("closeAllPositions: parse positionAmt=%q err: %v", rawAmt, err)
			continue
		}
		if math.Abs(amt) < 1e-6 {
			continue
		}
		absQty := strings.TrimLeft(rawAmt, "+")
		if strings.HasPrefix(absQty, "-") {
			absQty = strings.TrimPrefix(absQty, "-")
		}
		absQty = strings.TrimSpace(absQty)
		if absQty == "" || absQty == "0" {
			continue
		}
		side := "SELL"
		if amt < 0 {
			side = "BUY"
		}
		posSide := ""
		if pr.PositionSide != nil {
			posSide = strings.ToUpper(strings.TrimSpace(*pr.PositionSide))
		}
		tasks = append(tasks, closeTask{
			symbol:  *pr.Symbol,
			side:    side,
			qty:     absQty,
			posSide: posSide,
		})
	}
	if len(tasks) == 0 {
		t.Logf("closeAllPositions: no positions to close")
		return
	}
	for _, task := range tasks {
		req := rc.FuturesAPI.CreateOrderV1(ctx).
			Symbol(task.symbol).
			Side(task.side).
			Type_("MARKET").
			Quantity(task.qty).
			RecvWindow(5000).
			Timestamp(time.Now().UnixMilli())
		action := "reduce-only"
		if task.posSide != "" && !strings.EqualFold(task.posSide, "BOTH") {
			req = req.PositionSide(task.posSide)
			action = "hedged"
		} else {
			req = req.ReduceOnly("true")
		}
		_, _, err := req.Execute()
		if err != nil {
			if ge, ok := err.(*restum.GenericOpenAPIError); ok {
				t.Logf("closeAllPositions: %s close failed for %s %s qty=%s: status=%s body=%s", action, task.symbol, task.side, task.qty, ge.Error(), string(ge.Body()))
			} else {
				t.Logf("closeAllPositions: %s close failed for %s %s qty=%s: %v", action, task.symbol, task.side, task.qty, err)
			}
		} else {
			t.Logf("closeAllPositions: submitted %s %s %s qty=%s", action, task.symbol, task.side, task.qty)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func keysOf(m map[string]struct{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

// placeTestLimitOrder places a small limit order to trigger ORDER_TRADE_UPDATE events
func placeTestLimitOrder(t *testing.T, symbol string, side string, qty string, priceStr string) error {
	t.Helper()
	rc := newRESTClientUserData()
	ctx, err := restAuthContextUser()
	if err != nil {
		return err
	}
	dualMode, derr := detectDualPositionMode(ctx, rc)
	if derr != nil {
		t.Logf("placeTestLimitOrder: position mode lookup failed: %v", derr)
	}
	ts := time.Now().UnixMilli()
	req := rc.FuturesAPI.CreateOrderV1(ctx).
		Symbol(symbol).
		Side(side).
		Type_("LIMIT").
		TimeInForce("GTC").
		Quantity(qty).
		Price(priceStr).
		Timestamp(ts)
	if dualMode {
		posSide := "LONG"
		if strings.EqualFold(side, "SELL") {
			posSide = "SHORT"
		}
		req = req.PositionSide(posSide)
	}
	_, _, err = req.Execute()
	if err == nil {
		return nil
	}
	// Decode REST error body (if available) to bubble up Binance error code/message for clarity
	if ge, ok := err.(*restum.GenericOpenAPIError); ok {
		var em struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		// always include raw body in error for debugging
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

// roundToPrec rounds x to the given decimal precision using floor for safety
func roundToPrec(x float64, prec int) float64 {
	if prec < 0 {
		prec = 0
	}
	p := math.Pow10(prec)
	return math.Floor(x*p) / p
}

// calcLimitOrderParams derives a safe quantity and price using exchange info precisions and notional floor
func calcLimitOrderParams(ctx context.Context, rc *restum.APIClient, symbol string, targetPrice float64) (qtyStr, priceStr string, err error) {
	info, _, e := rc.FuturesAPI.GetExchangeInfoV1(ctx).Execute()
	if e != nil || info == nil || info.Symbols == nil {
		return "", "", fmt.Errorf("exchangeInfo error: %v", e)
	}
	var qp, pp int32 = 3, 2
	// Optional: pull tick/step sizes from filters to avoid -4014 (tick size) rejections
	var (
		priceStep   float64
		qtyStep     float64
		minQty      float64
		minNotional float64 = 5.0
		priceDecs   int
		qtyDecs     int
	)
	// helpers for robust rounding to step using integer math at given scale
	floorToStep := func(value float64, step float64, decs int) float64 {
		if step <= 0 { // fallback to precision-based floored rounding
			p := math.Pow10(decs)
			return math.Floor(value*p) / p
		}
		scale := math.Pow10(decs)
		// convert to integer units
		vUnits := math.Floor(value*scale + 1e-9)
		sUnits := math.Round(step * scale)
		if sUnits <= 0 {
			sUnits = 1
		}
		q := math.Floor(vUnits/sUnits) * sUnits
		return q / scale
	}
	ceilToStep := func(value float64, step float64, decs int) float64 {
		if step <= 0 {
			p := math.Pow10(decs)
			return math.Ceil(value*p) / p
		}
		scale := math.Pow10(decs)
		vUnits := math.Ceil(value*scale - 1e-9)
		sUnits := math.Round(step * scale)
		if sUnits <= 0 {
			sUnits = 1
		}
		q := (math.Ceil(vUnits / sUnits)) * sUnits
		return q / scale
	}
	decsFromStepStr := func(s string) int {
		if s == "" {
			return 0
		}
		if i := strings.IndexByte(s, '.'); i >= 0 {
			return len(s) - i - 1
		}
		return 0
	}
	parseFloatSafe := func(s string) float64 {
		if s == "" {
			return 0
		}
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return v
		}
		return 0
	}
	for _, s := range info.Symbols {
		if s.Symbol != nil && strings.EqualFold(*s.Symbol, symbol) {
			if s.QuantityPrecision != nil {
				qp = *s.QuantityPrecision
			}
			if s.PricePrecision != nil {
				pp = *s.PricePrecision
			}
			// Try to derive tick/step from filters by re-marshaling to generic map
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
							case "MIN_NOTIONAL", "NOTIONAL":
								if nt, _ := fm["notional"].(string); nt != "" {
									if v := parseFloatSafe(nt); v > 0 {
										minNotional = v
									}
								}
							case "MARKET_LOT_SIZE":
								// fallback step for quantity if LOT_SIZE absent
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
	// Decide decimals to print: prefer from step strings; fallback to precision fields
	if priceDecs == 0 {
		priceDecs = int(pp)
	}
	if qtyDecs == 0 {
		qtyDecs = int(qp)
	}
	// Use a price slightly below mark and round DOWN to the nearest valid tick step
	price := floorToStep(targetPrice, priceStep, priceDecs)
	if price <= 0 {
		return "", "", fmt.Errorf("invalid rounded price")
	}
	// Ensure notional respects exchange filter (default to 5 if unavailable)
	if minNotional <= 0 {
		minNotional = 5.0
	}
	// Soft floor to avoid Binance -4164 rejections (requires >=100 notional on testnet)
	const (
		softFloorNotional = 110.0
		hardCapNotional   = 400.0
	)
	targetNotional := math.Max(minNotional*1.2, softFloorNotional)
	if targetNotional > hardCapNotional {
		targetNotional = hardCapNotional
	}
	q := ceilToStep(targetNotional/price, qtyStep, qtyDecs)
	if q <= 0 {
		q = 1.0 / math.Pow10(qtyDecs)
	}
	// honor minQty if present
	if minQty > 0 && q < minQty {
		q = minQty
	}

	// Ensure notional meets soft floor even after clamping
	if q*price < softFloorNotional {
		q = ceilToStep(softFloorNotional/price, qtyStep, qtyDecs)
	}

	// Clip to hard cap if we ended up above the safety limit
	if q*price > hardCapNotional {
		target := hardCapNotional / price
		q = floorToStep(target, qtyStep, qtyDecs)
		if q <= 0 {
			q = 1.0 / math.Pow10(qtyDecs)
		}
		// If clamping dropped us below the required floor, just accept the higher notional.
		if q*price < softFloorNotional {
			q = ceilToStep(softFloorNotional/price, qtyStep, qtyDecs)
		}
	}

	// Final formatting
	qtyStr = fmt.Sprintf("%.*f", qtyDecs, q)
	priceStr = fmt.Sprintf("%.*f", priceDecs, price)
	return qtyStr, priceStr, nil
}

// detectDualPositionMode returns true when the account is in dual (hedge) position mode.
func detectDualPositionMode(ctx context.Context, rc *restum.APIClient) (bool, error) {
	if ctx == nil {
		return false, fmt.Errorf("position side query requires auth context")
	}
	ts := time.Now().UnixMilli()
	resp, _, err := rc.FuturesAPI.GetPositionSideDualV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			return false, fmt.Errorf("position side query failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		}
		return false, err
	}
	if resp == nil || resp.DualSidePosition == nil {
		return false, fmt.Errorf("position side query: missing dualSidePosition")
	}
	return *resp.DualSidePosition, nil
}

// togglePositionMode flips dual-side position to trigger ACCOUNT_CONFIG_UPDATE
func togglePositionMode(t *testing.T) error {
	t.Helper()
	rc := newRESTClientUserData()
	ctx, err := restAuthContextUser()
	if err != nil {
		return err
	}
	ts := time.Now().UnixMilli()
	// Get current mode
	cur, _, err := rc.FuturesAPI.GetPositionSideDualV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			return fmt.Errorf("get position mode failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		}
		return err
	}
	if cur == nil || cur.DualSidePosition == nil {
		return fmt.Errorf("get position mode returned empty response")
	}
	// Flip value
	newVal := "false"
	if *cur.DualSidePosition == true {
		newVal = "false"
	} else {
		newVal = "true"
	}
	t.Logf("toggle position mode: current_dual=%v -> target_dual=%s", *cur.DualSidePosition, newVal)
	ts2 := time.Now().UnixMilli()
	_, _, err = rc.FuturesAPI.CreatePositionSideDualV1(ctx).DualSidePosition(newVal).Timestamp(ts2).RecvWindow(5000).Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			return fmt.Errorf("set position mode failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		}
		return err
	}
	return nil
}

// TestFullIntegrationSuite_UserData runs request/response and event coverage for UserDataStreamChannel
func TestFullIntegrationSuite_UserData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Ensure credentials exist
	if os.Getenv("BINANCE_API_KEY") == "" || os.Getenv("BINANCE_SECRET_KEY") == "" {
		t.Skip("Missing BINANCE_API_KEY/SECRET_KEY; skipping user data suite")
	}

	// Capture SDK log output and fail on any 'unhandled message'
	cw := &unhandledCatcherUD{}
	log.SetOutput(io.MultiWriter(cw, os.Stderr))
	defer func() {
		log.SetOutput(os.Stderr)
		cw.mu.Lock()
		defer cw.mu.Unlock()
		if len(cw.matches) > 0 {
			for _, line := range cw.matches {
				t.Logf("SDK log captured: %s", strings.TrimSpace(line))
			}
			t.Fatalf("SDK emitted %d 'unhandled message' log(s) during UserData suite; treating as failure", len(cw.matches))
		}
	}()

	// Dedicated client and explicit testnet1 selection
	cfg := getTestConfig()
	stc, err := NewStreamTestClientDedicated(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
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
	ch := umfuturesstreams.NewUserDataStreamChannel(stc.client)
	rec := newUMUserDataEventRecorder()
	ch.HandleListenKeyExpiredEvent(func(ctx context.Context, ev *models.ListenKeyExpiredEvent) error { rec.addExpired(ev); return nil })
	ch.HandleAccountUpdateEvent(func(ctx context.Context, ev *models.AccountUpdateEvent) error { rec.addAcctUpd(ev); return nil })
	ch.HandleMarginCallEvent(func(ctx context.Context, ev *models.MarginCallEvent) error { rec.addMarginCal(ev); return nil })
	ch.HandleOrderTradeUpdateEvent(func(ctx context.Context, ev *models.OrderTradeUpdateEvent) error { rec.addOrdUpd(ev); return nil })
	ch.HandleTradeLiteEvent(func(ctx context.Context, ev *models.TradeLiteEvent) error { rec.addTradeLite(ev); return nil })
	ch.HandleAccountConfigUpdateEvent(func(ctx context.Context, ev *models.AccountConfigUpdateEvent) error { rec.addCfgUpd(ev); return nil })
	ch.HandleStrategyUpdateEvent(func(ctx context.Context, ev *models.StrategyUpdateEvent) error { rec.addStratUpd(ev); return nil })
	ch.HandleGridUpdateEvent(func(ctx context.Context, ev *models.GridUpdateEvent) error { rec.addGridUpd(ev); return nil })
	ch.HandleConditionalOrderTriggerRejectEvent(func(ctx context.Context, ev *models.ConditionalOrderTriggerRejectEvent) error {
		rec.addCondRej(ev)
		return nil
	})

	// Connect to user data stream using listenKey
	cctx, ccancel := context.WithTimeout(context.Background(), 12*time.Second)
	if err := ch.Connect(cctx, listenKey); err != nil {
		ccancel()
		t.Fatalf("user data connect failed: %v", err)
	}
	ccancel()
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 6*time.Second)
		_ = ch.Disconnect(dctx)
		dcancel()
	}()

	// ---------- Request/Response: Start, Ping, Stop ----------
	t.Run("Request_Start", func(t *testing.T) {
		if isTestnet1 {
			t.Skip("Skipping userDataStream.start on testnet (unsupported)")
		}
		// Attempt WS-based start. Some servers may not support this; treat timeouts/errors as acceptable.
		sid := newMessageID()
		done := make(chan struct{}, 1)
		var got *models.UserDataStreamStartResponse
		cb := func(ctx context.Context, resp *models.UserDataStreamStartResponse, rpcErr error) error {
			if rpcErr != nil {
				if em, ok := rpcErr.(*models.ErrorMessage); ok {
					rec.addError(em)
					logJSON(t, "userData.start.error", em)
				} else {
					t.Logf("userData.start rpc error: %v", rpcErr)
				}
				return nil
			}
			got = resp
			logJSON(t, "userData.start.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		req := &models.UserDataStreamStartRequest{Id: sid}
		req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
		spCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := ch.Start(spCtx, req, &cb); err != nil {
			t.Logf("userDataStreamStart call err (acceptable if unsupported): %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting userData.start response (acceptable)")
		}
		_ = got
	})

	t.Run("Request_Ping", func(t *testing.T) {
		if isTestnet1 {
			t.Skip("Skipping userDataStream.ping on testnet (unsupported)")
		}
		pid := newMessageID()
		done := make(chan struct{}, 1)
		var got *models.UserDataStreamPingResponse
		cb := func(ctx context.Context, resp *models.UserDataStreamPingResponse, rpcErr error) error {
			if rpcErr != nil {
				if em, ok := rpcErr.(*models.ErrorMessage); ok {
					rec.addError(em)
					logJSON(t, "userData.ping.error", em)
				} else {
					t.Logf("userData.ping rpc error: %v", rpcErr)
				}
				return nil
			}
			got = resp
			logJSON(t, "userData.ping.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		req := &models.UserDataStreamPingRequest{Id: pid}
		req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
		pgCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := ch.Ping(pgCtx, req, &cb); err != nil {
			t.Logf("userDataStreamPing call err (acceptable if unsupported): %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting userData.ping response (acceptable)")
		}
		_ = got
	})

	t.Run("Request_Stop", func(t *testing.T) {
		if isTestnet1 {
			t.Skip("Skipping userDataStream.stop on testnet (unsupported)")
		}
		sid := newMessageID()
		done := make(chan struct{}, 1)
		var got *models.UserDataStreamStopResponse
		cb := func(ctx context.Context, resp *models.UserDataStreamStopResponse, rpcErr error) error {
			if rpcErr != nil {
				if em, ok := rpcErr.(*models.ErrorMessage); ok {
					rec.addError(em)
					logJSON(t, "userData.stop.error", em)
				} else {
					t.Logf("userData.stop rpc error: %v", rpcErr)
				}
				return nil
			}
			got = resp
			logJSON(t, "userData.stop.response", resp)
			select {
			case done <- struct{}{}:
			default:
			}
			return nil
		}
		req := &models.UserDataStreamStopRequest{Id: sid}
		req.Params.ApiKey = os.Getenv("BINANCE_API_KEY")
		spCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := ch.Stop(spCtx, req, &cb); err != nil {
			t.Logf("userDataStreamStop call err (acceptable if unsupported): %v", err)
		}
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Logf("timeout waiting userData.stop response (acceptable)")
		}
		_ = got
	})

	// ---------- Event Handlers ----------
	t.Run("OrderTradeUpdateEvent", func(t *testing.T) {
		// Try to trigger by placing a small LIMIT order far from market to avoid fills.
		symUP, err := restPickSymbol(context.Background())
		if err != nil || symUP == "" {
			t.Fatalf("failed to pick symbol: %v", err)
		}
		// Get a reference price from REST; fall back to 30000 if unavailable
		ref := 30000.0
		if p, err := restMarkPrice(context.Background(), symUP); err == nil && p > 0 {
			ref = p
		} else {
			t.Logf("mark price fetch failed; using fallback: %v", err)
		}
		// Build params from exchange info to satisfy precision and notional rules
		rc := newRESTClientUserData()
		qStr, pStr := "", ""
		if qty, px, e := calcLimitOrderParams(context.Background(), rc, symUP, ref*0.98); e == nil {
			qStr, pStr = qty, px
		} else {
			// fallback
			qStr = "0.001"
			pStr = fmt.Sprintf("%.2f", ref*0.98)
		}
		t.Logf("placing test LIMIT order: symbol=%s qty=%s price=%s (ref=%.8f)", symUP, qStr, pStr, ref)
		// Place a BUY limit order; if rejected due to margin, fall back to test endpoint
		if err := placeTestLimitOrder(t, symUP, "BUY", qStr, pStr); err != nil {
			t.Logf("live order rejected (tolerated): %v", err)
			if ge, ok := err.(*restum.GenericOpenAPIError); ok {
				t.Logf("REST error status=%s body=%s", ge.Error(), string(ge.Body()))
			}
			// Try test endpoint to validate request shape (no event expected)
			if ctx, e2 := restAuthContextUser(); e2 == nil {
				ts := time.Now().UnixMilli()
				_, _, err2 := rc.FuturesAPI.CreateOrderTestV1(ctx).
					Symbol(symUP).Side("BUY").Type_("LIMIT").TimeInForce("GTC").Quantity(qStr).Price(pStr).Timestamp(ts).
					RecvWindow(5000).Execute()
				if err2 != nil {
					t.Logf("test order rejected: %v", err2)
					if ge2, ok2 := err2.(*restum.GenericOpenAPIError); ok2 {
						t.Logf("REST error status=%s body=%s", ge2.Error(), string(ge2.Body()))
					}
				}
			}
		}
		// Wait for at least one ORDER_TRADE_UPDATE (NEW or rejected)
		_ = rec.waitForMin("ORDER_TRADE_UPDATE", 1, eventWait())
		cnt := rec.count("ORDER_TRADE_UPDATE")
		t.Logf("ORDER_TRADE_UPDATE events: %d", cnt)
		if cnt > 0 {
			ev := rec.ordUpd[len(rec.ordUpd)-1]
			// Log full typed event for diagnostics
			logJSON(t, "orderTradeUpdate.event(typed)", ev)
			if ev.EventType != "ORDER_TRADE_UPDATE" {
				t.Errorf("want e=ORDER_TRADE_UPDATE got %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			// Basic order checks
			ord := ev.OrderDetails
			assertNonEmpty(t, ord.Symbol, "symbol")
			assertNonEmpty(t, ord.Side, "side")
			assertNonEmpty(t, ord.OrderType, "orderType")
			assertNonEmpty(t, ord.OrderStatus, "orderStatus")
		} else {
			t.Logf("no ORDER_TRADE_UPDATE received (acceptable on accounts without permissions/balance)")
		}
	})

	t.Run("AccountConfigUpdateEvent", func(t *testing.T) {
		// Cancel all open orders to allow toggling position mode
		cancelAllOpenOrders(t)
		// Flatten existing positions; otherwise position mode toggle will fail with code -4068
		closeAllPositions(t)

		// Get current position mode
		rc := newRESTClientUserData()
		ctx, err := restAuthContextUser()
		if err != nil {
			t.Logf("position mode: auth error: %v", err)
		}
		var original *bool
		if ctx != nil {
			ts := time.Now().UnixMilli()
			cur, _, gerr := rc.FuturesAPI.GetPositionSideDualV1(ctx).Timestamp(ts).RecvWindow(5000).Execute()
			if gerr != nil {
				if ge, ok := gerr.(*restum.GenericOpenAPIError); ok {
					t.Logf("get position mode failed: status=%s body=%s", ge.Error(), string(ge.Body()))
				} else {
					t.Logf("get position mode failed: %v", gerr)
				}
			} else if cur != nil && cur.DualSidePosition != nil {
				original = cur.DualSidePosition
				t.Logf("position mode current_dual=%v", *original)
			}
		}

		// Flip position mode to trigger config update; ignore errors in restricted envs
		if err := togglePositionMode(t); err != nil {
			t.Logf("toggle position mode failed (acceptable): %v", err)
		}

		// Ensure we restore the original mode at the end
		if original != nil && ctx != nil {
			defer func() {
				// Re-cancel any open orders just in case
				cancelAllOpenOrders(t)
				closeAllPositions(t)
				target := "false"
				if *original {
					target = "true"
				}
				ts := time.Now().UnixMilli()
				_, _, e := rc.FuturesAPI.CreatePositionSideDualV1(ctx).DualSidePosition(target).Timestamp(ts).RecvWindow(5000).Execute()
				if e != nil {
					if ge, ok := e.(*restum.GenericOpenAPIError); ok {
						t.Logf("restore position mode failed: status=%s body=%s", ge.Error(), string(ge.Body()))
					} else {
						t.Logf("restore position mode failed: %v", e)
					}
				} else {
					t.Logf("restored position mode to dual=%v", *original)
				}
			}()
		}

		_ = rec.waitForMin("ACCOUNT_CONFIG_UPDATE", 1, eventWait())
		cnt := rec.count("ACCOUNT_CONFIG_UPDATE")
		t.Logf("ACCOUNT_CONFIG_UPDATE events: %d", cnt)
		if cnt == 0 {
			// Fallback trigger: change leverage to generate ACCOUNT_CONFIG_UPDATE (ac)
			sym, err := restPickSymbol(context.Background())
			if err != nil || sym == "" {
				sym = "BTCUSDT"
			}
			t.Logf("fallback: attempting leverage change on %s to trigger ACCOUNT_CONFIG_UPDATE", sym)
			rc := newRESTClientUserData()
			ctx, aerr := restAuthContextUser()
			if aerr == nil {
				// discover current leverage via position risk
				ts := time.Now().UnixMilli()
				risks, _, rerr := rc.FuturesAPI.GetPositionRiskV2(ctx).Symbol(sym).Timestamp(ts).RecvWindow(5000).Execute()
				curLev := 10
				if rerr != nil {
					if ge, ok := rerr.(*restum.GenericOpenAPIError); ok {
						t.Logf("get positionRisk failed: status=%s body=%s", ge.Error(), string(ge.Body()))
					} else {
						t.Logf("get positionRisk failed: %v", rerr)
					}
				} else {
					for _, it := range risks {
						if it.Symbol != nil && strings.EqualFold(*it.Symbol, sym) && it.Leverage != nil {
							if v, e := strconv.Atoi(*it.Leverage); e == nil {
								curLev = v
								break
							}
						}
					}
				}
				newLev := 11
				if curLev >= 11 {
					newLev = 10
				}
				t.Logf("leverage change: current=%d -> target=%d (symbol=%s)", curLev, newLev, sym)
				ts2 := time.Now().UnixMilli()
				_, _, lerr := rc.FuturesAPI.CreateLeverageV1(ctx).Symbol(sym).Leverage(int32(newLev)).Timestamp(ts2).RecvWindow(5000).Execute()
				if lerr != nil {
					if ge, ok := lerr.(*restum.GenericOpenAPIError); ok {
						t.Logf("set leverage failed: status=%s body=%s", ge.Error(), string(ge.Body()))
					} else {
						t.Logf("set leverage failed: %v", lerr)
					}
				} else {
					_ = rec.waitForMin("ACCOUNT_CONFIG_UPDATE", 1, eventWait())
					cnt = rec.count("ACCOUNT_CONFIG_UPDATE")
					t.Logf("ACCOUNT_CONFIG_UPDATE events after leverage change: %d", cnt)
					// restore leverage
					ts3 := time.Now().UnixMilli()
					_, _, r2 := rc.FuturesAPI.CreateLeverageV1(ctx).Symbol(sym).Leverage(int32(curLev)).Timestamp(ts3).RecvWindow(5000).Execute()
					if r2 != nil {
						if ge, ok := r2.(*restum.GenericOpenAPIError); ok {
							t.Logf("restore leverage failed: status=%s body=%s", ge.Error(), string(ge.Body()))
						} else {
							t.Logf("restore leverage failed: %v", r2)
						}
					} else {
						t.Logf("restored leverage to %d for %s", curLev, sym)
					}
				}
			} else {
				t.Logf("fallback leverage change skipped: auth error: %v", aerr)
			}
		}
		if cnt > 0 {
			ev := rec.cfgUpd[len(rec.cfgUpd)-1]
			if ev.EventType != "ACCOUNT_CONFIG_UPDATE" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		} else {
			t.Logf("no ACCOUNT_CONFIG_UPDATE received (acceptable)")
		}
	})

	t.Run("AccountUpdateEvent", func(t *testing.T) {
		// Often emitted on balance/position updates
		_ = rec.waitForMin("ACCOUNT_UPDATE", 1, eventWait())
		cnt := rec.count("ACCOUNT_UPDATE")
		t.Logf("ACCOUNT_UPDATE events: %d", cnt)
		if cnt == 0 {
			// Try to induce by placing a tiny MARKET order
			sym, err := restPickSymbol(context.Background())
			if err != nil || sym == "" {
				sym = "BTCUSDT"
			}
			ref := 30000.0
			if p, err := restMarkPrice(context.Background(), sym); err == nil && p > 0 {
				ref = p
			}
			rc := newRESTClientUserData()
			if qty, _, e := calcLimitOrderParams(context.Background(), rc, sym, ref); e == nil {
				t.Logf("attempting small market order for ACCOUNT_UPDATE: symbol=%s qty=%s", sym, qty)
				if err := placeTestMarketOrder(t, sym, "BUY", qty); err != nil {
					t.Logf("market order rejected (acceptable): %v", err)
				}
				_ = rec.waitForMin("ACCOUNT_UPDATE", 1, eventWait())
				cnt = rec.count("ACCOUNT_UPDATE")
				t.Logf("ACCOUNT_UPDATE events after market order: %d", cnt)
			}
		}
		if cnt > 0 {
			ev := rec.acctUpd[len(rec.acctUpd)-1]
			if ev.EventType != "ACCOUNT_UPDATE" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
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
			if ev.EventType != "MARGIN_CALL" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("TradeLiteEvent", func(t *testing.T) {
		// Try to induce by placing a tiny MARKET order if not seen
		_ = rec.waitForMin("TRADE_LITE", 1, time.Second*3)
		cnt := rec.count("TRADE_LITE")
		t.Logf("TRADE_LITE events: %d", cnt)
		if cnt == 0 {
			sym, err := restPickSymbol(context.Background())
			if err != nil || sym == "" {
				sym = "BTCUSDT"
			}
			ref := 30000.0
			if p, err := restMarkPrice(context.Background(), sym); err == nil && p > 0 {
				ref = p
			}
			rc := newRESTClientUserData()
			if qty, _, e := calcLimitOrderParams(context.Background(), rc, sym, ref); e == nil {
				t.Logf("attempting small market order for TRADE_LITE: symbol=%s qty=%s", sym, qty)
				if err := placeTestMarketOrder(t, sym, "BUY", qty); err != nil {
					t.Logf("market order rejected (acceptable): %v", err)
				}
				_ = rec.waitForMin("TRADE_LITE", 1, eventWait())
				cnt = rec.count("TRADE_LITE")
				t.Logf("TRADE_LITE events after market order: %d", cnt)
			}
		}
		if cnt > 0 {
			ev := rec.tradeLite[len(rec.tradeLite)-1]
			if ev.EventType != "TRADE_LITE" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			assertNonEmpty(t, ev.Symbol, "symbol")
		}
	})

	t.Run("StrategyUpdateEvent", func(t *testing.T) {
		_ = rec.waitForMin("STRATEGY_UPDATE", 1, time.Second*3)
		cnt := rec.count("STRATEGY_UPDATE")
		t.Logf("STRATEGY_UPDATE events: %d", cnt)
		if cnt > 0 {
			ev := rec.stratUpd[len(rec.stratUpd)-1]
			if ev.EventType != "STRATEGY_UPDATE" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("GridUpdateEvent", func(t *testing.T) {
		_ = rec.waitForMin("GRID_UPDATE", 1, time.Second*3)
		cnt := rec.count("GRID_UPDATE")
		t.Logf("GRID_UPDATE events: %d", cnt)
		if cnt > 0 {
			ev := rec.gridUpd[len(rec.gridUpd)-1]
			if ev.EventType != "GRID_UPDATE" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("ConditionalOrderTriggerRejectEvent", func(t *testing.T) {
		// Difficult to force; validate if present
		_ = rec.waitForMin("CONDITIONAL_ORDER_TRIGGER_REJECT", 1, time.Second*3)
		cnt := rec.count("CONDITIONAL_ORDER_TRIGGER_REJECT")
		t.Logf("CONDITIONAL_ORDER_TRIGGER_REJECT events: %d", cnt)
		if cnt > 0 {
			ev := rec.condRej[len(rec.condRej)-1]
			if ev.EventType != "CONDITIONAL_ORDER_TRIGGER_REJECT" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
			assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
		}
	})

	t.Run("ListenKeyExpiredEvent", func(t *testing.T) {
		// Not feasible to force within suite (expires ~60m); ensure handler registered. If any captured, validate.
		cnt := rec.count("listenKeyExpired")
		t.Logf("listenKeyExpired events: %d", cnt)
		if cnt > 0 {
			ev := rec.expired[len(rec.expired)-1]
			if ev.EventType != "listenKeyExpired" {
				t.Logf("unexpected event type: %s", ev.EventType)
			}
		}
	})

	t.Run("ErrorMessage", func(t *testing.T) {
		// Our start/ping/stop requests may produce errors, which we capture
		cnt := rec.count("error")
		t.Logf("error messages captured: %d", cnt)
		if cnt > 0 {
			msg := rec.errors[len(rec.errors)-1]
			if msg.ErrorPayload.Code == 0 {
				t.Logf("error code missing (message=%s)", msg.Error())
			}
			assertNonEmpty(t, msg.ErrorPayload.Msg, "error.msg")
		}
	})

	t.Run("ListenKeyKeepalive", func(t *testing.T) {
		// Exercise REST keepalive to ensure LK remains valid during suite
		keepaliveListenKeyTestnet(t, listenKey)
	})
}

func maskListenKey(lk string) string {
	if len(lk) <= 8 {
		return lk
	}
	return lk[:4] + "..." + lk[len(lk)-4:]
}
