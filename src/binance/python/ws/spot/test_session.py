"""
Integration tests for Binance Spot WebSocket API session endpoints.

These tests cover session management endpoints (Ed25519 authentication required).
Updated to work with the new SDK **params pattern.
"""

import os
import pytest
import pytest_asyncio
from conftest import setup_client_with_config, run_endpoint_test, BinanceTestConfig, AuthType, KeyType


def check_ed25519_credentials():
    """Check if Ed25519 credentials are properly available."""
    ed25519_api_key = os.getenv("BINANCE_ED25519_API_KEY")
    ed25519_key_path = os.getenv("BINANCE_ED25519_PRIVATE_KEY_PATH")
    
    # Check if Ed25519 environment variables are set and file exists
    if not (ed25519_api_key and ed25519_key_path and os.path.exists(ed25519_key_path)):
        return False
        
    # Additional check: make sure we're not using HMAC keys
    # If HMAC keys are the primary/only keys available, Ed25519 is not properly configured
    hmac_api_key = os.getenv("BINANCE_API_KEY")
    hmac_secret = os.getenv("BINANCE_SECRET_KEY")
    
    # If only HMAC credentials are available, Ed25519 is not properly configured
    if hmac_api_key and hmac_secret and not ed25519_api_key:
        return False
        
    return True


def should_skip_session_test():
    """Determine if session tests should be skipped."""
    return not check_ed25519_credentials()

def should_skip_session_status_logout():
    """Determine if session status/logout tests should be skipped.
    
    SDK FIX CONFIRMED: The SDK has been updated to handle zero-parameter methods.
    Session status and logout are now in the zero_param_methods list and will
    send 0 parameters as expected by the API.
    
    The _add_authentication method now bypasses parameter injection for:
    - session.status
    - session.logout
    - userDataStream.subscribe  
    - userDataStream.unsubscribe
    """
    # SDK is fixed - check for Ed25519 credentials instead
    return not check_ed25519_credentials()


class TestSessionEndpoints:
    """Test class for session management WebSocket API endpoints."""
    
    @pytest.mark.session
    @pytest_asyncio.fixture(autouse=True)
    async def setup(self):
        """Setup for session endpoint tests."""
        self.config = BinanceTestConfig(
            name="Ed25519-Session",
            description="Test session endpoints with Ed25519 authentication",
            auth_type=AuthType.USER_DATA,
            key_type=KeyType.ED25519,
            api_key="test_ed25519_api_key",
            private_key_path="/path/to/ed25519/private/key"
        )
        
        yield
    
    @pytest_asyncio.fixture
    async def ed25519_client(self):
        """Create an Ed25519 authenticated WebSocket client for session endpoints."""
        ed25519_api_key = os.getenv("BINANCE_ED25519_API_KEY")
        ed25519_key_path = os.getenv("BINANCE_ED25519_PRIVATE_KEY_PATH")
        
        if not ed25519_api_key or not ed25519_key_path or not os.path.exists(ed25519_key_path):
            pytest.skip("Ed25519 credentials not available for session tests")
        
        config = BinanceTestConfig(
            name="Ed25519-Session",
            description="Test session endpoints with Ed25519 authentication",
            auth_type=AuthType.USER_DATA,
            key_type=KeyType.ED25519,
            api_key=ed25519_api_key,
            private_key_path=ed25519_key_path
        )
        
        client = await setup_client_with_config(config)
        try:
            yield client
        finally:
            await client.disconnect()
    
    @pytest.mark.session
    @pytest.mark.integration
    @pytest.mark.asyncio
    @pytest.mark.skipif(should_skip_session_test(), reason="Ed25519 credentials not available - session endpoints require Ed25519 authentication")
    async def test_session_logon(self, ed25519_client):
        """Test session logon endpoint."""
        
        async def test_session_logon_request():
            from binance.ws.spot.models.session_logon_request import SessionLogonRequest
            import time
            
            # Session logon - let SDK handle authentication
            request = SessionLogonRequest()
            response = await ed25519_client.session_logon(request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            ed25519_client,
            "session_logon",
            test_session_logon_request,
            timeout=10.0
        )
        assert result["success"], f"Session logon test failed: {result.get('error')}"
    
    @pytest.mark.session
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_session_status(self, ed25519_client):
        """Test session status endpoint."""
        
        async def test_session_status_request():
            from binance.ws.spot.models.session_logon_request import SessionLogonRequest
            from binance.ws.spot.models.session_status_request import SessionStatusRequest
            import time
            
            # First logon to create a session
            logon_request = SessionLogonRequest()
            await ed25519_client.session_logon(logon_request)
            
            # Then check session status
            status_request = SessionStatusRequest()
            response = await ed25519_client.session_status(status_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            ed25519_client,
            "session_status",
            test_session_status_request,
            timeout=15.0
        )
        assert result["success"], f"Session status test failed: {result.get('error')}"
    
    @pytest.mark.session
    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_session_logout(self, ed25519_client):
        """Test session logout endpoint."""
        
        async def test_session_logout_request():
            from binance.ws.spot.models.session_logon_request import SessionLogonRequest
            from binance.ws.spot.models.session_logout_request import SessionLogoutRequest
            import time
            
            # First logon to create a session
            logon_request = SessionLogonRequest()
            await ed25519_client.session_logon(logon_request)
            
            # Then logout
            logout_request = SessionLogoutRequest()
            response = await ed25519_client.session_logout(logout_request)
            assert response.result is not None
            return response
        
        result = await run_endpoint_test(
            ed25519_client,
            "session_logout",
            test_session_logout_request,
            timeout=15.0
        )
        assert result["success"], f"Session logout test failed: {result.get('error')}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])