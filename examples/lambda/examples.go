package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"vasdeference/lambda"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ExampleBasicUsage demonstrates pure functional Lambda patterns with mo.Option and lo.Pipe
func ExampleBasicUsage() {
	fmt.Println("Functional Lambda usage patterns with monadic operations:")
	
	// Example 1: Immutable Lambda configuration with functional options
	functionalConfig := lambda.NewFunctionalLambdaConfig("my-test-function",
		lambda.WithInvocationType(types.InvocationTypeRequestResponse),
		lambda.WithLogType(types.LogTypeTail),
		lambda.WithTimeout(30*time.Second),
		lambda.WithRetryConfig(3, 1*time.Second),
		lambda.WithLambdaMetadata("environment", "production"),
	)
	
	// Create immutable payload using functional composition
	payload := `{"message": "Hello, Lambda!"}`
	
	// Simulate functional Lambda invocation with monadic operations
	result := lambda.SimulateFunctionalLambdaInvocation(payload, functionalConfig)
	
	// Demonstrate monadic operations with result
	processedResult := result.
		MapResult(func(r lambda.FunctionalInvokeResult) lambda.FunctionalInvokeResult {
			return lambda.FunctionalInvokeResult{
				StatusCode:      r.StatusCode,
				Payload:         fmt.Sprintf(`{"enriched": %s, "timestamp": "%s"}`, r.Payload, time.Now().UTC().Format(time.RFC3339)),
				LogResult:       r.LogResult,
				ExecutedVersion: r.ExecutedVersion,
				FunctionError:   r.FunctionError,
				ExecutionTime:   r.ExecutionTime,
			}
		})
	
	// Use mo.Option pattern for safe result handling
	processedResult.GetResult().
		Map(func(invokeResult lambda.FunctionalInvokeResult) (lambda.FunctionalInvokeResult, bool) {
			fmt.Printf("Lambda invocation successful: Status=%d, Duration=%v\n",
				invokeResult.StatusCode, invokeResult.ExecutionTime)
			return invokeResult, true
		}).OrElse(func() lambda.FunctionalInvokeResult {
			fmt.Printf("Lambda invocation failed: %v\n", processedResult.GetError().MustGet())
			return lambda.FunctionalInvokeResult{}
		})
	
	// Example 2: Functional S3 event processing with pipeline composition
	s3ProcessingPipeline := func(bucketName, objectKey string) mo.Result[string] {
		return lo.Pipe3(
			bucketName,
			func(bucket string) string {
				return lambda.BuildS3Event(bucket, objectKey, "s3:ObjectCreated:Put")
			},
			func(event string) (string, error) {
				if len(event) == 0 {
					return "", fmt.Errorf("empty S3 event")
				}
				return event, nil
			},
			func(event string) (string, error) {
				// Validate event structure
				if len(event) < 100 {
					return "", fmt.Errorf("invalid S3 event structure")
				}
				return event, nil
			},
		).Try()
	}
	
	// Execute S3 processing pipeline
	s3Result := s3ProcessingPipeline("test-bucket", "test-data/file.json")
	s3Result.
		Map(func(event string) string {
			fmt.Printf("S3 event generated successfully: %s...\n", event[:50])
			return event
		}).
		MapErr(func(err error) error {
			fmt.Printf("S3 event generation failed: %v\n", err)
			return err
		})
	
	// Example 3: Async invocation with functional configuration
	asyncConfig := lambda.NewFunctionalLambdaConfig("async-processor",
		lambda.WithInvocationType(types.InvocationTypeEvent),
		lambda.WithLambdaMetadata("async", true),
		lambda.WithLambdaMetadata("priority", "high"),
	)
	
	asyncResult := lambda.SimulateFunctionalLambdaInvocation(`{"data": "async-test"}`, asyncConfig)
	if asyncResult.IsSuccess() {
		fmt.Printf("Async invocation successful: Duration=%v\n", asyncResult.GetDuration())
	}
}

// ExampleTableDrivenValidation demonstrates functional validation with mo.Result and payload validators
func ExampleTableDrivenValidation() {
	fmt.Println("Functional table-driven validation patterns:")
	
	// Immutable validation scenario structure
	type ValidationScenario struct {
		name         string
		functionName string
		payload      string
		expectSuccess bool
		validator    func(string) error
		description  string
	}
	
	// Create validation functions using pure functions
	validateUser := func(payload string) error {
		if len(payload) < 10 {
			return fmt.Errorf("payload too short")
		}
		return nil
	}
	
	validateOrder := func(payload string) error {
		if payload == `{}` {
			return fmt.Errorf("missing required fields")
		}
		return nil
	}
	
	// Define scenarios as immutable data structures
	scenarios := []ValidationScenario{
		{
			name:         "ValidUserPayload",
			functionName: "user-validator",
			payload:      `{"name": "John", "age": 30}`,
			expectSuccess: true,
			validator:    validateUser,
			description:  "Valid user payload should be processed successfully",
		},
		{
			name:         "InvalidUserPayload",
			functionName: "user-validator",
			payload:      `{"name": ""}`,
			expectSuccess: false,
			validator:    validateUser,
			description:  "Invalid user payload should return validation error",
		},
		{
			name:         "EmptyOrderPayload",
			functionName: "order-validator",
			payload:      `{}`,
			expectSuccess: false,
			validator:    validateOrder,
			description:  "Empty order payload should return missing fields error",
		},
	}
	
	// Process scenarios using functional composition
	results := lo.Map(scenarios, func(scenario ValidationScenario, _ int) mo.Result[lambda.FunctionalOperationResult] {
		// Create functional config with validator
		config := lambda.NewFunctionalLambdaConfig(scenario.functionName,
			lambda.WithPayloadValidator(scenario.validator),
			lambda.WithLambdaMetadata("scenario", scenario.name),
		)
		
		// Execute functional Lambda invocation
		result := lambda.SimulateFunctionalLambdaInvocation(scenario.payload, config)
		
		// Validate expectations using mo.Result
		if scenario.expectSuccess && result.IsSuccess() {
			return mo.Ok(result)
		} else if !scenario.expectSuccess && result.IsError() {
			return mo.Ok(result)
		} else {
			return mo.Err[lambda.FunctionalOperationResult](fmt.Errorf("expectation mismatch for scenario %s", scenario.name))
		}
	})
	
	// Display results using functional operations
	lo.ForEach(lo.Zip2(scenarios, results), func(pair lo.Tuple2[ValidationScenario, mo.Result[lambda.FunctionalOperationResult]], _ int) {
		scenario := pair.A
		result := pair.B
		
		fmt.Printf("Scenario %s: %s\n", scenario.name, scenario.description)
		
		result.
			Map(func(opResult lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
				fmt.Printf("  ✓ Validation passed - Duration: %v\n", opResult.GetDuration())
				return opResult
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Validation failed: %v\n", err)
				return err
			})
	})
}

// ExampleAdvancedEventProcessing demonstrates functional event processing with monadic composition
func ExampleAdvancedEventProcessing() {
	fmt.Println("Functional event processing with monadic composition:")
	
	// Event processor type for functional composition
	type EventProcessor struct {
		eventType string
		config    lambda.FunctionalLambdaConfig
		builder   func() mo.Result[string]
	}
	
	// Create event processors using pure functions
	eventProcessors := []EventProcessor{
		{
			eventType: "S3",
			config: lambda.NewFunctionalLambdaConfig("s3-processor",
				lambda.WithLambdaMetadata("event_source", "s3"),
				lambda.WithTimeout(10*time.Second),
			),
			builder: func() mo.Result[string] {
				return mo.Ok(lambda.BuildS3Event("data-bucket", "input/data.json", "s3:ObjectCreated:Put"))
			},
		},
		{
			eventType: "DynamoDB",
			config: lambda.NewFunctionalLambdaConfig("dynamo-processor",
				lambda.WithLambdaMetadata("event_source", "dynamodb"),
				lambda.WithRetryConfig(5, 2*time.Second),
			),
			builder: func() mo.Result[string] {
				keys := map[string]interface{}{
					"id": map[string]interface{}{"S": "item-123"},
				}
				return mo.Ok(lambda.BuildDynamoDBEvent("users-table", "INSERT", keys))
			},
		},
		{
			eventType: "SQS",
			config: lambda.NewFunctionalLambdaConfig("sqs-processor",
				lambda.WithLambdaMetadata("event_source", "sqs"),
				lambda.WithLambdaMetadata("batch_processing", true),
			),
			builder: func() mo.Result[string] {
				queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
				messageBody := `{"order_id": "order-456", "status": "pending"}`
				return mo.Ok(lambda.BuildSQSEvent(queueUrl, messageBody))
			},
		},
		{
			eventType: "APIGateway",
			config: lambda.NewFunctionalLambdaConfig("api-processor",
				lambda.WithLambdaMetadata("event_source", "apigateway"),
				lambda.WithLambdaMetadata("http_method", "POST"),
			),
			builder: func() mo.Result[string] {
				return mo.Ok(lambda.BuildAPIGatewayEvent("POST", "/api/orders", `{"item": "widget", "quantity": 5}`))
			},
		},
	}
	
	// Process events using functional composition and monadic operations
	processingResults := lo.Map(eventProcessors, func(processor EventProcessor, _ int) mo.Result[lambda.FunctionalOperationResult] {
		return processor.builder().
			FlatMap(func(event string) mo.Result[lambda.FunctionalOperationResult] {
				// Execute Lambda invocation with the generated event
				result := lambda.SimulateFunctionalLambdaInvocation(event, processor.config)
				
				if result.IsSuccess() {
					return mo.Ok(result)
				} else {
					return mo.Err[lambda.FunctionalOperationResult](result.GetError().MustGet())
				}
			})
	})
	
	// Display results with functional operations
	lo.ForEach(lo.Zip2(eventProcessors, processingResults), func(pair lo.Tuple2[EventProcessor, mo.Result[lambda.FunctionalOperationResult]], _ int) {
		processor := pair.A
		result := pair.B
		
		fmt.Printf("%s event processing:\n", processor.eventType)
		
		result.
			Map(func(opResult lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
				opResult.GetResult().
					Map(func(invokeResult lambda.FunctionalInvokeResult) (lambda.FunctionalInvokeResult, bool) {
						fmt.Printf("  ✓ Processed successfully - Status: %d, Duration: %v\n",
							invokeResult.StatusCode, invokeResult.ExecutionTime)
						return invokeResult, true
					})
				return opResult
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Processing failed: %v\n", err)
				return err
			})
	})
	
	// Demonstrate function composition for multi-event workflow
	workflowPipeline := func(events []string) mo.Result[[]lambda.FunctionalOperationResult] {
		workflowConfig := lambda.NewFunctionalLambdaConfig("workflow-processor",
			lambda.WithLambdaMetadata("workflow", "multi-event"),
			lambda.WithRetryConfig(2, 500*time.Millisecond),
		)
		
		// Process all events in sequence using functional composition
		results := lo.Map(events, func(event string, _ int) lambda.FunctionalOperationResult {
			return lambda.SimulateFunctionalLambdaInvocation(event, workflowConfig)
		})
		
		// Check if all results are successful
		allSuccessful := lo.EveryBy(results, func(result lambda.FunctionalOperationResult) bool {
			return result.IsSuccess()
		})
		
		if allSuccessful {
			return mo.Ok(results)
		} else {
			failedCount := lo.CountBy(results, func(result lambda.FunctionalOperationResult) bool {
				return result.IsError()
			})
			return mo.Err[[]lambda.FunctionalOperationResult](fmt.Errorf("%d events failed in workflow", failedCount))
		}
	}
	
	// Execute workflow pipeline
	sampleEvents := []string{
		lambda.BuildS3Event("workflow-bucket", "step1.json", "s3:ObjectCreated:Put"),
		lambda.BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/workflow-queue", `{"step": 2}`),
	}
	
	workflowResult := workflowPipeline(sampleEvents)
	workflowResult.
		Map(func(results []lambda.FunctionalOperationResult) []lambda.FunctionalOperationResult {
			fmt.Printf("\n✓ Workflow completed successfully with %d events\n", len(results))
			return results
		}).
		MapErr(func(err error) error {
			fmt.Printf("\n✗ Workflow failed: %v\n", err)
			return err
		})
}

