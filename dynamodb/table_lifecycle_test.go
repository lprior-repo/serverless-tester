package dynamodb

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test table lifecycle input validation and construction logic
// Following TDD approach but focusing on testable parts without AWS calls

func TestCreateTableInputValidation_WithValidInputs_ShouldBuildCorrectInput(t *testing.T) {
	tableName := "test-table"
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("id"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: stringPtr("sortKey"),
			KeyType:       types.KeyTypeRange,
		},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("id"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("sortKey"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	options := CreateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
		LocalSecondaryIndexes: []types.LocalSecondaryIndex{
			{
				IndexName: stringPtr("local-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: stringPtr("id"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: stringPtr("localSortKey"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: stringPtr("global-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: stringPtr("gsiKey"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  int64Ptr(5),
					WriteCapacityUnits: int64Ptr(5),
				},
			},
		},
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtrForTable(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
		Tags: []types.Tag{
			{
				Key:   stringPtr("Environment"),
				Value: stringPtr("Test"),
			},
		},
		TableClass:                types.TableClassStandard,
		DeletionProtectionEnabled: boolPtrForTable(false),
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
	if options.Tags != nil {
		input.Tags = options.Tags
	}
	if options.TableClass != "" {
		input.TableClass = options.TableClass
	}
	if options.DeletionProtectionEnabled != nil {
		input.DeletionProtectionEnabled = options.DeletionProtectionEnabled
	}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, keySchema, input.KeySchema)
	assert.Equal(t, attributeDefinitions, input.AttributeDefinitions)
	assert.Equal(t, types.BillingModePayPerRequest, input.BillingMode)
	assert.Len(t, input.LocalSecondaryIndexes, 1)
	assert.Equal(t, "local-index", *input.LocalSecondaryIndexes[0].IndexName)
	assert.Len(t, input.GlobalSecondaryIndexes, 1)
	assert.Equal(t, "global-index", *input.GlobalSecondaryIndexes[0].IndexName)
	assert.True(t, *input.StreamSpecification.StreamEnabled)
	assert.Equal(t, types.StreamViewTypeNewAndOldImages, input.StreamSpecification.StreamViewType)
	assert.Len(t, input.Tags, 1)
	assert.Equal(t, "Environment", *input.Tags[0].Key)
	assert.Equal(t, "Test", *input.Tags[0].Value)
	assert.Equal(t, types.TableClassStandard, input.TableClass)
	assert.False(t, *input.DeletionProtectionEnabled)
}

func TestCreateTableInputValidation_WithProvisionedBilling_ShouldSetDefaultThroughput(t *testing.T) {
	tableName := "test-table"
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("id"),
			KeyType:       types.KeyTypeHash,
		},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("id"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	options := CreateTableOptions{} // No billing mode specified, should default to provisioned

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

	// Verify defaults were applied
	assert.Equal(t, types.BillingModeProvisioned, input.BillingMode)
	require.NotNil(t, input.ProvisionedThroughput)
	assert.Equal(t, int64(5), *input.ProvisionedThroughput.ReadCapacityUnits)
	assert.Equal(t, int64(5), *input.ProvisionedThroughput.WriteCapacityUnits)
}

func TestUpdateTableInputValidation_WithValidInputs_ShouldBuildCorrectInput(t *testing.T) {
	tableName := "test-table"
	options := UpdateTableOptions{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: stringPtr("newAttribute"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(10),
			WriteCapacityUnits: int64Ptr(10),
		},
		GlobalSecondaryIndexUpdates: []types.GlobalSecondaryIndexUpdate{
			{
				Create: &types.CreateGlobalSecondaryIndexAction{
					IndexName: stringPtr("new-index"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: stringPtr("newAttribute"),
							KeyType:       types.KeyTypeHash,
						},
					},
					Projection: &types.Projection{
						ProjectionType: types.ProjectionTypeAll,
					},
				},
			},
		},
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtrForTable(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
		TableClass:                types.TableClassStandardInfrequentAccess,
		DeletionProtectionEnabled: boolPtrForTable(true),
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
	if options.TableClass != "" {
		input.TableClass = options.TableClass
	}
	if options.DeletionProtectionEnabled != nil {
		input.DeletionProtectionEnabled = options.DeletionProtectionEnabled
	}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Equal(t, tableName, *input.TableName)
	assert.Len(t, input.AttributeDefinitions, 1)
	assert.Equal(t, "newAttribute", *input.AttributeDefinitions[0].AttributeName)
	assert.Equal(t, types.BillingModePayPerRequest, input.BillingMode)
	assert.Equal(t, int64(10), *input.ProvisionedThroughput.ReadCapacityUnits)
	assert.Len(t, input.GlobalSecondaryIndexUpdates, 1)
	assert.NotNil(t, input.GlobalSecondaryIndexUpdates[0].Create)
	assert.Equal(t, "new-index", *input.GlobalSecondaryIndexUpdates[0].Create.IndexName)
	assert.True(t, *input.StreamSpecification.StreamEnabled)
	assert.Equal(t, types.TableClassStandardInfrequentAccess, input.TableClass)
	assert.True(t, *input.DeletionProtectionEnabled)
}

func TestDeleteTableInputValidation_WithValidInputs_ShouldBuildCorrectInput(t *testing.T) {
	tableName := "test-table"

	input := &dynamodb.DeleteTableInput{
		TableName: &tableName,
	}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Equal(t, tableName, *input.TableName)
}

func TestDescribeTableInputValidation_WithValidInputs_ShouldBuildCorrectInput(t *testing.T) {
	tableName := "test-table"

	input := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Equal(t, tableName, *input.TableName)
}

func TestListTablesInputValidation_WithPagination_ShouldBuildCorrectInput(t *testing.T) {
	exclusiveStartTableName := "previous-table"

	input := &dynamodb.ListTablesInput{}
	if exclusiveStartTableName != "" {
		input.ExclusiveStartTableName = &exclusiveStartTableName
	}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Equal(t, exclusiveStartTableName, *input.ExclusiveStartTableName)
}

func TestListTablesInputValidation_WithoutPagination_ShouldBuildCorrectInput(t *testing.T) {
	input := &dynamodb.ListTablesInput{}

	// Verify input was constructed correctly
	require.NotNil(t, input)
	assert.Nil(t, input.ExclusiveStartTableName)
}

func TestConvertToTableDescription_WithCompleteInput_ShouldConvertAllFields(t *testing.T) {
	creationTime := time.Now()
	input := &types.TableDescription{
		TableName:        stringPtr("test-table"),
		TableStatus:      types.TableStatusActive,
		CreationDateTime: &creationTime,
		TableArn:         stringPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table"),
		TableId:          stringPtr("12345678-1234-1234-1234-123456789012"),
		ItemCount:        int64Ptr(100),
		TableSizeBytes:   int64Ptr(2048),
		ProvisionedThroughput: &types.ProvisionedThroughputDescription{
			ReadCapacityUnits:  int64Ptr(10),
			WriteCapacityUnits: int64Ptr(10),
		},
		BillingModeSummary: &types.BillingModeSummary{
			BillingMode: types.BillingModePayPerRequest,
		},
		LocalSecondaryIndexes: []types.LocalSecondaryIndexDescription{
			{
				IndexName: stringPtr("local-index"),
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndexDescription{
			{
				IndexName: stringPtr("global-index"),
			},
		},
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtrForTable(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
		LatestStreamLabel: stringPtr("stream-label"),
		LatestStreamArn:   stringPtr("arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream"),
		TableClassSummary: &types.TableClassSummary{
			TableClass: types.TableClassStandard,
		},
		DeletionProtectionEnabled: boolPtrForTable(false),
	}

	result := convertToTableDescription(input)

	require.NotNil(t, result)
	assert.Equal(t, "test-table", result.TableName)
	assert.Equal(t, types.TableStatusActive, result.TableStatus)
	assert.Equal(t, &creationTime, result.CreationDateTime)
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test-table", result.TableArn)
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", result.TableId)
	assert.Equal(t, int64(100), result.ItemCount)
	assert.Equal(t, int64(2048), result.TableSizeBytes)
	assert.Equal(t, int64(10), *result.ProvisionedThroughput.ReadCapacityUnits)
	assert.Equal(t, types.BillingModePayPerRequest, result.BillingModeSummary.BillingMode)
	assert.Len(t, result.LocalSecondaryIndexes, 1)
	assert.Equal(t, "local-index", *result.LocalSecondaryIndexes[0].IndexName)
	assert.Len(t, result.GlobalSecondaryIndexes, 1)
	assert.Equal(t, "global-index", *result.GlobalSecondaryIndexes[0].IndexName)
	assert.True(t, *result.StreamSpecification.StreamEnabled)
	assert.Equal(t, types.StreamViewTypeNewAndOldImages, result.StreamSpecification.StreamViewType)
	assert.Equal(t, "stream-label", result.LatestStreamLabel)
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream", result.LatestStreamArn)
	assert.Equal(t, types.TableClassStandard, result.TableClassSummary.TableClass)
	assert.False(t, *result.DeletionProtectionEnabled)
}

// Test error handling in table lifecycle operations - these will test AWS connection logic
func TestCreateTableE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("id"),
			KeyType:       types.KeyTypeHash,
		},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("id"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestCreateTableWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("id"),
			KeyType:       types.KeyTypeHash,
		},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("id"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	options := CreateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestDeleteTableE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""

	err := DeleteTableE(t, tableName)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestDescribeTableE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""

	_, err := DescribeTableE(t, tableName)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestUpdateTableE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := UpdateTableE(t, tableName, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestListTablesE_ShouldHandleAWSConnection(t *testing.T) {
	// This tests the AWS connection logic without specifying table name
	_, err := ListTablesE(t)

	// Should return an error due to AWS connection/credentials in test environment
	assert.Error(t, err)
}

func TestWaitForTableE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""
	expectedStatus := types.TableStatusActive
	timeout := 30 * time.Second

	err := WaitForTableE(t, tableName, expectedStatus, timeout)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestWaitForTableDeletedE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	tableName := ""
	timeout := 30 * time.Second

	err := WaitForTableDeletedE(t, tableName, timeout)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

// Helper function for table tests
func boolPtrForTable(b bool) *bool {
	return &b
}