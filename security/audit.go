package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditLogger provides comprehensive audit logging capabilities
type AuditLogger struct {
	config     AuditConfig
	writer     AuditWriter
	mutex      sync.RWMutex
	session    *AuditSession
	processors []AuditProcessor
}

// AuditConfig holds configuration for audit logging
type AuditConfig struct {
	EnableAuditing        bool
	LogLevel              AuditLevel
	OutputFormat          AuditFormat
	OutputDestination     string
	RotateLogFiles       bool
	MaxLogFileSize       int64 // in bytes
	MaxLogFiles          int
	EnableEncryption     bool
	EncryptionKey        *EncryptionKey
	EnableCompression    bool
	BufferSize           int
	FlushInterval        time.Duration
	EnableRemoteLogging  bool
	RemoteEndpoint       string
	EnableIntegrityCheck bool
}

// AuditLevel represents the level of audit logging
type AuditLevel string

const (
	AuditLevelDebug   AuditLevel = "DEBUG"
	AuditLevelInfo    AuditLevel = "INFO"
	AuditLevelWarn    AuditLevel = "WARN"
	AuditLevelError   AuditLevel = "ERROR"
	AuditLevelCritical AuditLevel = "CRITICAL"
)

// AuditFormat represents the format of audit logs
type AuditFormat string

const (
	AuditFormatJSON AuditFormat = "json"
	AuditFormatCSV  AuditFormat = "csv"
	AuditFormatSyslog AuditFormat = "syslog"
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Level         AuditLevel             `json:"level"`
	Category      AuditCategory          `json:"category"`
	Action        string                 `json:"action"`
	Subject       string                 `json:"subject"` // Who performed the action
	Object        string                 `json:"object"`  // What was acted upon
	Result        AuditResult            `json:"result"`
	Details       map[string]interface{} `json:"details"`
	Source        AuditSource            `json:"source"`
	SessionID     string                 `json:"session_id"`
	RequestID     string                 `json:"request_id"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	Location      string                 `json:"location,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Compliance    []ComplianceFramework  `json:"compliance,omitempty"`
	Hash          string                 `json:"hash,omitempty"` // For integrity checking
}

// AuditCategory represents the category of audit event
type AuditCategory string

const (
	AuditCategoryAuthentication AuditCategory = "authentication"
	AuditCategoryAuthorization  AuditCategory = "authorization"
	AuditCategoryDataAccess     AuditCategory = "data_access"
	AuditCategoryDataModification AuditCategory = "data_modification"
	AuditCategorySystem         AuditCategory = "system"
	AuditCategorySecurity       AuditCategory = "security"
	AuditCategoryCompliance     AuditCategory = "compliance"
	AuditCategoryConfiguration  AuditCategory = "configuration"
	AuditCategoryError          AuditCategory = "error"
)

// AuditResult represents the result of an audited action
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultDenied  AuditResult = "denied"
	AuditResultWarning AuditResult = "warning"
)