// ExampleEventSourceMapping demonstrates functional event source mapping with immutable configurations
func ExampleEventSourceMapping() {
	fmt.Println("Functional event source mapping patterns:")
	
	// Immutable event source mapping configuration using functional options pattern
	type EventSourceMappingConfig struct {
		eventSourceArn   string
		functionName     string
		batchSize        int
		enabled          bool
		startingPosition types.EventSourcePosition
		metadata         map[string]interface{}
	}
	
	// Event source mapping option type
	type EventSourceMappingOption func(EventSourceMappingConfig) EventSourceMappingConfig
	
	// Create functional event source mapping configuration
	createEventSourceMappingConfig := func(eventSourceArn, functionName string, opts ...EventSourceMappingOption) EventSourceMappingConfig {
		base := EventSourceMappingConfig{
			eventSourceArn:   eventSourceArn,
			functionName:     functionName,
			batchSize:        10,
			enabled:          true,
			startingPosition: types.EventSourcePositionLatest,
			metadata:         make(map[string]interface{}),
		}
		
		return lo.Reduce(opts, func(config EventSourceMappingConfig, opt EventSourceMappingOption, _ int) EventSourceMappingConfig {
			return opt(config)
		}, base)
	}
	
	// Functional options for event source mapping
	withBatchSize := func(size int) EventSourceMappingOption {
		return func(config EventSourceMappingConfig) EventSourceMappingConfig {
			config.batchSize = size
			return config
		}
	}
	
	withStartingPosition := func(position types.EventSourcePosition) EventSourceMappingOption {
		return func(config EventSourceMappingConfig) EventSourceMappingConfig {
			config.startingPosition = position
			return config
		}
	}
	
	withMappingMetadata := func(key string, value interface{}) EventSourceMappingOption {
		return func(config EventSourceMappingConfig) EventSourceMappingConfig {
			config.metadata = lo.Assign(config.metadata, map[string]interface{}{key: value})
			return config
		}
	}
	
	// Create multiple mapping configurations using functional composition
	mappingConfigs := []EventSourceMappingConfig{
		createEventSourceMappingConfig(
			"arn:aws:sqs:us-east-1:123456789012:high-priority-queue",
			"high-priority-processor",
			withBatchSize(5),
			withStartingPosition(types.EventSourcePositionLatest),
			withMappingMetadata("priority", "high"),
			withMappingMetadata("environment", "production"),
		),
		createEventSourceMappingConfig(
			"arn:aws:sqs:us-east-1:123456789012:batch-processing-queue",
			"batch-processor",
			withBatchSize(20),
			withMappingMetadata("processing_mode", "batch"),
		),
		createEventSourceMappingConfig(
			"arn:aws:kinesis:us-east-1:123456789012:stream/data-stream",
			"stream-processor",
			withBatchSize(100),
			withStartingPosition(types.EventSourcePositionTrimHorizon),
			withMappingMetadata("stream_type", "kinesis"),
		),
	}
	
	// Simulate mapping validation using functional composition
	validateMappingConfig := func(config EventSourceMappingConfig) mo.Result[EventSourceMappingConfig] {
		return lo.Pipe4(
			config,
			func(c EventSourceMappingConfig) (EventSourceMappingConfig, error) {
				if c.functionName == "" {
					return c, fmt.Errorf("function name cannot be empty")
				}
				return c, nil
			},
			func(c EventSourceMappingConfig) (EventSourceMappingConfig, error) {
				if c.batchSize < 1 || c.batchSize > 1000 {
					return c, fmt.Errorf("batch size must be between 1 and 1000")
				}
				return c, nil
			},
			func(c EventSourceMappingConfig) (EventSourceMappingConfig, error) {
				if c.eventSourceArn == "" {
					return c, fmt.Errorf("event source ARN cannot be empty")
				}
				return c, nil
			},
		).Try()
	}
	
	// Validate all configurations using monadic operations
	validationResults := lo.Map(mappingConfigs, func(config EventSourceMappingConfig, _ int) mo.Result[EventSourceMappingConfig] {
		return validateMappingConfig(config)
	})
	
	// Display results using functional operations
	lo.ForEach(lo.Zip2(mappingConfigs, validationResults), func(pair lo.Tuple2[EventSourceMappingConfig, mo.Result[EventSourceMappingConfig]], _ int) {
		config := pair.A
		result := pair.B
		
		fmt.Printf("Event source mapping for %s:\n", config.functionName)
		
		result.
			Map(func(validConfig EventSourceMappingConfig) EventSourceMappingConfig {
				fmt.Printf("  ✓ Configuration valid - Batch Size: %d, ARN: %s...\n",
					validConfig.batchSize, validConfig.eventSourceArn[:50])
				if len(validConfig.metadata) > 0 {
					fmt.Printf("  ✓ Metadata: %v\n", validConfig.metadata)
				}
				return validConfig
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Configuration invalid: %v\n", err)
				return err
			})
	})
}

// ExamplePerformancePatterns demonstrates functional concurrent Lambda invocations and performance monitoring
func ExamplePerformancePatterns() {
	fmt.Println("Functional performance and load patterns:")
	
	// Create high-performance Lambda configuration
	highLoadConfig := lambda.NewFunctionalLambdaConfig("high-load-function",
		lambda.WithTimeout(5*time.Second),
		lambda.WithRetryConfig(2, 200*time.Millisecond),
		lambda.WithLambdaMetadata("load_test", true),
		lambda.WithLambdaMetadata("performance_tier", "high"),
	)
	
	// Functional concurrent invocation pattern
	concurrentInvocations := func(concurrency int, payload string, config lambda.FunctionalLambdaConfig) mo.Result[[]lambda.FunctionalOperationResult] {
		// Create concurrent invocation tasks
		tasks := lo.Times(concurrency, func(_ int) func() lambda.FunctionalOperationResult {
			return func() lambda.FunctionalOperationResult {
				return lambda.SimulateFunctionalLambdaInvocation(payload, config)
			}
		})
		
		// Execute tasks concurrently using goroutines
		results := make(chan lambda.FunctionalOperationResult, concurrency)
		
		for _, task := range tasks {
			go func(t func() lambda.FunctionalOperationResult) {
				results <- t()
			}(task)
		}
		
		// Collect all results
		collectedResults := lo.Times(concurrency, func(_ int) lambda.FunctionalOperationResult {
			return <-results
		})
		
		// Check if all invocations were successful
		allSuccessful := lo.EveryBy(collectedResults, func(result lambda.FunctionalOperationResult) bool {
			return result.IsSuccess()
		})
		
		if allSuccessful {
			return mo.Ok(collectedResults)
		} else {
			failedCount := lo.CountBy(collectedResults, func(result lambda.FunctionalOperationResult) bool {
				return result.IsError()
			})
			return mo.Err[[]lambda.FunctionalOperationResult](fmt.Errorf("%d out of %d concurrent invocations failed", failedCount, concurrency))
		}
	}
	
	// Execute concurrent performance test
	payload := `{"load_test": true, "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"}`
	concurrencyLevels := []int{5, 10, 20}
	
	performanceResults := lo.Map(concurrencyLevels, func(concurrency int, _ int) mo.Result[[]lambda.FunctionalOperationResult] {
		startTime := time.Now()
		result := concurrentInvocations(concurrency, payload, highLoadConfig)
		duration := time.Since(startTime)
		
		// Log performance metrics
		result.
			Map(func(results []lambda.FunctionalOperationResult) []lambda.FunctionalOperationResult {
				avgDuration := lo.Reduce(results, func(acc time.Duration, result lambda.FunctionalOperationResult, _ int) time.Duration {
					return acc + result.GetDuration()
				}, 0) / time.Duration(len(results))
				
				fmt.Printf("✓ Concurrency %d: %d invocations completed in %v (avg: %v)\n",
					concurrency, len(results), duration, avgDuration)
				return results
			}).
			MapErr(func(err error) error {
				fmt.Printf("✗ Concurrency %d failed in %v: %v\n", concurrency, duration, err)
				return err
			})
		
		return result
	})
	
	// Analyze overall performance using functional operations
	successfulTests := lo.Filter(performanceResults, func(result mo.Result[[]lambda.FunctionalOperationResult], _ int) bool {
		return result.IsOk()
	})
	
	fmt.Printf("\nPerformance summary: %d/%d concurrency tests passed\n", len(successfulTests), len(performanceResults))
	
	// Demonstrate timeout and resilience patterns
	timeoutConfigs := []lambda.FunctionalLambdaConfig{
		lambda.NewFunctionalLambdaConfig("fast-function",
			lambda.WithTimeout(1*time.Second),
			lambda.WithLambdaMetadata("timeout_test", "fast"),
		),
		lambda.NewFunctionalLambdaConfig("medium-function",
			lambda.WithTimeout(5*time.Second),
			lambda.WithRetryConfig(3, 500*time.Millisecond),
			lambda.WithLambdaMetadata("timeout_test", "medium"),
		),
		lambda.NewFunctionalLambdaConfig("slow-function",
			lambda.WithTimeout(15*time.Second),
			lambda.WithRetryConfig(1, 1*time.Second),
			lambda.WithLambdaMetadata("timeout_test", "slow"),
		),
	}
	
	// Test timeout configurations using functional composition
	timeoutResults := lo.Map(timeoutConfigs, func(config lambda.FunctionalLambdaConfig, _ int) lambda.FunctionalOperationResult {
		testPayload := fmt.Sprintf(`{"delay_simulation": %d}`, int(config.timeout.Seconds())/2)
		return lambda.SimulateFunctionalLambdaInvocation(testPayload, config)
	})
	
	// Display timeout test results
	fmt.Println("\nTimeout configuration tests:")
	lo.ForEach(lo.Zip2(timeoutConfigs, timeoutResults), func(pair lo.Tuple2[lambda.FunctionalLambdaConfig, lambda.FunctionalOperationResult], _ int) {
		config := pair.A
		result := pair.B
		
		timeoutTest := config.metadata["timeout_test"].(string)
		if result.IsSuccess() {
			fmt.Printf("  ✓ %s timeout test passed - Duration: %v\n", timeoutTest, result.GetDuration())
		} else {
			fmt.Printf("  ✗ %s timeout test failed: %v\n", timeoutTest, result.GetError().MustGet())
		}
	})
}

