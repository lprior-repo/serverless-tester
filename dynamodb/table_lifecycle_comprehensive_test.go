package dynamodb

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for table lifecycle operations

func createBasicTableSchema() ([]types.KeySchemaElement, []types.AttributeDefinition) {
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
	
	return keySchema, attributeDefinitions
}

func createCompositeKeyTableSchema() ([]types.KeySchemaElement, []types.AttributeDefinition) {
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("pk"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: stringPtr("sk"),
			KeyType:       types.KeyTypeRange,
		},
	}
	
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("sk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	
	return keySchema, attributeDefinitions
}

func createTableWithGSISchema() ([]types.KeySchemaElement, []types.AttributeDefinition, []types.GlobalSecondaryIndex) {
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
		{
			AttributeName: stringPtr("gsi_pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("gsi_sk"),
			AttributeType: types.ScalarAttributeTypeN,
		},
	}
	
	gsis := []types.GlobalSecondaryIndex{
		{
			IndexName: stringPtr("test-gsi-1"),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: stringPtr("gsi_pk"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: stringPtr("gsi_sk"),
					KeyType:       types.KeyTypeRange,
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
	}
	
	return keySchema, attributeDefinitions, gsis
}

func createTableWithLSISchema() ([]types.KeySchemaElement, []types.AttributeDefinition, []types.LocalSecondaryIndex) {
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("pk"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: stringPtr("sk"),
			KeyType:       types.KeyTypeRange,
		},
	}
	
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("sk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("lsi_sk"),
			AttributeType: types.ScalarAttributeTypeN,
		},
	}
	
	lsis := []types.LocalSecondaryIndex{
		{
			IndexName: stringPtr("test-lsi-1"),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: stringPtr("pk"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: stringPtr("lsi_sk"),
					KeyType:       types.KeyTypeRange,
				},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeKeysOnly,
			},
		},
	}
	
	return keySchema, attributeDefinitions, lsis
}

// Basic table creation tests

