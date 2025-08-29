package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
	
	vasdeferns "vasdeference"
)

// TestContext alias for consistent API across all AWS service packages
type TestContext = vasdeferns.TestContext

// CreateTableOptions contains optional parameters for CreateTable operations
type CreateTableOptions struct {
	BillingMode                   types.BillingMode
	ProvisionedThroughput        *types.ProvisionedThroughput
	LocalSecondaryIndexes        []types.LocalSecondaryIndex
	GlobalSecondaryIndexes       []types.GlobalSecondaryIndex
	StreamSpecification          *types.StreamSpecification
	SSESpecification             *types.SSESpecification
	Tags                         []types.Tag
	TableClass                   types.TableClass
	DeletionProtectionEnabled    *bool
}

// UpdateTableOptions contains optional parameters for UpdateTable operations
type UpdateTableOptions struct {
	AttributeDefinitions         []types.AttributeDefinition
	BillingMode                  types.BillingMode
	ProvisionedThroughput        *types.ProvisionedThroughput
	GlobalSecondaryIndexUpdates  []types.GlobalSecondaryIndexUpdate
	StreamSpecification          *types.StreamSpecification
	SSESpecification             *types.SSESpecification
	ReplicaUpdates               []types.ReplicationGroupUpdate
	TableClass                   types.TableClass
	DeletionProtectionEnabled    *bool
}

// TableDescription represents table information
type TableDescription struct {
	TableName              string
	TableStatus            types.TableStatus
	CreationDateTime       *time.Time
	ProvisionedThroughput  *types.ProvisionedThroughputDescription
	TableSizeBytes         int64
	ItemCount              int64
	TableArn               string
	TableId                string
	BillingModeSummary     *types.BillingModeSummary
	LocalSecondaryIndexes  []types.LocalSecondaryIndexDescription
	GlobalSecondaryIndexes []types.GlobalSecondaryIndexDescription
	StreamSpecification    *types.StreamSpecification
	LatestStreamLabel      string
	LatestStreamArn        string
	SSEDescription         *types.SSEDescription
	ArchivalSummary        *types.ArchivalSummary
	TableClassSummary      *types.TableClassSummary
	DeletionProtectionEnabled *bool
}

// CreateTable creates a new DynamoDB table and fails the test if an error occurs
func CreateTable(ctx *TestContext, tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition) *TableDescription {
	result, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	require.NoError(ctx.T, err)
	return result
}

// CreateTableE creates a new DynamoDB table and returns the table description and any error
func CreateTableE(ctx *TestContext, tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition) (*TableDescription, error) {
	return CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, CreateTableOptions{})
}

// CreateTableWithOptions creates a new DynamoDB table with options and fails the test if an error occurs
func CreateTableWithOptions(ctx *TestContext, tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition, options CreateTableOptions) *TableDescription {
	result, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	require.NoError(ctx.T, err)
	return result
}

// CreateTableWithOptionsE creates a new DynamoDB table with options and returns the table description and any error
func CreateTableWithOptionsE(ctx *TestContext, tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition, options CreateTableOptions) (*TableDescription, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.CreateTableInput{
		TableName:            &tableName,
		KeySchema:           keySchema,
		AttributeDefinitions: attributeDefinitions,
	}

	// Set billing mode and provisioned throughput
	if options.BillingMode != "" {
		input.BillingMode = options.BillingMode
	} else {
		input.BillingMode = types.BillingModeProvisioned
	}

	if input.BillingMode == types.BillingModeProvisioned {
		if options.ProvisionedThroughput != nil {
			input.ProvisionedThroughput = options.ProvisionedThroughput
		} else {
			// Default provisioned throughput
			input.ProvisionedThroughput = &types.ProvisionedThroughput{
				ReadCapacityUnits:  int64Ptr(5),
				WriteCapacityUnits: int64Ptr(5),
			}
		}
	}

	if options.LocalSecondaryIndexes != nil {
		input.LocalSecondaryIndexes = options.LocalSecondaryIndexes
	}
	if options.GlobalSecondaryIndexes != nil {
		input.GlobalSecondaryIndexes = options.GlobalSecondaryIndexes
	}
	if options.StreamSpecification != nil {
		input.StreamSpecification = options.StreamSpecification
	}
	if options.SSESpecification != nil {
		input.SSESpecification = options.SSESpecification
	}
	if options.Tags != nil {
		input.Tags = options.Tags
	}
	if options.TableClass != "" {
		input.TableClass = options.TableClass
	}
	if options.DeletionProtectionEnabled != nil {
		input.DeletionProtectionEnabled = options.DeletionProtectionEnabled
	}

	result, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.TableDescription), nil
}

