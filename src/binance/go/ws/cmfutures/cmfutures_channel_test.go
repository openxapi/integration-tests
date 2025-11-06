package wstest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	restcm "github.com/openxapi/binance-go/rest/cmfutures"
	cmfutures "github.com/openxapi/binance-go/ws/cmfutures"
	"github.com/openxapi/binance-go/ws/cmfutures/models"
)

type unhandledCatcher struct {
	mu      sync.Mutex
	matches []string
}

func (c *unhandledCatcher) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("unhandled message:")) {
		c.mu.Lock()
		c.matches = append(c.matches, strings.TrimSpace(string(p)))
		c.mu.Unlock()
	}
	return len(p), nil
}

func newUnhandledCatcher() *unhandledCatcher { return &unhandledCatcher{} }

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultRequestTimeout)
}

func newResponseHandler[T any](t testing.TB, label string) (*func(context.Context, *T, error) error, <-chan *T) {
	t.Helper()
	ch := make(chan *T, 1)
	handler := func(ctx context.Context, resp *T, err error) error {
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", label, err)
			return nil
		}
		if resp == nil {
			t.Fatalf("%s: nil response", label)
			return nil
		}
		select {
		case ch <- resp:
		default:
		}
		return nil
	}
	return &handler, ch
}

func newErrorResponseHandler[T any](t testing.TB, label string) (*func(context.Context, *T, error) error, <-chan error) {
	t.Helper()
	ch := make(chan error, 1)
	handler := func(ctx context.Context, resp *T, err error) error {
		if err == nil {
			t.Fatalf("%s: expected error, got response %#v", label, resp)
			return nil
		}
		select {
		case ch <- err:
		default:
		}
		return nil
	}
	return &handler, ch
}

func awaitResponse[T any](t testing.TB, ch <-chan *T, label string) *T {
	t.Helper()
	select {
	case resp := <-ch:
		return resp
	case <-time.After(defaultRequestTimeout):
		t.Fatalf("%s: timeout waiting for response", label)
		return nil
	}
}

func awaitError(t testing.TB, ch <-chan error, label string) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(defaultRequestTimeout):
		t.Fatalf("%s: timeout waiting for error", label)
		return nil
	}
}

func TestFullIntegrationSuite_CmFutures(t *testing.T) {
	if testing.Short() {
		t.Skip("cmfutures WS integration suite slow; skipping in short mode")
	}

	catcher := newUnhandledCatcher()
	prevWriter := log.Writer()
	log.SetOutput(catcher)
	defer func() {
		log.SetOutput(prevWriter)
		catcher.mu.Lock()
		defer catcher.mu.Unlock()
		if len(catcher.matches) > 0 {
			for _, line := range catcher.matches {
				t.Logf("SDK log: %s", line)
			}
			t.Fatalf("SDK emitted %d unhandled message log(s)", len(catcher.matches))
		}
	}()

	creds := getCreds()

	publicHarness := newChannelHarness(t, creds.Public)
	publicHarness.connect(t)
	t.Cleanup(func() { publicHarness.disconnect(t) })

	var hmacHarness *channelHarness
	if creds.HMAC != nil {
		hmacHarness = newChannelHarness(t, creds.HMAC)
		hmacHarness.connect(t)
		t.Cleanup(func() { hmacHarness.disconnect(t) })
	}

	var edHarness *channelHarness
	if creds.Ed25519 != nil {
		edHarness = newChannelHarness(t, creds.Ed25519)
		edHarness.connect(t)
		t.Cleanup(func() { edHarness.disconnect(t) })
	}

	restClient := newCmFuturesRESTClient()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	symbol, err := pickCoinFuturesSymbol(ctx, restClient)
	if err != nil {
		t.Fatalf("failed to discover active symbol: %v", err)
	}
	if symbol == "" {
		t.Fatalf("futures symbol discovery returned empty value")
	}

	t.Run("AccountRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil || !hmacHarness.Config.supports(cmfutures.AuthTypeUserData) {
			t.Skip("HMAC credentials not configured for USER_DATA")
		}
		runAccountRequests(t, hmacHarness)
	})

	t.Run("TradingRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil || !hmacHarness.Config.supports(cmfutures.AuthTypeTrade) {
			t.Skip("HMAC credentials not configured for TRADE")
		}
		runTradingRequests(t, hmacHarness, symbol, restClient)
	})

	t.Run("SessionRequests", func(t *testing.T) {
		runSessionRequests(t, publicHarness, edHarness)
	})
}

