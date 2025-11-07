Binance Options WS SDK — Integration Coverage

SDK: `github.com/openxapi/binance-go/ws/options-streams`
REST: `github.com/openxapi/binance-go/rest/options`

Notes
- Request methods accept a third argument: pointer to per‑request callback (e.g., `*func(context.Context, *models.SubscribeResponse, error) error`). Tests pass concrete callbacks, assert ids/fields, and treat error payloads as first‑class.
- One IntegrationTestSuite per channel: Market, Combined, User Data. Field‑level assertions are performed for each event type where market activity exists.

Coverage Summary
- MarketStreamChannel: 100% of request methods and event handlers covered.
- CombinedMarketStreamChannel: 100% of request methods and event handlers (incl. wrapper) covered.
- UserDataStreamChannel: Connect/Disconnect and all event handlers covered (requires API keys).
- Stream builders and typed params: covered for all stream types used in tests.
- Core Client / Server management / WS Auth: pending (relied on SDK defaults in tests).

Client (type `Client`)
- [x] NewClient() *Client
- [ ] NewClientWithAuth(auth *Auth) *Client
- [ ] SetAuth(auth *Auth)
- [ ] AddServer(name, serverURL, title, description string) error
- [ ] AddOrUpdateServer(name, serverURL, title, description string) error
- [ ] RemoveServer(name string) error
- [ ] UpdateServer(name, serverURL, title, description string) error
- [ ] SetActiveServer(name string) error
- [x] GetActiveServer() *ServerInfo
- [ ] GetServer(name string) *ServerInfo
- [ ] ListServers() map[string]*ServerInfo
- [ ] GetCurrentURL() string
- [ ] GetURL() string (deprecated alias)
- [ ] RegisterHandlers(channel string, m map[string]func(context.Context, []byte) error)
- [ ] Wait(ctx context.Context) error

Server Management (type `ServerManager`, `ServerInfo`)
- [ ] NewServerManager() *ServerManager
- [ ] (sm) AddServer(name string, server *ServerInfo) error
- [ ] (sm) AddOrUpdateServer(name string, server *ServerInfo) error
- [ ] (sm) RemoveServer(name string) error
- [ ] (sm) UpdateServer(name string, server *ServerInfo) error
- [ ] (sm) UpdateServerPathname(name string, pathname string) error
- [ ] (sm) ResolveServerURL(name string, variables map[string]string) (string, error)
- [ ] (sm) SetActiveServer(name string) error
- [ ] (sm) GetActiveServer() *ServerInfo
- [ ] (sm) GetServer(name string) *ServerInfo
- [ ] (sm) ListServers() map[string]*ServerInfo
- [ ] (sm) GetActiveServerURL() string

Auth & Signing (types `Auth`, `RequestSigner`)
- [ ] NewAuth(apiKey string) *Auth
- [ ] (a) SetSecretKey(secretKey string)
- [ ] (a) SetPrivateKey(privateKey []byte) error
- [ ] (a) SetPrivateKeyPath(path string)
- [ ] (a) SetPrivateKeyReader(reader io.Reader)
- [ ] (a) SetPassphrase(passphrase string)
- [ ] (a) ContextWithValue(ctx context.Context) (context.Context, error)
- [ ] NewRequestSigner(auth *Auth) *RequestSigner
- [ ] (s) EnsureInitialized() error
- [ ] (s) SignRequest(params map[string]interface{}, authType AuthType) error
- [ ] GetAuthTypeFromMessageName(messageName string) AuthType
- [ ] RequiresSignature(authType AuthType) bool

Stream Builders & Typed Params (package functions/types)
- [x] BuildIndexPriceEventStream
- [x] BuildKlineEventStream
- [x] BuildMarkPriceEventStream
- [x] BuildTickerEventStream
- [x] BuildTickerByUnderlyingEventStream
- [x] BuildTradeEventStream
- [x] BuildPartialDepthEventStream
- [x] BuildNewSymbolInfoEventStream
- [x] BuildOpenInterestEventStream
- [x] IndexPriceEventStreamParams.Values()
- [x] KlineEventStreamParams.Values()
- [x] MarkPriceEventStreamParams.Values()
- [x] TickerEventStreamParams.Values()
- [x] TickerByUnderlyingEventStreamParams.Values()
- [x] TradeEventStreamParams.Values()
- [x] PartialDepthEventStreamParams.Values()
- [x] OpenInterestEventStreamParams.Values()

