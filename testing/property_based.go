package testutils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

// PropertyBasedTest provides property-based testing capabilities
type PropertyBasedTest struct {
	rng *RandomGenerator
}

// StringGenerator interface for string generation
type StringGenerator interface {
	Generate() string
	Shrink(value string) []string
}

// IntGenerator interface for integer generation
type IntGenerator interface {
	Generate() int
	Shrink(value int) []int
}

// ArnGenerator interface for ARN generation
type ArnGeneratorInterface interface {
	Generate() string
	Shrink(value string) []string
}

// LambdaNameGeneratorInterface interface for Lambda name generation
type LambdaNameGeneratorInterface interface {
	Generate() string
	Shrink(value string) []string
}

// EventGeneratorInterface interface for event generation
type EventGeneratorInterface interface {
	Generate() TestEvent
	Shrink(value TestEvent) []TestEvent
}

// PropertyResult represents the result of a property test
type PropertyResult struct {
	Passed         bool
	TestsRun       int
	CounterExample string
	ShrunkExample  string
	Error          error
}

// TestEvent represents a test event structure
type TestEvent struct {
	EventID   string
	Source    string
	Timestamp int64
	Payload   map[string]interface{}
}

// RandomGenerator provides cryptographically secure random generation
type RandomGenerator struct {
	seed int64
}

// NewPropertyBasedTest creates a new property-based testing instance
func NewPropertyBasedTest() *PropertyBasedTest {
	return &PropertyBasedTest{
		rng: &RandomGenerator{seed: time.Now().UnixNano()},
	}
}

// StringGenerator creates a generator for strings of specified length range
func (p *PropertyBasedTest) StringGenerator(minLen, maxLen int) StringGenerator {
	return &stringGenerator{
		rng:    p.rng,
		minLen: minLen,
		maxLen: maxLen,
	}
}

// IntGenerator creates a generator for integers in specified range
func (p *PropertyBasedTest) IntGenerator(min, max int) IntGenerator {
	return &intGenerator{
		rng: p.rng,
		min: min,
		max: max,
	}
}

// ArnGenerator creates a generator for AWS ARN strings
func (p *PropertyBasedTest) ArnGenerator() ArnGeneratorInterface {
	return &arnGenerator{
		rng: p.rng,
	}
}

// LambdaFunctionNameGenerator creates a generator for Lambda function names
func (p *PropertyBasedTest) LambdaFunctionNameGenerator() LambdaNameGeneratorInterface {
	return &lambdaNameGenerator{
		rng: p.rng,
	}
}

// EventGenerator creates a generator for test events
func (p *PropertyBasedTest) EventGenerator() EventGeneratorInterface {
	return &eventGenerator{
		rng: p.rng,
	}
}

// CheckStringProperty runs property-based tests for string properties
func (p *PropertyBasedTest) CheckStringProperty(
	name string,
	gen StringGenerator,
	property func(string) bool,
	iterations int,
) PropertyResult {
	for i := 0; i < iterations; i++ {
		value := gen.Generate()
		if !property(value) {
			// Property failed, try to shrink the counter-example
			shrunk := p.shrinkStringValue(gen, value, property)
			return PropertyResult{
				Passed:         false,
				TestsRun:       i + 1,
				CounterExample: fmt.Sprintf("%v", value),
				ShrunkExample:  fmt.Sprintf("%v", shrunk),
			}
		}
	}
	
	return PropertyResult{
		Passed:   true,
		TestsRun: iterations,
	}
}

// CheckProperty is a generic wrapper for different property types (for backwards compatibility)
func (p *PropertyBasedTest) CheckProperty(
	name string,
	generator interface{},
	property interface{},
	iterations int,
) PropertyResult {
	// Type switch to handle different generator types
	switch gen := generator.(type) {
	case StringGenerator:
		if prop, ok := property.(func(string) bool); ok {
			return p.CheckStringProperty(name, gen, prop, iterations)
		}
	}
	
	return PropertyResult{
		Passed: false,
		Error:  fmt.Errorf("unsupported generator or property type"),
	}
}

// shrinkStringValue attempts to find a smaller string counter-example
func (p *PropertyBasedTest) shrinkStringValue(
	gen StringGenerator,
	value string,
	property func(string) bool,
) string {
	return p.shrinkStringValueWithDepth(gen, value, property, 0, 100)
}

// shrinkStringValueWithDepth attempts to find a smaller string counter-example with depth limit
func (p *PropertyBasedTest) shrinkStringValueWithDepth(
	gen StringGenerator,
	value string,
	property func(string) bool,
	depth int,
	maxDepth int,
) string {
	if depth >= maxDepth {
		return value
	}
	
	candidates := gen.Shrink(value)
	for _, candidate := range candidates {
		if !property(candidate) {
			// Found a smaller counter-example, continue shrinking
			return p.shrinkStringValueWithDepth(gen, candidate, property, depth+1, maxDepth)
		}
	}
	return value
}

// secureRandomInt generates a cryptographically secure random integer
func (r *RandomGenerator) secureRandomInt(min, max int) int {
	if min >= max {
		return min
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return int(time.Now().UnixNano())%(max-min+1) + min
	}
	
	return int(n.Int64()) + min
}

