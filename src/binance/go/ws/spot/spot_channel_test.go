package wstest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
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

const (
	subscriptionPollAttempts = 5
	subscriptionPollDelay    = 400 * time.Millisecond
)

func (c *unhandledCatcher) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("unhandled message:")) {
		c.mu.Lock()
		c.matches = append(c.matches, strings.TrimSpace(string(p)))
		c.mu.Unlock()
	}
	return len(p), nil
}

func newUnhandledCatcher() *unhandledCatcher { return &unhandledCatcher{} }

type spotUserDataEventRecorder struct {
	mu           sync.Mutex
	terminated   []*models.EventStreamTerminatedEvent
	account      []*models.AccountUpdateEvent
	balance      []*models.BalanceUpdateEvent
	order        []*models.OrderUpdateEvent
	listStatus   []*models.ListStatusEvent
	externalLock []*models.ExternalLockUpdateEvent
}

func newSpotUserDataEventRecorder() *spotUserDataEventRecorder {
	return &spotUserDataEventRecorder{}
}

func (r *spotUserDataEventRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.terminated = nil
	r.account = nil
	r.balance = nil
	r.order = nil
	r.listStatus = nil
	r.externalLock = nil
}

func (r *spotUserDataEventRecorder) addTerminated(evt *models.EventStreamTerminatedEvent) {
	r.mu.Lock()
	r.terminated = append(r.terminated, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) addAccount(evt *models.AccountUpdateEvent) {
	r.mu.Lock()
	r.account = append(r.account, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) addBalance(evt *models.BalanceUpdateEvent) {
	r.mu.Lock()
	r.balance = append(r.balance, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) addOrder(evt *models.OrderUpdateEvent) {
	r.mu.Lock()
	r.order = append(r.order, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) addListStatus(evt *models.ListStatusEvent) {
	r.mu.Lock()
	r.listStatus = append(r.listStatus, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) addExternalLock(evt *models.ExternalLockUpdateEvent) {
	r.mu.Lock()
	r.externalLock = append(r.externalLock, evt)
	r.mu.Unlock()
}

func (r *spotUserDataEventRecorder) count(kind string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch kind {
	case "eventStreamTerminated":
		return len(r.terminated)
	case "accountUpdate":
		return len(r.account)
	case "balanceUpdate":
		return len(r.balance)
	case "orderUpdate":
		return len(r.order)
	case "listStatus":
		return len(r.listStatus)
	case "externalLockUpdate":
		return len(r.externalLock)
	default:
		return 0
	}
}

type exchangeInfoSummary struct {
	sorSymbols []string
}

func (s *exchangeInfoSummary) recordSorSymbols(symbols []string) {
	if s == nil || len(symbols) == 0 {
		return
	}
	for _, sym := range symbols {
		candidate := strings.ToUpper(strings.TrimSpace(sym))
		if candidate == "" {
			continue
		}
		found := false
		for _, existing := range s.sorSymbols {
			if existing == candidate {
				found = true
				break
			}
		}
		if !found {
			s.sorSymbols = append(s.sorSymbols, candidate)
		}
	}
}

func (s *exchangeInfoSummary) SorSymbols() []string {
	if s == nil || len(s.sorSymbols) == 0 {
		return nil
	}
	out := make([]string, len(s.sorSymbols))
	copy(out, s.sorSymbols)
	return out
}

func sanitizeClientOrderPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range prefix {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			// skip unsupported characters
		}
	}
	return b.String()
}

func newClientOrderID(prefix string, seed int64) string {
	clean := sanitizeClientOrderPrefix(prefix)
	if clean == "" {
		clean = "wstest"
	}
	if len(clean) > 12 {
		clean = clean[:12]
	}
	if seed < 0 {
		seed = -seed
	}
	suffix := strings.ToUpper(strconv.FormatInt(seed, 36))
	if len(suffix) < 6 {
		suffix = fmt.Sprintf("%06s", suffix)
	}
	return fmt.Sprintf("%s-%s", clean, suffix)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *spotUserDataEventRecorder) waitFor(kind string, min int, timeout time.Duration) bool {
	if min <= 0 {
		return true
	}
	deadline := time.Now().Add(timeout)
	for {
		if r.count(kind) >= min {
			return true
		}
		if time.Now().After(deadline) {
			return r.count(kind) >= min
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (r *spotUserDataEventRecorder) lastOrderUpdate() *models.OrderUpdateEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.order) == 0 {
		return nil
	}
	return r.order[len(r.order)-1]
}

func (r *spotUserDataEventRecorder) lastAccountUpdate() *models.AccountUpdateEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.account) == 0 {
		return nil
	}
	return r.account[len(r.account)-1]
}

func (r *spotUserDataEventRecorder) lastBalanceUpdate() *models.BalanceUpdateEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.balance) == 0 {
		return nil
	}
	return r.balance[len(r.balance)-1]
}

func (r *spotUserDataEventRecorder) lastListStatus() *models.ListStatusEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.listStatus) == 0 {
		return nil
	}
	return r.listStatus[len(r.listStatus)-1]
}

func (r *spotUserDataEventRecorder) lastTerminated() *models.EventStreamTerminatedEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.terminated) == 0 {
		return nil
	}
	return r.terminated[len(r.terminated)-1]
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

