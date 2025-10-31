package wstest

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	spot "github.com/openxapi/binance-go/ws/spot"
	"github.com/openxapi/binance-go/ws/spot/models"
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

func newResponseHandler[T any](t testing.TB, label string) (*func(context.Context, *T) error, <-chan *T) {
	t.Helper()
	ch := make(chan *T, 1)
	handler := func(ctx context.Context, resp *T) error {
		if resp == nil {
			t.Errorf("%s: nil response", label)
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

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultRequestTimeout)
}

func ensureSessionLoggedIn(t testing.TB, h *channelHarness) {
	t.Helper()
	if h.Signer == nil {
		// Without a signer we cannot perform the signed session.logon call.
		t.Skip("signer not available; cannot perform session.logon for user stream setup")
	}

	status := sessionStatusResponse(t, h, "session.status(pre-userStream)")
	if status.Status == 200 && status.Result.UserDataStream {
		return
	}

	logonID := time.Now().UnixNano()
	logonReq := &models.SessionLogonRequest{Id: models.NewMessageIDInt64(logonID)}
	params := map[string]interface{}{}
	signAndApply(t, h.Signer, spot.AuthTypeSigned, params, &logonReq.Params)
	logonHandler, logonCh := newResponseHandler[models.SessionLogonResponse](t, "session.logon(userStream)")
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.SessionLogon(ctx, logonReq, logonHandler); err != nil {
		t.Fatalf("session.logon(userStream) call failed: %v", err)
	}
	logonResp := awaitResponse(t, logonCh, "session.logon(userStream)")
	if got, ok := logonResp.Id.ValInt64(); ok && got != logonID {
		t.Fatalf("session.logon(userStream) id mismatch: want %d got %d", logonID, got)
	}
	if logonResp.Status != 200 {
		t.Fatalf("session.logon(userStream) status: want 200 got %d", logonResp.Status)
	}

	status = sessionStatusResponse(t, h, "session.status(post-userStream-logon)")
	if status.Status != 200 {
		t.Fatalf("session.status post-logon status: want 200 got %d", status.Status)
	}
	if !status.Result.UserDataStream {
		t.Logf("session.status post-logon indicates userDataStream inactive; proceeding with explicit subscription")
	}
}

func sessionStatusResponse(t testing.TB, h *channelHarness, label string) *models.SessionStatusResponse {
	t.Helper()
	id := time.Now().UnixNano()
	req := &models.SessionStatusRequest{Id: models.NewMessageIDInt64(id)}
	handler, ch := newResponseHandler[models.SessionStatusResponse](t, label)
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.SessionStatus(ctx, req, handler); err != nil {
		t.Fatalf("%s call failed: %v", label, err)
	}
	resp := awaitResponse(t, ch, label)
	if got, ok := resp.Id.ValInt64(); ok && got != id {
		t.Fatalf("%s id mismatch: want %d got %d", label, id, got)
	}
	return resp
}

func assertUserDataStreamState(t testing.TB, h *channelHarness, want bool, label string) {
	t.Helper()
	status := sessionStatusResponse(t, h, label)
	if status.Status != 200 {
		t.Fatalf("%s status: want 200 got %d", label, status.Status)
	}
	if status.Result.UserDataStream != want {
		t.Fatalf("%s userDataStream: want %v got %v", label, want, status.Result.UserDataStream)
	}
}

func ensureUserStreamInactive(t testing.TB, h *channelHarness) {
	t.Helper()
	status := sessionStatusResponse(t, h, "session.status(ensureUserStreamInactive)")
	if status.Status != 200 {
		t.Fatalf("session.status during ensureUserStreamInactive: want 200 got %d", status.Status)
	}
	if !status.Result.UserDataStream {
		return
	}
	resp := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(reset)", 0)
	if resp.Status != 200 {
		t.Fatalf("userDataStream.unsubscribe(reset) status: want 200 got %d", resp.Status)
	}
	status = sessionStatusResponse(t, h, "session.status(after ensureUserStreamInactive)")
	if status.Status != 200 {
		t.Fatalf("session.status after ensureUserStreamInactive: want 200 got %d", status.Status)
	}
	if status.Result.UserDataStream {
		t.Fatalf("user data stream remained active after reset unsubscribe")
	}
}

func userStreamSubscribe(t testing.TB, h *channelHarness, label string) *models.UserDataStreamSubscribeResponse {
	t.Helper()
	id := time.Now().UnixNano()
	req := &models.UserDataStreamSubscribeRequest{Id: models.NewMessageIDInt64(id)}
	req.Method = "userDataStream.subscribe"
	logRequestOnFailure(t, label+".request", req)
	handler, ch := newResponseHandler[models.UserDataStreamSubscribeResponse](t, label)
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.UserDataStreamSubscribe(ctx, req, handler); err != nil {
		t.Fatalf("%s call failed: %v", label, err)
	}
	resp := awaitResponse(t, ch, label)
	logRequestOnFailure(t, label+".response", resp)
	if got, ok := resp.Id.ValInt64(); !ok || got != id {
		t.Errorf("%s id mismatch: want %d got %d (ok=%v)", label, id, got, ok)
	}
	if resp.Status != 200 {
		t.Errorf("%s status: want 200 got %d", label, resp.Status)
	}
	return resp
}

func userStreamSubscribeSignature(t testing.TB, h *channelHarness, label string) *models.UserDataStreamSubscribeSignatureResponse {
	t.Helper()
	if h.Signer == nil {
		t.Fatalf("%s requires signer", label)
	}
	id := time.Now().UnixNano()
	req := &models.UserDataStreamSubscribeSignatureRequest{Id: models.NewMessageIDInt64(id)}
	req.Method = "userDataStream.subscribe.signature"
	params := map[string]interface{}{}
	signAndApply(t, h.Signer, spot.AuthTypeSigned, params, &req.Params)
	logRequestOnFailure(t, label+".request", req)
	handler, ch := newResponseHandler[models.UserDataStreamSubscribeSignatureResponse](t, label)
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.UserDataStreamSubscribeSignature(ctx, req, handler); err != nil {
		t.Fatalf("%s call failed: %v", label, err)
	}
	resp := awaitResponse(t, ch, label)
	logRequestOnFailure(t, label+".response", resp)
	if got, ok := resp.Id.ValInt64(); !ok || got != id {
		t.Errorf("%s id mismatch: want %d got %d (ok=%v)", label, id, got, ok)
	}
	if resp.Status != 200 {
		t.Errorf("%s status: want 200 got %d", label, resp.Status)
	}
	return resp
}

func userStreamUnsubscribe(t testing.TB, h *channelHarness, label string, subscriptionID int64) *models.UserDataStreamUnsubscribeResponse {
	t.Helper()
	id := time.Now().UnixNano()
	req := &models.UserDataStreamUnsubscribeRequest{Id: models.NewMessageIDInt64(id)}
	req.Method = "userDataStream.unsubscribe"
	if subscriptionID != 0 {
		req.Params.SubscriptionId = subscriptionID
	}
	logRequestOnFailure(t, label+".request", req)
	handler, ch := newResponseHandler[models.UserDataStreamUnsubscribeResponse](t, label)
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.UserDataStreamUnsubscribe(ctx, req, handler); err != nil {
		t.Fatalf("%s call failed: %v", label, err)
	}
	resp := awaitResponse(t, ch, label)
	logRequestOnFailure(t, label+".response", resp)
	if got, ok := resp.Id.ValInt64(); !ok || got != id {
		t.Errorf("%s id mismatch: want %d got %d (ok=%v)", label, id, got, ok)
	}
	if resp.Status != 200 {
		t.Logf("%s status: want 200 got %d", label, resp.Status)
	}
	return resp
}

func TestFullIntegrationSuite_Spot(t *testing.T) {
	if testing.Short() {
		t.Skip("spot WS integration suite slow; skipping in short mode")
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	symbol, err := restPickSymbol(ctx)
	cancel()
	if err != nil {
		t.Fatalf("failed to discover active symbol: %v", err)
	}
	if symbol == "" {
		t.Fatalf("restPickSymbol returned empty symbol")
	}
	symbolLower := strings.ToLower(symbol)

	type errorEnvelope struct {
		Source  string
		Message *models.ErrorMessage
	}
	errorCollector := make(chan errorEnvelope, 32)
	reportError := func(source string, msg *models.ErrorMessage) {
		if msg == nil {
			return
		}
		select {
		case errorCollector <- errorEnvelope{Source: source, Message: msg}:
		default:
			// drop oldest by draining one to keep channel non-blocking
			select {
			case <-errorCollector:
			default:
			}
			errorCollector <- errorEnvelope{Source: source, Message: msg}
		}
	}

	publicHarness.Channel.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error {
		reportError(publicHarness.Config.Name, msg)
		return nil
	})
	t.Cleanup(func() { publicHarness.Channel.UnregisterErrorMessage() })

	if hmacHarness != nil {
		hmacHarness.Channel.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error {
			reportError(hmacHarness.Config.Name, msg)
			return nil
		})
		t.Cleanup(func() { hmacHarness.Channel.UnregisterErrorMessage() })
	}
	if rsaHarness != nil {
		rsaHarness.Channel.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error {
			reportError(rsaHarness.Config.Name, msg)
			return nil
		})
		t.Cleanup(func() { rsaHarness.Channel.UnregisterErrorMessage() })
	}
	if edHarness != nil {
		edHarness.Channel.HandleErrorMessage(func(ctx context.Context, msg *models.ErrorMessage) error {
			reportError(edHarness.Config.Name, msg)
			return nil
		})
		t.Cleanup(func() { edHarness.Channel.UnregisterErrorMessage() })
	}

	t.Run("PublicRequests", func(t *testing.T) {
		runPublicRequests(t, publicHarness, symbol, symbolLower)
	})

	t.Run("UserDataRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil {
			t.Skip("HMAC credentials not configured")
		}
		runUserDataRequests(t, hmacHarness, symbol)
	})

	t.Run("TradingRequests_HMAC", func(t *testing.T) {
		if hmacHarness == nil {
			t.Skip("HMAC credentials not configured")
		}
		runTradingRequests(t, hmacHarness, symbol)
	})

	t.Run("SessionRequests_Ed25519", func(t *testing.T) {
		if edHarness == nil {
			t.Skip("Ed25519 credentials not configured")
		}
		runSessionRequests(t, edHarness)
	})

	t.Run("UserStreamRequests", func(t *testing.T) {
		var userStreamHarness *channelHarness
		switch {
		case edHarness != nil && edHarness.Config.supports(spot.AuthTypeSigned):
			userStreamHarness = edHarness
		case rsaHarness != nil && rsaHarness.Config.supports(spot.AuthTypeSigned):
			userStreamHarness = rsaHarness
		case hmacHarness != nil && hmacHarness.Config.supports(spot.AuthTypeUserStream):
			userStreamHarness = hmacHarness
		}
		if userStreamHarness == nil {
			t.Skip("no credentials available for user stream requests")
		}
		runUserStreamRequests(t, userStreamHarness)
	})

	t.Run("KlineResponseHandlers", func(t *testing.T) {
		runKlineResponseHandlers(t, publicHarness, symbol, symbolLower)
	})

	t.Run("UserDataEventHandlers", func(t *testing.T) {
		if hmacHarness == nil {
			t.Skip("HMAC credentials not configured")
		}
		runUserDataEventHandlers(t, hmacHarness, symbol)
	})

	// Log captured error messages for debugging context.
	close(errorCollector)
	var errorsSeen []errorEnvelope
	for env := range errorCollector {
		errorsSeen = append(errorsSeen, env)
	}
	if len(errorsSeen) > 0 {
		for _, env := range errorsSeen {
			data, _ := json.Marshal(env.Message)
			t.Logf("errorMessage[%s]: %s", env.Source, string(data))
		}
		t.Fatalf("received %d error message(s) from server during test run", len(errorsSeen))
	}
}

