package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
	"vasdeference/parallel"
	"vasdeference/snapshot"
)

// TestIntegration demonstrates how snapshot and parallel packages work together with vasdeference
func TestIntegration(t *testing.T) {
	t.Run("snapshot and parallel integration", func(t *testing.T) {
		// Create test context and arrange
		testCtx := sfx.NewTestContext(t)
		arrange := sfx.NewArrange(testCtx)
		defer arrange.Cleanup()

		// Create a snapshot tester
		snap := snapshot.New(t, snapshot.Options{
			SnapshotDir: t.TempDir(),
			Sanitizers: []snapshot.Sanitizer{
				snapshot.SanitizeTimestamps(),
				snapshot.SanitizeUUIDs(),
			},
		})

		// Create a parallel runner with the arrange context
		runner, err := parallel.NewRunnerWithArrange(arrange, parallel.Options{
			PoolSize: 2,
		})
		require.NoError(t, err)

		// Define test cases that use snapshots
		testCases := []parallel.TestCase{
			{
				Name: "lambda_response_test",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Simulate a Lambda response
					response := map[string]interface{}{
						"statusCode": 200,
						"body":       "Hello from test1",
						"timestamp":  "2023-01-01T10:00:00Z",
						"requestId":  "abc-123",
						"namespace":  tctx.GetFunctionName("handler"),
					}

					// Take a snapshot of the response
					snap.MatchJSON("lambda_response_test1", response)
					return nil
				},
			},
			{
				Name: "dynamodb_items_test", 
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Simulate DynamoDB items
					items := []map[string]interface{}{
						{
							"id":        "item-1",
							"tableName": tctx.GetTableName("users"),
							"timestamp": "2023-01-01T11:00:00Z",
						},
					}

					// Take a snapshot of the items
					snap.MatchDynamoDBItems("dynamodb_items_test1", items)
					return nil
				},
			},
		}

		// Run the test cases in parallel
		runner.Run(testCases)

		// The test cases run successfully if we get here
		// Individual snapshots are verified within each test case
	})

	t.Run("terratest style error handling", func(t *testing.T) {
		// Test the E (error-returning) variants
		testCtx := sfx.NewTestContext(t)
		
		snap, err := snapshot.NewE(testCtx.T)
		require.NoError(t, err)
		assert.NotNil(t, snap)

		// Test snapshot matching with error handling
		err = snap.MatchE("test", "test data")
		assert.NoError(t, err)

		err = snap.MatchJSONE("json_test", map[string]string{"key": "value"})
		assert.NoError(t, err)
	})
}

