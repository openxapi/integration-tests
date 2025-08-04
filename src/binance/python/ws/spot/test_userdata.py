"""
Integration tests for Binance Spot WebSocket API user data endpoints.

These tests cover all user data endpoints that require USER_DATA authentication.
Updated to work with the new SDK **params pattern.
"""

import pytest
import pytest_asyncio
from conftest import setup_client_with_config, run_endpoint_test, BinanceTestConfig, AuthType, KeyType


class TestUserDataEndpoints:
    """Test class for user data WebSocket API endpoints."""
    
    @pytest.mark.user_data
    @pytest_asyncio.fixture(autouse=True)
    async def setup(self):
        """Setup for user data endpoint tests."""
        self.config = BinanceTestConfig(
            name="HMAC-UserData",
            description="Test USER_DATA endpoints with HMAC authentication",
            auth_type=AuthType.USER_DATA,
            key_type=KeyType.HMAC,
            api_key="test_api_key",
            secret_key="test_secret_key"
        )
        
        yield
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_account_status(self, authenticated_client):
        """Test account status endpoint."""
        
        async def test_account_status_request():
            response = await authenticated_client.account_status()
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "account_status",
            test_account_status_request,
            timeout=10.0
        )
        assert result["success"], f"Account status test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_account_commission(self, authenticated_client, test_symbol):
        """Test account commission rates endpoint."""
        
        async def test_account_commission_request():
            from binance.ws.spot.models.account_commission_request import AccountCommissionRequest
            request = AccountCommissionRequest(symbol=test_symbol)
            response = await authenticated_client.account_commission(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "account_commission",
            test_account_commission_request,
            timeout=10.0
        )
        assert result["success"], f"Account commission test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_my_trades(self, authenticated_client, test_symbol):
        """Test my trades history endpoint."""
        
        async def test_my_trades_request():
            from binance.ws.spot.models.my_trades_request import MyTradesRequest
            request = MyTradesRequest(symbol=test_symbol, limit=10)
            response = await authenticated_client.mytrades(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "my_trades",
            test_my_trades_request,
            timeout=10.0
        )
        assert result["success"], f"My trades test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_all_orders(self, authenticated_client, test_symbol):
        """Test all orders history endpoint."""
        
        async def test_all_orders_request():
            from binance.ws.spot.models.all_orders_request import AllOrdersRequest
            request = AllOrdersRequest(symbol=test_symbol, limit=10)
            response = await authenticated_client.allorders(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "all_orders",
            test_all_orders_request,
            timeout=10.0
        )
        assert result["success"], f"All orders test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_open_orders_status(self, authenticated_client, test_symbol):
        """Test open orders status endpoint."""
        
        async def test_open_orders_request():
            from binance.ws.spot.models.open_orders_status_request import OpenOrdersStatusRequest
            request = OpenOrdersStatusRequest(symbol=test_symbol)
            response = await authenticated_client.openorders_status(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "open_orders_status",
            test_open_orders_request,
            timeout=10.0
        )
        assert result["success"], f"Open orders status test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_account_rate_limits_orders(self, authenticated_client):
        """Test account order rate limits endpoint."""
        
        async def test_rate_limits_request():
            response = await authenticated_client.account_ratelimits_orders()
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "account_rate_limits_orders",
            test_rate_limits_request,
            timeout=10.0
        )
        assert result["success"], f"Account rate limits test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_my_allocations(self, authenticated_client, test_symbol):
        """Test my allocations endpoint."""
        
        async def test_allocations_request():
            from binance.ws.spot.models.my_allocations_request import MyAllocationsRequest
            request = MyAllocationsRequest(symbol=test_symbol, limit=10)
            response = await authenticated_client.myallocations(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "my_allocations",
            test_allocations_request,
            timeout=10.0
        )
        assert result["success"], f"My allocations test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_trades_recent(self, authenticated_client, test_symbol):
        """Test recent trades endpoint."""
        
        async def test_recent_trades_request():
            from binance.ws.spot.models.trades_recent_request import TradesRecentRequest
            request = TradesRecentRequest(symbol=test_symbol, limit=10)
            response = await authenticated_client.trades_recent(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "trades_recent",
            test_recent_trades_request,
            timeout=10.0
        )
        assert result["success"], f"Recent trades test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_all_order_lists(self, authenticated_client):
        """Test all order lists endpoint."""
        
        async def test_all_order_lists_request():
            from binance.ws.spot.models.all_order_lists_request import AllOrderListsRequest
            request = AllOrderListsRequest(limit=10)
            response = await authenticated_client.allorderlists(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "all_order_lists",
            test_all_order_lists_request,
            timeout=10.0
        )
        assert result["success"], f"All order lists test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_open_order_lists_status(self, authenticated_client):
        """Test open order lists status endpoint."""
        
        async def test_open_order_lists_request():
            response = await authenticated_client.openorderlists_status()
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "open_order_lists_status",
            test_open_order_lists_request,
            timeout=10.0
        )
        assert result["success"], f"Open order lists status test failed: {result.get('error')}"
    
    @pytest.mark.user_data
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_my_prevented_matches(self, authenticated_client, test_symbol):
        """Test my prevented matches endpoint."""
        
        async def test_prevented_matches_request():
            from binance.ws.spot.models.my_prevented_matches_request import MyPreventedMatchesRequest
            # API requires either orderId or preventedMatchId - using orderId=1 as a test value
            request = MyPreventedMatchesRequest(symbol=test_symbol, orderId=1, limit=10)
            response = await authenticated_client.mypreventedmatches(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "my_prevented_matches",
            test_prevented_matches_request,
            timeout=10.0
        )
        assert result["success"], f"My prevented matches test failed: {result.get('error')}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])