// ExampleErrorHandlingAndRecovery demonstrates functional error handling with mo.Result and recovery patterns
func ExampleErrorHandlingAndRecovery() {
	fmt.Println("Functional error handling and recovery patterns:")
	
	// Error scenario definitions using immutable structures
	type ErrorScenario struct {
		name         string
		functionName string
		payload      string
		errorType    string
		recoverable  bool
		config       lambda.FunctionalLambdaConfig
	}
	
	// Create error scenarios with functional configurations
	errorScenarios := []ErrorScenario{
		{
			name:         "DivisionByZero",
			functionName: "error-test-function",
			payload:      `{"trigger_error": "division_by_zero"}`,
			errorType:    "runtime_error",
			recoverable:  false,
			config: lambda.NewFunctionalLambdaConfig("error-test-function",
				lambda.WithRetryConfig(1, 100*time.Millisecond),
				lambda.WithLambdaMetadata("error_handling", true),
				lambda.WithLambdaMetadata("error_type", "runtime"),
			),
		},
		{
			name:         "NetworkTimeout",
			functionName: "network-function",
			payload:      `{"simulate_timeout": true}`,
			errorType:    "network_error",
			recoverable:  true,
			config: lambda.NewFunctionalLambdaConfig("network-function",
				lambda.WithRetryConfig(3, 500*time.Millisecond),
				lambda.WithTimeout(2*time.Second),
				lambda.WithLambdaMetadata("error_handling", true),
				lambda.WithLambdaMetadata("error_type", "network"),
			),
		},
		{
			name:         "TransientDatabaseError",
			functionName: "database-function",
			payload:      `{"simulate_db_error": "connection_lost"}`,
			errorType:    "database_error",
			recoverable:  true,
			config: lambda.NewFunctionalLambdaConfig("database-function",
				lambda.WithRetryConfig(5, 1*time.Second),
				lambda.WithLambdaMetadata("error_handling", true),
				lambda.WithLambdaMetadata("error_type", "database"),
			),
		},
	}
	
	// Simulate error conditions with functional recovery patterns
	processErrorScenario := func(scenario ErrorScenario) mo.Result[lambda.FunctionalOperationResult] {
		// Simulate different error conditions based on scenario type
		switch scenario.errorType {
		case "runtime_error":
			// Simulate non-recoverable runtime error
			errorResult := lambda.NewErrorFunctionalOperationResult(
				fmt.Errorf("runtime error: %s", scenario.payload),
				10*time.Millisecond,
				scenario.config.metadata,
			)
			return mo.Err[lambda.FunctionalOperationResult](fmt.Errorf("unrecoverable runtime error"))
			
		case "network_error":
			// Simulate recoverable network error with retry
			if scenario.recoverable {
				// Simulate successful recovery after retry
				successResult := lambda.SimulateFunctionalLambdaInvocation(scenario.payload, scenario.config)
				if successResult.IsSuccess() {
					return mo.Ok(successResult)
				}
			}
			return mo.Err[lambda.FunctionalOperationResult](fmt.Errorf("network error: timeout"))
			
		case "database_error":
			// Simulate transient database error with eventual recovery
			if scenario.recoverable {
				// Simulate recovery after multiple retries
				retryResult := lambda.SimulateFunctionalLambdaInvocation(scenario.payload, scenario.config)
				if retryResult.IsSuccess() {
					return mo.Ok(retryResult)
				}
			}
			return mo.Err[lambda.FunctionalOperationResult](fmt.Errorf("database error: connection failed"))
			
		default:
			// Default successful execution
			result := lambda.SimulateFunctionalLambdaInvocation(scenario.payload, scenario.config)
			return mo.Ok(result)
		}
	}
	
	// Process all error scenarios using functional composition
	scenarioResults := lo.Map(errorScenarios, func(scenario ErrorScenario, _ int) mo.Result[lambda.FunctionalOperationResult] {
		return processErrorScenario(scenario)
	})
	
	// Display results with recovery patterns
	lo.ForEach(lo.Zip2(errorScenarios, scenarioResults), func(pair lo.Tuple2[ErrorScenario, mo.Result[lambda.FunctionalOperationResult]], _ int) {
		scenario := pair.A
		result := pair.B
		
		fmt.Printf("Error scenario '%s' (%s):\n", scenario.name, scenario.errorType)
		
		result.
			Map(func(opResult lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
				if scenario.recoverable {
					fmt.Printf("  ✓ Recovered successfully after retries - Duration: %v\n", opResult.GetDuration())
				} else {
					fmt.Printf("  ✓ Handled gracefully - Duration: %v\n", opResult.GetDuration())
				}
				return opResult
			}).
			MapErr(func(err error) error {
				if scenario.recoverable {
					fmt.Printf("  ✗ Recovery failed: %v\n", err)
				} else {
					fmt.Printf("  ✗ Error handled (expected failure): %v\n", err)
				}
				return err
			})
	})
	
	// Demonstrate circuit breaker pattern using functional composition
	circuitBreakerPattern := func(invocations []string, config lambda.FunctionalLambdaConfig) mo.Result[[]lambda.FunctionalOperationResult] {
		const failureThreshold = 2
		var failureCount int
		
		results := lo.Map(invocations, func(payload string, index int) lambda.FunctionalOperationResult {
			// Simulate circuit breaker logic
			if failureCount >= failureThreshold {
				// Circuit is open - fail fast
				return lambda.NewErrorFunctionalOperationResult(
					fmt.Errorf("circuit breaker open - failing fast"),
					1*time.Millisecond,
					map[string]interface{}{"circuit_breaker": "open", "invocation": index},
				)
			}
			
			// Simulate invocation (some fail, some succeed)
			if index%3 == 0 { // Every third invocation fails
				failureCount++
				return lambda.NewErrorFunctionalOperationResult(
					fmt.Errorf("simulated failure %d", index),
					50*time.Millisecond,
					map[string]interface{}{"invocation": index, "status": "failed"},
				)
			}
			
			// Successful invocation
			return lambda.SimulateFunctionalLambdaInvocation(payload, config)
		})
		
		return mo.Ok(results)
	}
	
	// Test circuit breaker pattern
	circuitConfig := lambda.NewFunctionalLambdaConfig("circuit-breaker-function",
		lambda.WithLambdaMetadata("pattern", "circuit_breaker"),
		lambda.WithRetryConfig(1, 100*time.Millisecond),
	)
	
	testInvocations := lo.Times(8, func(i int) string {
		return fmt.Sprintf(`{"test_invocation": %d}`, i)
	})
	
	circuitResult := circuitBreakerPattern(testInvocations, circuitConfig)
	circuitResult.
		Map(func(results []lambda.FunctionalOperationResult) []lambda.FunctionalOperationResult {
			successful := lo.CountBy(results, func(result lambda.FunctionalOperationResult) bool {
				return result.IsSuccess()
			})
			failed := len(results) - successful
			fmt.Printf("\n✓ Circuit breaker test: %d successful, %d failed out of %d invocations\n",
				successful, failed, len(results))
			return results
		}).
		MapErr(func(err error) error {
			fmt.Printf("\n✗ Circuit breaker test failed: %v\n", err)
			return err
		})
}

