package testutils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ModelBasedTest provides model-based testing with state machines
type ModelBasedTest struct {
	rng *RandomGenerator
}

// StateMachine represents a finite state machine for testing
type StateMachine struct {
	name           string
	states         map[string]bool  // state name -> is initial state
	currentState   string
	transitions    map[string]map[string]string  // from state -> command -> to state
	invariants     []StateInvariant
	commandHistory []CommandResult
	mutex          sync.RWMutex
}

// StateInvariant represents an invariant that must hold in any state
type StateInvariant struct {
	Name  string
	Check func(state string, lastCommand string) bool
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	Command    string
	FromState  string
	ToState    string
	Success    bool
	Timestamp  time.Time
	Error      error
}

// ParallelTestResult represents the result of parallel state machine testing
type ParallelTestResult struct {
	MachineID           int
	ExecutedCommands    []CommandResult
	InvariantViolations []InvariantViolation
	FinalState          string
}

// InvariantViolation represents a violation of a state invariant
type InvariantViolation struct {
	InvariantName string
	State         string
	Command       string
	Timestamp     time.Time
}

// NewModelBasedTest creates a new model-based testing instance
func NewModelBasedTest() *ModelBasedTest {
	return &ModelBasedTest{
		rng: &RandomGenerator{seed: time.Now().UnixNano()},
	}
}

// CreateStateMachine creates a new state machine for testing
func (m *ModelBasedTest) CreateStateMachine(name string) *StateMachine {
	return &StateMachine{
		name:           name,
		states:         make(map[string]bool),
		transitions:    make(map[string]map[string]string),
		invariants:     make([]StateInvariant, 0),
		commandHistory: make([]CommandResult, 0),
	}
}

// AddState adds a state to the state machine
func (sm *StateMachine) AddState(stateName string, isInitial bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.states[stateName] = isInitial
	if isInitial {
		sm.currentState = stateName
	}
	
	// Initialize transitions map for this state
	if sm.transitions[stateName] == nil {
		sm.transitions[stateName] = make(map[string]string)
	}
}

// AddTransition adds a transition between states
func (sm *StateMachine) AddTransition(fromState, command, toState string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if sm.transitions[fromState] == nil {
		sm.transitions[fromState] = make(map[string]string)
	}
	sm.transitions[fromState][command] = toState
}

// AddInvariants adds state invariants to validate
func (sm *StateMachine) AddInvariants(invariants []StateInvariant) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.invariants = append(sm.invariants, invariants...)
}

// GetCurrentState returns the current state of the machine
func (sm *StateMachine) GetCurrentState() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	return sm.currentState
}

// CanTransition checks if a command can be executed from current state
func (sm *StateMachine) CanTransition(command string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	if stateTransitions, exists := sm.transitions[sm.currentState]; exists {
		_, canTransition := stateTransitions[command]
		return canTransition
	}
	return false
}

// ExecuteCommand executes a command and transitions state if valid
func (sm *StateMachine) ExecuteCommand(command string) CommandResult {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	result := CommandResult{
		Command:   command,
		FromState: sm.currentState,
		ToState:   sm.currentState,
		Success:   false,
		Timestamp: time.Now(),
	}
	
	// Check if transition is valid
	if stateTransitions, exists := sm.transitions[sm.currentState]; exists {
		if toState, canTransition := stateTransitions[command]; canTransition {
			// Execute transition
			sm.currentState = toState
			result.ToState = toState
			result.Success = true
			
			// Check invariants
			for _, invariant := range sm.invariants {
				if !invariant.Check(sm.currentState, command) {
					result.Error = fmt.Errorf("invariant violation: %s", invariant.Name)
					result.Success = false
					break
				}
			}
		} else {
			result.Error = fmt.Errorf("invalid command %s from state %s", command, sm.currentState)
		}
	} else {
		result.Error = fmt.Errorf("no transitions defined for state %s", sm.currentState)
	}
	
	// Record command in history
	sm.commandHistory = append(sm.commandHistory, result)
	
	return result
}