// DeleteTable deletes a DynamoDB table and fails the test if an error occurs
func DeleteTable(ctx *TestContext, tableName string) {
	err := DeleteTableE(t, tableName)
	require.NoError(ctx.T, err)
}

// DeleteTableE deletes a DynamoDB table and returns any error
func DeleteTableE(ctx *TestContext, tableName string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: &tableName,
	})

	return err
}

// DescribeTable gets information about a DynamoDB table and fails the test if an error occurs
func DescribeTable(ctx *TestContext, tableName string) *TableDescription {
	result, err := DescribeTableE(t, tableName)
	require.NoError(ctx.T, err)
	return result
}

// DescribeTableE gets information about a DynamoDB table and returns the table description and any error
func DescribeTableE(ctx *TestContext, tableName string) (*TableDescription, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.Table), nil
}

// UpdateTable modifies a DynamoDB table and fails the test if an error occurs
func UpdateTable(ctx *TestContext, tableName string, options UpdateTableOptions) *TableDescription {
	result, err := UpdateTableE(t, tableName, options)
	require.NoError(ctx.T, err)
	return result
}

// UpdateTableE modifies a DynamoDB table and returns the table description and any error
func UpdateTableE(ctx *TestContext, tableName string, options UpdateTableOptions) (*TableDescription, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.UpdateTableInput{
		TableName: &tableName,
	}

	if options.AttributeDefinitions != nil {
		input.AttributeDefinitions = options.AttributeDefinitions
	}
	if options.BillingMode != "" {
		input.BillingMode = options.BillingMode
	}
	if options.ProvisionedThroughput != nil {
		input.ProvisionedThroughput = options.ProvisionedThroughput
	}
	if options.GlobalSecondaryIndexUpdates != nil {
		input.GlobalSecondaryIndexUpdates = options.GlobalSecondaryIndexUpdates
	}
	if options.StreamSpecification != nil {
		input.StreamSpecification = options.StreamSpecification
	}
	if options.SSESpecification != nil {
		input.SSESpecification = options.SSESpecification
	}
	if options.ReplicaUpdates != nil {
		input.ReplicaUpdates = options.ReplicaUpdates
	}
	if options.TableClass != "" {
		input.TableClass = options.TableClass
	}
	if options.DeletionProtectionEnabled != nil {
		input.DeletionProtectionEnabled = options.DeletionProtectionEnabled
	}

	result, err := client.UpdateTable(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.TableDescription), nil
}

// WaitForTable waits for a table to reach the specified status and fails the test if an error occurs
func WaitForTable(ctx *TestContext, tableName string, expectedStatus types.TableStatus, timeout time.Duration) {
	err := WaitForTableE(t, tableName, expectedStatus, timeout)
	require.NoError(ctx.T, err)
}

// WaitForTableE waits for a table to reach the specified status and returns any error
func WaitForTableE(ctx *TestContext, tableName string, expectedStatus types.TableStatus, timeout time.Duration) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	maxRetries := int(timeout.Seconds() / 2) // 2 second intervals
	description := fmt.Sprintf("Waiting for table %s to reach status %s", tableName, expectedStatus)
	
	retry.DoWithRetry(
		t,
		description,
		maxRetries,
		2*time.Second,
		func() (string, error) {
			result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
				TableName: &tableName,
			})
			if err != nil {
				// If table doesn't exist and we're waiting for it to be deleted, that's success
				if expectedStatus == types.TableStatusDeleting {
					var notFoundErr *types.ResourceNotFoundException
					if errors.As(err, &notFoundErr) {
						return "deleted", nil
					}
				}
				return "", fmt.Errorf("failed to describe table %s: %w", tableName, err)
			}

			if result.Table.TableStatus == expectedStatus {
				return "ready", nil
			}

			// Check if table is in a terminal error state
			if result.Table.TableStatus == types.TableStatusInaccessibleEncryptionCredentials {
				return "", fmt.Errorf("table %s is in inaccessible encryption credentials state", tableName)
			}

			return "", fmt.Errorf("table %s is in status %s, expected %s", tableName, result.Table.TableStatus, expectedStatus)
		},
	)
	
	return nil
}