// ExampleComprehensiveIntegration demonstrates functional serverless workflow with monadic composition
func ExampleComprehensiveIntegration() {
	fmt.Println("Functional comprehensive serverless workflow with monadic operations:")
	
	// Workflow step definition using immutable structures
	type WorkflowStep struct {
		name         string
		functionName string
		eventBuilder func() mo.Result[string]
		config       lambda.FunctionalLambdaConfig
		nextStep     mo.Option[string]
	}
	
	// Order data as immutable structure
	orderData := map[string]interface{}{
		"customer_id": "cust-123",
		"items": []map[string]interface{}{
			{"sku": "item-001", "quantity": 2, "price": 29.99},
			{"sku": "item-002", "quantity": 1, "price": 49.99},
		},
		"shipping_address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "Anytown",
			"state":  "CA",
			"zip":    "12345",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	// Create immutable payload using functional composition
	serializeOrderData := func(data map[string]interface{}) mo.Result[string] {
		bytes, err := json.Marshal(data)
		if err != nil {
			return mo.Err[string](fmt.Errorf("failed to serialize order data: %w", err))
		}
		return mo.Ok(string(bytes))
	}
	
	// Define workflow steps using functional composition
	workflowSteps := []WorkflowStep{
		{
			name:         "OrderValidation",
			functionName: "order-validator",
			eventBuilder: func() mo.Result[string] {
				return serializeOrderData(orderData).
					Map(func(orderPayload string) string {
						return lambda.BuildAPIGatewayEvent("POST", "/api/orders", orderPayload)
					})
			},
			config: lambda.NewFunctionalLambdaConfig("order-validator",
				lambda.WithLambdaMetadata("workflow_step", "validation"),
				lambda.WithTimeout(10*time.Second),
				lambda.WithRetryConfig(3, 1*time.Second),
			),
			nextStep: mo.Some("OrderProcessing"),
		},
		{
			name:         "OrderProcessing",
			functionName: "order-processor",
			eventBuilder: func() mo.Result[string] {
				processingPayload := lo.Assign(orderData, map[string]interface{}{
					"status":      "processing",
					"order_id":    "order-" + fmt.Sprintf("%d", time.Now().UnixNano()),
					"validated":   true,
					"created_at":  time.Now().UTC().Format(time.RFC3339),
				})
				return serializeOrderData(processingPayload).
					Map(func(payload string) string {
						return lambda.BuildSQSEvent(
							"https://sqs.us-east-1.amazonaws.com/123456789012/order-processing-queue",
							payload,
						)
					})
			},
			config: lambda.NewFunctionalLambdaConfig("order-processor",
				lambda.WithLambdaMetadata("workflow_step", "processing"),
				lambda.WithTimeout(30*time.Second),
				lambda.WithRetryConfig(5, 2*time.Second),
			),
			nextStep: mo.Some("InventoryUpdate"),
		},
		{
			name:         "InventoryUpdate",
			functionName: "inventory-updater",
			eventBuilder: func() mo.Result[string] {
				keys := map[string]interface{}{
					"order_id": map[string]interface{}{"S": "order-456"},
					"status":   map[string]interface{}{"S": "processed"},
				}
				return mo.Ok(lambda.BuildDynamoDBEvent("orders", "INSERT", keys))
			},
			config: lambda.NewFunctionalLambdaConfig("inventory-updater",
				lambda.WithLambdaMetadata("workflow_step", "inventory"),
				lambda.WithTimeout(15*time.Second),
			),
			nextStep: mo.Some("NotificationSender"),
		},
		{
			name:         "NotificationSender",
			functionName: "notification-sender",
			eventBuilder: func() mo.Result[string] {
				return mo.Ok(lambda.BuildSNSEvent(
					"arn:aws:sns:us-east-1:123456789012:order-confirmations",
					"Order processed successfully - Customer: cust-123",
				))
			},
			config: lambda.NewFunctionalLambdaConfig("notification-sender",
				lambda.WithLambdaMetadata("workflow_step", "notification"),
				lambda.WithTimeout(5*time.Second),
			),
			nextStep: mo.None[string](),
		},
	}
	
	// Execute workflow using functional composition and monadic operations
	type WorkflowResult struct {
		step   WorkflowStep
		result lambda.FunctionalOperationResult
		successful bool
	}
	
	executeWorkflowStep := func(step WorkflowStep) mo.Result[WorkflowResult] {
		return step.eventBuilder().
			FlatMap(func(event string) mo.Result[WorkflowResult] {
				// Execute the Lambda function with the generated event
				result := lambda.SimulateFunctionalLambdaInvocation(event, step.config)
				
				workflowResult := WorkflowResult{
					step:       step,
					result:     result,
					successful: result.IsSuccess(),
				}
				
				if result.IsSuccess() {
					return mo.Ok(workflowResult)
				} else {
					return mo.Err[WorkflowResult](fmt.Errorf("workflow step %s failed: %v", step.name, result.GetError().MustGet()))
				}
			})
	}
	
	// Execute all workflow steps in sequence using functional composition
	workflowResults := lo.Map(workflowSteps, func(step WorkflowStep, _ int) mo.Result[WorkflowResult] {
		return executeWorkflowStep(step)
	})
	
	// Analyze workflow results using functional operations
	successfulSteps := lo.Filter(workflowResults, func(result mo.Result[WorkflowResult], _ int) bool {
		return result.IsOk()
	})
	
	failedSteps := lo.Filter(workflowResults, func(result mo.Result[WorkflowResult], _ int) bool {
		return result.IsError()
	})
	
	// Display workflow execution results
	fmt.Printf("Workflow execution summary: %d/%d steps completed successfully\n", len(successfulSteps), len(workflowSteps))
	
	lo.ForEach(lo.Zip2(workflowSteps, workflowResults), func(pair lo.Tuple2[WorkflowStep, mo.Result[WorkflowResult]], index int) {
		step := pair.A
		result := pair.B
		
		fmt.Printf("Step %d - %s:\n", index+1, step.name)
		
		result.
			Map(func(workflowResult WorkflowResult) WorkflowResult {
				fmt.Printf("  ✓ Completed successfully - Duration: %v\n", workflowResult.result.GetDuration())
				// Check if there's a next step
				step.nextStep.
					Map(func(nextStepName string) (string, bool) {
						fmt.Printf("  → Next step: %s\n", nextStepName)
						return nextStepName, true
					}).OrElse(func() string {
						fmt.Printf("  ✓ Workflow completed\n")
						return ""
					})
				return workflowResult
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Failed: %v\n", err)
				return err
			})
	})
	
	// Calculate total workflow metrics using functional operations
	if len(failedSteps) == 0 {
		totalDuration := lo.Reduce(successfulSteps, func(acc time.Duration, result mo.Result[WorkflowResult], _ int) time.Duration {
			return acc + result.MustGet().result.GetDuration()
		}, 0)
		
		fmt.Printf("\n✓ Complete workflow executed successfully in %v\n", totalDuration)
		fmt.Println("✓ Order processing pipeline: API Gateway → SQS → DynamoDB → SNS")
	} else {
		fmt.Printf("\n✗ Workflow partially failed - %d steps failed\n", len(failedSteps))
	}
}

// ExampleDataFactories demonstrates functional payload creation with immutable patterns
func ExampleDataFactories() {
	fmt.Println("Functional data factory patterns with immutable payload creation:")
	
	// Immutable payload builder using functional composition
	type PayloadBuilder[T any] struct {
		data     T
		validators []func(T) error
		transforms []func(T) T
		metadata   map[string]interface{}
	}
	
	// Create new payload builder with functional options
	newPayloadBuilder := func(data any) PayloadBuilder[any] {
		return PayloadBuilder[any]{
			data:       data,
			validators: make([]func(any) error, 0),
			transforms: make([]func(any) any, 0),
			metadata:   make(map[string]interface{}),
		}
	}
	
	// Functional payload creation for different entity types
	type UserData struct {
		Name      string    `json:"name"`
		Age       int       `json:"age"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}
	
	type OrderData struct {
		CustomerID string    `json:"customer_id"`
		Amount     float64   `json:"amount"`
		Currency   string    `json:"currency"`
		Status     string    `json:"status"`
		Items      []Item    `json:"items"`
		CreatedAt  time.Time `json:"created_at"`
	}
	
	type Item struct {
		SKU      string  `json:"sku"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
	}
	
	// Pure functional payload creators using mo.Result for error handling
	createUserPayload := func(name string, age int, email string) mo.Result[string] {
		// Validate inputs using functional composition
		validationPipeline := lambda.ExecuteFunctionalValidationPipeline[UserData](UserData{
			Name:      name,
			Age:       age,
			Email:     email,
			CreatedAt: time.Now().UTC(),
		},
			func(user UserData) (UserData, error) {
				if user.Name == "" {
					return user, fmt.Errorf("name cannot be empty")
				}
				return user, nil
			},
			func(user UserData) (UserData, error) {
				if user.Age < 0 || user.Age > 150 {
					return user, fmt.Errorf("age must be between 0 and 150")
				}
				return user, nil
			},
			func(user UserData) (UserData, error) {
				if user.Email == "" {
					return user, fmt.Errorf("email cannot be empty")
				}
				return user, nil
			},
		)
		
		if validationPipeline.IsError() {
			return mo.Err[string](validationPipeline.GetError().MustGet())
		}
		
		// Serialize to JSON using pure function
		bytes, err := json.Marshal(validationPipeline.GetValue())
		if err != nil {
			return mo.Err[string](fmt.Errorf("failed to serialize user payload: %w", err))
		}
		
		return mo.Ok(string(bytes))
	}
	
	createOrderPayload := func(customerID string, amount float64, items []Item) mo.Result[string] {
		// Create immutable order with functional validation
		orderData := OrderData{
			CustomerID: customerID,
			Amount:     amount,
			Currency:   "USD",
			Status:     "pending",
			Items:      items,
			CreatedAt:  time.Now().UTC(),
		}
		
		// Validate order using functional composition
		validationPipeline := lambda.ExecuteFunctionalValidationPipeline[OrderData](orderData,
			func(order OrderData) (OrderData, error) {
				if order.CustomerID == "" {
					return order, fmt.Errorf("customer ID cannot be empty")
				}
				return order, nil
			},
			func(order OrderData) (OrderData, error) {
				if order.Amount <= 0 {
					return order, fmt.Errorf("amount must be positive")
				}
				return order, nil
			},
			func(order OrderData) (OrderData, error) {
				if len(order.Items) == 0 {
					return order, fmt.Errorf("order must have at least one item")
				}
				return order, nil
			},
			func(order OrderData) (OrderData, error) {
				// Validate total amount matches items
				itemsTotal := lo.Reduce(order.Items, func(acc float64, item Item, _ int) float64 {
					return acc + (item.Price * float64(item.Quantity))
				}, 0)
				
				if math.Abs(order.Amount-itemsTotal) > 0.01 {
					return order, fmt.Errorf("order amount %.2f does not match items total %.2f", order.Amount, itemsTotal)
				}
				return order, nil
			},
		)
		
		if validationPipeline.IsError() {
			return mo.Err[string](validationPipeline.GetError().MustGet())
		}
		
		// Serialize to JSON using pure function
		bytes, err := json.Marshal(validationPipeline.GetValue())
		if err != nil {
			return mo.Err[string](fmt.Errorf("failed to serialize order payload: %w", err))
		}
		
		return mo.Ok(string(bytes))
	}
	
	// Factory function composition for complex payload creation
	createLambdaInvocationPayload := func(functionName string, eventType string, data interface{}) mo.Result[string] {
		payloadConfig := lambda.NewFunctionalLambdaConfig(functionName,
			lambda.WithLambdaMetadata("event_type", eventType),
			lambda.WithLambdaMetadata("created_at", time.Now().UTC().Format(time.RFC3339)),
		)
		
		// Create event-specific payload
		switch eventType {
		case "s3":
			return mo.Ok(lambda.BuildS3Event("payload-bucket", "data.json", "s3:ObjectCreated:Put"))
		case "sqs":
			if str, ok := data.(string); ok {
				return mo.Ok(lambda.BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/payload-queue", str))
			}
			return mo.Err[string](fmt.Errorf("SQS payload data must be string"))
		case "api":
			if str, ok := data.(string); ok {
				return mo.Ok(lambda.BuildAPIGatewayEvent("POST", "/api/data", str))
			}
			return mo.Err[string](fmt.Errorf("API payload data must be string"))
		default:
			return mo.Err[string](fmt.Errorf("unsupported event type: %s", eventType))
		}
	}
	
	// Example usage with functional composition and error handling
	fmt.Println("Creating user payloads:")
	userExamples := []struct {
		name  string
		age   int
		email string
	}{
		{"Alice Johnson", 28, "alice@example.com"},
		{"Bob Smith", 35, "bob@example.com"},
		{"", 25, "invalid@example.com"}, // Invalid - empty name
		{"Charlie Brown", -5, "charlie@example.com"}, // Invalid - negative age
	}
	
	userResults := lo.Map(userExamples, func(example struct {
		name  string
		age   int
		email string
	}, _ int) mo.Result[string] {
		return createUserPayload(example.name, example.age, example.email)
	})
	
	lo.ForEach(lo.Zip2(userExamples, userResults), func(pair lo.Tuple2[struct {
		name  string
		age   int
		email string
	}, mo.Result[string]], _ int) {
		example := pair.A
		result := pair.B
		
		result.
			Map(func(payload string) string {
				fmt.Printf("✓ User payload created for %s: %s...\n", example.name, payload[:50])
				return payload
			}).
			MapErr(func(err error) error {
				fmt.Printf("✗ Failed to create user payload for %s: %v\n", example.name, err)
				return err
			})
	})
	
	// Create order payload with items
	fmt.Println("\nCreating order payloads:")
	orderItems := []Item{
		{SKU: "WIDGET-001", Quantity: 2, Price: 19.99},
		{SKU: "GADGET-002", Quantity: 1, Price: 49.99},
	}
	
	orderResult := createOrderPayload("cust-456", 89.97, orderItems)
	orderResult.
		Map(func(payload string) string {
			fmt.Printf("✓ Order payload created: %s...\n", payload[:80])
			
			// Demonstrate functional Lambda invocation with created payload
			orderConfig := lambda.NewFunctionalLambdaConfig("order-processor",
				lambda.WithLambdaMetadata("payload_source", "factory"),
				lambda.WithTimeout(15*time.Second),
			)
			
			invocationResult := lambda.SimulateFunctionalLambdaInvocation(payload, orderConfig)
			if invocationResult.IsSuccess() {
				fmt.Printf("✓ Order processing simulation successful\n")
			}
			
			return payload
		}).
		MapErr(func(err error) error {
			fmt.Printf("✗ Failed to create order payload: %v\n", err)
			return err
		})
	
	// Demonstrate event-specific payload factory
	fmt.Println("\nCreating event-specific payloads:")
	eventTypes := []string{"s3", "sqs", "api"}
	eventData := []interface{}{
		nil, // S3 doesn't need additional data
		`{"message": "Hello from SQS"}`,
		`{"action": "create", "resource": "user"}`,
	}
	
	lo.ForEach(lo.Zip2(eventTypes, eventData), func(pair lo.Tuple2[string, interface{}], _ int) {
		eventType := pair.A
		data := pair.B
		
		payloadResult := createLambdaInvocationPayload("event-handler", eventType, data)
		payloadResult.
			Map(func(payload string) string {
				fmt.Printf("✓ %s event payload created: %s...\n", strings.ToUpper(eventType), payload[:60])
				return payload
			}).
			MapErr(func(err error) error {
				fmt.Printf("✗ Failed to create %s event payload: %v\n", eventType, err)
				return err
			})
	})
}

// ExampleValidationHelpers demonstrates functional configuration validation with mo.Result patterns
func ExampleValidationHelpers() {
	fmt.Println("Functional configuration validation patterns:")
	
	// Immutable function configuration structure
	type FunctionalConfigSpec struct {
		functionName string
		runtime      types.Runtime
		handler      string
		timeout      int32
		memorySize   int32
		state        types.State
		environment  map[string]string
		metadata     map[string]interface{}
	}
	
	// Configuration validation function using functional composition
	validateConfigSpec := func(actual FunctionalConfigSpec, expected FunctionalConfigSpec) mo.Result[FunctionalConfigSpec] {
		return lambda.ExecuteFunctionalValidationPipeline[FunctionalConfigSpec](actual,
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				if config.functionName != expected.functionName {
					return config, fmt.Errorf("function name mismatch: expected %s, got %s", expected.functionName, config.functionName)
				}
				return config, nil
			},
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				if config.runtime != expected.runtime {
					return config, fmt.Errorf("runtime mismatch: expected %s, got %s", expected.runtime, config.runtime)
				}
				return config, nil
			},
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				if config.handler != expected.handler {
					return config, fmt.Errorf("handler mismatch: expected %s, got %s", expected.handler, config.handler)
				}
				return config, nil
			},
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				if config.timeout != expected.timeout {
					return config, fmt.Errorf("timeout mismatch: expected %d, got %d", expected.timeout, config.timeout)
				}
				return config, nil
			},
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				if config.memorySize != expected.memorySize {
					return config, fmt.Errorf("memory size mismatch: expected %d, got %d", expected.memorySize, config.memorySize)
				}
				return config, nil
			},
			func(config FunctionalConfigSpec) (FunctionalConfigSpec, error) {
				// Validate required environment variables
				for key, expectedValue := range expected.environment {
					if actualValue, exists := config.environment[key]; !exists {
						return config, fmt.Errorf("missing environment variable: %s", key)
					} else if actualValue != expectedValue {
						return config, fmt.Errorf("environment variable %s mismatch: expected %s, got %s", key, expectedValue, actualValue)
					}
				}
				return config, nil
			},
		).Try()
	}
	
	// Expected configuration specification
	expectedConfig := FunctionalConfigSpec{
		functionName: "production-function",
		runtime:      types.RuntimeNodejs18x,
		handler:      "index.handler",
		timeout:      30,
		memorySize:   256,
		state:        types.StateActive,
		environment: map[string]string{
			"NODE_ENV":  "production",
			"LOG_LEVEL": "info",
		},
		metadata: map[string]interface{}{
			"validation_type": "expected",
		},
	}
	
	// Test configurations with different validation scenarios
	testConfigurations := []FunctionalConfigSpec{
		{
			// Valid configuration
			functionName: "production-function",
			runtime:      types.RuntimeNodejs18x,
			handler:      "index.handler",
			timeout:      30,
			memorySize:   256,
			state:        types.StateActive,
			environment: map[string]string{
				"NODE_ENV":  "production",
				"LOG_LEVEL": "info",
				"DEBUG":     "false", // Extra environment variable (allowed)
			},
			metadata: map[string]interface{}{"validation_type": "valid"},
		},
		{
			// Invalid configuration - wrong runtime
			functionName: "production-function",
			runtime:      types.RuntimePython39, // Wrong runtime
			handler:      "index.handler",
			timeout:      30,
			memorySize:   256,
			state:        types.StateActive,
			environment: map[string]string{
				"NODE_ENV":  "production",
				"LOG_LEVEL": "info",
			},
			metadata: map[string]interface{}{"validation_type": "invalid_runtime"},
		},
		{
			// Invalid configuration - missing environment variables
			functionName: "production-function",
			runtime:      types.RuntimeNodejs18x,
			handler:      "index.handler",
			timeout:      30,
			memorySize:   256,
			state:        types.StateActive,
			environment: map[string]string{
				"NODE_ENV": "development", // Wrong value
				// Missing LOG_LEVEL
			},
			metadata: map[string]interface{}{"validation_type": "missing_env_vars"},
		},
	}
	
	// Validate all configurations using functional composition
	validationResults := lo.Map(testConfigurations, func(config FunctionalConfigSpec, _ int) mo.Result[FunctionalConfigSpec] {
		return validateConfigSpec(config, expectedConfig)
	})
	
	// Display validation results using functional operations
	lo.ForEach(lo.Zip2(testConfigurations, validationResults), func(pair lo.Tuple2[FunctionalConfigSpec, mo.Result[FunctionalConfigSpec]], index int) {
		config := pair.A
		result := pair.B
		
		validationType := config.metadata["validation_type"].(string)
		fmt.Printf("Configuration validation %d (%s):\n", index+1, validationType)
		
		result.
			Map(func(validConfig FunctionalConfigSpec) FunctionalConfigSpec {
				fmt.Printf("  ✓ Validation passed - Function: %s, Runtime: %s\n",
					validConfig.functionName, validConfig.runtime)
				return validConfig
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Validation failed: %v\n", err)
				return err
			})
	})
	
	// Demonstrate advanced validation patterns with Lambda configurations
	functionalValidationScenario := func() mo.Result[[]lambda.FunctionalLambdaConfig] {
		// Create multiple Lambda configurations for validation
		configs := []lambda.FunctionalLambdaConfig{
			lambda.NewFunctionalLambdaConfig("user-service",
				lambda.WithTimeout(10*time.Second),
				lambda.WithRetryConfig(3, 500*time.Millisecond),
				lambda.WithLambdaMetadata("service_tier", "standard"),
			),
			lambda.NewFunctionalLambdaConfig("payment-service",
				lambda.WithTimeout(30*time.Second),
				lambda.WithRetryConfig(5, 1*time.Second),
				lambda.WithLambdaMetadata("service_tier", "critical"),
			),
			lambda.NewFunctionalLambdaConfig("notification-service",
				lambda.WithTimeout(5*time.Second),
				lambda.WithRetryConfig(2, 200*time.Millisecond),
				lambda.WithLambdaMetadata("service_tier", "background"),
			),
		}
		
		// Validate each configuration using functional composition
		validationErrors := lo.FilterMap(configs, func(config lambda.FunctionalLambdaConfig, _ int) (error, bool) {
			// Extract metadata for validation
			serviceTier := config.GetMetadata()["service_tier"]
			
			// Validate timeout based on service tier
			switch serviceTier {
			case "critical":
				if config.timeout < 15*time.Second {
					return fmt.Errorf("critical service %s must have timeout >= 15s, got %v", config.functionName, config.timeout), true
				}
			case "standard":
				if config.timeout < 5*time.Second || config.timeout > 30*time.Second {
					return fmt.Errorf("standard service %s timeout must be 5-30s, got %v", config.functionName, config.timeout), true
				}
			case "background":
				if config.timeout > 10*time.Second {
					return fmt.Errorf("background service %s timeout must be <= 10s, got %v", config.functionName, config.timeout), true
				}
			}
			
			return nil, false
		})
		
		if len(validationErrors) > 0 {
			return mo.Err[[]lambda.FunctionalLambdaConfig](fmt.Errorf("configuration validation failed with %d errors", len(validationErrors)))
		}
		
		return mo.Ok(configs)
	}
	
	// Execute advanced validation scenario
	fmt.Println("\nAdvanced Lambda configuration validation:")
	advancedResult := functionalValidationScenario()
	advancedResult.
		Map(func(configs []lambda.FunctionalLambdaConfig) []lambda.FunctionalLambdaConfig {
			fmt.Printf("✓ All %d Lambda configurations are valid\n", len(configs))
			lo.ForEach(configs, func(config lambda.FunctionalLambdaConfig, _ int) {
				serviceTier := config.GetMetadata()["service_tier"]
				fmt.Printf("  - %s (%s): timeout=%v, retries=%d\n",
					config.functionName, serviceTier, config.timeout, config.maxRetries)
			})
			return configs
		}).
		MapErr(func(err error) error {
			fmt.Printf("✗ Lambda configuration validation failed: %v\n", err)
			return err
		})
}