// secureRandomString generates a cryptographically secure random string
func (r *RandomGenerator) secureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to simple method if crypto/rand fails
			result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			result[i] = charset[n.Int64()]
		}
	}
	
	return string(result)
}

// ========== Generator Implementations ==========

// stringGenerator implements string generation
type stringGenerator struct {
	rng    *RandomGenerator
	minLen int
	maxLen int
}

func (g *stringGenerator) Generate() string {
	length := g.rng.secureRandomInt(g.minLen, g.maxLen)
	return g.rng.secureRandomString(length)
}

func (g *stringGenerator) Shrink(value string) []string {
	if len(value) <= 1 {
		return []string{}
	}
	
	var candidates []string
	// Try empty string
	candidates = append(candidates, "")
	
	// Try half the length
	if len(value) > 2 {
		candidates = append(candidates, value[:len(value)/2])
	}
	
	// Try removing one character
	if len(value) > 1 {
		candidates = append(candidates, value[:len(value)-1])
	}
	
	// Try simpler string with same length
	if len(value) > 0 && value != strings.Repeat("a", len(value)) {
		candidates = append(candidates, strings.Repeat("a", len(value)))
	}
	
	return candidates
}

// intGenerator implements integer generation
type intGenerator struct {
	rng *RandomGenerator
	min int
	max int
}

func (g *intGenerator) Generate() int {
	return g.rng.secureRandomInt(g.min, g.max)
}

func (g *intGenerator) Shrink(value int) []int {
	var candidates []int
	
	// Shrink towards zero
	if value > 0 {
		candidates = append(candidates, 0)
		candidates = append(candidates, value/2)
	} else if value < 0 {
		candidates = append(candidates, 0)
		candidates = append(candidates, value/2)
	}
	
	// Try smaller absolute values
	if math.Abs(float64(value)) > 1 {
		if value > 0 {
			candidates = append(candidates, value-1)
		} else {
			candidates = append(candidates, value+1)
		}
	}
	
	return candidates
}

// arnGenerator implements AWS ARN generation
type arnGenerator struct {
	rng *RandomGenerator
}

func (g *arnGenerator) Generate() string {
	services := []string{"lambda", "dynamodb", "s3", "iam", "ec2", "rds"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	
	service := services[g.rng.secureRandomInt(0, len(services)-1)]
	region := regions[g.rng.secureRandomInt(0, len(regions)-1)]
	accountId := fmt.Sprintf("%012d", g.rng.secureRandomInt(100000000000, 999999999999))
	resourceType := g.rng.secureRandomString(8)
	resourceId := g.rng.secureRandomString(16)
	
	return fmt.Sprintf("arn:aws:%s:%s:%s:%s/%s", service, region, accountId, resourceType, resourceId)
}

func (g *arnGenerator) Shrink(value string) []string {
	// For ARNs, we can try simpler resource names
	parts := strings.Split(value, "/")
	if len(parts) > 1 {
		return []string{
			strings.Join(parts[:len(parts)-1], "/") + "/test",
			strings.Join(parts[:len(parts)-1], "/") + "/a",
		}
	}
	return []string{}
}

// lambdaNameGenerator implements Lambda function name generation
type lambdaNameGenerator struct {
	rng *RandomGenerator
}

func (g *lambdaNameGenerator) Generate() string {
	length := g.rng.secureRandomInt(1, 64) // AWS Lambda name limits
	name := g.rng.secureRandomString(length)
	
	// Ensure valid Lambda naming (alphanumeric, hyphens, underscores)
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	result := make([]byte, len(name))
	for i, char := range name {
		if strings.ContainsRune(validChars, char) {
			result[i] = byte(char)
		} else {
			result[i] = validChars[g.rng.secureRandomInt(0, len(validChars)-1)]
		}
	}
	
	return string(result)
}

func (g *lambdaNameGenerator) Shrink(value string) []string {
	if len(value) <= 1 {
		return []string{}
	}
	
	return []string{
		value[:len(value)/2],
		"test",
		"a",
	}
}

// eventGenerator implements test event generation
type eventGenerator struct {
	rng *RandomGenerator
}

func (g *eventGenerator) Generate() TestEvent {
	sources := []string{"aws.s3", "aws.lambda", "aws.dynamodb", "custom.service"}
	source := sources[g.rng.secureRandomInt(0, len(sources)-1)]
	
	return TestEvent{
		EventID:   g.rng.secureRandomString(16),
		Source:    source,
		Timestamp: time.Now().Unix() + int64(g.rng.secureRandomInt(-86400, 86400)), // +/- 1 day
		Payload: map[string]interface{}{
			"key1": g.rng.secureRandomString(10),
			"key2": g.rng.secureRandomInt(1, 1000),
		},
	}
}

func (g *eventGenerator) Shrink(value TestEvent) []TestEvent {
	return []TestEvent{
		{
			EventID:   "test",
			Source:    "test.source",
			Timestamp: 0,
			Payload:   map[string]interface{}{"key": "value"},
		},
	}
}