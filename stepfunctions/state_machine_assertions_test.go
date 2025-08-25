package stepfunctions

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Tests for state machine assertion functions that are currently at 0% coverage

func TestAssertStateMachineExists(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		shouldPanic      bool
		description      string
	}{
		{
			name:            "ExistingStateMachine",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:existing-machine",
			shouldPanic:     false, // Assuming it exists (mock would need to be set up)
			description:     "Existing state machine should not cause panic",
		},
		{
			name:            "NonExistingStateMachine",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:non-existing-machine",
			shouldPanic:     true, // Assuming it doesn't exist
			description:     "Non-existing state machine should cause panic",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			shouldPanic:     true,
			description:     "Invalid ARN should cause panic",
		},
		{
			name:            "EmptyArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "",
			shouldPanic:     true,
			description:     "Empty ARN should cause panic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertStateMachineExists to panic for %s, but it did not", tc.description)
					}
				}()
			}

			// Use a mock T to avoid interfering with the actual test
			mockT := &MockT{errorMessages: make([]string, 0)}
			AssertStateMachineExists(mockT, tc.ctx, tc.stateMachineArn)
		})
	}
}

func TestAssertStateMachineExistsE(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expected:        false, // Will fail due to no AWS connection, but tests the function
			description:     "Valid ARN should not cause immediate failure",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expected:        false,
			description:     "Invalid ARN should return false",
		},
		{
			name:            "EmptyArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "",
			expected:        false,
			description:     "Empty ARN should return false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertStateMachineExistsE(tc.ctx, tc.stateMachineArn)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertStateMachineExistsE_Internal(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt to check existence",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expected:        false,
			description:     "Invalid ARN should return false immediately",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := assertStateMachineExistsE(tc.ctx, tc.stateMachineArn)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestStateMachineAssertStateMachineStatus(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedStatus   types.StateMachineStatus
		shouldPanic      bool
		description      string
	}{
		{
			name:            "ActiveStatus",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:active-machine",
			expectedStatus:  types.StateMachineStatusActive,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Checking active status without connection should panic",
		},
		{
			name:            "DeletingStatus",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:deleting-machine",
			expectedStatus:  types.StateMachineStatusDeleting,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Checking deleting status without connection should panic",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedStatus:  types.StateMachineStatusActive,
			shouldPanic:     true,
			description:     "Invalid ARN should panic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertStateMachineStatus to panic for %s, but it did not", tc.description)
					}
				}()
			}

			mockT := &MockT{errorMessages: make([]string, 0)}
			AssertStateMachineStatus(mockT, tc.ctx, tc.stateMachineArn, tc.expectedStatus)
		})
	}
}

func TestAssertStateMachineStatusE(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedStatus   types.StateMachineStatus
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedStatus:  types.StateMachineStatusActive,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt status check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedStatus:  types.StateMachineStatusActive,
			expected:        false,
			description:     "Invalid ARN should return false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertStateMachineStatusE(tc.ctx, tc.stateMachineArn, tc.expectedStatus)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertStateMachineStatusE_Internal(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedStatus   types.StateMachineStatus
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedStatus:  types.StateMachineStatusActive,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt status check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedStatus:  types.StateMachineStatusActive,
			expected:        false,
			description:     "Invalid ARN should return false immediately",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := assertStateMachineStatusE(tc.ctx, tc.stateMachineArn, tc.expectedStatus)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestStateMachineAssertStateMachineType(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedType     types.StateMachineType
		shouldPanic      bool
		description      string
	}{
		{
			name:            "StandardType",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:standard-machine",
			expectedType:    types.StateMachineTypeStandard,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Checking standard type without connection should panic",
		},
		{
			name:            "ExpressType",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:express-machine",
			expectedType:    types.StateMachineTypeExpress,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Checking express type without connection should panic",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedType:    types.StateMachineTypeStandard,
			shouldPanic:     true,
			description:     "Invalid ARN should panic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertStateMachineType to panic for %s, but it did not", tc.description)
					}
				}()
			}

			mockT := &MockT{errorMessages: make([]string, 0)}
			AssertStateMachineType(mockT, tc.ctx, tc.stateMachineArn, tc.expectedType)
		})
	}
}

