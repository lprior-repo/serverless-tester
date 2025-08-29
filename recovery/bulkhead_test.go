package recovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBulkhead tests bulkhead isolation pattern for different service types
func TestBulkhead(t *testing.T) {
	t.Run("creates bulkhead with correct configuration", func(t *testing.T) {
		config := BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 10,
			QueueSize:    20,
			Timeout:      30 * time.Second,
		}

		bulkhead := NewBulkhead(config)

		assert.Equal(t, "test-bulkhead", bulkhead.Name())
		assert.Equal(t, 10, bulkhead.MaxConcurrent())
		assert.Equal(t, 20, bulkhead.QueueSize())
	})

	t.Run("allows execution within limits", func(t *testing.T) {
		bulkhead := NewBulkhead(BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 2,
			QueueSize:    1,
			Timeout:      100 * time.Millisecond,
		})

		ctx := context.Background()
		
		result, err := ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("queues requests when concurrent limit reached", func(t *testing.T) {
		bulkhead := NewBulkhead(BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    2,
			Timeout:      1 * time.Second,
		})

		ctx := context.Background()
		
		// Start a long-running operation to fill the slot
		done := make(chan bool)
		go func() {
			ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "long-running", nil
			})
			done <- true
		}()

		// Give first operation time to start
		time.Sleep(10 * time.Millisecond)

		// This should queue
		result, err := ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			return "queued", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "queued", result)

		// Wait for first operation to complete
		<-done
	})

	t.Run("rejects requests when queue is full", func(t *testing.T) {
		bulkhead := NewBulkhead(BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    1,
			Timeout:      100 * time.Millisecond,
		})

		ctx := context.Background()

		// Fill the concurrent slot
		longRunningDone := make(chan bool)
		go func() {
			ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
				time.Sleep(200 * time.Millisecond)
				return "long-running", nil
			})
			longRunningDone <- true
		}()

		// Fill the queue
		queuedDone := make(chan bool)
		go func() {
			ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
				time.Sleep(50 * time.Millisecond)
				return "queued", nil
			})
			queuedDone <- true
		}()

		// Give operations time to start
		time.Sleep(10 * time.Millisecond)

		// This should be rejected
		result, err := ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			return "should-be-rejected", nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bulkhead is full")
		assert.Empty(t, result)

		// Cleanup
		<-queuedDone
		<-longRunningDone
	})

	t.Run("respects timeout", func(t *testing.T) {
		bulkhead := NewBulkhead(BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    0,
			Timeout:      50 * time.Millisecond,
		})

		ctx := context.Background()

		result, err := ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return "timeout-test", nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Empty(t, result)
	})

	t.Run("tracks metrics correctly", func(t *testing.T) {
		bulkhead := NewBulkhead(BulkheadConfig{
			Name:         "test-bulkhead",
			MaxConcurrent: 5,
			QueueSize:    10,
			Timeout:      1 * time.Second,
		})

		ctx := context.Background()

		// Execute some operations
		ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			return "success1", nil
		})
		ExecuteWithBulkhead(ctx, bulkhead, func() (string, error) {
			return "", errors.New("failure1")
		})

		metrics := bulkhead.GetMetrics()

		assert.Equal(t, "test-bulkhead", metrics.Name)
		assert.Equal(t, int64(1), metrics.TotalSuccesses)
		assert.Equal(t, int64(1), metrics.TotalFailures)
		assert.Equal(t, int64(2), metrics.TotalRequests)
	})
}

// TestServiceTypeBulkheads tests bulkhead isolation for different AWS service types
func TestServiceTypeBulkheads(t *testing.T) {
	t.Run("creates service-specific bulkheads", func(t *testing.T) {
		manager := NewBulkheadManager()

		// Configure different bulkheads for different services
		configs := map[string]BulkheadConfig{
			"lambda": {
				Name:         "lambda-bulkhead",
				MaxConcurrent: 100,
				QueueSize:    200,
				Timeout:      30 * time.Second,
			},
			"dynamodb": {
				Name:         "dynamodb-bulkhead",
				MaxConcurrent: 50,
				QueueSize:    100,
				Timeout:      10 * time.Second,
			},
			"s3": {
				Name:         "s3-bulkhead",
				MaxConcurrent: 20,
				QueueSize:    50,
				Timeout:      60 * time.Second,
			},
		}

		for serviceType, config := range configs {
			manager.AddBulkhead(serviceType, config)
		}

		// Test that each service has its own bulkhead
		lambdaBulkhead := manager.GetBulkhead("lambda")
		dynamoBulkhead := manager.GetBulkhead("dynamodb")
		s3Bulkhead := manager.GetBulkhead("s3")

		assert.NotNil(t, lambdaBulkhead)
		assert.NotNil(t, dynamoBulkhead)
		assert.NotNil(t, s3Bulkhead)

		assert.Equal(t, 100, lambdaBulkhead.MaxConcurrent())
		assert.Equal(t, 50, dynamoBulkhead.MaxConcurrent())
		assert.Equal(t, 20, s3Bulkhead.MaxConcurrent())
	})

	t.Run("executes operations with service-specific bulkheads", func(t *testing.T) {
		manager := NewBulkheadManager()

		manager.AddBulkhead("lambda", BulkheadConfig{
			Name:         "lambda-bulkhead",
			MaxConcurrent: 2,
			QueueSize:    1,
			Timeout:      1 * time.Second,
		})

		manager.AddBulkhead("dynamodb", BulkheadConfig{
			Name:         "dynamodb-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    1,
			Timeout:      1 * time.Second,
		})

		ctx := context.Background()

		// Lambda operations should not affect DynamoDB operations
		lambdaResult, err := ExecuteWithServiceBulkhead(ctx, manager, "lambda", func() (string, error) {
			return "lambda-success", nil
		})

		dynamoResult, err2 := ExecuteWithServiceBulkhead(ctx, manager, "dynamodb", func() (string, error) {
			return "dynamo-success", nil
		})

		assert.NoError(t, err)
		assert.NoError(t, err2)
		assert.Equal(t, "lambda-success", lambdaResult)
		assert.Equal(t, "dynamo-success", dynamoResult)
	})

	t.Run("isolates failures between services", func(t *testing.T) {
		manager := NewBulkheadManager()

		manager.AddBulkhead("service1", BulkheadConfig{
			Name:         "service1-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    0,
			Timeout:      100 * time.Millisecond,
		})

		manager.AddBulkhead("service2", BulkheadConfig{
			Name:         "service2-bulkhead",
			MaxConcurrent: 1,
			QueueSize:    0,
			Timeout:      100 * time.Millisecond,
		})

		ctx := context.Background()

		// Fill service1 bulkhead with long-running operation
		go func() {
			ExecuteWithServiceBulkhead(ctx, manager, "service1", func() (string, error) {
				time.Sleep(200 * time.Millisecond)
				return "service1-slow", nil
			})
		}()

		// Give service1 operation time to start
		time.Sleep(10 * time.Millisecond)

		// Service1 should be busy, but service2 should still work
		service1Result, err1 := ExecuteWithServiceBulkhead(ctx, manager, "service1", func() (string, error) {
			return "service1-should-fail", nil
		})

		service2Result, err2 := ExecuteWithServiceBulkhead(ctx, manager, "service2", func() (string, error) {
			return "service2-should-work", nil
		})

		// Service1 should fail due to bulkhead being full
		assert.Error(t, err1)
		assert.Empty(t, service1Result)

		// Service2 should succeed
		assert.NoError(t, err2)
		assert.Equal(t, "service2-should-work", service2Result)
	})
}