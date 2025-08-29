package security

import (
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	config ValidationConfig
}

// ValidationConfig holds configuration for input validation
type ValidationConfig struct {
	MaxStringLength    int
	AllowHTML         bool
	AllowScripts      bool
	RequireHTTPS      bool
	AllowedDomains    []string
	BlockedPatterns   []string
	EnableSanitization bool
	StrictMode        bool
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid     bool
	Value       interface{}
	Errors      []string
	Warnings    []string
	Sanitized   bool
	OriginalValue interface{}
}

// SanitizationLevel defines the level of sanitization to apply
type SanitizationLevel string

const (
	SanitizationNone   SanitizationLevel = "none"
	SanitizationBasic  SanitizationLevel = "basic"
	SanitizationStrict SanitizationLevel = "strict"
	SanitizationCustom SanitizationLevel = "custom"
)

// NewInputValidator creates a new input validator with default configuration
func NewInputValidator(config ValidationConfig) *InputValidator {
	// Set defaults
	if config.MaxStringLength == 0 {
		config.MaxStringLength = 1000
	}

	return &InputValidator{
		config: config,
	}
}

// ValidateString validates and sanitizes string input
func (iv *InputValidator) ValidateString(input string, rules StringValidationRules) ValidationResult {
	result := ValidationResult{
		IsValid:       true,
		OriginalValue: input,
	}

	// Check length constraints
	if rules.MinLength > 0 && len(input) < rules.MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("String must be at least %d characters long", rules.MinLength))
	}

	maxLength := rules.MaxLength
	if maxLength == 0 {
		maxLength = iv.config.MaxStringLength
	}
	
	if len(input) > maxLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("String exceeds maximum length of %d characters", maxLength))
	}

	// Check for required patterns
	if rules.RequiredPattern != "" {
		matched, err := regexp.MatchString(rules.RequiredPattern, input)
		if err != nil || !matched {
			result.IsValid = false
			result.Errors = append(result.Errors, "String does not match required pattern")
		}
	}

	// Check for forbidden patterns
	for _, pattern := range rules.ForbiddenPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("String contains forbidden pattern: %s", pattern))
		}
	}

	// Check for blocked patterns from config
	for _, pattern := range iv.config.BlockedPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			result.IsValid = false
			result.Errors = append(result.Errors, "String contains blocked content")
		}
	}

	// Sanitize if enabled
	sanitizedValue := input
	if iv.config.EnableSanitization || rules.Sanitize {
		sanitizedValue = iv.sanitizeString(input, rules.SanitizationLevel)
		if sanitizedValue != input {
			result.Sanitized = true
		}
	}

	result.Value = sanitizedValue

	return result
}

// StringValidationRules defines validation rules for strings
type StringValidationRules struct {
	MinLength         int
	MaxLength         int
	RequiredPattern   string
	ForbiddenPatterns []string
	Sanitize          bool
	SanitizationLevel SanitizationLevel
	AllowEmpty        bool
}

// ValidateEmail validates email addresses
func (iv *InputValidator) ValidateEmail(email string) ValidationResult {
	result := ValidationResult{
		OriginalValue: email,
		Value:         email,
	}

	// Basic email validation using net/mail
	_, err := mail.ParseAddress(email)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid email format")
		return result
	}

	// Additional validation rules
	if len(email) > 254 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Email address exceeds maximum length")
		return result
	}

	// Check domain if allowed domains are specified
	if len(iv.config.AllowedDomains) > 0 {
		domain := email[strings.LastIndex(email, "@")+1:]
		allowed := false
		for _, allowedDomain := range iv.config.AllowedDomains {
			if domain == allowedDomain || strings.HasSuffix(domain, "."+allowedDomain) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.IsValid = false
			result.Errors = append(result.Errors, "Email domain is not allowed")
			return result
		}
	}

	result.IsValid = true
	return result
}

