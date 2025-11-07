package wstest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	restspot "github.com/openxapi/binance-go/rest/spot"
)

var errSpotCredentialsMissing = errors.New("spot REST credentials not configured")

const defaultSpotRestTestnetServer = "https://testnet.binance.vision"

func newRESTClient() *restspot.APIClient {
	cfg := restspot.NewConfiguration()
	if s := strings.TrimSpace(os.Getenv("BINANCE_SPOT_REST_SERVER")); s != "" {
		cfg.Servers[0].URL = s
	} else if len(cfg.Servers) > 0 && strings.TrimSpace(cfg.Servers[0].URL) != "" {
		cfg.Servers[0].URL = defaultSpotRestTestnetServer
	} else {
		cfg.Servers = []restspot.ServerConfiguration{{
			URL:         defaultSpotRestTestnetServer,
			Description: "Binance Spot Testnet",
		}}
	}
	return restspot.NewAPIClient(cfg)
}

type spotSymbolConstraints struct {
	priceStep     float64
	qtyStep       float64
	minQty        float64
	minNotional   float64
	priceDecimals int
	qtyDecimals   int
}

func loadSpotSymbolConstraints(ctx context.Context, rc *restspot.APIClient, symbol string) (*spotSymbolConstraints, error) {
	if rc == nil {
		rc = newRESTClient()
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	info, _, err := rc.SpotTradingAPI.GetExchangeInfoV3(cctx).Symbol(symbol).Execute()
	if err != nil || info == nil || len(info.Symbols) == 0 {
		if err != nil {
			return nil, fmt.Errorf("exchangeInfo error: %w", err)
		}
		return nil, fmt.Errorf("exchangeInfo returned no symbols for %s", symbol)
	}
	var sym *restspot.SpotGetExchangeInfoV3RespSymbolsInner
	for i := range info.Symbols {
		if info.Symbols[i].Symbol != nil && strings.EqualFold(*info.Symbols[i].Symbol, symbol) {
			sym = &info.Symbols[i]
			break
		}
	}
	if sym == nil {
		sym = &info.Symbols[0]
	}
	constraints := &spotSymbolConstraints{
		priceDecimals: 8,
		qtyDecimals:   8,
		minNotional:   10.0,
	}
	for _, f := range sym.Filters {
		switch strings.ToUpper(f.GetFilterType()) {
		case "PRICE_FILTER":
			if ts := f.GetTickSize(); ts != "" {
				constraints.priceStep = parseFloat64(ts)
				if dec := decimals(ts); dec > 0 {
					constraints.priceDecimals = dec
				}
			}
		case "LOT_SIZE":
			if ss := f.GetStepSize(); ss != "" {
				constraints.qtyStep = parseFloat64(ss)
				if dec := decimals(ss); dec > 0 {
					constraints.qtyDecimals = dec
				}
			}
			if mq := f.GetMinQty(); mq != "" {
				constraints.minQty = parseFloat64(mq)
			}
		case "MIN_NOTIONAL":
			if mv := f.GetMinNotionalValue(); mv != "" {
				if v := parseFloat64(mv); v > 0 {
					constraints.minNotional = v
				}
			}
		}
	}
	return constraints, nil
}

func fetchSorSymbols(ctx context.Context, rc *restspot.APIClient) ([]string, error) {
	client := rc
	if client == nil {
		client = newRESTClient()
	}
	base := ctx
	if base == nil {
		base = context.Background()
	}
	cctx, cancel := context.WithTimeout(base, 8*time.Second)
	defer cancel()
	resp, _, err := client.SpotTradingAPI.GetExchangeInfoV3(cctx).Execute()
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Sors) == 0 {
		return nil, nil
	}
	var symbols []string
	for _, sor := range resp.Sors {
		for _, sym := range sor.Symbols {
			if sym == "" {
				continue
			}
			symbols = append(symbols, strings.ToUpper(sym))
		}
	}
	return symbols, nil
}

