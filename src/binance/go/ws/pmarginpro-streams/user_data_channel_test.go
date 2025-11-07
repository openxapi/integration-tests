package streamstest

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	pmarginpro "github.com/openxapi/binance-go/ws/pmarginpro-streams"
	"github.com/openxapi/binance-go/ws/pmarginpro-streams/models"
)

// TestFullIntegrationSuite_UserData exercises connection, request/response, and event handlers
// for the Portfolio Margin Pro user data stream channel.
func TestFullIntegrationSuite_UserData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping portfolio margin pro WS tests in short mode")
	}

	catcher := &unhandledCatcher{}
	log.SetOutput(catcher)
	defer func() {
		log.SetOutput(os.Stderr)
		catcher.mu.Lock()
		defer catcher.mu.Unlock()
		if len(catcher.matches) > 0 {
			for _, line := range catcher.matches {
				t.Logf("SDK log captured: %s", strings.TrimSpace(line))
			}
			t.Fatalf("SDK emitted %d 'unhandled message' log(s) during user data suite", len(catcher.matches))
		}
	}()

	apiKey := apiKeyFromEnv()
	secret := secretKeyFromEnv()
	listenKey := preferredListenKeyFromEnv()

	var (
		handle  *listenKeyHandle
		authCtx context.Context
	)

	if listenKey == "" {
		if apiKey == "" || secret == "" {
			t.Skip("BINANCE_API_KEY/SECRET_KEY not set and no BINANCE_LISTEN_KEY provided; skipping")
		}
		restClient := newPMarginProRESTClient()
		baseCtx := context.Background()
		var err error
		authCtx, _, err = restAuthContext(baseCtx, apiKey, secret)
		if err != nil {
			t.Fatalf("failed to construct auth context: %v", err)
		}
		handle, err = restCreateListenKey(authCtx, restClient)
		if err != nil {
			t.Skipf("failed to create portfolio margin pro listen key: %v", err)
			return
		}
		listenKey = handle.ListenKey
		t.Cleanup(func() {
			if handle != nil && authCtx != nil {
				ctx, cancel := context.WithTimeout(authCtx, 5*time.Second)
				defer cancel()
				if err := handle.Close(ctx); err != nil {
					t.Logf("listen key close error: %v", err)
				}
			}
		})
	}

	if strings.TrimSpace(listenKey) == "" {
		t.Skip("listen key unavailable; skipping user data stream tests")
	}

	stc, err := NewStreamTestClientDedicated(getTestConfig())
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	client := stc.Client()

	if as := client.GetActiveServer(); as != nil {
		t.Logf("active pmargin pro WS server: name=%s url=%s", as.Name, as.URL)
	} else {
		t.Logf("active pmargin pro WS server: <nil>")
	}

	ud := pmarginpro.NewUserDataStreamChannel(client)
	var connected bool

	if !t.Run("Connect", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := ud.Connect(ctx, listenKey); err != nil {
			t.Fatalf("connect failed: %v", err)
		}
		connected = true
	}) {
		return
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := ud.Disconnect(ctx); err != nil {
			t.Logf("disconnect cleanup error: %v", err)
		}
		cancel()
	}()

	runRiskLevelChangeEventTests(t, ud)

	t.Run("ListenKeyKeepAlive", func(t *testing.T) {
		if handle == nil || authCtx == nil {
			t.Skip("no REST listen key handle available")
		}
		if err := handle.Renew(authCtx); err != nil {
			t.Logf("listen key keepalive failed (non-fatal): %v", err)
		}
	})

	t.Run("Disconnect", func(t *testing.T) {
		if !connected {
			t.Skip("channel was never connected")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ud.Disconnect(ctx); err != nil {
			t.Fatalf("disconnect failed: %v", err)
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		if err := ud.Disconnect(ctx2); err != nil {
			t.Fatalf("second disconnect failed: %v", err)
		}
		connected = false
	})
}

func runRiskLevelChangeEventTests(t *testing.T, ud *pmarginpro.UserDataStreamChannel) {
	t.Run("RiskLevelChangeEvent", func(t *testing.T) {
		events := make(chan *models.RiskLevelChangeEvent, 1)
		ud.HandleRiskLevelChangeEvent(func(ctx context.Context, ev *models.RiskLevelChangeEvent) error {
			if ev.EventType != "" && ev.EventType != "riskLevelChange" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.UniMMRLevel != "" {
				tryParseFloat(t, ev.UniMMRLevel, "uniMMRLevel")
			}
			if ev.AccountEquityInUSD != "" {
				tryParseFloat(t, ev.AccountEquityInUSD, "accountEquityInUSD")
			}
			if ev.ActualEquityWithoutCollateralRateInUSD != "" {
				tryParseFloat(t, ev.ActualEquityWithoutCollateralRateInUSD, "actualEquityWithoutCollateralRateInUSD")
			}
			if ev.TotalMaintenanceMarginInUSD != "" {
				tryParseFloat(t, ev.TotalMaintenanceMarginInUSD, "totalMaintenanceMarginInUSD")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterRiskLevelChangeEvent() })
		if ev, ok := waitForEvent(t, "risk level change", events); ok {
			logJSON(t, "event.riskLevelChange", ev)
		}
	})
}
