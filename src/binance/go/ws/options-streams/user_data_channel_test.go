package streamstest

import (
    "context"
    "log"
    "os"
    "strings"
    "testing"
    "time"

    restoptions "github.com/openxapi/binance-go/rest/options"
    optionsstreams "github.com/openxapi/binance-go/ws/options-streams"
    "github.com/openxapi/binance-go/ws/options-streams/models"
)

// TestFullIntegrationSuite_UserData runs request/response and event coverage for UserDataStreamsChannel
func TestFullIntegrationSuite_UserData(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }

    // Capture SDK log output and fail on any 'unhandled message' during this suite
    cw := &unhandledCatcher{}
    log.SetOutput(cw)
    defer func() {
        // Restore default output
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

    apiKey := os.Getenv("BINANCE_API_KEY")
    secret := os.Getenv("BINANCE_SECRET_KEY")
    if apiKey == "" || secret == "" {
        t.Skip("BINANCE_API_KEY/SECRET_KEY not set; skipping user data stream tests")
    }

	// REST client for listenKey and (optionally) order ops
	cfg := restoptions.NewConfiguration()
	if s := os.Getenv("BINANCE_OPTIONS_REST_SERVER"); s != "" {
		cfg.Servers[0].URL = s
	}
	rc := restoptions.NewAPIClient(cfg)
	auth := restoptions.NewAuth(apiKey)
	auth.SetSecretKey(secret)
	ctx, err := auth.ContextWithValue(context.Background())
	if err != nil {
		t.Fatalf("auth context: %v", err)
	}

	lk, _, err := rc.OptionsAPI.CreateListenKeyV1(ctx).Execute()
	if err != nil || lk == nil || lk.ListenKey == nil || *lk.ListenKey == "" {
		t.Skipf("could not create listen key: %v", err)
		return
	}
	defer func() { _, _, _ = rc.OptionsAPI.DeleteListenKeyV1(ctx).Execute() }()

    // Dedicated WS client
    stc, err := NewStreamTestClientDedicated(getTestConfig())
    if err != nil {
        t.Fatalf("failed to create ws client: %v", err)
    }

    // User data channel
    ud := optionsstreams.NewUserDataStreamsChannel(stc.client)

    // Log current active server from SDK defaults
    if as := stc.client.GetActiveServer(); as != nil {
        t.Logf("Active WS server: name=%s url=%s", as.Name, as.URL)
    } else {
        t.Logf("Active WS server: <nil>")
    }

    // ---------- Connect ----------
    t.Run("Connect", func(t *testing.T) {
        cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        if err := ud.Connect(cctx, *lk.ListenKey); err != nil {
            cancel()
            t.Fatalf("connect user data: %v", err)
        }
        cancel()
    })

    // Ensure we cleanly disconnect at the end of the suite
    defer func() {
        dctx, dcancel := context.WithTimeout(context.Background(), 3*time.Second)
        _ = ud.Disconnect(dctx)
        dcancel()
    }()

    // ---------- Event Handlers ----------
    t.Run("AccountUpdateEvent", func(t *testing.T) {
        accCh := make(chan *models.AccountUpdateEvent, 1)
        ud.HandleAccountUpdateEvent(func(ctx context.Context, ev *models.AccountUpdateEvent) error {
            if ev.EventType != "ACCOUNT_UPDATE" {
                t.Errorf("account update e mismatch: %s", ev.EventType)
            }
            assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
            if len(ev.AccountBalanceArray) > 0 {
                bal := ev.AccountBalanceArray[0]
                if bal.MarginAsset != "" {
                    _ = tryParseFloat(t, bal.AccountBalance, "accountBalance")
                    _ = tryParseFloat(t, bal.PositionValue, "positionValue")
                }
            }
            if len(ev.PositionsArray) > 0 {
                p := ev.PositionsArray[0]
                if p.ContractSymbol != "" {
                    assertOptionSymbolFormat(t, p.ContractSymbol)
                    _ = tryParseFloat(t, p.CurrentPositions, "currentPositions")
                }
            }
            select {
            case accCh <- ev:
            default:
            }
            return nil
        })
        // Wait for one event or timeout (acceptable due to low activity)
        select {
        case ev := <-accCh:
            logJSON(t, "event.accountUpdate", ev)
        case <-time.After(eventWait()):
            t.Logf("timeout waiting ACCOUNT_UPDATE (acceptable)")
        }
    })

    t.Run("OrderTradeUpdateEvent", func(t *testing.T) {
        ordCh := make(chan *models.OrderTradeUpdateEvent, 1)
        ud.HandleOrderTradeUpdateEvent(func(ctx context.Context, ev *models.OrderTradeUpdateEvent) error {
            if ev.EventType != "ORDER_TRADE_UPDATE" {
                t.Errorf("order update e mismatch: %s", ev.EventType)
            }
            assertRecentMs(t, ev.EventTime, 24*time.Hour, "eventTime")
            if len(ev.OrderDetailsArray) > 0 {
                od := ev.OrderDetailsArray[0]
                if od.Symbol != "" {
                    assertOptionSymbolFormat(t, od.Symbol)
                }
                if od.OrderPrice != "" {
                    _ = tryParseFloat(t, od.OrderPrice, "orderPrice")
                }
                if od.OrderQuantity != "" {
                    _ = tryParseFloat(t, od.OrderQuantity, "orderQty")
                }
                if od.CompletedTradeVolume != "" {
                    _ = tryParseFloat(t, od.CompletedTradeVolume, "completedVol")
                }
                if od.CompletedTradeAmount != "" {
                    _ = tryParseFloat(t, od.CompletedTradeAmount, "completedAmt")
                }
                if od.OrderCreateTime > 0 && od.OrderUpdateTime > 0 && od.OrderCreateTime > od.OrderUpdateTime {
                    t.Logf("order times inverted: create %d update %d", od.OrderCreateTime, od.OrderUpdateTime)
                }
            }
            select {
            case ordCh <- ev:
            default:
            }
            return nil
        })

        // Optionally try to trigger an order update event (guarded by env)
        if os.Getenv("ENABLE_USERDATA_ORDER_TESTS") == "1" {
            // Pick an underlying and symbol
            underlying := "ETH"
            if u := os.Getenv("PREFERRED_UNDERLYING"); u != "" {
                underlying = strings.Split(u, ",")[0]
            }
            symbol, selErr := selectATMSymbol(underlying, "")
            if selErr == nil && symbol != "" {
                // Place a tiny limit order far from market to avoid fill
                qty := os.Getenv("TEST_ORDER_QTY")
                if qty == "" {
                    qty = "1"
                }
                price := os.Getenv("TEST_ORDER_PRICE")
                if price == "" {
                    price = "1"
                }
                side := os.Getenv("TEST_ORDER_SIDE")
                if side == "" {
                    side = "BUY"
                }
                tif := os.Getenv("TEST_ORDER_TIF")
                if tif == "" {
                    tif = "GTC"
                }
                _, _, _ = rc.OptionsAPI.CreateOrderV1(ctx).
                    Symbol(symbol).
                    Side(side).
                    Type_("LIMIT").
                    Quantity(qty).
                    Price(price).
                    TimeInForce(tif).
                    NewOrderRespType("RESULT").
                    Timestamp(time.Now().UnixMilli()).
                    Execute()
                // brief delay to allow event propagation
                time.Sleep(2 * time.Second)
            }
        }

        // Wait for one event or timeout (acceptable)
        select {
        case ev := <-ordCh:
            if len(ev.OrderDetailsArray) > 0 && len(ev.OrderDetailsArray[0].FillsArray) > 0 {
                f := ev.OrderDetailsArray[0].FillsArray[0]
                if f.TradePrice != "" {
                    _ = tryParseFloat(t, f.TradePrice, "fill.price")
                }
                if f.TradeQuantity != "" {
                    _ = tryParseFloat(t, f.TradeQuantity, "fill.qty")
                }
            }
            logJSON(t, "event.orderTradeUpdate", ev)
        case <-time.After(eventWait()):
            t.Logf("timeout waiting ORDER_TRADE_UPDATE (acceptable)")
        }
    })

    t.Run("RiskLevelChangeEvent", func(t *testing.T) {
        riskCh := make(chan *models.RiskLevelChangeEvent, 1)
        ud.HandleRiskLevelChangeEvent(func(ctx context.Context, ev *models.RiskLevelChangeEvent) error {
            if ev.EventType != "RISK_LEVEL_CHANGE" {
                t.Errorf("risk level e mismatch: %s", ev.EventType)
            }
            if ev.MarginBalance != "" {
                _ = tryParseFloat(t, ev.MarginBalance, "marginBalance")
            }
            if ev.MaintenanceMargin != "" {
                _ = tryParseFloat(t, ev.MaintenanceMargin, "maintenanceMargin")
            }
            select {
            case riskCh <- ev:
            default:
            }
            return nil
        })
        // Wait for one event or timeout (acceptable)
        select {
        case ev := <-riskCh:
            logJSON(t, "event.riskLevelChange", ev)
        case <-time.After(eventWait()):
            t.Logf("timeout waiting RISK_LEVEL_CHANGE (acceptable)")
        }
    })
}