func spotSymbolSupportsSOR(ctx context.Context, rc *restspot.APIClient, symbol string) bool {
	if strings.TrimSpace(symbol) == "" {
		return false
	}
	symbol = strings.ToUpper(symbol)
	sors, err := fetchSorSymbols(ctx, rc)
	if err != nil || len(sors) == 0 {
		return false
	}
	for _, sym := range sors {
		if strings.EqualFold(sym, symbol) {
			return true
		}
	}
	return false
}

func spotPickSorSymbol(ctx context.Context, rc *restspot.APIClient) (string, error) {
	sors, err := fetchSorSymbols(ctx, rc)
	if err != nil {
		return "", err
	}
	if len(sors) == 0 {
		return "", fmt.Errorf("no SOR-supported symbols available")
	}
	return sors[0], nil
}

func prepareSorLimitOrder(ctx context.Context, rc *restspot.APIClient, symbol string) (string, string, error) {
	if rc == nil {
		rc = newRESTClient()
	}
	base := ctx
	if base == nil {
		base = context.Background()
	}
	priceRef, err := restTickerPrice(base, rc, symbol)
	if err != nil || priceRef <= 0 {
		if err != nil {
			return "", "", fmt.Errorf("ticker price fetch failed for %s: %w", symbol, err)
		}
		return "", "", fmt.Errorf("ticker price unavailable for %s", symbol)
	}
	target := priceRef * 0.99
	if target <= 0 {
		target = priceRef
	}
	qty, price, _, calcErr := calcSpotLimitOrderParams(base, rc, symbol, target)
	if calcErr != nil {
		return "", "", calcErr
	}
	return qty, price, nil
}

// restPickSymbol returns an actively traded symbol (prefers BTCUSDT).
func restPickSymbol(ctx context.Context) (string, error) {
	if pref := os.Getenv("PREFERRED_SYMBOL"); pref != "" {
		parts := strings.Split(pref, ",")
		for _, part := range parts {
			sym := strings.TrimSpace(part)
			if sym == "" {
				continue
			}
			if symbolExists(ctx, strings.ToUpper(sym)) {
				return strings.ToUpper(sym), nil
			}
		}
	}
	rc := newRESTClient()
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	info, _, err := rc.SpotTradingAPI.GetExchangeInfoV3(cctx).Execute()
	if err != nil || info == nil || info.Symbols == nil || len(info.Symbols) == 0 {
		return "BTCUSDT", nil
	}
	for _, s := range info.Symbols {
		if s.Symbol != nil && strings.EqualFold(*s.Symbol, "BTCUSDT") {
			return "BTCUSDT", nil
		}
	}
	if info.Symbols[0].Symbol != nil {
		return strings.ToUpper(*info.Symbols[0].Symbol), nil
	}
	return "BTCUSDT", nil
}

func symbolExists(ctx context.Context, symbol string) bool {
	rc := newRESTClient()
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	info, _, err := rc.SpotTradingAPI.GetExchangeInfoV3(cctx).Execute()
	if err != nil || info == nil || info.Symbols == nil {
		return false
	}
	for _, s := range info.Symbols {
		if s.Symbol != nil && strings.EqualFold(*s.Symbol, symbol) {
			return true
		}
	}
	return false
}

func restAuthContext() (context.Context, error) {
	apiKey := strings.TrimSpace(os.Getenv("BINANCE_API_KEY"))
	secret := strings.TrimSpace(os.Getenv("BINANCE_SECRET_KEY"))
	if apiKey == "" || secret == "" {
		return nil, errSpotCredentialsMissing
	}
	auth := restspot.NewAuth(apiKey)
	auth.SetSecretKey(secret)
	return auth.ContextWithValue(context.Background())
}

