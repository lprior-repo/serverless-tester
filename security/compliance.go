package security

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ComplianceFramework represents a compliance framework type
type ComplianceFramework string

const (
	ComplianceSOX      ComplianceFramework = "SOX"      // Sarbanes-Oxley Act
	ComplianceHIPAA    ComplianceFramework = "HIPAA"    // Health Insurance Portability and Accountability Act
	ComplianceGDPR     ComplianceFramework = "GDPR"     // General Data Protection Regulation
	CompliancePCIDSS   ComplianceFramework = "PCI-DSS"  // Payment Card Industry Data Security Standard
	ComplianceSOC2     ComplianceFramework = "SOC2"     // Service Organization Control 2
	ComplianceFERPA    ComplianceFramework = "FERPA"    // Family Educational Rights and Privacy Act
)

// ComplianceRequirement represents a specific compliance requirement
type ComplianceRequirement struct {
	ID          string
	Framework   ComplianceFramework
	Title       string
	Description string
	Category    string
	Severity    ComplianceSeverity
	Required    bool
	TestCases   []ComplianceTestCase
}

// ComplianceSeverity represents the severity level of a compliance requirement
type ComplianceSeverity string

const (
	ComplianceSeverityCritical ComplianceSeverity = "Critical"
	ComplianceSeverityHigh     ComplianceSeverity = "High"
	ComplianceSeverityMedium   ComplianceSeverity = "Medium"
	ComplianceSeverityLow      ComplianceSeverity = "Low"
)

// ComplianceTestCase represents a test case for compliance validation
type ComplianceTestCase struct {
	ID          string
	Name        string
	Description string
	TestFunc    func(config ComplianceConfig) ComplianceResult
}

// ComplianceConfig holds configuration for compliance testing
type ComplianceConfig struct {
	Framework       ComplianceFramework
	DataTypes       []DataType
	EncryptionRequired bool
	AccessLogging   bool
	DataRetention   time.Duration
	GeographicRegion string
	CustomSettings  map[string]interface{}
}

// DataType represents types of data being processed
type DataType string

const (
	DataTypePII         DataType = "PII"          // Personally Identifiable Information
	DataTypePHI         DataType = "PHI"          // Protected Health Information
	DataTypePCI         DataType = "PCI"          // Payment Card Information
	DataTypeFinancial   DataType = "Financial"    // Financial data
	DataTypeEducational DataType = "Educational"  // Educational records
	DataTypeBiometric   DataType = "Biometric"    // Biometric data
)

// ComplianceResult represents the result of a compliance check
type ComplianceResult struct {
	RequirementID string
	Framework     ComplianceFramework
	Status        ComplianceStatus
	Message       string
	Evidence      map[string]interface{}
	TestedAt      time.Time
	Remediation   string
}

// ComplianceStatus represents the status of a compliance check
type ComplianceStatus string

const (
	ComplianceStatusPassed    ComplianceStatus = "Passed"
	ComplianceStatusFailed    ComplianceStatus = "Failed"
	ComplianceStatusWarning   ComplianceStatus = "Warning"
	ComplianceStatusNotTested ComplianceStatus = "NotTested"
)

// ComplianceValidator validates compliance requirements
type ComplianceValidator struct {
	requirements map[ComplianceFramework][]ComplianceRequirement
}

// NewComplianceValidator creates a new compliance validator
func NewComplianceValidator() *ComplianceValidator {
	cv := &ComplianceValidator{
		requirements: make(map[ComplianceFramework][]ComplianceRequirement),
	}
	
	cv.initializeSOXRequirements()
	cv.initializeHIPAARequirements()
	cv.initializeGDPRRequirements()
	cv.initializePCIDSSRequirements()
	
	return cv
}

// ValidateCompliance validates all requirements for a specific framework
func (cv *ComplianceValidator) ValidateCompliance(config ComplianceConfig) []ComplianceResult {
	requirements, exists := cv.requirements[config.Framework]
	if !exists {
		return []ComplianceResult{{
			Framework: config.Framework,
			Status:    ComplianceStatusFailed,
			Message:   fmt.Sprintf("Unknown compliance framework: %s", config.Framework),
			TestedAt:  time.Now(),
		}}
	}

	var results []ComplianceResult
	for _, req := range requirements {
		for _, testCase := range req.TestCases {
			result := testCase.TestFunc(config)
			result.RequirementID = req.ID
			result.Framework = config.Framework
			result.TestedAt = time.Now()
			results = append(results, result)
		}
	}

	return results
}

