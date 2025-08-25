package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TableManagementMockClient implements the DynamoDB API interface for testing table operations
type TableManagementMockClient struct {
	mock.Mock
}

func (m *TableManagementMockClient) CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.CreateTableOutput), args.Error(1)
}

func (m *TableManagementMockClient) DeleteTable(ctx context.Context, params *dynamodb.DeleteTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteTableOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DeleteTableOutput), args.Error(1)
}

func (m *TableManagementMockClient) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DescribeTableOutput), args.Error(1)
}

func (m *TableManagementMockClient) UpdateTable(ctx context.Context, params *dynamodb.UpdateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateTableOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateTableOutput), args.Error(1)
}

func (m *TableManagementMockClient) ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.ListTablesOutput), args.Error(1)
}

// TableManagerAPI interface for dependency injection
type TableManagerAPI interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DeleteTable(ctx context.Context, params *dynamodb.DeleteTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	UpdateTable(ctx context.Context, params *dynamodb.UpdateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateTableOutput, error)
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
}

// ==============================================================================
// PURE FUNCTIONS FOR TABLE INPUT BUILDING
// ==============================================================================

// buildCreateTableInput builds a CreateTableInput from parameters
func buildCreateTableInput(tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition, options CreateTableOptions) *dynamodb.CreateTableInput {
	if tableName == "" {
		return nil
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
			input.ProvisionedThroughput = &types.ProvisionedThroughput{
				ReadCapacityUnits:  &[]int64{5}[0],
				WriteCapacityUnits: &[]int64{5}[0],
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

	return input
}

// buildUpdateTableInput builds an UpdateTableInput from parameters
func buildUpdateTableInput(tableName string, options UpdateTableOptions) *dynamodb.UpdateTableInput {
	if tableName == "" {
		return nil
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

	return input
}

// ==============================================================================
// PURE FUNCTIONS FOR TABLE OPERATIONS WITH DEPENDENCY INJECTION
// ==============================================================================

// CreateTableWithClient creates a table using the provided client
func CreateTableWithClient(ctx context.Context, client TableManagerAPI, tableName string, keySchema []types.KeySchemaElement, attributeDefinitions []types.AttributeDefinition, options CreateTableOptions) (*TableDescription, error) {
	input := buildCreateTableInput(tableName, keySchema, attributeDefinitions, options)
	if input == nil {
		return nil, errors.New("invalid table name")
	}

	result, err := client.CreateTable(ctx, input)
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.TableDescription), nil
}

// DeleteTableWithClient deletes a table using the provided client
func DeleteTableWithClient(ctx context.Context, client TableManagerAPI, tableName string) error {
	if tableName == "" {
		return errors.New("invalid table name")
	}

	input := &dynamodb.DeleteTableInput{
		TableName: &tableName,
	}

	_, err := client.DeleteTable(ctx, input)
	return err
}

// DescribeTableWithClient describes a table using the provided client
func DescribeTableWithClient(ctx context.Context, client TableManagerAPI, tableName string) (*TableDescription, error) {
	if tableName == "" {
		return nil, errors.New("invalid table name")
	}

	input := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	result, err := client.DescribeTable(ctx, input)
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.Table), nil
}

// UpdateTableWithClient updates a table using the provided client
func UpdateTableWithClient(ctx context.Context, client TableManagerAPI, tableName string, options UpdateTableOptions) (*TableDescription, error) {
	input := buildUpdateTableInput(tableName, options)
	if input == nil {
		return nil, errors.New("invalid table name")
	}

	result, err := client.UpdateTable(ctx, input)
	if err != nil {
		return nil, err
	}

	return convertToTableDescription(result.TableDescription), nil
}

// ListTablesWithClient lists all tables using the provided client
func ListTablesWithClient(ctx context.Context, client TableManagerAPI) ([]string, error) {
	var allTables []string
	var exclusiveStartTableName *string

	for {
		input := &dynamodb.ListTablesInput{}
		if exclusiveStartTableName != nil {
			input.ExclusiveStartTableName = exclusiveStartTableName
		}

		result, err := client.ListTables(ctx, input)
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

// WaitForTableStatusWithClient waits for a table to reach the specified status
func WaitForTableStatusWithClient(ctx context.Context, client TableManagerAPI, tableName string, expectedStatus types.TableStatus, timeout time.Duration) error {
	if tableName == "" {
		return errors.New("invalid table name")
	}

	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Since(start) > timeout {
				return fmt.Errorf("timeout waiting for table %s to reach status %s", tableName, expectedStatus)
			}

			result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: &tableName,
			})
			if err != nil {
				// If table doesn't exist and we're waiting for it to be deleted, that's success
				if expectedStatus == types.TableStatusDeleting {
					var notFoundErr *types.ResourceNotFoundException
					if errors.As(err, &notFoundErr) {
						return nil
					}
				}
				return fmt.Errorf("failed to describe table %s: %w", tableName, err)
			}

			if result.Table.TableStatus == expectedStatus {
				return nil
			}

			// Check if table is in a terminal error state
			if result.Table.TableStatus == types.TableStatusInaccessibleEncryptionCredentials {
				return fmt.Errorf("table %s is in inaccessible encryption credentials state", tableName)
			}
		}
	}
}