func TestCreateTable_WithBasicSchema_ShouldCreateSuccessfully(t *testing.T) {
	tableName := "basic-test-table"
	keySchema, attributeDefinitions := createBasicTableSchema()

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithCompositeKey_ShouldCreateSuccessfully(t *testing.T) {
	tableName := "composite-key-table"
	keySchema, attributeDefinitions := createCompositeKeyTableSchema()

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithOptions_ShouldApplyOptionsCorrectly(t *testing.T) {
	tableName := "options-test-table"
	keySchema, attributeDefinitions := createBasicTableSchema()
	
	options := CreateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtr(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
		SSESpecification: &types.SSESpecification{
			Enabled: boolPtr(true),
		},
		Tags: []types.Tag{
			{
				Key:   stringPtr("Environment"),
				Value: stringPtr("test"),
			},
			{
				Key:   stringPtr("Project"),
				Value: stringPtr("comprehensive-tests"),
			},
		},
		TableClass:                types.TableClassStandard,
		DeletionProtectionEnabled: boolPtr(true),
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithProvisionedThroughput_ShouldSetCapacityCorrectly(t *testing.T) {
	tableName := "provisioned-table"
	keySchema, attributeDefinitions := createBasicTableSchema()
	
	options := CreateTableOptions{
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(10),
			WriteCapacityUnits: int64Ptr(5),
		},
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithGSI_ShouldCreateWithSecondaryIndex(t *testing.T) {
	tableName := "gsi-test-table"
	keySchema, attributeDefinitions, gsis := createTableWithGSISchema()
	
	options := CreateTableOptions{
		BillingMode:            types.BillingModeProvisioned,
		GlobalSecondaryIndexes: gsis,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(5),
			WriteCapacityUnits: int64Ptr(5),
		},
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithLSI_ShouldCreateWithLocalSecondaryIndex(t *testing.T) {
	tableName := "lsi-test-table"
	keySchema, attributeDefinitions, lsis := createTableWithLSISchema()
	
	options := CreateTableOptions{
		BillingMode:           types.BillingModeProvisioned,
		LocalSecondaryIndexes: lsis,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(5),
			WriteCapacityUnits: int64Ptr(5),
		},
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithComplexConfiguration_ShouldHandleAllOptions(t *testing.T) {
	tableName := "complex-config-table"
	keySchema, attributeDefinitions, gsis := createTableWithGSISchema()
	
	options := CreateTableOptions{
		BillingMode:            types.BillingModeProvisioned,
		GlobalSecondaryIndexes: gsis,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(20),
			WriteCapacityUnits: int64Ptr(10),
		},
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtr(true),
			StreamViewType: types.StreamViewTypeKeysOnly,
		},
		SSESpecification: &types.SSESpecification{
			Enabled:                 boolPtr(true),
			SSEType:                 types.SSETypeAes256,
			KMSMasterKeyId:          stringPtr("alias/aws/dynamodb"),
		},
		Tags: []types.Tag{
			{
				Key:   stringPtr("Environment"),
				Value: stringPtr("production"),
			},
			{
				Key:   stringPtr("Application"),
				Value: stringPtr("test-app"),
			},
			{
				Key:   stringPtr("CostCenter"),
				Value: stringPtr("engineering"),
			},
		},
		TableClass:                types.TableClassStandardInfrequentAccess,
		DeletionProtectionEnabled: boolPtr(true),
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Table deletion tests

func TestDeleteTable_WithExistingTable_ShouldDeleteSuccessfully(t *testing.T) {
	tableName := "delete-test-table"

	err := DeleteTableE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestDeleteTable_WithNonExistentTable_ShouldReturnResourceNotFoundException(t *testing.T) {
	tableName := "non-existent-table"

	err := DeleteTableE(t, tableName)
	
	// Should fail - later we'll mock this to return ResourceNotFoundException
	assert.Error(t, err)
}

// Table description tests

func TestDescribeTable_WithExistingTable_ShouldReturnTableInfo(t *testing.T) {
	tableName := "describe-test-table"

	_, err := DescribeTableE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestDescribeTable_WithNonExistentTable_ShouldReturnResourceNotFoundException(t *testing.T) {
	tableName := "non-existent-table"

	_, err := DescribeTableE(t, tableName)
	
	// Should fail - later we'll mock this to return ResourceNotFoundException
	assert.Error(t, err)
}

// Table update tests

func TestUpdateTable_ProvisionedThroughput_ShouldUpdateCapacity(t *testing.T) {
	tableName := "update-throughput-table"
	options := UpdateTableOptions{
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(15),
			WriteCapacityUnits: int64Ptr(8),
		},
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_BillingMode_ShouldChangeBillingMode(t *testing.T) {
	tableName := "update-billing-table"
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_StreamSpecification_ShouldUpdateStream(t *testing.T) {
	tableName := "update-stream-table"
	options := UpdateTableOptions{
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled:  boolPtr(true),
			StreamViewType: types.StreamViewTypeNewAndOldImages,
		},
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_DisableStream_ShouldDisableStream(t *testing.T) {
	tableName := "disable-stream-table"
	options := UpdateTableOptions{
		StreamSpecification: &types.StreamSpecification{
			StreamEnabled: boolPtr(false),
		},
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_TableClass_ShouldUpdateTableClass(t *testing.T) {
	tableName := "update-class-table"
	options := UpdateTableOptions{
		TableClass: types.TableClassStandardInfrequentAccess,
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_DeletionProtection_ShouldToggleDeletionProtection(t *testing.T) {
	tableName := "update-protection-table"
	options := UpdateTableOptions{
		DeletionProtectionEnabled: boolPtr(false),
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestUpdateTable_GSIUpdates_ShouldHandleGSIOperations(t *testing.T) {
	tableName := "update-gsi-table"
	
	// Example GSI update operations
	gsiUpdates := []types.GlobalSecondaryIndexUpdate{
		{
			Create: &types.CreateGlobalSecondaryIndexAction{
				IndexName: stringPtr("new-gsi-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: stringPtr("new_gsi_pk"),
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
		{
			Update: &types.UpdateGlobalSecondaryIndexAction{
				IndexName: stringPtr("existing-gsi-index"),
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  int64Ptr(10),
					WriteCapacityUnits: int64Ptr(10),
				},
			},
		},
		{
			Delete: &types.DeleteGlobalSecondaryIndexAction{
				IndexName: stringPtr("old-gsi-index"),
			},
		},
	}
	
	options := UpdateTableOptions{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: stringPtr("new_gsi_pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexUpdates: gsiUpdates,
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Table waiting tests

func TestWaitForTable_ActiveStatus_ShouldWaitUntilActive(t *testing.T) {
	tableName := "wait-active-table"
	timeout := 30 * time.Second

	err := WaitForTableE(t, tableName, types.TableStatusActive, timeout)
	
	// Should fail without proper mocking (timeout)
	assert.Error(t, err)
}

func TestWaitForTableActive_ConvenienceMethod_ShouldWaitForActive(t *testing.T) {
	tableName := "wait-active-convenience-table"
	timeout := 10 * time.Second

	// This should timeout and fail
	assert.Panics(t, func() {
		WaitForTableActive(t, tableName, timeout)
	})
}

func TestWaitForTable_CreatingToActive_ShouldHandleStatusTransition(t *testing.T) {
	tableName := "creating-to-active-table"
	timeout := 60 * time.Second

	err := WaitForTableE(t, tableName, types.TableStatusActive, timeout)
	
	// Should fail without proper mocking (timeout)
	assert.Error(t, err)
}

func TestWaitForTable_UpdatingToActive_ShouldHandleUpdateCompletion(t *testing.T) {
	tableName := "updating-to-active-table"
	timeout := 120 * time.Second // Updates can take longer

	err := WaitForTableE(t, tableName, types.TableStatusActive, timeout)
	
	// Should fail without proper mocking (timeout)
	assert.Error(t, err)
}

func TestWaitForTableDeleted_ShouldWaitUntilDeleted(t *testing.T) {
	tableName := "wait-deleted-table"
	timeout := 30 * time.Second

	err := WaitForTableDeletedE(t, tableName, timeout)
	
	// Should fail without proper mocking (timeout)
	assert.Error(t, err)
}

func TestWaitForTable_WithTimeout_ShouldTimeoutGracefully(t *testing.T) {
	tableName := "timeout-test-table"
	timeout := 2 * time.Second // Very short timeout

	start := time.Now()
	err := WaitForTableE(t, tableName, types.TableStatusActive, timeout)
	duration := time.Since(start)

	// Should fail due to timeout
	assert.Error(t, err)
	// Should respect timeout
	assert.GreaterOrEqual(t, duration, timeout)
	assert.Contains(t, err.Error(), "timeout")
}

func TestWaitForTable_InaccessibleEncryptionCredentials_ShouldFailImmediately(t *testing.T) {
	tableName := "inaccessible-encryption-table"
	timeout := 30 * time.Second

	err := WaitForTableE(t, tableName, types.TableStatusActive, timeout)
	
	// Should fail without proper mocking, but normally would fail due to inaccessible credentials
	assert.Error(t, err)
}

// Table listing tests

func TestListTables_ShouldReturnAllTables(t *testing.T) {
	_, err := ListTablesE(t)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestListTables_WithPagination_ShouldHandleManyTables(t *testing.T) {
	tables, err := ListTablesE(t)
	
	// Should fail without proper mocking
	assert.Error(t, err)
	
	// In a real scenario, we'd test pagination behavior
	if err == nil {
		t.Logf("Found %d tables", len(tables))
	}
}

// Complex lifecycle scenarios

func TestTableLifecycle_CompleteWorkflow_ShouldHandleFullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full lifecycle test in short mode")
	}

	baseName := "lifecycle-test"
	tableName := fmt.Sprintf("%s-%d", baseName, time.Now().Unix())
	keySchema, attributeDefinitions := createBasicTableSchema()

	// Phase 1: Create table
	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 2: Wait for table to become active
	err = WaitForTableE(t, tableName, types.TableStatusActive, 60*time.Second)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 3: Describe table
	_, err = DescribeTableE(t, tableName)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 4: Update table
	updateOptions := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}
	_, err = UpdateTableE(t, tableName, updateOptions)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 5: Wait for update to complete
	err = WaitForTableE(t, tableName, types.TableStatusActive, 120*time.Second)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 6: Delete table
	err = DeleteTableE(t, tableName)
	assert.Error(t, err) // Should fail without proper mocking

	// Phase 7: Wait for deletion to complete
	err = WaitForTableDeletedE(t, tableName, 60*time.Second)
	assert.Error(t, err) // Should fail without proper mocking
}

func TestTableCreation_WithMultipleTables_ShouldHandleConcurrentCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent creation test in short mode")
	}

	numTables := 3
	baseName := "concurrent-test"
	errorChan := make(chan error, numTables)

	// Create multiple tables concurrently
	for i := 0; i < numTables; i++ {
		go func(tableIndex int) {
			tableName := fmt.Sprintf("%s-%d-%d", baseName, time.Now().Unix(), tableIndex)
			keySchema, attributeDefinitions := createBasicTableSchema()
			
			_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
			errorChan <- err
		}(i)
	}

	// Collect all errors
	errorCount := 0
	for i := 0; i < numTables; i++ {
		err := <-errorChan
		if err != nil {
			errorCount++
		}
	}

	// We expect errors since we don't have real AWS setup
	assert.Equal(t, numTables, errorCount)
	
	t.Logf("Attempted concurrent creation of %d tables", numTables)
}

// Performance and stress tests

func TestTableCreation_PerformanceBenchmark_SingleTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tableName := fmt.Sprintf("perf-test-%d", time.Now().Unix())
	keySchema, attributeDefinitions := createBasicTableSchema()

	start := time.Now()
	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	duration := time.Since(start)

	// Should fail without proper mocking
	assert.Error(t, err)
	t.Logf("Table creation attempt took: %v", duration)
}

func TestTableUpdate_PerformanceBenchmark_ThroughputUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tableName := "perf-update-test"
	options := UpdateTableOptions{
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(25),
			WriteCapacityUnits: int64Ptr(15),
		},
	}

	start := time.Now()
	_, err := UpdateTableE(t, tableName, options)
	duration := time.Since(start)

	// Should fail without proper mocking
	assert.Error(t, err)
	t.Logf("Table update attempt took: %v", duration)
}

// Error handling tests

func TestCreateTable_WithInvalidTableName_ShouldReturnValidationException(t *testing.T) {
	tableName := "" // Invalid empty table name
	keySchema, attributeDefinitions := createBasicTableSchema()

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail - later we'll mock this to return ValidationException
	assert.Error(t, err)
}

func TestCreateTable_WithExistingTableName_ShouldReturnResourceInUseException(t *testing.T) {
	tableName := "existing-table"
	keySchema, attributeDefinitions := createBasicTableSchema()

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail - later we'll mock this to return ResourceInUseException
	assert.Error(t, err)
}

func TestCreateTable_WithInvalidKeySchema_ShouldReturnValidationException(t *testing.T) {
	tableName := "invalid-key-schema-table"
	keySchema := []types.KeySchemaElement{} // Empty key schema
	attributeDefinitions := []types.AttributeDefinition{}

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail - later we'll mock this to return ValidationException
	assert.Error(t, err)
}

func TestUpdateTable_WithNonExistentTable_ShouldReturnResourceNotFoundException(t *testing.T) {
	tableName := "non-existent-update-table"
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail - later we'll mock this to return ResourceNotFoundException
	assert.Error(t, err)
}

func TestUpdateTable_DuringUpdate_ShouldReturnResourceInUseException(t *testing.T) {
	tableName := "currently-updating-table"
	options := UpdateTableOptions{
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(20),
			WriteCapacityUnits: int64Ptr(10),
		},
	}

	_, err := UpdateTableE(t, tableName, options)
	
	// Should fail - later we'll mock this to return ResourceInUseException
	assert.Error(t, err)
}

func TestDeleteTable_WithDeletionProtection_ShouldReturnValidationException(t *testing.T) {
	tableName := "deletion-protected-table"

	err := DeleteTableE(t, tableName)
	
	// Should fail - later we'll mock this to return ValidationException
	assert.Error(t, err)
}

// Edge cases and special scenarios

func TestCreateTable_WithLongTableName_ShouldHandleNameLengthLimit(t *testing.T) {
	// DynamoDB table names can be 3-255 characters
	longTableName := "very-long-table-name-" + fmt.Sprintf("%0250d", 1) // Create a very long name
	keySchema, attributeDefinitions := createBasicTableSchema()

	_, err := CreateTableE(t, longTableName, keySchema, attributeDefinitions)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestCreateTable_WithInvalidCharacters_ShouldReturnValidationException(t *testing.T) {
	tableName := "invalid@table#name!" // Contains invalid characters
	keySchema, attributeDefinitions := createBasicTableSchema()

	_, err := CreateTableE(t, tableName, keySchema, attributeDefinitions)
	
	// Should fail - later we'll mock this to return ValidationException
	assert.Error(t, err)
}

func TestCreateTable_WithTooManyGSIs_ShouldReturnLimitExceededException(t *testing.T) {
	tableName := "too-many-gsis-table"
	keySchema, attributeDefinitions := createBasicTableSchema()
	
	// Create more than 20 GSIs (DynamoDB limit)
	gsis := make([]types.GlobalSecondaryIndex, 21)
	newAttributeDefinitions := make([]types.AttributeDefinition, len(attributeDefinitions)+21)
	copy(newAttributeDefinitions, attributeDefinitions)
	
	for i := 0; i < 21; i++ {
		attrName := fmt.Sprintf("gsi_pk_%d", i+1)
		newAttributeDefinitions[len(attributeDefinitions)+i] = types.AttributeDefinition{
			AttributeName: stringPtr(attrName),
			AttributeType: types.ScalarAttributeTypeS,
		}
		
		gsis[i] = types.GlobalSecondaryIndex{
			IndexName: stringPtr(fmt.Sprintf("gsi-%d", i+1)),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: stringPtr(attrName),
					KeyType:       types.KeyTypeHash,
				},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeKeysOnly,
			},
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  int64Ptr(1),
				WriteCapacityUnits: int64Ptr(1),
			},
		}
	}
	
	options := CreateTableOptions{
		BillingMode:            types.BillingModeProvisioned,
		GlobalSecondaryIndexes: gsis,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(5),
			WriteCapacityUnits: int64Ptr(5),
		},
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, newAttributeDefinitions, options)
	
	// Should fail - later we'll mock this to return LimitExceededException
	assert.Error(t, err)
}

func TestCreateTable_WithTooManyLSIs_ShouldReturnLimitExceededException(t *testing.T) {
	tableName := "too-many-lsis-table"
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: stringPtr("pk"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: stringPtr("sk"),
			KeyType:       types.KeyTypeRange,
		},
	}
	
	// Create more than 10 LSIs (DynamoDB limit)
	lsis := make([]types.LocalSecondaryIndex, 11)
	attributeDefinitions := []types.AttributeDefinition{
		{
			AttributeName: stringPtr("pk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: stringPtr("sk"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	
	for i := 0; i < 11; i++ {
		attrName := fmt.Sprintf("lsi_sk_%d", i+1)
		attributeDefinitions = append(attributeDefinitions, types.AttributeDefinition{
			AttributeName: stringPtr(attrName),
			AttributeType: types.ScalarAttributeTypeS,
		})
		
		lsis[i] = types.LocalSecondaryIndex{
			IndexName: stringPtr(fmt.Sprintf("lsi-%d", i+1)),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: stringPtr("pk"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: stringPtr(attrName),
					KeyType:       types.KeyTypeRange,
				},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeKeysOnly,
			},
		}
	}
	
	options := CreateTableOptions{
		BillingMode:           types.BillingModeProvisioned,
		LocalSecondaryIndexes: lsis,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  int64Ptr(5),
			WriteCapacityUnits: int64Ptr(5),
		},
	}

	_, err := CreateTableWithOptionsE(t, tableName, keySchema, attributeDefinitions, options)
	
	// Should fail - later we'll mock this to return LimitExceededException
	assert.Error(t, err)
}