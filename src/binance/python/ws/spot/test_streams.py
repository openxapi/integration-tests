"""
Integration tests for Binance Spot WebSocket API user data stream endpoints.

These tests cover user data stream management endpoints.
Updated to work with the new SDK **params pattern.
"""

import os
import pytest
import pytest_asyncio
from conftest import setup_client_with_config, run_endpoint_test, BinanceTestConfig, AuthType, KeyType


def should_skip_subscribe_unsubscribe_test():
    """Determine if subscribe/unsubscribe tests should be skipped.
    
    These methods require:
    1. Ed25519 authentication (session.logon only works with Ed25519)
    2. Session logon before calling subscribe/unsubscribe
    3. Empty request objects (no parameters)
    
    Reference: Go SDK implementation requires Ed25519 and session logon first.
    """
    ed25519_api_key = os.getenv("BINANCE_ED25519_API_KEY")
    ed25519_key_path = os.getenv("BINANCE_ED25519_PRIVATE_KEY_PATH")
    
    # If either Ed25519 variable is missing or file doesn't exist, skip
    if not ed25519_api_key or not ed25519_key_path:
        return True
    if not os.path.exists(ed25519_key_path):
        return True
        
    return False


class TestUserDataStreamEndpoints:
    """Test class for user data stream WebSocket API endpoints."""
    
    @pytest.mark.streams
    @pytest_asyncio.fixture(autouse=True)
    async def setup(self):
        """Setup for stream endpoint tests."""
        self.config = BinanceTestConfig(
            name="HMAC-Streams",
            description="Test user data stream endpoints with HMAC authentication",
            auth_type=AuthType.TRADE,
            key_type=KeyType.HMAC,
            api_key="test_api_key",
            secret_key="test_secret_key"
        )
        
        yield
    
    @pytest_asyncio.fixture
    async def ed25519_client(self):
        """Create an Ed25519 authenticated WebSocket client for session-based endpoints."""
        ed25519_api_key = os.getenv("BINANCE_ED25519_API_KEY")
        ed25519_key_path = os.getenv("BINANCE_ED25519_PRIVATE_KEY_PATH")
        
        if not ed25519_api_key or not ed25519_key_path or not os.path.exists(ed25519_key_path):
            pytest.skip("Ed25519 credentials not available for session-based tests")
        
        config = BinanceTestConfig(
            name="Ed25519-Session",
            description="Test session endpoints with Ed25519 authentication",
            auth_type=AuthType.TRADE,
            key_type=KeyType.ED25519,
            api_key=ed25519_api_key,
            private_key_path=ed25519_key_path
        )
        
        client = await setup_client_with_config(config)
        try:
            yield client
        finally:
            await client.disconnect()
    
    @pytest.mark.streams
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_user_data_stream_start(self, authenticated_client):
        """Test starting a user data stream."""
        
        async def test_stream_start_request():
            from binance.ws.spot.models.user_data_stream_start_request import UserDataStreamStartRequest
            # SDK Bug: USER_STREAM auth doesn't add apiKey, so we add it manually
            api_key = os.getenv("BINANCE_API_KEY")
            if not api_key:
                pytest.skip("BINANCE_API_KEY environment variable not set")
            
            request = UserDataStreamStartRequest(apiKey=api_key)
            response = await authenticated_client.userdatastream_start(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "user_data_stream_start",
            test_stream_start_request,
            timeout=10.0
        )
        assert result["success"], f"User data stream start test failed: {result.get('error')}"
    
    @pytest.mark.streams
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_user_data_stream_ping(self, authenticated_client):
        """Test pinging a user data stream."""
        
        async def test_stream_ping_request():
            # First start a stream
            from binance.ws.spot.models.user_data_stream_start_request import UserDataStreamStartRequest
            from binance.ws.spot.models.user_data_stream_ping_request import UserDataStreamPingRequest
            
            # SDK Bug: USER_STREAM auth doesn't add apiKey, so we add it manually
            api_key = os.getenv("BINANCE_API_KEY")
            if not api_key:
                pytest.skip("BINANCE_API_KEY environment variable not set")
            
            start_request = UserDataStreamStartRequest(apiKey=api_key)
            start_response = await authenticated_client.userdatastream_start(start_request)
            listen_key = start_response.result.listen_key
            
            # Then ping it (ping requires both listen_key and apiKey)
            ping_request = UserDataStreamPingRequest(listenKey=listen_key, apiKey=api_key)
            response = await authenticated_client.userdatastream_ping(ping_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "user_data_stream_ping",
            test_stream_ping_request,
            timeout=10.0
        )
        assert result["success"], f"User data stream ping test failed: {result.get('error')}"
    
    @pytest.mark.streams
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_user_data_stream_stop(self, authenticated_client):
        """Test stopping a user data stream."""
        
        async def test_stream_stop_request():
            # First start a stream
            from binance.ws.spot.models.user_data_stream_start_request import UserDataStreamStartRequest
            from binance.ws.spot.models.user_data_stream_stop_request import UserDataStreamStopRequest
            
            # SDK Bug: USER_STREAM auth doesn't add apiKey, so we add it manually
            api_key = os.getenv("BINANCE_API_KEY")
            if not api_key:
                pytest.skip("BINANCE_API_KEY environment variable not set")
            
            start_request = UserDataStreamStartRequest(apiKey=api_key)
            start_response = await authenticated_client.userdatastream_start(start_request)
            listen_key = start_response.result.listen_key
            
            # Then stop it (stop requires both listen_key and apiKey)
            stop_request = UserDataStreamStopRequest(listenKey=listen_key, apiKey=api_key)
            response = await authenticated_client.userdatastream_stop(stop_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            authenticated_client,
            "user_data_stream_stop",
            test_stream_stop_request,
            timeout=10.0
        )
        assert result["success"], f"User data stream stop test failed: {result.get('error')}"
    
    @pytest.mark.streams
    @pytest.mark.integration
    @pytest.mark.asyncio
    @pytest.mark.skipif(should_skip_subscribe_unsubscribe_test(), reason="Subscribe/unsubscribe methods require Ed25519 authentication and session logon")
    async def test_user_data_stream_subscribe(self, ed25519_client):
        """Test subscribing to a user data stream."""
        
        async def test_stream_subscribe_request():
            # First perform session logon (required for subscribe/unsubscribe)
            from binance.ws.spot.models.session_logon_request import SessionLogonRequest
            from binance.ws.spot.models.user_data_stream_start_request import UserDataStreamStartRequest
            from binance.ws.spot.models.user_data_stream_subscribe_request import UserDataStreamSubscribeRequest
            
            import time
            
            # Session logon - let SDK handle authentication
            logon_request = SessionLogonRequest()
            await ed25519_client.session_logon(logon_request)
            
            # Start a user data stream
            start_request = UserDataStreamStartRequest()
            start_response = await ed25519_client.userdatastream_start(start_request)
            listen_key = start_response.result.listen_key
            
            # Then subscribe to it (subscribe takes no parameters after session logon)
            subscribe_request = UserDataStreamSubscribeRequest()
            response = await ed25519_client.userdatastream_subscribe(subscribe_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            ed25519_client,
            "user_data_stream_subscribe",
            test_stream_subscribe_request,
            timeout=15.0
        )
        assert result["success"], f"User data stream subscribe test failed: {result.get('error')}"
    
    @pytest.mark.streams
    @pytest.mark.integration
    @pytest.mark.asyncio
    @pytest.mark.skipif(should_skip_subscribe_unsubscribe_test(), reason="Subscribe/unsubscribe methods require Ed25519 authentication and session logon")
    async def test_user_data_stream_unsubscribe(self, ed25519_client):
        """Test unsubscribing from a user data stream."""
        
        async def test_stream_unsubscribe_request():
            # First perform session logon (required for subscribe/unsubscribe)
            from binance.ws.spot.models.session_logon_request import SessionLogonRequest
            from binance.ws.spot.models.user_data_stream_start_request import UserDataStreamStartRequest
            from binance.ws.spot.models.user_data_stream_subscribe_request import UserDataStreamSubscribeRequest
            from binance.ws.spot.models.user_data_stream_unsubscribe_request import UserDataStreamUnsubscribeRequest
            
            import time
            
            # Session logon - let SDK handle authentication
            logon_request = SessionLogonRequest()
            await ed25519_client.session_logon(logon_request)
            
            # Start a user data stream
            start_request = UserDataStreamStartRequest()
            start_response = await ed25519_client.userdatastream_start(start_request)
            listen_key = start_response.result.listen_key
            
            # Subscribe to it first
            subscribe_request = UserDataStreamSubscribeRequest()
            await ed25519_client.userdatastream_subscribe(subscribe_request)
            
            # Then unsubscribe (unsubscribe takes no parameters after session logon)
            unsubscribe_request = UserDataStreamUnsubscribeRequest()
            response = await ed25519_client.userdatastream_unsubscribe(unsubscribe_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            ed25519_client,
            "user_data_stream_unsubscribe",
            test_stream_unsubscribe_request,
            timeout=20.0
        )
        assert result["success"], f"User data stream unsubscribe test failed: {result.get('error')}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])