func runAccountRequests(t *testing.T, h *channelHarness) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer unavailable for account requests")
	}

	t.Run("AccountBalance", func(t *testing.T) {
		req := &models.AccountBalanceRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, cmfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.balance.request", req)
		handler, ch := newResponseHandler[models.AccountBalanceResponse](t, "account.balance.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountBalance(ctx, req, handler); err != nil {
			t.Fatalf("account.balance call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.balance.response")
		validateAccountBalanceResponse(t, resp)
	})

	t.Run("AccountPosition", func(t *testing.T) {
		req := &models.AccountPositionRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, cmfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.position.request", req)
		handler, ch := newResponseHandler[models.AccountPositionResponse](t, "account.position.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountPosition(ctx, req, handler); err != nil {
			t.Fatalf("account.position call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.position.response")
		validateAccountPositionResponse(t, resp)
	})

	t.Run("AccountStatus", func(t *testing.T) {
		req := &models.AccountStatusRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, cmfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.status.request", req)
		handler, ch := newResponseHandler[models.AccountStatusResponse](t, "account.status.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountStatus(ctx, req, handler); err != nil {
			t.Fatalf("account.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.status.response")
		validateAccountStatusResponse(t, resp)
	})
}

func runTradingRequests(t *testing.T, h *channelHarness, symbol string, rc *restcm.APIClient) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer unavailable for trading requests")
	}

	params, err := prepareFuturesLimitOrderParams(context.Background(), rc, symbol, "BUY")
	if err != nil {
		t.Fatalf("failed to prepare order parameters: %v", err)
	}

	orderState := struct {
		orderID       int64
		clientOrderID string
		positionSide  string
		side          string
	}{
		orderID:       0,
		clientOrderID: "",
		positionSide:  "",
		side:          "",
	}

	t.Run("OrderCancelUnknownOrder", func(t *testing.T) {
		req := &models.OrderCancelRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = time.Now().UnixNano()
		req.Params.RecvWindow = 5000
		body := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, cmfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.cancel.unknown.request", req)
		handler, errCh := newErrorResponseHandler[models.OrderCancelResponse](t, "order.cancel.unknown.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderCancel(ctx, req, handler); err != nil {
			t.Fatalf("order.cancel unknown call failed: %v", err)
		}
		err := awaitError(t, errCh, "order.cancel.unknown.response")
		var apiErr *models.ErrorMessage
		if !errors.As(err, &apiErr) {
			t.Fatalf("order.cancel unknown error not ErrorMessage: %T %v", err, err)
		}
		if apiErr.Status == 0 {
			t.Errorf("order.cancel unknown status unset")
		}
		if apiErr.ErrorPayload.Code == 0 {
			t.Errorf("order.cancel unknown code unset")
		}
		if strings.TrimSpace(apiErr.ErrorPayload.Msg) == "" {
			t.Errorf("order.cancel unknown message empty")
		}
	})

	t.Run("OrderPlace", func(t *testing.T) {
		req := &models.OrderPlaceRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = params.Quantity
		req.Params.Price = params.InitialPrice
		req.Params.NewClientOrderId = fmt.Sprintf("wstest-%d", time.Now().UnixNano())
		req.Params.RecvWindow = 5000
		body := map[string]interface{}{
			"symbol":           req.Params.Symbol,
			"side":             req.Params.Side,
			"type":             req.Params.Type,
			"timeInForce":      req.Params.TimeInForce,
			"quantity":         req.Params.Quantity,
			"price":            req.Params.Price,
			"newClientOrderId": req.Params.NewClientOrderId,
			"recvWindow":       req.Params.RecvWindow,
		}
		orderState.positionSide = determineOrderPositionSide(t, rc, h.Config, req.Params.Side)
		orderState.side = req.Params.Side
		if orderState.positionSide != "" {
			req.Params.PositionSide = orderState.positionSide
			body["positionSide"] = orderState.positionSide
		}
		signAndApply(t, h.Signer, cmfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.place.request", req)
		handler, ch := newResponseHandler[models.OrderPlaceResponse](t, "order.place.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderPlace(ctx, req, handler); err != nil {
			t.Fatalf("order.place call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.place.response")
		validateOrderPlaceResponse(t, resp, req.Params.Symbol, req.Params.Side, req.Params.Type)
		orderState.orderID = resp.Result.OrderId
		orderState.clientOrderID = resp.Result.ClientOrderId
	})

	t.Run("OrderStatus", func(t *testing.T) {
		if orderState.orderID == 0 {
			t.Skip("order not placed")
		}
		req := &models.OrderStatusRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = orderState.orderID
		req.Params.RecvWindow = 5000
		body := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, cmfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.status.request", req)
		handler, ch := newResponseHandler[models.OrderStatusResponse](t, "order.status.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderStatus(ctx, req, handler); err != nil {
			t.Fatalf("order.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.status.response")
		validateOrderStatusResponse(t, resp, req.Params.Symbol, orderState.orderID)
	})

	t.Run("OrderModify", func(t *testing.T) {
		if orderState.orderID == 0 {
			t.Skip("order not placed")
		}
		req := &models.OrderModifyRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = orderState.orderID
		req.Params.Price = params.ModifyPrice
		req.Params.Quantity = params.Quantity
		if orderState.side == "" {
			orderState.side = "BUY"
		}
		req.Params.Side = orderState.side
		req.Params.RecvWindow = 5000
		body := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"price":      req.Params.Price,
			"quantity":   req.Params.Quantity,
			"side":       req.Params.Side,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, cmfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.modify.request", req)
		handler, ch := newResponseHandler[models.OrderModifyResponse](t, "order.modify.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderModify(ctx, req, handler); err != nil {
			t.Fatalf("order.modify call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.modify.response")
		validateOrderModifyResponse(t, resp, req.Params.Symbol, orderState.orderID, req.Params.Price)
	})

	t.Run("OrderCancel", func(t *testing.T) {
		if orderState.orderID == 0 {
			t.Skip("order not placed")
		}
		req := &models.OrderCancelRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = orderState.orderID
		req.Params.RecvWindow = 5000
		body := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, cmfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.cancel.request", req)
		handler, ch := newResponseHandler[models.OrderCancelResponse](t, "order.cancel.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderCancel(ctx, req, handler); err != nil {
			t.Fatalf("order.cancel call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.cancel.response")
		validateOrderCancelResponse(t, resp, req.Params.Symbol, orderState.orderID)
	})
}

func runSessionRequests(t *testing.T, public *channelHarness, ed *channelHarness) {
	t.Helper()

	t.Run("SessionStatus", func(t *testing.T) {
		req := &models.SessionStatusRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		logRequestOnFailure(t, "session.status.request", req)
		handler, ch := newResponseHandler[models.SessionStatusResponse](t, "session.status.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := public.Channel.SessionStatus(ctx, req, handler); err != nil {
			t.Fatalf("session.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.status.response")
		validateSessionStatusResponse(t, resp)
	})

	t.Run("SessionLogout", func(t *testing.T) {
		req := &models.SessionLogoutRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		logRequestOnFailure(t, "session.logout.request", req)
		handler, ch := newResponseHandler[models.SessionLogoutResponse](t, "session.logout.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := public.Channel.SessionLogout(ctx, req, handler); err != nil {
			t.Fatalf("session.logout call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.logout.response")
		validateSessionLogoutResponse(t, resp)
	})

	t.Run("SessionLogon", func(t *testing.T) {
		if ed == nil || ed.Signer == nil || !ed.Config.supports(cmfutures.AuthTypeSigned) {
			t.Skip("Ed25519 credentials not configured for SIGNED session.logon")
		}
		if shouldSkipSessionLogon(ed.Client) {
			t.Skip("session.logon not available on testnet")
		}
		req := &models.SessionLogonRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		payload := map[string]interface{}{}
		signAndApply(t, ed.Signer, cmfutures.AuthTypeSigned, payload, &req.Params)
		logRequestOnFailure(t, "session.logon.request", req)
		handler, ch := newResponseHandler[models.SessionLogonResponse](t, "session.logon.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := ed.Channel.SessionLogon(ctx, req, handler); err != nil {
			t.Fatalf("session.logon call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.logon.response")
		validateSessionLogonResponse(t, resp, ed.Config.APIKey)
	})
}

func validateAccountBalanceResponse(t *testing.T, resp *models.AccountBalanceResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("account.balance status: want 200 got %d", resp.Status)
	}
	if len(resp.Result) == 0 {
		t.Error("account.balance result empty")
	}
	for i, bal := range resp.Result {
		assertNonEmpty(t, bal.Asset, fmt.Sprintf("account.balance[%d].asset", i))
		tryParseFloat(t, bal.Balance, fmt.Sprintf("account.balance[%d].balance", i))
		tryParseFloat(t, bal.AvailableBalance, fmt.Sprintf("account.balance[%d].availableBalance", i))
		if bal.UpdateTime < 0 {
			t.Errorf("account.balance[%d].updateTime < 0", i)
		}
	}
}

func validateAccountPositionResponse(t *testing.T, resp *models.AccountPositionResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("account.position status: want 200 got %d", resp.Status)
	}
	for i, pos := range resp.Result {
		assertNonEmpty(t, pos.Symbol, fmt.Sprintf("account.position[%d].symbol", i))
		tryParseFloat(t, pos.PositionAmt, fmt.Sprintf("account.position[%d].positionAmt", i))
		tryParseFloat(t, pos.EntryPrice, fmt.Sprintf("account.position[%d].entryPrice", i))
		tryParseFloat(t, pos.MarkPrice, fmt.Sprintf("account.position[%d].markPrice", i))
		if pos.UpdateTime < 0 {
			t.Errorf("account.position[%d].updateTime < 0", i)
		}
	}
}

func validateAccountStatusResponse(t *testing.T, resp *models.AccountStatusResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("account.status status: want 200 got %d", resp.Status)
	}
	if resp.Result.FeeTier < 0 {
		t.Errorf("account.status fee tier < 0: %d", resp.Result.FeeTier)
	}
	if len(resp.Result.Assets) > 0 {
		for i, asset := range resp.Result.Assets {
			assertNonEmpty(t, asset.Asset, fmt.Sprintf("account.status.assets[%d].asset", i))
			tryParseFloat(t, asset.MarginBalance, fmt.Sprintf("account.status.assets[%d].marginBalance", i))
		}
	}
	if len(resp.Result.Positions) > 0 {
		for i, pos := range resp.Result.Positions {
			assertNonEmpty(t, pos.Symbol, fmt.Sprintf("account.status.positions[%d].symbol", i))
			tryParseFloat(t, pos.EntryPrice, fmt.Sprintf("account.status.positions[%d].entryPrice", i))
			tryParseFloat(t, pos.PositionAmt, fmt.Sprintf("account.status.positions[%d].positionAmt", i))
		}
	}
}

func validateOrderPlaceResponse(t *testing.T, resp *models.OrderPlaceResponse, symbol string, side string, typ string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.place status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId == 0 {
		t.Fatalf("order.place missing orderId")
	}
	if resp.Result.Symbol != "" && !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Errorf("order.place symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	if !strings.EqualFold(resp.Result.Side, side) && resp.Result.Side != "" {
		t.Errorf("order.place side mismatch: want %s got %s", side, resp.Result.Side)
	}
	if !strings.EqualFold(resp.Result.Type, typ) && resp.Result.Type != "" {
		t.Errorf("order.place type mismatch: want %s got %s", typ, resp.Result.Type)
	}
	assertNonEmpty(t, resp.Result.ClientOrderId, "order.place clientOrderId")
}

func validateOrderStatusResponse(t *testing.T, resp *models.OrderStatusResponse, symbol string, orderID int64) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.status status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Errorf("order.status orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if resp.Result.Symbol != "" && !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Errorf("order.status symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
}

func validateOrderModifyResponse(t *testing.T, resp *models.OrderModifyResponse, symbol string, orderID int64, price string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.modify status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Errorf("order.modify orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if resp.Result.Symbol != "" && !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Errorf("order.modify symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	if resp.Result.Price != "" && resp.Result.Price != price {
		t.Errorf("order.modify price mismatch: want %s got %s", price, resp.Result.Price)
	}
}

func validateOrderCancelResponse(t *testing.T, resp *models.OrderCancelResponse, symbol string, orderID int64) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.cancel status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Errorf("order.cancel orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if resp.Result.Symbol != "" && !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Errorf("order.cancel symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
}

func validateSessionStatusResponse(t *testing.T, resp *models.SessionStatusResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.status status: want 200 got %d", resp.Status)
	}
	if resp.Result.ServerTime <= 0 {
		t.Errorf("session.status serverTime <= 0: %d", resp.Result.ServerTime)
	} else {
		assertRecentMs(t, resp.Result.ServerTime, 2*time.Minute, "session.status serverTime")
	}
}

func validateSessionLogoutResponse(t *testing.T, resp *models.SessionLogoutResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.logout status: want 200 got %d", resp.Status)
	}
	if resp.Result.ServerTime <= 0 {
		t.Errorf("session.logout serverTime <= 0: %d", resp.Result.ServerTime)
	} else {
		assertRecentMs(t, resp.Result.ServerTime, 2*time.Minute, "session.logout serverTime")
	}
}

func validateSessionLogonResponse(t *testing.T, resp *models.SessionLogonResponse, expectedAPIKey string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.logon status: want 200 got %d", resp.Status)
	}
	if expectedAPIKey != "" && !strings.EqualFold(resp.Result.ApiKey, expectedAPIKey) && resp.Result.ApiKey != "" {
		t.Errorf("session.logon apiKey mismatch: want %s got %s", expectedAPIKey, resp.Result.ApiKey)
	}
	if resp.Result.ServerTime <= 0 {
		t.Errorf("session.logon serverTime <= 0: %d", resp.Result.ServerTime)
	} else {
		assertRecentMs(t, resp.Result.ServerTime, 2*time.Minute, "session.logon serverTime")
	}
}

func pickCoinFuturesSymbol(ctx context.Context, rc *restcm.APIClient) (string, error) {
	candidates := []string{}
	if pref := strings.TrimSpace(os.Getenv("BINANCE_CMFUTURES_SYMBOL")); pref != "" {
		for _, part := range strings.Split(pref, ",") {
			if trimmed := strings.ToUpper(strings.TrimSpace(part)); trimmed != "" {
				candidates = append(candidates, trimmed)
			}
		}
	}
	if len(candidates) == 0 {
		if pref := strings.TrimSpace(os.Getenv("PREFERRED_SYMBOL")); pref != "" {
			for _, part := range strings.Split(pref, ",") {
				if trimmed := strings.ToUpper(strings.TrimSpace(part)); trimmed != "" {
					candidates = append(candidates, trimmed)
				}
			}
		}
	}
	candidates = append(candidates, "BTCUSD_PERP")
	for _, sym := range candidates {
		if sym == "" {
			continue
		}
		if _, err := loadFuturesSymbolConstraints(ctx, rc, sym); err == nil {
			return sym, nil
		}
	}
	return "", fmt.Errorf("no suitable symbol found")
}

func shouldSkipSessionLogon(client *cmfutures.Client) bool {
	if client == nil {
		return true
	}
	info := client.GetActiveServer()
	if info == nil {
		return true
	}
	name := strings.TrimSpace(strings.ToLower(info.Name))
	if name == strings.ToLower(defaultTestnetServer) || strings.Contains(name, "testnet") {
		return true
	}
	if strings.Contains(strings.ToLower(info.URL), "testnet") {
		return true
	}
	return false
}
