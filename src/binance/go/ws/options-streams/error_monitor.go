package streamstest

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ErrorMonitor captures log messages that indicate SDK parsing errors
type ErrorMonitor struct {
	errors []string
	mu     sync.RWMutex
	originalOutput *os.File
	isActive bool
}

var globalErrorMonitor = &ErrorMonitor{}

// StartErrorMonitoring starts monitoring for SDK parsing errors
func (em *ErrorMonitor) StartErrorMonitoring() {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if em.isActive {
		return // Already monitoring
	}
	
	em.errors = make([]string, 0)
	em.isActive = true
	
	// Override log output to capture SDK errors
	em.originalOutput = log.Writer().(*os.File)
	
	// Create a custom writer that captures errors
	pr, pw, _ := os.Pipe()
	log.SetOutput(pw)
	
	// Start goroutine to read and filter log messages
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := pr.Read(buffer)
			if err != nil {
				break
			}
			
			message := string(buffer[:n])
			
			// Check if this is an SDK parsing error
			if em.isSDKError(message) {
				em.mu.Lock()
				em.errors = append(em.errors, strings.TrimSpace(message))
				em.mu.Unlock()
			}
			
			// Also write to original output
			if em.originalOutput != nil {
				em.originalOutput.Write(buffer[:n])
			}
		}
	}()
}

// StopErrorMonitoring stops monitoring and returns captured errors
func (em *ErrorMonitor) StopErrorMonitoring() []string {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if !em.isActive {
		return nil
	}
	
	em.isActive = false
	
	// Restore original log output
	if em.originalOutput != nil {
		log.SetOutput(em.originalOutput)
	}
	
	// Return captured errors
	result := make([]string, len(em.errors))
	copy(result, em.errors)
	em.errors = nil
	
	return result
}

// GetCurrentErrors returns current errors without stopping monitoring
func (em *ErrorMonitor) GetCurrentErrors() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	result := make([]string, len(em.errors))
	copy(result, em.errors)
	return result
}

// ClearErrors clears current errors
func (em *ErrorMonitor) ClearErrors() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.errors = nil
}

// isSDKError checks if a log message indicates an SDK parsing error
func (em *ErrorMonitor) isSDKError(message string) bool {
	errorIndicators := []string{
		"Error handling message:",
		"json: cannot unmarshal",
		"Failed to parse",
		"Unknown message type:",
		"Unknown options stream message format:",
		"Error parsing",
		"Failed to marshal",
		"Failed to unmarshal",
	}
	
	for _, indicator := range errorIndicators {
		if strings.Contains(message, indicator) {
			return true
		}
	}
	
	return false
}

// Global convenience functions
func StartErrorMonitoring() {
	globalErrorMonitor.StartErrorMonitoring()
}

func StopErrorMonitoring() []string {
	return globalErrorMonitor.StopErrorMonitoring()
}

func GetCurrentErrors() []string {
	return globalErrorMonitor.GetCurrentErrors()
}

func ClearErrors() {
	globalErrorMonitor.ClearErrors()
}

// WaitForErrorsOrTimeout waits for either errors to occur or timeout
func WaitForErrorsOrTimeout(timeout time.Duration) []string {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		errors := GetCurrentErrors()
		if len(errors) > 0 {
			return errors
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	return GetCurrentErrors()
}