Market Streams (type `MarketStreamChannel`, key: `marketStream`)
- [x] NewMarketStreamChannel(client *Client) *MarketStreamChannel
- [x] Connect(ctx context.Context, streamName string) error
- [x] Disconnect(ctx context.Context) error
- [x] Subscribe(ctx context.Context, req *models.SubscribeRequest, cb *func(context.Context, *models.SubscribeResponse, error) error) error
- [x] Unsubscribe(ctx context.Context, req *models.UnsubscribeRequest, cb *func(context.Context, *models.UnsubscribeResponse, error) error) error
- [x] ListSubscriptions(ctx context.Context, req *models.ListSubscriptionsRequest, cb *func(context.Context, *models.ListSubscriptionsResponse, error) error) error
- [x] SetProperty(ctx context.Context, req *models.SetPropertyRequest, cb *func(context.Context, *models.SetPropertyResponse, error) error) error
- [x] GetProperty(ctx context.Context, req *models.GetPropertyRequest, cb *func(context.Context, *models.GetPropertyResponse, error) error) error
- [x] HandleNewSymbolInfoEvent(fn func(context.Context, *models.NewSymbolInfoEvent) error)
- [x] HandleOpenInterestEvent(fn func(context.Context, *models.OpenInterestEvent) error)
- [x] HandleMarkPriceEvent(fn func(context.Context, *models.MarkPriceEvent) error)
- [x] HandleKlineEvent(fn func(context.Context, *models.KlineEvent) error)
- [x] HandleTickerByUnderlyingEvent(fn func(context.Context, *models.TickerByUnderlyingEvent) error)
- [x] HandleIndexPriceEvent(fn func(context.Context, *models.IndexPriceEvent) error)
- [x] HandleTickerEvent(fn func(context.Context, *models.TickerEvent) error)
- [x] HandleTradeEvent(fn func(context.Context, *models.TradeEvent) error)
- [x] HandlePartialDepthEvent(fn func(context.Context, *models.PartialDepthEvent) error)

Combined Market Streams (type `CombinedMarketStreamChannel`, key: `combinedMarketStream`)
- [x] NewCombinedMarketStreamChannel(client *Client) *CombinedMarketStreamChannel
- [x] Connect(ctx context.Context, streams string) error
- [x] Disconnect(ctx context.Context) error
- [x] Subscribe(ctx context.Context, req *models.SubscribeRequest, cb *func(context.Context, *models.SubscribeResponse, error) error) error
- [x] Unsubscribe(ctx context.Context, req *models.UnsubscribeRequest, cb *func(context.Context, *models.UnsubscribeResponse, error) error) error
- [x] ListSubscriptions(ctx context.Context, req *models.ListSubscriptionsRequest, cb *func(context.Context, *models.ListSubscriptionsResponse, error) error) error
- [x] SetProperty(ctx context.Context, req *models.SetPropertyRequest, cb *func(context.Context, *models.SetPropertyResponse, error) error) error
- [x] GetProperty(ctx context.Context, req *models.GetPropertyRequest, cb *func(context.Context, *models.GetPropertyResponse, error) error) error
- [x] HandleCombinedMarketStreamEvent(fn func(context.Context, *models.CombinedMarketStreamEvent) error)
- [x] HandleNewSymbolInfoEvent(fn func(context.Context, *models.NewSymbolInfoEvent) error)
- [x] HandleOpenInterestEvent(fn func(context.Context, *models.OpenInterestEvent) error)
- [x] HandleMarkPriceEvent(fn func(context.Context, *models.MarkPriceEvent) error)
- [x] HandleKlineEvent(fn func(context.Context, *models.KlineEvent) error)
- [x] HandleTickerByUnderlyingEvent(fn func(context.Context, *models.TickerByUnderlyingEvent) error)
- [x] HandleIndexPriceEvent(fn func(context.Context, *models.IndexPriceEvent) error)
- [x] HandleTickerEvent(fn func(context.Context, *models.TickerEvent) error)
- [x] HandleTradeEvent(fn func(context.Context, *models.TradeEvent) error)
- [x] HandlePartialDepthEvent(fn func(context.Context, *models.PartialDepthEvent) error)

User Data Streams (type `UserDataStreamChannel`, key: `userDataStream`)
- [x] NewUserDataStreamChannel(client *Client) *UserDataStreamChannel
- [x] Connect(ctx context.Context, listenKey string) error
- [x] Disconnect(ctx context.Context) error
- [x] HandleAccountUpdateEvent(fn func(context.Context, *models.AccountUpdateEvent) error)
- [x] HandleOrderTradeUpdateEvent(fn func(context.Context, *models.OrderTradeUpdateEvent) error)
- [x] HandleRiskLevelChangeEvent(fn func(context.Context, *models.RiskLevelChangeEvent) error)

Validation Approach
- Request/response: assert `id` echo, presence and shape of result arrays, and error messages when present.
- Events: assert `e` types, symbol formats, timestamps recency, and numeric fields; optional REST cross‑checks for index/mark/ticker when `ENABLE_REST_VALIDATION=1`.
- Robustness: suites fail if the SDK logs any `unhandled message:` during execution.

Known Behaviors
- `SET_PROPERTY`/`GET_PROPERTY` may not ACK in all environments; tests treat timeouts as informational when they occur under short deadlines.
- Options market activity is variable; suites use graceful timeouts to avoid false negatives when streams are quiet.
- Latest SDK drop removed the standalone `HandleErrorMessage` hooks from market/combined channels; tests assert error envelopes via per-request callbacks instead.

Next Steps
- Add coverage for core `Client` methods (server CRUD, `RegisterHandlers`, `Wait`).
- Cover ServerManager helpers and URL resolution.
- Add WS auth/signing coverage if/when private WS messages require it.
- Expand negative/edge‑case tests (bad params, unsubscribed list checks, etc.).
