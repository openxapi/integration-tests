package streamstest

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	restoptions "github.com/openxapi/binance-go/rest/options"
	"strings"
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

func newRESTClient() *restoptions.APIClient {
	cfg := restoptions.NewConfiguration()
	if s := os.Getenv("BINANCE_OPTIONS_REST_SERVER"); s != "" {
		cfg.Servers[0].URL = s
	}
	return restoptions.NewAPIClient(cfg)
}

// restPickUnderlying returns an active underlying asset (upper and lower case forms)
func restPickUnderlying(ctx context.Context) (string, string, error) {
	// Try to honor PREFERRED_UNDERLYING if it exists and is active
	preferred := os.Getenv("PREFERRED_UNDERLYING")
	if preferred != "" {
		parts := strings.Split(preferred, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			// Verify activity by checking at least one symbol exists for this underlying
			syms, err := getActiveOptionsSymbols(strings.ToUpper(p))
			if err == nil && len(syms) > 0 {
				up := strings.ToUpper(p)
				return up, strings.ToLower(up), nil
			}
		}
	}

	// Fallback: pull all active symbols and derive underlyings
	syms, err := getActiveOptionsSymbols("")
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch active option symbols: %w", err)
	}
	if len(syms) == 0 {
		return "", "", fmt.Errorf("no active option symbols found")
	}
	seen := map[string]bool{}
	var order []string
	for _, s := range syms {
		if i := strings.Index(s, "-"); i > 0 {
			u := s[:i]
			if !seen[u] {
				seen[u] = true
				order = append(order, u)
			}
		}
	}
	if len(order) == 0 {
		return "", "", fmt.Errorf("no underlyings derived from active symbols")
	}
	up := order[0]
	return up, strings.ToLower(up), nil
}

// restPickTwoUnderlyings returns two distinct active underlyings (upper and lower forms)
func restPickTwoUnderlyings(ctx context.Context) (string, string, string, string, error) {
	// Try preferred list first, honoring order
	if pref := os.Getenv("PREFERRED_UNDERLYING"); pref != "" {
		parts := strings.Split(pref, ",")
		active := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			syms, _ := getActiveOptionsSymbols(strings.ToUpper(p))
			if len(syms) > 0 {
				active = append(active, strings.ToUpper(p))
			}
		}
		if len(active) > 0 {
			upA := active[0]
			lowA := strings.ToLower(upA)
			upB := ""
			lowB := ""
			if len(active) > 1 {
				upB = active[1]
				lowB = strings.ToLower(upB)
			}
			return upA, lowA, upB, lowB, nil
		}
	}

	syms, err := getActiveOptionsSymbols("")
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to fetch active option symbols: %w", err)
	}
	if len(syms) == 0 {
		return "", "", "", "", fmt.Errorf("no active option symbols found")
	}
	seen := map[string]bool{}
	order := make([]string, 0, 4)
	for _, s := range syms {
		if i := strings.Index(s, "-"); i > 0 {
			u := s[:i]
			if !seen[u] {
				seen[u] = true
				order = append(order, u)
			}
		}
	}
	if len(order) == 0 {
		return "", "", "", "", fmt.Errorf("no underlyings derived from active symbols")
	}
	// Prefer BTC then ETH if present, then fall back to first two
	pick := func(want string) (string, bool) {
		want = strings.ToUpper(want)
		for _, u := range order {
			if u == want {
				return u, true
			}
		}
		return "", false
	}
	upA := ""
	upB := ""
	if u, ok := pick("BTC"); ok {
		upA = u
	}
	if upA == "" {
		upA = order[0]
	}
	// pick second distinct
	for _, u := range order {
		if u != upA {
			upB = u
			break
		}
	}
	return upA, strings.ToLower(upA), upB, strings.ToLower(upB), nil
}

// restIndexPrice fetches the current underlying index price via REST
func restIndexPrice(ctx context.Context, underlying string) (float64, error) {
	// Reuse existing helper; it already respects BINANCE_OPTIONS_REST_SERVER
	return getCurrentUnderlyingPrice(underlying)
}

// restMarkPrice fetches the option mark price via REST
func restMarkPrice(ctx context.Context, symbol string) (float64, error) {
	rc := newRESTClient()
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, _, err := rc.OptionsAPI.GetMarkV1(cctx).Symbol(symbol).Execute()
	if err != nil || len(resp) == 0 || resp[0].MarkPrice == nil {
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	return strconv.ParseFloat(*resp[0].MarkPrice, 64)
}

// restTickerLast fetches the option last price via REST
func restTickerLast(ctx context.Context, symbol string) (float64, error) {
	rc := newRESTClient()
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, _, err := rc.OptionsAPI.GetTickerV1(cctx).Symbol(symbol).Execute()
	if err != nil || len(resp) == 0 || resp[0].LastPrice == nil {
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	return strconv.ParseFloat(*resp[0].LastPrice, 64)
}
