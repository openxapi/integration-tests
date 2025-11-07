package streamstest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	restpmargin "github.com/openxapi/binance-go/rest/pmargin"
)

func newPMarginProRESTClient() *restpmargin.APIClient {
	cfg := restpmargin.NewConfiguration()
	if override := strings.TrimSpace(os.Getenv("BINANCE_PM_REST_SERVER")); override != "" {
		cfg.Servers[0].URL = override
	} else if override := strings.TrimSpace(os.Getenv("BINANCE_REST_SERVER_URL")); override != "" {
		cfg.Servers[0].URL = override
	}
	return restpmargin.NewAPIClient(cfg)
}

func restAuthContext(ctx context.Context, apiKey, secret string) (context.Context, *restpmargin.Auth, error) {
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(secret) == "" {
		return nil, nil, fmt.Errorf("missing API credentials")
	}
	auth := restpmargin.NewAuth(apiKey)
	auth.SetSecretKey(secret)
	authCtx, err := auth.ContextWithValue(ctx)
	if err != nil {
		return nil, nil, err
	}
	return authCtx, auth, nil
}

type listenKeyHandle struct {
	ListenKey string
	Renew     func(context.Context) error
	Close     func(context.Context) error
}

func restCreateListenKey(ctx context.Context, client *restpmargin.APIClient) (*listenKeyHandle, error) {
	resp, _, err := client.PortfolioMarginAPI.CreateListenKeyV1(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("create listen key: %w", err)
	}
	if resp == nil || resp.ListenKey == nil || *resp.ListenKey == "" {
		return nil, fmt.Errorf("listen key response empty")
	}
	key := strings.TrimSpace(*resp.ListenKey)
	handle := &listenKeyHandle{
		ListenKey: key,
	}
	handle.Renew = func(rctx context.Context) error {
		renewCtx, cancel := context.WithTimeout(rctx, 5*time.Second)
		defer cancel()
		if _, _, err := client.PortfolioMarginAPI.UpdateListenKeyV1(renewCtx).Execute(); err != nil {
			return fmt.Errorf("keepalive listen key: %w", err)
		}
		return nil
	}
	handle.Close = func(rctx context.Context) error {
		closeCtx, cancel := context.WithTimeout(rctx, 5*time.Second)
		defer cancel()
		if _, _, err := client.PortfolioMarginAPI.DeleteListenKeyV1(closeCtx).Execute(); err != nil {
			return fmt.Errorf("delete listen key: %w", err)
		}
		return nil
	}
	return handle, nil
}