// ValidateURL validates and sanitizes URLs
func (iv *InputValidator) ValidateURL(urlString string) ValidationResult {
	result := ValidationResult{
		OriginalValue: urlString,
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid URL format")
		return result
	}

	// Check scheme requirements
	if iv.config.RequireHTTPS && parsedURL.Scheme != "https" {
		result.IsValid = false
		result.Errors = append(result.Errors, "HTTPS is required")
		return result
	}

	// Validate allowed schemes
	allowedSchemes := []string{"http", "https", "ftp"}
	schemeAllowed := false
	for _, scheme := range allowedSchemes {
		if parsedURL.Scheme == scheme {
			schemeAllowed = true
			break
		}
	}
	if !schemeAllowed {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("URL scheme '%s' is not allowed", parsedURL.Scheme))
		return result
	}

	// Check for suspicious URL patterns
	suspiciousPatterns := []string{
		`javascript:`,
		`data:`,
		`vbscript:`,
		`file:`,
	}

	lowerURL := strings.ToLower(urlString)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerURL, pattern) {
			result.IsValid = false
			result.Errors = append(result.Errors, "URL contains potentially dangerous scheme")
			return result
		}
	}

	result.IsValid = true
	result.Value = parsedURL.String() // Use parsed URL for consistency
	return result
}

// ValidateIPAddress validates IP addresses (IPv4 and IPv6)
func (iv *InputValidator) ValidateIPAddress(ip string) ValidationResult {
	result := ValidationResult{
		OriginalValue: ip,
		Value:         ip,
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid IP address format")
		return result
	}

	// Check for private/internal IP ranges in strict mode
	if iv.config.StrictMode {
		if parsedIP.IsPrivate() || parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() {
			result.Warnings = append(result.Warnings, "IP address is in private/internal range")
		}
	}

	result.IsValid = true
	return result
}

// ValidateJSON validates JSON structure and content
func (iv *InputValidator) ValidateJSON(jsonString string, maxDepth int) ValidationResult {
	result := ValidationResult{
		OriginalValue: jsonString,
	}

	if maxDepth == 0 {
		maxDepth = 10 // Default max depth
	}

	// Check for basic JSON validity
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid JSON: %v", err))
		return result
	}

	// Check JSON depth to prevent deeply nested attacks
	depth := calculateJSONDepth(jsonData, 0)
	if depth > maxDepth {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON depth %d exceeds maximum allowed depth %d", depth, maxDepth))
		return result
	}

	// Check JSON size
	if len(jsonString) > iv.config.MaxStringLength {
		result.IsValid = false
		result.Errors = append(result.Errors, "JSON string exceeds maximum size")
		return result
	}

	result.IsValid = true
	result.Value = jsonData
	return result
}

// ValidateNumeric validates numeric input
func (iv *InputValidator) ValidateNumeric(input string, rules NumericValidationRules) ValidationResult {
	result := ValidationResult{
		OriginalValue: input,
	}

	// Parse based on type
	var numValue float64
	var err error

	if rules.IsInteger {
		intVal, parseErr := strconv.ParseInt(input, 10, 64)
		if parseErr != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid integer format")
			return result
		}
		numValue = float64(intVal)
		result.Value = intVal
	} else {
		numValue, err = strconv.ParseFloat(input, 64)
		if err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid numeric format")
			return result
		}
		result.Value = numValue
	}

	// Validate range
	if rules.MinValue != nil && numValue < *rules.MinValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %.2f is below minimum %.2f", numValue, *rules.MinValue))
	}

	if rules.MaxValue != nil && numValue > *rules.MaxValue {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %.2f exceeds maximum %.2f", numValue, *rules.MaxValue))
	}

	if result.IsValid {
		result.IsValid = true
	}

	return result
}

// NumericValidationRules defines validation rules for numeric values
type NumericValidationRules struct {
	IsInteger bool
	MinValue  *float64
	MaxValue  *float64
}

// ValidateDateTime validates date and time strings
func (iv *InputValidator) ValidateDateTime(input string, format string) ValidationResult {
	result := ValidationResult{
		OriginalValue: input,
	}

	if format == "" {
		format = time.RFC3339 // Default to RFC3339
	}

	parsedTime, err := time.Parse(format, input)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid datetime format: %v", err))
		return result
	}

	// Check for reasonable date ranges in strict mode
	if iv.config.StrictMode {
		now := time.Now()
		hundredYearsAgo := now.AddDate(-100, 0, 0)
		hundredYearsFromNow := now.AddDate(100, 0, 0)

		if parsedTime.Before(hundredYearsAgo) || parsedTime.After(hundredYearsFromNow) {
			result.Warnings = append(result.Warnings, "Date is outside reasonable range")
		}
	}

	result.IsValid = true
	result.Value = parsedTime
	return result
}