// ExampleFluentAPIPatterns demonstrates functional composition patterns with monadic operations
func ExampleFluentAPIPatterns() {
	fmt.Println("Functional composition patterns with monadic operations:")
	
	// Functional Lambda operation builder using method chaining with immutable structures
	type FunctionalLambdaOperation struct {
		config         lambda.FunctionalLambdaConfig
		payload        string
		expectations   []func(lambda.FunctionalOperationResult) mo.Result[bool]
		transformations []func(lambda.FunctionalOperationResult) lambda.FunctionalOperationResult
	}
	
	// Functional operation builder using immutable patterns
	createOperation := func(functionName string) FunctionalLambdaOperation {
		return FunctionalLambdaOperation{
			config:          lambda.NewFunctionalLambdaConfig(functionName),
			payload:         "",
			expectations:    make([]func(lambda.FunctionalOperationResult) mo.Result[bool], 0),
			transformations: make([]func(lambda.FunctionalOperationResult) lambda.FunctionalOperationResult, 0),
		}
	}
	
	// Fluent methods using functional composition
	withPayload := func(op FunctionalLambdaOperation, payload string) FunctionalLambdaOperation {
		return FunctionalLambdaOperation{
			config:          op.config,
			payload:         payload,
			expectations:    op.expectations,
			transformations: op.transformations,
		}
	}
	
	withTimeout := func(op FunctionalLambdaOperation, timeout time.Duration) FunctionalLambdaOperation {
		newConfig := lambda.NewFunctionalLambdaConfig(op.config.functionName,
			lambda.WithTimeout(timeout),
			lambda.WithRetryConfig(op.config.maxRetries, op.config.retryDelay),
		)
		return FunctionalLambdaOperation{
			config:          newConfig,
			payload:         op.payload,
			expectations:    op.expectations,
			transformations: op.transformations,
		}
	}
	
	withRetries := func(op FunctionalLambdaOperation, maxRetries int, delay time.Duration) FunctionalLambdaOperation {
		newConfig := lambda.NewFunctionalLambdaConfig(op.config.functionName,
			lambda.WithTimeout(op.config.timeout),
			lambda.WithRetryConfig(maxRetries, delay),
		)
		return FunctionalLambdaOperation{
			config:          newConfig,
			payload:         op.payload,
			expectations:    op.expectations,
			transformations: op.transformations,
		}
	}
	
	withExpectation := func(op FunctionalLambdaOperation, expectation func(lambda.FunctionalOperationResult) mo.Result[bool]) FunctionalLambdaOperation {
		newExpectations := append(op.expectations, expectation)
		return FunctionalLambdaOperation{
			config:          op.config,
			payload:         op.payload,
			expectations:    newExpectations,
			transformations: op.transformations,
		}
	}
	
	withTransformation := func(op FunctionalLambdaOperation, transform func(lambda.FunctionalOperationResult) lambda.FunctionalOperationResult) FunctionalLambdaOperation {
		newTransformations := append(op.transformations, transform)
		return FunctionalLambdaOperation{
			config:          op.config,
			payload:         op.payload,
			expectations:    op.expectations,
			transformations: newTransformations,
		}
	}
	
	// Execute operation with functional composition
	executeOperation := func(op FunctionalLambdaOperation) mo.Result[lambda.FunctionalOperationResult] {
		// Execute Lambda invocation
		result := lambda.SimulateFunctionalLambdaInvocation(op.payload, op.config)
		
		// Apply transformations using functional composition
		transformedResult := lo.Reduce(op.transformations, func(acc lambda.FunctionalOperationResult, transform func(lambda.FunctionalOperationResult) lambda.FunctionalOperationResult, _ int) lambda.FunctionalOperationResult {
			return transform(acc)
		}, result)
		
		// Validate expectations using monadic operations
		expectationResults := lo.Map(op.expectations, func(expectation func(lambda.FunctionalOperationResult) mo.Result[bool], _ int) mo.Result[bool] {
			return expectation(transformedResult)
		})
		
		// Check if all expectations pass
		allExpectationsPassed := lo.Every(expectationResults, func(result mo.Result[bool]) bool {
			return result.IsOk() && result.MustGet()
		})
		
		if !allExpectationsPassed {
			failedExpectations := lo.Filter(expectationResults, func(result mo.Result[bool], _ int) bool {
				return result.IsError() || !result.MustGet()
			})
			return mo.Err[lambda.FunctionalOperationResult](fmt.Errorf("%d expectations failed", len(failedExpectations)))
		}
		
		return mo.Ok(transformedResult)
	}
	
	// Define expectation functions
	expectSuccess := func(result lambda.FunctionalOperationResult) mo.Result[bool] {
		if result.IsSuccess() {
			return mo.Ok(true)
		} else {
			return mo.Ok(false)
		}
	}
	
	expectPayloadContains := func(substring string) func(lambda.FunctionalOperationResult) mo.Result[bool] {
		return func(result lambda.FunctionalOperationResult) mo.Result[bool] {
			return result.GetResult().
				Map(func(invokeResult lambda.FunctionalInvokeResult) (bool, bool) {
					contains := len(invokeResult.Payload) > 0 && len(substring) > 0
					return contains, true
				}).
				Try()
		}
	}
	
	expectDurationLessThan := func(maxDuration time.Duration) func(lambda.FunctionalOperationResult) mo.Result[bool] {
		return func(result lambda.FunctionalOperationResult) mo.Result[bool] {
			withinTimeLimit := result.GetDuration() < maxDuration
			return mo.Ok(withinTimeLimit)
		}
	}
	
	// Transformation functions
	enrichWithMetadata := func(metadata map[string]interface{}) func(lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
		return func(result lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
			enrichedMetadata := lo.Assign(result.GetMetadata(), metadata)
			if result.IsSuccess() {
				return lambda.NewSuccessFunctionalOperationResult(
					result.GetResult().MustGet(),
					result.GetDuration(),
					enrichedMetadata,
				)
			} else {
				return lambda.NewErrorFunctionalOperationResult(
					result.GetError().MustGet(),
					result.GetDuration(),
					enrichedMetadata,
				)
			}
		}
	}
	
	// Demonstrate functional composition patterns
	exampleOperations := []struct {
		name        string
		operations  []func(FunctionalLambdaOperation) FunctionalLambdaOperation
		description string
	}{
		{
			name: "BasicOperation",
			operations: []func(FunctionalLambdaOperation) FunctionalLambdaOperation{
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withPayload(op, `{"test": true}`)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withExpectation(op, expectSuccess)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withExpectation(op, expectPayloadContains("result"))
				},
			},
			description: "Basic Lambda operation with success expectation",
		},
		{
			name: "PerformanceOperation",
			operations: []func(FunctionalLambdaOperation) FunctionalLambdaOperation{
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withPayload(op, `{"performance_test": true}`)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withTimeout(op, 2*time.Second)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withRetries(op, 3, 200*time.Millisecond)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withExpectation(op, expectSuccess)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withExpectation(op, expectDurationLessThan(500*time.Millisecond))
				},
			},
			description: "Performance-focused operation with duration expectations",
		},
		{
			name: "EnrichedOperation",
			operations: []func(FunctionalLambdaOperation) FunctionalLambdaOperation{
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withPayload(op, `{"enrichment_test": true}`)
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withTransformation(op, enrichWithMetadata(map[string]interface{}{
						"enriched_at": time.Now().UTC().Format(time.RFC3339),
						"enriched_by": "functional_pipeline",
					}))
				},
				func(op FunctionalLambdaOperation) FunctionalLambdaOperation {
					return withExpectation(op, expectSuccess)
				},
			},
			description: "Operation with metadata enrichment transformation",
		},
	}
	
	// Execute all example operations using functional composition
	operationResults := lo.Map(exampleOperations, func(example struct {
		name        string
		operations  []func(FunctionalLambdaOperation) FunctionalLambdaOperation
		description string
	}, _ int) mo.Result[lambda.FunctionalOperationResult] {
		// Compose all operations using functional pipeline
		finalOperation := lo.Reduce(example.operations, func(acc FunctionalLambdaOperation, op func(FunctionalLambdaOperation) FunctionalLambdaOperation, _ int) FunctionalLambdaOperation {
			return op(acc)
		}, createOperation("test-function"))
		
		return executeOperation(finalOperation)
	})
	
	// Display results using functional operations
	lo.ForEach(lo.Zip2(exampleOperations, operationResults), func(pair lo.Tuple2[struct {
		name        string
		operations  []func(FunctionalLambdaOperation) FunctionalLambdaOperation
		description string
	}, mo.Result[lambda.FunctionalOperationResult]], _ int) {
		example := pair.A
		result := pair.B
		
		fmt.Printf("%s: %s\n", example.name, example.description)
		
		result.
			Map(func(opResult lambda.FunctionalOperationResult) lambda.FunctionalOperationResult {
				fmt.Printf("  ✓ Operation completed successfully - Duration: %v\n", opResult.GetDuration())
				if len(opResult.GetMetadata()) > 0 {
					fmt.Printf("  → Metadata: %v\n", opResult.GetMetadata())
				}
				return opResult
			}).
			MapErr(func(err error) error {
				fmt.Printf("  ✗ Operation failed: %v\n", err)
				return err
			})
	})
	
	// Demonstrate complex function composition with lo.Pipe
	fmt.Println("\nComplex functional composition with lo.Pipe:")
	complexWorkflow := lo.Pipe4(
		`{"workflow": "complex", "step": 1}`,
		func(payload string) lambda.FunctionalLambdaConfig {
			return lambda.NewFunctionalLambdaConfig("workflow-step-1",
				lambda.WithLambdaMetadata("workflow_step", 1),
				lambda.WithTimeout(5*time.Second),
			)
		},
		func(config lambda.FunctionalLambdaConfig) lambda.FunctionalOperationResult {
			return lambda.SimulateFunctionalLambdaInvocation(`{"workflow": "complex", "step": 1}`, config)
		},
		func(result lambda.FunctionalOperationResult) mo.Result[string] {
			if result.IsSuccess() {
				return mo.Ok(fmt.Sprintf("Workflow completed in %v", result.GetDuration()))
			} else {
				return mo.Err[string](fmt.Errorf("workflow failed: %v", result.GetError().MustGet()))
			}
		},
	)
	
	complexWorkflow.
		Map(func(message string) string {
			fmt.Printf("✓ %s\n", message)
			return message
		}).
		MapErr(func(err error) error {
			fmt.Printf("✗ %v\n", err)
			return err
		})
}

