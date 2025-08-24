package eventbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

// CreateArchive creates an EventBridge archive and panics on error (Terratest pattern)
func CreateArchive(ctx *TestContext, config ArchiveConfig) ArchiveResult {
	result, err := CreateArchiveE(ctx, config)
	if err != nil {
		ctx.T.Errorf("CreateArchive failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// CreateArchiveE creates an EventBridge archive and returns error
func CreateArchiveE(ctx *TestContext, config ArchiveConfig) (ArchiveResult, error) {
	startTime := time.Now()
	
	logOperation("create_archive", map[string]interface{}{
		"archive_name":     config.ArchiveName,
		"event_source_arn": config.EventSourceArn,
		"retention_days":   config.RetentionDays,
	})
	
	if config.ArchiveName == "" {
		err := fmt.Errorf("archive name cannot be empty")
		logResult("create_archive", false, time.Since(startTime), err)
		return ArchiveResult{}, err
	}
	
	if config.EventSourceArn == "" {
		err := fmt.Errorf("event source ARN cannot be empty")
		logResult("create_archive", false, time.Since(startTime), err)
		return ArchiveResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Create the archive
	input := &eventbridge.CreateArchiveInput{
		ArchiveName:    &config.ArchiveName,
		EventSourceArn: &config.EventSourceArn,
		Description:    &config.Description,
	}
	
	// Add event pattern if specified
	if config.EventPattern != "" {
		if !ValidateEventPattern(config.EventPattern) {
			err := fmt.Errorf("%w: %s", ErrInvalidEventPattern, config.EventPattern)
			logResult("create_archive", false, time.Since(startTime), err)
			return ArchiveResult{}, err
		}
		input.EventPattern = &config.EventPattern
	}
	
	// Set retention days (default to 0 for indefinite retention)
	if config.RetentionDays > 0 {
		input.RetentionDays = aws.Int32(config.RetentionDays)
	}
	
	// Execute with retry logic
	retryConfig := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, retryConfig)
			time.Sleep(delay)
		}
		
		response, err := client.CreateArchive(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := ArchiveResult{
			ArchiveName:    config.ArchiveName,
			ArchiveArn:     safeString(response.ArchiveArn),
			State:          response.State,
			EventSourceArn: config.EventSourceArn,
			CreationTime:   response.CreationTime,
		}
		
		logResult("create_archive", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("create_archive", false, time.Since(startTime), lastErr)
	return ArchiveResult{}, fmt.Errorf("failed to create archive after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}

// DeleteArchive deletes an EventBridge archive and panics on error (Terratest pattern)
func DeleteArchive(ctx *TestContext, archiveName string) {
	err := DeleteArchiveE(ctx, archiveName)
	if err != nil {
		ctx.T.Errorf("DeleteArchive failed: %v", err)
		ctx.T.FailNow()
	}
}

// DeleteArchiveE deletes an EventBridge archive and returns error
func DeleteArchiveE(ctx *TestContext, archiveName string) error {
	startTime := time.Now()
	
	logOperation("delete_archive", map[string]interface{}{
		"archive_name": archiveName,
	})
	
	if archiveName == "" {
		err := fmt.Errorf("archive name cannot be empty")
		logResult("delete_archive", false, time.Since(startTime), err)
		return err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DeleteArchiveInput{
			ArchiveName: &archiveName,
		}
		
		_, err := client.DeleteArchive(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("delete_archive", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("delete_archive", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to delete archive after %d attempts: %w", config.MaxAttempts, lastErr)
}

// DescribeArchive describes an EventBridge archive and panics on error (Terratest pattern)
func DescribeArchive(ctx *TestContext, archiveName string) ArchiveResult {
	result, err := DescribeArchiveE(ctx, archiveName)
	if err != nil {
		ctx.T.Errorf("DescribeArchive failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DescribeArchiveE describes an EventBridge archive and returns error
func DescribeArchiveE(ctx *TestContext, archiveName string) (ArchiveResult, error) {
	startTime := time.Now()
	
	logOperation("describe_archive", map[string]interface{}{
		"archive_name": archiveName,
	})
	
	if archiveName == "" {
		err := fmt.Errorf("archive name cannot be empty")
		logResult("describe_archive", false, time.Since(startTime), err)
		return ArchiveResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DescribeArchiveInput{
			ArchiveName: &archiveName,
		}
		
		response, err := client.DescribeArchive(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := ArchiveResult{
			ArchiveName:    safeString(response.ArchiveName),
			ArchiveArn:     safeString(response.ArchiveArn),
			State:          response.State,
			EventSourceArn: safeString(response.EventSourceArn),
			CreationTime:   response.CreationTime,
		}
		
		logResult("describe_archive", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("describe_archive", false, time.Since(startTime), lastErr)
	return ArchiveResult{}, fmt.Errorf("failed to describe archive after %d attempts: %w", config.MaxAttempts, lastErr)
}

// ListArchives lists EventBridge archives and panics on error (Terratest pattern)
func ListArchives(ctx *TestContext, eventSourceArn string) []ArchiveResult {
	results, err := ListArchivesE(ctx, eventSourceArn)
	if err != nil {
		ctx.T.Errorf("ListArchives failed: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// ListArchivesE lists EventBridge archives and returns error
func ListArchivesE(ctx *TestContext, eventSourceArn string) ([]ArchiveResult, error) {
	startTime := time.Now()
	
	logOperation("list_archives", map[string]interface{}{
		"event_source_arn": eventSourceArn,
	})
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.ListArchivesInput{}
		
		// Add event source ARN filter if specified
		if eventSourceArn != "" {
			input.EventSourceArn = &eventSourceArn
		}
		
		response, err := client.ListArchives(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Convert results
		results := make([]ArchiveResult, len(response.Archives))
		for i, archive := range response.Archives {
			results[i] = ArchiveResult{
				ArchiveName:    safeString(archive.ArchiveName),
				State:          archive.State,
				EventSourceArn: safeString(archive.EventSourceArn),
				CreationTime:   archive.CreationTime,
			}
		}
		
		logResult("list_archives", true, time.Since(startTime), nil)
		return results, nil
	}
	
	logResult("list_archives", false, time.Since(startTime), lastErr)
	return nil, fmt.Errorf("failed to list archives after %d attempts: %w", config.MaxAttempts, lastErr)
}

// StartReplay starts an EventBridge replay and panics on error (Terratest pattern)
func StartReplay(ctx *TestContext, config ReplayConfig) ReplayResult {
	result, err := StartReplayE(ctx, config)
	if err != nil {
		ctx.T.Errorf("StartReplay failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// StartReplayE starts an EventBridge replay and returns error
func StartReplayE(ctx *TestContext, config ReplayConfig) (ReplayResult, error) {
	startTime := time.Now()
	
	logOperation("start_replay", map[string]interface{}{
		"replay_name":      config.ReplayName,
		"event_source_arn": config.EventSourceArn,
		"event_start_time": config.EventStartTime,
		"event_end_time":   config.EventEndTime,
	})
	
	if config.ReplayName == "" {
		err := fmt.Errorf("replay name cannot be empty")
		logResult("start_replay", false, time.Since(startTime), err)
		return ReplayResult{}, err
	}
	
	if config.EventSourceArn == "" {
		err := fmt.Errorf("event source ARN cannot be empty")
		logResult("start_replay", false, time.Since(startTime), err)
		return ReplayResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Create the replay
	input := &eventbridge.StartReplayInput{
		ReplayName:     &config.ReplayName,
		EventSourceArn: &config.EventSourceArn,
		EventStartTime: &config.EventStartTime,
		EventEndTime:   &config.EventEndTime,
		Destination:    &config.Destination,
		Description:    &config.Description,
	}
	
	// Execute with retry logic
	retryConfig := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, retryConfig)
			time.Sleep(delay)
		}
		
		response, err := client.StartReplay(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := ReplayResult{
			ReplayName:      config.ReplayName,
			ReplayArn:       safeString(response.ReplayArn),
			State:           response.State,
			EventStartTime:  &config.EventStartTime,
			EventEndTime:    &config.EventEndTime,
			ReplayStartTime: response.ReplayStartTime,
		}
		
		logResult("start_replay", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("start_replay", false, time.Since(startTime), lastErr)
	return ReplayResult{}, fmt.Errorf("failed to start replay after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}

// DescribeReplay describes an EventBridge replay and panics on error (Terratest pattern)
func DescribeReplay(ctx *TestContext, replayName string) ReplayResult {
	result, err := DescribeReplayE(ctx, replayName)
	if err != nil {
		ctx.T.Errorf("DescribeReplay failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DescribeReplayE describes an EventBridge replay and returns error
func DescribeReplayE(ctx *TestContext, replayName string) (ReplayResult, error) {
	startTime := time.Now()
	
	logOperation("describe_replay", map[string]interface{}{
		"replay_name": replayName,
	})
	
	if replayName == "" {
		err := fmt.Errorf("replay name cannot be empty")
		logResult("describe_replay", false, time.Since(startTime), err)
		return ReplayResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DescribeReplayInput{
			ReplayName: &replayName,
		}
		
		response, err := client.DescribeReplay(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := ReplayResult{
			ReplayName:      safeString(response.ReplayName),
			ReplayArn:       safeString(response.ReplayArn),
			State:           response.State,
			StateReason:     safeString(response.StateReason),
			EventStartTime:  response.EventStartTime,
			EventEndTime:    response.EventEndTime,
			ReplayStartTime: response.ReplayStartTime,
			ReplayEndTime:   response.ReplayEndTime,
		}
		
		logResult("describe_replay", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("describe_replay", false, time.Since(startTime), lastErr)
	return ReplayResult{}, fmt.Errorf("failed to describe replay after %d attempts: %w", config.MaxAttempts, lastErr)
}

// CancelReplay cancels an EventBridge replay and panics on error (Terratest pattern)
func CancelReplay(ctx *TestContext, replayName string) {
	err := CancelReplayE(ctx, replayName)
	if err != nil {
		ctx.T.Errorf("CancelReplay failed: %v", err)
		ctx.T.FailNow()
	}
}

// CancelReplayE cancels an EventBridge replay and returns error
func CancelReplayE(ctx *TestContext, replayName string) error {
	startTime := time.Now()
	
	logOperation("cancel_replay", map[string]interface{}{
		"replay_name": replayName,
	})
	
	if replayName == "" {
		err := fmt.Errorf("replay name cannot be empty")
		logResult("cancel_replay", false, time.Since(startTime), err)
		return err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.CancelReplayInput{
			ReplayName: &replayName,
		}
		
		_, err := client.CancelReplay(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("cancel_replay", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("cancel_replay", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to cancel replay after %d attempts: %w", config.MaxAttempts, lastErr)
}