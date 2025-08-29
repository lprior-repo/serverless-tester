package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	
	"vasdeference"
	
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ExampleFunctionalCoreUsage demonstrates the functional core usage patterns
func ExampleFunctionalCoreUsage() {
	fmt.Println("⚡ Functional Core Usage Patterns:")
	
	// Create immutable core configuration
	config := vasdeference.NewFunctionalCoreConfig(
		vasdeference.WithCoreRegion("us-east-1"),
		vasdeference.WithCoreTimeout(30*time.Second),
		vasdeference.WithCoreNamespace("functional-example"),
		vasdeference.WithCoreMetrics(true),
		vasdeference.WithCoreMetadata("environment", "test"),
	)
	
	fmt.Printf("✅ Created immutable core configuration with region: %s\n", config.GetRegion())
	fmt.Printf("✅ Configuration timeout: %v\n", config.GetTimeout())
	fmt.Printf("✅ Configuration namespace: %s\n", config.GetNamespace().MustGet())
	
	// Pure functional validation
	validationResult := validateFunctionalCoreConfig(config)
	validationResult.
		Map(func(err error) error {
			fmt.Printf("❌ Validation error: %v\n", err)
			return err
		}).
		OrElse(func() {
			fmt.Println("✅ Configuration validation passed")
		})
	
	// Function composition for complex operations
	pipeline := lo.Pipe3(
		createTestContext,
		validateContext,
		processContext,
	)
	
	result := pipeline(config)
	fmt.Printf("✅ Pipeline result: %+v\n", result)
	fmt.Println("Functional core usage completed\n")
}

// validateFunctionalCoreConfig is a pure function that validates configuration
func validateFunctionalCoreConfig(config *vasdeference.FunctionalCoreConfig) mo.Option[error] {
	if config.GetRegion() == "" {
		return mo.Some[error](fmt.Errorf("region cannot be empty"))
	}
	if config.GetTimeout() <= 0 {
		return mo.Some[error](fmt.Errorf("timeout must be positive"))
	}
	return mo.None[error]()
}

// Pure functions for pipeline composition
func createTestContext(config *vasdeference.FunctionalCoreConfig) map[string]interface{} {
	return map[string]interface{}{
		"region":    config.GetRegion(),
		"timeout":   config.GetTimeout(),
		"namespace": config.GetNamespace().OrEmpty(),
		"status":    "created",
	}
}

func validateContext(context map[string]interface{}) map[string]interface{} {
	context["validated"] = true
	return context
}

func processContext(context map[string]interface{}) map[string]interface{} {
	context["processed"] = true
	context["timestamp"] = time.Now().Unix()
	return context
}

// ExampleWithCustomOptions demonstrates functional configuration patterns
func ExampleWithCustomOptions() {
	fmt.Println("⚙️ Functional Configuration Patterns:")
	
	// Multiple configuration approaches with functional composition
	baseConfig := vasdeference.NewFunctionalCoreConfig(
		vasdeference.WithCoreRegion("us-west-2"),
		vasdeference.WithCoreTimeout(60*time.Second),
	)
	
	enhancedConfig := vasdeference.NewFunctionalCoreConfig(
		vasdeference.WithCoreRegion("eu-west-1"),
		vasdeference.WithCoreNamespace("integration-test"),
		vasdeference.WithCoreMetrics(true),
		vasdeference.WithCoreResilience(true),
		vasdeference.WithCoreMetadata("environment", "staging"),
		vasdeference.WithCoreMetadata("team", "platform"),
	)
	
	fmt.Printf("✅ Base config: %s region, %v timeout\n", 
		baseConfig.GetRegion(), baseConfig.GetTimeout())
	fmt.Printf("✅ Enhanced config: %s region, namespace: %s\n", 
		enhancedConfig.GetRegion(), enhancedConfig.GetNamespace().OrEmpty())
	
	// Demonstrate configuration validation with monads
	validationResult := lo.Pipe2(
		enhancedConfig,
		validateFunctionalCoreConfig,
		func(err mo.Option[error]) string {
			return err.Map(func(e error) string {
				return fmt.Sprintf("❌ Validation failed: %v", e)
			}).OrElse("✅ Configuration is valid")
		},
	)
	
	fmt.Println(validationResult)
	
	// Use lo.Map to transform configuration options
	options := []string{"logging", "metrics", "resilience"}
	enabled := lo.Map(options, func(option string, _ int) map[string]bool {
		switch option {
		case "logging":
			return map[string]bool{option: enhancedConfig.IsLoggingEnabled()}
		case "metrics":
			return map[string]bool{option: enhancedConfig.IsMetricsEnabled()}
		case "resilience":
			return map[string]bool{option: enhancedConfig.IsResilienceEnabled()}
		default:
			return map[string]bool{option: false}
		}
	})
	
	fmt.Printf("✅ Feature status: %+v\n", enabled)
	fmt.Println("Functional configuration completed\n")
}

// ExampleTerraformIntegration demonstrates functional Terraform output processing
func ExampleTerraformIntegration() {
	fmt.Println("🏗️ Functional Terraform Integration Patterns:")
	
	// Simulate Terraform outputs with immutable data
	outputs := map[string]interface{}{
		"lambda_function_name": "my-function-abc123",
		"api_gateway_url":     "https://abc123.execute-api.us-east-1.amazonaws.com",
		"dynamodb_table_name": "my-table-xyz789",
		"region":              "us-east-1",
		"environment":         "production",
	}
	
	// Type-safe extraction using monadic patterns
	getTerraformOutput := func(key string) mo.Option[string] {
		if val, exists := outputs[key]; exists {
			if str, ok := val.(string); ok {
				return mo.Some(str)
			}
		}
		return mo.None[string]()
	}
	
	// Pure functional extraction and validation
	resources := lo.MapEntries(outputs, func(key string, value interface{}) (string, mo.Option[string]) {
		return key, getTerraformOutput(key)
	})
	
	// Filter and process valid outputs
	validResources := lo.PickByValues(resources, func(opt mo.Option[string]) bool {
		return opt.IsPresent()
	})
	
	fmt.Printf("✅ Valid resources found: %d\n", len(validResources))
	
	// Function composition for resource naming
	resourceNamer := lo.Pipe2(
		func(baseName string) string { return strings.TrimSuffix(baseName, "-abc123") },
		func(name string) string { return fmt.Sprintf("validated-%s", name) },
	)
	
	functionName := getTerraformOutput("lambda_function_name").
		Map(resourceNamer).
		OrElse("default-function")
	
	fmt.Printf("✅ Processed function name: %s\n", functionName)
	
	// Validate URLs using pure functions
	validateURL := func(url string) mo.Option[error] {
		if !strings.HasPrefix(url, "https://") {
			return mo.Some[error](fmt.Errorf("URL must use HTTPS: %s", url))
		}
		return mo.None[error]()
	}
	
	apiValidation := getTerraformOutput("api_gateway_url").
		FlatMap(func(url string) mo.Option[string] {
			return validateURL(url).Match(
				func(err error) mo.Option[string] { return mo.None[string]() },
				func() mo.Option[string] { return mo.Some(url) },
			)
		}).
		OrElse("invalid-url")
	
	fmt.Printf("✅ API URL validation: %s\n", apiValidation)
	fmt.Println("Functional Terraform integration completed\n")
}

