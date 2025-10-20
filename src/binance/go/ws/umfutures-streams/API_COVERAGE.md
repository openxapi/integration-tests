Binance UM Futures WS SDK — Integration Coverage

SDK: `github.com/openxapi/binance-go/ws/umfutures-streams`
REST: `github.com/openxapi/binance-go/rest/umfutures`

Notes
- Request methods accept a per-request callback pointer (e.g., `*func(context.Context, *models.SubscribeResponse) error`). Tests pass concrete callbacks and assert ids/fields where ACKs are returned. Some property methods may not ACK on all servers; timeouts are tolerated when reasonable.
- One IntegrationTestSuite per channel: Market, Combined. Field-level assertions are performed for each event type where market activity exists.
- For Market and Combined suites we force `mainnet1` to increase event availability; default behavior remains testnet for other tests.

Coverage Summary
- MarketStreamsChannel: requests (Subscribe, Unsubscribe, List, SetProperty, GetProperty) covered. Event handlers validated: aggTrade, markPrice, kline, continuousKline, ticker, bookTicker, partialDepth, diffDepth, allMarkPrices, allMiniTickers, allTickers, allBookTickers, allLiquidations, compositeIndex, assetIndex, allAssetIndexes. MiniTicker and contractInfo are listed below for completeness (see status).
- CombinedMarketStreamsChannel: requests covered; wrapper `HandleCombinedMarketStreamsEvent` installed. Event handlers validated: markPrice, kline, allTickers, allMiniTickers, allBookTickers, allMarkPrices, allLiquidations, compositeIndex, contractInfo, allAssetIndexes. Additional single‑symbol handlers are listed below (see status).
 - UserDataStreamsChannel: requests (Start, Ping, Stop) exercised with graceful tolerance on unsupported ACKs; channel Connect/Disconnect validated using REST listenKey on testnet. Event handlers installed; ORDER_TRADE_UPDATE and ACCOUNT_CONFIG_UPDATE triggered via REST and validated when account/permissions allow. Other events registered and tolerated if not emitted during runs.
- Stream builders and typed params: covered for the streams used in tests.

Client / Server
- [x] NewClient() *Client
- [x] GetActiveServer() *ServerInfo (logged in tests)
- [x] SetActiveServer("mainnet1") in Market/Combined suites; default testnet validated in a separate test

Stream Builders & Typed Params (subset used by tests)
- [x] BuildAggregateTradeEventStream
- [x] BuildMarkPriceEventStream
- [x] BuildKlineEventStream
- [x] BuildTickerEventStream
- [x] BuildBookTickerEventStream
- [x] BuildPartialDepthEventStream
- [x] BuildDiffDepthEventStream

Market Streams (type `MarketStreamsChannel`, key: `marketStreams`)
- [x] NewMarketStreamsChannel(client *Client) *MarketStreamsChannel
- [x] Connect(ctx, streamName string) / Disconnect(ctx)
- [x] Subscribe(ctx, req, cb)
- [x] Unsubscribe(ctx, req, cb)
- [x] ListSubscriptions(ctx, req, cb)
- [x] SetProperty(ctx, req, cb) (timeout tolerated)
- [x] GetProperty(ctx, req, cb) (optional ACK)
- [x] HandleAggregateTradeEvent / HandleMarkPriceEvent / HandleKlineEvent / HandleTickerEvent / HandleBookTickerEvent / HandlePartialDepthEvent / HandleDiffDepthEvent / HandleErrorMessage

Combined Market Streams (type `CombinedMarketStreamsChannel`, key: `combinedMarketStreams`)
- [x] NewCombinedMarketStreamsChannel(client *Client) *CombinedMarketStreamsChannel
- [x] Connect(ctx, streams string) / Disconnect(ctx)
- [x] Subscribe(ctx, req, cb)
- [x] Unsubscribe(ctx, req, cb)
- [x] ListSubscriptions(ctx, req, cb)
- [x] SetProperty(ctx, req, cb) (timeout tolerated)
- [x] GetProperty(ctx, req, cb) (optional ACK)
- Handlers (Complete List):
  - [x] HandleCombinedMarketStreamsEvent (wrapper)
  - [x] HandleErrorMessage
  - [x] HandleAggregateTradeEvent
  - [x] HandleMarkPriceEvent
  - [x] HandleAllMarkPricesEvent
  - [x] HandleKlineEvent
  - [x] HandleContinuousKlineEvent
  - [x] HandleMiniTickerEvent
  - [x] HandleAllMiniTickersEvent
  - [x] HandleTickerEvent
  - [x] HandleAllTickersEvent
  - [x] HandleBookTickerEvent
  - [x] HandleAllBookTickersEvent
  - [x] HandleLiquidationEvent (single‑symbol; tolerant to sparse events)
  - [x] HandleAllLiquidationsEvent (updated tests validate payload)
  - [x] HandlePartialDepthEvent
  - [x] HandleDiffDepthEvent
  - [x] HandleCompositeIndexEvent
  - [x] HandleContractInfoEvent
  - [x] HandleAssetIndexEvent
  - [x] HandleAllAssetIndexesEvent

Handlers By Channel (MarketStreamsChannel)
- [x] HandleErrorMessage
- [x] HandleAggregateTradeEvent
- [x] HandleMarkPriceEvent
- [x] HandleAllMarkPricesEvent
- [x] HandleKlineEvent
- [x] HandleContinuousKlineEvent
- [x] HandleMiniTickerEvent
- [x] HandleAllMiniTickersEvent
- [x] HandleTickerEvent
- [x] HandleAllTickersEvent
- [x] HandleBookTickerEvent
- [x] HandleAllBookTickersEvent
- [x] HandleLiquidationEvent (single‑symbol; tolerant to sparse events)
- [x] HandleAllLiquidationsEvent (updated tests validate payload)
- [x] HandlePartialDepthEvent
- [x] HandleDiffDepthEvent
- [x] HandleCompositeIndexEvent
- [ ] HandleContractInfoEvent (available in Combined suite; pending in Market)
- [x] HandleAssetIndexEvent
- [x] HandleAllAssetIndexesEvent

Notes
- All HandleXxx methods are registered in at least one suite; most have focused subscription tests. Single‑symbol events with sparse traffic (MiniTicker, Liquidation, BookTicker, Partial/Diff Depth in combined) are marked pending for targeted subscribe/validate blocks.
- AllLiquidationsEvent tests updated to assert eventType, timestamps, and core fields in `o` (order) payload.

Next Steps
- Add tests for additional market events: compositeIndex, contractInfo, all-@arr streams, liquidation events where feasible.
- Expand UserData coverage: trigger MARGIN_CALL (simulated), TRADE_LITE, STRATEGY_UPDATE, GRID_UPDATE, CONDITIONAL_ORDER_TRIGGER_REJECT when feasible on testnet accounts; add ListenKey expiry scenario.
