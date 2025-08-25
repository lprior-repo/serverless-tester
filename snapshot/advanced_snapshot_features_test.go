package snapshot

import (
	"encoding/json" 
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Following TDD: Add advanced snapshot features for AWS resource state management

func TestAWSResourceStateCapture(t *testing.T) {
	t.Run("captures Lambda function configuration state", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		lambdaConfig := map[string]interface{}{
			"FunctionName":    "my-test-function",
			"Runtime":         "nodejs18.x",
			"Handler":         "index.handler",
			"MemorySize":      128,
			"Timeout":         30,
			"Environment": map[string]interface{}{
				"Variables": map[string]string{
					"NODE_ENV": "production",
					"API_URL":  "https://api.example.com",
				},
			},
			"LastModified": "2023-01-01T12:00:00.000Z",
			"Version":      "$LATEST",
		}

		snapshot.MatchLambdaResponse("lambda_config", mustMarshal(lambdaConfig))

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		// Verify timestamp sanitization occurred
		assert.Contains(t, string(content), "<TIMESTAMP>")
		assert.NotContains(t, string(content), "2023-01-01T12:00:00.000Z")
	})

	t.Run("captures DynamoDB table configuration state", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		tableConfig := map[string]interface{}{
			"TableName": "test-table",
			"TableStatus": "ACTIVE",
			"KeySchema": []map[string]string{
				{"AttributeName": "id", "KeyType": "HASH"},
			},
			"AttributeDefinitions": []map[string]string{
				{"AttributeName": "id", "AttributeType": "S"},
			},
			"BillingModeSummary": map[string]string{
				"BillingMode": "PAY_PER_REQUEST",
			},
			"CreationDateTime": time.Now().Format(time.RFC3339),
			"TableSizeBytes":   0,
			"ItemCount":        0,
		}

		// Use DynamoDB-specific matcher
		snapshot.MatchJSON("dynamodb_table_config", tableConfig)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test-table", parsed["TableName"])
		assert.Equal(t, "ACTIVE", parsed["TableStatus"])
	})

	t.Run("captures EventBridge rule configuration state", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		ruleConfig := map[string]interface{}{
			"Name":         "test-rule",
			"Description":  "Test rule for snapshot testing",
			"State":        "ENABLED",
			"EventPattern": `{"source": ["myapp"], "detail-type": ["Order Placed"]}`,
			"ScheduleExpression": "",
			"Targets": []map[string]interface{}{
				{
					"Id":  "1",
					"Arn": "arn:aws:lambda:us-east-1:123456789012:function:test-handler",
					"RoleArn": "arn:aws:iam::123456789012:role/service-role/test-role",
				},
			},
			"CreatedBy": "550e8400-e29b-41d4-a716-446655440000",
		}

		// Apply EventBridge-specific sanitizers
		sanitized := snapshot.WithSanitizers(
			SanitizeUUIDs(),
			SanitizeCustom(`arn:aws:lambda:[^:]+:\d+:function:[^"]+`, "<LAMBDA_ARN>"),
			SanitizeCustom(`arn:aws:iam::\d+:role/[^"]+`, "<IAM_ROLE_ARN>"),
		)

		sanitized.MatchJSON("eventbridge_rule_config", ruleConfig)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "<UUID>")
		assert.Contains(t, contentStr, "<LAMBDA_ARN>")
		assert.Contains(t, contentStr, "<IAM_ROLE_ARN>")
	})
}

func TestInfrastructureDiffDetection(t *testing.T) {
	t.Run("detects changes in Lambda environment variables", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Initial state
		initialConfig := map[string]interface{}{
			"Environment": map[string]interface{}{
				"Variables": map[string]string{
					"NODE_ENV": "development",
					"DEBUG":    "true",
				},
			},
		}

		// Create baseline snapshot
		snapshot.MatchJSON("lambda_env_baseline", initialConfig)

		// Modified state
		modifiedConfig := map[string]interface{}{
			"Environment": map[string]interface{}{
				"Variables": map[string]string{
					"NODE_ENV": "production", // Changed
					"DEBUG":    "false",      // Changed
					"NEW_VAR":  "added",      // Added
				},
			},
		}

		// This should create a diff when compared
		snapshot2 := New(t, Options{SnapshotDir: tempDir})
		
		err := snapshot2.MatchE("lambda_env_baseline", modifiedConfig)
		assert.Error(t, err, "Should detect differences")
		assert.Contains(t, err.Error(), "snapshot mismatch")
	})

	t.Run("detects changes in DynamoDB table configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Initial table config
		initialTable := map[string]interface{}{
			"ProvisionedThroughput": map[string]int{
				"ReadCapacityUnits":  5,
				"WriteCapacityUnits": 5,
			},
			"BillingMode": "PROVISIONED",
		}

		snapshot.MatchJSON("table_baseline", initialTable)

		// Modified table config - switched to on-demand
		modifiedTable := map[string]interface{}{
			"BillingMode": "PAY_PER_REQUEST",
			// ProvisionedThroughput removed in on-demand mode
		}

		err := snapshot.MatchE("table_baseline", modifiedTable)
		assert.Error(t, err, "Should detect billing mode change")
	})
}

