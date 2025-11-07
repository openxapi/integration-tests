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

	restum "github.com/openxapi/binance-go/rest/umfutures"
	umfutures "github.com/openxapi/binance-go/ws/umfutures"
	"github.com/openxapi/binance-go/ws/umfutures/models"
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

type responseResult[T any] struct {
	resp *T
	err  error
}

func newResponseHandler[T any](t testing.TB) (*func(context.Context, *T, error) error, <-chan responseResult[T]) {
	t.Helper()
	ch := make(chan responseResult[T], 1)
	handler := func(ctx context.Context, resp *T, err error) error {
		select {
		case ch <- responseResult[T]{resp: resp, err: err}:
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

func awaitResponse[T any](t testing.TB, ch <-chan responseResult[T], label string) *T {
	t.Helper()
	timer := time.NewTimer(defaultRequestTimeout)
	defer timer.Stop()
	select {
	case result := <-ch:
		if result.err != nil {
			t.Fatalf("%s: unexpected error: %v", label, result.err)
			return nil
		}
		if result.resp == nil {
			t.Fatalf("%s: nil response", label)
			return nil
		}
		return result.resp
	case <-timer.C:
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

func TestFullIntegrationSuite_UmFutures(t *testing.T) {
	if testing.Short() {
		t.Skip("umfutures WS integration suite slow; skipping in short mode")
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

	var rsaHarness *channelHarness
	if creds.RSA != nil {
		rsaHarness = newChannelHarness(t, creds.RSA)
		rsaHarness.connect(t)
		t.Cleanup(func() { rsaHarness.disconnect(t) })
	}

	var edHarness *channelHarness
	if creds.Ed25519 != nil {
		edHarness = newChannelHarness(t, creds.Ed25519)
		edHarness.connect(t)
		t.Cleanup(func() { edHarness.disconnect(t) })
	}

	restClient := newUmFuturesRESTClient()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	symbol, err := pickFuturesSymbol(ctx, restClient)
	if err != nil {
		t.Fatalf("failed to discover active symbol: %v", err)
	}
	if symbol == "" {
		t.Fatalf("futures symbol discovery returned empty value")
	}

	t.Run("PublicRequests", func(t *testing.T) {
		runPublicRequests(t, publicHarness, symbol)
	})

	t.Run("UserDataRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil || !hmacHarness.Config.supports(umfutures.AuthTypeUserData) {
			t.Skip("HMAC credentials not configured for USER_DATA")
		}
		runUserDataRequests(t, hmacHarness)
	})

	t.Run("TradingRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil || !hmacHarness.Config.supports(umfutures.AuthTypeTrade) {
			t.Skip("HMAC credentials not configured for TRADE")
		}
		runTradingRequests(t, hmacHarness, symbol, restClient)
	})

	t.Run("SessionRequests", func(t *testing.T) {
		runSessionRequests(t, publicHarness, hmacHarness, rsaHarness, edHarness)
	})
}

func runPublicRequests(t *testing.T, h *channelHarness, symbol string) {
	t.Helper()
	symbol = strings.ToUpper(symbol)

	t.Run("TickerPrice", func(t *testing.T) {
		req := &models.TickerPriceRequest{
			Id: models.NewMessageIDInt64(time.Now().UnixNano()),
		}
		req.Params.Symbol = symbol
		logRequestOnFailure(t, "ticker.price.request", req)
		handler, ch := newResponseHandler[models.TickerPriceResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TickerPrice(ctx, req, handler); err != nil {
			t.Fatalf("ticker.price call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "ticker.price.response")
		validateTickerPriceResponse(t, resp, symbol)
	})

	t.Run("TickerBook", func(t *testing.T) {
		req := &models.TickerBookRequest{
			Id: models.NewMessageIDInt64(time.Now().UnixNano()),
		}
		req.Params.Symbol = symbol
		logRequestOnFailure(t, "ticker.book.request", req)
		handler, ch := newResponseHandler[models.TickerBookResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TickerBook(ctx, req, handler); err != nil {
			t.Fatalf("ticker.book call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "ticker.book.response")
		validateTickerBookResponse(t, resp, symbol)
	})

	t.Run("Depth", func(t *testing.T) {
		req := &models.DepthRequest{
			Id: models.NewMessageIDInt64(time.Now().UnixNano()),
		}
		req.Params.Symbol = symbol
		req.Params.Limit = 20
		logRequestOnFailure(t, "depth.request", req)
		handler, ch := newResponseHandler[models.DepthResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Depth(ctx, req, handler); err != nil {
			t.Fatalf("depth call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "depth.response")
		validateDepthResponse(t, resp)
	})

	t.Run("DepthInvalidSymbol", func(t *testing.T) {
		req := &models.DepthRequest{
			Id: models.NewMessageIDInt64(time.Now().UnixNano()),
		}
		req.Params.Symbol = "INVALIDPAIR"
		req.Params.Limit = 5
		logRequestOnFailure(t, "depth.invalid.request", req)
		handler, errCh := newErrorResponseHandler[models.DepthResponse](t, "depth.invalid.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Depth(ctx, req, handler); err != nil {
			t.Fatalf("depth invalid call failed: %v", err)
		}
		err := awaitError(t, errCh, "depth.invalid.response")
		var apiErr *models.ErrorMessage
		if !errors.As(err, &apiErr) {
			t.Fatalf("depth invalid error not ErrorMessage: %T %v", err, err)
		}
		if apiErr.Status == 0 {
			t.Errorf("depth invalid error status unset")
		}
		if apiErr.ErrorPayload.Code == 0 {
			t.Errorf("depth invalid error code unset")
		}
		if strings.TrimSpace(apiErr.ErrorPayload.Msg) == "" {
			t.Errorf("depth invalid error message empty")
		}
	})
}

func runUserDataRequests(t *testing.T, h *channelHarness) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer unavailable for user data requests")
	}

	t.Run("AccountBalance", func(t *testing.T) {
		req := &models.AccountBalanceRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.balance.request", req)
		handler, ch := newResponseHandler[models.AccountBalanceResponse](t)
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
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.position.request", req)
		handler, ch := newResponseHandler[models.AccountPositionResponse](t)
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
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.status.request", req)
		handler, ch := newResponseHandler[models.AccountStatusResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountStatus(ctx, req, handler); err != nil {
			t.Fatalf("account.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.status.response")
		validateAccountStatusResponse(t, resp)
	})

	t.Run("V2AccountBalance", func(t *testing.T) {
		req := &models.V2AccountBalanceRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "v2.account.balance.request", req)
		handler, ch := newResponseHandler[models.V2AccountBalanceResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.V2AccountBalance(ctx, req, handler); err != nil {
			t.Fatalf("v2/account.balance call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "v2.account.balance.response")
		validateV2AccountBalanceResponse(t, resp)
	})

	t.Run("V2AccountPosition", func(t *testing.T) {
		req := &models.V2AccountPositionRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "v2.account.position.request", req)
		handler, ch := newResponseHandler[models.V2AccountPositionResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.V2AccountPosition(ctx, req, handler); err != nil {
			t.Fatalf("v2/account.position call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "v2.account.position.response")
		validateV2AccountPositionResponse(t, resp)
	})

	t.Run("V2AccountStatus", func(t *testing.T) {
		req := &models.V2AccountStatusRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, umfutures.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "v2.account.status.request", req)
		handler, ch := newResponseHandler[models.V2AccountStatusResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.V2AccountStatus(ctx, req, handler); err != nil {
			t.Fatalf("v2/account.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "v2.account.status.response")
		validateV2AccountStatusResponse(t, resp)
	})
}

func runTradingRequests(t *testing.T, h *channelHarness, symbol string, rc *restum.APIClient) {
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
	}{
		orderID:       0,
		clientOrderID: "",
		positionSide:  "",
	}

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
		if posSide := determineOrderPositionSide(t, rc, h.Config, req.Params.Side); posSide != "" {
			req.Params.PositionSide = posSide
			body["positionSide"] = posSide
			orderState.positionSide = posSide
		} else {
			orderState.positionSide = ""
		}
		signAndApply(t, h.Signer, umfutures.AuthTypeTrade, body, &req.Params)
		logRequestOnFailure(t, "order.place.request", req)
		handler, ch := newResponseHandler[models.OrderPlaceResponse](t)
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
		payload := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, umfutures.AuthTypeTrade, payload, &req.Params)
		logRequestOnFailure(t, "order.status.request", req)
		handler, ch := newResponseHandler[models.OrderStatusResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderStatus(ctx, req, handler); err != nil {
			t.Fatalf("order.status call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.status.response")
		validateOrderStatusResponse(t, resp, orderState.orderID, req.Params.Symbol)
	})

	t.Run("OrderModify", func(t *testing.T) {
		if orderState.orderID == 0 {
			t.Skip("order not placed")
		}
		req := &models.OrderModifyRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = orderState.orderID
		req.Params.Side = "BUY"
		req.Params.Quantity = params.Quantity
		req.Params.Price = params.ModifyPrice
		req.Params.RecvWindow = 5000
		payload := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"side":       req.Params.Side,
			"quantity":   req.Params.Quantity,
			"price":      req.Params.Price,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, umfutures.AuthTypeTrade, payload, &req.Params)
		logRequestOnFailure(t, "order.modify.request", req)
		handler, ch := newResponseHandler[models.OrderModifyResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderModify(ctx, req, handler); err != nil {
			t.Fatalf("order.modify call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.modify.response")
		validateOrderModifyResponse(t, resp, orderState.orderID, req.Params.Symbol, params.ModifyPrice)
	})

	t.Run("OrderCancel", func(t *testing.T) {
		if orderState.orderID == 0 {
			t.Skip("order not placed")
		}
		req := &models.OrderCancelRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.Symbol = strings.ToUpper(symbol)
		req.Params.OrderId = orderState.orderID
		req.Params.RecvWindow = 5000
		payload := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, umfutures.AuthTypeTrade, payload, &req.Params)
		logRequestOnFailure(t, "order.cancel.request", req)
		handler, ch := newResponseHandler[models.OrderCancelResponse](t)
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderCancel(ctx, req, handler); err != nil {
			t.Fatalf("order.cancel call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "order.cancel.response")
		validateOrderCancelResponse(t, resp, orderState.orderID, req.Params.Symbol)
	})
}

func determineOrderPositionSide(t *testing.T, rc *restum.APIClient, cfg *TestConfig, side string) string {
	t.Helper()
	if cfg == nil || strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return ""
	}
	if rc == nil {
		return ""
	}
	auth := restum.NewAuth(cfg.APIKey)
	auth.SetSecretKey(cfg.SecretKey)
	authCtx, err := auth.ContextWithValue(context.Background())
	if err != nil {
		t.Logf("order.place position side: auth context error: %v", err)
		return ""
	}
	ts := time.Now().UnixMilli()
	resp, _, err := rc.FuturesAPI.GetPositionSideDualV1(authCtx).
		Timestamp(ts).
		RecvWindow(5000).
		Execute()
	if err != nil {
		if ge, ok := err.(*restum.GenericOpenAPIError); ok {
			t.Logf("position side query failed: status=%s body=%s", ge.Error(), string(ge.Body()))
		} else {
			t.Logf("position side query failed: %v", err)
		}
		return ""
	}
	if resp == nil || resp.DualSidePosition == nil {
		return ""
	}
	if *resp.DualSidePosition {
		if strings.EqualFold(side, "SELL") {
			return "SHORT"
		}
		return "LONG"
	}
	return ""
}

func shouldSkipSessionLogon(client *umfutures.Client) bool {
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

func runSessionRequests(t *testing.T, public *channelHarness, hmac *channelHarness, rsa *channelHarness, ed *channelHarness) {
	t.Helper()

	t.Run("SessionStatus", func(t *testing.T) {
		req := &models.SessionStatusRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		logRequestOnFailure(t, "session.status.request", req)
		handler, ch := newResponseHandler[models.SessionStatusResponse](t)
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
		handler, ch := newResponseHandler[models.SessionLogoutResponse](t)
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
		if ed == nil || ed.Signer == nil || !ed.Config.supports(umfutures.AuthTypeSigned) {
			t.Skip("Ed25519 credentials not configured for SIGNED session.logon")
		}
		if shouldSkipSessionLogon(ed.Client) {
			t.Skip("session.logon not available on testnet")
		}
		req := &models.SessionLogonRequest{Id: models.NewMessageIDInt64(time.Now().UnixNano())}
		req.Params.RecvWindow = 5000
		payload := map[string]interface{}{
			"apiKey":     ed.Config.APIKey,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, ed.Signer, umfutures.AuthTypeSigned, payload, &req.Params)
		logRequestOnFailure(t, "session.logon.request", req)
		handler, ch := newResponseHandler[models.SessionLogonResponse](t)
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

func validateTickerPriceResponse(t *testing.T, resp *models.TickerPriceResponse, symbol string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("ticker.price status: want 200 got %d", resp.Status)
	}
	if !strings.EqualFold(resp.Result.Symbol, symbol) && resp.Result.Symbol != "" {
		t.Fatalf("ticker.price symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	assertNonEmpty(t, resp.Result.Price, "ticker.price price")
	if price := tryParseFloat(t, resp.Result.Price, "ticker.price price"); price <= 0 {
		t.Errorf("ticker.price price <= 0: %f", price)
	}
	if resp.Result.Time > 0 {
		assertRecentMs(t, resp.Result.Time, 2*time.Minute, "ticker.price time")
	}
}

func validateTickerBookResponse(t *testing.T, resp *models.TickerBookResponse, symbol string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("ticker.book status: want 200 got %d", resp.Status)
	}
	if !strings.EqualFold(resp.Result.Symbol, symbol) && resp.Result.Symbol != "" {
		t.Fatalf("ticker.book symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	assertNonEmpty(t, resp.Result.BidPrice, "ticker.book bidPrice")
	assertNonEmpty(t, resp.Result.BidQty, "ticker.book bidQty")
	assertNonEmpty(t, resp.Result.AskPrice, "ticker.book askPrice")
	assertNonEmpty(t, resp.Result.AskQty, "ticker.book askQty")
	if bp := tryParseFloat(t, resp.Result.BidPrice, "ticker.book bidPrice"); bp <= 0 {
		t.Errorf("ticker.book bidPrice <= 0: %f", bp)
	}
	if ap := tryParseFloat(t, resp.Result.AskPrice, "ticker.book askPrice"); ap <= 0 {
		t.Errorf("ticker.book askPrice <= 0: %f", ap)
	}
	tryParseFloat(t, resp.Result.BidQty, "ticker.book bidQty")
	tryParseFloat(t, resp.Result.AskQty, "ticker.book askQty")
}

func validateDepthResponse(t *testing.T, resp *models.DepthResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("depth status: want 200 got %d", resp.Status)
	}
	if resp.Result.LastUpdateId <= 0 {
		t.Errorf("depth lastUpdateId <= 0: %d", resp.Result.LastUpdateId)
	}
	if len(resp.Result.Bids) == 0 {
		t.Error("depth bids empty")
	}
	if len(resp.Result.Asks) == 0 {
		t.Error("depth asks empty")
	}
	checkDepthSide := func(side string, levels [][]string) {
		for i, level := range levels {
			if len(level) < 2 {
				t.Fatalf("%s level %d malformed: %#v", side, i, level)
			}
			if price := tryParseFloat(t, level[0], fmt.Sprintf("%s price[%d]", side, i)); price <= 0 {
				t.Errorf("%s price[%d] <= 0: %f", side, i, price)
			}
			tryParseFloat(t, level[1], fmt.Sprintf("%s qty[%d]", side, i))
		}
	}
	checkDepthSide("bids", resp.Result.Bids)
	checkDepthSide("asks", resp.Result.Asks)
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
		tryParseFloat(t, pos.MarkPrice, fmt.Sprintf("account.position[%d].markPrice", i))
		tryParseFloat(t, pos.Notional, fmt.Sprintf("account.position[%d].notional", i))
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
	if len(resp.Result.Assets) == 0 {
		t.Error("account.status assets empty")
	}
	for i, asset := range resp.Result.Assets {
		assertNonEmpty(t, asset.Asset, fmt.Sprintf("account.status.assets[%d].asset", i))
		tryParseFloat(t, asset.AvailableBalance, fmt.Sprintf("account.status.assets[%d].availableBalance", i))
		tryParseFloat(t, asset.WalletBalance, fmt.Sprintf("account.status.assets[%d].walletBalance", i))
		if asset.UpdateTime < 0 {
			t.Errorf("account.status.assets[%d].updateTime < 0", i)
		}
	}
}

func validateV2AccountBalanceResponse(t *testing.T, resp *models.V2AccountBalanceResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("v2/account.balance status: want 200 got %d", resp.Status)
	}
	if len(resp.Result) == 0 {
		t.Error("v2/account.balance result empty")
	}
	for i, bal := range resp.Result {
		assertNonEmpty(t, bal.Asset, fmt.Sprintf("v2/account.balance[%d].asset", i))
		tryParseFloat(t, bal.Balance, fmt.Sprintf("v2/account.balance[%d].balance", i))
		if bal.UpdateTime < 0 {
			t.Errorf("v2/account.balance[%d].updateTime < 0", i)
		}
	}
}

func validateV2AccountPositionResponse(t *testing.T, resp *models.V2AccountPositionResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("v2/account.position status: want 200 got %d", resp.Status)
	}
	for i, pos := range resp.Result {
		assertNonEmpty(t, pos.Symbol, fmt.Sprintf("v2/account.position[%d].symbol", i))
		tryParseFloat(t, pos.PositionAmt, fmt.Sprintf("v2/account.position[%d].positionAmt", i))
		tryParseFloat(t, pos.MarkPrice, fmt.Sprintf("v2/account.position[%d].markPrice", i))
	}
}

func validateV2AccountStatusResponse(t *testing.T, resp *models.V2AccountStatusResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("v2/account.status status: want 200 got %d", resp.Status)
	}
	if len(resp.Result.Assets) == 0 {
		t.Error("v2/account.status assets empty")
	}
	for i, asset := range resp.Result.Assets {
		assertNonEmpty(t, asset.Asset, fmt.Sprintf("v2/account.status.assets[%d].asset", i))
		tryParseFloat(t, asset.AvailableBalance, fmt.Sprintf("v2/account.status.assets[%d].availableBalance", i))
	}
}

func validateOrderPlaceResponse(t *testing.T, resp *models.OrderPlaceResponse, symbol, side, typ string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.place status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId == 0 {
		t.Fatal("order.place returned orderId=0")
	}
	assertNonEmpty(t, resp.Result.ClientOrderId, "order.place clientOrderId")
	if !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Fatalf("order.place symbol: want %s got %s", symbol, resp.Result.Symbol)
	}
	if !strings.EqualFold(resp.Result.Side, side) {
		t.Fatalf("order.place side: want %s got %s", side, resp.Result.Side)
	}
	if !strings.EqualFold(resp.Result.Type, typ) {
		t.Fatalf("order.place type: want %s got %s", typ, resp.Result.Type)
	}
}

func validateOrderStatusResponse(t *testing.T, resp *models.OrderStatusResponse, orderID int64, symbol string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.status status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Fatalf("order.status orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Fatalf("order.status symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	if resp.Result.Time > 0 {
		assertRecentMs(t, resp.Result.Time, 10*time.Minute, "order.status time")
	}
}

func validateOrderModifyResponse(t *testing.T, resp *models.OrderModifyResponse, orderID int64, symbol, price string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.modify status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Fatalf("order.modify orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Fatalf("order.modify symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
	if resp.Result.Price != "" && resp.Result.Price != price {
		t.Fatalf("order.modify price mismatch: want %s got %s", price, resp.Result.Price)
	}
}

func validateOrderCancelResponse(t *testing.T, resp *models.OrderCancelResponse, orderID int64, symbol string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("order.cancel status: want 200 got %d", resp.Status)
	}
	if resp.Result.OrderId != orderID {
		t.Fatalf("order.cancel orderId mismatch: want %d got %d", orderID, resp.Result.OrderId)
	}
	if !strings.EqualFold(resp.Result.Symbol, symbol) {
		t.Fatalf("order.cancel symbol mismatch: want %s got %s", symbol, resp.Result.Symbol)
	}
}

func validateSessionStatusResponse(t *testing.T, resp *models.SessionStatusResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.status status: want 200 got %d", resp.Status)
	}
	if resp.Result.ServerTime > 0 {
		assertRecentMs(t, resp.Result.ServerTime, 2*time.Minute, "session.status serverTime")
	}
}

func validateSessionLogoutResponse(t *testing.T, resp *models.SessionLogoutResponse) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.logout status: want 200 got %d", resp.Status)
	}
}

func validateSessionLogonResponse(t *testing.T, resp *models.SessionLogonResponse, apiKey string) {
	t.Helper()
	if resp.Status != 200 {
		t.Fatalf("session.logon status: want 200 got %d", resp.Status)
	}
	if !strings.EqualFold(resp.Result.ApiKey, apiKey) {
		t.Fatalf("session.logon apiKey mismatch: want %s got %s", apiKey, resp.Result.ApiKey)
	}
	if resp.Result.ServerTime > 0 {
		assertRecentMs(t, resp.Result.ServerTime, 2*time.Minute, "session.logon serverTime")
	}
}

func pickFuturesSymbol(ctx context.Context, rc *restum.APIClient) (string, error) {
	candidates := []string{}
	if pref := strings.TrimSpace(os.Getenv("BINANCE_UMFUTURES_SYMBOL")); pref != "" {
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
	candidates = append(candidates, "BTCUSDT")
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