func sessionSubscriptionsResponse(t testing.TB, h *channelHarness, label string) *models.SessionSubscriptionsResponse {
	t.Helper()
	id := time.Now().UnixNano()
	req := &models.SessionSubscriptionsRequest{Id: models.NewMessageIDInt64(id)}
	logRequestOnFailure(t, label+".request", req)
	handler, ch := newResponseHandler[models.SessionSubscriptionsResponse](t, label)
	ctx, cancel := requestContext()
	defer cancel()
	throttleWS()
	if err := h.Channel.SessionSubscriptions(ctx, req, handler); err != nil {
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

func resolveUserStreamSubscriptionID(t testing.TB, h *channelHarness, explicit int64, label string) int64 {
	t.Helper()
	if explicit != 0 {
		return explicit
	}
	labelBase := label
	for attempt := 0; attempt < subscriptionPollAttempts; attempt++ {
		labelAttempt := labelBase
		if attempt > 0 {
			labelAttempt = fmt.Sprintf("%s.retry%d", labelBase, attempt+1)
		}
		resp := sessionSubscriptionsResponse(t, h, labelAttempt)
		for _, entry := range resp.Result {
			if entry.SubscriptionId != 0 {
				return entry.SubscriptionId
			}
		}
		if attempt < subscriptionPollAttempts-1 {
			time.Sleep(subscriptionPollDelay)
		}
	}
	return 0
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

func registerUserStreamEventHandlers(t testing.TB, h *channelHarness) (*spotUserDataEventRecorder, chan *models.EventStreamTerminatedEvent) {
	t.Helper()
	rec := newSpotUserDataEventRecorder()
	terminatedCh := make(chan *models.EventStreamTerminatedEvent, 8)

	h.Channel.HandleEventStreamTerminatedEvent(func(ctx context.Context, evt *models.EventStreamTerminatedEvent) error {
		if evt == nil {
			t.Errorf("eventStreamTerminated handler received nil event")
			return nil
		}
		if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "eventStreamTerminated") {
			t.Errorf("eventStreamTerminated handler received wrong event type %q", evt.Event.EventType)
		}
		rec.addTerminated(evt)
		select {
		case terminatedCh <- evt:
		default:
		}
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterEventStreamTerminatedEvent() })

	h.Channel.HandleAccountUpdateEvent(func(ctx context.Context, evt *models.AccountUpdateEvent) error {
		if evt == nil {
			t.Errorf("accountUpdate handler received nil event")
			return nil
		}
		rec.addAccount(evt)
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterAccountUpdateEvent() })

	h.Channel.HandleBalanceUpdateEvent(func(ctx context.Context, evt *models.BalanceUpdateEvent) error {
		if evt == nil {
			t.Errorf("balanceUpdate handler received nil event")
			return nil
		}
		rec.addBalance(evt)
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterBalanceUpdateEvent() })

	h.Channel.HandleOrderUpdateEvent(func(ctx context.Context, evt *models.OrderUpdateEvent) error {
		if evt == nil {
			t.Errorf("orderUpdate handler received nil event")
			return nil
		}
		rec.addOrder(evt)
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterOrderUpdateEvent() })

	h.Channel.HandleListStatusEvent(func(ctx context.Context, evt *models.ListStatusEvent) error {
		if evt == nil {
			t.Errorf("listStatus handler received nil event")
			return nil
		}
		rec.addListStatus(evt)
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterListStatusEvent() })

	h.Channel.HandleExternalLockUpdateEvent(func(ctx context.Context, evt *models.ExternalLockUpdateEvent) error {
		if evt == nil {
			t.Errorf("externalLockUpdate handler received nil event")
			return nil
		}
		rec.addExternalLock(evt)
		return nil
	})
	t.Cleanup(func() { h.Channel.UnregisterExternalLockUpdateEvent() })

	return rec, terminatedCh
}

func drainEventStreamTerminated(ch chan *models.EventStreamTerminatedEvent) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func awaitEventStreamTerminated(t testing.TB, ch chan *models.EventStreamTerminatedEvent, label string) *models.EventStreamTerminatedEvent {
	t.Helper()
	select {
	case evt := <-ch:
		if evt == nil {
			t.Fatalf("%s: received nil event stream termination", label)
		}
		if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "eventStreamTerminated") {
			t.Errorf("%s: unexpected event type %q", label, evt.Event.EventType)
		}
		if evt.Event.EventTime == 0 {
			t.Logf("%s: eventStreamTerminated missing event time", label)
		}
		return evt
	case <-time.After(eventWait()):
		t.Fatalf("%s: timeout waiting for eventStreamTerminated event", label)
		return nil
	}
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
	exchangeSnapshot := &exchangeInfoSummary{}

	t.Run("PublicRequests", func(t *testing.T) {
		runPublicRequests(t, publicHarness, symbol, symbolLower, exchangeSnapshot)
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
		runTradingRequests(t, hmacHarness, symbol, exchangeSnapshot)
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
		var eventHarness *channelHarness
		if edHarness != nil && edHarness.Config.supports(spot.AuthTypeSigned) {
			eventHarness = edHarness
		} else if rsaHarness != nil && rsaHarness.Config.supports(spot.AuthTypeSigned) {
			eventHarness = rsaHarness
		} else if hmacHarness != nil && hmacHarness.Config.supports(spot.AuthTypeSigned) {
			eventHarness = hmacHarness
		}
		if eventHarness == nil {
			t.Skip("no credentials available for signed user data event handling")
		}
		if eventHarness.Config.KeyType != spot.KeyTypeED25519 {
			t.Skipf("%s key type %s not supported for session.logon; requires Ed25519 credentials", eventHarness.Config.Name, eventHarness.Config.KeyType)
		}
		runUserDataEventHandlers(t, eventHarness, symbol)
	})

}

func runPublicRequests(t *testing.T, h *channelHarness, symbolUpper string, symbolLower string, info *exchangeInfoSummary) {
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
		if info != nil {
			for _, sor := range resp.Result.Sors {
				info.recordSorSymbols(sor.Symbols)
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

	t.Run("DepthInvalidSymbolError", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.DepthRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = "INVALIDPAIR"
		req.Params.Limit = 5
		handler, errCh := newErrorResponseHandler[models.DepthResponse](t, "depth.error")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.Depth(ctx, req, handler); err != nil {
			t.Fatalf("depth call failed: %v", err)
		}
		err := awaitError(t, errCh, "depth.error")
		t.Logf("depth invalid symbol returned error: %v", err)
		var apiErr *models.ErrorMessage
		if !errors.As(err, &apiErr) {
			t.Fatalf("depth error not ErrorMessage: %T %v", err, err)
		}
		if apiErr.Status == 0 {
			t.Errorf("depth error status unset")
		}
		if apiErr.ErrorPayload.Code == 0 {
			t.Errorf("depth error code unset")
		}
		if apiErr.ErrorPayload.Msg == "" {
			t.Errorf("depth error message empty")
		} else if !strings.Contains(strings.ToLower(apiErr.ErrorPayload.Msg), "symbol") {
			t.Errorf("depth error message does not mention symbol: %q", apiErr.ErrorPayload.Msg)
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
		if !h.Config.supports(spot.AuthTypeTrade) {
			t.Skip("credentials do not include TRADE permission for order amendments")
		}

		rc := newRESTClient()
		baseCtx := context.Background()
		priceRef, err := restTickerPrice(baseCtx, rc, symbol)
		if err != nil || priceRef <= 0 {
			t.Logf("ticker price fetch failed (%v); falling back to default price", err)
			priceRef = 100.0
		}
		targetPrice := priceRef * 0.5
		if targetPrice <= 0 {
			targetPrice = 50.0
		}
		qtyStr, priceStr, constraints, err := calcSpotLimitOrderParams(baseCtx, rc, symbol, targetPrice)
		if err != nil {
			t.Skipf("calcSpotLimitOrderParams failed: %v", err)
		}
		if constraints != nil {
			priceVal := parseFloat64(priceStr)
			qtyVal := parseFloat64(qtyStr)
			if priceVal > 0 && qtyVal > 0 {
				minNotional := constraints.minNotional
				if minNotional <= 0 {
					minNotional = 10.0
				}
				targetNotional := minNotional * 2
				if targetNotional < minNotional+1 {
					targetNotional = minNotional + 1
				}
				currentNotional := qtyVal * priceVal
				if currentNotional < targetNotional {
					adjustedQty := ceilToStep(targetNotional/priceVal, constraints.qtyStep, constraints.qtyDecimals)
					if adjustedQty*priceVal > 50.0 {
						adjustedQty = floorToStep(50.0/priceVal, constraints.qtyStep, constraints.qtyDecimals)
					}
					if adjustedQty > qtyVal && adjustedQty > 0 {
						qtyStr = fmt.Sprintf("%.*f", constraints.qtyDecimals, adjustedQty)
					}
				}
			}
		}
		order, err := placeSpotLimitOrder(rc, symbol, "BUY", qtyStr, priceStr)
		if err != nil {
			if errors.Is(err, errSpotCredentialsMissing) {
				t.Skip("spot REST credentials not configured for trading operations")
			}
			if isInvalidSpotKeyOrPermissionError(err) {
				t.Skipf("spot credentials lack trade permission: %v", err)
			}
			if isInsufficientSpotBalanceError(err) {
				t.Skipf("spot account insufficient to place order: %v", err)
			}
			t.Fatalf("placeSpotLimitOrder failed: %v", err)
		}
		if order == nil {
			t.Skip("placeSpotLimitOrder returned nil order")
		}
		defer func() {
			if order != nil {
				if cerr := cancelSpotOrder(rc, order); cerr != nil {
					t.Logf("cancelSpotOrder cleanup error: %v", cerr)
				}
			}
		}()
		if order.orderID == 0 && order.clientOrderID == "" {
			t.Skip("order identifiers unavailable; cannot request amendments")
		}

		newQtyStr, err := deriveAmendedQuantity(qtyStr, parseFloat64(priceStr), constraints)
		if err != nil {
			t.Skipf("deriveAmendedQuantity failed: %v", err)
		}
		if newQtyStr == qtyStr || newQtyStr == "" {
			t.Skip("unable to produce distinct amendment quantity")
		}

		amendID := time.Now().UnixNano()
		amendReq := &models.OrderAmendKeepPriorityRequest{Id: models.NewMessageIDInt64(amendID)}
		amendReq.Params.Symbol = symbol
		amendReq.Params.Timestamp = time.Now().UnixMilli()
		amendReq.Params.RecvWindow = 5000
		amendReq.Params.NewQty = newQtyStr
		params := map[string]interface{}{
			"symbol":     amendReq.Params.Symbol,
			"timestamp":  amendReq.Params.Timestamp,
			"recvWindow": amendReq.Params.RecvWindow,
			"newQty":     amendReq.Params.NewQty,
		}
		if order.orderID != 0 {
			amendReq.Params.OrderId = order.orderID
			params["orderId"] = order.orderID
		} else if order.clientOrderID != "" {
			amendReq.Params.OrigClientOrderId = order.clientOrderID
			params["origClientOrderId"] = order.clientOrderID
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &amendReq.Params)
		logRequestOnFailure(t, "orderAmendKeepPriority.request", amendReq)
		amendHandler, amendCh := newResponseHandler[models.OrderAmendKeepPriorityResponse](t, "orderAmendKeepPriority")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderAmendKeepPriority(ctx, amendReq, amendHandler); err != nil {
			t.Fatalf("orderAmendKeepPriority call failed: %v", err)
		}
		amendResp := awaitResponse(t, amendCh, "orderAmendKeepPriority")
		cancel()
		if got, ok := amendResp.Id.ValInt64(); !ok || got != amendID {
			t.Errorf("orderAmendKeepPriority id mismatch: want %d got %d (ok=%v)", amendID, got, ok)
		}
		if amendResp.Status != 200 {
			t.Fatalf("orderAmendKeepPriority status: want 200 got %d", amendResp.Status)
		}
		if amendResp.Result.AmendedOrder.Symbol == "" {
			t.Errorf("orderAmendKeepPriority amended order missing symbol")
		}
		if amendResp.Result.AmendedOrder.Qty == "" {
			t.Errorf("orderAmendKeepPriority amended order missing qty")
		}
		if amendResp.Result.AmendedOrder.Qty != newQtyStr {
			t.Errorf("orderAmendKeepPriority qty mismatch: want %s got %s", newQtyStr, amendResp.Result.AmendedOrder.Qty)
		}
		if amendResp.Result.ExecutionId == 0 {
			t.Errorf("orderAmendKeepPriority missing executionId")
		}
		if amendResp.Result.AmendedOrder.OrderId != 0 {
			order.orderID = amendResp.Result.AmendedOrder.OrderId
		}
		if order.orderID == 0 {
			t.Skip("orderId not available after amendment; cannot query order history")
		}

		var historyResp *models.OrderAmendmentsResponse
		for attempt := 0; attempt < 3; attempt++ {
			historyID := time.Now().UnixNano()
			req := &models.OrderAmendmentsRequest{Id: models.NewMessageIDInt64(historyID)}
			req.Params.Symbol = symbol
			req.Params.OrderId = order.orderID
			req.Params.Timestamp = time.Now().UnixMilli()
			req.Params.Limit = 5
			req.Params.RecvWindow = "5000"
			hParams := map[string]interface{}{
				"symbol":     req.Params.Symbol,
				"orderId":    req.Params.OrderId,
				"timestamp":  req.Params.Timestamp,
				"limit":      req.Params.Limit,
				"recvWindow": req.Params.RecvWindow,
			}
			signAndApply(t, h.Signer, spot.AuthTypeUserData, hParams, &req.Params)
			label := fmt.Sprintf("orderAmendments[%d]", attempt+1)
			logRequestOnFailure(t, label+".request", req)
			handler, ch := newResponseHandler[models.OrderAmendmentsResponse](t, label)
			ctxHist, cancelHist := requestContext()
			throttleWS()
			if err := h.Channel.OrderAmendments(ctxHist, req, handler); err != nil {
				cancelHist()
				t.Fatalf("orderAmendments call failed: %v", err)
			}
			resp := awaitResponse(t, ch, label)
			cancelHist()
			logRequestOnFailure(t, label+".response", resp)
			if got, ok := resp.Id.ValInt64(); !ok || got != historyID {
				t.Errorf("%s id mismatch: want %d got %d (ok=%v)", label, historyID, got, ok)
			}
			if resp.Status == 200 && len(resp.Result) > 0 {
				historyResp = resp
				break
			}
			time.Sleep(1500 * time.Millisecond)
		}

		if historyResp == nil || len(historyResp.Result) == 0 {
			t.Fatalf("orderAmendments returned no results to validate")
		}

		found := false
		for _, entry := range historyResp.Result {
			if entry.OrderId != 0 && order.orderID != 0 && entry.OrderId != order.orderID {
				continue
			}
			if entry.NewQty == "" {
				continue
			}
			if entry.Symbol != "" && !strings.EqualFold(entry.Symbol, symbol) {
				continue
			}
			found = true
			if entry.NewQty != newQtyStr {
				t.Errorf("orderAmendments newQty mismatch: want %s got %s", newQtyStr, entry.NewQty)
			}
			if entry.OrigQty == "" {
				t.Errorf("orderAmendments entry missing origQty")
			} else if entry.OrigQty != qtyStr {
				t.Logf("orderAmendments origQty differs from initial quantity: want %s got %s", qtyStr, entry.OrigQty)
			}
			if entry.Time == 0 {
				t.Errorf("orderAmendments entry missing time")
			} else {
				assertRecentMs(t, entry.Time, 1*time.Minute, "orderAmendments.time")
			}
			break
		}
		if !found {
			t.Fatalf("orderAmendments did not include amendment for target order")
		}
	})
}

func runTradingRequests(t *testing.T, h *channelHarness, symbol string, info *exchangeInfoSummary) {
	t.Helper()
	if h.Signer == nil {
		t.Skip("signer not available for trading requests")
	}
	if h.Config == nil || !h.Config.supports(spot.AuthTypeTrade) {
		t.Skip("credentials do not include TRADE permission")
	}

	rc := newRESTClient()
	baseCtx := context.Background()

	priceRef, err := restTickerPrice(baseCtx, rc, symbol)
	if err != nil || priceRef <= 0 {
		t.Logf("restTickerPrice failed (%v); falling back to default price", err)
		priceRef = 100.0
	}
	targetPrice := priceRef * 0.5
	if targetPrice <= 0 {
		targetPrice = 50.0
	}

	qtyStr, priceStr, constraints, err := calcSpotLimitOrderParams(baseCtx, rc, symbol, targetPrice)
	if err != nil {
		t.Skipf("calcSpotLimitOrderParams failed: %v", err)
	}
	if qtyStr == "" || priceStr == "" {
		t.Skip("unable to determine valid quantity/price for trading requests")
	}

	priceDecimals := 8
	if constraints != nil && constraints.priceDecimals > 0 {
		priceDecimals = constraints.priceDecimals
	}

	formatPrice := func(v float64) string {
		if constraints != nil {
			rounded := floorToStep(v, constraints.priceStep, constraints.priceDecimals)
			if rounded <= 0 {
				rounded = floorToStep(targetPrice, constraints.priceStep, constraints.priceDecimals)
			}
			return fmt.Sprintf("%.*f", constraints.priceDecimals, rounded)
		}
		return fmt.Sprintf("%.*f", priceDecimals, v)
	}

	sorSymbol := strings.ToUpper(symbol)
	sorQtyStr := qtyStr
	sorPriceStr := priceStr
	sorSupported := false
	sorDiscoveryErrors := make([]string, 0)

	sorSymbols, err := fetchSorSymbols(baseCtx, rc)
	if err != nil {
		t.Logf("fetchSorSymbols error: %v", err)
	}
	if len(sorSymbols) == 0 && info != nil {
		sorSymbols = info.SorSymbols()
	}
	var sorCandidates []string
	for _, sym := range sorSymbols {
		candidate := strings.ToUpper(strings.TrimSpace(sym))
		if candidate == "" {
			continue
		}
		if strings.EqualFold(candidate, symbol) {
			sorCandidates = append([]string{candidate}, sorCandidates...)
		} else {
			sorCandidates = append(sorCandidates, candidate)
		}
	}
	sorAvailable := len(sorCandidates) > 0
	seenSorCandidates := make(map[string]struct{})

	trySorCandidate := func(candidate string) bool {
		candidate = strings.ToUpper(strings.TrimSpace(candidate))
		if candidate == "" {
			return false
		}
		if _, seen := seenSorCandidates[candidate]; seen {
			return false
		}
		seenSorCandidates[candidate] = struct{}{}
		if strings.EqualFold(candidate, symbol) {
			if qtyStr == "" || priceStr == "" {
				sorDiscoveryErrors = append(sorDiscoveryErrors, fmt.Sprintf("%s: base order parameters unavailable", candidate))
				return false
			}
			sorSymbol = candidate
			sorQtyStr = qtyStr
			sorPriceStr = priceStr
			sorSupported = true
			return true
		}
		q, p, prepErr := prepareSorLimitOrder(baseCtx, rc, candidate)
		if prepErr != nil {
			sorDiscoveryErrors = append(sorDiscoveryErrors, fmt.Sprintf("%s: %v", candidate, prepErr))
			return false
		}
		if q == "" || p == "" {
			sorDiscoveryErrors = append(sorDiscoveryErrors, fmt.Sprintf("%s: derived order parameters empty", candidate))
			return false
		}
		sorSymbol = candidate
		sorQtyStr = q
		sorPriceStr = p
		sorSupported = true
		return true
	}

	for _, candidate := range sorCandidates {
		if trySorCandidate(candidate) {
			break
		}
	}

	if !sorSupported {
		if alt, err := spotPickSorSymbol(baseCtx, rc); err == nil {
			alt = strings.ToUpper(strings.TrimSpace(alt))
			if alt != "" {
				sorAvailable = true
				_ = trySorCandidate(alt)
			}
		} else if err != nil {
			sorDiscoveryErrors = append(sorDiscoveryErrors, fmt.Sprintf("spotPickSorSymbol: %v", err))
		}
	}
	if !sorSupported && spotSymbolSupportsSOR(baseCtx, rc, strings.ToUpper(symbol)) && qtyStr != "" && priceStr != "" {
		sorSymbol = strings.ToUpper(symbol)
		sorQtyStr = qtyStr
		sorPriceStr = priceStr
		sorSupported = true
		sorAvailable = true
	}

	type orderInfo struct {
		symbol        string
		orderID       int64
		clientOrderID string
		price         string
		qty           string
	}

	placeOrder := func(t *testing.T, label string, orderSymbol string, quantity string, price string) *orderInfo {
		t.Helper()
		id := time.Now().UnixNano()
		req := &models.OrderPlaceRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = orderSymbol
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = quantity
		req.Params.Price = price
		req.Params.Timestamp = time.Now().UnixMilli()
		req.Params.RecvWindow = 5000
		req.Params.NewClientOrderId = newClientOrderID(label, id)
		params := map[string]interface{}{
			"symbol":           req.Params.Symbol,
			"side":             req.Params.Side,
			"type":             req.Params.Type,
			"timeInForce":      req.Params.TimeInForce,
			"quantity":         req.Params.Quantity,
			"price":            req.Params.Price,
			"timestamp":        req.Params.Timestamp,
			"recvWindow":       req.Params.RecvWindow,
			"newClientOrderId": req.Params.NewClientOrderId,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, label+".orderPlace.request", req)
		handler, ch := newResponseHandler[models.OrderPlaceResponse](t, label+".orderPlace.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderPlace(ctx, req, handler); err != nil {
			t.Fatalf("%s: orderPlace call failed: %v", label, err)
		}
		resp := awaitResponse(t, ch, label+".orderPlace")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("%s: orderPlace id mismatch: want %d got %d (ok=%v)", label, id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("%s: orderPlace status want 200 got %d", label, resp.Status)
		}
		if resp.Result.OrderId == 0 {
			t.Fatalf("%s: orderPlace missing orderId", label)
		}
		if resp.Result.ClientOrderId == "" {
			t.Fatalf("%s: orderPlace missing clientOrderId", label)
		}
		assertRecentMs(t, resp.Result.TransactTime, 1*time.Minute, label+".orderPlace.transactTime")
		return &orderInfo{
			symbol:        orderSymbol,
			orderID:       resp.Result.OrderId,
			clientOrderID: resp.Result.ClientOrderId,
			price:         req.Params.Price,
			qty:           req.Params.Quantity,
		}
	}

	cancelOrder := func(t *testing.T, label string, info *orderInfo) *models.OrderCancelResponse {
		t.Helper()
		if info == nil || info.orderID == 0 {
			t.Fatal(label + ": order info missing for cancel")
		}
		id := time.Now().UnixNano()
		req := &models.OrderCancelRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = info.symbol
		req.Params.OrderId = info.orderID
		req.Params.Timestamp = time.Now().UnixMilli()
		req.Params.RecvWindow = 5000
		params := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"timestamp":  req.Params.Timestamp,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, label+".orderCancel.request", req)
		handler, ch := newResponseHandler[models.OrderCancelResponse](t, label+".orderCancel.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderCancel(ctx, req, handler); err != nil {
			t.Fatalf("%s: orderCancel call failed: %v", label, err)
		}
		resp := awaitResponse(t, ch, label+".orderCancel")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("%s: orderCancel id mismatch: want %d got %d (ok=%v)", label, id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("%s: orderCancel status want 200 got %d", label, resp.Status)
		}
		if resp.Result.OrderId != info.orderID {
			t.Errorf("%s: orderCancel orderId mismatch: want %d got %d", label, info.orderID, resp.Result.OrderId)
		}
		return resp
	}

	ensureOrderClosed := func(info *orderInfo) {
		if info == nil || info.orderID == 0 {
			return
		}
		if err := cancelSpotOrder(rc, &spotPlacedOrder{
			symbol:        info.symbol,
			orderID:       info.orderID,
			clientOrderID: info.clientOrderID,
		}); err != nil {
			log.Printf("cleanup cancel order %d (%s) failed: %v", info.orderID, info.symbol, err)
		}
	}

	t.Run("OrderTest", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.OrderTestRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = symbol
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = qtyStr
		req.Params.Price = priceStr
		req.Params.Timestamp = time.Now().UnixMilli()
		params := map[string]interface{}{
			"symbol":      req.Params.Symbol,
			"side":        req.Params.Side,
			"type":        req.Params.Type,
			"timeInForce": req.Params.TimeInForce,
			"quantity":    req.Params.Quantity,
			"price":       req.Params.Price,
			"timestamp":   req.Params.Timestamp,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "orderTest.request", req)
		handler, ch := newResponseHandler[models.OrderTestResponse](t, "orderTest.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderTest(ctx, req, handler); err != nil {
			t.Fatalf("orderTest call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "orderTest")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("orderTest id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("orderTest status want 200 got %d", resp.Status)
		}
	})

	t.Run("OrderPlace", func(t *testing.T) {
		info := placeOrder(t, "orderPlace", symbol, qtyStr, priceStr)
		t.Cleanup(func() { ensureOrderClosed(info) })
		if info.price != priceStr {
			t.Errorf("orderPlace price mismatch: want %s got %s", priceStr, info.price)
		}
	})

	t.Run("OrderStatus", func(t *testing.T) {
		info := placeOrder(t, "orderStatus", symbol, qtyStr, priceStr)
		t.Cleanup(func() { ensureOrderClosed(info) })

		id := time.Now().UnixNano()
		req := &models.OrderStatusRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = info.symbol
		req.Params.OrderId = info.orderID
		req.Params.RecvWindow = "5000"
		req.Params.Timestamp = time.Now().UnixMilli()
		params := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"orderId":    req.Params.OrderId,
			"recvWindow": req.Params.RecvWindow,
			"timestamp":  req.Params.Timestamp,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "orderStatus.request", req)
		handler, ch := newResponseHandler[models.OrderStatusResponse](t, "orderStatus.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderStatus(ctx, req, handler); err != nil {
			t.Fatalf("orderStatus call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "orderStatus")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("orderStatus id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("orderStatus status want 200 got %d", resp.Status)
		}
		if resp.Result.OrderId != info.orderID {
			t.Errorf("orderStatus orderId mismatch: want %d got %d", info.orderID, resp.Result.OrderId)
		}
		if !strings.EqualFold(resp.Result.Symbol, info.symbol) {
			t.Errorf("orderStatus symbol mismatch: want %s got %s", info.symbol, resp.Result.Symbol)
		}
		if resp.Result.Price == "" {
			t.Error("orderStatus missing price")
		}
	})

	t.Run("OrderCancel", func(t *testing.T) {
		info := placeOrder(t, "orderCancel", symbol, qtyStr, priceStr)
		resp := cancelOrder(t, "orderCancel", info)
		if resp.Result.Status == "" {
			t.Errorf("orderCancel result missing status")
		}
	})

	t.Run("OrderCancelReplace", func(t *testing.T) {
		info := placeOrder(t, "orderCancelReplace", symbol, qtyStr, priceStr)
		t.Cleanup(func() { ensureOrderClosed(info) })

		basePrice := parseFloat64(info.price)
		if basePrice <= 0 {
			basePrice = targetPrice
		}
		newPriceStr := formatPrice(basePrice * 0.95)
		if newPriceStr == info.price {
			newPriceStr = formatPrice(basePrice * 0.9)
		}

		id := time.Now().UnixNano()
		req := &models.OrderCancelReplaceRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = info.symbol
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = info.qty
		req.Params.Price = newPriceStr
		req.Params.CancelReplaceMode = "STOP_ON_FAILURE"
		req.Params.CancelOrderId = info.orderID
		req.Params.Timestamp = time.Now().UnixMilli()
		req.Params.RecvWindow = 5000
		params := map[string]interface{}{
			"symbol":            req.Params.Symbol,
			"side":              req.Params.Side,
			"type":              req.Params.Type,
			"timeInForce":       req.Params.TimeInForce,
			"quantity":          req.Params.Quantity,
			"price":             req.Params.Price,
			"cancelReplaceMode": req.Params.CancelReplaceMode,
			"cancelOrderId":     req.Params.CancelOrderId,
			"timestamp":         req.Params.Timestamp,
			"recvWindow":        req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "orderCancelReplace.request", req)
		handler, ch := newResponseHandler[models.OrderCancelReplaceResponse](t, "orderCancelReplace.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OrderCancelReplace(ctx, req, handler); err != nil {
			t.Fatalf("orderCancelReplace call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "orderCancelReplace")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("orderCancelReplace id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("orderCancelReplace status want 200 got %d", resp.Status)
		}
		if resp.Result.CancelResult != "" && resp.Result.CancelResult != "SUCCESS" {
			t.Errorf("orderCancelReplace cancel result: want SUCCESS got %s", resp.Result.CancelResult)
		}
		if resp.Result.NewOrderResult != "" && resp.Result.NewOrderResult != "SUCCESS" {
			t.Errorf("orderCancelReplace new order result: want SUCCESS got %s", resp.Result.NewOrderResult)
		}
		if resp.Result.NewOrderResponse.OrderId == 0 {
			t.Fatal("orderCancelReplace missing replacement order id")
		}
		if resp.Result.NewOrderResponse.Price != "" && resp.Result.NewOrderResponse.Price != newPriceStr {
			t.Errorf("orderCancelReplace price mismatch: want %s got %s", newPriceStr, resp.Result.NewOrderResponse.Price)
		}
		newInfo := &orderInfo{
			symbol:        info.symbol,
			orderID:       resp.Result.NewOrderResponse.OrderId,
			clientOrderID: resp.Result.NewOrderResponse.ClientOrderId,
			price:         resp.Result.NewOrderResponse.Price,
			qty:           resp.Result.NewOrderResponse.OrigQty,
		}
		t.Cleanup(func() { ensureOrderClosed(newInfo) })
	})

	t.Run("OpenOrdersCancelAll", func(t *testing.T) {
		info := placeOrder(t, "openOrdersCancelAll", symbol, qtyStr, priceStr)
		defer ensureOrderClosed(info)

		id := time.Now().UnixNano()
		req := &models.OpenOrdersCancelAllRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = info.symbol
		req.Params.Timestamp = time.Now().UnixMilli()
		req.Params.RecvWindow = 5000
		params := map[string]interface{}{
			"symbol":     req.Params.Symbol,
			"timestamp":  req.Params.Timestamp,
			"recvWindow": req.Params.RecvWindow,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "openOrdersCancelAll.request", req)
		handler, ch := newResponseHandler[models.OpenOrdersCancelAllResponse](t, "openOrdersCancelAll.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.OpenOrdersCancelAll(ctx, req, handler); err != nil {
			t.Fatalf("openOrdersCancelAll call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "openOrdersCancelAll")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("openOrdersCancelAll id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("openOrdersCancelAll status want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Fatalf("openOrdersCancelAll returned no cancellations")
		}
		found := false
		for _, entry := range resp.Result {
			if entry.OrderId == info.orderID || entry.ClientOrderId == info.clientOrderID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("openOrdersCancelAll did not report cancellation for order %d", info.orderID)
		}
	})

	t.Run("SorOrderTest", func(t *testing.T) {
		if !sorSupported {
			if !sorAvailable {
				t.Skip("no SOR-supported symbols available for trading requests")
			}
			if len(sorDiscoveryErrors) > 0 {
				t.Skipf("unable to prepare SOR order parameters: %s", strings.Join(sorDiscoveryErrors, "; "))
			}
			t.Skip("unable to resolve SOR order parameters for current environment")
		}
		if sorSymbol == "" {
			t.Skip("SOR symbol not resolved")
		}
		t.Logf("sorOrderTest using %s (qty=%s price=%s)", sorSymbol, sorQtyStr, sorPriceStr)
		qtyFloat := parseFloat64(sorQtyStr)
		priceFloat := parseFloat64(sorPriceStr)
		if qtyFloat <= 0 || priceFloat <= 0 {
			t.Skip("quantity or price invalid for SOR test")
		}
		id := time.Now().UnixNano()
		req := &models.SorOrderTestRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = sorSymbol
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = qtyFloat
		req.Params.Price = priceFloat
		req.Params.Timestamp = time.Now().UnixMilli()
		params := map[string]interface{}{
			"symbol":      req.Params.Symbol,
			"side":        req.Params.Side,
			"type":        req.Params.Type,
			"timeInForce": req.Params.TimeInForce,
			"quantity":    req.Params.Quantity,
			"price":       req.Params.Price,
			"timestamp":   req.Params.Timestamp,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "sorOrderTest.request", req)
		handler, ch := newResponseHandler[models.SorOrderTestResponse](t, "sorOrderTest.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SorOrderTest(ctx, req, handler); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "sor") {
				t.Skipf("sorOrderTest not supported: %v", err)
			}
			t.Fatalf("sorOrderTest call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "sorOrderTest")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sorOrderTest id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("sorOrderTest status want 200 got %d", resp.Status)
		}
	})

	t.Run("SorOrderPlace", func(t *testing.T) {
		if !sorSupported {
			if !sorAvailable {
				t.Skip("no SOR-supported symbols available for trading requests")
			}
			if len(sorDiscoveryErrors) > 0 {
				t.Skipf("unable to prepare SOR order parameters: %s", strings.Join(sorDiscoveryErrors, "; "))
			}
			t.Skip("unable to resolve SOR order parameters for current environment")
		}
		if sorSymbol == "" {
			t.Skip("SOR symbol not resolved")
		}
		t.Logf("sorOrderPlace using %s (qty=%s price=%s)", sorSymbol, sorQtyStr, sorPriceStr)
		qtyFloat := parseFloat64(sorQtyStr)
		priceFloat := parseFloat64(sorPriceStr)
		if qtyFloat <= 0 || priceFloat <= 0 {
			t.Skip("quantity or price invalid for SOR order")
		}

		id := time.Now().UnixNano()
		req := &models.SorOrderPlaceRequest{Id: models.NewMessageIDInt64(id)}
		req.Params.Symbol = sorSymbol
		req.Params.Side = "BUY"
		req.Params.Type = "LIMIT"
		req.Params.TimeInForce = "GTC"
		req.Params.Quantity = sorQtyStr
		req.Params.Price = sorPriceStr
		req.Params.Timestamp = time.Now().UnixMilli()
		req.Params.RecvWindow = 5000
		req.Params.NewClientOrderId = newClientOrderID("sor-order", id)
		params := map[string]interface{}{
			"symbol":           req.Params.Symbol,
			"side":             req.Params.Side,
			"type":             req.Params.Type,
			"timeInForce":      req.Params.TimeInForce,
			"quantity":         req.Params.Quantity,
			"price":            req.Params.Price,
			"timestamp":        req.Params.Timestamp,
			"recvWindow":       req.Params.RecvWindow,
			"newClientOrderId": req.Params.NewClientOrderId,
		}
		signAndApply(t, h.Signer, spot.AuthTypeTrade, params, &req.Params)
		logRequestOnFailure(t, "sorOrderPlace.request", req)
		handler, ch := newResponseHandler[models.SorOrderPlaceResponse](t, "sorOrderPlace.response")
		ctx, cancel := requestContext()
		defer cancel()
		throttleWS()
		if err := h.Channel.SorOrderPlace(ctx, req, handler); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "sor") {
				t.Skipf("sorOrderPlace not supported: %v", err)
			}
			t.Fatalf("sorOrderPlace call failed: %v", err)
		}
		resp := awaitResponse(t, ch, "sorOrderPlace")
		if got, ok := resp.Id.ValInt64(); !ok || got != id {
			t.Errorf("sorOrderPlace id mismatch: want %d got %d (ok=%v)", id, got, ok)
		}
		if resp.Status != 200 {
			t.Fatalf("sorOrderPlace status want 200 got %d", resp.Status)
		}
		if len(resp.Result) == 0 {
			t.Fatalf("sorOrderPlace result empty")
		}
		entry := resp.Result[0]
		assertRecentMs(t, entry.TransactTime, 1*time.Minute, "sorOrderPlace.transactTime")
		if !strings.EqualFold(entry.Symbol, sorSymbol) {
			t.Errorf("sorOrderPlace symbol mismatch: want %s got %s", sorSymbol, entry.Symbol)
		}
		if entry.OrderId == 0 {
			t.Errorf("sorOrderPlace missing orderId")
		}
		if entry.Price == "" {
			t.Errorf("sorOrderPlace missing price")
		}
		ensureOrderClosed(&orderInfo{
			symbol:        sorSymbol,
			orderID:       entry.OrderId,
			clientOrderID: entry.ClientOrderId,
		})
	})
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

	_, terminatedCh := registerUserStreamEventHandlers(t, h)
	drainTerminated := func() { drainEventStreamTerminated(terminatedCh) }

	t.Run("UserDataStreamSubscribe", func(t *testing.T) {
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(after userDataStream.subscribe)")
		assertUserDataStreamState(t, h, true, "session.status(after userDataStream.subscribe)")

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(cleanup)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("cleanup unsubscribe status: want 200 got %d", respUnsub.Status)
		}
		evt := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after userDataStream.subscribe)")
		if evt.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after userDataStream.subscribe): subscriptionId mismatch: want %d got %d", subscriptionID, evt.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after userDataStream.subscribe cleanup)")
		drainTerminated()
	})

	t.Run("UserDataStreamSubscribeSignature", func(t *testing.T) {
		if h.Signer == nil {
			t.Skip("signer not available for signature subscription")
		}

		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribeSignature(t, h, "userDataStream.subscribe.signature")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(after userDataStream.subscribe.signature)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe.signature: subscription id not reported; will verify via session status and terminate event")
		}
		assertUserDataStreamState(t, h, true, "session.status(after userDataStream.subscribe.signature)")

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(cleanup.signature)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("cleanup(signature) unsubscribe status: want 200 got %d", respUnsub.Status)
		}
		evt := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after userDataStream.subscribe.signature)")
		if subscriptionID != 0 && evt.SubscriptionId != 0 && evt.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after userDataStream.subscribe.signature): subscriptionId mismatch: want %d got %d", subscriptionID, evt.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after userDataStream.subscribe.signature cleanup)")
		drainTerminated()
	})

	t.Run("UserDataStreamUnsubscribe", func(t *testing.T) {
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-unsubscribe)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before userDataStream.unsubscribe)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-unsubscribe): subscription id not reported; unsubscribing all")
		}
		assertUserDataStreamState(t, h, true, "session.status(before userDataStream.unsubscribe)")

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("unsubscribe status: want 200 got %d", respUnsub.Status)
		}
		evt := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after userDataStream.unsubscribe)")
		if subscriptionID != 0 && evt.SubscriptionId != 0 && evt.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after userDataStream.unsubscribe): subscriptionId mismatch: want %d got %d", subscriptionID, evt.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after userDataStream.unsubscribe)")
		drainTerminated()

		// Validate that unsubscribe without an explicit subscription id closes all streams.
		respAll := userStreamSubscribe(t, h, "userDataStream.subscribe(for-unsubscribe-all)")
		subscriptionIDAll := resolveUserStreamSubscriptionID(t, h, respAll.Result.SubscriptionId, "session.subscriptions(before userDataStream.unsubscribe(all))")
		if subscriptionIDAll == 0 {
			t.Log("userDataStream.subscribe(for-unsubscribe-all): subscription id not reported; unsubscribing all without id")
		}
		assertUserDataStreamState(t, h, true, "session.status(before userDataStream.unsubscribe(all))")

		respAllUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(all)", 0)
		if respAllUnsub.Status != 200 {
			t.Errorf("unsubscribe(all) status: want 200 got %d", respAllUnsub.Status)
		}
		evtAll := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after userDataStream.unsubscribe(all))")
		if subscriptionIDAll != 0 && evtAll.SubscriptionId != 0 && evtAll.SubscriptionId != subscriptionIDAll {
			t.Logf("eventStreamTerminated(after userDataStream.unsubscribe(all)): subscriptionId mismatch: want %d got %d", subscriptionIDAll, evtAll.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after userDataStream.unsubscribe(all))")
		drainTerminated()
	})
}

