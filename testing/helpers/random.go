package helpers

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/random"
)

// randomSource holds the random source for deterministic testing
var randomSource *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// UniqueID generates a unique identifier using Terratest's random module
// Prioritizing Terratest patterns for consistency across all test infrastructure
func UniqueID() string {
	return random.UniqueId()
}

// UUIDUniqueId generates a unique identifier using UUID (legacy support)
// This is maintained for backward compatibility but UniqueID() is preferred
func UUIDUniqueId() string {
	return uuid.New().String()
}

// TerratestUniqueId generates a unique identifier using Terratest's random module
// This is now the same as UniqueID() for consistency
func TerratestUniqueId() string {
	return random.UniqueId()
}

// RandomResourceName generates a random resource name with prefix using Terratest
func RandomResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, random.UniqueId())
}

// RandomString generates a random string of specified length using alphanumeric characters
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}
	
	// Generate random string with alphanumeric characters
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[randomSource.Intn(len(charset))]
	}
	return string(result)
}

// RandomStringLower generates a random lowercase string of specified length
func RandomStringLower(length int) string {
	if length <= 0 {
		return ""
	}
	
	return strings.ToLower(RandomString(length))
}

// RandomInt generates a random integer between min and max (inclusive)
func RandomInt(min, max int) int {
	if min == max {
		return min
	}
	
	if min > max {
		min, max = max, min
	}
	
	return randomSource.Intn(max-min+1) + min
}

// RandomAWSResourceName generates a random AWS resource name with given prefix
// Follows AWS naming conventions: alphanumeric characters and hyphens only
func RandomAWSResourceName(prefix string) string {
	if prefix == "" {
		prefix = "resource"
	}
	
	// Generate a random suffix
	suffix := RandomStringLower(8)
	
	// Combine prefix and suffix with hyphen
	return fmt.Sprintf("%s-%s", prefix, suffix)
}

// TimestampID generates a unique ID based on timestamp
func TimestampID() string {
	timestamp := time.Now().UnixNano()
	randomPart := RandomString(4)
	return fmt.Sprintf("%d-%s", timestamp, randomPart)
}

// TerratestTimestampID generates a unique ID using Terratest's random module
func TerratestTimestampID() string {
	return fmt.Sprintf("ts-%s", random.UniqueId())
}

// SeedRandom sets the random seed for deterministic testing
func SeedRandom(seed int64) {
	randomSource = rand.New(rand.NewSource(seed))
}