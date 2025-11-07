package streamstest

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	pmarginstreams "github.com/openxapi/binance-go/ws/pmargin-streams"
	"github.com/openxapi/binance-go/ws/pmargin-streams/models"
)

// TestFullIntegrationSuite_UserData exercises connection, request/response, and event handlers
// for the Portfolio Margin user data stream channel.
func TestFullIntegrationSuite_UserData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping portfolio margin WS tests in short mode")
	}

	// Capture SDK log output and fail on any "unhandled message" during the run
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
	if apiKey == "" || secret == "" {
		t.Skip("BINANCE_API_KEY/SECRET_KEY not set; skipping portfolio margin user data tests")
	}

	restClient := newPMarginRESTClient()
	baseCtx := context.Background()
	authCtx, _, err := restAuthContext(baseCtx, apiKey, secret)
	if err != nil {
		t.Fatalf("failed to construct auth context: %v", err)
	}

	handle, err := restCreateListenKey(authCtx, restClient)
	if err != nil {
		t.Skipf("failed to create portfolio margin listen key: %v", err)
		return
	}
	defer func() {
		if handle != nil {
			if err := handle.Close(authCtx); err != nil {
				t.Logf("listen key close error: %v", err)
			}
		}
	}()
	if handle.ListenKey == "" {
		t.Skip("listen key response empty; skipping user data stream tests")
	}

	stc, err := NewStreamTestClientDedicated(getTestConfig())
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	client := stc.Client()

	if as := client.GetActiveServer(); as != nil {
		t.Logf("active pmargin WS server: name=%s url=%s", as.Name, as.URL)
	} else {
		t.Logf("active pmargin WS server: <nil>")
	}

	ud := pmarginstreams.NewUserDataStreamChannel(client)
	var connected bool

	if !t.Run("Connect", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := ud.Connect(ctx, handle.ListenKey); err != nil {
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

	runUserDataEventTests(t, ud)

	t.Run("ListenKeyKeepAlive", func(t *testing.T) {
		if handle == nil {
			t.Skip("no listen key handle available")
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

		// Disconnect again to ensure idempotency
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		if err := ud.Disconnect(ctx2); err != nil {
			t.Fatalf("second disconnect failed: %v", err)
		}
		connected = false
	})
}

func runUserDataEventTests(t *testing.T, ud *pmarginstreams.UserDataStreamChannel) {
	t.Run("ConditionalOrderTradeUpdateEvent", func(t *testing.T) {
		events := make(chan *models.ConditionalOrderTradeUpdateEvent, 1)
		ud.HandleConditionalOrderTradeUpdateEvent(func(ctx context.Context, ev *models.ConditionalOrderTradeUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "CONDITIONAL_ORDER_TRADE_UPDATE" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.StrategyOrderDetails.Quantity != "" {
				_ = tryParseFloat(t, ev.StrategyOrderDetails.Quantity, "strategy.quantity")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterConditionalOrderTradeUpdateEvent() })
		if ev, ok := waitForEvent(t, "conditional order trade update", events); ok {
			logJSON(t, "event.conditionalOrderTradeUpdate", ev)
		}
	})

	t.Run("OpenOrderLossEvent", func(t *testing.T) {
		events := make(chan *models.OpenOrderLossEvent, 1)
		ud.HandleOpenOrderLossEvent(func(ctx context.Context, ev *models.OpenOrderLossEvent) error {
			if ev.EventType != "" && ev.EventType != "openOrderLoss" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterOpenOrderLossEvent() })
		if ev, ok := waitForEvent(t, "open order loss", events); ok {
			for _, loss := range ev.ArrayOfLossUpdates {
				if loss.Asset != "" {
					assertNonEmpty(t, loss.Asset, "loss.asset")
				}
				if loss.LossAmount != "" {
					_ = tryParseFloat(t, loss.LossAmount, "loss.amount")
				}
			}
			logJSON(t, "event.openOrderLoss", ev)
		}
	})

	t.Run("MarginAccountUpdateEvent", func(t *testing.T) {
		events := make(chan *models.MarginAccountUpdateEvent, 1)
		ud.HandleMarginAccountUpdateEvent(func(ctx context.Context, ev *models.MarginAccountUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "outboundAccountPosition" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			for _, bal := range ev.BalancesArray {
				if bal.Asset != "" {
					assertNonEmpty(t, bal.Asset, "balance.asset")
				}
				if bal.FreeBalance != "" {
					_ = tryParseFloat(t, bal.FreeBalance, "balance.free")
				}
				if bal.LockedBalance != "" {
					_ = tryParseFloat(t, bal.LockedBalance, "balance.locked")
				}
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterMarginAccountUpdateEvent() })
		if ev, ok := waitForEvent(t, "margin account update", events); ok {
			logJSON(t, "event.marginAccountUpdate", ev)
		}
	})

	t.Run("LiabilityUpdateEvent", func(t *testing.T) {
		events := make(chan *models.LiabilityUpdateEvent, 1)
		ud.HandleLiabilityUpdateEvent(func(ctx context.Context, ev *models.LiabilityUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "liabilityChange" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.Asset != "" {
				assertNonEmpty(t, ev.Asset, "asset")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterLiabilityUpdateEvent() })
		if ev, ok := waitForEvent(t, "liability update", events); ok {
			logJSON(t, "event.liabilityUpdate", ev)
		}
	})

	t.Run("MarginOrderUpdateEvent", func(t *testing.T) {
		events := make(chan *models.MarginOrderUpdateEvent, 1)
		ud.HandleMarginOrderUpdateEvent(func(ctx context.Context, ev *models.MarginOrderUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "executionReport" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.OrderPrice != "" {
				_ = tryParseFloat(t, ev.OrderPrice, "orderPrice")
			}
			if ev.OrderQuantity != "" {
				_ = tryParseFloat(t, ev.OrderQuantity, "orderQty")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterMarginOrderUpdateEvent() })
		if ev, ok := waitForEvent(t, "margin order update", events); ok {
			logJSON(t, "event.marginOrderUpdate", ev)
		}
	})

	t.Run("FuturesOrderUpdateEvent", func(t *testing.T) {
		events := make(chan *models.FuturesOrderUpdateEvent, 1)
		ud.HandleFuturesOrderUpdateEvent(func(ctx context.Context, ev *models.FuturesOrderUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "ORDER_TRADE_UPDATE" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.OrderDetails.Symbol != "" {
				assertNonEmpty(t, ev.OrderDetails.Symbol, "order.symbol")
			}
			if ev.OrderDetails.OriginalPrice != "" {
				_ = tryParseFloat(t, ev.OrderDetails.OriginalPrice, "order.originalPrice")
			}
			if ev.OrderDetails.AveragePrice != "" {
				_ = tryParseFloat(t, ev.OrderDetails.AveragePrice, "order.averagePrice")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterFuturesOrderUpdateEvent() })
		if ev, ok := waitForEvent(t, "futures order update", events); ok {
			logJSON(t, "event.futuresOrderUpdate", ev)
		}
	})

	t.Run("FuturesBalancePositionUpdateEvent", func(t *testing.T) {
		events := make(chan *models.FuturesBalancePositionUpdateEvent, 1)
		ud.HandleFuturesBalancePositionUpdateEvent(func(ctx context.Context, ev *models.FuturesBalancePositionUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "ACCOUNT_UPDATE" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if det := ev.AccountUpdateDetails; det.EventReasonType != "" {
				assertNonEmpty(t, det.EventReasonType, "accountUpdate.reason")
			}
			for _, bal := range ev.AccountUpdateDetails.Balances {
				if bal.Asset != "" {
					assertNonEmpty(t, bal.Asset, "balance.asset")
				}
				if bal.WalletBalance != "" {
					_ = tryParseFloat(t, bal.WalletBalance, "balance.wallet")
				}
				if bal.CrossWalletBalance != "" {
					_ = tryParseFloat(t, bal.CrossWalletBalance, "balance.crossWallet")
				}
				if bal.BalanceChange != "" {
					_ = tryParseFloat(t, bal.BalanceChange, "balance.change")
				}
			}
			for _, pos := range ev.AccountUpdateDetails.Positions {
				if pos.Symbol != "" {
					assertNonEmpty(t, pos.Symbol, "position.symbol")
				}
				if pos.PositionAmount != "" {
					_ = tryParseFloat(t, pos.PositionAmount, "position.amount")
				}
				if pos.EntryPrice != "" {
					_ = tryParseFloat(t, pos.EntryPrice, "position.entryPrice")
				}
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterFuturesBalancePositionUpdateEvent() })
		if ev, ok := waitForEvent(t, "futures balance/position update", events); ok {
			logJSON(t, "event.futuresBalancePositionUpdate", ev)
		}
	})

	t.Run("FuturesAccountConfigUpdateEvent", func(t *testing.T) {
		events := make(chan *models.FuturesAccountConfigUpdateEvent, 1)
		ud.HandleFuturesAccountConfigUpdateEvent(func(ctx context.Context, ev *models.FuturesAccountConfigUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "ACCOUNT_CONFIG_UPDATE" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterFuturesAccountConfigUpdateEvent() })
		if ev, ok := waitForEvent(t, "futures account config update", events); ok {
			logJSON(t, "event.futuresAccountConfigUpdate", ev)
		}
	})

	t.Run("RiskLevelChangeEvent", func(t *testing.T) {
		events := make(chan *models.RiskLevelChangeEvent, 1)
		ud.HandleRiskLevelChangeEvent(func(ctx context.Context, ev *models.RiskLevelChangeEvent) error {
			if ev.EventType != "" && ev.EventType != "riskLevelChange" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.RiskLevelStatus != "" {
				assertNonEmpty(t, ev.RiskLevelStatus, "riskLevel.status")
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

	t.Run("MarginBalanceUpdateEvent", func(t *testing.T) {
		events := make(chan *models.MarginBalanceUpdateEvent, 1)
		ud.HandleMarginBalanceUpdateEvent(func(ctx context.Context, ev *models.MarginBalanceUpdateEvent) error {
			if ev.EventType != "" && ev.EventType != "balanceUpdate" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			if ev.BalanceDelta != "" {
				_ = tryParseFloat(t, ev.BalanceDelta, "balanceDelta")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterMarginBalanceUpdateEvent() })
		if ev, ok := waitForEvent(t, "margin balance update", events); ok {
			logJSON(t, "event.marginBalanceUpdate", ev)
		}
	})

	t.Run("UserDataStreamExpiredEvent", func(t *testing.T) {
		events := make(chan *models.UserDataStreamExpiredEvent, 1)
		ud.HandleUserDataStreamExpiredEvent(func(ctx context.Context, ev *models.UserDataStreamExpiredEvent) error {
			if ev.EventType != "" && ev.EventType != "listenKeyExpired" {
				t.Errorf("unexpected event type: %s", ev.EventType)
			}
			if ev.EventTime > 0 {
				assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
			}
			select {
			case events <- ev:
			default:
			}
			return nil
		})
		t.Cleanup(func() { ud.UnregisterUserDataStreamExpiredEvent() })
		if ev, ok := waitForEvent(t, "listen key expired", events); ok {
			logJSON(t, "event.listenKeyExpired", ev)
		}
	})
}