// ==============================================================================
// TEST HELPER FUNCTIONS
// ==============================================================================

func createMockTableDescription(tableName string, status types.TableStatus) *types.TableDescription {
	creationTime := time.Now()
	return &types.TableDescription{
		TableName:        &tableName,
		TableStatus:      status,
		CreationDateTime: &creationTime,
		TableArn:         stringPtrForTableMgmt(fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s", tableName)),
		TableId:          stringPtrForTableMgmt("12345678-1234-1234-1234-123456789012"),
		ItemCount:        &[]int64{0}[0],
		TableSizeBytes:   &[]int64{0}[0],
		BillingModeSummary: &types.BillingModeSummary{
			BillingMode: types.BillingModePayPerRequest,
		},
		DeletionProtectionEnabled: &[]bool{false}[0],
	}
}

func createTableNotFoundError() *types.ResourceNotFoundException {
	return &types.ResourceNotFoundException{
		Message: stringPtrForTableMgmt("Table not found"),
	}
}

func createResourceInUseError() *types.ResourceInUseException {
	return &types.ResourceInUseException{
		Message: stringPtrForTableMgmt("Table already exists"),
	}
}

func createValidationErrorForTable() error {
	// Using generic error since ValidationException isn't properly exported
	return fmt.Errorf("ValidationException: Invalid table name")
}

func createLimitExceededErrorForTable() error {
	return fmt.Errorf("LimitExceededException: Too many tables")
}

// stringPtrForTableMgmt returns a pointer to a string
func stringPtrForTableMgmt(s string) *string {
	return &s
}

// ==============================================================================
// TEST PHASE: RED - FAILING TESTS FOR PURE FUNCTIONS
// ==============================================================================

func TestBuildCreateTableInput_WithValidBasicParameters_ShouldCreateCorrectInput(t *testing.T) {
	tableName := "test-table"
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{}

	result := buildCreateTableInput(tableName, keySchema, attributeDefinitions, options)

	require.NotNil(t, result)
	assert.Equal(t, tableName, *result.TableName)
	assert.Equal(t, keySchema, result.KeySchema)
	assert.Equal(t, attributeDefinitions, result.AttributeDefinitions)
	assert.Equal(t, types.BillingModeProvisioned, result.BillingMode)
	require.NotNil(t, result.ProvisionedThroughput)
	assert.Equal(t, int64(5), *result.ProvisionedThroughput.ReadCapacityUnits)
	assert.Equal(t, int64(5), *result.ProvisionedThroughput.WriteCapacityUnits)
}

func TestBuildCreateTableInput_WithPayPerRequestBilling_ShouldNotSetThroughput(t *testing.T) {
	tableName := "pay-per-request-table"
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	result := buildCreateTableInput(tableName, keySchema, attributeDefinitions, options)

	require.NotNil(t, result)
	assert.Equal(t, types.BillingModePayPerRequest, result.BillingMode)
	assert.Nil(t, result.ProvisionedThroughput)
}

func TestBuildCreateTableInput_WithCustomProvisionedThroughput_ShouldUseCustomValues(t *testing.T) {
	tableName := "custom-throughput-table"
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  &[]int64{10}[0],
			WriteCapacityUnits: &[]int64{15}[0],
		},
	}

	result := buildCreateTableInput(tableName, keySchema, attributeDefinitions, options)

	require.NotNil(t, result)
	require.NotNil(t, result.ProvisionedThroughput)
	assert.Equal(t, int64(10), *result.ProvisionedThroughput.ReadCapacityUnits)
	assert.Equal(t, int64(15), *result.ProvisionedThroughput.WriteCapacityUnits)
}

