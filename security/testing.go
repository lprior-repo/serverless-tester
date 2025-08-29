package security

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// VulnerabilityScanner provides security testing capabilities
type VulnerabilityScanner struct {
	config SecurityTestConfig
}

// SecurityTestConfig holds configuration for security testing
type SecurityTestConfig struct {
	EnableVulnerabilityScanning bool
	EnableCodeAnalysis         bool
	EnableDependencyScanning   bool
	EnableSecretsDetection     bool
	ScanTimeout               time.Duration
	ReportFormat              string
	OutputDirectory           string
}

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

const (
	SeverityCritical VulnerabilitySeverity = "Critical"
	SeverityHigh     VulnerabilitySeverity = "High"
	SeverityMedium   VulnerabilitySeverity = "Medium"
	SeverityLow      VulnerabilitySeverity = "Low"
	SeverityInfo     VulnerabilitySeverity = "Info"
)

// VulnerabilityType represents the type of vulnerability
type VulnerabilityType string

const (
	VulnTypeInjection          VulnerabilityType = "Injection"
	VulnTypeBrokenAuth         VulnerabilityType = "BrokenAuthentication"
	VulnTypeSensitiveData      VulnerabilityType = "SensitiveDataExposure"
	VulnTypeXXE               VulnerabilityType = "XXE"
	VulnTypeBrokenAccessControl VulnerabilityType = "BrokenAccessControl"
	VulnTypeSecurityMisconfig  VulnerabilityType = "SecurityMisconfiguration"
	VulnTypeXSS               VulnerabilityType = "XSS"
	VulnTypeInsecureDeserialization VulnerabilityType = "InsecureDeserialization"
	VulnTypeKnownVulns        VulnerabilityType = "KnownVulnerabilities"
	VulnTypeInsufficientLogging VulnerabilityType = "InsufficientLogging"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string
	Type        VulnerabilityType
	Severity    VulnerabilitySeverity
	Title       string
	Description string
	Location    string
	LineNumber  int
	Evidence    string
	CVE         string
	CVSS        float64
	Solution    string
	References  []string
	DetectedAt  time.Time
}

// SecurityTestResult holds the results of security testing
type SecurityTestResult struct {
	TestID        string
	TestType      string
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Status        string
	Vulnerabilities []Vulnerability
	Summary       SecurityTestSummary
	Recommendations []string
}

// SecurityTestSummary provides a summary of security test results
type SecurityTestSummary struct {
	TotalVulnerabilities int
	CriticalCount       int
	HighCount          int
	MediumCount        int
	LowCount           int
	InfoCount          int
	RiskScore          float64
}

// NewVulnerabilityScanner creates a new vulnerability scanner
func NewVulnerabilityScanner(config SecurityTestConfig) *VulnerabilityScanner {
	if config.ScanTimeout == 0 {
		config.ScanTimeout = 10 * time.Minute
	}
	
	if config.ReportFormat == "" {
		config.ReportFormat = "json"
	}
	
	if config.OutputDirectory == "" {
		config.OutputDirectory = "./security-reports"
	}

	return &VulnerabilityScanner{
		config: config,
	}
}

// ScanCode performs static code analysis for security vulnerabilities
func (vs *VulnerabilityScanner) ScanCode(ctx context.Context, codebasePath string) (*SecurityTestResult, error) {
	result := &SecurityTestResult{
		TestID:    generateTestID("code-scan"),
		TestType:  "Static Code Analysis",
		StartTime: time.Now(),
		Status:    "Running",
	}

	var vulnerabilities []Vulnerability

	// Walk through the codebase
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-code files
		if info.IsDir() || !isCodeFile(path) {
			return nil
		}

		// Scan file for vulnerabilities
		fileVulns, err := vs.scanFile(path)
		if err != nil {
			return fmt.Errorf("failed to scan file %s: %w", path, err)
		}

		vulnerabilities = append(vulnerabilities, fileVulns...)
		return nil
	})

	if err != nil {
		result.Status = "Failed"
		return result, fmt.Errorf("code scan failed: %w", err)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Vulnerabilities = vulnerabilities
	result.Summary = vs.generateSummary(vulnerabilities)
	result.Recommendations = vs.generateRecommendations(vulnerabilities)
	result.Status = "Completed"

	return result, nil
}