func runPublicRequests(t *testing.T, h *channelHarness, symbolUpper string, symbolLower string) {
	t.Helper()

	t.Run("Ping", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.PingRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.PingResponse](t, "ping.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Ping(ctx, req, handler); err != nil {
			t.Fatalf("ping call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "ping")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("ping id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("ping status: want 200 got %d", resp.Status)
		}
	})

	t.Run("ServerTime", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TimeRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.TimeResponse](t, "time.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Time(ctx, req, handler); err != nil {
			t.Fatalf("time call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "time")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("time id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("time status: want 200 got %d", resp.Status)
		}
		if resp.Result.ServerTime <= 0 {
			t.Errorf("time result missing server time: %+v", resp.Result)
		} else {
			assertRecentMs(t, resp.Result.ServerTime, 5*time.Minute, "serverTime")
		}
	})

	t.Run("ExchangeInfo", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.ExchangeInfoRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.ExchangeInfoResponse](t, "exchangeInfo.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.ExchangeInfo(ctx, req, handler); err != nil {
			t.Fatalf("exchangeInfo call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "exchangeInfo")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("exchangeInfo id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("exchangeInfo status: want 200 got %d", resp.Status)
		}
		if len(resp.Result.Symbols) == 0 {
			t.Errorf("exchangeInfo returned no symbols")
		} else {
			found := false
			for _, sym := range resp.Result.Symbols {
				if strings.EqualFold(sym.Symbol, symbolUpper) {
					found = true
					break
				}
			}
			if !found {
				t.Logf("exchangeInfo did not include %s symbol", symbolUpper)
			}
		}
	})

	t.Run("AvgPrice", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.AvgPriceRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.AvgPriceResponse](t, "avgPrice.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AvgPrice(ctx, req, handler); err != nil {
			t.Fatalf("avgPrice call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "avgPrice")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("avgPrice id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("avgPrice status: want 200 got %d", resp.Status)
		}
		assertRecentMs(t, resp.Result.CloseTime, 30*time.Minute, "avgPrice.closeTime")
		assertNonEmpty(t, resp.Result.Price, "avgPrice.price")
		tryParseFloat(t, resp.Result.Price, "avgPrice.price")
	})

	t.Run("Depth", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.DepthRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		req.Params.Limit = 20
		handler, ch := newResponseHandler[models.DepthResponse](t, "depth.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Depth(ctx, req, handler); err != nil {
			t.Fatalf("depth call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "depth")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("depth id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("depth status: want 200 got %d", resp.Status)
		}
		if resp.Result.LastUpdateId <= 0 {
			t.Errorf("depth lastUpdateId <= 0")
		}
		if len(resp.Result.Bids) == 0 || len(resp.Result.Asks) == 0 {
			t.Errorf("depth bids/asks empty (bids=%d asks=%d)", len(resp.Result.Bids), len(resp.Result.Asks))
		}
	})

	t.Run("Ticker24hr", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.Ticker24hrRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.Ticker24hrResponse](t, "ticker24hr.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Ticker24hr(ctx, req, handler); err != nil {
			t.Fatalf("ticker24hr call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "ticker24hr")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("ticker24hr id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("ticker24hr status: want 200 got %d", resp.Status)
		}
		assertNonEmpty(t, resp.Result.Symbol, "ticker24hr.symbol")
		tryParseFloat(t, resp.Result.LastPrice, "ticker24hr.lastPrice")
		assertRecentMs(t, resp.Result.CloseTime, 24*time.Hour, "ticker24hr.closeTime")
	})

	t.Run("Ticker", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TickerRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.TickerResponse](t, "ticker.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Ticker(ctx, req, handler); err != nil {
			t.Fatalf("ticker call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "ticker")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("ticker id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("ticker status: want 200 got %d", resp.Status)
		}
		assertNonEmpty(t, resp.Result.Symbol, "ticker.symbol")
		tryParseFloat(t, resp.Result.LastPrice, "ticker.lastPrice")
		assertRecentMs(t, resp.Result.CloseTime, 24*time.Hour, "ticker.closeTime")
	})

	t.Run("TickerPrice", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TickerPriceRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.TickerPriceResponse](t, "tickerPrice.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TickerPrice(ctx, req, handler); err != nil {
			t.Fatalf("tickerPrice call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tickerPrice")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tickerPrice id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tickerPrice status: want 200 got %d", resp.Status)
		}
		assertNonEmpty(t, resp.Result.Symbol, "tickerPrice.symbol")
		tryParseFloat(t, resp.Result.Price, "tickerPrice.price")
	})

	t.Run("TickerBook", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TickerBookRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.TickerBookResponse](t, "tickerBook.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TickerBook(ctx, req, handler); err != nil {
			t.Fatalf("tickerBook call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tickerBook")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tickerBook id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tickerBook status: want 200 got %d", resp.Status)
		}
		assertNonEmpty(t, resp.Result.Symbol, "tickerBook.symbol")
		tryParseFloat(t, resp.Result.BidPrice, "tickerBook.bidPrice")
		tryParseFloat(t, resp.Result.AskPrice, "tickerBook.askPrice")
	})

	t.Run("TickerTradingDay", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TickerTradingDayRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		handler, ch := newResponseHandler[models.TickerTradingDayResponse](t, "tickerTradingDay.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TickerTradingDay(ctx, req, handler); err != nil {
			t.Fatalf("tickerTradingDay call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tickerTradingDay")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tickerTradingDay id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tickerTradingDay status: want 200 got %d", resp.Status)
		}
		assertNonEmpty(t, resp.Result.Symbol, "tickerTradingDay.symbol")
		tryParseFloat(t, resp.Result.LastPrice, "tickerTradingDay.lastPrice")
		if resp.Result.CloseTime > 0 {
			assertRecentMs(t, resp.Result.CloseTime, 24*time.Hour, "tickerTradingDay.closeTime")
		}
	})

	t.Run("TradesAggregate", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TradesAggregateRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		req.Params.Limit = 5
		handler, ch := newResponseHandler[models.TradesAggregateResponse](t, "tradesAggregate.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TradesAggregate(ctx, req, handler); err != nil {
			t.Fatalf("tradesAggregate call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tradesAggregate")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tradesAggregate id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tradesAggregate status: want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Errorf("tradesAggregate result empty")
		} else {
			item := resp.Result[0]
			eventTime := int64(item.T)
			if eventTime > 0 {
				assertRecentMs(t, eventTime, 24*time.Hour, "aggregateTrade.time")
			}
			tryParseFloat(t, item.P, "aggregateTrade.price")
		}
	})

	t.Run("TradesHistorical", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TradesHistoricalRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		req.Params.Limit = 5
		handler, ch := newResponseHandler[models.TradesHistoricalResponse](t, "tradesHistorical.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TradesHistorical(ctx, req, handler); err != nil {
			t.Fatalf("tradesHistorical call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tradesHistorical")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tradesHistorical id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tradesHistorical status: want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Errorf("tradesHistorical result empty")
		} else {
			item := resp.Result[0]
			assertRecentMs(t, item.Time, 48*time.Hour, "historicalTrade.time")
			tryParseFloat(t, item.Price, "historicalTrade.price")
		}
	})

	t.Run("TradesRecent", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.TradesRecentRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbolUpper
		req.Params.Limit = 5
		handler, ch := newResponseHandler[models.TradesRecentResponse](t, "tradesRecent.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.TradesRecent(ctx, req, handler); err != nil {
			t.Fatalf("tradesRecent call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "tradesRecent")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("tradesRecent id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("tradesRecent status: want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Errorf("tradesRecent result empty")
		} else {
			item := resp.Result[0]
			assertRecentMs(t, item.Time, 5*time.Minute, "recentTrade.time")
			tryParseFloat(t, item.Price, "recentTrade.price")
		}
	})
}