func TestRegressionTesting(t *testing.T) {
	t.Run("prevents regression in Lambda response structure", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Expected Lambda response structure
		expectedResponse := map[string]interface{}{
			"statusCode": 200,
			"headers": map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Requested-With",
			},
			"body": map[string]interface{}{
				"success": true,
				"data":    []string{"item1", "item2"},
				"count":   2,
			},
			"isBase64Encoded": false,
		}

		snapshot.MatchLambdaResponse("api_response_structure", mustMarshal(expectedResponse))

		// Later, if someone accidentally changes the structure...
		brokenResponse := map[string]interface{}{
			"statusCode": 200,
			"body":       "plain string instead of object", // Regression!
		}

		err := snapshot.MatchE("api_response_structure", mustMarshal(brokenResponse))
		assert.Error(t, err, "Should catch regression in response structure")
	})

	t.Run("prevents regression in Step Functions state machine definition", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Expected state machine definition
		stateMachineDefinition := map[string]interface{}{
			"Comment": "Order processing workflow",
			"StartAt": "ValidateOrder",
			"States": map[string]interface{}{
				"ValidateOrder": map[string]interface{}{
					"Type":     "Task",
					"Resource": "arn:aws:lambda:region:account:function:validate-order",
					"Next":     "ProcessPayment",
				},
				"ProcessPayment": map[string]interface{}{
					"Type":     "Task", 
					"Resource": "arn:aws:lambda:region:account:function:process-payment",
					"End":      true,
				},
			},
		}

		snapshot.MatchStepFunctionExecution("order_workflow_definition", stateMachineDefinition)

		// Verify snapshot can detect structural changes
		modifiedDefinition := map[string]interface{}{
			"Comment": "Order processing workflow",
			"StartAt": "ValidateOrder",
			"States": map[string]interface{}{
				"ValidateOrder": map[string]interface{}{
					"Type":     "Task",
					"Resource": "arn:aws:lambda:region:account:function:validate-order",
					// "Next" field removed - this breaks the workflow!
					"End": true,
				},
			},
		}

		err := snapshot.MatchE("order_workflow_definition", modifiedDefinition)
		assert.Error(t, err, "Should catch workflow structure changes")
	})
}

func TestStateComparisonAndValidation(t *testing.T) {
	t.Run("validates AWS resource state consistency", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{
			SnapshotDir: tempDir,
			Sanitizers: []Sanitizer{
				SanitizeTimestamps(),
				SanitizeUUIDs(),
				SanitizeCustom(`arn:aws:[^:]+:[^:]+:\d+:[^"]+`, "<AWS_ARN>"),
			},
		})

		// Multi-resource state snapshot
		resourceState := map[string]interface{}{
			"lambda_functions": []map[string]interface{}{
				{
					"FunctionName": "api-handler",
					"Runtime":      "nodejs18.x",
					"FunctionArn":  "arn:aws:lambda:us-east-1:123456789012:function:api-handler",
				},
			},
			"dynamodb_tables": []map[string]interface{}{
				{
					"TableName":   "users",
					"TableStatus": "ACTIVE",
					"TableArn":    "arn:aws:dynamodb:us-east-1:123456789012:table/users",
				},
			},
			"eventbridge_rules": []map[string]interface{}{
				{
					"Name":  "user-events",
					"State": "ENABLED",
					"Arn":   "arn:aws:events:us-east-1:123456789012:rule/user-events",
				},
			},
			"timestamp": "2023-01-01T12:00:00Z",
			"deployment_id": "550e8400-e29b-41d4-a716-446655440000",
		}

		snapshot.MatchJSON("infrastructure_state", resourceState)

		// Verify ARNs and dynamic values were sanitized
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "<AWS_ARN>")
		assert.Contains(t, contentStr, "<TIMESTAMP>")
		assert.Contains(t, contentStr, "<UUID>")
		assert.NotContains(t, contentStr, "123456789012")
		assert.NotContains(t, contentStr, "2023-01-01T12:00:00Z")
	})
}

// Helper function for marshaling data without error handling in tests
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}