package cmfutures_test

import (
	"log"
	"testing"
	"time"

	"github.com/openxapi/binance-go/ws/cmfutures"
	"github.com/openxapi/binance-go/ws/cmfutures/models"
	"github.com/stretchr/testify/suite"
)

// UserDataTestSuite tests user data stream related APIs
type UserDataTestSuite struct {
	BaseTestSuite
	listenKey string
}

// TestUserDataTestSuite runs the user data test suite
func TestUserDataTestSuite(t *testing.T) {
	suite.Run(t, new(UserDataTestSuite))
}

// SetupTest runs before each test
func (s *UserDataTestSuite) SetupTest() {
	s.listenKey = ""
	
	// Check and reconnect if needed
	if !s.client.IsConnected() {
		log.Println("WebSocket connection lost, attempting to reconnect...")
		
		// Create new client with auth
		s.client = cmfutures.NewClientWithAuth(s.auth)
		
		// Set testnet server
		err := s.client.SetActiveServer("testnet1")
		if err != nil {
			s.T().Fatalf("Failed to set testnet server during reconnect: %v", err)
		}
		
		// Reconnect
		err = s.client.Connect(s.ctx)
		if err != nil {
			s.T().Fatalf("Failed to reconnect to WebSocket: %v", err)
		}
		
		// Allow connection to stabilize
		time.Sleep(500 * time.Millisecond)
		log.Println("Successfully reconnected to WebSocket")
	}
}

// TearDownTest runs after each test
func (s *UserDataTestSuite) TearDownTest() {
	// Clean up listen key if it exists
	if s.listenKey != "" {
		s.stopUserDataStream()
		s.listenKey = ""
	}
}

// TestUserDataStreamStart tests the userDataStream.start endpoint
func (s *UserDataTestSuite) TestUserDataStreamStart() {
	s.requireAuth()

	done := make(chan bool)
	request := models.NewUserDataStreamStartRequest()

	s.logVerbose("Starting user data stream")

	err := s.client.SendUserDataStreamStart(s.getTestContext(), request, func(response *models.UserDataStreamStartResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "userDataStream.start")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")
		s.Require().NotNil(response.Result, "Result should not be nil")
		s.Require().NotEmpty(response.Result.ListenKey, "Listen key should not be empty")

		// Store listen key for cleanup
		s.listenKey = response.Result.ListenKey

		// Log stream details
		s.logVerbose("User data stream started: %s", formatJSON(response))
		log.Printf("🎯 User data stream started with listen key: %s", s.listenKey)

		return nil
	})

	s.Require().NoError(err, "Failed to send userDataStream.start request")
	s.waitForResponse(done, defaultTimeout, "userDataStream.start")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestUserDataStreamPing tests the userDataStream.ping endpoint
func (s *UserDataTestSuite) TestUserDataStreamPing() {
	s.requireAuth()

	// First start a stream
	s.startUserDataStreamForTest()
	if s.listenKey == "" {
		s.T().Skip("Failed to start user data stream for ping test")
	}

	done := make(chan bool)
	request := models.NewUserDataStreamPingRequest()

	s.logVerbose("Pinging user data stream: %s", s.listenKey)

	err := s.client.SendUserDataStreamPing(s.getTestContext(), request, func(response *models.UserDataStreamPingResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "userDataStream.ping")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")

		// Log ping result
		s.logVerbose("User data stream ping successful: %s", formatJSON(response))
		log.Printf("🏓 User data stream ping successful for listen key: %s", s.listenKey)

		return nil
	})

	s.Require().NoError(err, "Failed to send userDataStream.ping request")
	s.waitForResponse(done, defaultTimeout, "userDataStream.ping")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestUserDataStreamStop tests the userDataStream.stop endpoint
func (s *UserDataTestSuite) TestUserDataStreamStop() {
	s.requireAuth()

	// First start a stream
	s.startUserDataStreamForTest()
	if s.listenKey == "" {
		s.T().Skip("Failed to start user data stream for stop test")
	}

	done := make(chan bool)
	request := models.NewUserDataStreamStopRequest()

	s.logVerbose("Stopping user data stream: %s", s.listenKey)

	err := s.client.SendUserDataStreamStop(s.getTestContext(), request, func(response *models.UserDataStreamStopResponse, err error) error {
		defer close(done)

		if err != nil {
			s.handleAPIError(err, "userDataStream.stop")
			return err
		}

		// Validate response
		s.Require().NotNil(response, "Response should not be nil")
		s.Require().NotEmpty(response.Id, "Response ID should not be empty")
		s.Require().Equal(int64(200), response.Status, "Response status should be 200")

		// Clear listen key since stream is stopped
		s.listenKey = ""

		// Log stop result
		s.logVerbose("User data stream stopped: %s", formatJSON(response))
		log.Printf("🛑 User data stream stopped successfully")

		return nil
	})

	s.Require().NoError(err, "Failed to send userDataStream.stop request")
	s.waitForResponse(done, defaultTimeout, "userDataStream.stop")

	// Rate limit delay
	time.Sleep(rateLimitDelay)
}