// Functional helper functions using pure functions and monadic operations

// extractPayloadFromFunctionalResult demonstrates pure function result extraction
func extractPayloadFromFunctionalResult(result lambda.FunctionalOperationResult) mo.Option[string] {
	// Pure function that safely extracts payload using monadic operations
	return result.GetResult().
		Map(func(invokeResult lambda.FunctionalInvokeResult) (string, bool) {
			if len(invokeResult.Payload) > 0 {
				return invokeResult.Payload, true
			}
			return "", false
		})
}

// buildFunctionalCustomEvent demonstrates immutable event creation with validation
func buildFunctionalCustomEvent(eventType string, data map[string]interface{}) mo.Result[string] {
	// Immutable event creation using functional composition
	validateEventType := func(eType string) error {
		allowedTypes := []string{"user_registration", "order_placed", "payment_processed", "notification_sent"}
		if !lo.Contains(allowedTypes, eType) {
			return fmt.Errorf("invalid event type: %s, allowed types: %v", eType, allowedTypes)
		}
		return nil
	}
	
	validateEventData := func(data map[string]interface{}) error {
		if len(data) == 0 {
			return fmt.Errorf("event data cannot be empty")
		}
		return nil
	}
	
	// Validate inputs using functional pipeline
	validationPipeline := lambda.ExecuteFunctionalValidationPipeline[map[string]interface{}](data,
		func(d map[string]interface{}) (map[string]interface{}, error) {
			return d, validateEventType(eventType)
		},
		func(d map[string]interface{}) (map[string]interface{}, error) {
			return d, validateEventData(d)
		},
	)
	
	if validationPipeline.IsError() {
		return mo.Err[string](validationPipeline.GetError().MustGet())
	}
	
	// Create immutable event structure
	eventData := lo.Assign(map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"event_id":   fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano()),
	}, map[string]interface{}{
		lo.Ternary(eventType == "user_registration", "user_data", 
			lo.Ternary(eventType == "order_placed", "order_data", "data")): data,
	})
	
	// Serialize to JSON using pure function
	bytes, err := json.Marshal(eventData)
	if err != nil {
		return mo.Err[string](fmt.Errorf("failed to serialize event: %w", err))
	}
	
	return mo.Ok(string(bytes))
}