func TestBuildCreateTableInput_WithEmptyTableName_ShouldReturnNil(t *testing.T) {
	tableName := ""
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{}

	result := buildCreateTableInput(tableName, keySchema, attributeDefinitions, options)

	assert.Nil(t, result)
}

func TestBuildUpdateTableInput_WithValidBasicParameters_ShouldCreateCorrectInput(t *testing.T) {
	tableName := "update-test-table"
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	result := buildUpdateTableInput(tableName, options)

	require.NotNil(t, result)
	assert.Equal(t, tableName, *result.TableName)
	assert.Equal(t, types.BillingModePayPerRequest, result.BillingMode)
}

func TestBuildUpdateTableInput_WithEmptyTableName_ShouldReturnNil(t *testing.T) {
	tableName := ""
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	result := buildUpdateTableInput(tableName, options)

	assert.Nil(t, result)
}

// ==============================================================================
// TEST PHASE: RED - FAILING TESTS FOR MOCKED TABLE OPERATIONS
// ==============================================================================

func TestCreateTableWithClient_WithValidInput_ShouldCreateTableSuccessfully(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "test-table"
	
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	expectedOutput := &dynamodb.CreateTableOutput{
		TableDescription: createMockTableDescription(tableName, types.TableStatusCreating),
	}

	mockClient.On("CreateTable", ctx, mock.MatchedBy(func(input *dynamodb.CreateTableInput) bool {
		return *input.TableName == tableName &&
			input.BillingMode == types.BillingModePayPerRequest
	})).Return(expectedOutput, nil)

	result, err := CreateTableWithClient(ctx, mockClient, tableName, keySchema, attributeDefinitions, options)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tableName, result.TableName)
	assert.Equal(t, types.TableStatusCreating, result.TableStatus)
	mockClient.AssertExpectations(t)
}

func TestCreateTableWithClient_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := ""
	
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{}

	result, err := CreateTableWithClient(ctx, mockClient, tableName, keySchema, attributeDefinitions, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid table name")
	mockClient.AssertExpectations(t)
}

func TestCreateTableWithClient_WithExistingTable_ShouldReturnResourceInUseError(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "existing-table"
	
	keySchema := []types.KeySchemaElement{
		{AttributeName: stringPtrForTableMgmt("id"), KeyType: types.KeyTypeHash},
	}
	attributeDefinitions := []types.AttributeDefinition{
		{AttributeName: stringPtrForTableMgmt("id"), AttributeType: types.ScalarAttributeTypeS},
	}
	options := CreateTableOptions{}

	mockClient.On("CreateTable", ctx, mock.AnythingOfType("*dynamodb.CreateTableInput")).
		Return(nil, createResourceInUseError())

	result, err := CreateTableWithClient(ctx, mockClient, tableName, keySchema, attributeDefinitions, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	var resourceInUseErr *types.ResourceInUseException
	assert.True(t, errors.As(err, &resourceInUseErr))
	mockClient.AssertExpectations(t)
}

func TestDeleteTableWithClient_WithValidInput_ShouldDeleteTableSuccessfully(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "delete-test-table"

	expectedOutput := &dynamodb.DeleteTableOutput{
		TableDescription: createMockTableDescription(tableName, types.TableStatusDeleting),
	}

	mockClient.On("DeleteTable", ctx, mock.MatchedBy(func(input *dynamodb.DeleteTableInput) bool {
		return *input.TableName == tableName
	})).Return(expectedOutput, nil)

	err := DeleteTableWithClient(ctx, mockClient, tableName)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDeleteTableWithClient_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := ""

	err := DeleteTableWithClient(ctx, mockClient, tableName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
	mockClient.AssertExpectations(t)
}

func TestDescribeTableWithClient_WithValidInput_ShouldReturnTableInfo(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "describe-test-table"

	expectedOutput := &dynamodb.DescribeTableOutput{
		Table: createMockTableDescription(tableName, types.TableStatusActive),
	}

	mockClient.On("DescribeTable", ctx, mock.MatchedBy(func(input *dynamodb.DescribeTableInput) bool {
		return *input.TableName == tableName
	})).Return(expectedOutput, nil)

	result, err := DescribeTableWithClient(ctx, mockClient, tableName)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tableName, result.TableName)
	assert.Equal(t, types.TableStatusActive, result.TableStatus)
	mockClient.AssertExpectations(t)
}

func TestDescribeTableWithClient_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := ""

	result, err := DescribeTableWithClient(ctx, mockClient, tableName)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid table name")
	mockClient.AssertExpectations(t)
}

