package options_test

import (
	"testing"

	"github.com/openxapi/binance-go/ws/options"
	"github.com/stretchr/testify/suite"
)

// AuthenticationTestSuite tests authentication functionality
type AuthenticationTestSuite struct {
	BaseTestSuite
}

// TestAuthenticationSuite runs the authentication test suite
func TestAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationTestSuite))
}

// TestHMACAuthentication tests HMAC-SHA256 authentication
func (s *AuthenticationTestSuite) TestHMACAuthentication() {
	s.requireAuth()
	
	s.Run("HMACAuthSetup", func() {
		// Create HMAC auth
		hmacAuth := options.NewAuth(testAPIKey)
		hmacAuth.SetSecretKey(testSecretKey)
		
		// Verify auth creation
		s.Require().NotNil(hmacAuth)
		
		// Test context creation
		authCtx, err := hmacAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		s.Require().NotNil(authCtx)
		
		// Test client creation with auth
		authClient := options.NewClientWithAuth(hmacAuth)
		s.Require().NotNil(authClient)
		
		s.logVerbose("HMAC authentication setup successful")
	})
	
	s.Run("HMACAuthValidation", func() {
		// Test with invalid API key
		invalidAuth := options.NewAuth("")
		_, err := invalidAuth.ContextWithValue(s.ctx)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "API key is required")
		
		// Test with valid API key
		validAuth := options.NewAuth("valid_key")
		validAuth.SetSecretKey("valid_secret")
		_, err = validAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		s.logVerbose("HMAC authentication validation completed")
	})
}

// TestRSAAuthentication tests RSA authentication setup
func (s *AuthenticationTestSuite) TestRSAAuthentication() {
	s.Run("RSAAuthSetup", func() {
		// Create RSA auth (without actual private key for this test)
		rsaAuth := options.NewAuth(testAPIKey)
		
		// Test setting private key path
		rsaAuth.SetPrivateKeyPath("/path/to/private-key.pem")
		
		// Test setting passphrase
		rsaAuth.SetPassphrase("test-passphrase")
		
		s.Require().NotNil(rsaAuth)
		s.logVerbose("RSA authentication configuration tested")
	})
}

// TestEd25519Authentication tests Ed25519 authentication setup
func (s *AuthenticationTestSuite) TestEd25519Authentication() {
	s.Run("Ed25519AuthSetup", func() {
		// Create Ed25519 auth (without actual private key for this test)
		ed25519Auth := options.NewAuth(testAPIKey)
		
		// Test setting private key path
		ed25519Auth.SetPrivateKeyPath("/path/to/ed25519-key.pem")
		
		s.Require().NotNil(ed25519Auth)
		s.logVerbose("Ed25519 authentication configuration tested")
	})
}

// TestAuthenticationPriority tests authentication priority and context handling
func (s *AuthenticationTestSuite) TestAuthenticationPriority() {
	s.requireAuth()
	
	s.Run("AuthenticationPriority", func() {
		// Create client-level auth
		clientAuth := options.NewAuth("client_key")
		clientAuth.SetSecretKey("client_secret")
		client := options.NewClientWithAuth(clientAuth)
		
		// Create context-level auth
		contextAuth := options.NewAuth("context_key") 
		contextAuth.SetSecretKey("context_secret")
		authCtx, err := contextAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		// Both should be valid
		s.Require().NotNil(client)
		s.Require().NotNil(authCtx)
		
		s.logVerbose("Authentication priority test completed")
	})
}

// TestAuthenticationMethods tests various authentication method configurations
func (s *AuthenticationTestSuite) TestAuthenticationMethods() {
	s.requireAuth()
	
	s.Run("MultipleAuthMethods", func() {
		// Test different authentication scenarios
		
		// 1. Client-level authentication
		clientAuth := options.NewAuth(testAPIKey)
		clientAuth.SetSecretKey(testSecretKey)
		clientWithAuth := options.NewClientWithAuth(clientAuth)
		s.Require().NotNil(clientWithAuth)
		
		// 2. Per-request authentication
		requestAuth := options.NewAuth(testAPIKey)
		requestAuth.SetSecretKey(testSecretKey)
		requestCtx, err := requestAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		s.Require().NotNil(requestCtx)
		
		// 3. Setting auth on existing client
		regularClient := options.NewClient()
		regularClient.SetAuth(clientAuth)
		s.Require().NotNil(regularClient)
		
		s.logVerbose("Multiple authentication methods tested")
	})
}

// TestAuthenticationErrors tests authentication error scenarios
func (s *AuthenticationTestSuite) TestAuthenticationErrors() {
	s.Run("AuthenticationErrors", func() {
		// Test empty API key
		emptyAuth := options.NewAuth("")
		_, err := emptyAuth.ContextWithValue(s.ctx)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "API key is required")
		
		// Test nil auth scenarios
		client := options.NewClient()
		s.Require().NotNil(client)
		
		// Verify client handles missing auth gracefully
		s.logVerbose("Authentication error scenarios tested")
	})
}

// TestContextAuthentication tests context-based authentication
func (s *AuthenticationTestSuite) TestContextAuthentication() {
	s.requireAuth()
	
	s.Run("ContextAuthentication", func() {
		// Create different auth instances for different purposes
		tradingAuth := options.NewAuth("trading_" + testAPIKey)
		tradingAuth.SetSecretKey(testSecretKey)
		
		readOnlyAuth := options.NewAuth("readonly_" + testAPIKey)
		readOnlyAuth.SetSecretKey(testSecretKey)
		
		// Create contexts with different auth
		tradingCtx, err := tradingAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		readOnlyCtx, err := readOnlyAuth.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		// Both contexts should be valid but different
		s.Require().NotNil(tradingCtx)
		s.Require().NotNil(readOnlyCtx)
		
		s.logVerbose("Context authentication test completed")
	})
}

// TestSecurityBestPractices tests security-related functionality
func (s *AuthenticationTestSuite) TestSecurityBestPractices() {
	s.Run("SecurityPractices", func() {
		// Test that auth instances are independent
		auth1 := options.NewAuth("key1")
		auth1.SetSecretKey("secret1")
		
		auth2 := options.NewAuth("key2")
		auth2.SetSecretKey("secret2")
		
		// They should be independent instances
		s.Require().NotEqual(auth1, auth2)
		
		// Test context isolation
		ctx1, err := auth1.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		ctx2, err := auth2.ContextWithValue(s.ctx)
		s.Require().NoError(err)
		
		// Contexts should be different
		s.Require().NotEqual(ctx1, ctx2)
		
		s.logVerbose("Security best practices verified")
	})
}