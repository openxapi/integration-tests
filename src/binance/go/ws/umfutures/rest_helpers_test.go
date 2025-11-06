package wstest

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	restum "github.com/openxapi/binance-go/rest/umfutures"
)

const defaultUmFuturesRestTestnetServer = "https://testnet.binancefuture.com"

func newUmFuturesRESTClient() *restum.APIClient {
	cfg := restum.NewConfiguration()
	if s := strings.TrimSpace(os.Getenv("BINANCE_UMFUTURES_REST_SERVER")); s != "" {
		cfg.Servers[0].URL = s
	} else if len(cfg.Servers) > 0 && strings.TrimSpace(cfg.Servers[0].URL) != "" {
		cfg.Servers[0].URL = defaultUmFuturesRestTestnetServer
	} else {
		cfg.Servers = []restum.ServerConfiguration{{
			URL:         defaultUmFuturesRestTestnetServer,
			Description: "Binance Futures Testnet",
		}}
	}
	return restum.NewAPIClient(cfg)
}

type futuresSymbolConstraints struct {
	priceStep     float64
	qtyStep       float64
	minQty        float64
	minNotional   float64
	priceDecimals int
	qtyDecimals   int
}

func loadFuturesSymbolConstraints(ctx context.Context, rc *restum.APIClient, symbol string) (*futuresSymbolConstraints, error) {
	if rc == nil {
		rc = newUmFuturesRESTClient()
	}
	base := ctx
	if base == nil {
		base = context.Background()
	}
	cctx, cancel := context.WithTimeout(base, 8*time.Second)
	defer cancel()
	resp, _, err := rc.FuturesAPI.GetExchangeInfoV1(cctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("exchangeInfo error: %w", err)
	}
	if resp == nil || len(resp.Symbols) == 0 {
		return nil, fmt.Errorf("exchangeInfo returned no symbols")
	}
	var entry *restum.UmfuturesGetExchangeInfoV1RespSymbolsInner
	symbol = strings.ToUpper(symbol)
	for i := range resp.Symbols {
		if strings.EqualFold(resp.Symbols[i].GetSymbol(), symbol) {
			entry = &resp.Symbols[i]
			break
		}
	}
	if entry == nil {
		return nil, fmt.Errorf("symbol %s not found in exchangeInfo", symbol)
	}
	constraints := &futuresSymbolConstraints{
		priceDecimals: int(entry.GetPricePrecision()),
		qtyDecimals:   int(entry.GetQuantityPrecision()),
		minNotional:   5.0,
	}
	for _, f := range entry.Filters {
		if strings.EqualFold(f.GetFilterType(), "PRICE_FILTER") {
			if ts := f.GetTickSize(); ts != "" {
				constraints.priceStep = parseFloat64(ts)
				if dec := decimals(ts); dec > 0 {
					constraints.priceDecimals = dec
				}
			}
		}
	}
	if constraints.priceStep <= 0 && constraints.priceDecimals > 0 {
		constraints.priceStep = 1.0 / math.Pow10(constraints.priceDecimals)
	}
	if constraints.qtyStep <= 0 && constraints.qtyDecimals > 0 {
		constraints.qtyStep = 1.0 / math.Pow10(constraints.qtyDecimals)
	}
	if constraints.qtyStep <= 0 {
		constraints.qtyStep = 0.001
	}
	if constraints.qtyDecimals <= 0 {
		constraints.qtyDecimals = 8
	}
	if constraints.priceDecimals <= 0 {
		constraints.priceDecimals = 8
	}
	if constraints.minQty <= 0 {
		constraints.minQty = constraints.qtyStep
	}
	return constraints, nil
}

