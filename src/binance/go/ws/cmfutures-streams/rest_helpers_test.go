package streamstest

import (
    "context"
    "fmt"
    "os"
    "strings"
    "time"

    restcm "github.com/openxapi/binance-go/rest/cmfutures"
    "strconv"
)

func restValidationEnabled() bool {
    v := os.Getenv("ENABLE_REST_VALIDATION")
    if v == "" { return false }
    switch strings.ToLower(v) {
    case "1", "true", "yes", "on":
        return true
    default:
        return false
    }
}

func newRESTClient() *restcm.APIClient {
    cfg := restcm.NewConfiguration()
    if s := os.Getenv("BINANCE_CMFUTURES_REST_SERVER"); s != "" {
        cfg.Servers[0].URL = s
    } else {
        // Default to testnet host
        cfg.Servers[0].URL = "https://testnet.binancefuture.com"
    }
    return restcm.NewAPIClient(cfg)
}

// restPickSymbol returns an active COIN-M perpetual symbol (upper-case), prefer env PREFERRED_SYMBOL
func restPickSymbol(ctx context.Context) (string, error) {
    if pref := os.Getenv("PREFERRED_SYMBOL"); pref != "" {
        parts := strings.Split(pref, ",")
        for _, p := range parts {
            s := strings.TrimSpace(p)
            if s == "" { continue }
            if symbolExists(ctx, strings.ToUpper(s)) { return strings.ToUpper(s), nil }
        }
    }
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
    defer cancel()
    info, _, err := rc.FuturesAPI.GetExchangeInfoV1(cctx).Execute()
    if err != nil || info == nil || info.Symbols == nil || len(info.Symbols) == 0 {
        return "BTCUSD_PERP", nil
    }
    // Prefer BTCUSD_PERP if present
    for _, s := range info.Symbols {
        if s.Symbol != nil && strings.EqualFold(*s.Symbol, "BTCUSD_PERP") {
            return "BTCUSD_PERP", nil
        }
    }
    if info.Symbols[0].Symbol != nil {
        return strings.ToUpper(*info.Symbols[0].Symbol), nil
    }
    return "BTCUSD_PERP", nil
}

func symbolExists(ctx context.Context, sym string) bool {
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    info, _, err := rc.FuturesAPI.GetExchangeInfoV1(cctx).Execute()
    if err != nil || info == nil || info.Symbols == nil { return false }
    for _, s := range info.Symbols {
        if s.Symbol != nil && strings.EqualFold(*s.Symbol, sym) { return true }
    }
    return false
}

func restMarkPrice(ctx context.Context, symbol string) (float64, error) {
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    resp, _, err := rc.FuturesAPI.GetPremiumIndexV1(cctx).Symbol(symbol).Execute()
    if err != nil {
        return 0, err
    }
    // For COIN-M, the SDK returns a slice []GetPremiumIndexV1RespItem
    if len(resp) == 0 {
        return 0, fmt.Errorf("no items")
    }
    // Try to find matching symbol first
    for _, it := range resp {
        if it.Symbol != nil && strings.EqualFold(*it.Symbol, symbol) && it.MarkPrice != nil {
            return strconvParseFloat(*it.MarkPrice, 64)
        }
    }
    // Fallback to first item's markPrice
    if resp[0].MarkPrice == nil {
        return 0, fmt.Errorf("no markPrice in first item")
    }
    return strconvParseFloat(*resp[0].MarkPrice, 64)
}

// alias used by rest_helpers for parsing floats
var strconvParseFloat = func(s string, bitSize int) (float64, error) { return strconv.ParseFloat(s, bitSize) }