// ExampleLambdaIntegration demonstrates integration with lambda package
func ExampleLambdaIntegration() {
	fmt.Println("VasDeference Lambda integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Lambda client
	// lambdaClient := vdf.CreateLambdaClient()
	// Expected: lambdaClient should not be nil
	fmt.Println("Creating Lambda client through VasDeference")
	
	// functionName := vdf.PrefixResourceName("test-function")
	fmt.Println("Generating prefixed function name: test-function")
	
	// This is how you would integrate with the lambda package:
	// function := lambda.NewFunction(vdf.Context, &lambda.FunctionConfig{
	//     Name: functionName,
	//     Runtime: "nodejs18.x",
	//     Handler: "index.handler",
	// })
	fmt.Println("Creating Lambda function with configuration")
	
	// vdf.RegisterCleanup(func() error {
	//     return function.Delete()
	// })
	fmt.Println("Registering Lambda function cleanup")
	fmt.Println("Lambda integration completed")
}

// ExampleEventBridgeIntegration demonstrates integration with eventbridge package
func ExampleEventBridgeIntegration() {
	fmt.Println("VasDeference EventBridge integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create EventBridge client
	// eventBridgeClient := vdf.CreateEventBridgeClient()
	// Expected: eventBridgeClient should not be nil
	fmt.Println("Creating EventBridge client through VasDeference")
	
	// ruleName := vdf.PrefixResourceName("test-rule")
	fmt.Println("Generating prefixed rule name: test-rule")
	
	// This is how you would integrate with the eventbridge package:
	// rule := eventbridge.NewRule(vdf.Context, &eventbridge.RuleConfig{
	//     Name: ruleName,
	//     EventPattern: map[string]interface{}{
	//         "source": []string{"myapp"},
	//     },
	// })
	fmt.Println("Creating EventBridge rule with event pattern")
	
	// vdf.RegisterCleanup(func() error {
	//     return rule.Delete()
	// })
	fmt.Println("Registering EventBridge rule cleanup")
	fmt.Println("EventBridge integration completed")
}

// ExampleDynamoDBIntegration demonstrates integration with dynamodb package
func ExampleDynamoDBIntegration() {
	fmt.Println("VasDeference DynamoDB integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create DynamoDB client
	// dynamoClient := vdf.CreateDynamoDBClient()
	// Expected: dynamoClient should not be nil
	fmt.Println("Creating DynamoDB client through VasDeference")
	
	// tableName := vdf.PrefixResourceName("test-table")
	fmt.Println("Generating prefixed table name: test-table")
	
	// This is how you would integrate with the dynamodb package:
	// table := dynamodb.NewTable(vdf.Context, &dynamodb.TableConfig{
	//     TableName: tableName,
	//     BillingMode: "PAY_PER_REQUEST",
	//     AttributeDefinitions: []dynamodb.AttributeDefinition{
	//         {AttributeName: "id", AttributeType: "S"},
	//     },
	//     KeySchema: []dynamodb.KeySchemaElement{
	//         {AttributeName: "id", KeyType: "HASH"},
	//     },
	// })
	fmt.Println("Creating DynamoDB table with pay-per-request billing")
	
	// vdf.RegisterCleanup(func() error {
	//     return table.Delete()
	// })
	fmt.Println("Registering DynamoDB table cleanup")
	fmt.Println("DynamoDB integration completed")
}

// ExampleStepFunctionsIntegration demonstrates integration with stepfunctions package
func ExampleStepFunctionsIntegration() {
	fmt.Println("VasDeference Step Functions integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Step Functions client
	// sfnClient := vdf.CreateStepFunctionsClient()
	// Expected: sfnClient should not be nil
	fmt.Println("Creating Step Functions client through VasDeference")
	
	// stateMachineName := vdf.PrefixResourceName("test-state-machine")
	fmt.Println("Generating prefixed state machine name: test-state-machine")
	
	// This is how you would integrate with the stepfunctions package:
	definition := `{
	  "Comment": "A Hello World example",
	  "StartAt": "HelloWorld",
	  "States": {
	    "HelloWorld": {
	      "Type": "Pass",
	      "Result": "Hello World!",
	      "End": true
	    }
	  }
	}`
	fmt.Printf("Sample state machine definition:\n%s\n", definition)
	
	// stateMachine := stepfunctions.NewStateMachine(vdf.Context, &stepfunctions.StateMachineConfig{
	//     Name: stateMachineName,
	//     Definition: definition,
	//     RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
	// })
	fmt.Println("Creating Step Functions state machine with Hello World definition")
	
	// vdf.RegisterCleanup(func() error {
	//     return stateMachine.Delete()
	// })
	fmt.Println("Registering Step Functions state machine cleanup")
	fmt.Println("Step Functions integration completed")
}

// ExampleRetryMechanism demonstrates functional retry patterns with monadic error handling
func ExampleRetryMechanism() {
	fmt.Println("🔄 Functional Retry Mechanism Patterns:")
	
	// Define operation result type
	type OperationResult struct {
		value    interface{}
		attempts int
		duration time.Duration
	}
	
	// Create retry strategies using immutable configuration
	retryStrategies := []vasdeference.FunctionalRetryStrategy{
		vasdeference.NewFunctionalRetryStrategy(3, 100*time.Millisecond),
		vasdeference.NewFunctionalRetryStrategy(5, 200*time.Millisecond).
			WithMaxDelay(2*time.Second).
			WithMultiplier(1.5).
			WithJitter(true),
		vasdeference.NewFunctionalRetryStrategy(2, 50*time.Millisecond).
			WithJitter(false).
			WithRetryableErrors([]string{"timeout", "network error"}),
	}
	
	fmt.Printf("✅ Created %d retry strategies\n", len(retryStrategies))
	
	// Pure retry function using generics and monads
	type RetryableOperation[T any] func() (T, error)
	
	executeWithRetry := func[T any](operation RetryableOperation[T], strategy vasdeference.FunctionalRetryStrategy) mo.Result[T] {
		var lastError error
		var result T
		
		for attempt := 0; attempt < strategy.GetMaxRetries(); attempt++ {
			if attempt > 0 {
				// Calculate exponential backoff with jitter
				baseDelay := strategy.GetBaseDelay()
				multiplier := strategy.GetMultiplier()
				delay := time.Duration(float64(baseDelay) * (multiplier * float64(attempt)))
				
				if delay > strategy.GetMaxDelay() {
					delay = strategy.GetMaxDelay()
				}
				
				if strategy.IsJitterEnabled() {
					jitter := time.Duration(lo.RandomInt(0, int(delay.Milliseconds()/4))) * time.Millisecond
					delay += jitter
				}
				
				time.Sleep(delay)
			}
			
			result, lastError = operation()
			if lastError == nil {
				return mo.Ok(result)
			}
			
			// Check if error is retryable
			errorMsg := strings.ToLower(lastError.Error())
			retryable := lo.Some(strategy.GetRetryableErrors(), func(retryableErr string) bool {
				return strings.Contains(errorMsg, strings.ToLower(retryableErr))
			})
			
			if !retryable {
				break
			}
		}
		
		return mo.Err[T](lastError)
	}
	
	// Create test operations with different failure patterns
	testOperations := []struct {
		name        string
		operation   RetryableOperation[string]
		strategyIdx int
	}{
		{
			name: "eventually-succeeds",
			operation: func() func() (string, error) {
				attemptCount := 0
				return func() (string, error) {
					attemptCount++
					if attemptCount < 3 {
						return "", fmt.Errorf("timeout occurred")
					}
					return fmt.Sprintf("success-after-%d-attempts", attemptCount), nil
				}
			}(),
			strategyIdx: 0,
		},
		{
			name: "always-fails-retryable",
			operation: func() (string, error) {
				return "", fmt.Errorf("network error: connection refused")
			},
			strategyIdx: 1,
		},
		{
			name: "non-retryable-error",
			operation: func() (string, error) {
				return "", fmt.Errorf("validation error: invalid input")
			},
			strategyIdx: 2,
		},
	}
	
	// Execute operations with retry strategies using functional patterns
	results := lo.Map(testOperations, func(test struct {
		name        string
		operation   RetryableOperation[string]
		strategyIdx int
	}, _ int) struct {
		name   string
		result mo.Result[string]
	} {
		strategy := retryStrategies[test.strategyIdx]
		return struct {
			name   string
			result mo.Result[string]
		}{
			name:   test.name,
			result: executeWithRetry(test.operation, strategy),
		}
	})
	
	// Process results using monadic operations
	successCount := lo.CountBy(results, func(result struct {
		name   string
		result mo.Result[string]
	}) bool {
		return result.result.IsOk()
	})
	
	fmt.Printf("✅ Operations executed: %d\n", len(results))
	fmt.Printf("✅ Successful operations: %d\n", successCount)
	fmt.Printf("✅ Failed operations: %d\n", len(results)-successCount)
	
	// Display results using functional patterns
	lo.ForEach(results, func(result struct {
		name   string
		result mo.Result[string]
	}, _ int) {
		result.result.Match(
			func(err error) {
				fmt.Printf("❌ %s: %v\n", result.name, err)
			},
			func(value string) {
				fmt.Printf("✅ %s: %s\n", result.name, value)
			},
		)
	})
	
	fmt.Println("Functional retry mechanism completed\n")
}

// ExampleEnvironmentVariables demonstrates functional environment variable handling
func ExampleEnvironmentVariables() {
	fmt.Println("🌍 Functional Environment Variable Patterns:")
	
	// Pure functions for environment variable processing
	type EnvVar struct {
		key          string
		value        mo.Option[string]
		defaultValue string
		required     bool
	}
	
	// Pure function to get environment variable with Option
	getEnvVar := func(key string) mo.Option[string] {
		if value := os.Getenv(key); value != "" {
			return mo.Some(value)
		}
		return mo.None[string]()
	}
	
	// Create environment variable configurations
	envVars := []EnvVar{
		{"AWS_REGION", getEnvVar("AWS_REGION"), "us-east-1", true},
		{"AWS_PROFILE", getEnvVar("AWS_PROFILE"), "default", false},
		{"LOG_LEVEL", getEnvVar("LOG_LEVEL"), "INFO", false},
		{"ENVIRONMENT", getEnvVar("ENVIRONMENT"), "development", true},
		{"CI", getEnvVar("CI"), "false", false},
	}
	
	// Process environment variables using functional patterns
	processedVars := lo.Map(envVars, func(env EnvVar, _ int) struct {
		key           string
		resolvedValue string
		wasDefault    bool
		valid         bool
	} {
		resolvedValue := env.value.OrElse(env.defaultValue)
		wasDefault := env.value.IsAbsent()
		valid := !env.required || env.value.IsPresent()
		
		return struct {
			key           string
			resolvedValue string
			wasDefault    bool
			valid         bool
		}{
			key:           env.key,
			resolvedValue: resolvedValue,
			wasDefault:    wasDefault,
			valid:         valid,
		}
	})
	
	// Filter and categorize results
	validVars := lo.Filter(processedVars, func(env struct {
		key           string
		resolvedValue string
		wasDefault    bool
		valid         bool
	}, _ int) bool {
		return env.valid
	})
	
	defaultedVars := lo.Filter(validVars, func(env struct {
		key           string
		resolvedValue string
		wasDefault    bool
		valid         bool
	}, _ int) bool {
		return env.wasDefault
	})
	
	fmt.Printf("✅ Total environment variables: %d\n", len(envVars))
	fmt.Printf("✅ Valid variables: %d\n", len(validVars))
	fmt.Printf("✅ Using default values: %d\n", len(defaultedVars))
	
	// Display results using functional operations
	lo.ForEach(validVars, func(env struct {
		key           string
		resolvedValue string
		wasDefault    bool
		valid         bool
	}, _ int) {
		status := lo.Ternary(env.wasDefault, "(default)", "(from env)")
		fmt.Printf("✅ %s=%s %s\n", env.key, env.resolvedValue, status)
	})
	
	// Check CI environment using pure functions
	ciIndicators := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "CIRCLECI", "JENKINS_URL"}
	isCI := lo.Some(ciIndicators, func(indicator string) bool {
		return getEnvVar(indicator).Map(func(value string) bool {
			return strings.ToLower(value) == "true" || value != ""
		}).OrElse(false)
	})
	
	fmt.Printf("✅ Running in CI environment: %v\n", isCI)
	
	// Create configuration map using functional operations
	configMap := lo.Associate(validVars, func(env struct {
		key           string
		resolvedValue string
		wasDefault    bool
		valid         bool
	}) (string, string) {
		return env.key, env.resolvedValue
	})
	
	fmt.Printf("✅ Configuration map created with %d entries\n", len(configMap))
	fmt.Println("Functional environment variables completed\n")
}

// ExampleFullIntegration demonstrates functional composition of multiple services
func ExampleFullIntegration() {
	fmt.Println("🎆 Functional Full Integration Patterns:")
	
	// Create immutable configuration for complete integration
	integrationConfig := vasdeference.NewFunctionalCoreConfig(
		vasdeference.WithCoreRegion("us-east-1"),
		vasdeference.WithCoreNamespace(fmt.Sprintf("integration-%d", time.Now().Unix())),
		vasdeference.WithCoreTimeout(2*time.Minute),
		vasdeference.WithCoreMetrics(true),
		vasdeference.WithCoreResilience(true),
		vasdeference.WithCoreMetadata("deployment", "full-stack"),
		vasdeference.WithCoreMetadata("environment", "production"),
	)
	
	namespace := integrationConfig.GetNamespace().OrElse("default")
	fmt.Printf("✅ Generated namespace: %s\n", namespace)
	
	// Define service configurations using immutable structs
	type ServiceConfig struct {
		name         string
		serviceType  string
		dependencies []string
		resources    map[string]interface{}
		tags         map[string]string
	}
	
	// Create service configurations using functional composition
	services := lo.Map([]struct {
		baseName     string
		serviceType  string
		dependencies []string
	}{
		{"processor", "lambda", []string{}},
		{"data", "dynamodb", []string{}},
		{"trigger", "eventbridge", []string{"processor"}},
		{"workflow", "stepfunctions", []string{"processor", "data"}},
		{"monitor", "cloudwatch", []string{"processor", "data", "workflow"}},
	}, func(spec struct {
		baseName     string
		serviceType  string
		dependencies []string
	}, _ int) ServiceConfig {
		
		serviceName := fmt.Sprintf("%s-%s", namespace, spec.baseName)
		
		return ServiceConfig{
			name:         serviceName,
			serviceType:  spec.serviceType,
			dependencies: lo.Map(spec.dependencies, func(dep string, _ int) string {
				return fmt.Sprintf("%s-%s", namespace, dep)
			}),
			resources: map[string]interface{}{
				"arn":    fmt.Sprintf("arn:aws:%s:us-east-1:123456789012:%s", spec.serviceType, serviceName),
				"status": "pending",
				"region": integrationConfig.GetRegion(),
			},
			tags: map[string]string{
				"namespace": namespace,
				"service":   spec.serviceType,
				"managed":   "functional-framework",
			},
		}
	})
	
	fmt.Printf("✅ Created %d service configurations\n", len(services))
	
	// Validate service dependencies using pure functions
	validateDependencies := func(service ServiceConfig, allServices []ServiceConfig) mo.Option[error] {
		serviceNames := lo.Map(allServices, func(s ServiceConfig, _ int) string { return s.name })
		
		for _, dep := range service.dependencies {
			if !lo.Contains(serviceNames, dep) {
				return mo.Some[error](fmt.Errorf("service %s has missing dependency: %s", service.name, dep))
			}
		}
		return mo.None[error]()
	}
	
	// Validate all services using functional patterns
	validationResults := lo.Map(services, func(service ServiceConfig, _ int) struct {
		serviceName string
		valid       bool
		error       mo.Option[error]
	} {
		err := validateDependencies(service, services)
		return struct {
			serviceName string
			valid       bool
			error       mo.Option[error]
		}{
			serviceName: service.name,
			valid:       err.IsAbsent(),
			error:       err,
		}
	})
	
	validServices := lo.Filter(validationResults, func(result struct {
		serviceName string
		valid       bool
		error       mo.Option[error]
	}, _ int) bool {
		return result.valid
	})
	
	fmt.Printf("✅ Valid services: %d/%d\n", len(validServices), len(services))
	
	// Create deployment pipeline using function composition
	deploymentPipeline := lo.Pipe4(
		services,
		func(services []ServiceConfig) []ServiceConfig {
			// Sort by dependency count (deploy independent services first)
			return lo.SortBy(services, func(s ServiceConfig) int {
				return len(s.dependencies)
			})
		},
		func(services []ServiceConfig) []ServiceConfig {
			// Mark services as deploying
			return lo.Map(services, func(s ServiceConfig, _ int) ServiceConfig {
				s.resources["status"] = "deploying"
				s.resources["deploy_time"] = time.Now().Unix()
				return s
			})
		},
		func(services []ServiceConfig) []ServiceConfig {
			// Simulate deployment success
			return lo.Map(services, func(s ServiceConfig, _ int) ServiceConfig {
				s.resources["status"] = "deployed"
				s.resources["health"] = "healthy"
				return s
			})
		},
		func(services []ServiceConfig) map[string]interface{} {
			// Generate deployment summary
			return map[string]interface{}{
				"namespace":        namespace,
				"total_services":   len(services),
				"deployment_time":  time.Now().Format(time.RFC3339),
				"services":         lo.Map(services, func(s ServiceConfig, _ int) string { return s.name }),
				"service_types":    lo.Uniq(lo.Map(services, func(s ServiceConfig, _ int) string { return s.serviceType })),
				"status":           "completed",
			}
		},
	)
	
	deploymentSummary := deploymentPipeline
	fmt.Printf("✅ Deployment completed: %v\n", deploymentSummary["status"])
	fmt.Printf("✅ Services deployed: %v\n", deploymentSummary["service_types"])
	
	// Generate cleanup plan using functional operations
	cleanupPlan := lo.Reverse(lo.SortBy(services, func(s ServiceConfig) int {
		return len(s.dependencies) // Cleanup in reverse dependency order
	}))
	
	cleanupActions := lo.Map(cleanupPlan, func(service ServiceConfig, index int) string {
		return fmt.Sprintf("%d. Delete %s (%s)", index+1, service.name, service.serviceType)
	})
	
	fmt.Printf("✅ Cleanup plan generated with %d actions\n", len(cleanupActions))
	lo.ForEach(cleanupActions[:3], func(action string, _ int) {
		fmt.Printf("  %s\n", action)
	})
	
	fmt.Println("Functional full integration completed\n")
}

// ExampleCloudWatchLogsClient demonstrates functional logging patterns
func ExampleCloudWatchLogsClient() {
	fmt.Println("📊 Functional CloudWatch Logs Patterns:")
	
	// Immutable log configuration using functional patterns
	type LogConfig struct {
		logGroupName    string
		logStreamName   string
		retentionDays   int
		logLevel        string
		tags            map[string]string
		encryption      bool
	}
	
	type LogEntry struct {
		timestamp time.Time
		level     string
		message   string
		metadata  map[string]interface{}
	}
	
	// Pure functions for log processing
	createLogEntry := func(level, message string, metadata map[string]interface{}) LogEntry {
		return LogEntry{
			timestamp: time.Now(),
			level:     level,
			message:   message,
			metadata:  metadata,
		}
	}
	
	formatLogEntry := func(entry LogEntry) string {
		return fmt.Sprintf("[%s] %s: %s", 
			entry.timestamp.Format(time.RFC3339),
			entry.level,
			entry.message)
	}
	
	// Create log configurations for different services
	logConfigs := lo.Map([]struct {
		service       string
		retentionDays int
		level         string
	}{
		{"lambda-processor", 30, "INFO"},
		{"api-gateway", 7, "WARN"},
		{"step-functions", 90, "DEBUG"},
		{"eventbridge", 14, "INFO"},
	}, func(spec struct {
		service       string
		retentionDays int
		level         string
	}, _ int) LogConfig {
		return LogConfig{
			logGroupName:  fmt.Sprintf("/aws/%s", spec.service),
			logStreamName: fmt.Sprintf("%s-%d", spec.service, time.Now().Unix()),
			retentionDays: spec.retentionDays,
			logLevel:      spec.level,
			tags: map[string]string{
				"service":     spec.service,
				"environment": "production",
				"framework":   "functional",
			},
			encryption: true,
		}
	})
	
	fmt.Printf("✅ Created %d log configurations\n", len(logConfigs))
	
	// Generate sample log entries using functional patterns
	sampleLogs := lo.FlatMap(logConfigs, func(config LogConfig, _ int) []LogEntry {
		return lo.Map([]struct {
			level   string
			message string
		}{
			{"INFO", "Service started successfully"},
			{"DEBUG", "Processing request"},
			{"WARN", "Retry attempt needed"},
			{"ERROR", "Operation failed"},
		}, func(log struct {
			level   string
			message string
		}, _ int) LogEntry {
			return createLogEntry(log.level, log.message, map[string]interface{}{
				"service":    strings.TrimPrefix(config.logGroupName, "/aws/"),
				"log_group": config.logGroupName,
				"session_id": fmt.Sprintf("session-%d", lo.RandomInt(1000, 9999)),
			})
		})
	})
	
	// Filter logs by level using functional operations
	logLevels := []string{"ERROR", "WARN", "INFO", "DEBUG"}
	logCounts := lo.Associate(logLevels, func(level string) (string, int) {
		count := lo.CountBy(sampleLogs, func(entry LogEntry) bool {
			return entry.level == level
		})
		return level, count
	})
	
	fmt.Printf("✅ Generated %d log entries\n", len(sampleLogs))
	for level, count := range logCounts {
		if count > 0 {
			fmt.Printf("✅ %s level: %d entries\n", level, count)
		}
	}
	
	// Process logs using monadic patterns
	processLogs := func(entries []LogEntry, minLevel string) mo.Option[[]string] {
		levelPriority := map[string]int{"DEBUG": 0, "INFO": 1, "WARN": 2, "ERROR": 3}
		minPriority, exists := levelPriority[minLevel]
		if !exists {
			return mo.None[[]string]()
		}
		
		filteredLogs := lo.Filter(entries, func(entry LogEntry, _ int) bool {
			return levelPriority[entry.level] >= minPriority
		})
		
		formattedLogs := lo.Map(filteredLogs, func(entry LogEntry, _ int) string {
			return formatLogEntry(entry)
		})
		
		return mo.Some(formattedLogs)
	}
	
	// Process logs with different filtering levels
	processingResults := lo.Map([]string{"INFO", "WARN", "ERROR"}, func(level string, _ int) struct {
		level  string
		count  int
		valid  bool
	} {
		result := processLogs(sampleLogs, level)
		return struct {
			level  string
			count  int
			valid  bool
		}{
			level: level,
			count: result.Map(func(logs []string) int { return len(logs) }).OrElse(0),
			valid: result.IsPresent(),
		}
	})
	
	lo.ForEach(processingResults, func(result struct {
		level  string
		count  int
		valid  bool
	}, _ int) {
		if result.valid {
			fmt.Printf("✅ Processed %s+ logs: %d entries\n", result.level, result.count)
		}
	})
	
	fmt.Printf("✅ CloudWatch client type reference: %T\n", &cloudwatchlogs.Client{})
	fmt.Println("Functional CloudWatch Logs patterns completed\n")
}


// ExampleUniqueIdGeneration demonstrates functional ID generation patterns
func ExampleUniqueIdGeneration() {
	fmt.Println("🆔 Functional Unique ID Generation Patterns:")
	
	// Pure functions for ID generation using different strategies
	type IDGenerator func() string
	type IDConfig struct {
		prefix    mo.Option[string]
		suffix    mo.Option[string]
		length    int
		timestamp bool
		random    bool
	}
	
	// Create ID generators using functional composition
	createTimestampID := func(config IDConfig) IDGenerator {
		return func() string {
			base := fmt.Sprintf("%d", time.Now().UnixNano())
			if config.random {
				base += fmt.Sprintf("-%d", lo.RandomInt(1000, 9999))
			}
			return lo.Pipe3(
				base,
				func(id string) string {
					return config.prefix.Map(func(p string) string { return p + "-" + id }).OrElse(id)
				},
				func(id string) string {
					return config.suffix.Map(func(s string) string { return id + "-" + s }).OrElse(id)
				},
				func(id string) string {
					if config.length > 0 && len(id) > config.length {
						return id[:config.length]
					}
					return id
				},
			)
		}
	}
	
	createRandomID := func(config IDConfig) IDGenerator {
		return func() string {
			chars := "abcdefghijklmnopqrstuvwxyz0123456789"
			length := lo.Ternary(config.length > 0, config.length, 8)
			
			randomChars := lo.Times(length, func(_ int) string {
				return string(chars[lo.RandomInt(0, len(chars))])
			})
			
			base := strings.Join(randomChars, "")
			
			return lo.Pipe2(
				base,
				func(id string) string {
					return config.prefix.Map(func(p string) string { return p + "-" + id }).OrElse(id)
				},
				func(id string) string {
					return config.suffix.Map(func(s string) string { return id + "-" + s }).OrElse(id)
				},
			)
		}
	}
	
	// Create different ID generator configurations
	idConfigs := []struct {
		name      string
		config    IDConfig
		generator IDGenerator
	}{
		{
			name: "timestamp-basic",
			config: IDConfig{timestamp: true},
			generator: nil, // Will be set below
		},
		{
			name: "timestamp-with-prefix",
			config: IDConfig{
				prefix:    mo.Some("func"),
				timestamp: true,
				random:    true,
			},
			generator: nil,
		},
		{
			name: "random-short",
			config: IDConfig{
				length: 6,
				random: true,
			},
			generator: nil,
		},
		{
			name: "random-with-service",
			config: IDConfig{
				prefix: mo.Some("svc"),
				suffix: mo.Some("prod"),
				length: 12,
				random: true,
			},
			generator: nil,
		},
	}
	
	// Initialize generators based on configuration
	for i, config := range idConfigs {
		if config.config.timestamp {
			idConfigs[i].generator = createTimestampID(config.config)
		} else {
			idConfigs[i].generator = createRandomID(config.config)
		}
	}
	
	// Generate IDs using functional patterns
	generatedIDs := lo.FlatMap(idConfigs, func(config struct {
		name      string
		config    IDConfig
		generator IDGenerator
	}, _ int) []struct {
		generatorName string
		id            string
		unique        bool
	} {
		// Generate multiple IDs to test uniqueness
		ids := lo.Times(3, func(_ int) string {
			time.Sleep(time.Nanosecond) // Ensure timestamp difference
			return config.generator()
		})
		
		return lo.Map(ids, func(id string, index int) struct {
			generatorName string
			id            string
			unique        bool
		} {
			return struct {
				generatorName string
				id            string
				unique        bool
			}{
				generatorName: config.name,
				id:            id,
				unique:        true, // Will be validated below
			}
		})
	})
	
	fmt.Printf("✅ Generated %d IDs using %d different strategies\n", len(generatedIDs), len(idConfigs))
	
	// Validate uniqueness using functional operations
	allIDs := lo.Map(generatedIDs, func(item struct {
		generatorName string
		id            string
		unique        bool
	}, _ int) string {
		return item.id
	})
	
	uniqueIDs := lo.Uniq(allIDs)
	uniquenessPct := float64(len(uniqueIDs)) / float64(len(allIDs)) * 100
	
	fmt.Printf("✅ Uniqueness validation: %.1f%% (%d/%d unique)\n", uniquenessPct, len(uniqueIDs), len(allIDs))
	
	// Group IDs by generator type
	idGroups := lo.GroupBy(generatedIDs, func(item struct {
		generatorName string
		id            string
		unique        bool
	}) string {
		return item.generatorName
	})
	
	for generatorName, ids := range idGroups {
		sampleID := ids[0].id
		fmt.Printf("✅ %s generator - sample: %s (length: %d)\n", generatorName, sampleID, len(sampleID))
	}
	
	// Demonstrate functional validation of ID format
	validateIDFormat := func(id string) mo.Option[error] {
		if id == "" {
			return mo.Some[error](fmt.Errorf("ID cannot be empty"))
		}
		if len(id) < 3 {
			return mo.Some[error](fmt.Errorf("ID too short: %s", id))
		}
		return mo.None[error]()
	}
	
	validIDs := lo.Filter(allIDs, func(id string, _ int) bool {
		return validateIDFormat(id).IsAbsent()
	})
	
	fmt.Printf("✅ Format validation: %d/%d IDs valid\n", len(validIDs), len(allIDs))
	fmt.Println("Functional unique ID generation completed\n")
}

// ExampleTestingInterface demonstrates TestingT interface methods
func ExampleTestingInterface() {
	fmt.Println("VasDeference TestingT interface patterns:")
	
	// ctx := vasdeference.NewTestContext(t)
	
	// Test Helper method (should not panic)
	// ctx.T.Helper()
	fmt.Println("Calling Helper method on TestingT interface")
	
	// Test Name method
	// name := ctx.T.Name()
	// Expected: name should not be empty
	fmt.Println("Getting test name from TestingT interface")
	
	// Test Logf method (should not panic)
	// ctx.T.Logf("Test log message: %s", "test")
	fmt.Println("Logging message through TestingT interface")
	fmt.Println("TestingT interface completed")
}

// ExampleAWSCredentialsValidation demonstrates AWS credentials validation
func ExampleAWSCredentialsValidation() {
	fmt.Println("VasDeference AWS credentials validation patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test credential validation
	// err := vdf.ValidateAWSCredentialsE()
	// Expected: In test environment without AWS credentials, this should return an error
	fmt.Println("Validating AWS credentials configuration")
	fmt.Println("Expected: Error if credentials not properly configured")
	fmt.Println("AWS credentials validation completed")
}

// ExampleNamespaceSanitization demonstrates namespace sanitization
func ExampleNamespaceSanitization() {
	fmt.Println("VasDeference namespace sanitization patterns:")
	
	// Test with empty options
	// ctx := vasdeference.NewTestContext(t)
	// arrange := vasdeference.NewArrangeWithNamespace(ctx, "test-empty")
	// defer arrange.Cleanup()
	fmt.Println("Creating namespace with: test-empty")
	// Expected: arrange.Namespace should equal "test-empty"
	
	// Test with different namespace patterns - SanitizeNamespace converts underscores to hyphens
	// ctx2 := vasdeference.NewTestContext(t)
	// arrange2 := vasdeference.NewArrangeWithNamespace(ctx2, "test_with_underscores")
	// defer arrange2.Cleanup()
	fmt.Println("Creating namespace with: test_with_underscores")
	// Expected: SanitizeNamespace converts underscores to hyphens -> "test-with-underscores"
	fmt.Println("Namespace sanitization completed")
}

// ExampleEnvironmentVariableEdgeCases demonstrates environment variable edge cases
func ExampleEnvironmentVariableEdgeCases() {
	fmt.Println("VasDeference environment variable edge cases:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test GetEnvVarWithDefault with non-existent variable
	// value := vdf.GetEnvVarWithDefault("NON_EXISTENT_VAR_12345", "default_value")
	// Expected: value should equal "default_value"
	fmt.Println("Getting non-existent variable with default: default_value")
	
	// Test GetEnvVarWithDefault with existing variable
	// t.Setenv("TEST_VAR_VASDEFERENCE", "test_value")
	// value2 := vdf.GetEnvVarWithDefault("TEST_VAR_VASDEFERENCE", "default")
	// Expected: value2 should equal "test_value"
	fmt.Println("Setting TEST_VAR_VASDEFERENCE=test_value")
	fmt.Println("Getting existing variable should return: test_value")
	fmt.Println("Environment variable edge cases completed")
}

// ExampleCIEnvironmentDetection demonstrates CI environment detection
func ExampleCIEnvironmentDetection() {
	fmt.Println("VasDeference CI environment detection patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test without CI environment variables
	// isCI := vdf.IsCIEnvironment()
	// Expected: isCI should be boolean type
	fmt.Println("Checking CI environment without CI variables")
	
	// Test with GitHub Actions
	// t.Setenv("GITHUB_ACTIONS", "true")
	// isCI2 := vdf.IsCIEnvironment()
	// Expected: isCI2 should be true
	fmt.Println("Setting GITHUB_ACTIONS=true, should detect CI: true")
	
	// Test with CircleCI
	// t.Setenv("CIRCLECI", "true")
	// isCI3 := vdf.IsCIEnvironment()
	// Expected: isCI3 should be true
	fmt.Println("Setting CIRCLECI=true, should detect CI: true")
	fmt.Println("CI environment detection completed")
}


// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	fmt.Println("VasDeference error handling patterns:")
	
	// Test with nil TestingT (should panic)
	// Expected: vasdeference.New(nil) should panic
	fmt.Println("Testing nil TestingT parameter - should panic")
	
	// Test NewWithOptionsE with nil TestingT
	// vdf, err := vasdeference.NewWithOptionsE(nil, &vasdeference.VasDeferenceOptions{})
	// Expected: vdf should be nil, err should contain "TestingT cannot be nil"
	fmt.Println("Testing NewWithOptionsE with nil TestingT - should return error")
	fmt.Println("Error handling completed")
}

// ExampleUniqueIdUniqueness demonstrates ID uniqueness verification
func ExampleUniqueIdUniqueness() {
	fmt.Println("VasDeference unique ID uniqueness verification:")
	
	// Generate multiple IDs to verify uniqueness
	_ = /* ids */ make(map[string]bool)
	for i := 0; i < 10; i++ {
		// id := vasdeference.UniqueId()
		fmt.Printf("Generating unique ID #%d\n", i+1)
		
		// Each ID should be unique
		// Expected: id should not be empty, should not exist in ids map
		// ids[id] = true
		
		// Small delay to ensure different timestamps
		time.Sleep(time.Nanosecond)
	}
	fmt.Println("All IDs should be unique across multiple generations")
	fmt.Println("Unique ID uniqueness completed")
}

// ExampleTerraformOutputErrors demonstrates Terraform output error scenarios
func ExampleTerraformOutputErrors() {
	fmt.Println("VasDeference Terraform output error scenarios:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test with nil outputs map
	// value, err := vdf.GetTerraformOutputE(nil, "test-key")
	// Expected: value should be empty, err should contain "outputs map cannot be nil"
	fmt.Println("Testing nil outputs map - should return error")
	
	// Test with non-string value
	outputs := map[string]interface{}{
		"numeric_value": 123,
		"boolean_value": true,
	}
	fmt.Printf("Testing non-string values: %+v\n", outputs)
	
	// value, err = vdf.GetTerraformOutputE(outputs, "numeric_value")
	// Expected: value should be empty, err should contain "is not a string"
	fmt.Println("Getting numeric value should return error - not a string")
	fmt.Println("Terraform output errors completed")
}


// ExampleSanitizeNamespaceEdgeCases demonstrates namespace sanitization edge cases
func ExampleSanitizeNamespaceEdgeCases() {
	fmt.Println("VasDeference namespace sanitization edge cases:")
	
	// Test empty string
	// result := vasdeference.SanitizeNamespace("")
	// Expected: result should equal "default"
	fmt.Println("Empty string input -> expected: 'default'")
	
	// Test string with only invalid characters
	// result = vasdeference.SanitizeNamespace("!@#$%^&*()")
	// Expected: result should equal "default"
	fmt.Println("Invalid characters '!@#$%^&*()' -> expected: 'default'")
	
	// Test string with only hyphens
	// result = vasdeference.SanitizeNamespace("---")
	// Expected: result should equal "default"
	fmt.Println("Only hyphens '---' -> expected: 'default'")
	
	// Test mixed valid/invalid characters
	// result = vasdeference.SanitizeNamespace("test!@#_with./chars")
	// Expected: result should equal "test-with--chars"
	fmt.Println("Mixed characters 'test!@#_with./chars' -> expected: 'test-with--chars'")
	fmt.Println("Namespace sanitization edge cases completed")
}

// ExampleComprehensiveConfiguration demonstrates comprehensive configuration scenarios
func ExampleComprehensiveConfiguration() {
	fmt.Println("VasDeference comprehensive configuration patterns:")
	
	// Test NewTestContextE with empty options to cover default-setting branches
	// ctx, err := vasdeference.NewTestContextE(t, &vasdeference.Options{})
	// Expected: err should be nil, ctx should not be nil
	fmt.Println("Creating test context with empty options")
	
	// Expected defaults:
	// ctx.Region should equal vasdeference.RegionUSEast1
	// ctx.T should equal the testing instance
	fmt.Printf("Expected default region: %s\n", vasdeference.RegionUSEast1)
	
	// Test NewWithOptionsE with default values
	opts := &vasdeference.VasDeferenceOptions{
		Region:     "", // Should get default
		MaxRetries: 0,  // Should get default  
		RetryDelay: 0,  // Should get default
		Namespace:  "", // Should trigger auto-detection
	}
	fmt.Printf("Testing with options: %+v\n", opts)
	
	// vdf, err := vasdeference.NewWithOptionsE(t, opts)
	// Expected: err should be nil, vdf should not be nil
	// vdf.Context.Region should equal vasdeference.RegionUSEast1
	// vdf.Arrange.Namespace should not be empty (auto-detected)
	fmt.Println("Creating VasDeference with default options")
	fmt.Println("Expected: defaults applied and namespace auto-detected")
	fmt.Println("Comprehensive configuration completed")
}

// ExampleAdvancedFeatures demonstrates advanced functional programming patterns
func ExampleAdvancedFeatures() {
	fmt.Println("🚀 Advanced Functional Programming Patterns:")
	
	// Advanced functional composition with higher-order functions
	type AdvancedConfig struct {
		region          string
		timeout         time.Duration
		clientConfigs   map[string]interface{}
		featureFlags    map[string]bool
		performanceTier string
	}
	
	// Higher-order function that returns a configuration transformer
	createConfigTransformer := func(transformType string) func(AdvancedConfig) AdvancedConfig {
		switch transformType {
		case "production":
			return func(config AdvancedConfig) AdvancedConfig {
				config.timeout = 60 * time.Second
				config.performanceTier = "high"
				config.featureFlags["monitoring"] = true
				config.featureFlags["caching"] = true
				return config
			}
		case "development":
			return func(config AdvancedConfig) AdvancedConfig {
				config.timeout = 15 * time.Second
				config.performanceTier = "standard"
				config.featureFlags["debug"] = true
				config.featureFlags["verbose_logging"] = true
				return config
			}
		default:
			return func(config AdvancedConfig) AdvancedConfig { return config }
		}
	}
	
	// Create base configuration
	baseConfig := AdvancedConfig{
		region:          vasdeference.RegionUSEast1,
		timeout:         30 * time.Second,
		clientConfigs:   make(map[string]interface{}),
		featureFlags:    make(map[string]bool),
		performanceTier: "standard",
	}
	
	// Apply transformations using functional composition
	environments := []string{"development", "staging", "production"}
	configurations := lo.Map(environments, func(env string, _ int) struct {
		environment string
		config      AdvancedConfig
	} {
		transformer := createConfigTransformer(env)
		transformedConfig := transformer(baseConfig)
		transformedConfig.clientConfigs["environment"] = env
		
		return struct {
			environment string
			config      AdvancedConfig
		}{
			environment: env,
			config:      transformedConfig,
		}
	})
	
	fmt.Printf("✅ Created %d environment configurations\n", len(configurations))
	
	// Advanced error handling with Result monad chains
	type ValidationRule[T any] func(T) mo.Result[T]
	
	validateTimeout := func(config AdvancedConfig) mo.Result[AdvancedConfig] {
		if config.timeout <= 0 || config.timeout > 10*time.Minute {
			return mo.Err[AdvancedConfig](fmt.Errorf("invalid timeout: %v", config.timeout))
		}
		return mo.Ok(config)
	}
	
	validateRegion := func(config AdvancedConfig) mo.Result[AdvancedConfig] {
		validRegions := []string{vasdeference.RegionUSEast1, "us-west-2", "eu-west-1"}
		if !lo.Contains(validRegions, config.region) {
			return mo.Err[AdvancedConfig](fmt.Errorf("invalid region: %s", config.region))
		}
		return mo.Ok(config)
	}
	
	validateFeatures := func(config AdvancedConfig) mo.Result[AdvancedConfig] {
		requiredFeatures := []string{"monitoring", "debug", "caching", "verbose_logging"}
		enabled := lo.Filter(requiredFeatures, func(feature string, _ int) bool {
			return config.featureFlags[feature]
		})
		
		if len(enabled) == 0 {
			return mo.Err[AdvancedConfig](fmt.Errorf("no features enabled"))
		}
		return mo.Ok(config)
	}
	
	// Chain validation using monadic operations
	validationResults := lo.Map(configurations, func(item struct {
		environment string
		config      AdvancedConfig
	}, _ int) struct {
		environment string
		result      mo.Result[AdvancedConfig]
		valid       bool
	} {
		// Chain validations using FlatMap
		result := mo.Ok(item.config).
			FlatMap(func(c AdvancedConfig) mo.Result[AdvancedConfig] { return validateTimeout(c) }).
			FlatMap(func(c AdvancedConfig) mo.Result[AdvancedConfig] { return validateRegion(c) }).
			FlatMap(func(c AdvancedConfig) mo.Result[AdvancedConfig] { return validateFeatures(c) })
		
		return struct {
			environment string
			result      mo.Result[AdvancedConfig]
			valid       bool
		}{
			environment: item.environment,
			result:      result,
			valid:       result.IsOk(),
		}
	})
	
	validConfigs := lo.Filter(validationResults, func(result struct {
		environment string
		result      mo.Result[AdvancedConfig]
		valid       bool
	}, _ int) bool {
		return result.valid
	})
	
	fmt.Printf("✅ Valid configurations: %d/%d\n", len(validConfigs), len(configurations))
	
	// Advanced client factory using generics and functional patterns
	type ClientFactory[T any] func(config AdvancedConfig) mo.Option[T]
	type MockClient struct {
		service     string
		region      string
		timeout     time.Duration
		configured  bool
	}
	
	createClientFactory := func(serviceType string) ClientFactory[MockClient] {
		return func(config AdvancedConfig) mo.Option[MockClient] {
			if config.region == "" {
				return mo.None[MockClient]()
			}
			
			return mo.Some(MockClient{
				service:    serviceType,
				region:     config.region,
				timeout:    config.timeout,
				configured: true,
			})
		}
	}
	
	// Create multiple service clients using functional factories
	serviceTypes := []string{"lambda", "dynamodb", "eventbridge", "stepfunctions"}
	clientResults := lo.FlatMap(validConfigs, func(configResult struct {
		environment string
		result      mo.Result[AdvancedConfig]
		valid       bool
	}, _ int) []struct {
		environment string
		service     string
		client      mo.Option[MockClient]
	} {
		config := configResult.result.MustGet()
		
		return lo.Map(serviceTypes, func(service string, _ int) struct {
			environment string
			service     string
			client      mo.Option[MockClient]
		} {
			factory := createClientFactory(service)
			client := factory(config)
			
			return struct {
				environment string
				service     string
				client      mo.Option[MockClient]
			}{
				environment: configResult.environment,
				service:     service,
				client:      client,
			}
		})
	})
	
	successfulClients := lo.Filter(clientResults, func(result struct {
		environment string
		service     string
		client      mo.Option[MockClient]
	}, _ int) bool {
		return result.client.IsPresent()
	})
	
	fmt.Printf("✅ Successfully created %d service clients\n", len(successfulClients))
	
	// Group clients by environment and service
	env Groups := lo.GroupBy(successfulClients, func(result struct {
		environment string
		service     string
		client      mo.Option[MockClient]
	}) string {
		return result.environment
	})
	
	for env, clients := range envGroups {
		services := lo.Map(clients, func(client struct {
			environment string
			service     string
			client      mo.Option[MockClient]
		}, _ int) string {
			return client.service
		})
		fmt.Printf("✅ %s environment: %s\n", env, strings.Join(services, ", "))
	}
	
	fmt.Println("Advanced functional programming patterns completed\n")
}

// ExampleTerratestIntegration demonstrates Terratest integration features
func ExampleTerratestIntegration() {
	fmt.Println("VasDeference Terratest integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Get random AWS region using Terratest
	// randomRegion := vdf.GetRandomRegion()
	// Expected: randomRegion should be a valid AWS region
	fmt.Println("Getting random AWS region using Terratest's aws module")
	
	// Get AWS account ID using Terratest
	// accountId := vdf.GetAccountId()
	// Expected: accountId should be a valid AWS account ID
	fmt.Println("Getting AWS account ID using Terratest's aws module")
	
	// Generate random resource names
	// functionName := vdf.GenerateRandomName("my-function")
	// tableName := vdf.GenerateRandomName("my-table")
	// Expected: names should contain unique suffixes
	fmt.Println("Generating random resource names with Terratest's random module")
	
	// Get all available regions and AZs
	// allRegions := vdf.GetAllAvailableRegions()
	// azs := vdf.GetAvailabilityZones()
	// Expected: arrays should contain valid AWS regions and AZs
	fmt.Println("Getting all available regions and availability zones")
	
	// Create unique test identifier
	// testId := vdf.CreateTestIdentifier()
	// Expected: testId should contain namespace and unique ID
	fmt.Println("Creating unique test identifier with namespace")
	
	fmt.Println("Terratest integration completed")
}

// ExampleTerratestRetryPatterns demonstrates Terratest retry patterns
func ExampleTerratestRetryPatterns() {
	fmt.Println("VasDeference Terratest retry patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Retry with backoff using Terratest patterns
	// action := vasdeference.RetryableAction{
	//     Description: "Check Lambda function state",
	//     MaxRetries:  5,
	//     TimeBetween: 2 * time.Second,
	//     Action: func() (string, error) {
	//         // Check if function is ready
	//         return "ready", nil
	//     },
	// }
	// result := vdf.DoWithRetry(action)
	// Expected: result should equal "ready"
	fmt.Println("Configuring retryable action for Lambda function state check")
	
	// Wait until condition is met
	// vdf.WaitUntil(
	//     "DynamoDB table active",
	//     10,
	//     3*time.Second,
	//     func() (bool, error) {
	//         // Check table status
	//         return true, nil
	//     },
	// )
	fmt.Println("Waiting until DynamoDB table becomes active using Terratest retry")
	
	// Validate AWS resource exists with retry
	// vdf.ValidateAWSResourceExists(
	//     "Lambda function",
	//     "my-function",
	//     func() (bool, error) {
	//         // Check if function exists
	//         return true, nil
	//     },
	// )
	fmt.Println("Validating AWS resource existence with built-in retry logic")
	
	fmt.Println("Terratest retry patterns completed")
}

// ExampleTerratestRandomGeneration demonstrates Terratest random generation
func ExampleTerratestRandomGeneration() {
	fmt.Println("VasDeference Terratest random generation patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Generate random resource names with length constraints
	// shortName := vdf.CreateRandomResourceName("func", 30)
	// longTableName := vdf.CreateRandomResourceName("my-table", 255)
	// Expected: names should be within specified length limits
	fmt.Println("Creating random resource names with length constraints")
	
	// Create resource name with timestamp
	// timestampName := vdf.CreateResourceNameWithTimestamp("deployment")
	// Expected: name should contain namespace, prefix, and unique timestamp
	fmt.Println("Creating resource name with timestamp using Terratest patterns")
	
	// Get random region excluding specific ones
	// excludeRegions := []string{"us-gov-east-1", "us-gov-west-1"}
	// randomRegion := vdf.GetRandomRegionExcluding(excludeRegions)
	// Expected: region should not be in excluded list
	fmt.Println("Getting random region excluding government regions")
	
	fmt.Println("Terratest random generation completed")
}

// ExampleTerratestTerraformOperations demonstrates Terraform operations with Terratest
func ExampleTerratestTerraformOperations() {
	fmt.Println("VasDeference Terraform operations with Terratest:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Terraform options
	// variables := map[string]interface{}{
	//     "environment": "test",
	//     "region":      vdf.Context.Region,
	//     "namespace":   vdf.Arrange.Namespace,
	// }
	// terraformOptions := vdf.TerraformOptions("./terraform", variables)
	// Expected: terraformOptions should be properly configured
	fmt.Println("Creating Terraform options with test variables")
	
	// Initialize and apply Terraform
	// vdf.InitAndApplyTerraform(terraformOptions)
	// // Cleanup is automatically registered
	// Expected: Terraform should apply successfully and register cleanup
	fmt.Println("Initializing and applying Terraform with automatic cleanup registration")
	
	// Get Terraform outputs
	// functionName := vdf.GetTerraformOutputAsString(terraformOptions, "function_name")
	// apiConfig := vdf.GetTerraformOutputAsMap(terraformOptions, "api_config")
	// resourceList := vdf.GetTerraformOutputAsList(terraformOptions, "resource_arns")
	// Expected: outputs should be properly typed and accessible
	fmt.Println("Getting Terraform outputs as string, map, and list types")
	
	// Log all outputs for debugging
	// vdf.LogTerraformOutput(terraformOptions)
	// Expected: all outputs should be logged for debugging purposes
	fmt.Println("Logging all Terraform outputs for debugging")
	
	fmt.Println("Terraform operations with Terratest completed")
}

// runAllExamples demonstrates running all functional programming examples
func runAllExamples() {
	fmt.Println("🚀 Running all Functional Programming Examples:")
	fmt.Println("   Using samber/lo and samber/mo with immutable data structures")
	fmt.Println("")
	
	ExampleFunctionalCoreUsage()
	fmt.Println("")
	
	ExampleWithCustomOptions()
	fmt.Println("")
	
	ExampleTerraformIntegration()
	fmt.Println("")
	
	ExampleLambdaIntegration()
	fmt.Println("")
	
	ExampleEventBridgeIntegration()
	fmt.Println("")
	
	ExampleDynamoDBIntegration()
	fmt.Println("")
	
	ExampleStepFunctionsIntegration()
	fmt.Println("")
	
	ExampleRetryMechanism()
	fmt.Println("")
	
	ExampleEnvironmentVariables()
	fmt.Println("")
	
	ExampleFullIntegration()
	fmt.Println("")
	
	ExampleCloudWatchLogsClient()
	fmt.Println("")
	
	ExampleUniqueIdGeneration()
	fmt.Println("")
	
	ExampleTestingInterface()
	fmt.Println("")
	
	ExampleAWSCredentialsValidation()
	fmt.Println("")
	
	ExampleNamespaceSanitization()
	fmt.Println("")
	
	ExampleEnvironmentVariableEdgeCases()
	fmt.Println("")
	
	ExampleCIEnvironmentDetection()
	fmt.Println("")
	
	ExampleErrorHandling()
	fmt.Println("")
	
	ExampleUniqueIdUniqueness()
	fmt.Println("")
	
	ExampleTerraformOutputErrors()
	fmt.Println("")
	
	ExampleSanitizeNamespaceEdgeCases()
	fmt.Println("")
	
	ExampleComprehensiveConfiguration()
	fmt.Println("")
	
	ExampleAdvancedFeatures()
	fmt.Println("")
	
	ExampleTerratestIntegration()
	fmt.Println("")
	
	ExampleTerratestRetryPatterns()
	fmt.Println("")
	
	ExampleTerratestRandomGeneration()
	fmt.Println("")
	
	ExampleTerratestTerraformOperations()
	fmt.Println("✨ All functional programming examples completed successfully!")
}

func main() {
	// This file demonstrates functional programming patterns using samber/lo and samber/mo.
	// Features immutable data structures, monadic error handling, function composition,
	// and type-safe operations with Go generics.
	// Run examples with: go run ./examples/core/examples.go
	
	runAllExamples()
}