// WaitForTableActive waits for a table to become active and fails the test if an error occurs
func WaitForTableActive(ctx *TestContext, tableName string, timeout time.Duration) {
	WaitForTable(ctx, tableName, types.TableStatusActive, timeout)
}

// WaitForTableDeleted waits for a table to be deleted and fails the test if an error occurs
func WaitForTableDeleted(ctx *TestContext, tableName string, timeout time.Duration) {
	err := WaitForTableDeletedE(t, tableName, timeout)
	require.NoError(ctx.T, err)
}

// WaitForTableDeletedE waits for a table to be deleted and returns any error
func WaitForTableDeletedE(ctx *TestContext, tableName string, timeout time.Duration) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	maxRetries := int(timeout.Seconds() / 2) // 2 second intervals
	description := fmt.Sprintf("Waiting for table %s to be deleted", tableName)
	
	retry.DoWithRetry(
		t,
		description,
		maxRetries,
		2*time.Second,
		func() (string, error) {
			_, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
				TableName: &tableName,
			})
			
			if err != nil {
				var notFoundErr *types.ResourceNotFoundException
				if errors.As(err, &notFoundErr) {
					return "deleted", nil // Table has been deleted
				}
				return "", fmt.Errorf("unexpected error while waiting for table %s to be deleted: %w", tableName, err)
			}

			return "", fmt.Errorf("table %s still exists", tableName)
		},
	)
	
	return nil
}

// ListTables lists all DynamoDB tables and fails the test if an error occurs
func ListTables(ctx *TestContext) []string {
	tables, err := ListTablesE(ctx)
	require.NoError(ctx.T, err)
	return tables
}

// ListTablesE lists all DynamoDB tables and returns the table names and any error
func ListTablesE(ctx *TestContext) ([]string, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	var allTables []string
	var exclusiveStartTableName *string

	for {
		input := &dynamodb.ListTablesInput{}
		if exclusiveStartTableName != nil {
			input.ExclusiveStartTableName = exclusiveStartTableName
		}

		result, err := client.ListTables(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		allTables = append(allTables, result.TableNames...)

		if result.LastEvaluatedTableName == nil {
			break
		}
		exclusiveStartTableName = result.LastEvaluatedTableName
	}

	return allTables, nil
}

// convertToTableDescription converts AWS SDK TableDescription to our custom type
func convertToTableDescription(table *types.TableDescription) *TableDescription {
	if table == nil {
		return nil
	}

	desc := &TableDescription{
		TableName:     ptrToString(table.TableName),
		TableStatus:   table.TableStatus,
		TableSizeBytes: ptrToInt64(table.TableSizeBytes),
		ItemCount:     ptrToInt64(table.ItemCount),
		TableArn:      ptrToString(table.TableArn),
		TableId:       ptrToString(table.TableId),
	}

	if table.CreationDateTime != nil {
		desc.CreationDateTime = table.CreationDateTime
	}
	if table.ProvisionedThroughput != nil {
		desc.ProvisionedThroughput = table.ProvisionedThroughput
	}
	if table.BillingModeSummary != nil {
		desc.BillingModeSummary = table.BillingModeSummary
	}
	if table.LocalSecondaryIndexes != nil {
		desc.LocalSecondaryIndexes = table.LocalSecondaryIndexes
	}
	if table.GlobalSecondaryIndexes != nil {
		desc.GlobalSecondaryIndexes = table.GlobalSecondaryIndexes
	}
	if table.StreamSpecification != nil {
		desc.StreamSpecification = table.StreamSpecification
	}
	if table.LatestStreamLabel != nil {
		desc.LatestStreamLabel = *table.LatestStreamLabel
	}
	if table.LatestStreamArn != nil {
		desc.LatestStreamArn = *table.LatestStreamArn
	}
	if table.SSEDescription != nil {
		desc.SSEDescription = table.SSEDescription
	}
	if table.ArchivalSummary != nil {
		desc.ArchivalSummary = table.ArchivalSummary
	}
	if table.TableClassSummary != nil {
		desc.TableClassSummary = table.TableClassSummary
	}
	if table.DeletionProtectionEnabled != nil {
		desc.DeletionProtectionEnabled = table.DeletionProtectionEnabled
	}

	return desc
}

// Helper functions for pointer conversions
func int64Ptr(i int64) *int64 {
	return &i
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}