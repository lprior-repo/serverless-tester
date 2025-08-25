package eventbridge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// AssertRuleExists asserts that an EventBridge rule exists and panics if not (Terratest pattern)
func AssertRuleExists(ctx *TestContext, ruleName string, eventBusName string) {
	err := AssertRuleExistsE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("AssertRuleExists failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertRuleExistsE asserts that an EventBridge rule exists and returns error
func AssertRuleExistsE(ctx *TestContext, ruleName string, eventBusName string) error {
	_, err := DescribeRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("rule '%s' does not exist in event bus '%s': %w", ruleName, eventBusName, err)
	}
	return nil
}

// AssertRuleEnabled asserts that an EventBridge rule is enabled and panics if not (Terratest pattern)
func AssertRuleEnabled(ctx *TestContext, ruleName string, eventBusName string) {
	err := AssertRuleEnabledE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("AssertRuleEnabled failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertRuleEnabledE asserts that an EventBridge rule is enabled and returns error
func AssertRuleEnabledE(ctx *TestContext, ruleName string, eventBusName string) error {
	rule, err := DescribeRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("failed to describe rule '%s': %w", ruleName, err)
	}
	
	if rule.State != RuleStateEnabled {
		return fmt.Errorf("rule '%s' is not enabled, current state: %s", ruleName, string(rule.State))
	}
	
	return nil
}

// AssertRuleDisabled asserts that an EventBridge rule is disabled and panics if not (Terratest pattern)
func AssertRuleDisabled(ctx *TestContext, ruleName string, eventBusName string) {
	err := AssertRuleDisabledE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("AssertRuleDisabled failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertRuleDisabledE asserts that an EventBridge rule is disabled and returns error
func AssertRuleDisabledE(ctx *TestContext, ruleName string, eventBusName string) error {
	rule, err := DescribeRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("failed to describe rule '%s': %w", ruleName, err)
	}
	
	if rule.State != RuleStateDisabled {
		return fmt.Errorf("rule '%s' is not disabled, current state: %s", ruleName, string(rule.State))
	}
	
	return nil
}

// AssertTargetExists asserts that a target exists for an EventBridge rule and panics if not (Terratest pattern)
func AssertTargetExists(ctx *TestContext, ruleName string, eventBusName string, targetID string) {
	err := AssertTargetExistsE(ctx, ruleName, eventBusName, targetID)
	if err != nil {
		ctx.T.Errorf("AssertTargetExists failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertTargetExistsE asserts that a target exists for an EventBridge rule and returns error
func AssertTargetExistsE(ctx *TestContext, ruleName string, eventBusName string, targetID string) error {
	targets, err := ListTargetsForRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("failed to list targets for rule '%s': %w", ruleName, err)
	}
	
	for _, target := range targets {
		if target.ID == targetID {
			return nil
		}
	}
	
	return fmt.Errorf("target '%s' not found for rule '%s' in event bus '%s'", targetID, ruleName, eventBusName)
}

// AssertTargetCount asserts that a rule has the expected number of targets and panics if not (Terratest pattern)
func AssertTargetCount(ctx *TestContext, ruleName string, eventBusName string, expectedCount int) {
	err := AssertTargetCountE(ctx, ruleName, eventBusName, expectedCount)
	if err != nil {
		ctx.T.Errorf("AssertTargetCount failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertTargetCountE asserts that a rule has the expected number of targets and returns error
func AssertTargetCountE(ctx *TestContext, ruleName string, eventBusName string, expectedCount int) error {
	targets, err := ListTargetsForRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("failed to list targets for rule '%s': %w", ruleName, err)
	}
	
	actualCount := len(targets)
	if actualCount != expectedCount {
		return fmt.Errorf("rule '%s' has %d targets, expected %d", ruleName, actualCount, expectedCount)
	}
	
	return nil
}

// AssertEventBusExists asserts that an EventBridge event bus exists and panics if not (Terratest pattern)
func AssertEventBusExists(ctx *TestContext, eventBusName string) {
	err := AssertEventBusExistsE(ctx, eventBusName)
	if err != nil {
		ctx.T.Errorf("AssertEventBusExists failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertEventBusExistsE asserts that an EventBridge event bus exists and returns error
func AssertEventBusExistsE(ctx *TestContext, eventBusName string) error {
	_, err := DescribeEventBusE(ctx, eventBusName)
	if err != nil {
		return fmt.Errorf("event bus '%s' does not exist: %w", eventBusName, err)
	}
	return nil
}

// AssertArchiveExists asserts that an EventBridge archive exists and panics if not (Terratest pattern)
func AssertArchiveExists(ctx *TestContext, archiveName string) {
	err := AssertArchiveExistsE(ctx, archiveName)
	if err != nil {
		ctx.T.Errorf("AssertArchiveExists failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertArchiveExistsE asserts that an EventBridge archive exists and returns error
func AssertArchiveExistsE(ctx *TestContext, archiveName string) error {
	_, err := DescribeArchiveE(ctx, archiveName)
	if err != nil {
		return fmt.Errorf("archive '%s' does not exist: %w", archiveName, err)
	}
	return nil
}

// AssertArchiveState asserts that an EventBridge archive is in the expected state and panics if not (Terratest pattern)
func AssertArchiveState(ctx *TestContext, archiveName string, expectedState types.ArchiveState) {
	err := AssertArchiveStateE(ctx, archiveName, expectedState)
	if err != nil {
		ctx.T.Errorf("AssertArchiveState failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertArchiveStateE asserts that an EventBridge archive is in the expected state and returns error
func AssertArchiveStateE(ctx *TestContext, archiveName string, expectedState types.ArchiveState) error {
	archive, err := DescribeArchiveE(ctx, archiveName)
	if err != nil {
		return fmt.Errorf("failed to describe archive '%s': %w", archiveName, err)
	}
	
	if archive.State != expectedState {
		return fmt.Errorf("archive '%s' is in state '%s', expected '%s'", archiveName, string(archive.State), string(expectedState))
	}
	
	return nil
}

// AssertReplayExists asserts that an EventBridge replay exists and panics if not (Terratest pattern)
func AssertReplayExists(ctx *TestContext, replayName string) {
	err := AssertReplayExistsE(ctx, replayName)
	if err != nil {
		ctx.T.Errorf("AssertReplayExists failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertReplayExistsE asserts that an EventBridge replay exists and returns error
func AssertReplayExistsE(ctx *TestContext, replayName string) error {
	_, err := DescribeReplayE(ctx, replayName)
	if err != nil {
		return fmt.Errorf("replay '%s' does not exist: %w", replayName, err)
	}
	return nil
}

// AssertReplayCompleted asserts that an EventBridge replay has completed and panics if not (Terratest pattern)
func AssertReplayCompleted(ctx *TestContext, replayName string) {
	err := AssertReplayCompletedE(ctx, replayName)
	if err != nil {
		ctx.T.Errorf("AssertReplayCompleted failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertReplayCompletedE asserts that an EventBridge replay has completed and returns error
func AssertReplayCompletedE(ctx *TestContext, replayName string) error {
	replay, err := DescribeReplayE(ctx, replayName)
	if err != nil {
		return fmt.Errorf("failed to describe replay '%s': %w", replayName, err)
	}
	
	if replay.State != ReplayStateCompleted {
		return fmt.Errorf("replay '%s' is in state '%s', expected '%s'", replayName, string(replay.State), string(ReplayStateCompleted))
	}
	
	return nil
}

// AssertPatternMatches asserts that an event matches a given pattern and panics if not (Terratest pattern)
func AssertPatternMatches(ctx *TestContext, event CustomEvent, pattern string) {
	err := AssertPatternMatchesE(ctx, event, pattern)
	if err != nil {
		ctx.T.Errorf("AssertPatternMatches failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertPatternMatchesE asserts that an event matches a given pattern and returns error
func AssertPatternMatchesE(ctx *TestContext, event CustomEvent, pattern string) error {
	if !ValidateEventPattern(pattern) {
		return fmt.Errorf("invalid event pattern: %s", pattern)
	}
	
	// Use EventBridge's TestEventPattern API for accurate matching
	client := createEventBridgeClient(ctx)
	
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.TestEventPatternInput{
			EventPattern: &pattern,
			Event:        &event.Detail,
		}
		
		response, err := client.TestEventPattern(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		if !response.Result {
			return fmt.Errorf("event does not match pattern")
		}
		
		return nil
	}
	
	return fmt.Errorf("failed to test pattern after %d attempts: %w", config.MaxAttempts, lastErr)
}

// AssertEventSent asserts that an event was successfully sent and panics if not (Terratest pattern)
func AssertEventSent(ctx *TestContext, result PutEventResult) {
	err := AssertEventSentE(ctx, result)
	if err != nil {
		ctx.T.Errorf("AssertEventSent failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertEventSentE asserts that an event was successfully sent and returns error
func AssertEventSentE(ctx *TestContext, result PutEventResult) error {
	if !result.Success {
		return fmt.Errorf("event was not successfully sent: %s - %s", result.ErrorCode, result.ErrorMessage)
	}
	
	if result.EventID == "" {
		return fmt.Errorf("event ID is empty")
	}
	
	return nil
}

// AssertEventBatchSent asserts that all events in a batch were successfully sent and panics if not (Terratest pattern)
func AssertEventBatchSent(ctx *TestContext, result PutEventsResult) {
	err := AssertEventBatchSentE(ctx, result)
	if err != nil {
		ctx.T.Errorf("AssertEventBatchSent failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertEventBatchSentE asserts that all events in a batch were successfully sent and returns error
func AssertEventBatchSentE(ctx *TestContext, result PutEventsResult) error {
	if result.FailedEntryCount > 0 {
		var failedEntries []string
		for i, entry := range result.Entries {
			if !entry.Success {
				failedEntries = append(failedEntries, fmt.Sprintf("Entry %d: %s - %s", i, entry.ErrorCode, entry.ErrorMessage))
			}
		}
		totalEntries := len(result.Entries)
		return fmt.Errorf("%d out of %d events failed to send:\n%s", result.FailedEntryCount, totalEntries, strings.Join(failedEntries, "\n"))
	}
	
	for i, entry := range result.Entries {
		if entry.EventID == "" {
			return fmt.Errorf("entry %d has empty event ID", i)
		}
	}
	
	return nil
}

// AssertRuleHasPattern asserts that a rule has the expected event pattern and panics if not (Terratest pattern)
func AssertRuleHasPattern(ctx *TestContext, ruleName string, eventBusName string, expectedPattern string) {
	err := AssertRuleHasPatternE(ctx, ruleName, eventBusName, expectedPattern)
	if err != nil {
		ctx.T.Errorf("AssertRuleHasPattern failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertRuleHasPatternE asserts that a rule has the expected event pattern and returns error
func AssertRuleHasPatternE(ctx *TestContext, ruleName string, eventBusName string, expectedPattern string) error {
	rule, err := DescribeRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		return fmt.Errorf("failed to describe rule '%s': %w", ruleName, err)
	}
	
	// Check if rule has no event pattern (e.g., scheduled rule)
	if rule.EventPattern == "" {
		return fmt.Errorf("rule has no event pattern: '%s' (it may be a scheduled rule)", ruleName)
	}
	
	// Compare actual pattern with expected pattern
	if rule.EventPattern != expectedPattern {
		return fmt.Errorf("rule has different pattern. Rule '%s' - Actual: %s, Expected: %s", ruleName, rule.EventPattern, expectedPattern)
	}
	
	return nil
}

// WaitForReplayCompletion waits for an EventBridge replay to complete and panics on timeout (Terratest pattern)
func WaitForReplayCompletion(ctx *TestContext, replayName string, timeout time.Duration) {
	err := WaitForReplayCompletionE(ctx, replayName, timeout)
	if err != nil {
		ctx.T.Errorf("WaitForReplayCompletion failed: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForReplayCompletionE waits for an EventBridge replay to complete and returns error on timeout
func WaitForReplayCompletionE(ctx *TestContext, replayName string, timeout time.Duration) error {
	startTime := time.Now()
	checkInterval := 10 * time.Second
	
	for time.Since(startTime) < timeout {
		replay, err := DescribeReplayE(ctx, replayName)
		if err != nil {
			return fmt.Errorf("failed to describe replay: %w", err)
		}
		
		switch replay.State {
		case ReplayStateCompleted:
			return nil
		case ReplayStateFailed:
			return fmt.Errorf("replay failed: %s", replay.StateReason)
		case ReplayStateStarting, ReplayStateRunning:
			time.Sleep(checkInterval)
		default:
			return fmt.Errorf("unknown replay state: %s", string(replay.State))
		}
	}
	
	return fmt.Errorf("timeout waiting for replay '%s' to complete after %v", replayName, timeout)
}

// WaitForArchiveState waits for an EventBridge archive to reach the expected state and panics on timeout (Terratest pattern)
func WaitForArchiveState(ctx *TestContext, archiveName string, expectedState types.ArchiveState, timeout time.Duration) {
	err := WaitForArchiveStateE(ctx, archiveName, expectedState, timeout)
	if err != nil {
		ctx.T.Errorf("WaitForArchiveState failed: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForArchiveStateE waits for an EventBridge archive to reach the expected state and returns error on timeout
func WaitForArchiveStateE(ctx *TestContext, archiveName string, expectedState types.ArchiveState, timeout time.Duration) error {
	startTime := time.Now()
	checkInterval := 5 * time.Second
	
	for time.Since(startTime) < timeout {
		archive, err := DescribeArchiveE(ctx, archiveName)
		if err != nil {
			return fmt.Errorf("failed to describe archive: %w", err)
		}
		
		if archive.State == expectedState {
			return nil
		}
		
		time.Sleep(checkInterval)
	}
	
	return fmt.Errorf("timeout waiting for archive '%s' to reach state '%s' after %v", archiveName, string(expectedState), timeout)
}