// sanitizeString applies various sanitization techniques to strings
func (iv *InputValidator) sanitizeString(input string, level SanitizationLevel) string {
	switch level {
	case SanitizationNone:
		return input
		
	case SanitizationBasic:
		return iv.basicSanitization(input)
		
	case SanitizationStrict:
		return iv.strictSanitization(input)
		
	default:
		return iv.basicSanitization(input)
	}
}

// basicSanitization performs basic HTML escaping and whitespace normalization
func (iv *InputValidator) basicSanitization(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)
	
	// Normalize whitespace
	sanitized = strings.TrimSpace(sanitized)
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")
	
	return sanitized
}

// strictSanitization performs aggressive sanitization
func (iv *InputValidator) strictSanitization(input string) string {
	// Start with basic sanitization
	sanitized := iv.basicSanitization(input)
	
	// Remove potential script tags and event handlers
	scriptPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)on\w+\s*=`,
	}
	
	for _, pattern := range scriptPatterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "")
	}
	
	// Remove non-printable characters except common whitespace
	var result strings.Builder
	for _, r := range sanitized {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' || r == ' ' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// calculateJSONDepth calculates the nesting depth of JSON data
func calculateJSONDepth(data interface{}, currentDepth int) int {
	switch v := data.(type) {
	case map[string]interface{}:
		maxDepth := currentDepth
		for _, value := range v {
			depth := calculateJSONDepth(value, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth
		
	case []interface{}:
		maxDepth := currentDepth
		for _, value := range v {
			depth := calculateJSONDepth(value, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth
		
	default:
		return currentDepth
	}
}

// ValidateCreditCard validates credit card numbers using Luhn algorithm
func (iv *InputValidator) ValidateCreditCard(cardNumber string) ValidationResult {
	result := ValidationResult{
		OriginalValue: cardNumber,
	}

	// Remove spaces and hyphens
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Check if all characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Credit card number must contain only digits")
		return result
	}

	// Check length (13-19 digits for most cards)
	if len(cleaned) < 13 || len(cleaned) > 19 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Credit card number must be 13-19 digits long")
		return result
	}

	// Luhn algorithm validation
	if !iv.luhnCheck(cleaned) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Credit card number fails Luhn checksum validation")
		return result
	}

	result.IsValid = true
	result.Value = cleaned
	return result
}

// luhnCheck implements the Luhn algorithm for credit card validation
func (iv *InputValidator) luhnCheck(cardNumber string) bool {
	sum := 0
	isEven := false

	// Process digits from right to left
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// ValidatePhone validates phone numbers
func (iv *InputValidator) ValidatePhone(phoneNumber string) ValidationResult {
	result := ValidationResult{
		OriginalValue: phoneNumber,
	}

	// Remove common separators
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if all remaining characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Phone number must contain only digits and common separators")
		return result
	}

	// Check length (7-15 digits for international numbers)
	if len(cleaned) < 7 || len(cleaned) > 15 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Phone number must be 7-15 digits long")
		return result
	}

	result.IsValid = true
	result.Value = cleaned
	return result
}

// SanitizeFilename sanitizes filenames for safe filesystem operations
func (iv *InputValidator) SanitizeFilename(filename string) string {
	// Remove path separators
	sanitized := strings.ReplaceAll(filename, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "\\", "")
	
	// Remove dangerous characters
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}
	
	// Remove leading/trailing dots and spaces
	sanitized = strings.Trim(sanitized, ". ")
	
	// Limit length
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}
	
	// Ensure it's not empty or a reserved name
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperSanitized := strings.ToUpper(sanitized)
	for _, reserved := range reservedNames {
		if upperSanitized == reserved {
			sanitized = "file_" + sanitized
			break
		}
	}
	
	if sanitized == "" {
		sanitized = "untitled"
	}
	
	return sanitized
}

// ValidateBatch validates multiple inputs of the same type
func (iv *InputValidator) ValidateBatch(inputs []string, validateFunc func(string) ValidationResult) []ValidationResult {
	results := make([]ValidationResult, len(inputs))
	
	for i, input := range inputs {
		results[i] = validateFunc(input)
	}
	
	return results
}