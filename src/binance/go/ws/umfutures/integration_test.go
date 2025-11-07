package wstest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	umfutures "github.com/openxapi/binance-go/ws/umfutures"
)

const (
	defaultConnectTimeout    = 12 * time.Second
	defaultRequestTimeout    = 10 * time.Second
	defaultDisconnectTimeout = 6 * time.Second
	defaultTestnetServer     = "testnet"
	overrideServerName       = "override"
)

// TestConfig describes one credential bundle that can be used against the WS API.
type TestConfig struct {
	Name           string
	Description    string
	KeyType        umfutures.KeyType
	APIKey         string
	SecretKey      string
	PrivateKeyPath string
	PrivateKeyPass string
	SupportsAuth   []umfutures.AuthType
}

// CredentialStore lazily loads credential bundles from the environment.
type CredentialStore struct {
	Public  *TestConfig
	HMAC    *TestConfig
	RSA     *TestConfig
	Ed25519 *TestConfig
}

var (
	credsOnce sync.Once
	creds     CredentialStore
)

func getCreds() CredentialStore {
	credsOnce.Do(func() {
		creds = loadCredentialStore()
	})
	return creds
}

func loadCredentialStore() CredentialStore {
	store := CredentialStore{
		Public: &TestConfig{
			Name:        "Public-NoAuth",
			Description: "Public market-data requests (no authentication)",
			SupportsAuth: []umfutures.AuthType{
				umfutures.AuthTypeNone,
			},
		},
	}

	if apiKey := strings.TrimSpace(os.Getenv("BINANCE_API_KEY")); apiKey != "" {
		if secret := strings.TrimSpace(os.Getenv("BINANCE_SECRET_KEY")); secret != "" {
			store.HMAC = &TestConfig{
				Name:         "HMAC",
				Description:  "Signed USER_DATA / TRADE requests via HMAC",
				KeyType:      umfutures.KeyTypeHMAC,
				APIKey:       apiKey,
				SecretKey:    secret,
				SupportsAuth: []umfutures.AuthType{umfutures.AuthTypeUserData, umfutures.AuthTypeUserStream, umfutures.AuthTypeTrade},
			}
		}
	}

	if apiKey := strings.TrimSpace(os.Getenv("BINANCE_RSA_API_KEY")); apiKey != "" {
		if path := strings.TrimSpace(os.Getenv("BINANCE_RSA_PRIVATE_KEY_PATH")); path != "" {
			store.RSA = &TestConfig{
				Name:           "RSA",
				Description:    "Signed requests via RSA private key",
				KeyType:        umfutures.KeyTypeRSA,
				APIKey:         apiKey,
				PrivateKeyPath: path,
				PrivateKeyPass: strings.TrimSpace(os.Getenv("BINANCE_RSA_PRIVATE_KEY_PASSPHRASE")),
				SupportsAuth:   []umfutures.AuthType{umfutures.AuthTypeUserData, umfutures.AuthTypeUserStream, umfutures.AuthTypeTrade, umfutures.AuthTypeSigned},
			}
		}
	}

	if apiKey := strings.TrimSpace(os.Getenv("BINANCE_ED25519_API_KEY")); apiKey != "" {
		if path := strings.TrimSpace(os.Getenv("BINANCE_ED25519_PRIVATE_KEY_PATH")); path != "" {
			store.Ed25519 = &TestConfig{
				Name:           "Ed25519",
				Description:    "Signed requests using Ed25519 (required for session.logon)",
				KeyType:        umfutures.KeyTypeED25519,
				APIKey:         apiKey,
				PrivateKeyPath: path,
				PrivateKeyPass: strings.TrimSpace(os.Getenv("BINANCE_ED25519_PRIVATE_KEY_PASSPHRASE")),
				SupportsAuth:   []umfutures.AuthType{umfutures.AuthTypeUserData, umfutures.AuthTypeUserStream, umfutures.AuthTypeTrade, umfutures.AuthTypeSigned},
			}
		}
	}

	return store
}

func (cfg *TestConfig) supports(auth umfutures.AuthType) bool {
	if cfg == nil {
		return false
	}
	if auth == umfutures.AuthTypeNone {
		return true
	}
	if cfg.APIKey == "" {
		return false
	}
	if len(cfg.SupportsAuth) == 0 {
		// Assume all signed flows if unspecified.
		return true
	}
	for _, a := range cfg.SupportsAuth {
		if a == auth {
			return true
		}
	}
	return false
}

