package testutils

import (
	"fmt"
	"sync"
)

// TestingT provides interface compatibility with testing frameworks
// This is a local copy to avoid circular imports
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Error(args ...interface{}) // Added for Terratest compatibility
	Fail()                     // Added for Terratest compatibility
	FailNow()
	Logf(format string, args ...interface{})
	Name() string
}

// MockTestingT is a mock implementation of the TestingT interface for testing.
// It tracks all method calls and their arguments to allow verification of testing behavior
// without actually failing tests.
type MockTestingT struct {
	mu                sync.RWMutex
	name              string
	helperCallCount   int
	errorfCallCount   int
	logfCallCount     int
	failCallCount     int
	failNowCallCount  int
	errorMessages     []string
	logMessages       []string
}

// Ensure MockTestingT implements TestingT interface at compile time
var _ TestingT = (*MockTestingT)(nil)

// NewMockTestingT creates a new mock testing instance with the given name
func NewMockTestingT(name string) *MockTestingT {
	return &MockTestingT{
		name:          name,
		errorMessages: make([]string, 0),
		logMessages:   make([]string, 0),
	}
}

// Helper implements TestingT interface
func (m *MockTestingT) Helper() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.helperCallCount++
}

// Errorf implements TestingT interface
func (m *MockTestingT) Errorf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorfCallCount++
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

// Error implements TestingT interface for Terratest compatibility
func (m *MockTestingT) Error(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorfCallCount++
	m.errorMessages = append(m.errorMessages, fmt.Sprint(args...))
}

// Fail implements TestingT interface for Terratest compatibility
func (m *MockTestingT) Fail() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCallCount++
}

// FailNow implements TestingT interface
func (m *MockTestingT) FailNow() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNowCallCount++
}

// Logf implements TestingT interface
func (m *MockTestingT) Logf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logfCallCount++
	m.logMessages = append(m.logMessages, fmt.Sprintf(format, args...))
}

// Name implements TestingT interface
func (m *MockTestingT) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// Verification methods for testing

// WasHelperCalled returns true if Helper() was called at least once
func (m *MockTestingT) WasHelperCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.helperCallCount > 0
}

// WasErrorfCalled returns true if Errorf() was called at least once
func (m *MockTestingT) WasErrorfCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorfCallCount > 0
}

// WasLogfCalled returns true if Logf() was called at least once
func (m *MockTestingT) WasLogfCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logfCallCount > 0
}

// WasFailNowCalled returns true if FailNow() was called at least once
func (m *MockTestingT) WasFailNowCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failNowCallCount > 0
}

// GetHelperCallCount returns the number of times Helper() was called
func (m *MockTestingT) GetHelperCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.helperCallCount
}

// GetErrorfCallCount returns the number of times Errorf() was called
func (m *MockTestingT) GetErrorfCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorfCallCount
}

// GetLogfCallCount returns the number of times Logf() was called
func (m *MockTestingT) GetLogfCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logfCallCount
}

// GetFailNowCallCount returns the number of times FailNow() was called
func (m *MockTestingT) GetFailNowCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failNowCallCount
}

// GetErrorMessages returns a copy of all error messages captured
func (m *MockTestingT) GetErrorMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.errorMessages))
	copy(result, m.errorMessages)
	return result
}

// GetLogMessages returns a copy of all log messages captured
func (m *MockTestingT) GetLogMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.logMessages))
	copy(result, m.logMessages)
	return result
}

// Reset clears all captured state, allowing the mock to be reused
func (m *MockTestingT) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.helperCallCount = 0
	m.errorfCallCount = 0
	m.logfCallCount = 0
	m.failNowCallCount = 0
	m.errorMessages = m.errorMessages[:0] // Keep capacity, clear length
	m.logMessages = m.logMessages[:0]     // Keep capacity, clear length
}

// SetName allows changing the test name after creation
func (m *MockTestingT) SetName(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
}