package eventbridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateScheduleExpression(t *testing.T) {
	t.Run("ValidRateExpressions", func(t *testing.T) {
		validExpressions := []string{
			"rate(1 minute)",
			"rate(5 minutes)",
			"rate(1 hour)",
			"rate(12 hours)",
			"rate(1 day)",
			"rate(30 days)",
		}

		for _, expr := range validExpressions {
			t.Run(expr, func(t *testing.T) {
				err := ValidateScheduleExpressionE(expr)
				assert.NoError(t, err, "Expected valid rate expression: %s", expr)
			})
		}
	})

	t.Run("ValidCronExpressions", func(t *testing.T) {
		validExpressions := []string{
			"cron(0 12 * * ? *)",           // Daily at noon
			"cron(15 10 ? * MON-FRI *)",    // Weekdays at 10:15
			"cron(0 8 1 * ? *)",            // First of every month at 8 AM
			"cron(0/15 * * * ? *)",         // Every 15 minutes
			"cron(0 18 ? * SUN *)",         // Every Sunday at 6 PM
		}

		for _, expr := range validExpressions {
			t.Run(expr, func(t *testing.T) {
				err := ValidateScheduleExpressionE(expr)
				assert.NoError(t, err, "Expected valid cron expression: %s", expr)
			})
		}
	})

	t.Run("InvalidExpressions", func(t *testing.T) {
		testCases := []struct {
			expression    string
			expectedError error
		}{
			{"", ErrEmptyScheduleExpression},
			{"invalid", ErrInvalidScheduleFormat},
			{"rate()", ErrInvalidRateExpression},
			{"rate(0 minutes)", ErrInvalidRateValue},
			{"rate(-1 hours)", ErrInvalidRateValue},
			{"rate(1 seconds)", ErrInvalidRateUnit},
			{"rate(5 minute)", ErrInvalidRateUnit},
			{"rate(1 hours)", ErrInvalidRateUnit},
			{"cron()", ErrInvalidCronExpression},
			{"cron(0 12)", ErrCronTooFewFields},
			{"cron(0 12 * * ? * * extra)", ErrCronTooManyFields},
			{"cron(0 12 15 * MON *)", ErrCronInvalidDaySpecification},
		}

		for _, tc := range testCases {
			t.Run(tc.expression, func(t *testing.T) {
				err := ValidateScheduleExpressionE(tc.expression)
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
			})
		}
	})
}

func TestBuildRateExpression(t *testing.T) {
	testCases := []struct {
		name     string
		value    int
		unit     string
		expected string
	}{
		{"SingleMinute", 1, "minute", "rate(1 minute)"},
		{"MultipleMinutes", 5, "minutes", "rate(5 minutes)"},
		{"SingleHour", 1, "hour", "rate(1 hour)"},
		{"MultipleHours", 12, "hours", "rate(12 hours)"},
		{"SingleDay", 1, "day", "rate(1 day)"},
		{"MultipleDays", 30, "days", "rate(30 days)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildRateExpression(tc.value, tc.unit)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildCronExpression(t *testing.T) {
	result := BuildCronExpression("0", "12", "*", "*", "?", "*")
	expected := "cron(0 12 * * ? *)"
	assert.Equal(t, expected, result)
}

func TestIsRateExpression(t *testing.T) {
	testCases := []struct {
		expression string
		expected   bool
	}{
		{"rate(5 minutes)", true},
		{"cron(0 12 * * ? *)", false},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			result := IsRateExpression(tc.expression)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsCronExpression(t *testing.T) {
	testCases := []struct {
		expression string
		expected   bool
	}{
		{"cron(0 12 * * ? *)", true},
		{"rate(5 minutes)", false},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			result := IsCronExpression(tc.expression)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseRateExpression(t *testing.T) {
	t.Run("ValidExpressions", func(t *testing.T) {
		testCases := []struct {
			expression    string
			expectedValue int
			expectedUnit  string
		}{
			{"rate(1 minute)", 1, "minute"},
			{"rate(5 minutes)", 5, "minutes"},
			{"rate(1 hour)", 1, "hour"},
			{"rate(12 hours)", 12, "hours"},
			{"rate(1 day)", 1, "day"},
			{"rate(30 days)", 30, "days"},
		}

		for _, tc := range testCases {
			t.Run(tc.expression, func(t *testing.T) {
				value, unit, err := ParseRateExpression(tc.expression)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedValue, value)
				assert.Equal(t, tc.expectedUnit, unit)
			})
		}
	})

	t.Run("InvalidExpressions", func(t *testing.T) {
		testCases := []string{
			"invalid",
			"cron(0 12 * * ? *)",
			"rate()",
			"rate(abc minutes)",
		}

		for _, expr := range testCases {
			t.Run(expr, func(t *testing.T) {
				_, _, err := ParseRateExpression(expr)
				require.Error(t, err)
			})
		}
	})
}

func TestParseCronExpression(t *testing.T) {
	t.Run("ValidExpressions", func(t *testing.T) {
		testCases := []struct {
			expression     string
			expectedFields []string
		}{
			{"cron(0 12 * * ? *)", []string{"0", "12", "*", "*", "?", "*"}},
			{"cron(15 10 ? * MON-FRI *)", []string{"15", "10", "?", "*", "MON-FRI", "*"}},
			{"cron(0/15 * * * ? *)", []string{"0/15", "*", "*", "*", "?", "*"}},
		}

		for _, tc := range testCases {
			t.Run(tc.expression, func(t *testing.T) {
				fields, err := ParseCronExpression(tc.expression)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedFields, fields)
			})
		}
	})

	t.Run("InvalidExpressions", func(t *testing.T) {
		testCases := []string{
			"invalid",
			"rate(5 minutes)",
			"cron()",
			"cron(0 12)",
			"cron(0 12 * * ? * * extra)",
		}

		for _, expr := range testCases {
			t.Run(expr, func(t *testing.T) {
				_, err := ParseCronExpression(expr)
				require.Error(t, err)
			})
		}
	})
}