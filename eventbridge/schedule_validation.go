package eventbridge

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Common schedule validation errors
var (
	ErrEmptyScheduleExpression      = errors.New("schedule expression cannot be empty")
	ErrInvalidScheduleFormat        = errors.New("invalid schedule expression format")
	ErrInvalidRateExpression        = errors.New("invalid rate expression")
	ErrInvalidCronExpression        = errors.New("invalid cron expression")
	ErrInvalidRateValue             = errors.New("invalid rate value")
	ErrInvalidRateUnit              = errors.New("invalid rate unit")
	ErrInvalidCronField             = errors.New("invalid cron field")
	ErrCronTooManyFields            = errors.New("cron expression has too many fields")
	ErrCronTooFewFields             = errors.New("cron expression has too few fields")
	ErrCronInvalidDaySpecification  = errors.New("day of month and day of week cannot both be specified")
)

// Regular expressions for validation
var (
	rateExpressionRegex = regexp.MustCompile(`^rate\((-?\d+)\s+(\w+)\)$`)
	cronExpressionRegex = regexp.MustCompile(`^cron\((.*)\)$`)
	cronFieldRegex      = regexp.MustCompile(`^[0-9*?/,-L#]+$|^[A-Z]{3}(-[A-Z]{3})?$|^\*$|^\?$`)
)

// ValidateScheduleExpression validates a schedule expression and panics on error (Terratest pattern)
func ValidateScheduleExpression(t TestingT, expression string) {
	err := ValidateScheduleExpressionE(expression)
	if err != nil {
		t.Errorf("ValidateScheduleExpression failed: %v", err)
		t.FailNow()
	}
}

// ValidateScheduleExpressionE validates a schedule expression and returns error
func ValidateScheduleExpressionE(expression string) error {
	if expression == "" {
		return ErrEmptyScheduleExpression
	}
	
	// Check if it's a rate expression
	if strings.HasPrefix(expression, "rate(") {
		return validateRateExpressionE(expression)
	}
	
	// Check if it's a cron expression
	if strings.HasPrefix(expression, "cron(") {
		return validateCronExpressionE(expression)
	}
	
	// Neither rate nor cron format
	return fmt.Errorf("%w: expression must start with 'rate(' or 'cron('", ErrInvalidScheduleFormat)
}

// validateRateExpressionE validates rate expressions like rate(5 minutes)
func validateRateExpressionE(expression string) error {
	matches := rateExpressionRegex.FindStringSubmatch(expression)
	if len(matches) != 3 {
		return fmt.Errorf("%w: %s", ErrInvalidRateExpression, expression)
	}
	
	valueStr := matches[1]
	unit := matches[2]
	
	// Parse and validate the value
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fmt.Errorf("%w: value must be a positive integer", ErrInvalidRateValue)
	}
	
	if value <= 0 {
		return fmt.Errorf("%w: value must be greater than 0", ErrInvalidRateValue)
	}
	
	// Validate the unit
	validUnits := map[string]bool{
		"minute":  true,
		"minutes": true,
		"hour":    true,
		"hours":   true,
		"day":     true,
		"days":    true,
	}
	
	if !validUnits[unit] {
		return fmt.Errorf("%w: unit must be minute(s), hour(s), or day(s)", ErrInvalidRateUnit)
	}
	
	// Unit consistency check
	if value == 1 && strings.HasSuffix(unit, "s") {
		return fmt.Errorf("%w: use singular unit for value 1", ErrInvalidRateUnit)
	}
	if value > 1 && !strings.HasSuffix(unit, "s") {
		return fmt.Errorf("%w: use plural unit for value greater than 1", ErrInvalidRateUnit)
	}
	
	return nil
}