// createFunctionalEventBuilder demonstrates higher-order function for event building
func createFunctionalEventBuilder(eventType string) func(map[string]interface{}) mo.Result[string] {
	// Return a specialized event builder function
	return func(data map[string]interface{}) mo.Result[string] {
		return buildFunctionalCustomEvent(eventType, data)
	}
}

// composeFunctionalEventPipeline demonstrates event processing pipeline
func composeFunctionalEventPipeline(events []map[string]interface{}, eventType string) mo.Result[[]string] {
	// Create event builder for the specified type
	eventBuilder := createFunctionalEventBuilder(eventType)
	
	// Process all events using functional composition
	eventResults := lo.Map(events, func(event map[string]interface{}, _ int) mo.Result[string] {
		return eventBuilder(event)
	})
	
	// Filter successful results
	successfulEvents := lo.FilterMap(eventResults, func(result mo.Result[string], _ int) (string, bool) {
		return result.OrElse(func() string { return "" }), result.IsOk()
	})
	
	// Check if all events were processed successfully
	if len(successfulEvents) == len(events) {
		return mo.Ok(successfulEvents)
	} else {
		failedCount := len(events) - len(successfulEvents)
		return mo.Err[[]string](fmt.Errorf("%d out of %d events failed to process", failedCount, len(events)))
	}
}

