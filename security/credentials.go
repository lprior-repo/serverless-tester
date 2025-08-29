package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// CredentialConfig holds configuration for credential management
type CredentialConfig struct {
	Type            CredentialType
	Name            string
	Description     string
	RotationDays    int
	SecretARN       string
	ParameterName   string
	EncryptionKeyID string
	Tags            map[string]string
}

// CredentialType represents the type of credential
type CredentialType string

const (
	CredentialTypeSecret    CredentialType = "secret"
	CredentialTypeParameter CredentialType = "parameter"
	CredentialTypeAPIKey    CredentialType = "api_key"
	CredentialTypeDatabase  CredentialType = "database"
	CredentialTypeOAuth     CredentialType = "oauth"
)

// Credential represents a managed credential
type Credential struct {
	ID          string
	Config      CredentialConfig
	Value       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   *time.Time
	IsActive    bool
	RotationDue bool
}

// CredentialManager manages secure credentials
type CredentialManager struct {
	secretsClient *secretsmanager.Client
	ssmClient     *ssm.Client
	encryptionKey *EncryptionKey
}

// SecretsManagerClient interface for AWS Secrets Manager operations
type SecretsManagerClient interface {
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	UpdateSecret(ctx context.Context, params *secretsmanager.UpdateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.UpdateSecretOutput, error)
	DeleteSecret(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
	RotateSecret(ctx context.Context, params *secretsmanager.RotateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.RotateSecretOutput, error)
}

// SSMClient interface for AWS Systems Manager Parameter Store operations
type SSMClient interface {
	PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	DeleteParameter(ctx context.Context, params *ssm.DeleteParameterInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParameterOutput, error)
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(secretsClient SecretsManagerClient, ssmClient SSMClient, encKey *EncryptionKey) *CredentialManager {
	return &CredentialManager{
		secretsClient: secretsClient.(*secretsmanager.Client),
		ssmClient:     ssmClient.(*ssm.Client),
		encryptionKey: encKey,
	}
}

// CreateSecret creates a new secret in AWS Secrets Manager
func (cm *CredentialManager) CreateSecret(ctx context.Context, config CredentialConfig, value string) (*Credential, error) {
	if err := ValidateCredentialConfig(config); err != nil {
		return nil, fmt.Errorf("invalid credential config: %w", err)
	}

	// Encrypt the secret value
	encResult, err := EncryptData(cm.encryptionKey, []byte(value))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
	}

	// Create secret in AWS Secrets Manager
	input := &secretsmanager.CreateSecretInput{
		Name:        aws.String(config.Name),
		Description: aws.String(config.Description),
		SecretString: aws.String(encResult.EncryptedData),
		KmsKeyId:    aws.String(config.EncryptionKeyID),
	}

	// Add tags if provided
	if len(config.Tags) > 0 {
		for key, val := range config.Tags {
			input.Tags = append(input.Tags, secretsmanager.Tag{
				Key:   aws.String(key),
				Value: aws.String(val),
			})
		}
	}

	output, err := cm.secretsClient.CreateSecret(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	credential := &Credential{
		ID:        *output.ARN,
		Config:    config,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Set expiration if rotation is configured
	if config.RotationDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, config.RotationDays)
		credential.ExpiresAt = &expiresAt
	}

	config.SecretARN = *output.ARN
	credential.Config = config

	return credential, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (cm *CredentialManager) GetSecret(ctx context.Context, secretARN string) (*Credential, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	}

	output, err := cm.secretsClient.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Create encryption result from stored data
	encResult := &EncryptionResult{
		EncryptedData: *output.SecretString,
		KeyID:         cm.encryptionKey.ID,
		Algorithm:     cm.encryptionKey.Config.Algorithm,
	}

	// Decrypt the secret value
	decResult, err := DecryptData(cm.encryptionKey, encResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret value: %w", err)
	}

	credential := &Credential{
		ID:        secretARN,
		Value:     string(decResult.DecryptedData),
		CreatedAt: *output.CreatedDate,
		IsActive:  true,
	}

	return credential, nil
}

// RotateSecret rotates a secret in AWS Secrets Manager
func (cm *CredentialManager) RotateSecret(ctx context.Context, secretARN string, newValue string) error {
	// Encrypt the new secret value
	encResult, err := EncryptData(cm.encryptionKey, []byte(newValue))
	if err != nil {
		return fmt.Errorf("failed to encrypt new secret value: %w", err)
	}

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretARN),
		SecretString: aws.String(encResult.EncryptedData),
	}

	_, err = cm.secretsClient.UpdateSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to rotate secret: %w", err)
	}

	return nil
}

