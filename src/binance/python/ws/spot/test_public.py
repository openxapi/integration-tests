"""
Integration tests for Binance Spot WebSocket API public endpoints.

These tests cover all public endpoints that don't require authentication.
Updated to work with the new SDK **params pattern.
"""

import pytest
import pytest_asyncio
from conftest import setup_client_with_config, run_endpoint_test, BinanceTestConfig, AuthType, KeyType


def get_result_field(result_model, field_name, default=None):
    """Safely get field from result model, handling both defined fields and model_extra."""
    if hasattr(result_model, field_name):
        return getattr(result_model, field_name)
    if hasattr(result_model, 'model_extra') and result_model.model_extra:
        return result_model.model_extra.get(field_name, default)
    return default


def assert_field_exists(result_model, field_name, message=None):
    """Assert that a field exists in result model."""
    if message is None:
        message = f"Result should contain {field_name}"
    
    has_attr = hasattr(result_model, field_name)
    has_extra = (hasattr(result_model, 'model_extra') and 
                result_model.model_extra and 
                field_name in result_model.model_extra)
    
    assert has_attr or has_extra, message


class TestPublicEndpoints:
    """Test class for public WebSocket API endpoints."""
    
    @pytest.mark.public
    @pytest_asyncio.fixture(autouse=True)
    async def setup(self):
        """Setup for public endpoint tests."""
        self.config = BinanceTestConfig(
            name="Public-NoAuth",
            description="Test public endpoints that don't require authentication",
            auth_type=AuthType.NONE,
            key_type=KeyType.HMAC,
            api_key=None,
            secret_key=None
        )
        
        yield
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ping(self, websocket_client):
        """Test ping endpoint for connectivity."""
        
        async def test_ping_request():
            from binance.ws.spot.models.ping_request import PingRequest
            request = PingRequest()
            response = await websocket_client.ping(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation (ping can have null status for successful connectivity)
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Validate response structure
            assert hasattr(response, 'result'), "Response should have result attribute"
            assert hasattr(response, 'id'), "Response should have id attribute"
            
            # For ping, result should exist but can be empty
            # This validates the connection was successful
            return response
        
        result = await run_endpoint_test(
            websocket_client, 
            "ping",
            test_ping_request,
            timeout=5.0
        )
        assert result["success"], f"Ping test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_server_time(self, websocket_client):
        """Test server time endpoint."""
        
        async def test_time_request():
            from binance.ws.spot.models.time_request import TimeRequest
            request = TimeRequest()
            response = await websocket_client.time(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            # response.result is now a Pydantic model, not a dict
            from binance.ws.spot.models.time_response import TimeResponseResult
            assert isinstance(response.result, TimeResponseResult), "Result should be a TimeResponseResult model"
            
            # Server time validation
            assert hasattr(response.result, 'server_time'), "Result should have server_time attribute"
            server_time = response.result.server_time
            assert server_time is not None, "Server time should not be null"
            assert isinstance(server_time, int), "Server time should be an integer timestamp"
            assert server_time > 0, "Server time should be positive"
            
            # Validate timestamp is reasonable (within 5 minutes of current time)
            import time
            current_time_ms = int(time.time() * 1000)
            time_diff = abs(current_time_ms - server_time)
            assert time_diff < 5 * 60 * 1000, f"Server time differs too much from local time: {time_diff}ms"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "server_time",
            test_time_request,
            timeout=5.0
        )
        assert result["success"], f"Server time test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_exchange_info(self, websocket_client):
        """Test exchange information endpoint."""
        
        async def test_exchange_info_request():
            from binance.ws.spot.models.exchange_info_request import ExchangeInfoRequest
            request = ExchangeInfoRequest()
            response = await websocket_client.exchangeinfo(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            from binance.ws.spot.models.exchange_info_response import ExchangeInfoResponseResult
            assert isinstance(response.result, ExchangeInfoResponseResult), "Result should be an ExchangeInfoResponseResult model"
            
            # Exchange info specific validations
            result = response.result
            
            # Timezone validation
            assert_field_exists(result, 'timezone', "Result should contain timezone")
            timezone = get_result_field(result, 'timezone')
            assert timezone is not None, "Timezone should not be null"
            assert isinstance(timezone, str), "Timezone should be a string"
            assert timezone != "", "Timezone should not be empty"
            
            # Server time validation
            assert_field_exists(result, 'server_time', "Result should contain server_time")
            server_time = get_result_field(result, 'server_time')
            assert server_time is not None, "Server time should not be null"
            assert isinstance(server_time, int), "Server time should be an integer"
            assert server_time > 0, "Server time should be positive"
            
            # Symbols validation
            assert_field_exists(result, 'symbols', "Result should contain symbols")
            symbols = get_result_field(result, 'symbols')
            assert symbols is not None, "Symbols should not be null"
            assert isinstance(symbols, list), "Symbols should be a list"
            assert len(symbols) > 0, "Should have at least one symbol"
            
            # Validate first symbol structure - now properly deserialized as Pydantic models
            first_symbol = symbols[0]
            # Symbols are now properly deserialized as ExchangeInfoResponseSymbolsItem objects
            from binance.ws.spot.models.exchange_info_response import ExchangeInfoResponseSymbolsItem
            assert isinstance(first_symbol, ExchangeInfoResponseSymbolsItem), "Symbol should be an ExchangeInfoResponseSymbolsItem model"
            assert hasattr(first_symbol, 'symbol'), "Symbol should have symbol field"
            assert hasattr(first_symbol, 'status'), "Symbol should have status field"
            assert hasattr(first_symbol, 'base_asset'), "Symbol should have base_asset field"
            assert hasattr(first_symbol, 'quote_asset'), "Symbol should have quote_asset field"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "exchange_info",
            test_exchange_info_request,
            timeout=10.0
        )
        assert result["success"], f"Exchange info test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_depth(self, websocket_client, test_symbol):
        """Test order book depth endpoint."""
        
        async def test_depth_request():
            from binance.ws.spot.models.depth_request import DepthRequest
            request = DepthRequest(symbol=test_symbol, limit=10)
            response = await websocket_client.depth(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            from binance.ws.spot.models.depth_response import DepthResponseResult
            assert isinstance(response.result, DepthResponseResult), "Result should be a DepthResponseResult model"
            
            # Order book specific validations
            result = response.result
            
            # Last update ID validation
            assert_field_exists(result, 'last_update_id', "Result should contain last_update_id")
            last_update_id = get_result_field(result, 'last_update_id')
            assert last_update_id is not None, "Last update ID should not be null"
            assert isinstance(last_update_id, int), "Last update ID should be an integer"
            assert last_update_id > 0, "Last update ID should be positive"
            
            # Bids validation
            assert_field_exists(result, 'bids', "Result should contain bids")
            bids = get_result_field(result, 'bids')
            assert bids is not None, "Bids should not be null"
            assert isinstance(bids, list), "Bids should be a list"
            assert len(bids) <= 10, f"Should have at most 10 bids, got {len(bids)}"
            
            # Validate bid structure if bids exist
            if len(bids) > 0:
                first_bid = bids[0]
                assert isinstance(first_bid, list), "Bid should be a list [price, quantity]"
                assert len(first_bid) >= 2, "Bid should have at least price and quantity"
                # Validate price and quantity are numeric strings
                price, quantity = first_bid[0], first_bid[1]
                assert isinstance(price, str), "Bid price should be a string"
                assert isinstance(quantity, str), "Bid quantity should be a string"
                assert float(price) > 0, "Bid price should be positive"
                assert float(quantity) > 0, "Bid quantity should be positive"
            
            # Asks validation
            assert_field_exists(result, 'asks', "Result should contain asks")
            asks = get_result_field(result, 'asks')
            assert asks is not None, "Asks should not be null"
            assert isinstance(asks, list), "Asks should be a list"
            assert len(asks) <= 10, f"Should have at most 10 asks, got {len(asks)}"
            
            # Validate ask structure if asks exist
            if len(asks) > 0:
                first_ask = asks[0]
                assert isinstance(first_ask, list), "Ask should be a list [price, quantity]"
                assert len(first_ask) >= 2, "Ask should have at least price and quantity"
                # Validate price and quantity are numeric strings
                price, quantity = first_ask[0], first_ask[1]
                assert isinstance(price, str), "Ask price should be a string"
                assert isinstance(quantity, str), "Ask quantity should be a string"
                assert float(price) > 0, "Ask price should be positive"
                assert float(quantity) > 0, "Ask quantity should be positive"
            
            # Validate market structure (asks should be higher than bids)
            if len(bids) > 0 and len(asks) > 0:
                highest_bid = float(bids[0][0])
                lowest_ask = float(asks[0][0])
                assert lowest_ask > highest_bid, f"Lowest ask ({lowest_ask}) should be higher than highest bid ({highest_bid})"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "depth",
            test_depth_request,
            timeout=5.0
        )
        assert result["success"], f"Depth test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_klines(self, websocket_client, test_symbol):
        """Test klines (candlestick) endpoint."""
        
        async def test_klines_request():
            from binance.ws.spot.models.klines_request import KlinesRequest
            request = KlinesRequest(symbol=test_symbol, interval="1m", limit=10)
            response = await websocket_client.klines(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation - klines result is directly a List[List[Any]]
            assert response.result is not None, "Response should have result field"
            assert isinstance(response.result, list), "Result should be a list of klines"
            
            # The result is directly the klines data
            klines_data = response.result
            
            assert isinstance(klines_data, list), "Klines data should be a list"
            assert len(klines_data) <= 10, f"Should have at most 10 klines, got {len(klines_data)}"
            
            # Validate kline structure if klines exist
            if len(klines_data) > 0:
                first_kline = klines_data[0]
                assert isinstance(first_kline, list), "Kline should be a list"
                assert len(first_kline) >= 12, f"Kline should have at least 12 fields, got {len(first_kline)}"
                
                # Validate kline fields: [openTime, open, high, low, close, volume, closeTime, quoteVolume, count, takerBuyBaseVolume, takerBuyQuoteVolume, ignore]
                open_time, open_price, high, low, close, volume = first_kline[:6]
                close_time = first_kline[6]
                
                # Time validation
                assert isinstance(open_time, int), "Open time should be an integer timestamp"
                assert isinstance(close_time, int), "Close time should be an integer timestamp"
                assert open_time > 0, "Open time should be positive"
                assert close_time > 0, "Close time should be positive"
                assert close_time > open_time, "Close time should be after open time"
                
                # Price validation
                assert isinstance(open_price, str), "Open price should be a string"
                assert isinstance(high, str), "High price should be a string"
                assert isinstance(low, str), "Low price should be a string"
                assert isinstance(close, str), "Close price should be a string"
                
                open_val = float(open_price)
                high_val = float(high)
                low_val = float(low)
                close_val = float(close)
                
                assert open_val > 0, "Open price should be positive"
                assert high_val > 0, "High price should be positive"
                assert low_val > 0, "Low price should be positive"
                assert close_val > 0, "Close price should be positive"
                
                # OHLC relationship validation
                assert high_val >= max(open_val, close_val), "High should be >= max(open, close)"
                assert low_val <= min(open_val, close_val), "Low should be <= min(open, close)"
                
                # Volume validation
                assert isinstance(volume, str), "Volume should be a string"
                volume_val = float(volume)
                assert volume_val >= 0, "Volume should be non-negative"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "klines",
            test_klines_request,
            timeout=5.0
        )
        assert result["success"], f"Klines test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ui_klines(self, websocket_client, test_symbol):
        """Test UI klines endpoint."""
        
        async def test_ui_klines_request():
            from binance.ws.spot.models.ui_klines_request import UiKlinesRequest
            request = UiKlinesRequest(symbol=test_symbol, interval="1m", limit=10)
            response = await websocket_client.uiklines(request)
            
            # Basic response validation (same as klines)
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation - UI klines result is directly a List[List[Any]]
            assert response.result is not None, "Response should have result field"
            assert isinstance(response.result, list), "Result should be a list of UI klines"
            
            # The result is directly the UI klines data
            klines_data = response.result
            
            assert isinstance(klines_data, list), "UI klines data should be a list"
            assert len(klines_data) <= 10, f"Should have at most 10 UI klines, got {len(klines_data)}"
            
            # UI klines have same structure as regular klines
            if len(klines_data) > 0:
                first_kline = klines_data[0]
                assert isinstance(first_kline, list), "UI kline should be a list"
                assert len(first_kline) >= 12, f"UI kline should have at least 12 fields, got {len(first_kline)}"
                
                # Basic validation of key fields
                open_time, open_price, high, low, close = first_kline[:5]
                assert isinstance(open_time, int) and open_time > 0, "Open time should be positive integer"
                assert float(open_price) > 0, "Open price should be positive"
                assert float(high) >= float(open_price), "High should be >= open"
                assert float(low) <= float(open_price), "Low should be <= open"
                assert float(close) > 0, "Close price should be positive"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ui_klines",
            test_ui_klines_request,
            timeout=5.0
        )
        assert result["success"], f"UI Klines test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ticker_24hr(self, websocket_client, test_symbol):
        """Test 24hr ticker statistics endpoint."""
        
        async def test_ticker_24hr_request():
            from binance.ws.spot.models.ticker24hr_request import Ticker24hrRequest
            request = Ticker24hrRequest(symbol=test_symbol)
            response = await websocket_client.ticker_24hr(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            # Result is now a Pydantic model, not a dict
            from binance.ws.spot.models.ticker24hr_response import Ticker24hrResponseResult
            assert isinstance(response.result, Ticker24hrResponseResult), "Result should be a Ticker24hrResponseResult model"
            
            # 24hr ticker specific validations
            result = response.result
            
            # Symbol validation
            assert hasattr(result, 'symbol'), "Result should have symbol attribute"
            assert result.symbol == test_symbol, f"Expected symbol {test_symbol}, got {result.symbol}"
            
            # Price fields validation - use actual field names from the model
            price_fields = ['open_price', 'high_price', 'low_price', 'last_price', 'prev_close_price']
            for field in price_fields:
                assert hasattr(result, field), f"Result should have {field} attribute"
                price = getattr(result, field)
                assert price is not None, f"{field} should not be null"
                assert isinstance(price, str), f"{field} should be a string"
                assert float(price) > 0, f"{field} should be positive"
            
            # Volume fields validation
            volume_fields = ['volume', 'quote_volume']
            for field in volume_fields:
                assert hasattr(result, field), f"Result should have {field} attribute"
                volume = getattr(result, field)
                assert volume is not None, f"{field} should not be null"
                assert isinstance(volume, str), f"{field} should be a string"
                assert float(volume) >= 0, f"{field} should be non-negative"
            
            # Count validation
            assert hasattr(result, 'count'), "Result should have count attribute"
            count = result.count
            assert count is not None, "Count should not be null"
            assert isinstance(count, int), "Count should be an integer"
            assert count >= 0, "Count should be non-negative"
            
            # Price change validation
            if hasattr(result, 'price_change') and hasattr(result, 'price_change_percent'):
                price_change = result.price_change
                price_change_percent = result.price_change_percent
                assert isinstance(price_change, str), "Price change should be a string"
                assert isinstance(price_change_percent, str), "Price change percent should be a string"
            
            # Validate OHLC relationship
            open_price = float(result.open_price)
            high_price = float(result.high_price)
            low_price = float(result.low_price)
            last_price = float(result.last_price)
            
            assert high_price >= max(open_price, last_price), "High should be >= max(open, last)"
            assert low_price <= min(open_price, last_price), "Low should be <= min(open, last)"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ticker_24hr",
            test_ticker_24hr_request,
            timeout=5.0
        )
        assert result["success"], f"24hr ticker test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ticker_price(self, websocket_client, test_symbol):
        """Test ticker price endpoint."""
        
        async def test_ticker_price_request():
            from binance.ws.spot.models.ticker_price_request import TickerPriceRequest
            request = TickerPriceRequest(symbol=test_symbol)
            response = await websocket_client.ticker_price(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            from binance.ws.spot.models.ticker_price_response import TickerPriceResponseResult
            assert isinstance(response.result, TickerPriceResponseResult), "Result should be a TickerPriceResponseResult model"
            
            # Ticker price specific validations
            result = response.result
            
            # Symbol validation - this field is defined in the model
            symbol = result.symbol
            assert symbol is not None, "Symbol should not be null"
            assert isinstance(symbol, str), "Symbol should be a string"
            assert symbol == test_symbol, f"Expected symbol {test_symbol}, got {symbol}"
            
            # Price validation - this field is defined in the model
            price = result.price
            assert price is not None, "Price should not be null"
            assert isinstance(price, str), "Price should be a string"
            price_val = float(price)
            assert price_val > 0, "Price should be positive"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ticker_price",
            test_ticker_price_request,
            timeout=5.0
        )
        assert result["success"], f"Ticker price test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ticker_book(self, websocket_client, test_symbol):
        """Test book ticker endpoint."""
        
        async def test_ticker_book_request():
            from binance.ws.spot.models.ticker_book_request import TickerBookRequest
            request = TickerBookRequest(symbol=test_symbol)
            response = await websocket_client.ticker_book(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            assert response.result is not None, "Response should have result field"
            # Result is now a Pydantic model, not a dict
            from binance.ws.spot.models.ticker_book_response import TickerBookResponseResult
            assert isinstance(response.result, TickerBookResponseResult), "Result should be a TickerBookResponseResult model"
            
            # Book ticker specific validations
            result = response.result
            
            # Symbol validation
            assert hasattr(result, 'symbol'), "Result should have symbol attribute"
            assert result.symbol == test_symbol, f"Expected symbol {test_symbol}, got {result.symbol}"
            
            # Best bid validation
            assert hasattr(result, 'bid_price'), "Result should have bid_price attribute"
            assert hasattr(result, 'bid_qty'), "Result should have bid_qty attribute"
            bid_price = result.bid_price
            bid_qty = result.bid_qty
            assert isinstance(bid_price, str), "Bid price should be a string"
            assert isinstance(bid_qty, str), "Bid quantity should be a string"
            assert float(bid_price) > 0, "Bid price should be positive"
            assert float(bid_qty) > 0, "Bid quantity should be positive"
            
            # Best ask validation
            assert hasattr(result, 'ask_price'), "Result should have ask_price attribute"
            assert hasattr(result, 'ask_qty'), "Result should have ask_qty attribute"
            ask_price = result.ask_price
            ask_qty = result.ask_qty
            assert isinstance(ask_price, str), "Ask price should be a string"
            assert isinstance(ask_qty, str), "Ask quantity should be a string"
            assert float(ask_price) > 0, "Ask price should be positive"
            assert float(ask_qty) > 0, "Ask quantity should be positive"
            
            # Market structure validation
            assert float(ask_price) > float(bid_price), f"Ask price ({ask_price}) should be higher than bid price ({bid_price})"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ticker_book",
            test_ticker_book_request,
            timeout=5.0
        )
        assert result["success"], f"Book ticker test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ticker_trading_day(self, websocket_client, test_symbol):
        """Test trading day ticker endpoint."""
        
        async def test_trading_day_request():
            from binance.ws.spot.models.ticker_trading_day_request import TickerTradingDayRequest
            request = TickerTradingDayRequest(symbol=test_symbol)
            response = await websocket_client.ticker_tradingday(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            assert response.result is not None, "Response should have result field"
            # Result is now a Pydantic model, not a dict
            from binance.ws.spot.models.ticker_trading_day_response import TickerTradingDayResponseResult
            assert isinstance(response.result, TickerTradingDayResponseResult), "Result should be a TickerTradingDayResponseResult model"
            
            # Trading day ticker specific validations
            result = response.result
            
            # Symbol validation
            assert hasattr(result, 'symbol'), "Result should have symbol attribute"
            assert result.symbol == test_symbol, f"Expected symbol {test_symbol}, got {result.symbol}"
            
            # Price fields validation
            price_fields = ['openPrice', 'highPrice', 'lowPrice', 'lastPrice']
            for field in price_fields:
                if hasattr(result, field):
                    price = getattr(result, field)
                    assert isinstance(price, str), f"{field} should be a string"
                    assert float(price) > 0, f"{field} should be positive"
            
            # Volume validation
            if hasattr(result, 'volume'):
                volume = result.volume
                assert isinstance(volume, str), "Volume should be a string"
                assert float(volume) >= 0, "Volume should be non-negative"
            
            # Count validation
            if hasattr(result, 'count'):
                count = result.count
                assert isinstance(count, int), "Count should be an integer"
                assert count >= 0, "Count should be non-negative"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ticker_trading_day",
            test_trading_day_request,
            timeout=5.0
        )
        assert result["success"], f"Trading day ticker test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_ticker(self, websocket_client, test_symbol):
        """Test rolling window ticker endpoint."""
        
        async def test_ticker_request():
            from binance.ws.spot.models.ticker_request import TickerRequest
            request = TickerRequest(symbol=test_symbol, windowSize="1d")
            response = await websocket_client.ticker(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            assert response.result is not None, "Response should have result field"
            from binance.ws.spot.models.ticker_response import TickerResponseResult
            assert isinstance(response.result, TickerResponseResult), "Result should be a TickerResponseResult model"
            
            # Rolling window ticker validations
            result = response.result
            
            # Symbol validation
            assert hasattr(result, 'symbol'), "Result should have symbol attribute"
            assert result.symbol == test_symbol, f"Expected symbol {test_symbol}, got {result.symbol}"
            
            # Price change validation
            if hasattr(result, 'price_change') and result.price_change is not None:
                price_change = result.price_change
                assert isinstance(price_change, str), "Price change should be a string"
            
            if hasattr(result, 'price_change_percent') and result.price_change_percent is not None:
                price_change_percent = result.price_change_percent
                assert isinstance(price_change_percent, str), "Price change percent should be a string"
            
            # Open/close price validation
            if (hasattr(result, 'open_price') and result.open_price is not None and 
                hasattr(result, 'last_price') and result.last_price is not None):
                open_price = float(result.open_price)
                last_price = float(result.last_price)
                assert open_price > 0, "Open price should be positive"
                assert last_price > 0, "Last price should be positive"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "ticker",
            test_ticker_request,
            timeout=5.0
        )
        assert result["success"], f"Ticker test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_avg_price(self, websocket_client, test_symbol):
        """Test average price endpoint."""
        
        async def test_avg_price_request():
            from binance.ws.spot.models.avg_price_request import AvgPriceRequest
            request = AvgPriceRequest(symbol=test_symbol)
            response = await websocket_client.avgprice(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            assert response.result is not None, "Response should have result field"
            from binance.ws.spot.models.avg_price_response import AvgPriceResponseResult
            assert isinstance(response.result, AvgPriceResponseResult), "Result should be an AvgPriceResponseResult model"
            
            # Average price specific validations
            result = response.result
            
            # Minutes validation
            assert hasattr(result, 'mins'), "Result should have mins attribute"
            mins = result.mins
            assert mins is not None, "Minutes should not be null"
            assert isinstance(mins, int), "Minutes should be an integer"
            assert mins > 0, "Minutes should be positive"
            
            # Price validation
            assert hasattr(result, 'price'), "Result should have price attribute"
            price = result.price
            assert price is not None, "Price should not be null"  
            assert isinstance(price, str), "Price should be a string"
            assert float(price) > 0, "Price should be positive"
            
            # Close time validation
            if hasattr(result, 'close_time') and result.close_time is not None:
                close_time = result.close_time
                assert isinstance(close_time, int), "Close time should be an integer"
                assert close_time > 0, "Close time should be positive"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "avg_price",
            test_avg_price_request,
            timeout=5.0
        )
        assert result["success"], f"Average price test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_trades_aggregate(self, websocket_client, test_symbol):
        """Test aggregate trades endpoint."""
        
        async def test_trades_aggregate_request():
            from binance.ws.spot.models.trades_aggregate_request import TradesAggregateRequest
            request = TradesAggregateRequest(symbol=test_symbol, limit=10)
            response = await websocket_client.trades_aggregate(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            # Status validation
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            # Result structure validation
            assert response.result is not None, "Response should have result field"
            assert isinstance(response.result, list), "Result should be a list of trades"
            assert len(response.result) <= 10, f"Should have at most 10 trades, got {len(response.result)}"
            
            # Validate trade structure if trades exist
            if len(response.result) > 0:
                first_trade = response.result[0]
                assert isinstance(first_trade, dict), "Trade should be a dictionary"
                
                # Trade ID validation
                assert 'a' in first_trade, "Trade should contain aggregate trade ID (a)"
                trade_id = first_trade['a']
                assert trade_id is not None, "Trade ID should not be null"
                assert isinstance(trade_id, int), "Trade ID should be an integer"
                assert trade_id > 0, "Trade ID should be positive"
                
                # Price validation
                assert 'p' in first_trade, "Trade should contain price (p)"
                price = first_trade['p']
                assert price is not None, "Price should not be null"
                assert isinstance(price, str), "Price should be a string"
                assert float(price) > 0, "Price should be positive"
                
                # Quantity validation
                assert 'q' in first_trade, "Trade should contain quantity (q)"
                quantity = first_trade['q']
                assert quantity is not None, "Quantity should not be null"
                assert isinstance(quantity, str), "Quantity should be a string"
                assert float(quantity) > 0, "Quantity should be positive"
                
                # Timestamp validation
                assert 'T' in first_trade, "Trade should contain timestamp (T)"
                timestamp = first_trade['T']
                assert timestamp is not None, "Timestamp should not be null"
                assert isinstance(timestamp, int), "Timestamp should be an integer"
                assert timestamp > 0, "Timestamp should be positive"
                
                # Buyer maker validation
                assert 'm' in first_trade, "Trade should contain buyer maker flag (m)"
                buyer_maker = first_trade['m']
                assert buyer_maker is not None, "Buyer maker should not be null"
                assert isinstance(buyer_maker, bool), "Buyer maker should be a boolean"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "trades_aggregate",
            test_trades_aggregate_request,
            timeout=5.0
        )
        assert result["success"], f"Aggregate trades test failed: {result.get('error')}"
    
    @pytest.mark.public
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_trades_historical(self, websocket_client, test_symbol):
        """Test historical trades endpoint."""
        
        async def test_trades_historical_request():
            from binance.ws.spot.models.trades_historical_request import TradesHistoricalRequest
            request = TradesHistoricalRequest(symbol=test_symbol, limit=10)
            response = await websocket_client.trades_historical(request)
            
            # Basic response validation
            assert response is not None, "Response should not be None"
            
            if response.status is not None:
                assert response.status == 200, f"Expected status 200, got {response.status}"
            
            assert response.result is not None, "Response should have result field"
            assert isinstance(response.result, list), "Result should be a list of historical trades"
            assert len(response.result) <= 10, f"Should have at most 10 trades, got {len(response.result)}"
            
            # Validate historical trade structure if trades exist
            if len(response.result) > 0:
                first_trade = response.result[0]
                assert isinstance(first_trade, dict), "Historical trade should be a dictionary"
                
                # Trade ID validation
                assert 'id' in first_trade, "Trade should contain ID"
                trade_id = first_trade['id']
                assert trade_id is not None, "Trade ID should not be null"
                assert isinstance(trade_id, int), "Trade ID should be an integer"
                assert trade_id > 0, "Trade ID should be positive"
                
                # Price validation
                assert 'price' in first_trade, "Trade should contain price"
                price = first_trade['price']
                assert price is not None, "Price should not be null"
                assert isinstance(price, str), "Price should be a string"
                assert float(price) > 0, "Price should be positive"
                
                # Quantity validation
                assert 'qty' in first_trade, "Trade should contain quantity"
                quantity = first_trade['qty']
                assert quantity is not None, "Quantity should not be null"
                assert isinstance(quantity, str), "Quantity should be a string"
                assert float(quantity) > 0, "Quantity should be positive"
                
                # Quote quantity validation
                assert 'quoteQty' in first_trade, "Trade should contain quote quantity"
                quote_qty = first_trade['quoteQty']
                assert quote_qty is not None, "Quote quantity should not be null"
                assert isinstance(quote_qty, str), "Quote quantity should be a string"
                assert float(quote_qty) > 0, "Quote quantity should be positive"
                
                # Time validation
                assert 'time' in first_trade, "Trade should contain time"
                timestamp = first_trade['time']
                assert timestamp is not None, "Timestamp should not be null"
                assert isinstance(timestamp, int), "Timestamp should be an integer"
                assert timestamp > 0, "Timestamp should be positive"
                
                # Buyer maker validation
                assert 'isBuyerMaker' in first_trade, "Trade should contain buyer maker flag"
                buyer_maker = first_trade['isBuyerMaker']
                assert buyer_maker is not None, "Buyer maker should not be null"
                assert isinstance(buyer_maker, bool), "Buyer maker should be a boolean"
            
            return response
        
        result = await run_endpoint_test(
            websocket_client,
            "trades_historical",
            test_trades_historical_request,
            timeout=5.0
        )
        assert result["success"], f"Historical trades test failed: {result.get('error')}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])