// validateCronExpressionE validates cron expressions like cron(0 12 * * ? *)
func validateCronExpressionE(expression string) error {
	matches := cronExpressionRegex.FindStringSubmatch(expression)
	if len(matches) != 2 {
		return fmt.Errorf("%w: %s", ErrInvalidCronExpression, expression)
	}
	
	cronFields := strings.Fields(matches[1])
	
	// Cron expressions must have exactly 6 fields
	if len(cronFields) < 6 {
		// Special case for empty cron() - return the more specific error first
		if len(cronFields) == 0 {
			return fmt.Errorf("%w: %s", ErrInvalidCronExpression, expression)
		}
		return fmt.Errorf("%w: expected 6 fields, got %d", ErrCronTooFewFields, len(cronFields))
	}
	if len(cronFields) > 6 {
		return fmt.Errorf("%w: expected 6 fields, got %d", ErrCronTooManyFields, len(cronFields))
	}
	
	// Validate each field
	if err := validateCronField(cronFields[0], "minute", 0, 59); err != nil {
		return err
	}
	if err := validateCronField(cronFields[1], "hour", 0, 23); err != nil {
		return err
	}
	if err := validateCronField(cronFields[2], "day", 1, 31); err != nil {
		return err
	}
	if err := validateCronField(cronFields[3], "month", 1, 12); err != nil {
		return err
	}
	if err := validateCronField(cronFields[4], "day-of-week", 1, 7); err != nil {
		return err
	}
	if err := validateCronField(cronFields[5], "year", 1970, 3000); err != nil {
		return err
	}
	
	// Validate day-of-month and day-of-week constraint
	dayOfMonth := cronFields[2]
	dayOfWeek := cronFields[4]
	
	if dayOfMonth != "?" && dayOfWeek != "?" {
		return ErrCronInvalidDaySpecification
	}
	
	return nil
}

// validateCronField validates individual cron fields
func validateCronField(field string, fieldName string, min int, max int) error {
	// Allow ? and * wildcards
	if field == "?" || field == "*" {
		return nil
	}
	
	// Check basic format
	if !cronFieldRegex.MatchString(field) {
		return fmt.Errorf("%w: invalid %s field format: %s", ErrInvalidCronField, fieldName, field)
	}
	
	// For numeric validation, try to parse basic numeric values
	// This is a simplified validation - full cron parsing would be much more complex
	if strings.ContainsAny(field, ",-/L#") {
		// Complex expressions - basic format check only
		return nil
	}
	
	// Simple numeric value
	if value, err := strconv.Atoi(field); err == nil {
		if value < min || value > max {
			return fmt.Errorf("%w: %s value %d out of range [%d-%d]", ErrInvalidCronField, fieldName, value, min, max)
		}
	}
	
	return nil
}

// BuildRateExpression creates a rate expression from value and unit
func BuildRateExpression(value int, unit string) string {
	return fmt.Sprintf("rate(%d %s)", value, unit)
}

// BuildCronExpression creates a cron expression from individual fields
func BuildCronExpression(minute, hour, day, month, dayOfWeek, year string) string {
	return fmt.Sprintf("cron(%s %s %s %s %s %s)", minute, hour, day, month, dayOfWeek, year)
}

// IsRateExpression checks if the expression is a rate expression
func IsRateExpression(expression string) bool {
	return strings.HasPrefix(expression, "rate(")
}

// IsCronExpression checks if the expression is a cron expression
func IsCronExpression(expression string) bool {
	return strings.HasPrefix(expression, "cron(")
}

// ParseRateExpression parses a rate expression and returns the value and unit
func ParseRateExpression(expression string) (int, string, error) {
	matches := rateExpressionRegex.FindStringSubmatch(expression)
	if len(matches) != 3 {
		return 0, "", fmt.Errorf("%w: %s", ErrInvalidRateExpression, expression)
	}
	
	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("%w: value must be an integer", ErrInvalidRateValue)
	}
	
	return value, matches[2], nil
}

// ParseCronExpression parses a cron expression and returns the individual fields
func ParseCronExpression(expression string) ([]string, error) {
	matches := cronExpressionRegex.FindStringSubmatch(expression)
	if len(matches) != 2 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCronExpression, expression)
	}
	
	fields := strings.Fields(matches[1])
	if len(fields) != 6 {
		return nil, fmt.Errorf("%w: expected 6 fields, got %d", ErrInvalidCronExpression, len(fields))
	}
	
	return fields, nil
}