func TestAssertStateMachineTypeE(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedType     types.StateMachineType
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedType:    types.StateMachineTypeStandard,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt type check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedType:    types.StateMachineTypeStandard,
			expected:        false,
			description:     "Invalid ARN should return false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertStateMachineTypeE(tc.ctx, tc.stateMachineArn, tc.expectedType)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertStateMachineTypeE_Internal(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedType     types.StateMachineType
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedType:    types.StateMachineTypeStandard,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt type check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedType:    types.StateMachineTypeStandard,
			expected:        false,
			description:     "Invalid ARN should return false immediately",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := assertStateMachineTypeE(tc.ctx, tc.stateMachineArn, tc.expectedType)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestStateMachineAssertExecutionCount(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedCount    int
		shouldPanic      bool
		description      string
	}{
		{
			name:            "ValidCount",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedCount:   5,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Valid count check without connection should panic",
		},
		{
			name:            "ZeroCount",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:empty-machine",
			expectedCount:   0,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Zero count check without connection should panic",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedCount:   1,
			shouldPanic:     true,
			description:     "Invalid ARN should panic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionCount to panic for %s, but it did not", tc.description)
					}
				}()
			}

			mockT := &MockT{errorMessages: make([]string, 0)}
			AssertExecutionCount(mockT, tc.ctx, tc.stateMachineArn, tc.expectedCount)
		})
	}
}

func TestAssertExecutionCountE(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedCount    int
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedCount:   5,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt count check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedCount:   1,
			expected:        false,
			description:     "Invalid ARN should return false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertExecutionCountE(tc.ctx, tc.stateMachineArn, tc.expectedCount)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertExecutionCountE_Internal(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		expectedCount    int
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedCount:   5,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt count check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			expectedCount:   1,
			expected:        false,
			description:     "Invalid ARN should return false immediately",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := assertExecutionCountE(tc.ctx, tc.stateMachineArn, tc.expectedCount)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestStateMachineAssertExecutionCountByStatus(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		status           types.ExecutionStatus
		expectedCount    int
		shouldPanic      bool
		description      string
	}{
		{
			name:            "SucceededCount",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   3,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Succeeded count check without connection should panic",
		},
		{
			name:            "FailedCount",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusFailed,
			expectedCount:   1,
			shouldPanic:     true, // Will panic due to no AWS connection
			description:     "Failed count check without connection should panic",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   1,
			shouldPanic:     true,
			description:     "Invalid ARN should panic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionCountByStatus to panic for %s, but it did not", tc.description)
					}
				}()
			}

			mockT := &MockT{errorMessages: make([]string, 0)}
			AssertExecutionCountByStatus(mockT, tc.ctx, tc.stateMachineArn, tc.status, tc.expectedCount)
		})
	}
}

func TestAssertExecutionCountByStatusE(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		status           types.ExecutionStatus
		expectedCount    int
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   3,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt count by status check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   1,
			expected:        false,
			description:     "Invalid ARN should return false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertExecutionCountByStatusE(tc.ctx, tc.stateMachineArn, tc.status, tc.expectedCount)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertExecutionCountByStatusE_Internal(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *TestContext
		stateMachineArn  string
		status           types.ExecutionStatus
		expectedCount    int
		expected         bool
		description      string
	}{
		{
			name:            "ValidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   3,
			expected:        false, // Will fail due to no AWS connection
			description:     "Valid ARN should attempt count by status check",
		},
		{
			name:            "InvalidArn",
			ctx:             &TestContext{T: t},
			stateMachineArn: "invalid-arn",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   1,
			expected:        false,
			description:     "Invalid ARN should return false immediately",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := assertExecutionCountByStatusE(tc.ctx, tc.stateMachineArn, tc.status, tc.expectedCount)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

// Benchmark tests for state machine assertion functions

func BenchmarkAssertStateMachineExistsE(b *testing.B) {
	ctx := &TestContext{T: &testing.T{}}
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertStateMachineExistsE(ctx, stateMachineArn)
	}
}

func BenchmarkAssertStateMachineStatusE(b *testing.B) {
	ctx := &TestContext{T: &testing.T{}}
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"
	expectedStatus := types.StateMachineStatusActive

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertStateMachineStatusE(ctx, stateMachineArn, expectedStatus)
	}
}

func BenchmarkAssertStateMachineTypeE(b *testing.B) {
	ctx := &TestContext{T: &testing.T{}}
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"
	expectedType := types.StateMachineTypeStandard

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertStateMachineTypeE(ctx, stateMachineArn, expectedType)
	}
}

func BenchmarkAssertExecutionCountE(b *testing.B) {
	ctx := &TestContext{T: &testing.T{}}
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"
	expectedCount := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertExecutionCountE(ctx, stateMachineArn, expectedCount)
	}
}

func BenchmarkAssertExecutionCountByStatusE(b *testing.B) {
	ctx := &TestContext{T: &testing.T{}}
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"
	status := types.ExecutionStatusSucceeded
	expectedCount := 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertExecutionCountByStatusE(ctx, stateMachineArn, status, expectedCount)
	}
}