// ScanDependencies checks for known vulnerabilities in dependencies
func (vs *VulnerabilityScanner) ScanDependencies(ctx context.Context, projectPath string) (*SecurityTestResult, error) {
	result := &SecurityTestResult{
		TestID:    generateTestID("dependency-scan"),
		TestType:  "Dependency Vulnerability Scan",
		StartTime: time.Now(),
		Status:    "Running",
	}

	var vulnerabilities []Vulnerability

	// Check for Go dependencies
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		goVulns, err := vs.scanGoModules(goModPath)
		if err != nil {
			result.Status = "Failed"
			return result, fmt.Errorf("failed to scan Go modules: %w", err)
		}
		vulnerabilities = append(vulnerabilities, goVulns...)
	}

	// Check for package.json (Node.js)
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		nodeVulns, err := vs.scanNodeModules(packageJSONPath)
		if err != nil {
			result.Status = "Failed"
			return result, fmt.Errorf("failed to scan Node modules: %w", err)
		}
		vulnerabilities = append(vulnerabilities, nodeVulns...)
	}

	// Check for requirements.txt (Python)
	requirementsPath := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		pythonVulns, err := vs.scanPythonRequirements(requirementsPath)
		if err != nil {
			result.Status = "Failed"
			return result, fmt.Errorf("failed to scan Python requirements: %w", err)
		}
		vulnerabilities = append(vulnerabilities, pythonVulns...)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Vulnerabilities = vulnerabilities
	result.Summary = vs.generateSummary(vulnerabilities)
	result.Recommendations = vs.generateRecommendations(vulnerabilities)
	result.Status = "Completed"

	return result, nil
}

// DetectSecrets scans for hardcoded secrets and credentials
func (vs *VulnerabilityScanner) DetectSecrets(ctx context.Context, codebasePath string) (*SecurityTestResult, error) {
	result := &SecurityTestResult{
		TestID:    generateTestID("secrets-scan"),
		TestType:  "Secrets Detection",
		StartTime: time.Now(),
		Status:    "Running",
	}

	var vulnerabilities []Vulnerability

	// Walk through the codebase
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and binary files
		if info.IsDir() || isBinaryFile(path) {
			return nil
		}

		// Scan file for secrets
		secrets, err := vs.scanFileForSecrets(path)
		if err != nil {
			return fmt.Errorf("failed to scan file for secrets %s: %w", path, err)
		}

		vulnerabilities = append(vulnerabilities, secrets...)
		return nil
	})

	if err != nil {
		result.Status = "Failed"
		return result, fmt.Errorf("secrets scan failed: %w", err)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Vulnerabilities = vulnerabilities
	result.Summary = vs.generateSummary(vulnerabilities)
	result.Recommendations = vs.generateRecommendations(vulnerabilities)
	result.Status = "Completed"

	return result, nil
}

// ScanWebApplication performs web application security testing
func (vs *VulnerabilityScanner) ScanWebApplication(ctx context.Context, baseURL string) (*SecurityTestResult, error) {
	result := &SecurityTestResult{
		TestID:    generateTestID("web-scan"),
		TestType:  "Web Application Security Scan",
		StartTime: time.Now(),
		Status:    "Running",
	}

	var vulnerabilities []Vulnerability

	// Test for common web vulnerabilities
	webVulns := []func(string) []Vulnerability{
		vs.testForXSS,
		vs.testForSQLInjection,
		vs.testForDirectoryTraversal,
		vs.testForCommandInjection,
		vs.testForInsecureHeaders,
	}

	for _, testFunc := range webVulns {
		vulns := testFunc(baseURL)
		vulnerabilities = append(vulnerabilities, vulns...)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Vulnerabilities = vulnerabilities
	result.Summary = vs.generateSummary(vulnerabilities)
	result.Recommendations = vs.generateRecommendations(vulnerabilities)
	result.Status = "Completed"

	return result, nil
}

// scanFile scans a single file for security vulnerabilities
func (vs *VulnerabilityScanner) scanFile(filepath string) ([]Vulnerability, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var vulnerabilities []Vulnerability
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		lineVulns := vs.scanLineForVulnerabilities(filepath, i+1, line)
		vulnerabilities = append(vulnerabilities, lineVulns...)
	}

	return vulnerabilities, nil
}