// newClientAndSigner constructs a client (always switching to testnet unless overridden)
// and returns an optional signer when credentials are available.
func (cfg *TestConfig) newClientAndSigner(t testing.TB) (*umfutures.Client, *umfutures.RequestSigner) {
	t.Helper()
	if cfg == nil {
		t.Fatalf("test configuration is required")
	}

	client := umfutures.NewClient()
	if err := ensureDefaultServer(client); err != nil {
		t.Fatalf("failed to select WS server for %s: %v", cfg.Name, err)
	}

	if cfg.APIKey == "" {
		return client, nil
	}

	auth := umfutures.NewAuth(cfg.APIKey)
	switch cfg.KeyType {
	case umfutures.KeyTypeHMAC:
		if cfg.SecretKey == "" {
			t.Skipf("skipping %s: BINANCE_SECRET_KEY not configured", cfg.Name)
		}
		auth.SetSecretKey(cfg.SecretKey)
	case umfutures.KeyTypeRSA, umfutures.KeyTypeED25519:
		if cfg.PrivateKeyPath == "" {
			t.Skipf("skipping %s: private key path not configured", cfg.Name)
		}
		auth.SetPrivateKeyPath(cfg.PrivateKeyPath)
		if cfg.PrivateKeyPass != "" {
			auth.SetPassphrase(cfg.PrivateKeyPass)
		}
	default:
		t.Fatalf("unsupported key type %q for config %s", cfg.KeyType, cfg.Name)
	}

	client.SetAuth(auth)
	signer := umfutures.NewRequestSigner(auth)
	if err := signer.EnsureInitialized(); err != nil {
		t.Skipf("credentials for %s are not usable: %v", cfg.Name, err)
	}

	return client, signer
}

// channelHarness bundles a client, signer, and channel for a specific credential set.
type channelHarness struct {
	Config  *TestConfig
	Client  *umfutures.Client
	Signer  *umfutures.RequestSigner
	Channel *umfutures.UmfuturesChannel
}

func newChannelHarness(t testing.TB, cfg *TestConfig) *channelHarness {
	t.Helper()
	client, signer := cfg.newClientAndSigner(t)
	return &channelHarness{
		Config:  cfg,
		Client:  client,
		Signer:  signer,
		Channel: umfutures.NewUmfuturesChannel(client),
	}
}

func (h *channelHarness) connect(t testing.TB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()
	if err := h.Channel.Connect(ctx); err != nil {
		active := "<unknown>"
		if info := h.Client.GetActiveServer(); info != nil && info.URL != "" {
			active = info.URL
		}
		detailParts := []string{fmt.Sprintf("server=%s", active)}
		if errors.Is(err, websocket.ErrBadHandshake) {
			detailParts = append(detailParts, "handshake_err=websocket.ErrBadHandshake")
		}
		if chain := unwrapErrorChain(err); chain != "" {
			detailParts = append(detailParts, fmt.Sprintf("error_chain=%s", chain))
		}
		if probe := probeHTTPContext(active); probe != "" {
			detailParts = append(detailParts, probe)
		}
		t.Fatalf("connect %s failed: %v (%s)", h.Config.Name, err, strings.Join(detailParts, " "))
	}
}

func (h *channelHarness) disconnect(t testing.TB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultDisconnectTimeout)
	defer cancel()
	if err := h.Channel.Disconnect(ctx); err != nil && !isContextCanceled(err) {
		t.Errorf("disconnect %s: %v", h.Config.Name, err)
	}
}

func isContextCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func ensureDefaultServer(client *umfutures.Client) error {
	if v := strings.TrimSpace(os.Getenv("BINANCE_UMFUTURES_WS_SERVER")); v != "" {
		if err := client.AddOrUpdateServer(overrideServerName, v, "Override", "Test override server"); err != nil {
			return fmt.Errorf("register override server: %w", err)
		}
		if err := client.SetActiveServer(overrideServerName); err != nil {
			return fmt.Errorf("activate override server: %w", err)
		}
		return nil
	}
	if err := client.SetActiveServer(defaultTestnetServer); err != nil {
		return fmt.Errorf("set active server %q: %w", defaultTestnetServer, err)
	}
	return nil
}

func unwrapErrorChain(err error) string {
	if err == nil {
		return ""
	}
	seen := make(map[string]struct{})
	var parts []string
	for current := err; current != nil; current = errors.Unwrap(current) {
		part := current.Error()
		if _, ok := seen[part]; ok {
			break
		}
		seen[part] = struct{}{}
		parts = append(parts, part)
	}
	return strings.Join(parts, " -> ")
}

func probeHTTPContext(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Sprintf("probe=parse_err:%v", err)
	}
	switch u.Scheme {
	case "wss":
		u.Scheme = "https"
	case "ws":
		u.Scheme = "http"
	case "https", "http":
	default:
		return fmt.Sprintf("probe=unsupported_scheme:%s", u.Scheme)
	}

	probeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Sprintf("probe=request_err:%v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			return fmt.Sprintf("probe=network_err:%v", netErr)
		}
		return fmt.Sprintf("probe=do_err:%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}
	return fmt.Sprintf("probe_status=%d %s", resp.StatusCode, snippet)
}
