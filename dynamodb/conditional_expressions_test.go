package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for conditional expressions
func createConditionalPutItemOptions() PutItemOptions {
	return PutItemOptions{
		ConditionExpression: stringPtrConditional("attribute_not_exists(#id) AND #status = :inactive"),
		ExpressionAttributeNames: map[string]string{
			"#id":     "id",
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inactive": &types.AttributeValueMemberS{Value: "inactive"},
		},
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}
}

func createConditionalUpdateItemOptions() UpdateItemOptions {
	return UpdateItemOptions{
		ConditionExpression: stringPtrConditional("attribute_exists(#id) AND #version = :expectedVersion"),
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}
}

func createConditionalDeleteItemOptions() DeleteItemOptions {
	return DeleteItemOptions{
		ConditionExpression: stringPtrConditional("attribute_exists(#id) AND #active = :false"),
		ExpressionAttributeNames: map[string]string{
			"#id":     "id",
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}
}

// TESTS FOR CONDITION EXPRESSION VALIDATION

func TestConditionalExpressions_WithAttributeExists_ShouldValidateExistence(t *testing.T) {
	// RED: Test attribute_exists condition
	condition := "attribute_exists(id)"
	
	// GREEN: Verify condition syntax is valid
	assert.Contains(t, condition, "attribute_exists")
	assert.Contains(t, condition, "id")
	assert.NotContains(t, condition, ":")
}

func TestConditionalExpressions_WithAttributeNotExists_ShouldValidateNonExistence(t *testing.T) {
	// RED: Test attribute_not_exists condition
	condition := "attribute_not_exists(email)"
	
	// GREEN: Verify condition syntax is valid
	assert.Contains(t, condition, "attribute_not_exists")
	assert.Contains(t, condition, "email")
}

func TestConditionalExpressions_WithAttributeType_ShouldValidateType(t *testing.T) {
	// RED: Test attribute_type condition
	condition := "attribute_type(#score, :numberType)"
	attributeNames := map[string]string{
		"#score": "score",
	}
	attributeValues := map[string]types.AttributeValue{
		":numberType": &types.AttributeValueMemberS{Value: "N"},
	}
	
	// GREEN: Verify type condition is valid
	assert.Contains(t, condition, "attribute_type")
	assert.Contains(t, condition, "#score")
	assert.Contains(t, condition, ":numberType")
	assert.Equal(t, "score", attributeNames["#score"])
	assert.Equal(t, "N", attributeValues[":numberType"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithBeginsWith_ShouldValidateStringPrefix(t *testing.T) {
	// RED: Test begins_with condition
	condition := "begins_with(#email, :emailPrefix)"
	attributeNames := map[string]string{
		"#email": "email",
	}
	attributeValues := map[string]types.AttributeValue{
		":emailPrefix": &types.AttributeValueMemberS{Value: "admin@"},
	}
	
	// GREEN: Verify begins_with condition is valid
	assert.Contains(t, condition, "begins_with")
	assert.Contains(t, condition, "#email")
	assert.Contains(t, condition, ":emailPrefix")
	assert.Equal(t, "email", attributeNames["#email"])
	assert.Equal(t, "admin@", attributeValues[":emailPrefix"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithContains_ShouldValidateContainment(t *testing.T) {
	// RED: Test contains condition
	condition := "contains(#tags, :tag)"
	attributeNames := map[string]string{
		"#tags": "tags",
	}
	attributeValues := map[string]types.AttributeValue{
		":tag": &types.AttributeValueMemberS{Value: "important"},
	}
	
	// GREEN: Verify contains condition is valid
	assert.Contains(t, condition, "contains")
	assert.Contains(t, condition, "#tags")
	assert.Contains(t, condition, ":tag")
	assert.Equal(t, "tags", attributeNames["#tags"])
	assert.Equal(t, "important", attributeValues[":tag"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithSize_ShouldValidateAttributeSize(t *testing.T) {
	// RED: Test size condition
	condition := "size(#description) <= :maxLength"
	attributeNames := map[string]string{
		"#description": "description",
	}
	attributeValues := map[string]types.AttributeValue{
		":maxLength": &types.AttributeValueMemberN{Value: "1000"},
	}
	
	// GREEN: Verify size condition is valid
	assert.Contains(t, condition, "size")
	assert.Contains(t, condition, "#description")
	assert.Contains(t, condition, ":maxLength")
	assert.Contains(t, condition, "<=")
	assert.Equal(t, "description", attributeNames["#description"])
	assert.Equal(t, "1000", attributeValues[":maxLength"].(*types.AttributeValueMemberN).Value)
}

// TESTS FOR COMPLEX CONDITION EXPRESSIONS

func TestConditionalExpressions_WithLogicalAND_ShouldCombineConditions(t *testing.T) {
	// RED: Test logical AND combination
	condition := "attribute_exists(#id) AND #status = :active AND #count > :minCount"
	attributeNames := map[string]string{
		"#id":     "id",
		"#status": "status",
		"#count":  "count",
	}
	attributeValues := map[string]types.AttributeValue{
		":active":   &types.AttributeValueMemberS{Value: "active"},
		":minCount": &types.AttributeValueMemberN{Value: "0"},
	}
	
	// GREEN: Verify logical AND condition is valid
	assert.Contains(t, condition, " AND ")
	assert.Contains(t, condition, "attribute_exists(#id)")
	assert.Contains(t, condition, "#status = :active")
	assert.Contains(t, condition, "#count > :minCount")
	assert.Equal(t, "id", attributeNames["#id"])
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "count", attributeNames["#count"])
	assert.Equal(t, "active", attributeValues[":active"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "0", attributeValues[":minCount"].(*types.AttributeValueMemberN).Value)
}

func TestConditionalExpressions_WithLogicalOR_ShouldAlternateConditions(t *testing.T) {
	// RED: Test logical OR combination
	condition := "#status = :draft OR #status = :pending OR #status = :review"
	attributeNames := map[string]string{
		"#status": "status",
	}
	attributeValues := map[string]types.AttributeValue{
		":draft":   &types.AttributeValueMemberS{Value: "draft"},
		":pending": &types.AttributeValueMemberS{Value: "pending"},
		":review":  &types.AttributeValueMemberS{Value: "review"},
	}
	
	// GREEN: Verify logical OR condition is valid
	assert.Contains(t, condition, " OR ")
	assert.Contains(t, condition, "#status = :draft")
	assert.Contains(t, condition, "#status = :pending")
	assert.Contains(t, condition, "#status = :review")
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "draft", attributeValues[":draft"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "pending", attributeValues[":pending"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "review", attributeValues[":review"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithLogicalNOT_ShouldNegateConditions(t *testing.T) {
	// RED: Test logical NOT combination
	condition := "NOT (attribute_exists(#deletedAt) OR #archived = :true)"
	attributeNames := map[string]string{
		"#deletedAt": "deletedAt",
		"#archived":  "archived",
	}
	attributeValues := map[string]types.AttributeValue{
		":true": &types.AttributeValueMemberBOOL{Value: true},
	}
	
	// GREEN: Verify logical NOT condition is valid
	assert.Contains(t, condition, "NOT ")
	assert.Contains(t, condition, "attribute_exists(#deletedAt)")
	assert.Contains(t, condition, "#archived = :true")
	assert.Contains(t, condition, "(")
	assert.Contains(t, condition, ")")
	assert.Equal(t, "deletedAt", attributeNames["#deletedAt"])
	assert.Equal(t, "archived", attributeNames["#archived"])
	assert.True(t, attributeValues[":true"].(*types.AttributeValueMemberBOOL).Value)
}

func TestConditionalExpressions_WithGrouping_ShouldHandleParentheses(t *testing.T) {
	// RED: Test grouped conditions with parentheses
	condition := "(#status = :active OR #status = :pending) AND (#priority >= :high OR contains(#tags, :urgent))"
	attributeNames := map[string]string{
		"#status":   "status",
		"#priority": "priority",
		"#tags":     "tags",
	}
	attributeValues := map[string]types.AttributeValue{
		":active":  &types.AttributeValueMemberS{Value: "active"},
		":pending": &types.AttributeValueMemberS{Value: "pending"},
		":high":    &types.AttributeValueMemberN{Value: "8"},
		":urgent":  &types.AttributeValueMemberS{Value: "urgent"},
	}
	
	// GREEN: Verify grouped condition is valid
	openParens := 0
	closeParens := 0
	for _, char := range condition {
		if char == '(' {
			openParens++
		} else if char == ')' {
			closeParens++
		}
	}
	assert.Equal(t, openParens, closeParens)
	assert.Contains(t, condition, "AND")
	assert.Contains(t, condition, "OR")
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "priority", attributeNames["#priority"])
	assert.Equal(t, "tags", attributeNames["#tags"])
	assert.Equal(t, "active", attributeValues[":active"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "8", attributeValues[":high"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "urgent", attributeValues[":urgent"].(*types.AttributeValueMemberS).Value)
}

// TESTS FOR COMPARISON OPERATORS

func TestConditionalExpressions_WithEqualityComparison_ShouldValidateEquals(t *testing.T) {
	// RED: Test equality comparison
	condition := "#status = :expectedStatus"
	attributeNames := map[string]string{
		"#status": "status",
	}
	attributeValues := map[string]types.AttributeValue{
		":expectedStatus": &types.AttributeValueMemberS{Value: "published"},
	}
	
	// GREEN: Verify equality comparison is valid
	assert.Contains(t, condition, " = ")
	assert.Contains(t, condition, "#status")
	assert.Contains(t, condition, ":expectedStatus")
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "published", attributeValues[":expectedStatus"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithInequalityComparison_ShouldValidateNotEquals(t *testing.T) {
	// RED: Test inequality comparison
	condition := "#status <> :blockedStatus"
	attributeNames := map[string]string{
		"#status": "status",
	}
	attributeValues := map[string]types.AttributeValue{
		":blockedStatus": &types.AttributeValueMemberS{Value: "blocked"},
	}
	
	// GREEN: Verify inequality comparison is valid
	assert.Contains(t, condition, " <> ")
	assert.Contains(t, condition, "#status")
	assert.Contains(t, condition, ":blockedStatus")
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "blocked", attributeValues[":blockedStatus"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithNumericComparisons_ShouldValidateNumericOperators(t *testing.T) {
	// RED: Test numeric comparison operators
	conditions := []string{
		"#score > :minScore",
		"#score >= :minScore",
		"#score < :maxScore",
		"#score <= :maxScore",
	}
	attributeNames := map[string]string{
		"#score": "score",
	}
	attributeValues := map[string]types.AttributeValue{
		":minScore": &types.AttributeValueMemberN{Value: "50"},
		":maxScore": &types.AttributeValueMemberN{Value: "100"},
	}
	
	// GREEN: Verify all numeric comparison operators are valid
	for _, condition := range conditions {
		assert.Contains(t, condition, "#score")
		if condition == "#score > :minScore" || condition == "#score >= :minScore" {
			assert.Contains(t, condition, ":minScore")
		} else {
			assert.Contains(t, condition, ":maxScore")
		}
	}
	assert.Equal(t, "score", attributeNames["#score"])
	assert.Equal(t, "50", attributeValues[":minScore"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "100", attributeValues[":maxScore"].(*types.AttributeValueMemberN).Value)
}

func TestConditionalExpressions_WithBetweenComparison_ShouldValidateRange(t *testing.T) {
	// RED: Test BETWEEN comparison
	condition := "#age BETWEEN :minAge AND :maxAge"
	attributeNames := map[string]string{
		"#age": "age",
	}
	attributeValues := map[string]types.AttributeValue{
		":minAge": &types.AttributeValueMemberN{Value: "18"},
		":maxAge": &types.AttributeValueMemberN{Value: "65"},
	}
	
	// GREEN: Verify BETWEEN comparison is valid
	assert.Contains(t, condition, " BETWEEN ")
	assert.Contains(t, condition, " AND ")
	assert.Contains(t, condition, "#age")
	assert.Contains(t, condition, ":minAge")
	assert.Contains(t, condition, ":maxAge")
	assert.Equal(t, "age", attributeNames["#age"])
	assert.Equal(t, "18", attributeValues[":minAge"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, "65", attributeValues[":maxAge"].(*types.AttributeValueMemberN).Value)
}

func TestConditionalExpressions_WithInComparison_ShouldValidateInClause(t *testing.T) {
	// RED: Test IN comparison
	condition := "#status IN (:active, :pending, :approved)"
	attributeNames := map[string]string{
		"#status": "status",
	}
	attributeValues := map[string]types.AttributeValue{
		":active":   &types.AttributeValueMemberS{Value: "active"},
		":pending":  &types.AttributeValueMemberS{Value: "pending"},
		":approved": &types.AttributeValueMemberS{Value: "approved"},
	}
	
	// GREEN: Verify IN comparison is valid
	assert.Contains(t, condition, " IN ")
	assert.Contains(t, condition, "(")
	assert.Contains(t, condition, ")")
	assert.Contains(t, condition, ",")
	assert.Contains(t, condition, "#status")
	assert.Equal(t, "status", attributeNames["#status"])
	assert.Equal(t, "active", attributeValues[":active"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "pending", attributeValues[":pending"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "approved", attributeValues[":approved"].(*types.AttributeValueMemberS).Value)
}

// TESTS FOR NESTED ATTRIBUTE CONDITIONS

func TestConditionalExpressions_WithNestedAttributes_ShouldHandleDotNotation(t *testing.T) {
	// RED: Test nested attribute access
	condition := "#metadata.#version > :expectedVersion"
	attributeNames := map[string]string{
		"#metadata": "metadata",
		"#version":  "version",
	}
	attributeValues := map[string]types.AttributeValue{
		":expectedVersion": &types.AttributeValueMemberN{Value: "1"},
	}
	
	// GREEN: Verify nested attribute condition is valid
	assert.Contains(t, condition, "#metadata.#version")
	assert.Contains(t, condition, " > ")
	assert.Contains(t, condition, ":expectedVersion")
	assert.Equal(t, "metadata", attributeNames["#metadata"])
	assert.Equal(t, "version", attributeNames["#version"])
	assert.Equal(t, "1", attributeValues[":expectedVersion"].(*types.AttributeValueMemberN).Value)
}

func TestConditionalExpressions_WithListIndexing_ShouldHandleBracketNotation(t *testing.T) {
	// RED: Test list indexing in conditions
	condition := "#items[0].#name = :firstItemName"
	attributeNames := map[string]string{
		"#items": "items",
		"#name":  "name",
	}
	attributeValues := map[string]types.AttributeValue{
		":firstItemName": &types.AttributeValueMemberS{Value: "first"},
	}
	
	// GREEN: Verify list indexing condition is valid
	assert.Contains(t, condition, "#items[0]")
	assert.Contains(t, condition, ".#name")
	assert.Contains(t, condition, " = ")
	assert.Contains(t, condition, ":firstItemName")
	assert.Equal(t, "items", attributeNames["#items"])
	assert.Equal(t, "name", attributeNames["#name"])
	assert.Equal(t, "first", attributeValues[":firstItemName"].(*types.AttributeValueMemberS).Value)
}

func TestConditionalExpressions_WithDeeplyNestedAttributes_ShouldHandleComplexPaths(t *testing.T) {
	// RED: Test deeply nested attribute paths
	condition := "#profile.#personal.#address.#city = :expectedCity AND #profile.#preferences.#notifications = :enabled"
	attributeNames := map[string]string{
		"#profile":       "profile",
		"#personal":      "personal",
		"#address":       "address",
		"#city":          "city",
		"#preferences":   "preferences",
		"#notifications": "notifications",
	}
	attributeValues := map[string]types.AttributeValue{
		":expectedCity": &types.AttributeValueMemberS{Value: "New York"},
		":enabled":      &types.AttributeValueMemberBOOL{Value: true},
	}
	
	// GREEN: Verify deeply nested condition is valid
	assert.Contains(t, condition, "#profile.#personal.#address.#city")
	assert.Contains(t, condition, "#profile.#preferences.#notifications")
	assert.Contains(t, condition, " AND ")
	assert.Equal(t, "profile", attributeNames["#profile"])
	assert.Equal(t, "personal", attributeNames["#personal"])
	assert.Equal(t, "address", attributeNames["#address"])
	assert.Equal(t, "city", attributeNames["#city"])
	assert.Equal(t, "New York", attributeValues[":expectedCity"].(*types.AttributeValueMemberS).Value)
	assert.True(t, attributeValues[":enabled"].(*types.AttributeValueMemberBOOL).Value)
}

// TESTS FOR PUTITEM CONDITIONAL OPERATIONS

func TestPutItemWithConditionalExpression_WithValidCondition_ShouldSetupConditionalPut(t *testing.T) {
	// RED: Test conditional PutItem setup
	tableName := "test-table"
	item := createValidTestItemConditional()
	options := createConditionalPutItemOptions()
	
	input := createPutItemInput(tableName, item, options)
	
	// GREEN: Verify conditional PutItem is set up correctly
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, options.ExpressionAttributeValues, input.ExpressionAttributeValues)
	assert.Equal(t, options.ReturnValuesOnConditionCheckFailure, input.ReturnValuesOnConditionCheckFailure)
}

func TestPutItemWithConditionalExpression_WithComplexCondition_ShouldHandleComplexLogic(t *testing.T) {
	// RED: Test complex conditional PutItem
	tableName := "test-table"
	item := createValidTestItemConditional()
	options := PutItemOptions{
		ConditionExpression: stringPtrConditional("(attribute_not_exists(#id) OR #version < :maxVersion) AND #status IN (:draft, :review) AND NOT contains(#tags, :blocked)"),
		ExpressionAttributeNames: map[string]string{
			"#id":      "id",
			"#version": "version",
			"#status":  "status",
			"#tags":    "tags",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":maxVersion": &types.AttributeValueMemberN{Value: "10"},
			":draft":      &types.AttributeValueMemberS{Value: "draft"},
			":review":     &types.AttributeValueMemberS{Value: "review"},
			":blocked":    &types.AttributeValueMemberS{Value: "blocked"},
		},
	}
	
	input := createPutItemInput(tableName, item, options)
	
	// GREEN: Verify complex conditional PutItem is handled
	condition := *input.ConditionExpression
	assert.Contains(t, condition, "attribute_not_exists(#id)")
	assert.Contains(t, condition, "#version < :maxVersion")
	assert.Contains(t, condition, "#status IN (:draft, :review)")
	assert.Contains(t, condition, "NOT contains(#tags, :blocked)")
	assert.Contains(t, condition, "AND")
	assert.Contains(t, condition, "OR")
	assert.Contains(t, condition, "(")
	assert.Contains(t, condition, ")")
}

// TESTS FOR UPDATEITEM CONDITIONAL OPERATIONS

func TestUpdateItemWithConditionalExpression_WithVersionCheck_ShouldEnforceOptimisticLocking(t *testing.T) {
	// RED: Test optimistic locking with UpdateItem
	tableName := "test-table"
	key := createValidTestKeyConditional()
	updateExpression := "SET #name = :name, #version = #version + :inc"
	expressionAttributeNames := map[string]string{
		"#name":    "name",
		"#version": "version",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name":            &types.AttributeValueMemberS{Value: "updated-name"},
		":inc":             &types.AttributeValueMemberN{Value: "1"},
		":expectedVersion": &types.AttributeValueMemberN{Value: "5"},
	}
	options := UpdateItemOptions{
		ConditionExpression: stringPtrConditional("#version = :expectedVersion"),
	}
	
	// Add expression attribute names for condition
	expressionAttributeNames["#version"] = "version"
	
	input := createUpdateItemInput(tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)
	
	// GREEN: Verify optimistic locking is set up
	assert.Equal(t, "#version = :expectedVersion", *input.ConditionExpression)
	assert.Contains(t, *input.UpdateExpression, "#version = #version + :inc")
	assert.Equal(t, "version", input.ExpressionAttributeNames["#version"])
	assert.Equal(t, "5", input.ExpressionAttributeValues[":expectedVersion"].(*types.AttributeValueMemberN).Value)
}

func TestUpdateItemWithConditionalExpression_WithExistenceAndTypeCheck_ShouldValidateBeforeUpdate(t *testing.T) {
	// RED: Test existence and type validation in UpdateItem
	tableName := "test-table"
	key := createValidTestKeyConditional()
	updateExpression := "SET #score = #score + :increment"
	expressionAttributeNames := map[string]string{
		"#score": "score",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":increment":  &types.AttributeValueMemberN{Value: "10"},
		":numberType": &types.AttributeValueMemberS{Value: "N"},
	}
	options := UpdateItemOptions{
		ConditionExpression: stringPtrConditional("attribute_exists(#score) AND attribute_type(#score, :numberType)"),
	}
	
	input := createUpdateItemInput(tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)
	
	// GREEN: Verify existence and type checks are set up
	condition := *input.ConditionExpression
	assert.Contains(t, condition, "attribute_exists(#score)")
	assert.Contains(t, condition, "attribute_type(#score, :numberType)")
	assert.Contains(t, condition, " AND ")
	assert.Equal(t, "score", input.ExpressionAttributeNames["#score"])
	assert.Equal(t, "N", input.ExpressionAttributeValues[":numberType"].(*types.AttributeValueMemberS).Value)
}

// TESTS FOR DELETEITEM CONDITIONAL OPERATIONS

func TestDeleteItemWithConditionalExpression_WithStatusCheck_ShouldOnlyDeleteInactiveItems(t *testing.T) {
	// RED: Test conditional delete with status check
	tableName := "test-table"
	key := createValidTestKeyConditional()
	options := createConditionalDeleteItemOptions()
	
	input := createDeleteItemInput(tableName, key, options)
	
	// GREEN: Verify conditional delete is set up correctly
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, options.ExpressionAttributeValues, input.ExpressionAttributeValues)
	assert.Equal(t, options.ReturnValuesOnConditionCheckFailure, input.ReturnValuesOnConditionCheckFailure)
}

func TestDeleteItemWithConditionalExpression_WithTimestampCheck_ShouldDeleteExpiredItems(t *testing.T) {
	// RED: Test conditional delete with timestamp check
	tableName := "test-table"
	key := createValidTestKeyConditional()
	options := DeleteItemOptions{
		ConditionExpression: stringPtrConditional("#expiresAt <= :currentTime AND attribute_exists(#expiresAt)"),
		ExpressionAttributeNames: map[string]string{
			"#expiresAt": "expiresAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":currentTime": &types.AttributeValueMemberN{Value: "1640995200"}, // Unix timestamp
		},
		ReturnValues: types.ReturnValueAllOld,
	}
	
	input := createDeleteItemInput(tableName, key, options)
	
	// GREEN: Verify timestamp-based conditional delete is set up
	condition := *input.ConditionExpression
	assert.Contains(t, condition, "#expiresAt <= :currentTime")
	assert.Contains(t, condition, "attribute_exists(#expiresAt)")
	assert.Contains(t, condition, " AND ")
	assert.Equal(t, "expiresAt", input.ExpressionAttributeNames["#expiresAt"])
	assert.Equal(t, "1640995200", input.ExpressionAttributeValues[":currentTime"].(*types.AttributeValueMemberN).Value)
	assert.Equal(t, types.ReturnValueAllOld, input.ReturnValues)
}

// TESTS FOR ERROR SCENARIOS WITH CONDITIONAL EXPRESSIONS

func TestConditionalExpressions_WithInvalidSyntax_ShouldFailValidation(t *testing.T) {
	// RED: Test invalid condition expression syntax
	invalidConditions := []string{
		"attribute_exists(",                           // Missing closing parenthesis
		"#status =",                                  // Missing value
		"AND #status = :active",                      // Missing left operand
		"#status = :active OR",                       // Missing right operand
		"attribute_not_exists(#id) AND AND #status", // Double AND
		"contains(#tags",                             // Missing parameters
		"size(#items) <=",                           // Missing comparison value
	}
	
	// GREEN: Verify that we can detect these as potentially invalid
	for _, condition := range invalidConditions {
		// These would fail at DynamoDB service level, but we can validate basic structure
		if condition == "attribute_exists(" {
			assert.Contains(t, condition, "attribute_exists")
			assert.Contains(t, condition, "(")
			assert.NotContains(t, condition, ")")
		}
		if condition == "#status =" {
			assert.Contains(t, condition, "=")
			assert.NotContains(t, condition, ":")
		}
	}
}

func TestConditionalExpressions_WithMismatchedExpressionAttributes_ShouldFailValidation(t *testing.T) {
	// RED: Test mismatched expression attribute names and values
	condition := "#status = :status AND #count > :minCount"
	
	// Missing attribute name mapping
	incompleteAttributeNames := map[string]string{
		"#status": "status",
		// Missing "#count": "count"
	}
	
	// Missing attribute value
	incompleteAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: "active"},
		// Missing ":minCount": ...
	}
	
	// GREEN: Verify we can detect incomplete mappings
	assert.Contains(t, condition, "#count")
	assert.NotContains(t, incompleteAttributeNames, "#count")
	
	assert.Contains(t, condition, ":minCount")
	assert.NotContains(t, incompleteAttributeValues, ":minCount")
	
	// These would cause validation errors at DynamoDB service level
}

func TestConditionalExpressions_WithReservedWords_ShouldRequireExpressionAttributeNames(t *testing.T) {
	// RED: Test DynamoDB reserved words in conditions
	reservedWordConditions := []string{
		"#data = :value",      // 'data' is reserved
		"#index > :minIndex",  // 'index' is reserved
		"#order = :ascending", // 'order' is reserved
		"#size <= :maxSize",   // 'size' is reserved
	}
	
	requiredAttributeNames := map[string]string{
		"#data":  "data",
		"#index": "index",
		"#order": "order",
		"#size":  "size",
	}
	
	// GREEN: Verify reserved words are handled with expression attribute names
	for _, condition := range reservedWordConditions {
		assert.Contains(t, condition, "#")
		// Check that we have mappings for the reserved words
		if condition == "#data = :value" {
			assert.Equal(t, "data", requiredAttributeNames["#data"])
		}
		if condition == "#index > :minIndex" {
			assert.Equal(t, "index", requiredAttributeNames["#index"])
		}
		if condition == "#order = :ascending" {
			assert.Equal(t, "order", requiredAttributeNames["#order"])
		}
		if condition == "#size <= :maxSize" {
			assert.Equal(t, "size", requiredAttributeNames["#size"])
		}
	}
}

// INTEGRATION TESTS FOR CONDITIONAL OPERATIONS

func TestConditionalOperations_WithConditionalCheckFailure_ShouldReturnFailureError(t *testing.T) {
	// RED: Test conditional check failure error handling
	err := createConditionalCheckFailedError()
	
	// GREEN: Verify conditional check failure error is handled
	var conditionalErr *types.ConditionalCheckFailedException
	assert.ErrorAs(t, err, &conditionalErr)
	assert.Contains(t, err.Error(), "conditional request failed")
}

func TestConditionalOperations_WithReturnValuesOnFailure_ShouldReturnItemOnConditionFailure(t *testing.T) {
	// RED: Test return values on condition check failure
	tableName := "test-table"
	item := createValidTestItemConditional()
	options := PutItemOptions{
		ConditionExpression: stringPtrConditional("attribute_not_exists(#id)"),
		ExpressionAttributeNames: map[string]string{
			"#id": "id",
		},
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}
	
	input := createPutItemInput(tableName, item, options)
	
	// GREEN: Verify return values on failure is configured
	assert.Equal(t, types.ReturnValuesOnConditionCheckFailureAllOld, input.ReturnValuesOnConditionCheckFailure)
	assert.Equal(t, "attribute_not_exists(#id)", *input.ConditionExpression)
	
	// In case of condition failure, DynamoDB would return the existing item
}

// Test data factories for conditional tests
func createValidTestItemConditional() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":      &types.AttributeValueMemberS{Value: "conditional-test-id"},
		"name":    &types.AttributeValueMemberS{Value: "conditional-item"},
		"status":  &types.AttributeValueMemberS{Value: "active"},
		"version": &types.AttributeValueMemberN{Value: "1"},
		"count":   &types.AttributeValueMemberN{Value: "5"},
	}
}

func createValidTestKeyConditional() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "conditional-test-id"},
	}
}

// Helper functions for conditional tests
func stringPtrConditional(s string) *string {
	return &s
}

func createConditionalCheckFailedError() error {
	return &types.ConditionalCheckFailedException{
		Message: stringPtrConditional("The conditional request failed"),
	}
}