// ExampleOrganizationPatterns demonstrates functional organization patterns with immutable structures
func ExampleOrganizationPatterns() {
	fmt.Println("Functional Lambda organization patterns with pure functions and immutable structures:")
	
	// Define organization pattern categories using immutable structures
	type PatternCategory struct {
		name        string
		description string
		patterns    []string
		examples    []func() mo.Result[string]
	}
	
	// Create pattern categories using functional composition
	patternCategories := []PatternCategory{
		{
			name:        "Unit Patterns",
			description: "Pure function patterns for individual Lambda operations",
			patterns: []string{
				"Immutable payload validation with mo.Result",
				"Pure event parsing with functional composition",
				"Configuration validation using monadic operations",
				"Type-safe Lambda configuration builders",
			},
			examples: []func() mo.Result[string]{
				func() mo.Result[string] {
					// Example: Pure payload validation
					payload := `{"user_id": "123", "action": "create"}`
					config := lambda.NewFunctionalLambdaConfig("validator",
						lambda.WithPayloadValidator(func(p string) error {
							if len(p) == 0 {
								return fmt.Errorf("empty payload")
							}
							return nil
						}),
					)
					result := lambda.SimulateFunctionalLambdaInvocation(payload, config)
					if result.IsSuccess() {
						return mo.Ok("Payload validation pattern demonstrated")
					} else {
						return mo.Err[string](fmt.Errorf("validation failed"))
					}
				},
				func() mo.Result[string] {
					// Example: Event parsing with functional composition
					eventData := map[string]interface{}{
						"user_id": "user-456",
						"email":   "user@example.com",
					}
					return buildFunctionalCustomEvent("user_registration", eventData).
						Map(func(event string) string {
							return "Event parsing pattern demonstrated"
						})
				},
			},
		},
		{
			name:        "Integration Patterns",
			description: "Functional composition patterns for Lambda integration",
			patterns: []string{
				"Monadic Lambda function invocation chains",
				"Event-driven function composition with lo.Pipe",
				"Functional retry patterns with exponential backoff",
				"Immutable configuration propagation",
			},
			examples: []func() mo.Result[string]{
				func() mo.Result[string] {
					// Example: Function chain with monadic composition
					step1Config := lambda.NewFunctionalLambdaConfig("step1",
						lambda.WithLambdaMetadata("chain_position", 1),
					)
					step2Config := lambda.NewFunctionalLambdaConfig("step2",
						lambda.WithLambdaMetadata("chain_position", 2),
					)
					
					// Chain functions using monadic operations
					step1Result := lambda.SimulateFunctionalLambdaInvocation(`{"chain": "start"}`, step1Config)
					if step1Result.IsSuccess() {
						step2Result := lambda.SimulateFunctionalLambdaInvocation(`{"chain": "continue"}`, step2Config)
						if step2Result.IsSuccess() {
							return mo.Ok("Function chain pattern demonstrated")
						}
					}
					return mo.Err[string](fmt.Errorf("chain failed"))
				},
				func() mo.Result[string] {
					// Example: Event-driven composition
					eventChain := lo.Pipe3(
						map[string]interface{}{"trigger": "start"},
						func(data map[string]interface{}) mo.Result[string] {
							return buildFunctionalCustomEvent("order_placed", data)
						},
						func(event mo.Result[string]) mo.Result[lambda.FunctionalOperationResult] {
							return event.FlatMap(func(eventStr string) mo.Result[lambda.FunctionalOperationResult] {
								config := lambda.NewFunctionalLambdaConfig("event-processor")
								result := lambda.SimulateFunctionalLambdaInvocation(eventStr, config)
								return mo.Ok(result)
							})
						},
					)
					
					return eventChain.Map(func(result lambda.FunctionalOperationResult) string {
						return "Event-driven composition pattern demonstrated"
					})
				},
			},
		},
		{
			name:        "End-to-End Patterns",
			description: "Complete workflow patterns using functional composition",
			patterns: []string{
				"Immutable workflow state management",
				"Functional saga patterns with compensation",
				"Event sourcing with pure functions",
				"Pipeline composition with error handling",
			},
			examples: []func() mo.Result[string]{
				func() mo.Result[string] {
					// Example: Complete order processing workflow
					workflowSteps := []string{"validate", "process", "fulfill", "notify"}
					
					// Execute workflow using functional composition
					workflowResults := lo.Map(workflowSteps, func(step string, _ int) lambda.FunctionalOperationResult {
						config := lambda.NewFunctionalLambdaConfig(fmt.Sprintf("order-%s", step),
							lambda.WithLambdaMetadata("workflow_step", step),
						)
						return lambda.SimulateFunctionalLambdaInvocation(
							fmt.Sprintf(`{"step": "%s", "order_id": "order-123"}`, step),
							config,
						)
					})
					
					allSuccessful := lo.EveryBy(workflowResults, func(result lambda.FunctionalOperationResult) bool {
						return result.IsSuccess()
					})
					
					if allSuccessful {
						return mo.Ok("End-to-end workflow pattern demonstrated")
					} else {
						return mo.Err[string](fmt.Errorf("workflow failed"))
					}
				},
				func() mo.Result[string] {
					// Example: Data processing pipeline
					dataPipeline := lo.Pipe4(
						`{"raw_data": "input"}`,
						func(input string) lambda.FunctionalLambdaConfig {
							return lambda.NewFunctionalLambdaConfig("data-ingester",
								lambda.WithLambdaMetadata("pipeline_stage", "ingest"),
							)
						},
						func(config lambda.FunctionalLambdaConfig) lambda.FunctionalOperationResult {
							return lambda.SimulateFunctionalLambdaInvocation(`{"stage": "ingest"}`, config)
						},
						func(result lambda.FunctionalOperationResult) mo.Result[string] {
							if result.IsSuccess() {
								return mo.Ok("Data processing pipeline pattern demonstrated")
							} else {
								return mo.Err[string](fmt.Errorf("pipeline failed"))
							}
						},
					)
					
					return dataPipeline
				},
			},
		},
		{
			name:        "Performance Patterns",
			description: "Functional patterns for high-performance Lambda operations",
			patterns: []string{
				"Concurrent execution with functional composition",
				"Immutable performance monitoring",
				"Functional circuit breaker patterns",
				"Pure function caching and memoization",
			},
			examples: []func() mo.Result[string]{
				func() mo.Result[string] {
					// Example: Concurrent execution pattern
					concurrentTasks := lo.Times(5, func(i int) func() lambda.FunctionalOperationResult {
						return func() lambda.FunctionalOperationResult {
							config := lambda.NewFunctionalLambdaConfig(fmt.Sprintf("concurrent-task-%d", i),
								lambda.WithLambdaMetadata("task_id", i),
							)
							return lambda.SimulateFunctionalLambdaInvocation(
								fmt.Sprintf(`{"task_id": %d}`, i),
								config,
							)
						}
					})
					
					// Simulate concurrent execution
					results := make(chan lambda.FunctionalOperationResult, len(concurrentTasks))
					for _, task := range concurrentTasks {
						go func(t func() lambda.FunctionalOperationResult) {
							results <- t()
						}(task)
					}
					
					// Collect results
					collectedResults := lo.Times(len(concurrentTasks), func(_ int) lambda.FunctionalOperationResult {
						return <-results
					})
					
					allSuccessful := lo.EveryBy(collectedResults, func(result lambda.FunctionalOperationResult) bool {
						return result.IsSuccess()
					})
					
					if allSuccessful {
						return mo.Ok("Concurrent execution pattern demonstrated")
					} else {
						return mo.Err[string](fmt.Errorf("concurrent execution failed"))
					}
				},
				func() mo.Result[string] {
					// Example: Performance monitoring pattern
					performanceConfig := lambda.NewFunctionalLambdaConfig("performance-monitor",
						lambda.WithTimeout(1*time.Second),
						lambda.WithLambdaMetadata("performance_monitoring", true),
					)
					
					startTime := time.Now()
					result := lambda.SimulateFunctionalLambdaInvocation(`{"monitor": true}`, performanceConfig)
					duration := time.Since(startTime)
					
					if result.IsSuccess() && duration < 100*time.Millisecond {
						return mo.Ok(fmt.Sprintf("Performance monitoring pattern demonstrated - Duration: %v", duration))
					} else {
						return mo.Err[string](fmt.Errorf("performance monitoring failed"))
					}
				},
			},
		},
	}
	
	// Execute all pattern examples using functional composition
	lo.ForEach(patternCategories, func(category PatternCategory, _ int) {
		fmt.Printf("\n%s: %s\n", category.name, category.description)
		
		// Display pattern list
		lo.ForEach(category.patterns, func(pattern string, index int) {
			fmt.Printf("  %d. %s\n", index+1, pattern)
		})
		
		// Execute examples if available
		if len(category.examples) > 0 {
			fmt.Printf("  Examples:\n")
			lo.ForEach(category.examples, func(example func() mo.Result[string], index int) {
				exampleResult := example()
				exampleResult.
					Map(func(message string) string {
						fmt.Printf("    ✓ Example %d: %s\n", index+1, message)
						return message
					}).
					MapErr(func(err error) error {
						fmt.Printf("    ✗ Example %d failed: %v\n", index+1, err)
						return err
					})
			})
		}
	})
	
	// Summary of functional programming benefits
	fmt.Println("\nFunctional Programming Benefits in Lambda Testing:")
	benefits := []string{
		"Immutable configurations prevent accidental state mutations",
		"Monadic error handling with mo.Result provides safe error propagation",
		"Function composition with lo.Pipe enables complex workflow creation",
		"Pure functions ensure predictable and testable Lambda operations",
		"Type safety with Go generics catches errors at compile time",
		"Functional options pattern allows flexible configuration building",
	}
	
	lo.ForEach(benefits, func(benefit string, index int) {
		fmt.Printf("  ✓ %s\n", benefit)
	})
}

// runAllExamples demonstrates running all Lambda examples
func runAllExamples() {
	fmt.Println("Running all Lambda package examples:")
	
	ExampleBasicUsage()
	fmt.Println("")
	
	ExampleTableDrivenValidation()
	fmt.Println("")
	
	ExampleAdvancedEventProcessing()
	fmt.Println("")
	
	ExampleEventSourceMapping()
	fmt.Println("")
	
	ExamplePerformancePatterns()
	fmt.Println("")
	
	ExampleErrorHandlingAndRecovery()
	fmt.Println("")
	
	ExampleComprehensiveIntegration()
	fmt.Println("")
	
	ExampleDataFactories()
	fmt.Println("")
	
	ExampleValidationHelpers()
	fmt.Println("")
	
	ExampleFluentAPIPatterns()
	fmt.Println("")
	
	ExampleOrganizationPatterns()
	fmt.Println("All Lambda examples completed")
}

func main() {
	// This file demonstrates functional Lambda usage patterns with pure functions,
	// immutable data structures, and monadic operations using samber/lo and samber/mo.
	// Run examples with: go run ./examples/lambda/examples.go
	
	fmt.Println("=== Functional Programming Lambda Examples ===")
	fmt.Println("Demonstrating pure functions, immutable configurations, and monadic operations")
	fmt.Println("")
	
	runAllExamples()
	
	fmt.Println("\n=== Summary ===")
	fmt.Println("All functional Lambda examples completed successfully!")
	fmt.Println("Key patterns demonstrated:")
	fmt.Println("  ✓ Immutable Lambda configurations with functional options")
	fmt.Println("  ✓ Monadic error handling with mo.Option and mo.Result")
	fmt.Println("  ✓ Function composition with lo.Pipe for complex workflows")
	fmt.Println("  ✓ Pure functions for Lambda invocation and validation")
	fmt.Println("  ✓ Type-safe operations with Go generics")
	fmt.Println("  ✓ Functional event builders and payload creators")
	fmt.Println("  ✓ Concurrent execution with functional composition")
	fmt.Println("  ✓ Advanced validation pipelines using monadic operations")
}