func runUserDataRequests(t *testing.T, h *channelHarness, symbol string) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer not available for user data requests")
	}
	var lastOrderID int64

	t.Run("AccountStatus", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.AccountStatusRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.OmitZeroBalances = true
		params := map[string]interface{}{
			"omitZeroBalances": req.Params.OmitZeroBalances,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.status.request", req)
		handler, ch := newResponseHandler[models.AccountStatusResponse](t, "account.status")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountStatus(ctx, req, handler); err != nil {
			t.Fatalf("accountStatus call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.status")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("accountStatus id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("accountStatus status: want 200 got %d", resp.Status)
		}
	})

	t.Run("AccountRateLimitsOrders", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.AccountRateLimitsOrdersRequest{Id: models.NewMessageIDInt64(id)}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "account.rateLimits.orders.request", req)
		handler, ch := newResponseHandler[models.AccountRateLimitsOrdersResponse](t, "account.rateLimits.orders")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AccountRateLimitsOrders(ctx, req, handler); err != nil {
			t.Fatalf("accountRateLimitsOrders call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "account.rateLimits.orders")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("accountRateLimitsOrders id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("accountRateLimitsOrders status: want 200 got %d", resp.Status)
		}
	})

	t.Run("AllOrders", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.AllOrdersRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		req.Params.Limit = 5
		params := map[string]interface{}{
			"symbol": req.Params.Symbol,
			"limit":  req.Params.Limit,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "allOrders.request", req)
		handler, ch := newResponseHandler[models.AllOrdersResponse](t, "allOrders")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AllOrders(ctx, req, handler); err != nil {
			t.Fatalf("allOrders call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "allOrders")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("allOrders id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("allOrders status: want 200 got %d", resp.Status)
		}
		for _, ord := range resp.Result {
			if ord.OrderId != 0 {
				lastOrderID = ord.OrderId
				break
			}
		}
	})

	t.Run("OpenOrdersStatus", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.OpenOrdersStatusRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		params := map[string]interface{}{
			"symbol": req.Params.Symbol,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "openOrders.status.request", req)
		handler, ch := newResponseHandler[models.OpenOrdersStatusResponse](t, "openOrders.status")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OpenOrdersStatus(ctx, req, handler); err != nil {
			t.Fatalf("openOrdersStatus call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "openOrders.status")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("openOrdersStatus id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("openOrdersStatus status: want 200 got %d", resp.Status)
		}
	})

	t.Run("AllOrderLists", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.AllOrderListsRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Limit = 5
		params := map[string]interface{}{
			"limit": req.Params.Limit,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "allOrderLists.request", req)
		handler, ch := newResponseHandler[models.AllOrderListsResponse](t, "allOrderLists")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.AllOrderLists(ctx, req, handler); err != nil {
			t.Fatalf("allOrderLists call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "allOrderLists")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("allOrderLists id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("allOrderLists status: want 200 got %d", resp.Status)
		}
	})

	t.Run("MyTrades", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.MyTradesRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		req.Params.Limit = 5
		params := map[string]interface{}{
			"symbol": req.Params.Symbol,
			"limit":  req.Params.Limit,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "myTrades.request", req)
		handler, ch := newResponseHandler[models.MyTradesResponse](t, "myTrades")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.MyTrades(ctx, req, handler); err != nil {
			t.Fatalf("myTrades call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "myTrades")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("myTrades id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("myTrades status: want 200 got %d", resp.Status)
		}
	})

	t.Run("MyAllocations", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.MyAllocationsRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		req.Params.Limit = 5
		req.Params.Timestamp = time.Now().UnixMilli()
		params := map[string]interface{}{
			"symbol":    req.Params.Symbol,
			"limit":     req.Params.Limit,
			"timestamp": req.Params.Timestamp,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "myAllocations.request", req)
		handler, ch := newResponseHandler[models.MyAllocationsResponse](t, "myAllocations")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.MyAllocations(ctx, req, handler); err != nil {
			t.Fatalf("myAllocations call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "myAllocations")
		logRequestOnFailure(t, "myAllocations.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("myAllocations id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("myAllocations status: want 200 got %d", resp.Status)
		}
	})

	t.Run("MyPreventedMatches", func(t *testing.T) {
		if lastOrderID == 0 {
			t.Skip("no historical order id available for prevented matches query")
		}
		id := time.Now().UnixNano()
		req := &models.MyPreventedMatchesRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		req.Params.OrderId = lastOrderID
		req.Params.Limit = 5
		req.Params.Timestamp = time.Now().UnixMilli()
		params := map[string]interface{}{
			"symbol":    req.Params.Symbol,
			"limit":     req.Params.Limit,
			"timestamp": req.Params.Timestamp,
			"orderId":   req.Params.OrderId,
		}
		signAndApply(t, h.Signer, spot.AuthTypeUserData, params, &req.Params)
		logRequestOnFailure(t, "myPreventedMatches.request", req)
		handler, ch := newResponseHandler[models.MyPreventedMatchesResponse](t, "myPreventedMatches")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.MyPreventedMatches(ctx, req, handler); err != nil {
			t.Fatalf("myPreventedMatches call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "myPreventedMatches")
		logRequestOnFailure(t, "myPreventedMatches.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("myPreventedMatches id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("myPreventedMatches status: want 200 got %d", resp.Status)
		}
	})

	t.Run("OrderAmendments", func(t *testing.T) {
		t.Skip("orderAmendments requires an existing order id; coverage pending trade fixtures")
	})
}

func runTradingRequests(t *testing.T, h *channelHarness, symbol string) {
	t.Helper()
	_ = h
	_ = symbol
	t.Skip("trading request tests not yet implemented")
}

func runSessionRequests(t *testing.T, h *channelHarness) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer not available for session requests")
	}

	var loggedOn bool
	var logonResp *models.SessionLogonResponse

	t.Run("SessionLogon", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.SessionLogonRequest{Id: models.NewMessageIDInt64(id)}
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, spot.AuthTypeSigned, params, &req.Params)
		handler, ch := newResponseHandler[models.SessionLogonResponse](t, "session.logon")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SessionLogon(ctx, req, handler); err != nil {
			t.Fatalf("sessionLogon call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.logon")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sessionLogon id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("sessionLogon status: want 200 got %d", resp.Status)
		}
		if resp.Result.ApiKey == "" {
			t.Errorf("sessionLogon result missing apiKey")
		}
		assertRecentMs(t, resp.Result.ServerTime, 1*time.Minute, "sessionLogon.serverTime")
		loggedOn = true
		logonResp = resp
	})

	t.Run("SessionStatus", func(t *testing.T) {
		if !loggedOn {
			t.Skip("session not logged in")
		}
		id := time.Now().UnixNano()
		req := &models.SessionStatusRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.SessionStatusResponse](t, "session.status")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SessionStatus(ctx, req, handler); err != nil {
			t.Fatalf("sessionStatus call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.status")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sessionStatus id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("sessionStatus status: want 200 got %d", resp.Status)
		}
		if logonResp != nil && resp.Result.ApiKey != logonResp.Result.ApiKey {
			t.Errorf("sessionStatus apiKey mismatch: want %s got %s", logonResp.Result.ApiKey, resp.Result.ApiKey)
		}
	})

	t.Run("SessionSubscriptions", func(t *testing.T) {
		if !loggedOn {
			t.Skip("session not logged in")
		}
		id := time.Now().UnixNano()
		req := &models.SessionSubscriptionsRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.SessionSubscriptionsResponse](t, "session.subscriptions")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SessionSubscriptions(ctx, req, handler); err != nil {
			t.Fatalf("sessionSubscriptions call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.subscriptions")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sessionSubscriptions id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("sessionSubscriptions status: want 200 got %d", resp.Status)
		}
	})

	t.Run("SessionLogout", func(t *testing.T) {
		if !loggedOn {
			t.Skip("session not logged in")
		}
		id := time.Now().UnixNano()
		req := &models.SessionLogoutRequest{Id: models.NewMessageIDInt64(id)}
		handler, ch := newResponseHandler[models.SessionLogoutResponse](t, "session.logout")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SessionLogout(ctx, req, handler); err != nil {
			t.Fatalf("sessionLogout call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "session.logout")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sessionLogout id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("sessionLogout status: want 200 got %d", resp.Status)
		}
		loggedOn = false
	})
}