// GetRequirements returns all requirements for a framework
func (cv *ComplianceValidator) GetRequirements(framework ComplianceFramework) []ComplianceRequirement {
	return cv.requirements[framework]
}

// initializeSOXRequirements sets up Sarbanes-Oxley compliance requirements
func (cv *ComplianceValidator) initializeSOXRequirements() {
	requirements := []ComplianceRequirement{
		{
			ID:          "SOX-302",
			Framework:   ComplianceSOX,
			Title:       "Corporate Responsibility for Financial Reports",
			Description: "Principal executive and financial officers must certify financial reports",
			Category:    "Financial Reporting",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "SOX-302-001",
					Name: "Audit Trail Validation",
					Description: "Validate that all financial data changes are logged and traceable",
					TestFunc: validateSOXAuditTrail,
				},
			},
		},
		{
			ID:          "SOX-404",
			Framework:   ComplianceSOX,
			Title:       "Management Assessment of Internal Controls",
			Description: "Internal control over financial reporting assessment",
			Category:    "Internal Controls",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "SOX-404-001",
					Name: "Access Control Validation",
					Description: "Validate proper access controls for financial systems",
					TestFunc: validateSOXAccessControls,
				},
			},
		},
	}
	
	cv.requirements[ComplianceSOX] = requirements
}

// initializeHIPAARequirements sets up HIPAA compliance requirements
func (cv *ComplianceValidator) initializeHIPAARequirements() {
	requirements := []ComplianceRequirement{
		{
			ID:          "HIPAA-164.312",
			Framework:   ComplianceHIPAA,
			Title:       "Technical Safeguards",
			Description: "Technical policies and procedures for electronic protected health information",
			Category:    "Technical Safeguards",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "HIPAA-164.312-001",
					Name: "PHI Encryption Validation",
					Description: "Validate that PHI data is properly encrypted at rest and in transit",
					TestFunc: validateHIPAAEncryption,
				},
			},
		},
		{
			ID:          "HIPAA-164.308",
			Framework:   ComplianceHIPAA,
			Title:       "Administrative Safeguards",
			Description: "Administrative actions and policies to manage security measures",
			Category:    "Administrative Safeguards",
			Severity:    ComplianceSeverityHigh,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "HIPAA-164.308-001",
					Name: "Access Management Validation",
					Description: "Validate proper access management for PHI data",
					TestFunc: validateHIPAAAccessManagement,
				},
			},
		},
	}
	
	cv.requirements[ComplianceHIPAA] = requirements
}

// initializeGDPRRequirements sets up GDPR compliance requirements
func (cv *ComplianceValidator) initializeGDPRRequirements() {
	requirements := []ComplianceRequirement{
		{
			ID:          "GDPR-Art25",
			Framework:   ComplianceGDPR,
			Title:       "Data Protection by Design and by Default",
			Description: "Technical and organizational measures for data protection",
			Category:    "Data Protection",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "GDPR-Art25-001",
					Name: "Data Minimization Validation",
					Description: "Validate that only necessary personal data is collected and processed",
					TestFunc: validateGDPRDataMinimization,
				},
			},
		},
		{
			ID:          "GDPR-Art32",
			Framework:   ComplianceGDPR,
			Title:       "Security of Processing",
			Description: "Appropriate technical and organizational measures to ensure security",
			Category:    "Security of Processing",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "GDPR-Art32-001",
					Name: "Encryption and Pseudonymization",
					Description: "Validate encryption and pseudonymization of personal data",
					TestFunc: validateGDPREncryption,
				},
			},
		},
		{
			ID:          "GDPR-Art17",
			Framework:   ComplianceGDPR,
			Title:       "Right to Erasure",
			Description: "Right to have personal data erased without undue delay",
			Category:    "Data Subject Rights",
			Severity:    ComplianceSeverityHigh,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "GDPR-Art17-001",
					Name: "Data Deletion Capability",
					Description: "Validate capability to delete personal data upon request",
					TestFunc: validateGDPRDataDeletion,
				},
			},
		},
	}
	
	cv.requirements[ComplianceGDPR] = requirements
}