// AuditSource contains information about the source of the audit event
type AuditSource struct {
	Service     string `json:"service"`
	Component   string `json:"component"`
	Function    string `json:"function"`
	Filename    string `json:"filename,omitempty"`
	LineNumber  int    `json:"line_number,omitempty"`
	Version     string `json:"version,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// AuditSession represents an audit session
type AuditSession struct {
	ID        string
	StartTime time.Time
	Subject   string
	Context   map[string]interface{}
	Active    bool
}

// AuditWriter interface for writing audit logs
type AuditWriter interface {
	Write(event AuditEvent) error
	Flush() error
	Close() error
}

// AuditProcessor interface for processing audit events
type AuditProcessor interface {
	Process(event AuditEvent) (AuditEvent, error)
	Name() string
}

// GovernancePolicy represents a governance policy
type GovernancePolicy struct {
	ID            string
	Name          string
	Description   string
	Category      string
	Rules         []GovernanceRule
	Compliance    []ComplianceFramework
	Severity      ComplianceSeverity
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	ApprovalStatus string
	Version       string
}

// GovernanceRule represents a single governance rule
type GovernanceRule struct {
	ID          string
	Name        string
	Description string
	Condition   string
	Action      string
	Parameters  map[string]interface{}
	Active      bool
}

// GovernanceEngine manages governance policies and compliance
type GovernanceEngine struct {
	policies   map[string]*GovernancePolicy
	auditLog   *AuditLogger
	mutex      sync.RWMutex
	validators map[string]func(interface{}) error
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditConfig) (*AuditLogger, error) {
	if !config.EnableAuditing {
		return &AuditLogger{config: config}, nil
	}

	// Set defaults
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.MaxLogFileSize == 0 {
		config.MaxLogFileSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxLogFiles == 0 {
		config.MaxLogFiles = 10
	}

	// Create audit writer
	writer, err := createAuditWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit writer: %w", err)
	}

	logger := &AuditLogger{
		config:  config,
		writer:  writer,
		session: &AuditSession{},
	}

	// Add default processors
	logger.AddProcessor(&HashProcessor{})
	logger.AddProcessor(&ComplianceProcessor{})

	return logger, nil
}

// StartSession starts a new audit session
func (al *AuditLogger) StartSession(subject string, context map[string]interface{}) string {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	sessionID := generateSessionID()
	al.session = &AuditSession{
		ID:        sessionID,
		StartTime: time.Now(),
		Subject:   subject,
		Context:   context,
		Active:    true,
	}

	// Log session start
	al.logEvent(AuditEvent{
		Level:     AuditLevelInfo,
		Category:  AuditCategorySystem,
		Action:    "session_start",
		Subject:   subject,
		SessionID: sessionID,
		Details:   context,
	})

	return sessionID
}

// EndSession ends the current audit session
func (al *AuditLogger) EndSession() {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	if al.session != nil && al.session.Active {
		// Log session end
		al.logEvent(AuditEvent{
			Level:     AuditLevelInfo,
			Category:  AuditCategorySystem,
			Action:    "session_end",
			Subject:   al.session.Subject,
			SessionID: al.session.ID,
			Details: map[string]interface{}{
				"duration": time.Since(al.session.StartTime).String(),
			},
		})

		al.session.Active = false
	}
}

// LogAuthentication logs authentication events
func (al *AuditLogger) LogAuthentication(subject, action string, result AuditResult, details map[string]interface{}) {
	al.logEvent(AuditEvent{
		Level:    al.getLevelForResult(result),
		Category: AuditCategoryAuthentication,
		Action:   action,
		Subject:  subject,
		Result:   result,
		Details:  details,
		Compliance: []ComplianceFramework{
			ComplianceSOX, ComplianceHIPAA, ComplianceGDPR, CompliancePCIDSS,
		},
	})
}

// LogDataAccess logs data access events
func (al *AuditLogger) LogDataAccess(subject, object, action string, result AuditResult, details map[string]interface{}) {
	al.logEvent(AuditEvent{
		Level:    al.getLevelForResult(result),
		Category: AuditCategoryDataAccess,
		Action:   action,
		Subject:  subject,
		Object:   object,
		Result:   result,
		Details:  details,
		Compliance: []ComplianceFramework{
			ComplianceHIPAA, ComplianceGDPR, ComplianceSOX,
		},
	})
}

// LogDataModification logs data modification events
func (al *AuditLogger) LogDataModification(subject, object, action string, result AuditResult, before, after interface{}) {
	details := map[string]interface{}{
		"before": before,
		"after":  after,
	}

	al.logEvent(AuditEvent{
		Level:    al.getLevelForResult(result),
		Category: AuditCategoryDataModification,
		Action:   action,
		Subject:  subject,
		Object:   object,
		Result:   result,
		Details:  details,
		Compliance: []ComplianceFramework{
			ComplianceSOX, ComplianceHIPAA, ComplianceGDPR,
		},
	})
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(action string, result AuditResult, details map[string]interface{}) {
	al.logEvent(AuditEvent{
		Level:    AuditLevelCritical,
		Category: AuditCategorySecurity,
		Action:   action,
		Result:   result,
		Details:  details,
		Compliance: []ComplianceFramework{
			ComplianceSOX, ComplianceHIPAA, ComplianceGDPR, CompliancePCIDSS,
		},
	})
}

// LogComplianceEvent logs compliance-related events
func (al *AuditLogger) LogComplianceEvent(framework ComplianceFramework, action string, result AuditResult, details map[string]interface{}) {
	al.logEvent(AuditEvent{
		Level:      al.getLevelForResult(result),
		Category:   AuditCategoryCompliance,
		Action:     action,
		Result:     result,
		Details:    details,
		Compliance: []ComplianceFramework{framework},
	})
}

// LogError logs error events
func (al *AuditLogger) LogError(action, message string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()
	details["message"] = message

	al.logEvent(AuditEvent{
		Level:    AuditLevelError,
		Category: AuditCategoryError,
		Action:   action,
		Result:   AuditResultFailure,
		Details:  details,
	})
}

// logEvent is the internal method for logging events
func (al *AuditLogger) logEvent(event AuditEvent) {
	if !al.config.EnableAuditing {
		return
	}

	// Set default values
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.SessionID == "" && al.session != nil && al.session.Active {
		event.SessionID = al.session.ID
	}

	// Process the event through processors
	processedEvent := event
	for _, processor := range al.processors {
		var err error
		processedEvent, err = processor.Process(processedEvent)
		if err != nil {
			// Log processor error but continue
			fmt.Printf("Audit processor error: %v\n", err)
		}
	}

	// Write the event
	if al.writer != nil {
		if err := al.writer.Write(processedEvent); err != nil {
			fmt.Printf("Audit write error: %v\n", err)
		}
	}
}

// AddProcessor adds an audit event processor
func (al *AuditLogger) AddProcessor(processor AuditProcessor) {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	al.processors = append(al.processors, processor)
}

// getLevelForResult determines the audit level based on the result
func (al *AuditLogger) getLevelForResult(result AuditResult) AuditLevel {
	switch result {
	case AuditResultSuccess:
		return AuditLevelInfo
	case AuditResultWarning:
		return AuditLevelWarn
	case AuditResultFailure, AuditResultDenied:
		return AuditLevelError
	default:
		return AuditLevelInfo
	}
}

// Flush flushes the audit writer
func (al *AuditLogger) Flush() error {
	if al.writer != nil {
		return al.writer.Flush()
	}
	return nil
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	if al.writer != nil {
		return al.writer.Close()
	}
	return nil
}

// FileAuditWriter implements file-based audit writing
type FileAuditWriter struct {
	config   AuditConfig
	file     *os.File
	encoder  *json.Encoder
	mutex    sync.Mutex
	buffer   []AuditEvent
	size     int64
}

// Write writes an audit event to file
func (faw *FileAuditWriter) Write(event AuditEvent) error {
	faw.mutex.Lock()
	defer faw.mutex.Unlock()

	// Add to buffer
	faw.buffer = append(faw.buffer, event)

	// Check if we need to flush
	if len(faw.buffer) >= faw.config.BufferSize {
		return faw.flushBuffer()
	}

	return nil
}

// flushBuffer flushes the event buffer to file
func (faw *FileAuditWriter) flushBuffer() error {
	if len(faw.buffer) == 0 {
		return nil
	}

	// Check if we need to rotate the log file
	if faw.config.RotateLogFiles && faw.size >= faw.config.MaxLogFileSize {
		if err := faw.rotateLog(); err != nil {
			return err
		}
	}

	// Write events to file
	for _, event := range faw.buffer {
		if err := faw.encoder.Encode(event); err != nil {
			return err
		}
		
		// Estimate size increase
		faw.size += int64(len(event.ID) + 500) // Rough estimate
	}

	// Clear buffer
	faw.buffer = nil

	return faw.file.Sync()
}

// rotateLog rotates the current log file
func (faw *FileAuditWriter) rotateLog() error {
	// Close current file
	if faw.file != nil {
		faw.file.Close()
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	oldPath := faw.config.OutputDestination
	newPath := fmt.Sprintf("%s.%s", oldPath, timestamp)
	
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(faw.config.OutputDestination, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	faw.file = file
	faw.encoder = json.NewEncoder(file)
	faw.size = 0

	// Clean up old log files if needed
	return faw.cleanupOldLogs()
}

// cleanupOldLogs removes old log files beyond the retention limit
func (faw *FileAuditWriter) cleanupOldLogs() error {
	dir := filepath.Dir(faw.config.OutputDestination)
	base := filepath.Base(faw.config.OutputDestination)
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var logFiles []os.DirEntry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), base+".") && !entry.IsDir() {
			logFiles = append(logFiles, entry)
		}
	}

	// If we have more than the maximum allowed, remove the oldest
	if len(logFiles) > faw.config.MaxLogFiles {
		// Sort by modification time would be better, but for simplicity, assume naming convention
		for i := 0; i < len(logFiles)-faw.config.MaxLogFiles; i++ {
			filePath := filepath.Join(dir, logFiles[i].Name())
			os.Remove(filePath)
		}
	}

	return nil
}

// Flush flushes buffered events
func (faw *FileAuditWriter) Flush() error {
	faw.mutex.Lock()
	defer faw.mutex.Unlock()
	return faw.flushBuffer()
}

// Close closes the file writer
func (faw *FileAuditWriter) Close() error {
	faw.Flush()
	if faw.file != nil {
		return faw.file.Close()
	}
	return nil
}

// HashProcessor adds integrity hashes to audit events
type HashProcessor struct{}

func (hp *HashProcessor) Name() string {
	return "HashProcessor"
}

func (hp *HashProcessor) Process(event AuditEvent) (AuditEvent, error) {
	// Create a hash of the event for integrity checking
	eventData, err := json.Marshal(event)
	if err != nil {
		return event, err
	}

	hash := fmt.Sprintf("%x", hashData(eventData))
	event.Hash = hash

	return event, nil
}

// ComplianceProcessor adds compliance tags to events
type ComplianceProcessor struct{}

func (cp *ComplianceProcessor) Name() string {
	return "ComplianceProcessor"
}

func (cp *ComplianceProcessor) Process(event AuditEvent) (AuditEvent, error) {
	// Add compliance-specific tags based on event category
	switch event.Category {
	case AuditCategoryAuthentication, AuditCategoryAuthorization:
		event.Tags = append(event.Tags, "access_control", "identity_management")
		
	case AuditCategoryDataAccess, AuditCategoryDataModification:
		event.Tags = append(event.Tags, "data_governance", "privacy")
		
	case AuditCategorySecurity:
		event.Tags = append(event.Tags, "security_incident", "threat_detection")
		
	case AuditCategoryCompliance:
		event.Tags = append(event.Tags, "regulatory_compliance", "policy_enforcement")
	}

	return event, nil
}

// NewGovernanceEngine creates a new governance engine
func NewGovernanceEngine(auditLogger *AuditLogger) *GovernanceEngine {
	return &GovernanceEngine{
		policies:   make(map[string]*GovernancePolicy),
		auditLog:   auditLogger,
		validators: make(map[string]func(interface{}) error),
	}
}

// AddPolicy adds a governance policy
func (ge *GovernanceEngine) AddPolicy(policy *GovernancePolicy) error {
	ge.mutex.Lock()
	defer ge.mutex.Unlock()

	if policy.ID == "" {
		policy.ID = generatePolicyID()
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	ge.policies[policy.ID] = policy

	// Log policy addition
	if ge.auditLog != nil {
		ge.auditLog.LogComplianceEvent(
			ComplianceFramework(policy.Category),
			"policy_added",
			AuditResultSuccess,
			map[string]interface{}{
				"policy_id":   policy.ID,
				"policy_name": policy.Name,
				"compliance":  policy.Compliance,
			},
		)
	}

	return nil
}

// ValidatePolicy validates a governance policy
func (ge *GovernanceEngine) ValidatePolicy(policyID string, data interface{}) error {
	ge.mutex.RLock()
	policy, exists := ge.policies[policyID]
	ge.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	if !policy.Active {
		return fmt.Errorf("policy %s is not active", policyID)
	}

	// Validate against each rule in the policy
	for _, rule := range policy.Rules {
		if !rule.Active {
			continue
		}

		if err := ge.validateRule(rule, data); err != nil {
			// Log policy violation
			if ge.auditLog != nil {
				ge.auditLog.LogComplianceEvent(
					ComplianceFramework(policy.Category),
					"policy_violation",
					AuditResultFailure,
					map[string]interface{}{
						"policy_id": policyID,
						"rule_id":   rule.ID,
						"error":     err.Error(),
					},
				)
			}
			return fmt.Errorf("policy validation failed: %w", err)
		}
	}

	return nil
}

// validateRule validates data against a single rule
func (ge *GovernanceEngine) validateRule(rule GovernanceRule, data interface{}) error {
	// This is a simplified implementation
	// In a real system, this would parse and execute the rule condition
	
	if validator, exists := ge.validators[rule.ID]; exists {
		return validator(data)
	}

	// Default validation - check if required fields exist
	if rule.Condition == "required_field" {
		if fieldName, ok := rule.Parameters["field"].(string); ok {
			// Simple validation - check if field exists in map
			if dataMap, ok := data.(map[string]interface{}); ok {
				if _, exists := dataMap[fieldName]; !exists {
					return fmt.Errorf("required field %s is missing", fieldName)
				}
			}
		}
	}

	return nil
}

// Utility functions

func createAuditWriter(config AuditConfig) (AuditWriter, error) {
	switch config.OutputFormat {
	case AuditFormatJSON:
		file, err := os.OpenFile(config.OutputDestination, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		return &FileAuditWriter{
			config:  config,
			file:    file,
			encoder: json.NewEncoder(file),
			buffer:  make([]AuditEvent, 0, config.BufferSize),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported audit format: %s", config.OutputFormat)
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("event-%d", time.Now().UnixNano())
}

func generatePolicyID() string {
	return fmt.Sprintf("policy-%d", time.Now().UnixNano())
}

func hashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}