// scanLineForVulnerabilities scans a single line for vulnerabilities
func (vs *VulnerabilityScanner) scanLineForVulnerabilities(filepath string, lineNumber int, line string) []Vulnerability {
	var vulnerabilities []Vulnerability

	// Check for SQL injection patterns
	sqlPatterns := []string{
		`(?i)select\s+.*from\s+.*where.*=.*\$\{`,
		`(?i)insert\s+into.*values.*\$\{`,
		`(?i)update\s+.*set.*=.*\$\{`,
		`(?i)delete\s+from.*where.*=.*\$\{`,
	}

	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeInjection,
				Severity:    SeverityHigh,
				Title:       "Potential SQL Injection",
				Description: "Line contains patterns that may be vulnerable to SQL injection",
				Location:    filepath,
				LineNumber:  lineNumber,
				Evidence:    strings.TrimSpace(line),
				Solution:    "Use parameterized queries or prepared statements",
				DetectedAt:  time.Now(),
			})
		}
	}

	// Check for XSS patterns
	xssPatterns := []string{
		`(?i)innerHTML\s*=.*\$\{`,
		`(?i)document\.write\s*\(.*\$\{`,
		`(?i)eval\s*\(.*\$\{`,
	}

	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeXSS,
				Severity:    SeverityMedium,
				Title:       "Potential XSS Vulnerability",
				Description: "Line contains patterns that may be vulnerable to XSS",
				Location:    filepath,
				LineNumber:  lineNumber,
				Evidence:    strings.TrimSpace(line),
				Solution:    "Sanitize user input and use proper output encoding",
				DetectedAt:  time.Now(),
			})
		}
	}

	// Check for hardcoded credentials patterns
	credentialPatterns := []string{
		`(?i)password\s*[:=]\s*["'][^"']{8,}["']`,
		`(?i)api_key\s*[:=]\s*["'][^"']{20,}["']`,
		`(?i)secret\s*[:=]\s*["'][^"']{16,}["']`,
		`(?i)token\s*[:=]\s*["'][^"']{20,}["']`,
	}

	for _, pattern := range credentialPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeSensitiveData,
				Severity:    SeverityCritical,
				Title:       "Hardcoded Credentials",
				Description: "Line appears to contain hardcoded credentials",
				Location:    filepath,
				LineNumber:  lineNumber,
				Evidence:    "[REDACTED]", // Don't include actual credentials
				Solution:    "Use environment variables or secure credential management",
				DetectedAt:  time.Now(),
			})
		}
	}

	// Check for weak cryptographic algorithms
	weakCryptoPatterns := []string{
		`(?i)md5\.new\(`,
		`(?i)sha1\.new\(`,
		`(?i)des\.new\(`,
		`(?i)rc4\.new\(`,
	}

	for _, pattern := range weakCryptoPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeSecurityMisconfig,
				Severity:    SeverityMedium,
				Title:       "Weak Cryptographic Algorithm",
				Description: "Line uses weak or deprecated cryptographic algorithm",
				Location:    filepath,
				LineNumber:  lineNumber,
				Evidence:    strings.TrimSpace(line),
				Solution:    "Use strong cryptographic algorithms like SHA-256, AES-256",
				DetectedAt:  time.Now(),
			})
		}
	}

	return vulnerabilities
}

