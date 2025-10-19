package streamstest

import (
    "context"
    "fmt"
    "os"
    "strings"
    "time"
    "strconv"

    restum "github.com/openxapi/binance-go/rest/umfutures"
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

func newRESTClient() *restum.APIClient {
    cfg := restum.NewConfiguration()
    if s := os.Getenv("BINANCE_UMFUTURES_REST_SERVER"); s != "" {
        cfg.Servers[0].URL = s
    }
    return restum.NewAPIClient(cfg)
}

// restPickSymbol returns an active symbol; prefer PREFERRED_SYMBOL env (comma-separated)
func restPickSymbol(ctx context.Context) (string, error) {
    // honor preferred if active
    if pref := os.Getenv("PREFERRED_SYMBOL"); pref != "" {
        parts := strings.Split(pref, ",")
        for _, p := range parts {
            s := strings.TrimSpace(p)
            if s == "" { continue }
            if ok := symbolExists(ctx, strings.ToUpper(s)); ok {
                return strings.ToUpper(s), nil
            }
        }
    }
    // fallback: first active from exchange info
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
    defer cancel()
    info, _, err := rc.FuturesAPI.GetExchangeInfoV1(cctx).Execute()
    if err != nil || info == nil || info.Symbols == nil || len(info.Symbols) == 0 {
        return "BTCUSDT", nil // sensible fallback
    }
    // prefer BTCUSDT if present
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

func restTickerPrice(ctx context.Context, symbol string) (float64, error) {
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    resp, _, err := rc.FuturesAPI.GetTickerPriceV2(cctx).Symbol(symbol).Execute()
    if err != nil || resp == nil {
        if err != nil { return 0, err }
        return 0, fmt.Errorf("no price")
    }
    // Handle oneOf: either a single item or an array of items
    if v := resp.GetActualInstance(); v != nil {
        switch x := v.(type) {
        case *restum.UmfuturesGetTickerPriceV2RespItem:
            if x.Price == nil { return 0, fmt.Errorf("no price") }
            return parseFloat(*x.Price)
        case *[]restum.UmfuturesGetTickerPriceV2RespItem:
            // try to find matching symbol; otherwise first
            if len(*x) == 0 { return 0, fmt.Errorf("no price items") }
            for _, it := range *x {
                if it.Symbol != nil && strings.EqualFold(*it.Symbol, symbol) && it.Price != nil {
                    return parseFloat(*it.Price)
                }
            }
            if (*x)[0].Price == nil { return 0, fmt.Errorf("no price in first item") }
            return parseFloat(*(*x)[0].Price)
        }
    }
    return 0, fmt.Errorf("unexpected response type")
}

func restMarkPrice(ctx context.Context, symbol string) (float64, error) {
    rc := newRESTClient()
    cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    resp, _, err := rc.FuturesAPI.GetPremiumIndexV1(cctx).Symbol(symbol).Execute()
    if err != nil || resp == nil {
        if err != nil { return 0, err }
        return 0, fmt.Errorf("no mark resp")
    }
    // Handle oneOf container
    if v := resp.GetActualInstance(); v != nil {
        switch x := v.(type) {
        case *restum.UmfuturesGetPremiumIndexV1RespItem:
            if x.MarkPrice == nil { return 0, fmt.Errorf("no markPrice") }
            return parseFloat(*x.MarkPrice)
        case *[]restum.UmfuturesGetPremiumIndexV1RespItem:
            if len(*x) == 0 { return 0, fmt.Errorf("no mark items") }
            for _, it := range *x {
                if it.Symbol != nil && strings.EqualFold(*it.Symbol, symbol) && it.MarkPrice != nil {
                    return parseFloat(*it.MarkPrice)
                }
            }
            if (*x)[0].MarkPrice == nil { return 0, fmt.Errorf("no markPrice in first item") }
            return parseFloat(*(*x)[0].MarkPrice)
        }
    }
    return 0, fmt.Errorf("unexpected response type")
}

func parseFloat(s string) (float64, error) {
    return strconvParseFloat(s, 64)
}

// alias to avoid extra imports inside patch chunk
var strconvParseFloat = func(s string, bitSize int) (float64, error) { return strconv.ParseFloat(s, bitSize) }