func runUserStreamRequests(t *testing.T, h *channelHarness) {
	t.Helper()
	ensureSessionLoggedIn(t, h)
	t.Run("UserDataStreamSubscribe", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.UserDataStreamSubscribeRequest{Id: models.NewMessageIDInt64(id)}
		req.Method = "userDataStream.subscribe"
		logRequestOnFailure(t, "userDataStream.subscribe.request", req)
		handler, ch := newResponseHandler[models.UserDataStreamSubscribeResponse](t, "userDataStream.subscribe")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.UserDataStreamSubscribe(ctx, req, handler); err != nil {
			t.Fatalf("subscribe call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "userDataStream.subscribe")
		logRequestOnFailure(t, "userDataStream.subscribe.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("subscribe id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("subscribe status: want 200 got %d", resp.Status)
		}
		if resp.Result.SubscriptionId == 0 {
			t.Errorf("subscribe returned zero subscription id")
		}
	})

	t.Run("UserDataStreamSubscribeSignature", func(t *testing.T) {
		if h.Signer == nil {
			t.Skip("signer not available for signature subscription")
		}
		id := time.Now().UnixNano()
		req := &models.UserDataStreamSubscribeSignatureRequest{Id: models.NewMessageIDInt64(id)}
		req.Method = "userDataStream.subscribe.signature"
		params := map[string]interface{}{}
		signAndApply(t, h.Signer, spot.AuthTypeSigned, params, &req.Params)
		logRequestOnFailure(t, "userDataStream.subscribe.signature.request", req)

		handler, ch := newResponseHandler[models.UserDataStreamSubscribeSignatureResponse](t, "userDataStream.subscribe.signature")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.UserDataStreamSubscribeSignature(ctx, req, handler); err != nil {
			t.Fatalf("subscribe.signature call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "userDataStream.subscribe.signature")
		logRequestOnFailure(t, "userDataStream.subscribe.signature.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("subscribe.signature id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("subscribe.signature status: want 200 got %d", resp.Status)
		}
		if resp.Result.SubscriptionId == 0 {
			t.Errorf("subscribe.signature returned zero subscription id")
		}
	})

	t.Run("UserDataStreamUnsubscribe", func(t *testing.T) {
		// First acquire a subscription id we can close.
		id := time.Now().UnixNano()
		reqSub := &models.UserDataStreamSubscribeRequest{Id: models.NewMessageIDInt64(id)}
		reqSub.Method = "userDataStream.subscribe"
		logRequestOnFailure(t, "userDataStream.subscribe.setup.request", reqSub)
		subHandler, subCh := newResponseHandler[models.UserDataStreamSubscribeResponse](t, "userDataStream.subscribe(for-unsubscribe)")
		ctx, cancel := requestContext()
		throttleWS()
		if err := h.Channel.UserDataStreamSubscribe(ctx, reqSub, subHandler); err != nil {
			cancel()
			t.Fatalf("subscribe (setup) failed: %v", err)
		}
		resp := awaitResponse(t, subCh, "userDataStream.subscribe(for-unsubscribe)")
		logRequestOnFailure(t, "userDataStream.subscribe(for-unsubscribe).response", resp)
		cancel()
		subscriptionID := resp.Result.SubscriptionId
		if subscriptionID == 0 {
			t.Fatalf("setup subscription returned zero id")
		}

		idUnsub := time.Now().UnixNano()
		req := &models.UserDataStreamUnsubscribeRequest{Id: models.NewMessageIDInt64(idUnsub)}
		req.Method = "userDataStream.unsubscribe"
		req.Params.SubscriptionId = subscriptionID
		logRequestOnFailure(t, "userDataStream.unsubscribe.request", req)
		handler, ch := newResponseHandler[models.UserDataStreamUnsubscribeResponse](t, "userDataStream.unsubscribe")
		ctxUnsub, cancelUnsub := requestContext()
		defer cancelUnsub()
		throttleWS()
		if err := h.Channel.UserDataStreamUnsubscribe(ctxUnsub, req, handler); err != nil {
			t.Fatalf("unsubscribe call failed: %v", err)
		}
		respUnsub := awaitResponse(t, ch, "userDataStream.unsubscribe")
		logRequestOnFailure(t, "userDataStream.unsubscribe.response", respUnsub)
		if got, ok := respUnsub.Id.ValInt64(); !ok || got != idUnsub {
			t.Errorf("unsubscribe id mismatch: want %d got %d (ok=%v)", idUnsub, got, ok)
		}
		if respUnsub.Status != 200 {
			t.Errorf("unsubscribe status: want 200 got %d", respUnsub.Status)
		}
	})
}

func runKlineResponseHandlers(t *testing.T, h *channelHarness, symbolUpper string, symbolLower string) {
	t.Helper()
	_ = symbolLower // reserved for potential combined stream tests

	t.Run("Klines", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.KlinesRequest{Id: models.NewMessageIDInt64(id)}
		req.Method = "klines"
		req.Params.Symbol = symbolUpper
		req.Params.Interval = "1m"
		req.Params.Limit = 5
		logRequestOnFailure(t, "klines.request", req)

		handler, ch := newResponseHandler[models.KlinesResponse](t, "klines.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Klines(ctx, req, handler); err != nil {
			t.Fatalf("klines call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "klines")
		logRequestOnFailure(t, "klines.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("klines id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("klines status: want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Errorf("klines result empty")
		}
	})

	t.Run("UiKlines", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.UiKlinesRequest{Id: models.NewMessageIDInt64(id)}
		req.Method = "uiKlines"
		req.Params.Symbol = symbolUpper
		req.Params.Interval = "1m"
		req.Params.Limit = 3
		logRequestOnFailure(t, "uiKlines.request", req)

		handler, ch := newResponseHandler[models.UiKlinesResponse](t, "uiKlines.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.UiKlines(ctx, req, handler); err != nil {
			t.Fatalf("uiKlines call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "uiKlines")
		logRequestOnFailure(t, "uiKlines.response", resp)
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("uiKlines id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Errorf("uiKlines status: want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Errorf("uiKlines result empty")
		}
	})
}

func runUserDataEventHandlers(t *testing.T, h *channelHarness, symbol string) {
	t.Helper()
	_ = h
	_ = symbol
	t.Skip("user data event handler tests not yet implemented")
}
