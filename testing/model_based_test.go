package testutils

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

// ========== TDD: RED PHASE - Model-Based Testing Tests ==========

// RED: Test state machine creation and basic operations
func TestModelBasedTesting_ShouldCreateStateMachine(t *testing.T) {
	modelTest := NewModelBasedTest()
	
	// Create a simple state machine for Lambda function lifecycle
	sm := modelTest.CreateStateMachine("LambdaLifecycle")
	
	// Define states
	sm.AddState("Created", true)  // initial state
	sm.AddState("Deployed", false)
	sm.AddState("Running", false)
	sm.AddState("Stopped", false)
	sm.AddState("Deleted", false)
	
	// Add transitions
	sm.AddTransition("Created", "Deploy", "Deployed")
	sm.AddTransition("Deployed", "Start", "Running")
	sm.AddTransition("Running", "Stop", "Stopped")
	sm.AddTransition("Stopped", "Start", "Running")
	sm.AddTransition("Deployed", "Delete", "Deleted")
	sm.AddTransition("Stopped", "Delete", "Deleted")
	
	assert.Equal(t, "Created", sm.GetCurrentState(), "Should start in initial state")
	assert.True(t, sm.CanTransition("Deploy"), "Should allow valid transition")
	assert.False(t, sm.CanTransition("Start"), "Should not allow invalid transition")
}

// RED: Test state machine execution with commands
func TestModelBasedTesting_ShouldExecuteCommands(t *testing.T) {
	modelTest := NewModelBasedTest()
	sm := modelTest.CreateStateMachine("LambdaLifecycle")
	
	// Setup basic Lambda lifecycle state machine
	setupLambdaStateMachine(sm)
	
	// Execute a sequence of commands
	result := sm.ExecuteCommand("Deploy")
	assert.True(t, result.Success, "Deploy should succeed from Created state")
	assert.Equal(t, "Deployed", sm.GetCurrentState(), "Should transition to Deployed")
	
	result = sm.ExecuteCommand("Start")
	assert.True(t, result.Success, "Start should succeed from Deployed state")
	assert.Equal(t, "Running", sm.GetCurrentState(), "Should transition to Running")
	
	result = sm.ExecuteCommand("Delete")
	assert.False(t, result.Success, "Delete should fail from Running state")
	assert.Equal(t, "Running", sm.GetCurrentState(), "Should remain in Running state")
}

// RED: Test model-based property validation
func TestModelBasedTesting_ShouldValidateStateInvariants(t *testing.T) {
	modelTest := NewModelBasedTest()
	sm := modelTest.CreateStateMachine("LambdaLifecycle")
	setupLambdaStateMachine(sm)
	
	// Define invariants that should always hold
	invariants := []StateInvariant{
		{
			Name: "Cannot delete running function",
			Check: func(state string, lastCommand string) bool {
				return !(state == "Running" && lastCommand == "Delete")
			},
		},
		{
			Name: "Cannot start from deleted state",
			Check: func(state string, lastCommand string) bool {
				return !(state == "Deleted" && lastCommand == "Start")
			},
		},
	}
	
	sm.AddInvariants(invariants)
	
	// Run random testing to validate invariants
	violations := modelTest.RunRandomTesting(sm, 100)
	assert.Empty(t, violations, "Should not violate any invariants")
}

// RED: Test parallel state machine execution
func TestModelBasedTesting_ShouldHandleParallelExecution(t *testing.T) {
	modelTest := NewModelBasedTest()
	
	// Create multiple state machines for parallel testing
	machines := make([]*StateMachine, 3)
	for i := 0; i < 3; i++ {
		machines[i] = modelTest.CreateStateMachine(fmt.Sprintf("ParallelLambda-%d", i))
		setupLambdaStateMachine(machines[i])
	}
	
	// Run parallel execution test
	results := modelTest.RunParallelTesting(machines, 10) // Reduced iterations for testing
	
	assert.Len(t, results, 3, "Should have results for all machines")
	
	// Check that at least some machines executed commands
	hasExecutedCommands := false
	for _, result := range results {
		if len(result.ExecutedCommands) > 0 {
			hasExecutedCommands = true
		}
		assert.Empty(t, result.InvariantViolations, "Machine %d should not violate invariants", result.MachineID)
	}
	
	assert.True(t, hasExecutedCommands, "At least one machine should have executed commands")
}

// RED: Test state machine with AWS service integration
func TestModelBasedTesting_ShouldIntegrateWithAWSServices(t *testing.T) {
	modelTest := NewModelBasedTest()
	sm := modelTest.CreateStateMachine("DynamoDBTable")
	
	// Define DynamoDB table lifecycle
	sm.AddState("NotExists", true)
	sm.AddState("Creating", false)
	sm.AddState("Active", false)
	sm.AddState("Deleting", false)
	sm.AddState("Deleted", false)
	
	// Add transitions with AWS operations
	sm.AddTransition("NotExists", "CreateTable", "Creating")
	sm.AddTransition("Creating", "TableReady", "Active")
	sm.AddTransition("Active", "DeleteTable", "Deleting")
	sm.AddTransition("Deleting", "TableDeleted", "Deleted")
	
	// Add AWS-specific invariants
	awsInvariants := []StateInvariant{
		{
			Name: "Cannot operate on non-existent table",
			Check: func(state string, lastCommand string) bool {
				if state == "NotExists" || state == "Deleted" {
					return lastCommand != "PutItem" && lastCommand != "GetItem"
				}
				return true
			},
		},
	}
	
	sm.AddInvariants(awsInvariants)
	
	// Test with AWS-specific commands
	result := sm.ExecuteCommand("CreateTable")
	assert.True(t, result.Success, "Should create table successfully")
	assert.Equal(t, "Creating", sm.GetCurrentState(), "Should be in Creating state")
}

// Helper function to setup Lambda lifecycle state machine
func setupLambdaStateMachine(sm *StateMachine) {
	sm.AddState("Created", true)
	sm.AddState("Deployed", false)
	sm.AddState("Running", false)
	sm.AddState("Stopped", false)
	sm.AddState("Deleted", false)
	
	sm.AddTransition("Created", "Deploy", "Deployed")
	sm.AddTransition("Deployed", "Start", "Running")
	sm.AddTransition("Running", "Stop", "Stopped")
	sm.AddTransition("Stopped", "Start", "Running")
	sm.AddTransition("Deployed", "Delete", "Deleted")
	sm.AddTransition("Stopped", "Delete", "Deleted")
}