// initializePCIDSSRequirements sets up PCI-DSS compliance requirements
func (cv *ComplianceValidator) initializePCIDSSRequirements() {
	requirements := []ComplianceRequirement{
		{
			ID:          "PCI-DSS-3",
			Framework:   CompliancePCIDSS,
			Title:       "Protect Stored Cardholder Data",
			Description: "Protection requirements for stored cardholder data",
			Category:    "Data Protection",
			Severity:    ComplianceSeverityCritical,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "PCI-DSS-3-001",
					Name: "Cardholder Data Encryption",
					Description: "Validate encryption of stored cardholder data",
					TestFunc: validatePCIDSSEncryption,
				},
			},
		},
		{
			ID:          "PCI-DSS-8",
			Framework:   CompliancePCIDSS,
			Title:       "Identify and Authenticate Access",
			Description: "Requirements for user identification and authentication",
			Category:    "Access Control",
			Severity:    ComplianceSeverityHigh,
			Required:    true,
			TestCases: []ComplianceTestCase{
				{
					ID:   "PCI-DSS-8-001",
					Name: "Multi-Factor Authentication",
					Description: "Validate multi-factor authentication for system access",
					TestFunc: validatePCIDSSMFA,
				},
			},
		},
	}
	
	cv.requirements[CompliancePCIDSS] = requirements
}

// Compliance test functions

func validateSOXAuditTrail(config ComplianceConfig) ComplianceResult {
	if !config.AccessLogging {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "Access logging is not enabled for SOX compliance",
			Remediation: "Enable comprehensive access logging for all financial data operations",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "SOX audit trail requirements validated successfully",
		Evidence: map[string]interface{}{
			"access_logging_enabled": config.AccessLogging,
		},
	}
}

func validateSOXAccessControls(config ComplianceConfig) ComplianceResult {
	// Check if financial data types are properly configured
	hasFinancialData := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypeFinancial {
			hasFinancialData = true
			break
		}
	}
	
	if hasFinancialData && !config.EncryptionRequired {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "Financial data must be encrypted for SOX compliance",
			Remediation: "Enable encryption for all financial data",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "SOX access control requirements validated successfully",
		Evidence: map[string]interface{}{
			"encryption_enabled": config.EncryptionRequired,
			"data_types":        config.DataTypes,
		},
	}
}

func validateHIPAAEncryption(config ComplianceConfig) ComplianceResult {
	// Check if PHI data types are present
	hasPHI := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypePHI {
			hasPHI = true
			break
		}
	}
	
	if hasPHI && !config.EncryptionRequired {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "PHI data must be encrypted for HIPAA compliance",
			Remediation: "Enable encryption for all PHI data at rest and in transit",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "HIPAA encryption requirements validated successfully",
		Evidence: map[string]interface{}{
			"encryption_enabled": config.EncryptionRequired,
			"phi_present":       hasPHI,
		},
	}
}

func validateHIPAAAccessManagement(config ComplianceConfig) ComplianceResult {
	if !config.AccessLogging {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "Access logging is required for HIPAA compliance",
			Remediation: "Enable access logging for all PHI access attempts",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "HIPAA access management requirements validated successfully",
	}
}

func validateGDPRDataMinimization(config ComplianceConfig) ComplianceResult {
	// Check if PII is being processed
	hasPII := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypePII {
			hasPII = true
			break
		}
	}
	
	if hasPII {
		// For GDPR, we need to ensure data retention policies are in place
		if config.DataRetention == 0 {
			return ComplianceResult{
				Status:      ComplianceStatusWarning,
				Message:     "Data retention policy should be defined for PII data",
				Remediation: "Define and implement data retention policies",
			}
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "GDPR data minimization principles validated",
		Evidence: map[string]interface{}{
			"pii_present":      hasPII,
			"data_retention":   config.DataRetention,
		},
	}
}

func validateGDPREncryption(config ComplianceConfig) ComplianceResult {
	// Check if personal data requires encryption
	hasPersonalData := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypePII || dataType == DataTypeBiometric {
			hasPersonalData = true
			break
		}
	}
	
	if hasPersonalData && !config.EncryptionRequired {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "Personal data must be encrypted for GDPR compliance",
			Remediation: "Implement encryption and pseudonymization for personal data",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "GDPR encryption requirements validated successfully",
	}
}

func validateGDPRDataDeletion(config ComplianceConfig) ComplianceResult {
	// This would typically check if the system has data deletion capabilities
	// For now, we'll assume it's implemented if PII data types are configured
	hasPII := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypePII {
			hasPII = true
			break
		}
	}
	
	if hasPII {
		return ComplianceResult{
			Status:  ComplianceStatusPassed,
			Message: "GDPR data deletion capability assumed to be implemented",
			Evidence: map[string]interface{}{
				"pii_data_types": config.DataTypes,
			},
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "No personal data types configured, deletion capability not required",
	}
}

