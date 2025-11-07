package streamstest

import (
	"context"
	"strings"
	"testing"
	"time"

	pmarginpro "github.com/openxapi/binance-go/ws/pmarginpro-streams"
)

func TestClient_DefaultConfiguration(t *testing.T) {
	client := pmarginpro.NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if active := client.GetActiveServer(); active == nil {
		t.Fatal("expected default active server configured by SDK")
	} else if client.GetCurrentURL() == "" {
		t.Fatal("GetCurrentURL returned empty string")
	}

	if got := client.GetURL(); got != client.GetCurrentURL() {
		t.Errorf("GetURL mismatch: %s vs %s", got, client.GetCurrentURL())
	}

	if srv := client.ListServers(); len(srv) == 0 {
		t.Errorf("expected default servers from SDK, got %d", len(srv))
	}
}

func TestClient_WithOptionsAndAuth(t *testing.T) {
	opts := &pmarginpro.ClientOptions{HandlerWorkers: 2}
	client := pmarginpro.NewClientWithOptions(opts)
	if client == nil {
		t.Fatal("NewClientWithOptions returned nil")
	}
	client.RegisterHandlers("noop", map[string]func(context.Context, []byte) error{
		"evt:test": func(ctx context.Context, payload []byte) error { return nil },
	})

	auth := pmarginpro.NewAuth("test-key")
	auth.SetSecretKey("test-secret")
	clientWithAuth := pmarginpro.NewClientWithAuth(auth)
	if clientWithAuth == nil {
		t.Fatal("NewClientWithAuth returned nil")
	}
	clientWithAuth.SetAuth(auth)
	if clientWithAuth.GetActiveServer() == nil {
		t.Fatal("clientWithAuth missing active server")
	}
	clientWithAuth.StopReadLoop()
	if err := clientWithAuth.Wait(context.Background()); err != nil {
		t.Fatalf("Wait returned error without active connection: %v", err)
	}
}

func TestClient_ServerLifecycle(t *testing.T) {
	client := pmarginpro.NewClient()
	defaultActive := client.GetActiveServer()
	if defaultActive == nil {
		t.Fatal("expected default active server")
	}

	err := client.AddServer("testnet", "wss://example.test/pm-pro", "Test Server", "Test description")
	if err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	if srv := client.GetServer("testnet"); srv == nil {
		t.Fatalf("GetServer missing entry for testnet")
	} else if srv.URL != "wss://example.test/pm-pro" {
		t.Errorf("GetServer URL mismatch: %s", srv.URL)
	}

	err = client.AddOrUpdateServer("testnet", "wss://example.test/pm-pro/ws", "Test Server (updated)", "Updated description")
	if err != nil {
		t.Fatalf("AddOrUpdateServer failed: %v", err)
	}

	err = client.SetActiveServer("testnet")
	if err != nil {
		t.Fatalf("SetActiveServer failed: %v", err)
	}

	if got := client.GetCurrentURL(); !strings.HasPrefix(got, "wss://example.test/pm-pro") {
		t.Errorf("unexpected active URL: %s", got)
	}

	err = client.UpdateServer("testnet", "wss://example.test/pm-pro/ws", "Test Server (final)", "Final description")
	if err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	if err := client.RemoveServer("testnet"); err == nil {
		t.Fatalf("RemoveServer succeeded for active server")
	}

	if defaultActive != nil {
		if err := client.SetActiveServer(defaultActive.Name); err != nil {
			t.Fatalf("restore default active server: %v", err)
		}
	}

	if err := client.RemoveServer("testnet"); err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	if srv := client.GetServer("testnet"); srv != nil {
		t.Fatalf("server entry still present after removal: %+v", srv)
	}
}

func TestServerManagerLifecycle(t *testing.T) {
	sm := pmarginpro.NewServerManager()
	primary := &pmarginpro.ServerInfo{
		URL:         "wss://primary.example/ws",
		Protocol:    "wss",
		Host:        "primary.example",
		Pathname:    "/ws",
		Title:       "Primary",
		Description: "Primary server",
	}
	if err := sm.AddServer("primary", primary); err != nil {
		t.Fatalf("AddServer primary failed: %v", err)
	}

	secondary := &pmarginpro.ServerInfo{
		URL:         "wss://secondary.example/pm/ws/{listenKey}",
		Protocol:    "wss",
		Host:        "secondary.example",
		Pathname:    "/pm/ws/{listenKey}",
		Title:       "Secondary",
		Description: "Secondary server",
	}
	if err := sm.AddServer("secondary", secondary); err != nil {
		t.Fatalf("AddServer secondary failed: %v", err)
	}

	if err := sm.SetActiveServer("secondary"); err != nil {
		t.Fatalf("SetActiveServer failed: %v", err)
	}

	if resolved, err := sm.ResolveServerURL("secondary", map[string]string{"listenKey": "abc123"}); err != nil {
		t.Fatalf("ResolveServerURL failed: %v", err)
	} else if !strings.Contains(resolved, "abc123") {
		t.Errorf("resolved URL missing listen key: %s", resolved)
	}

	if err := sm.UpdateServerPathname("secondary", "/pm/ws/updated"); err != nil {
		t.Fatalf("UpdateServerPathname failed: %v", err)
	}

	if info := sm.GetServer("secondary"); info == nil || info.Pathname != "/pm/ws/updated" {
		t.Fatalf("GetServer failed to reflect updated pathname: %+v", info)
	}

	if err := sm.UpdateServer("secondary", &pmarginpro.ServerInfo{
		URL:         "wss://secondary.example/pm/ws/{listenKey}",
		Protocol:    "wss",
		Host:        "secondary.example",
		Pathname:    "/pm/ws/{listenKey}",
		Title:       "Secondary Updated",
		Description: "Updated description",
	}); err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	if err := sm.RemoveServer("secondary"); err == nil {
		t.Fatalf("RemoveServer succeeded for active server")
	}

	if err := sm.SetActiveServer("primary"); err != nil {
		t.Fatalf("SetActiveServer primary failed: %v", err)
	}

	if err := sm.RemoveServer("secondary"); err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	if info := sm.GetServer("secondary"); info != nil {
		t.Fatalf("expected secondary to be removed, got %+v", info)
	}

	if url := sm.GetActiveServerURL(); !strings.HasPrefix(url, "wss://primary.example") {
		t.Errorf("unexpected active server URL: %s", url)
	}
}

func TestClientStopWaitWithoutConnection(t *testing.T) {
	client := pmarginpro.NewClient()
	client.StopReadLoop()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := client.Wait(ctx); err != nil && err != context.Canceled {
		t.Fatalf("Wait returned unexpected error: %v", err)
	}
}

func TestAuthContextWithValue(t *testing.T) {
	auth := pmarginpro.NewAuth("api-key")
	auth.SetSecretKey("secret-key")
	ctx, err := auth.ContextWithValue(context.Background())
	if err != nil {
		t.Fatalf("ContextWithValue failed: %v", err)
	}
	val, ok := ctx.Value(pmarginpro.ContextBinanceAuth).(pmarginpro.Auth)
	if !ok {
		t.Fatalf("auth not found in context")
	}
	if val.APIKey != "api-key" {
		t.Errorf("API key mismatch: %s", val.APIKey)
	}
}
