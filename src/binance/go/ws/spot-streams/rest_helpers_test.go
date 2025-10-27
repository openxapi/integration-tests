package streamstest

import (
	"context"
	"os"
	"strings"
	"time"

	restspot "github.com/openxapi/binance-go/rest/spot"
)

func restValidationEnabled() bool {
	v := os.Getenv("ENABLE_REST_VALIDATION")
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func newRESTClient() *restspot.APIClient {
	cfg := restspot.NewConfiguration()
	if s := os.Getenv("BINANCE_SPOT_REST_SERVER"); s != "" {
		cfg.Servers[0].URL = s
	}
	return restspot.NewAPIClient(cfg)
}

// restPickSymbol returns an active symbol; prefer PREFERRED_SYMBOL env (comma-separated)
func restPickSymbol(ctx context.Context) (string, error) {
	if pref := os.Getenv("PREFERRED_SYMBOL"); pref != "" {
		parts := strings.Split(pref, ",")
		for _, p := range parts {
			s := strings.TrimSpace(p)
			if s == "" {
				continue
			}
			if ok := symbolExists(ctx, strings.ToUpper(s)); ok {
				return strings.ToUpper(s), nil
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

func symbolExists(ctx context.Context, sym string) bool {
	rc := newRESTClient()
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	info, _, err := rc.SpotTradingAPI.GetExchangeInfoV3(cctx).Execute()
	if err != nil || info == nil || info.Symbols == nil {
		return false
	}
	for _, s := range info.Symbols {
		if s.Symbol != nil && strings.EqualFold(*s.Symbol, sym) {
			return true
		}
	}
	return false
}