// CreateParameter creates a new parameter in AWS Systems Manager Parameter Store
func (cm *CredentialManager) CreateParameter(ctx context.Context, config CredentialConfig, value string) (*Credential, error) {
	if err := ValidateCredentialConfig(config); err != nil {
		return nil, fmt.Errorf("invalid credential config: %w", err)
	}

	// Encrypt the parameter value
	encResult, err := EncryptData(cm.encryptionKey, []byte(value))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt parameter value: %w", err)
	}

	input := &ssm.PutParameterInput{
		Name:        aws.String(config.ParameterName),
		Description: aws.String(config.Description),
		Value:       aws.String(encResult.EncryptedData),
		Type:        ssm.ParameterTypeSecureString,
		KeyId:       aws.String(config.EncryptionKeyID),
	}

	// Add tags if provided
	if len(config.Tags) > 0 {
		for key, val := range config.Tags {
			input.Tags = append(input.Tags, ssm.Tag{
				Key:   aws.String(key),
				Value: aws.String(val),
			})
		}
	}

	_, err = cm.ssmClient.PutParameter(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create parameter: %w", err)
	}

	credential := &Credential{
		ID:        config.ParameterName,
		Config:    config,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Set expiration if rotation is configured
	if config.RotationDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, config.RotationDays)
		credential.ExpiresAt = &expiresAt
	}

	return credential, nil
}

// GetParameter retrieves a parameter from AWS Systems Manager Parameter Store
func (cm *CredentialManager) GetParameter(ctx context.Context, parameterName string) (*Credential, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(false), // We handle our own encryption
	}

	output, err := cm.ssmClient.GetParameter(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameter: %w", err)
	}

	// Create encryption result from stored data
	encResult := &EncryptionResult{
		EncryptedData: *output.Parameter.Value,
		KeyID:         cm.encryptionKey.ID,
		Algorithm:     cm.encryptionKey.Config.Algorithm,
	}

	// Decrypt the parameter value
	decResult, err := DecryptData(cm.encryptionKey, encResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt parameter value: %w", err)
	}

	credential := &Credential{
		ID:        parameterName,
		Value:     string(decResult.DecryptedData),
		CreatedAt: *output.Parameter.LastModifiedDate,
		IsActive:  true,
	}

	return credential, nil
}

// ValidateCredentialConfig validates credential configuration
func ValidateCredentialConfig(config CredentialConfig) error {
	if config.Name == "" {
		return fmt.Errorf("credential name cannot be empty")
	}

	if !IsValidCredentialName(config.Name) {
		return fmt.Errorf("invalid credential name format: %s", config.Name)
	}

	if config.Type == "" {
		return fmt.Errorf("credential type cannot be empty")
	}

	validTypes := map[CredentialType]bool{
		CredentialTypeSecret:    true,
		CredentialTypeParameter: true,
		CredentialTypeAPIKey:    true,
		CredentialTypeDatabase:  true,
		CredentialTypeOAuth:     true,
	}

	if !validTypes[config.Type] {
		return fmt.Errorf("invalid credential type: %s", config.Type)
	}

	if config.RotationDays < 0 {
		return fmt.Errorf("rotation days cannot be negative: %d", config.RotationDays)
	}

	if config.RotationDays > 0 && config.RotationDays < 1 {
		return fmt.Errorf("rotation days must be at least 1 if specified: %d", config.RotationDays)
	}

	return nil
}

// IsValidCredentialName checks if a credential name follows naming conventions
func IsValidCredentialName(name string) bool {
	if len(name) == 0 || len(name) > 512 {
		return false
	}

	// AWS Secrets Manager naming pattern: alphanumeric, hyphens, underscores, periods, forward slashes
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	return pattern.MatchString(name)
}

// SanitizeCredentialName sanitizes a credential name for AWS services
func SanitizeCredentialName(name string) string {
	// Replace invalid characters with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "@", "_")
	sanitized = strings.ReplaceAll(sanitized, "#", "_")
	sanitized = strings.ReplaceAll(sanitized, "%", "_")
	
	// Remove any remaining invalid characters
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || 
		   (char >= 'A' && char <= 'Z') || 
		   (char >= '0' && char <= '9') || 
		   char == '_' || char == '-' || char == '.' || char == '/' {
			result.WriteRune(char)
		}
	}
	
	final := result.String()
	if len(final) == 0 {
		final = "default_credential"
	}
	
	return final
}

// CheckCredentialRotation checks if credentials need rotation
func CheckCredentialRotation(credential *Credential) bool {
	if credential.ExpiresAt == nil {
		return false
	}

	now := time.Now()
	
	// Check if credential has expired
	if now.After(*credential.ExpiresAt) {
		return true
	}

	// Check if within rotation warning period (7 days)
	warningPeriod := credential.ExpiresAt.Add(-7 * 24 * time.Hour)
	if now.After(warningPeriod) {
		return true
	}

	return false
}

// GenerateSecurePassword generates a secure password with specified criteria
func GenerateSecurePassword(length int, includeSymbols bool) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		symbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	var charset string = lowercase + uppercase + digits
	if includeSymbols {
		charset += symbols
	}

	password := make([]byte, length)
	for i := range password {
		randIndex := make([]byte, 1)
		if _, err := rand.Read(randIndex); err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = charset[int(randIndex[0])%len(charset)]
	}

	return string(password), nil
}

// ValidatePasswordStrength validates password strength according to security policies
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`).MatchString(password)

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if !hasSymbol {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common patterns
	commonPatterns := []string{
		"password", "123456", "qwerty", "admin", "letmein",
		"welcome", "monkey", "dragon", "pass", "master",
	}

	lowerPassword := strings.ToLower(password)
	for _, pattern := range commonPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return fmt.Errorf("password contains common pattern: %s", pattern)
		}
	}

	return nil
}