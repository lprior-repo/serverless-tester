// Package security provides comprehensive security features and compliance validation
// for the serverless testing framework. It implements:
//
//   - Advanced encryption with key rotation
//   - Secure credential management
//   - Network security validation
//   - Input validation and sanitization
//   - Compliance framework validation (SOX, HIPAA, GDPR, PCI-DSS)
//   - Security testing utilities
//   - Audit logging and governance
//
// The package follows strict functional programming principles and integrates
// seamlessly with the existing testing framework architecture.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

// EncryptionConfig holds configuration for encryption operations
type EncryptionConfig struct {
	Algorithm    string
	KeySize      int
	RotationDays int
	KeyID        string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// EncryptionKey represents an encryption key with metadata
type EncryptionKey struct {
	ID        string
	Key       []byte
	Config    EncryptionConfig
	CreatedAt time.Time
	ExpiresAt time.Time
	IsActive  bool
}

// EncryptionResult holds the result of an encryption operation
type EncryptionResult struct {
	EncryptedData string
	KeyID         string
	Algorithm     string
	Nonce         string
	CreatedAt     time.Time
}

// DecryptionResult holds the result of a decryption operation
type DecryptionResult struct {
	DecryptedData []byte
	KeyID         string
	Algorithm     string
	DecryptedAt   time.Time
}

// KeyRotationSchedule defines when and how keys should be rotated
type KeyRotationSchedule struct {
	Interval      time.Duration
	NextRotation  time.Time
	AutoRotate    bool
	NotifyBefore  time.Duration
	WarningThreshold time.Duration
}

// CreateEncryptionKey generates a new encryption key with the specified configuration
func CreateEncryptionKey(config EncryptionConfig) (*EncryptionKey, error) {
	if config.KeySize <= 0 {
		config.KeySize = 32 // Default to 256-bit key
	}

	if config.Algorithm == "" {
		config.Algorithm = "AES-256-GCM"
	}

	if config.RotationDays <= 0 {
		config.RotationDays = 90 // Default rotation every 90 days
	}

	key := make([]byte, config.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	keyID := generateKeyID()
	createdAt := time.Now()
	expiresAt := createdAt.AddDate(0, 0, config.RotationDays)

	config.KeyID = keyID
	config.CreatedAt = createdAt
	config.ExpiresAt = expiresAt

	return &EncryptionKey{
		ID:        keyID,
		Key:       key,
		Config:    config,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}, nil
}

// EncryptData encrypts data using AES-256-GCM
func EncryptData(key *EncryptionKey, plaintext []byte) (*EncryptionResult, error) {
	if !key.IsActive {
		return nil, fmt.Errorf("encryption key is not active")
	}

	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("encryption key has expired")
	}

	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	encodedData := base64.StdEncoding.EncodeToString(ciphertext)
	encodedNonce := base64.StdEncoding.EncodeToString(nonce)

	return &EncryptionResult{
		EncryptedData: encodedData,
		KeyID:         key.ID,
		Algorithm:     key.Config.Algorithm,
		Nonce:         encodedNonce,
		CreatedAt:     time.Now(),
	}, nil
}

// DecryptData decrypts data using AES-256-GCM
func DecryptData(key *EncryptionKey, encResult *EncryptionResult) (*DecryptionResult, error) {
	if key.ID != encResult.KeyID {
		return nil, fmt.Errorf("key ID mismatch")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encResult.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return &DecryptionResult{
		DecryptedData: plaintext,
		KeyID:         key.ID,
		Algorithm:     encResult.Algorithm,
		DecryptedAt:   time.Now(),
	}, nil
}

// RotateKey creates a new key to replace an existing key
func RotateKey(oldKey *EncryptionKey) (*EncryptionKey, error) {
	config := oldKey.Config
	config.CreatedAt = time.Time{}
	config.ExpiresAt = time.Time{}
	config.KeyID = ""

	newKey, err := CreateEncryptionKey(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create rotated key: %w", err)
	}

	// Mark old key as inactive
	oldKey.IsActive = false

	return newKey, nil
}

// ValidateKeyRotation checks if a key needs rotation
func ValidateKeyRotation(key *EncryptionKey, schedule KeyRotationSchedule) bool {
	now := time.Now()
	
	// Check if key has expired
	if now.After(key.ExpiresAt) {
		return true
	}

	// Check if within warning threshold
	warningTime := key.ExpiresAt.Add(-schedule.WarningThreshold)
	if now.After(warningTime) {
		return true
	}

	// Check if scheduled rotation is due
	if schedule.AutoRotate && now.After(schedule.NextRotation) {
		return true
	}

	return false
}

// CreateRotationSchedule creates a key rotation schedule
func CreateRotationSchedule(interval time.Duration, autoRotate bool) KeyRotationSchedule {
	now := time.Now()
	return KeyRotationSchedule{
		Interval:         interval,
		NextRotation:     now.Add(interval),
		AutoRotate:       autoRotate,
		NotifyBefore:     24 * time.Hour, // Notify 24 hours before rotation
		WarningThreshold: 7 * 24 * time.Hour, // Warning 7 days before expiration
	}
}

// generateKeyID creates a unique key identifier
func generateKeyID() string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("key_%x", hash.Sum(nil))[:16]
}

// ValidateEncryptionConfig validates encryption configuration parameters
func ValidateEncryptionConfig(config EncryptionConfig) error {
	if config.KeySize < 16 {
		return fmt.Errorf("key size must be at least 16 bytes, got %d", config.KeySize)
	}

	if config.KeySize%8 != 0 {
		return fmt.Errorf("key size must be multiple of 8, got %d", config.KeySize)
	}

	validAlgorithms := map[string]bool{
		"AES-128-GCM": true,
		"AES-192-GCM": true,
		"AES-256-GCM": true,
	}

	if config.Algorithm != "" && !validAlgorithms[config.Algorithm] {
		return fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
	}

	if config.RotationDays < 1 {
		return fmt.Errorf("rotation days must be positive, got %d", config.RotationDays)
	}

	if config.RotationDays > 365 {
		return fmt.Errorf("rotation days should not exceed 365, got %d", config.RotationDays)
	}

	return nil
}

// SecureEraseKey securely overwrites key material in memory
func SecureEraseKey(key *EncryptionKey) {
	if key == nil || key.Key == nil {
		return
	}

	// Overwrite key material with random data
	rand.Read(key.Key)
	
	// Clear key reference
	key.Key = nil
	key.IsActive = false
}