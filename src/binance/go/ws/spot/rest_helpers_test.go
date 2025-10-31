package wstest

import (
	"context"
	"os"
	"strings"
	"time"

	restspot "github.com/openxapi/binance-go/rest/spot"
)

func newRESTClient() *restspot.APIClient {
	cfg := restspot.NewConfiguration()
	if s := strings.TrimSpace(os.Getenv("BINANCE_SPOT_REST_SERVER")); s != "" {
		cfg.Servers[0].URL = s
	}
	return restspot.NewAPIClient(cfg)
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