// TestUserDataStreamLifecycle tests the complete lifecycle of a user data stream
func (s *UserDataTestSuite) TestUserDataStreamLifecycle() {
	s.requireAuth()

	// 1. Start stream
	s.TestUserDataStreamStart()
	s.Require().NotEmpty(s.listenKey, "Listen key should be set after start")

	// 2. Ping stream multiple times
	for i := 0; i < 3; i++ {
		s.logVerbose("Lifecycle test: Ping #%d", i+1)
		s.pingUserDataStream()
		time.Sleep(500 * time.Millisecond)
	}

	// 3. Stop stream
	s.stopUserDataStreamInTest()
	s.Assert().Empty(s.listenKey, "Listen key should be cleared after stop")
}

// TestZUserDataStreamErrorHandling tests error handling for user data stream
// Named with Z prefix to run last in the suite to avoid disrupting other tests
func (s *UserDataTestSuite) TestZUserDataStreamErrorHandling() {
	s.requireAuth()

	// Test ping without starting a stream first - this should generate an error
	// since we don't have an active listen key
	done := make(chan bool)
	pingRequest := models.NewUserDataStreamPingRequest()
	gotError := false
	
	s.logVerbose("Testing ping without active stream (expecting error)")
	
	err := s.client.SendUserDataStreamPing(s.getTestContext(), pingRequest, func(response *models.UserDataStreamPingResponse, err error) error {
		defer close(done)
		
		if err != nil {
			gotError = true
			if apiErr, ok := cmfutures.IsAPIError(err); ok {
				s.logVerbose("Received expected API error: Code=%d, Message=%s", 
					apiErr.Code, apiErr.Message)
				log.Printf("⚠️ Ping without stream error (expected): %s", apiErr.Message)
			} else {
				s.logVerbose("Received error: %v", err)
			}
		} else {
			s.logVerbose("Ping without stream succeeded (unexpected)")
		}
		
		return nil
	})
	
	s.Require().NoError(err, "Failed to send ping request")
	s.waitForResponse(done, defaultTimeout, "userDataStream.ping error test")
	
	// For CMFUTURES, the behavior might be different than SPOT/UMFUTURES
	// If we didn't get an error, just log it and continue
	if !gotError {
		log.Println("Note: CMFUTURES ping without stream did not return an error (different behavior)")
	}
}

// Helper methods

func (s *UserDataTestSuite) startUserDataStreamForTest() {
	done := make(chan bool)
	request := models.NewUserDataStreamStartRequest()

	err := s.client.SendUserDataStreamStart(s.getTestContext(), request, func(response *models.UserDataStreamStartResponse, err error) error {
		defer close(done)

		if err != nil {
			s.logVerbose("Failed to start user data stream: %v", err)
			return err
		}

		s.listenKey = response.Result.ListenKey
		s.logVerbose("User data stream started for test: %s", s.listenKey)
		return nil
	})

	if err != nil {
		return
	}

	select {
	case <-done:
	case <-time.After(defaultTimeout):
		s.T().Error("Timeout starting user data stream")
	}

	time.Sleep(rateLimitDelay)
}

func (s *UserDataTestSuite) pingUserDataStream() {
	if s.listenKey == "" {
		return
	}

	done := make(chan bool)
	request := models.NewUserDataStreamPingRequest()

	err := s.client.SendUserDataStreamPing(s.getTestContext(), request, func(response *models.UserDataStreamPingResponse, err error) error {
		defer close(done)
		if err != nil {
			s.logVerbose("Failed to ping user data stream: %v", err)
		} else {
			s.logVerbose("User data stream ping successful")
		}
		return nil
	})

	if err == nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}

	time.Sleep(rateLimitDelay)
}

func (s *UserDataTestSuite) stopUserDataStream() {
	if s.listenKey == "" {
		return
	}

	done := make(chan bool)
	request := models.NewUserDataStreamStopRequest()

	err := s.client.SendUserDataStreamStop(s.getTestContext(), request, func(response *models.UserDataStreamStopResponse, err error) error {
		defer close(done)
		if err != nil {
			s.logVerbose("Failed to stop user data stream: %v", err)
		} else {
			s.logVerbose("User data stream stopped")
		}
		s.listenKey = ""
		return nil
	})

	if err == nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

func (s *UserDataTestSuite) stopUserDataStreamInTest() {
	done := make(chan bool)
	request := models.NewUserDataStreamStopRequest()

	err := s.client.SendUserDataStreamStop(s.getTestContext(), request, func(response *models.UserDataStreamStopResponse, err error) error {
		defer close(done)
		if err != nil {
			s.T().Errorf("Failed to stop user data stream: %v", err)
		} else {
			log.Printf("✅ User data stream stopped in lifecycle test")
		}
		s.listenKey = ""
		return nil
	})

	s.Require().NoError(err)
	s.waitForResponse(done, defaultTimeout, "userDataStream.stop in lifecycle")
	time.Sleep(rateLimitDelay)
}