// scanGoModules scans Go modules for known vulnerabilities
func (vs *VulnerabilityScanner) scanGoModules(goModPath string) ([]Vulnerability, error) {
	// This would typically integrate with a vulnerability database
	// For now, we'll implement basic checks
	var vulnerabilities []Vulnerability

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		// Check for known vulnerable versions (simplified example)
		if strings.Contains(line, "github.com/gin-gonic/gin") && strings.Contains(line, "v1.6.") {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeKnownVulns,
				Severity:    SeverityMedium,
				Title:       "Known Vulnerability in Gin Framework",
				Description: "This version of Gin has known security issues",
				Location:    goModPath,
				LineNumber:  i + 1,
				Evidence:    strings.TrimSpace(line),
				Solution:    "Update to the latest version of Gin",
				References:  []string{"https://github.com/gin-gonic/gin/security/advisories"},
				DetectedAt:  time.Now(),
			})
		}
	}

	return vulnerabilities, nil
}

// scanNodeModules scans Node.js modules for vulnerabilities
func (vs *VulnerabilityScanner) scanNodeModules(packageJSONPath string) ([]Vulnerability, error) {
	// Placeholder for Node.js vulnerability scanning
	// This would integrate with npm audit or similar tools
	return []Vulnerability{}, nil
}

// scanPythonRequirements scans Python requirements for vulnerabilities
func (vs *VulnerabilityScanner) scanPythonRequirements(requirementsPath string) ([]Vulnerability, error) {
	// Placeholder for Python vulnerability scanning
	// This would integrate with safety or similar tools
	return []Vulnerability{}, nil
}

// scanFileForSecrets scans a file for hardcoded secrets
func (vs *VulnerabilityScanner) scanFileForSecrets(filepath string) ([]Vulnerability, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var vulnerabilities []Vulnerability
	lines := strings.Split(string(content), "\n")

	secretPatterns := map[string]string{
		"AWS Access Key":    `AKIA[0-9A-Z]{16}`,
		"AWS Secret Key":    `[0-9a-zA-Z/+=]{40}`,
		"GitHub Token":      `ghp_[0-9a-zA-Z]{36}`,
		"Slack Token":       `xoxb-[0-9]+-[0-9]+-[0-9a-zA-Z]+`,
		"Google API Key":    `AIza[0-9A-Za-z\\-_]{35}`,
		"Private Key":       `-----BEGIN\s+.*PRIVATE\s+KEY-----`,
		"JWT Token":         `eyJ[0-9a-zA-Z_-]*\.eyJ[0-9a-zA-Z_-]*\.[0-9a-zA-Z_-]*`,
	}

	for i, line := range lines {
		for secretType, pattern := range secretPatterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				vulnerabilities = append(vulnerabilities, Vulnerability{
					ID:          generateVulnerabilityID(),
					Type:        VulnTypeSensitiveData,
					Severity:    SeverityCritical,
					Title:       fmt.Sprintf("Hardcoded %s", secretType),
					Description: fmt.Sprintf("File contains what appears to be a hardcoded %s", secretType),
					Location:    filepath,
					LineNumber:  i + 1,
					Evidence:    "[REDACTED]", // Don't include actual secrets
					Solution:    "Remove hardcoded secrets and use secure credential management",
					DetectedAt:  time.Now(),
				})
			}
		}
	}

	return vulnerabilities, nil
}

// Web application vulnerability testing functions
func (vs *VulnerabilityScanner) testForXSS(baseURL string) []Vulnerability {
	// This would perform actual XSS testing
	// For now, return empty slice
	return []Vulnerability{}
}

func (vs *VulnerabilityScanner) testForSQLInjection(baseURL string) []Vulnerability {
	// This would perform actual SQL injection testing
	// For now, return empty slice
	return []Vulnerability{}
}

func (vs *VulnerabilityScanner) testForDirectoryTraversal(baseURL string) []Vulnerability {
	// This would test for directory traversal vulnerabilities
	// For now, return empty slice
	return []Vulnerability{}
}

func (vs *VulnerabilityScanner) testForCommandInjection(baseURL string) []Vulnerability {
	// This would test for command injection vulnerabilities
	// For now, return empty slice
	return []Vulnerability{}
}