func runKlineResponseHandlers(t *testing.T, h *channelHarness, symbolUpper string, symbolLower string) {
	t.Helper()
	_ = symbolLower // reserved for potential combined stream tests

	t.Run("Klines", func(t *testing.T) {
		id := time.Now().UnixNano()
		req := &models.KlinesRequest{Id: models.NewMessageIDInt64(id)}
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

	ensureSessionLoggedIn(t, h)

	recorder, terminatedCh := registerUserStreamEventHandlers(t, h)
	drainTerminated := func() { drainEventStreamTerminated(terminatedCh) }

	t.Run("EventStreamTerminated", func(t *testing.T) {
		recorder.reset()
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-eventStreamTerminated)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before eventStreamTerminated)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-eventStreamTerminated): subscription id not reported; unsubscribing all")
		}

		assertUserDataStreamState(t, h, true, "session.status(before eventStreamTerminated)")
		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(eventStreamTerminated)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("userDataStream.unsubscribe(eventStreamTerminated) status: want 200 got %d", respUnsub.Status)
		}

		evt := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(event handler)")
		if subscriptionID != 0 && evt.SubscriptionId != 0 && evt.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(event handler): subscriptionId mismatch: want %d got %d", subscriptionID, evt.SubscriptionId)
		}
		if evt.Event.EventTime == 0 {
			t.Errorf("eventStreamTerminated(event handler): missing event time")
		}
		if recorder.count("eventStreamTerminated") == 0 {
			t.Errorf("eventStreamTerminated(event handler): recorder missing event")
		}

		assertUserDataStreamState(t, h, false, "session.status(after eventStreamTerminated)")
		recorder.reset()
		drainTerminated()
	})

	t.Run("OrderUpdate", func(t *testing.T) {
		if !h.Config.supports(spot.AuthTypeTrade) {
			t.Skip("credentials do not include TRADE permission")
		}

		recorder.reset()
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-orderUpdate)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before orderUpdate)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-orderUpdate): subscription id not reported; proceeding with stream-wide unsubscribe")
		}
		assertUserDataStreamState(t, h, true, "session.status(before orderUpdate)")

		rc := newRESTClient()
		baseCtx := context.Background()
		priceRef, err := restTickerPrice(baseCtx, rc, symbol)
		if err != nil || priceRef <= 0 {
			t.Logf("ticker price fetch failed (%v); falling back to default price", err)
			priceRef = 100.0
		}
		targetPrice := priceRef * 0.5
		if targetPrice <= 0 {
			targetPrice = 50.0
		}
		qtyStr, priceStr, _, err := calcSpotLimitOrderParams(baseCtx, rc, symbol, targetPrice)
		if err != nil {
			t.Skipf("calcSpotLimitOrderParams failed: %v", err)
		}
		order, err := placeSpotLimitOrder(rc, symbol, "BUY", qtyStr, priceStr)
		if err != nil {
			if errors.Is(err, errSpotCredentialsMissing) {
				t.Skip("spot REST credentials not configured for trading operations")
			}
			if isInsufficientSpotBalanceError(err) {
				t.Skipf("spot account insufficient to place order: %v", err)
			}
			t.Fatalf("placeSpotLimitOrder failed: %v", err)
		}
		defer func() {
			if order != nil {
				if cerr := cancelSpotOrder(rc, order); cerr != nil {
					t.Logf("cancelSpotOrder cleanup error: %v", cerr)
				}
			}
		}()

		if !recorder.waitFor("orderUpdate", 1, eventWait()) {
			t.Log("orderUpdate event not observed within timeout; order may have been rejected before reaching stream")
		} else if evt := recorder.lastOrderUpdate(); evt != nil {
			if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "executionReport") {
				t.Errorf("orderUpdate event type mismatch: %s", evt.Event.EventType)
			}
			if evt.Event.EventTime == 0 {
				t.Log("orderUpdate event missing event time")
			}
			logJSON(t, "orderUpdate.event", evt)
		}

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(orderUpdate)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("userDataStream.unsubscribe(orderUpdate) status: want 200 got %d", respUnsub.Status)
		}
		evt := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after orderUpdate)")
		if subscriptionID != 0 && evt.SubscriptionId != 0 && evt.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after orderUpdate): subscriptionId mismatch: want %d got %d", subscriptionID, evt.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after orderUpdate)")
		recorder.reset()
		drainTerminated()
	})

	t.Run("AccountUpdate", func(t *testing.T) {
		if !h.Config.supports(spot.AuthTypeTrade) {
			t.Skip("credentials do not include TRADE permission")
		}

		recorder.reset()
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-accountUpdate)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before accountUpdate)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-accountUpdate): subscription id not reported; proceeding with stream-wide unsubscribe")
		}
		assertUserDataStreamState(t, h, true, "session.status(before accountUpdate)")

		rc := newRESTClient()
		baseCtx := context.Background()
		priceRef, err := restTickerPrice(baseCtx, rc, symbol)
		if err != nil || priceRef <= 0 {
			t.Logf("ticker price fetch failed (%v); using fallback for accountUpdate test", err)
			priceRef = 100.0
		}
		targetPrice := priceRef * 0.4
		if targetPrice <= 0 {
			targetPrice = 40.0
		}
		qtyStr, priceStr, _, err := calcSpotLimitOrderParams(baseCtx, rc, symbol, targetPrice)
		if err != nil {
			t.Skipf("calcSpotLimitOrderParams(accountUpdate) failed: %v", err)
		}
		order, err := placeSpotLimitOrder(rc, symbol, "BUY", qtyStr, priceStr)
		if err != nil {
			if errors.Is(err, errSpotCredentialsMissing) {
				t.Skip("spot REST credentials not configured for trading operations")
			}
			if isInsufficientSpotBalanceError(err) {
				t.Skipf("spot account insufficient to place limit order for accountUpdate: %v", err)
			}
			t.Fatalf("placeSpotLimitOrder(accountUpdate) failed: %v", err)
		}
		gotAccountUpdate := recorder.waitFor("accountUpdate", 1, eventWait())
		if !gotAccountUpdate {
			t.Log("accountUpdate event not observed immediately after limit order placement; will check after cancel")
		}
		if order != nil {
			if cerr := cancelSpotOrder(rc, order); cerr != nil {
				t.Logf("cancelSpotOrder(accountUpdate) error: %v", cerr)
			}
			order = nil
		}
		if !gotAccountUpdate {
			if recorder.waitFor("accountUpdate", 1, 5*time.Second) {
				gotAccountUpdate = true
			}
		}
		if !gotAccountUpdate {
			t.Fatalf("accountUpdate event not observed after limit order placement and cancellation")
		}
		if evt := recorder.lastAccountUpdate(); evt != nil {
			if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "outboundAccountPosition") {
				t.Errorf("accountUpdate event type mismatch: %s", evt.Event.EventType)
			}
			if evt.Event.EventTime == 0 {
				t.Log("accountUpdate event missing event time")
			}
			if len(evt.Event.BalancesArray) == 0 {
				t.Log("accountUpdate event missing balances array entries")
			}
			logJSON(t, "accountUpdate.event", evt)
		}

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(accountUpdate)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("userDataStream.unsubscribe(accountUpdate) status: want 200 got %d", respUnsub.Status)
		}
		evtTerm := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after accountUpdate)")
		if subscriptionID != 0 && evtTerm.SubscriptionId != 0 && evtTerm.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after accountUpdate): subscriptionId mismatch: want %d got %d", subscriptionID, evtTerm.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after accountUpdate)")
		recorder.reset()
		drainTerminated()
	})

	t.Run("BalanceUpdate", func(t *testing.T) {
		if !h.Config.supports(spot.AuthTypeTrade) {
			t.Skip("credentials do not include TRADE permission")
		}

		recorder.reset()
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-balanceUpdate)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before balanceUpdate)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-balanceUpdate): subscription id not reported; proceeding with stream-wide unsubscribe")
		}
		assertUserDataStreamState(t, h, true, "session.status(before balanceUpdate)")

		rc := newRESTClient()
		baseCtx := context.Background()
		constraints, err := loadSpotSymbolConstraints(baseCtx, rc, symbol)
		if err != nil {
			t.Skipf("loadSpotSymbolConstraints(balanceUpdate) failed: %v", err)
		}
		quote := constraints.minNotional
		if quote <= 0 {
			quote = 15.0
		}
		if quote < 12.0 {
			quote = 12.0
		}
		if quote > 50.0 {
			quote = 50.0
		}
		quoteStr := fmt.Sprintf("%.*f", maxInt(2, constraints.priceDecimals), quote)
		if _, err := placeSpotMarketOrder(rc, symbol, "BUY", "", quoteStr); err != nil {
			if errors.Is(err, errSpotCredentialsMissing) {
				t.Skip("spot REST credentials not configured for trading operations")
			}
			if isInsufficientSpotBalanceError(err) {
				t.Skipf("spot account insufficient to place market order for balanceUpdate: %v", err)
			}
			t.Fatalf("placeSpotMarketOrder(balanceUpdate) failed: %v", err)
		}

		gotBalanceUpdate := recorder.waitFor("balanceUpdate", 1, eventWait())
		if !gotBalanceUpdate {
			t.Log("balanceUpdate event not observed after market order; acceptable when balance remains unchanged or settlement delayed")
		} else if evt := recorder.lastBalanceUpdate(); evt != nil {
			if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "balanceUpdate") {
				t.Errorf("balanceUpdate event type mismatch: %s", evt.Event.EventType)
			}
			if evt.Event.EventTime == 0 {
				t.Log("balanceUpdate event missing event time")
			}
			if evt.Event.Asset == "" {
				t.Log("balanceUpdate event missing asset")
			}
			if evt.Event.BalanceDelta == "" {
				t.Log("balanceUpdate event missing balance delta")
			}
			logJSON(t, "balanceUpdate.event", evt)
		}

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(balanceUpdate)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("userDataStream.unsubscribe(balanceUpdate) status: want 200 got %d", respUnsub.Status)
		}
		evtTerm := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after balanceUpdate)")
		if subscriptionID != 0 && evtTerm.SubscriptionId != 0 && evtTerm.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after balanceUpdate): subscriptionId mismatch: want %d got %d", subscriptionID, evtTerm.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after balanceUpdate)")
		recorder.reset()
		drainTerminated()
	})

	t.Run("ListStatus", func(t *testing.T) {
		if !h.Config.supports(spot.AuthTypeTrade) {
			t.Skip("credentials do not include TRADE permission")
		}

		recorder.reset()
		drainTerminated()
		ensureUserStreamInactive(t, h)
		drainTerminated()

		resp := userStreamSubscribe(t, h, "userDataStream.subscribe(for-listStatus)")
		subscriptionID := resolveUserStreamSubscriptionID(t, h, resp.Result.SubscriptionId, "session.subscriptions(before listStatus)")
		if subscriptionID == 0 {
			t.Log("userDataStream.subscribe(for-listStatus): subscription id not reported; proceeding with stream-wide unsubscribe")
		}
		assertUserDataStreamState(t, h, true, "session.status(before listStatus)")

		rc := newRESTClient()
		baseCtx := context.Background()
		priceRef, err := restTickerPrice(baseCtx, rc, symbol)
		if err != nil || priceRef <= 0 {
			t.Logf("ticker price fetch failed (%v); using fallback for listStatus test", err)
			priceRef = 100.0
		}

		belowTarget := priceRef * 0.9
		if belowTarget <= 0 {
			belowTarget = 50.0
		}
		qtyStr, belowPrice, constraints, err := calcSpotLimitOrderParams(baseCtx, rc, symbol, belowTarget)
		if err != nil {
			t.Skipf("calcSpotLimitOrderParams(listStatus) failed: %v", err)
		}
		if constraints == nil {
			t.Skip("spot symbol constraints unavailable for listStatus")
		}
		belowPriceVal := parseFloat64(belowPrice)
		if belowPriceVal <= 0 {
			t.Skipf("invalid below price derived for listStatus: %s", belowPrice)
		}

		step := constraints.priceStep
		if step <= 0 {
			step = 1.0 / math.Pow10(constraints.priceDecimals)
		}

		stopTarget := priceRef * 1.05
		if stopTarget <= belowPriceVal {
			stopTarget = belowPriceVal + step*4
		}
		stopPriceVal := ceilToStep(stopTarget, constraints.priceStep, constraints.priceDecimals)
		if stopPriceVal <= belowPriceVal {
			stopPriceVal = ceilToStep(belowPriceVal+step*4, constraints.priceStep, constraints.priceDecimals)
		}
		stopLimitVal := ceilToStep(stopPriceVal*1.001, constraints.priceStep, constraints.priceDecimals)
		if stopLimitVal < stopPriceVal {
			stopLimitVal = stopPriceVal
		}
		aboveStopPrice := fmt.Sprintf("%.*f", constraints.priceDecimals, stopPriceVal)
		aboveLimitPrice := fmt.Sprintf("%.*f", constraints.priceDecimals, stopLimitVal)

		orderList, err := placeSpotOcoOrder(rc, symbol, "BUY", qtyStr, belowPrice, aboveStopPrice, aboveLimitPrice)
		if err != nil {
			if errors.Is(err, errSpotCredentialsMissing) {
				t.Skip("spot REST credentials not configured for trading operations")
			}
			if isInsufficientSpotBalanceError(err) {
				t.Skipf("spot account insufficient to place OCO order for listStatus: %v", err)
			}
			t.Fatalf("placeSpotOcoOrder(listStatus) failed: %v", err)
		}
		defer func() {
			if orderList != nil {
				if cerr := cancelSpotOrderList(rc, orderList); cerr != nil {
					t.Logf("cancelSpotOrderList(listStatus) error: %v", cerr)
				}
			}
		}()

		if !recorder.waitFor("listStatus", 1, eventWait()) {
			t.Fatalf("listStatus event not observed after OCO order placement")
		}
		if evt := recorder.lastListStatus(); evt != nil {
			if evt.Event.EventType != "" && !strings.EqualFold(evt.Event.EventType, "listStatus") {
				t.Errorf("listStatus event type mismatch: %s", evt.Event.EventType)
			}
			if evt.Event.EventTime == 0 {
				t.Log("listStatus event missing event time")
			}
			if evt.Event.ListClientOrderID == "" {
				t.Log("listStatus event missing list client order id")
			} else if orderList != nil && orderList.listClientOrderID != "" && !strings.EqualFold(evt.Event.ListClientOrderID, orderList.listClientOrderID) {
				t.Errorf("listStatus event list client order id mismatch: want %s got %s", orderList.listClientOrderID, evt.Event.ListClientOrderID)
			}
			if evt.Event.Symbol == "" {
				t.Log("listStatus event missing symbol")
			}
			if len(evt.Event.ArrayOfOrdersInTheList) == 0 {
				t.Log("listStatus event missing orders array")
			}
			if evt.Event.OrderListId == 0 {
				t.Log("listStatus event missing order list id")
			} else if orderList != nil && orderList.orderListID != 0 && evt.Event.OrderListId != orderList.orderListID {
				t.Errorf("listStatus event order list id mismatch: want %d got %d", orderList.orderListID, evt.Event.OrderListId)
			}
			if evt.Event.ListStatusType == "" {
				t.Log("listStatus event missing list status type")
			}
			logJSON(t, "listStatus.event", evt)
		}

		respUnsub := userStreamUnsubscribe(t, h, "userDataStream.unsubscribe(listStatus)", subscriptionID)
		if respUnsub.Status != 200 {
			t.Errorf("userDataStream.unsubscribe(listStatus) status: want 200 got %d", respUnsub.Status)
		}
		evtTerm := awaitEventStreamTerminated(t, terminatedCh, "eventStreamTerminated(after listStatus)")
		if subscriptionID != 0 && evtTerm.SubscriptionId != 0 && evtTerm.SubscriptionId != subscriptionID {
			t.Logf("eventStreamTerminated(after listStatus): subscriptionId mismatch: want %d got %d", subscriptionID, evtTerm.SubscriptionId)
		}
		assertUserDataStreamState(t, h, false, "session.status(after listStatus)")
		recorder.reset()
		drainTerminated()
	})

	t.Run("ExternalLockUpdate", func(t *testing.T) {
		t.Skip("triggering externalLockUpdate events requires external locking; pending coverage")
	})
}