func restTickerPrice(ctx context.Context, rc *restspot.APIClient, symbol string) (float64, error) {
	if rc == nil {
		rc = newRESTClient()
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, _, err := rc.SpotTradingAPI.GetTickerPriceV3(cctx).Symbol(symbol).Execute()
	if err != nil || resp == nil {
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("ticker price response empty")
	}
	if inst := resp.GetActualInstance(); inst != nil {
		switch v := inst.(type) {
		case *restspot.SpotGetTickerPriceV3RespItem:
			return parseFloat64(v.GetPrice()), nil
		case *[]restspot.SpotGetTickerPriceV3RespItem:
			items := *v
			if len(items) == 0 {
				return 0, fmt.Errorf("ticker price response missing entries")
			}
			for _, it := range items {
				if strings.EqualFold(it.GetSymbol(), symbol) {
					return parseFloat64(it.GetPrice()), nil
				}
			}
			return parseFloat64(items[0].GetPrice()), nil
		}
	}
	return 0, fmt.Errorf("unexpected ticker price response type")
}

func calcSpotLimitOrderParams(ctx context.Context, rc *restspot.APIClient, symbol string, targetPrice float64) (qtyStr, priceStr string, constraints *spotSymbolConstraints, err error) {
	constraints, err = loadSpotSymbolConstraints(ctx, rc, symbol)
	if err != nil {
		return "", "", nil, err
	}

	price := floorToStep(targetPrice, constraints.priceStep, constraints.priceDecimals)
	if price <= 0 {
		return "", "", nil, fmt.Errorf("invalid rounded price for %s", symbol)
	}

	requiredQty := constraints.minNotional / price
	if constraints.minQty > 0 && requiredQty < constraints.minQty {
		requiredQty = constraints.minQty
	}
	qty := ceilToStep(requiredQty, constraints.qtyStep, constraints.qtyDecimals)
	if qty <= 0 {
		qty = 1.0 / math.Pow10(constraints.qtyDecimals)
	}
	if qty*price > 50.0 {
		target := 50.0 / price
		qty = floorToStep(target, constraints.qtyStep, constraints.qtyDecimals)
		if qty <= 0 {
			qty = 1.0 / math.Pow10(constraints.qtyDecimals)
		}
	}
	qtyStr = fmt.Sprintf("%.*f", constraints.qtyDecimals, qty)
	priceStr = fmt.Sprintf("%.*f", constraints.priceDecimals, price)
	return qtyStr, priceStr, constraints, nil
}

type spotPlacedOrder struct {
	symbol        string
	orderID       int64
	clientOrderID string
}

type spotPlacedOrderList struct {
	symbol            string
	orderListID       int64
	listClientOrderID string
}

func placeSpotLimitOrder(rc *restspot.APIClient, symbol, side, qty, price string) (*spotPlacedOrder, error) {
	if rc == nil {
		rc = newRESTClient()
	}
	ctx, err := restAuthContext()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ts := time.Now().UnixMilli()
	clientOrderID := fmt.Sprintf("wstest-%d", ts)
	req := rc.SpotTradingAPI.CreateOrderV3(ctx).
		Symbol(symbol).
		Side(side).
		Type_("LIMIT").
		TimeInForce("GTC").
		Quantity(qty).
		Price(price).
		NewClientOrderId(clientOrderID).
		Timestamp(ts).
		RecvWindow(5000)
	resp, _, err := req.Execute()
	if err != nil {
		return nil, decodeSpotError(err)
	}
	order := &spotPlacedOrder{symbol: symbol, clientOrderID: clientOrderID}
	if resp != nil {
		if resp.HasOrderId() {
			order.orderID = resp.GetOrderId()
		}
		if resp.HasClientOrderId() && resp.GetClientOrderId() != "" {
			order.clientOrderID = resp.GetClientOrderId()
		}
	}
	return order, nil
}

func placeSpotOcoOrder(rc *restspot.APIClient, symbol, side, quantity, belowPrice, aboveStopPrice, aboveLimitPrice string) (*spotPlacedOrderList, error) {
	if rc == nil {
		rc = newRESTClient()
	}
	ctx, err := restAuthContext()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	ts := time.Now().UnixMilli()
	listClientOrderID := newClientOrderID("oco", ts)
	aboveClientOrderID := newClientOrderID("oco-above", ts+1)
	belowClientOrderID := newClientOrderID("oco-below", ts+2)
	req := rc.SpotTradingAPI.CreateOrderListOcoV3(ctx).
		Symbol(symbol).
		Side(side).
		Quantity(quantity).
		AboveType("STOP_LOSS_LIMIT").
		AboveStopPrice(aboveStopPrice).
		AbovePrice(aboveLimitPrice).
		AboveTimeInForce("GTC").
		AboveClientOrderId(aboveClientOrderID).
		BelowType("LIMIT_MAKER").
		BelowPrice(belowPrice).
		BelowClientOrderId(belowClientOrderID).
		ListClientOrderId(listClientOrderID).
		Timestamp(ts).
		RecvWindow(5000)
	resp, _, err := req.Execute()
	if err != nil {
		return nil, decodeSpotError(err)
	}
	orderList := &spotPlacedOrderList{symbol: symbol, listClientOrderID: listClientOrderID}
	if resp != nil {
		if resp.HasOrderListId() {
			orderList.orderListID = resp.GetOrderListId()
		}
		if resp.HasListClientOrderId() && resp.GetListClientOrderId() != "" {
			orderList.listClientOrderID = resp.GetListClientOrderId()
		}
	}
	return orderList, nil
}

func placeSpotMarketOrder(rc *restspot.APIClient, symbol, side, quantity, quoteQty string) (*spotPlacedOrder, error) {
	if strings.TrimSpace(quantity) == "" && strings.TrimSpace(quoteQty) == "" {
		return nil, fmt.Errorf("market order requires quantity or quoteQty")
	}
	if rc == nil {
		rc = newRESTClient()
	}
	ctx, err := restAuthContext()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ts := time.Now().UnixMilli()
	clientOrderID := fmt.Sprintf("wstest-mkt-%d", ts)
	req := rc.SpotTradingAPI.CreateOrderV3(ctx).
		Symbol(symbol).
		Side(side).
		Type_("MARKET").
		NewClientOrderId(clientOrderID).
		Timestamp(ts).
		RecvWindow(5000)
	if strings.TrimSpace(quantity) != "" {
		req = req.Quantity(quantity)
	}
	if strings.TrimSpace(quoteQty) != "" {
		req = req.QuoteOrderQty(quoteQty)
	}
	resp, _, err := req.Execute()
	if err != nil {
		return nil, decodeSpotError(err)
	}
	order := &spotPlacedOrder{symbol: symbol, clientOrderID: clientOrderID}
	if resp != nil {
		if resp.HasOrderId() {
			order.orderID = resp.GetOrderId()
		}
		if resp.HasClientOrderId() && resp.GetClientOrderId() != "" {
			order.clientOrderID = resp.GetClientOrderId()
		}
	}
	return order, nil
}

func cancelSpotOrder(rc *restspot.APIClient, order *spotPlacedOrder) error {
	if order == nil {
		return nil
	}
	if rc == nil {
		rc = newRESTClient()
	}
	ctx, err := restAuthContext()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ts := time.Now().UnixMilli()
	req := rc.SpotTradingAPI.DeleteOrderV3(ctx).
		Symbol(order.symbol).
		Timestamp(ts).
		RecvWindow(5000)
	if order.orderID != 0 {
		req = req.OrderId(order.orderID)
	} else if order.clientOrderID != "" {
		req = req.OrigClientOrderId(order.clientOrderID)
	}
	_, _, err = req.Execute()
	if err != nil {
		return decodeSpotError(err)
	}
	return nil
}

func cancelSpotOrderList(rc *restspot.APIClient, orderList *spotPlacedOrderList) error {
	if orderList == nil {
		return nil
	}
	if rc == nil {
		rc = newRESTClient()
	}
	ctx, err := restAuthContext()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ts := time.Now().UnixMilli()
	req := rc.SpotTradingAPI.DeleteOrderListV3(ctx).
		Symbol(orderList.symbol).
		Timestamp(ts).
		RecvWindow(5000)
	if orderList.orderListID != 0 {
		req = req.OrderListId(orderList.orderListID)
	} else if orderList.listClientOrderID != "" {
		req = req.ListClientOrderId(orderList.listClientOrderID)
	}
	_, _, err = req.Execute()
	if err != nil {
		return decodeSpotError(err)
	}
	return nil
}

func decodeSpotError(err error) error {
	if err == nil {
		return nil
	}
	if ge, ok := err.(*restspot.GenericOpenAPIError); ok {
		body := ge.Body()
		if len(body) > 0 {
			var em struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if e2 := json.Unmarshal(body, &em); e2 == nil && (em.Code != 0 || em.Msg != "") {
				return fmt.Errorf("status=%s code=%d msg=%s body=%s", ge.Error(), em.Code, em.Msg, string(body))
			}
			return fmt.Errorf("status=%s body=%s", ge.Error(), string(body))
		}
		return fmt.Errorf("status=%s", ge.Error())
	}
	return err
}

func parseFloat64(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func decimals(s string) int {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return len(s) - i - 1
	}
	return 0
}

func isInsufficientSpotBalanceError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "code=-2010") ||
		strings.Contains(msg, "code=-2011") ||
		strings.Contains(msg, "insufficient balance")
}

func isInvalidSpotKeyOrPermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "code=-2015") ||
		strings.Contains(msg, "invalid api-key") ||
		strings.Contains(msg, "permissions for action")
}

func floorToStep(val float64, step float64, decs int) float64 {
	if step <= 0 {
		scale := math.Pow10(decs)
		return math.Floor(val*scale) / scale
	}
	scale := math.Pow10(decs)
	valUnits := math.Floor(val*scale + 1e-9)
	stepUnits := math.Round(step * scale)
	if stepUnits <= 0 {
		stepUnits = 1
	}
	q := math.Floor(valUnits/stepUnits) * stepUnits
	return q / scale
}

func ceilToStep(val float64, step float64, decs int) float64 {
	if step <= 0 {
		scale := math.Pow10(decs)
		return math.Ceil(val*scale) / scale
	}
	scale := math.Pow10(decs)
	valUnits := math.Ceil(val*scale - 1e-9)
	stepUnits := math.Round(step * scale)
	if stepUnits <= 0 {
		stepUnits = 1
	}
	q := math.Ceil(valUnits/stepUnits) * stepUnits
	return q / scale
}

func deriveAmendedQuantity(origQty string, price float64, constraints *spotSymbolConstraints) (string, error) {
	if constraints == nil {
		return "", fmt.Errorf("constraints not available")
	}
	origVal := parseFloat64(origQty)
	if origVal <= 0 {
		return "", fmt.Errorf("invalid original quantity: %s", origQty)
	}
	step := constraints.qtyStep
	if step <= 0 {
		step = 1.0 / math.Pow10(constraints.qtyDecimals)
	}
	var candidate float64
	for i := 1; i <= 5; i++ {
		next := floorToStep(origVal-step*float64(i), step, constraints.qtyDecimals)
		if next <= 0 || almostEqual(next, origVal) {
			continue
		}
		if constraints.minQty > 0 && next < constraints.minQty {
			continue
		}
		if constraints.minNotional > 0 && price > 0 && next*price < constraints.minNotional {
			continue
		}
		candidate = next
		break
	}
	if candidate <= 0 || almostEqual(candidate, origVal) {
		return "", fmt.Errorf("unable to derive amended quantity from %s", origQty)
	}
	return fmt.Sprintf("%.*f", constraints.qtyDecimals, candidate), nil
}

func almostEqual(a, b float64) bool {
	const epsilon = 1e-12
	diff := math.Abs(a - b)
	return diff <= epsilon || diff <= epsilon*math.Max(math.Abs(a), math.Abs(b))
}
