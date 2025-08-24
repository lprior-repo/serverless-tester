package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Mock testing context for unit tests
type mockTestingT struct {
	errorMessages []string
	failed        bool
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockTestingT) FailNow() {
	m.failed = true
}

func newMockTestingT() *mockTestingT {
	return &mockTestingT{
		errorMessages: make([]string, 0),
		failed:        false,
	}
}

// Test data factories following functional programming principles

func createTestStateMachineDefinition() *StateMachineDefinition {
	return &StateMachineDefinition{
		Name:       "test-state-machine",
		Definition: `{"Comment": "Test state machine", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "Result": "Hello World!", "End": true}}}`,
		RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:       types.StateMachineTypeStandard,
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

func createTestExecutionResult(status types.ExecutionStatus, output string) *ExecutionResult {
	return &ExecutionResult{
		ExecutionArn:     "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn:  "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:             "test-execution",
		Status:           status,
		StartDate:        time.Now().Add(-5 * time.Minute),
		StopDate:         time.Now(),
		Input:            `{"message": "Hello, World!"}`,
		Output:           output,
		ExecutionTime:    5 * time.Minute,
	}
}

func createTestHistoryEvent(eventType types.HistoryEventType, timestamp time.Time) HistoryEvent {
	return HistoryEvent{
		Timestamp:                    timestamp,
		Type:                         eventType,
		ID:                          1,
		PreviousEventID:             0,
		StateEnteredEventDetails:    nil,
		StateExitedEventDetails:     nil,
		TaskStateEnteredEventDetails: nil,
		ExecutionStartedEventDetails: &ExecutionStartedEventDetails{
			Input: `{"message": "Hello, World!"}`,
			RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
		},
	}
}

// Table-driven test structures

type validateStateMachineNameTestCase struct {
	name          string
	stateMachineName  string
	expectedError bool
	description   string
}

type inputBuilderTestCase struct {
	name            string
	builderFunc     func() Input
	expectedFields  []string
	description     string
}

type executionStatusTestCase struct {
	name           string
	result         *ExecutionResult
	expectedStatus types.ExecutionStatus
	shouldSucceed  bool
	description    string
}

// Unit tests for core validation functions

func TestValidateStateMachineName(t *testing.T) {
	tests := []validateStateMachineNameTestCase{
		{
			name:             "ValidStateMachineName",
			stateMachineName: "my-state-machine",
			expectedError:    false,
			description:      "Valid state machine name with hyphens should pass validation",
		},
		{
			name:             "ValidNameWithUnderscores",
			stateMachineName: "my_state_machine",
			expectedError:    false,
			description:      "Valid state machine name with underscores should pass validation",
		},
		{
			name:             "ValidNameWithNumbers",
			stateMachineName: "MyStateMachine123",
			expectedError:    false,
			description:      "Valid state machine name with numbers should pass validation",
		},
		{
			name:             "EmptyStateMachineName",
			stateMachineName: "",
			expectedError:    true,
			description:      "Empty state machine name should fail validation",
		},
		{
			name:             "TooLongStateMachineName",
			stateMachineName: "this-state-machine-name-is-way-too-long-and-exceeds-the-maximum-allowed-length-for-aws-step-functions-state-machines-which-is-eighty-characters-total",
			expectedError:    true,
			description:      "State machine name exceeding 80 characters should fail validation",
		},
		{
			name:             "InvalidCharacters",
			stateMachineName: "my-state-machine@invalid",
			expectedError:    true,
			description:      "State machine name with invalid characters should fail validation",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineName(tc.stateMachineName)
			
			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestValidateStateMachineDefinition(t *testing.T) {
	tests := []struct {
		name          string
		definition    string
		expectedError bool
		description   string
	}{
		{
			name:          "ValidDefinition",
			definition:    `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
			expectedError: false,
			description:   "Valid state machine definition should pass validation",
		},
		{
			name:          "EmptyDefinition",
			definition:    "",
			expectedError: true,
			description:   "Empty definition should fail validation",
		},
		{
			name:          "InvalidJSON",
			definition:    `{"Comment": "Test", "StartAt": "HelloWorld"`,
			expectedError: true,
			description:   "Invalid JSON definition should fail validation",
		},
		{
			name:          "MissingStartAt",
			definition:    `{"Comment": "Test", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
			expectedError: true,
			description:   "Definition missing StartAt should fail validation",
		},
		{
			name:          "MissingStates",
			definition:    `{"Comment": "Test", "StartAt": "HelloWorld"}`,
			expectedError: true,
			description:   "Definition missing States should fail validation",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineDefinition(tc.definition)
			
			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestValidateExecutionName(t *testing.T) {
	tests := []struct {
		name          string
		executionName string
		expectedError bool
		description   string
	}{
		{
			name:          "ValidExecutionName",
			executionName: "my-execution-123",
			expectedError: false,
			description:   "Valid execution name should pass validation",
		},
		{
			name:          "EmptyExecutionName",
			executionName: "",
			expectedError: true,
			description:   "Empty execution name should fail validation",
		},
		{
			name:          "TooLongExecutionName",
			executionName: "this-execution-name-is-way-too-long-and-exceeds-the-maximum-allowed-length-for-aws-step-functions-execution-names-which-is-eighty-characters-total",
			expectedError: true,
			description:   "Execution name exceeding 80 characters should fail validation",
		},
		{
			name:          "InvalidCharacters",
			executionName: "my-execution@invalid",
			expectedError: true,
			description:   "Execution name with invalid characters should fail validation",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateExecutionName(tc.executionName)
			
			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// Tests for input builders

func TestNewInput(t *testing.T) {
	// Act
	input := NewInput()
	
	// Assert
	assert.NotNil(t, input, "NewInput should return non-nil input")
	assert.Empty(t, input.data, "New input should have empty data map")
}

func TestInputSet(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       interface{}
		expected    interface{}
		description string
	}{
		{
			name:        "SetStringValue",
			key:         "message",
			value:       "Hello, World!",
			expected:    "Hello, World!",
			description: "Should set string value correctly",
		},
		{
			name:        "SetIntValue",
			key:         "count",
			value:       42,
			expected:    42,
			description: "Should set integer value correctly",
		},
		{
			name:        "SetMapValue",
			key:         "config",
			value:       map[string]string{"env": "test"},
			expected:    map[string]string{"env": "test"},
			description: "Should set map value correctly",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := NewInput()
			
			// Act
			result := input.Set(tc.key, tc.value)
			
			// Assert
			assert.Equal(t, input, result, "Set should return the same input for chaining")
			assert.Equal(t, tc.expected, input.data[tc.key], tc.description)
		})
	}
}

func TestInputSetIf(t *testing.T) {
	tests := []struct {
		name        string
		condition   bool
		key         string
		value       interface{}
		shouldSet   bool
		description string
	}{
		{
			name:        "SetIfTrue",
			condition:   true,
			key:         "message",
			value:       "Hello, World!",
			shouldSet:   true,
			description: "Should set value when condition is true",
		},
		{
			name:        "SetIfFalse",
			condition:   false,
			key:         "message",
			value:       "Hello, World!",
			shouldSet:   false,
			description: "Should not set value when condition is false",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := NewInput()
			
			// Act
			result := input.SetIf(tc.condition, tc.key, tc.value)
			
			// Assert
			assert.Equal(t, input, result, "SetIf should return the same input for chaining")
			
			if tc.shouldSet {
				assert.Equal(t, tc.value, input.data[tc.key], tc.description)
			} else {
				_, exists := input.data[tc.key]
				assert.False(t, exists, tc.description)
			}
		})
	}
}

func TestInputMerge(t *testing.T) {
	// Arrange
	input1 := NewInput().Set("key1", "value1").Set("key2", "value2")
	input2 := NewInput().Set("key2", "updated_value2").Set("key3", "value3")
	
	// Act
	result := input1.Merge(input2)
	
	// Assert
	assert.Equal(t, input1, result, "Merge should return the same input for chaining")
	assert.Equal(t, "value1", input1.data["key1"], "Original values should be preserved")
	assert.Equal(t, "updated_value2", input1.data["key2"], "Merged values should override originals")
	assert.Equal(t, "value3", input1.data["key3"], "New values should be added")
}

func TestInputToJSON(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() Input
		expectedResult string
		shouldError    bool
		description    string
	}{
		{
			name: "SimpleInput",
			setupFunc: func() Input {
				return *NewInput().Set("message", "Hello, World!")
			},
			expectedResult: `{"message":"Hello, World!"}`,
			shouldError:    false,
			description:    "Simple input should convert to JSON correctly",
		},
		{
			name: "EmptyInput",
			setupFunc: func() Input {
				return *NewInput()
			},
			expectedResult: `{}`,
			shouldError:    false,
			description:    "Empty input should convert to empty JSON object",
		},
		{
			name: "ComplexInput",
			setupFunc: func() Input {
				return *NewInput().
					Set("message", "Hello").
					Set("count", 42).
					Set("enabled", true)
			},
			expectedResult: "", // We'll check that it's valid JSON, not exact content
			shouldError:    false,
			description:    "Complex input should convert to valid JSON",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := tc.setupFunc()
			
			// Act
			result, err := input.ToJSON()
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				if tc.expectedResult != "" {
					assert.Equal(t, tc.expectedResult, result, tc.description)
				} else {
					// For complex inputs, just verify it's valid JSON
					assert.NotEmpty(t, result, "Result should not be empty")
					assert.Contains(t, result, "message", "Result should contain expected keys")
				}
			}
		})
	}
}

// Tests for execution status validation

func TestValidateExecutionResult(t *testing.T) {
	successResult := createTestExecutionResult(types.ExecutionStatusSucceeded, `{"result": "success"}`)
	failedResult := createTestExecutionResult(types.ExecutionStatusFailed, `{"error": "execution failed"}`)
	
	tests := []struct {
		name           string
		result         *ExecutionResult
		expectSuccess  bool
		expectedErrors int
		description    string
	}{
		{
			name:           "ValidSuccessResult",
			result:         successResult,
			expectSuccess:  true,
			expectedErrors: 0,
			description:    "Valid success result should have no errors",
		},
		{
			name:           "ValidFailedResult",
			result:         failedResult,
			expectSuccess:  false,
			expectedErrors: 0,
			description:    "Valid failed result should have no errors when failure expected",
		},
		{
			name:           "NilResult",
			result:         nil,
			expectSuccess:  true,
			expectedErrors: 1,
			description:    "Nil result should return one error",
		},
		{
			name:           "UnexpectedSuccess",
			result:         successResult,
			expectSuccess:  false,
			expectedErrors: 1,
			description:    "Unexpected success should return one error",
		},
		{
			name:           "UnexpectedFailure",
			result:         failedResult,
			expectSuccess:  true,
			expectedErrors: 1,
			description:    "Unexpected failure should return one error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := validateExecutionResult(tc.result, tc.expectSuccess)
			
			// Assert
			assert.Len(t, errors, tc.expectedErrors, tc.description)
		})
	}
}

// Tests for pattern builders

func TestCreateOrderInput(t *testing.T) {
	// Arrange
	orderID := "order-123"
	customerID := "customer-456"
	items := []string{"item1", "item2"}
	
	// Act
	input := CreateOrderInput(orderID, customerID, items)
	
	// Assert
	assert.Equal(t, orderID, input.data["orderId"], "Should set order ID")
	assert.Equal(t, customerID, input.data["customerId"], "Should set customer ID")
	assert.Equal(t, items, input.data["items"], "Should set items")
}

func TestCreateWorkflowInput(t *testing.T) {
	// Arrange
	workflowName := "test-workflow"
	params := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}
	
	// Act
	input := CreateWorkflowInput(workflowName, params)
	
	// Assert
	assert.Equal(t, workflowName, input.data["workflowName"], "Should set workflow name")
	assert.Equal(t, params, input.data["parameters"], "Should set parameters")
}

func TestCreateRetryableInput(t *testing.T) {
	// Arrange
	operation := "risky-operation"
	maxRetries := 3
	
	// Act
	input := CreateRetryableInput(operation, maxRetries)
	
	// Assert
	assert.Equal(t, operation, input.data["operation"], "Should set operation")
	assert.Equal(t, maxRetries, input.data["maxRetries"], "Should set max retries")
	assert.Equal(t, 0, input.data["currentAttempt"], "Should initialize current attempt to 0")
}

// Tests for history analysis

func TestFindHistoryEventsByType(t *testing.T) {
	// Arrange
	events := []HistoryEvent{
		createTestHistoryEvent(types.HistoryEventTypeExecutionStarted, time.Now()),
		createTestHistoryEvent(types.HistoryEventTypeTaskStateEntered, time.Now().Add(1*time.Second)),
		createTestHistoryEvent(types.HistoryEventTypeTaskStateExited, time.Now().Add(2*time.Second)),
		createTestHistoryEvent(types.HistoryEventTypeExecutionSucceeded, time.Now().Add(3*time.Second)),
	}
	
	// Act
	startEvents := findHistoryEventsByType(events, types.HistoryEventTypeExecutionStarted)
	stateEvents := findHistoryEventsByType(events, types.HistoryEventTypeTaskStateEntered)
	
	// Assert
	assert.Len(t, startEvents, 1, "Should find one execution started event")
	assert.Len(t, stateEvents, 1, "Should find one state entered event")
	assert.Equal(t, types.HistoryEventTypeExecutionStarted, startEvents[0].Type, "Should return correct event type")
}

func TestCalculateExecutionDuration(t *testing.T) {
	// Arrange
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)
	events := []HistoryEvent{
		{
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: startTime,
		},
		{
			Type:      types.HistoryEventTypeExecutionSucceeded,
			Timestamp: endTime,
		},
	}
	
	// Act
	duration := calculateExecutionDuration(events)
	
	// Assert
	assert.Equal(t, 5*time.Minute, duration, "Should calculate correct execution duration")
}

// Benchmark tests for performance-critical functions

func BenchmarkNewInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewInput()
	}
}

func BenchmarkInputToJSON(b *testing.B) {
	input := NewInput().
		Set("message", "Hello, World!").
		Set("count", 42).
		Set("enabled", true)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := input.ToJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateStateMachineName(b *testing.B) {
	stateMachineName := "my-test-state-machine"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateStateMachineName(stateMachineName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindHistoryEventsByType(b *testing.B) {
	events := make([]HistoryEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = createTestHistoryEvent(types.HistoryEventTypeTaskStateEntered, time.Now())
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findHistoryEventsByType(events, types.HistoryEventTypeTaskStateEntered)
	}
}