func (vs *VulnerabilityScanner) testForInsecureHeaders(baseURL string) []Vulnerability {
	var vulnerabilities []Vulnerability

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		return vulnerabilities
	}
	defer resp.Body.Close()

	// Check for missing security headers
	securityHeaders := map[string]string{
		"Content-Security-Policy":   "Content Security Policy header is missing",
		"X-Frame-Options":          "X-Frame-Options header is missing",
		"X-Content-Type-Options":   "X-Content-Type-Options header is missing",
		"Strict-Transport-Security": "Strict-Transport-Security header is missing",
	}

	for header, message := range securityHeaders {
		if resp.Header.Get(header) == "" {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnerabilityID(),
				Type:        VulnTypeSecurityMisconfig,
				Severity:    SeverityMedium,
				Title:       fmt.Sprintf("Missing %s Header", header),
				Description: message,
				Location:    baseURL,
				Evidence:    fmt.Sprintf("Response headers: %v", resp.Header),
				Solution:    fmt.Sprintf("Add %s header to HTTP responses", header),
				DetectedAt:  time.Now(),
			})
		}
	}

	return vulnerabilities
}

// generateSummary creates a summary of vulnerabilities
func (vs *VulnerabilityScanner) generateSummary(vulnerabilities []Vulnerability) SecurityTestSummary {
	summary := SecurityTestSummary{
		TotalVulnerabilities: len(vulnerabilities),
	}

	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case SeverityCritical:
			summary.CriticalCount++
		case SeverityHigh:
			summary.HighCount++
		case SeverityMedium:
			summary.MediumCount++
		case SeverityLow:
			summary.LowCount++
		case SeverityInfo:
			summary.InfoCount++
		}
	}

	// Calculate risk score (simplified calculation)
	summary.RiskScore = float64(summary.CriticalCount*10 + summary.HighCount*7 + summary.MediumCount*4 + summary.LowCount*1)

	return summary
}

// generateRecommendations generates security recommendations based on findings
func (vs *VulnerabilityScanner) generateRecommendations(vulnerabilities []Vulnerability) []string {
	recommendations := make(map[string]bool)

	for _, vuln := range vulnerabilities {
		switch vuln.Type {
		case VulnTypeInjection:
			recommendations["Implement input validation and parameterized queries"] = true
		case VulnTypeSensitiveData:
			recommendations["Use secure credential management systems"] = true
		case VulnTypeSecurityMisconfig:
			recommendations["Review and harden security configurations"] = true
		case VulnTypeXSS:
			recommendations["Implement proper output encoding and input sanitization"] = true
		case VulnTypeKnownVulns:
			recommendations["Update dependencies to latest secure versions"] = true
		}
	}

	var result []string
	for rec := range recommendations {
		result = append(result, rec)
	}

	return result
}

// Utility functions

func isCodeFile(filepath string) bool {
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])
	codeExtensions := map[string]bool{
		"go":   true,
		"js":   true,
		"ts":   true,
		"py":   true,
		"java": true,
		"cs":   true,
		"cpp":  true,
		"c":    true,
		"php":  true,
		"rb":   true,
		"rs":   true,
	}
	return codeExtensions[ext]
}

func isBinaryFile(filepath string) bool {
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])
	binaryExtensions := map[string]bool{
		"exe": true, "dll": true, "so": true, "dylib": true,
		"jpg": true, "jpeg": true, "png": true, "gif": true,
		"pdf": true, "zip": true, "tar": true, "gz": true,
	}
	return binaryExtensions[ext]
}

func generateTestID(prefix string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())))
	return fmt.Sprintf("%s-%x", prefix, hash[:8])
}

func generateVulnerabilityID() string {
	hash := md5.Sum([]byte(fmt.Sprintf("vuln-%d", time.Now().UnixNano())))
	return fmt.Sprintf("VULN-%s", strings.ToUpper(hex.EncodeToString(hash[:])[:8]))
}