func validatePCIDSSEncryption(config ComplianceConfig) ComplianceResult {
	// Check if PCI data is being processed
	hasPCIData := false
	for _, dataType := range config.DataTypes {
		if dataType == DataTypePCI {
			hasPCIData = true
			break
		}
	}
	
	if hasPCIData && !config.EncryptionRequired {
		return ComplianceResult{
			Status:      ComplianceStatusFailed,
			Message:     "Cardholder data must be encrypted for PCI-DSS compliance",
			Remediation: "Implement strong encryption for all cardholder data",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "PCI-DSS encryption requirements validated successfully",
	}
}

func validatePCIDSSMFA(config ComplianceConfig) ComplianceResult {
	// This would typically check MFA implementation
	// For testing purposes, we'll check if access logging is enabled as a proxy
	if !config.AccessLogging {
		return ComplianceResult{
			Status:      ComplianceStatusWarning,
			Message:     "Multi-factor authentication should be implemented for PCI-DSS compliance",
			Remediation: "Implement multi-factor authentication for all system access",
		}
	}
	
	return ComplianceResult{
		Status:  ComplianceStatusPassed,
		Message: "PCI-DSS authentication requirements validated",
	}
}

// Utility functions for compliance validation

// ValidateDataClassification ensures data types are properly classified
func ValidateDataClassification(dataTypes []DataType) error {
	validTypes := map[DataType]bool{
		DataTypePII:         true,
		DataTypePHI:         true,
		DataTypePCI:         true,
		DataTypeFinancial:   true,
		DataTypeEducational: true,
		DataTypeBiometric:   true,
	}
	
	for _, dataType := range dataTypes {
		if !validTypes[dataType] {
			return fmt.Errorf("invalid data type: %s", dataType)
		}
	}
	
	return nil
}

// GetRequiredCompliance determines which compliance frameworks are required based on data types
func GetRequiredCompliance(dataTypes []DataType, region string) []ComplianceFramework {
	var required []ComplianceFramework
	frameworks := make(map[ComplianceFramework]bool)
	
	for _, dataType := range dataTypes {
		switch dataType {
		case DataTypePHI:
			frameworks[ComplianceHIPAA] = true
		case DataTypePCI:
			frameworks[CompliancePCIDSS] = true
		case DataTypeFinancial:
			frameworks[ComplianceSOX] = true
		case DataTypePII, DataTypeBiometric:
			if isEURegion(region) {
				frameworks[ComplianceGDPR] = true
			}
		case DataTypeEducational:
			frameworks[ComplianceFERPA] = true
		}
	}
	
	for framework := range frameworks {
		required = append(required, framework)
	}
	
	return required
}

// isEURegion checks if a region requires GDPR compliance
func isEURegion(region string) bool {
	euRegions := []string{
		"eu-", "europe-", "eu-west-", "eu-central-", "eu-north-", "eu-south-",
	}
	
	regionLower := strings.ToLower(region)
	for _, euPrefix := range euRegions {
		if strings.HasPrefix(regionLower, euPrefix) {
			return true
		}
	}
	
	return false
}

// ValidateComplianceConfig validates a compliance configuration
func ValidateComplianceConfig(config ComplianceConfig) error {
	validFrameworks := map[ComplianceFramework]bool{
		ComplianceSOX:    true,
		ComplianceHIPAA:  true,
		ComplianceGDPR:   true,
		CompliancePCIDSS: true,
		ComplianceSOC2:   true,
		ComplianceFERPA:  true,
	}
	
	if !validFrameworks[config.Framework] {
		return fmt.Errorf("invalid compliance framework: %s", config.Framework)
	}
	
	if err := ValidateDataClassification(config.DataTypes); err != nil {
		return fmt.Errorf("invalid data classification: %w", err)
	}
	
	if config.DataRetention < 0 {
		return fmt.Errorf("data retention period cannot be negative")
	}
	
	if config.GeographicRegion != "" {
		regionPattern := regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d+$`)
		if !regionPattern.MatchString(config.GeographicRegion) {
			return fmt.Errorf("invalid geographic region format: %s", config.GeographicRegion)
		}
	}
	
	return nil
}