func restFuturesTickerPrice(ctx context.Context, rc *restum.APIClient, symbol string) (float64, error) {
	if rc == nil {
		rc = newUmFuturesRESTClient()
	}
	base := ctx
	if base == nil {
		base = context.Background()
	}
	cctx, cancel := context.WithTimeout(base, 5*time.Second)
	defer cancel()
	resp, _, err := rc.FuturesAPI.GetTickerPriceV1(cctx).Symbol(symbol).Execute()
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, fmt.Errorf("ticker price response empty")
	}
	if item := resp.UmfuturesGetTickerPriceV1RespItem; item != nil {
		price := parseFloat64(item.GetPrice())
		if price <= 0 {
			return 0, fmt.Errorf("ticker price invalid for %s", symbol)
		}
		return price, nil
	}
	if arr := resp.ArrayOfUmfuturesGetTickerPriceV1RespItem; arr != nil {
		items := *arr
		if len(items) == 0 {
			return 0, fmt.Errorf("ticker price array empty")
		}
		for _, it := range items {
			if strings.EqualFold(it.GetSymbol(), symbol) {
				price := parseFloat64(it.GetPrice())
				if price > 0 {
					return price, nil
				}
			}
		}
		price := parseFloat64(items[0].GetPrice())
		if price > 0 {
			return price, nil
		}
	}
	return 0, fmt.Errorf("ticker price not available for %s", symbol)
}

type futuresOrderParams struct {
	Quantity     string
	InitialPrice string
	ModifyPrice  string
}

func prepareFuturesLimitOrderParams(ctx context.Context, rc *restum.APIClient, symbol string, side string) (*futuresOrderParams, error) {
	constraints, err := loadFuturesSymbolConstraints(ctx, rc, symbol)
	if err != nil {
		return nil, err
	}
	priceRef, err := restFuturesTickerPrice(ctx, rc, symbol)
	if err != nil {
		return nil, fmt.Errorf("ticker price fetch failed: %w", err)
	}
	if priceRef <= 0 {
		return nil, fmt.Errorf("ticker price invalid for %s", symbol)
	}

	side = strings.ToUpper(side)
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("unsupported side %q", side)
	}

	var entryMultiplier, modifyMultiplier float64
	if side == "BUY" {
		entryMultiplier = 0.97
		modifyMultiplier = 0.98
	} else {
		entryMultiplier = 1.03
		modifyMultiplier = 1.02
	}

	initialPrice := floorToStep(priceRef*entryMultiplier, constraints.priceStep, constraints.priceDecimals)
	if side == "SELL" {
		initialPrice = ceilToStep(priceRef*entryMultiplier, constraints.priceStep, constraints.priceDecimals)
	}
	if initialPrice <= 0 {
		initialPrice = floorToStep(priceRef, constraints.priceStep, constraints.priceDecimals)
	}
	modifyPrice := floorToStep(priceRef*modifyMultiplier, constraints.priceStep, constraints.priceDecimals)
	if side == "SELL" {
		modifyPrice = ceilToStep(priceRef*modifyMultiplier, constraints.priceStep, constraints.priceDecimals)
	}
	if modifyPrice <= 0 || (side == "BUY" && modifyPrice >= initialPrice) || (side == "SELL" && modifyPrice <= initialPrice) {
		step := constraints.priceStep
		if step <= 0 {
			step = 1.0 / math.Pow10(constraints.priceDecimals)
		}
		if side == "BUY" {
			modifyPrice = initialPrice + step
		} else {
			modifyPrice = initialPrice - step
		}
	}

	minQty := constraints.minQty
	if minQty <= 0 {
		minQty = constraints.qtyStep
	}
	if minQty <= 0 {
		minQty = 1.0 / math.Pow10(constraints.qtyDecimals)
	}
	qty := ceilToStep(minQty, constraints.qtyStep, constraints.qtyDecimals)
	if qty <= 0 {
		qty = 1.0 / math.Pow10(constraints.qtyDecimals)
	}
	targetNotional := qty * initialPrice
	if constraints.minNotional > 0 && targetNotional < constraints.minNotional {
		multiplier := math.Ceil(constraints.minNotional / targetNotional)
		qty = ceilToStep(qty*multiplier, constraints.qtyStep, constraints.qtyDecimals)
		targetNotional = qty * initialPrice
	}
	if targetNotional > 100.0 {
		qty = ceilToStep(100.0/initialPrice, constraints.qtyStep, constraints.qtyDecimals)
	}
	if qty <= 0 {
		return nil, errors.New("failed to compute valid order quantity")
	}

	return &futuresOrderParams{
		Quantity:     fmt.Sprintf("%.*f", constraints.qtyDecimals, qty),
		InitialPrice: fmt.Sprintf("%.*f", constraints.priceDecimals, initialPrice),
		ModifyPrice:  fmt.Sprintf("%.*f", constraints.priceDecimals, modifyPrice),
	}, nil
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