func TestUpdateTableWithClient_WithValidInput_ShouldUpdateTableSuccessfully(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "update-test-table"
	options := UpdateTableOptions{
		BillingMode: types.BillingModePayPerRequest,
	}

	expectedOutput := &dynamodb.UpdateTableOutput{
		TableDescription: createMockTableDescription(tableName, types.TableStatusUpdating),
	}

	mockClient.On("UpdateTable", ctx, mock.MatchedBy(func(input *dynamodb.UpdateTableInput) bool {
		return *input.TableName == tableName &&
			input.BillingMode == types.BillingModePayPerRequest
	})).Return(expectedOutput, nil)

	result, err := UpdateTableWithClient(ctx, mockClient, tableName, options)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tableName, result.TableName)
	assert.Equal(t, types.TableStatusUpdating, result.TableStatus)
	mockClient.AssertExpectations(t)
}

func TestListTablesWithClient_WithValidInput_ShouldReturnTableList(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()

	expectedOutput := &dynamodb.ListTablesOutput{
		TableNames: []string{"table-1", "table-2", "table-3"},
	}

	mockClient.On("ListTables", ctx, mock.MatchedBy(func(input *dynamodb.ListTablesInput) bool {
		return input.ExclusiveStartTableName == nil
	})).Return(expectedOutput, nil)

	result, err := ListTablesWithClient(ctx, mockClient)

	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Contains(t, result, "table-1")
	assert.Contains(t, result, "table-2")
	assert.Contains(t, result, "table-3")
	mockClient.AssertExpectations(t)
}

func TestWaitForTableStatusWithClient_WithValidInput_ShouldWaitSuccessfully(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "wait-test-table"
	expectedStatus := types.TableStatusActive
	timeout := 10 * time.Second

	// Mock table in creating state first, then active
	creatingTable := createMockTableDescription(tableName, types.TableStatusCreating)
	activeTable := createMockTableDescription(tableName, types.TableStatusActive)

	mockClient.On("DescribeTable", ctx, mock.MatchedBy(func(input *dynamodb.DescribeTableInput) bool {
		return *input.TableName == tableName
	})).Return(&dynamodb.DescribeTableOutput{Table: creatingTable}, nil).Once()

	mockClient.On("DescribeTable", ctx, mock.MatchedBy(func(input *dynamodb.DescribeTableInput) bool {
		return *input.TableName == tableName
	})).Return(&dynamodb.DescribeTableOutput{Table: activeTable}, nil).Once()

	start := time.Now()
	err := WaitForTableStatusWithClient(ctx, mockClient, tableName, expectedStatus, timeout)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, timeout) // Should complete before timeout
	mockClient.AssertExpectations(t)
}

func TestWaitForTableStatusWithClient_WithTimeout_ShouldTimeoutGracefully(t *testing.T) {
	mockClient := &TableManagementMockClient{}
	ctx := context.Background()
	tableName := "timeout-test-table"
	expectedStatus := types.TableStatusActive
	timeout := 100 * time.Millisecond // Very short timeout

	// Mock table always in creating state
	creatingTable := createMockTableDescription(tableName, types.TableStatusCreating)

	mockClient.On("DescribeTable", ctx, mock.MatchedBy(func(input *dynamodb.DescribeTableInput) bool {
		return *input.TableName == tableName
	})).Return(&dynamodb.DescribeTableOutput{Table: creatingTable}, nil).Maybe()

	start := time.Now()
	err := WaitForTableStatusWithClient(ctx, mockClient, tableName, expectedStatus, timeout)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.GreaterOrEqual(t, duration, timeout)
	mockClient.AssertExpectations(t)
}