// RunRandomTesting runs random command sequences to test the state machine
func (m *ModelBasedTest) RunRandomTesting(sm *StateMachine, iterations int) []InvariantViolation {
	violations := make([]InvariantViolation, 0)
	
	for i := 0; i < iterations; i++ {
		// Get valid commands for current state
		validCommands := m.getValidCommandsForState(sm, sm.GetCurrentState())
		
		if len(validCommands) == 0 {
			// Reset to initial state if stuck
			sm.Reset()
			continue
		}
		
		// Choose a random valid command
		commandIndex := m.rng.secureRandomInt(0, len(validCommands)-1)
		command := validCommands[commandIndex]
		
		result := sm.ExecuteCommand(command)
		
		// Check for actual invariant violations (not just invalid transitions)
		if result.Error != nil && result.Success == false {
			// Only record if it's an invariant violation, not an invalid transition
			errorMsg := result.Error.Error()
			if !isInvalidTransitionError(errorMsg) {
				violation := InvariantViolation{
					InvariantName: errorMsg,
					State:         result.FromState,
					Command:       command,
					Timestamp:     result.Timestamp,
				}
				violations = append(violations, violation)
			}
		}
	}
	
	return violations
}

// getValidCommandsForState returns commands that are valid from the given state
func (m *ModelBasedTest) getValidCommandsForState(sm *StateMachine, state string) []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	if stateTransitions, exists := sm.transitions[state]; exists {
		commands := make([]string, 0, len(stateTransitions))
		for command := range stateTransitions {
			commands = append(commands, command)
		}
		return commands
	}
	
	return []string{}
}

// isInvalidTransitionError checks if error is due to invalid transition vs invariant violation
func isInvalidTransitionError(errorMsg string) bool {
	return strings.Contains(errorMsg, "invalid command") || 
		   strings.Contains(errorMsg, "no transitions defined")
}

// RunParallelTesting runs multiple state machines in parallel
func (m *ModelBasedTest) RunParallelTesting(machines []*StateMachine, commandsPerMachine int) []ParallelTestResult {
	results := make([]ParallelTestResult, len(machines))
	var wg sync.WaitGroup
	
	for i, machine := range machines {
		wg.Add(1)
		go func(idx int, sm *StateMachine) {
			defer wg.Done()
			
			result := ParallelTestResult{
				MachineID:           idx,
				ExecutedCommands:    make([]CommandResult, 0),
				InvariantViolations: make([]InvariantViolation, 0),
			}
			
			// Get command history before running tests
			sm.mutex.RLock()
			initialHistory := len(sm.commandHistory)
			sm.mutex.RUnlock()
			
			// Run random testing on this machine
			violations := m.RunRandomTesting(sm, commandsPerMachine)
			result.InvariantViolations = violations
			
			// Get command history
			sm.mutex.RLock()
			result.ExecutedCommands = append(result.ExecutedCommands, sm.commandHistory[initialHistory:]...)
			result.FinalState = sm.currentState
			sm.mutex.RUnlock()
			
			results[idx] = result
		}(i, machine)
	}
	
	wg.Wait()
	return results
}

// getAllCommands extracts all possible commands from state machine transitions
func (m *ModelBasedTest) getAllCommands(sm *StateMachine) []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	commandSet := make(map[string]bool)
	
	for _, stateTransitions := range sm.transitions {
		for command := range stateTransitions {
			commandSet[command] = true
		}
	}
	
	commands := make([]string, 0, len(commandSet))
	for command := range commandSet {
		commands = append(commands, command)
	}
	
	return commands
}

// Reset resets the state machine to its initial state
func (sm *StateMachine) Reset() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Find initial state
	for state, isInitial := range sm.states {
		if isInitial {
			sm.currentState = state
			break
		}
	}
	
	// Clear history
	sm.commandHistory = make([]CommandResult, 0)
}

// GetCommandHistory returns the history of executed commands
func (sm *StateMachine) GetCommandHistory() []CommandResult {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	history := make([]CommandResult, len(sm.commandHistory))
	copy(history, sm.commandHistory)
	return history
}

// GetStateTransitions returns all possible transitions from current state
func (sm *StateMachine) GetStateTransitions() map[string]string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	if stateTransitions, exists := sm.transitions[sm.currentState]; exists {
		// Return a copy
		result := make(map[string]string)
		for k, v := range stateTransitions {
			result[k] = v
		}
		return result
	}
	
	return make(map[string]string)
}

// ValidateInvariants checks all invariants against current state
func (sm *StateMachine) ValidateInvariants(lastCommand string) []InvariantViolation {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	violations := make([]InvariantViolation, 0)
	
	for _, invariant := range sm.invariants {
		if !invariant.Check(sm.currentState, lastCommand) {
			violation := InvariantViolation{
				InvariantName: invariant.Name,
				State:         sm.currentState,
				Command:       lastCommand,
				Timestamp:     time.Now(),
			}
			violations = append(violations, violation)
		}
	}
	
	return violations
}