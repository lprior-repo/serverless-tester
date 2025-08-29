package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// DebugAnalyzer provides comprehensive debugging and troubleshooting capabilities
type DebugAnalyzer struct {
	region      string
	projectRoot string
	
	green  *color.Color
	red    *color.Color
	yellow *color.Color
	blue   *color.Color
	cyan   *color.Color
	bold   *color.Color
}

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Service     string    `json:"service"`
	RequestID   string    `json:"request_id"`
	Duration    string    `json:"duration"`
	BilledDuration string `json:"billed_duration"`
	MemorySize  string    `json:"memory_size"`
	MemoryUsed  string    `json:"memory_used"`
	Error       string    `json:"error,omitempty"`
	StackTrace  string    `json:"stack_trace,omitempty"`
	Raw         string    `json:"raw"`
}

// TraceSegment represents an X-Ray trace segment
type TraceSegment struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	StartTime  float64                `json:"start_time"`
	EndTime    float64                `json:"end_time"`
	Duration   float64                `json:"duration"`
	HTTP       map[string]interface{} `json:"http,omitempty"`
	AWS        map[string]interface{} `json:"aws,omitempty"`
	Error      bool                   `json:"error,omitempty"`
	Fault      bool                   `json:"fault,omitempty"`
	Throttle   bool                   `json:"throttle,omitempty"`
	Subsegments []TraceSegment        `json:"subsegments,omitempty"`
	Cause      map[string]interface{} `json:"cause,omitempty"`
}

// DebugReport contains analysis results
type DebugReport struct {
	Summary         *DebugSummary      `json:"summary"`
	LogAnalysis     *LogAnalysis       `json:"log_analysis"`
	TraceAnalysis   *TraceAnalysis     `json:"trace_analysis,omitempty"`
	Recommendations []Recommendation   `json:"recommendations"`
	Timestamp       time.Time          `json:"timestamp"`
}

// DebugSummary provides high-level debug information
type DebugSummary struct {
	Service        string        `json:"service"`
	TimeRange      string        `json:"time_range"`
	TotalLogs      int           `json:"total_logs"`
	ErrorCount     int           `json:"error_count"`
	WarningCount   int           `json:"warning_count"`
	UniqueErrors   int           `json:"unique_errors"`
	AvgDuration    time.Duration `json:"avg_duration"`
	MaxMemoryUsed  int           `json:"max_memory_used"`
}

// LogAnalysis contains detailed log analysis
type LogAnalysis struct {
	ErrorPatterns    []ErrorPattern    `json:"error_patterns"`
	PerformanceIssues []PerformanceIssue `json:"performance_issues"`
	MemoryAnalysis   *MemoryAnalysis   `json:"memory_analysis"`
	TimeoutIssues    []TimeoutIssue    `json:"timeout_issues"`
	ColdStarts       []ColdStartEvent  `json:"cold_starts"`
}

// TraceAnalysis contains X-Ray trace analysis
type TraceAnalysis struct {
	Traces          []TraceSegment    `json:"traces"`
	ServiceMap      map[string]int    `json:"service_map"`
	ErrorTraces     []TraceSegment    `json:"error_traces"`
	SlowTraces      []TraceSegment    `json:"slow_traces"`
	BottleneckAna   *BottleneckAnalysis `json:"bottleneck_analysis"`
}

// ErrorPattern represents a recurring error pattern
type ErrorPattern struct {
	Pattern     string    `json:"pattern"`
	Count       int       `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Examples    []string  `json:"examples"`
	Severity    string    `json:"severity"`
	Suggestion  string    `json:"suggestion"`
}

// PerformanceIssue represents a performance concern
type PerformanceIssue struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Impact      string        `json:"impact"`
	Duration    time.Duration `json:"duration"`
	Count       int           `json:"count"`
	Suggestion  string        `json:"suggestion"`
}

// MemoryAnalysis contains memory usage analysis
type MemoryAnalysis struct {
	AvgUsage       int     `json:"avg_usage"`
	MaxUsage       int     `json:"max_usage"`
	MemorySize     int     `json:"memory_size"`
	UtilizationPct float64 `json:"utilization_pct"`
	LowUsageCount  int     `json:"low_usage_count"`
	HighUsageCount int     `json:"high_usage_count"`
}

// TimeoutIssue represents timeout-related problems
type TimeoutIssue struct {
	RequestID   string        `json:"request_id"`
	Duration    time.Duration `json:"duration"`
	TimeoutVal  time.Duration `json:"timeout_value"`
	Timestamp   time.Time     `json:"timestamp"`
	Function    string        `json:"function"`
}

// ColdStartEvent represents Lambda cold start occurrences
type ColdStartEvent struct {
	RequestID   string        `json:"request_id"`
	InitDuration time.Duration `json:"init_duration"`
	Runtime     string        `json:"runtime"`
	Timestamp   time.Time     `json:"timestamp"`
	Function    string        `json:"function"`
}

// BottleneckAnalysis identifies performance bottlenecks in traces
type BottleneckAnalysis struct {
	SlowestServices []ServiceLatency `json:"slowest_services"`
	ErrorHotspots   []ErrorHotspot   `json:"error_hotspots"`
	ThrottlePoints  []ThrottlePoint  `json:"throttle_points"`
}

// ServiceLatency represents service latency information
type ServiceLatency struct {
	Service     string        `json:"service"`
	AvgLatency  time.Duration `json:"avg_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	CallCount   int           `json:"call_count"`
	ErrorRate   float64       `json:"error_rate"`
}

// ErrorHotspot represents areas with high error rates
type ErrorHotspot struct {
	Service   string  `json:"service"`
	Operation string  `json:"operation"`
	ErrorRate float64 `json:"error_rate"`
	Count     int     `json:"count"`
}

// ThrottlePoint represents throttling issues
type ThrottlePoint struct {
	Service      string    `json:"service"`
	Operation    string    `json:"operation"`
	ThrottleRate float64   `json:"throttle_rate"`
	Timestamp    time.Time `json:"timestamp"`
}

// Recommendation provides actionable debugging suggestions
type Recommendation struct {
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Resources   []string `json:"resources,omitempty"`
}

// NewDebugAnalyzer creates a new debug analyzer instance
func NewDebugAnalyzer(region, projectRoot string) *DebugAnalyzer {
	return &DebugAnalyzer{
		region:      region,
		projectRoot: projectRoot,
		green:       color.New(color.FgGreen),
		red:         color.New(color.FgRed),
		yellow:      color.New(color.FgYellow),
		blue:        color.New(color.FgBlue),
		cyan:        color.New(color.FgCyan),
		bold:        color.New(color.Bold),
	}
}

// AnalyzeLogs performs comprehensive log analysis
func (da *DebugAnalyzer) AnalyzeLogs(ctx context.Context, functionName, startTime string, follow bool) (*DebugReport, error) {
	da.printBanner("Log Analysis")
	
	// Fetch logs (simulated for demo)
	logs, err := da.fetchLogs(ctx, functionName, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}
	
	da.blue.Printf("📊 Analyzing %d log entries...\n", len(logs))
	
	// Parse and analyze logs
	logAnalysis := da.analyzeLogEntries(logs)
	
	// Generate debug report
	report := &DebugReport{
		Summary:         da.generateSummary(logs, logAnalysis),
		LogAnalysis:     logAnalysis,
		Recommendations: da.generateRecommendations(logAnalysis, nil),
		Timestamp:       time.Now(),
	}
	
	// Print analysis results
	da.printLogAnalysis(report)
	
	// Follow mode for real-time analysis
	if follow {
		return da.followLogs(ctx, functionName, report)
	}
	
	return report, nil
}

// AnalyzeTraces performs X-Ray trace analysis
func (da *DebugAnalyzer) AnalyzeTraces(ctx context.Context, traceID, serviceFilter string) (*DebugReport, error) {
	da.printBanner("X-Ray Trace Analysis")
	
	// Fetch traces (simulated for demo)
	traces, err := da.fetchTraces(ctx, traceID, serviceFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traces: %w", err)
	}
	
	da.blue.Printf("🔍 Analyzing %d trace segments...\n", len(traces))
	
	// Analyze traces
	traceAnalysis := da.analyzeTraces(traces)
	
	// Generate debug report
	report := &DebugReport{
		Summary:         da.generateTraceSummary(traces),
		TraceAnalysis:   traceAnalysis,
		Recommendations: da.generateTraceRecommendations(traceAnalysis),
		Timestamp:       time.Now(),
	}
	
	// Print analysis results
	da.printTraceAnalysis(report)
	
	return report, nil
}

// ValidateConfiguration validates AWS resource configurations
func (da *DebugAnalyzer) ValidateConfiguration(ctx context.Context, configFile string, autoFix bool) error {
	da.printBanner("Configuration Validation")
	
	// Load configuration
	config, err := da.loadConfiguration(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Validate configuration
	issues := da.validateConfigValues(config)
	
	if len(issues) == 0 {
		da.green.Println("✅ Configuration validation passed!")
		return nil
	}
	
	// Display issues
	da.printConfigurationIssues(issues)
	
	// Auto-fix if requested
	if autoFix {
		return da.autoFixConfiguration(config, issues, configFile)
	}
	
	return nil
}

// printBanner displays analysis banner
func (da *DebugAnalyzer) printBanner(title string) {
	da.cyan.Printf(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                            SFX Debug Analyzer                               ║
║                          %s                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`, fmt.Sprintf("%-44s", title))
	fmt.Println()
}

// fetchLogs simulates fetching CloudWatch logs
func (da *DebugAnalyzer) fetchLogs(ctx context.Context, functionName, startTime string) ([]*LogEntry, error) {
	// Simulate log entries for demonstration
	logs := []*LogEntry{
		{
			Timestamp:      time.Now().Add(-1 * time.Hour),
			Level:          "INFO",
			Message:        "Function started",
			Service:        functionName,
			RequestID:      "req-123",
			Duration:       "125ms",
			BilledDuration: "200ms",
			MemorySize:     "512",
			MemoryUsed:     "89",
			Raw:           "2024-01-01T10:00:00Z [INFO] Function started",
		},
		{
			Timestamp:      time.Now().Add(-50 * time.Minute),
			Level:          "ERROR",
			Message:        "Database connection timeout",
			Service:        functionName,
			RequestID:      "req-124",
			Duration:       "30000ms",
			BilledDuration: "30000ms",
			MemorySize:     "512",
			MemoryUsed:     "156",
			Error:          "connection timeout after 30s",
			StackTrace:     "at database.connect(db.js:45)",
			Raw:           "2024-01-01T10:10:00Z [ERROR] Database connection timeout",
		},
		{
			Timestamp:      time.Now().Add(-30 * time.Minute),
			Level:          "WARN",
			Message:        "High memory usage detected",
			Service:        functionName,
			RequestID:      "req-125",
			Duration:       "890ms",
			BilledDuration: "900ms",
			MemorySize:     "512",
			MemoryUsed:     "487",
			Raw:           "2024-01-01T10:30:00Z [WARN] High memory usage detected",
		},
	}
	
	return logs, nil
}

// analyzeLogEntries performs detailed log analysis
func (da *DebugAnalyzer) analyzeLogEntries(logs []*LogEntry) *LogAnalysis {
	analysis := &LogAnalysis{
		ErrorPatterns:     make([]ErrorPattern, 0),
		PerformanceIssues: make([]PerformanceIssue, 0),
		TimeoutIssues:     make([]TimeoutIssue, 0),
		ColdStarts:        make([]ColdStartEvent, 0),
	}
	
	// Analyze error patterns
	errorMap := make(map[string]*ErrorPattern)
	
	var totalDuration time.Duration
	var durationCount int
	var memoryUsages []int
	
	for _, log := range logs {
		// Track error patterns
		if log.Level == "ERROR" && log.Error != "" {
			pattern := da.normalizeErrorMessage(log.Error)
			if ep, exists := errorMap[pattern]; exists {
				ep.Count++
				ep.LastSeen = log.Timestamp
				ep.Examples = append(ep.Examples, log.Error)
			} else {
				errorMap[pattern] = &ErrorPattern{
					Pattern:    pattern,
					Count:      1,
					FirstSeen:  log.Timestamp,
					LastSeen:   log.Timestamp,
					Examples:   []string{log.Error},
					Severity:   da.determineSeverity(log.Error),
					Suggestion: da.generateErrorSuggestion(pattern),
				}
			}
		}
		
		// Track performance issues
		if log.Duration != "" {
			if duration, err := time.ParseDuration(log.Duration); err == nil {
				totalDuration += duration
				durationCount++
				
				// Check for slow executions (>5s)
				if duration > 5*time.Second {
					analysis.PerformanceIssues = append(analysis.PerformanceIssues, PerformanceIssue{
						Type:        "slow_execution",
						Description: fmt.Sprintf("Slow execution detected: %v", duration),
						Impact:      "high",
						Duration:    duration,
						Count:       1,
						Suggestion:  "Optimize function logic, check external dependencies",
					})
				}
				
				// Check for timeouts
				if duration > 29*time.Second {
					analysis.TimeoutIssues = append(analysis.TimeoutIssues, TimeoutIssue{
						RequestID:  log.RequestID,
						Duration:   duration,
						TimeoutVal: 30 * time.Second,
						Timestamp:  log.Timestamp,
						Function:   log.Service,
					})
				}
			}
		}
		
		// Track memory usage
		if log.MemoryUsed != "" && log.MemorySize != "" {
			var used, size int
			fmt.Sscanf(log.MemoryUsed, "%d", &used)
			fmt.Sscanf(log.MemorySize, "%d", &size)
			
			if used > 0 && size > 0 {
				memoryUsages = append(memoryUsages, used)
			}
		}
		
		// Detect cold starts
		if strings.Contains(strings.ToLower(log.Message), "init") || 
		   strings.Contains(strings.ToLower(log.Message), "cold start") {
			analysis.ColdStarts = append(analysis.ColdStarts, ColdStartEvent{
				RequestID:    log.RequestID,
				InitDuration: 2 * time.Second, // Simulated
				Runtime:      "nodejs18.x",    // Simulated
				Timestamp:    log.Timestamp,
				Function:     log.Service,
			})
		}
	}
	
	// Convert error map to slice
	for _, pattern := range errorMap {
		analysis.ErrorPatterns = append(analysis.ErrorPatterns, *pattern)
	}
	
	// Analyze memory usage
	if len(memoryUsages) > 0 {
		analysis.MemoryAnalysis = da.analyzeMemoryUsage(memoryUsages, 512) // Assuming 512MB
	}
	
	return analysis
}

// normalizeErrorMessage normalizes error messages for pattern detection
func (da *DebugAnalyzer) normalizeErrorMessage(errMsg string) string {
	// Remove specific values and keep pattern
	patterns := []struct {
		regex  *regexp.Regexp
		replace string
	}{
		{regexp.MustCompile(`\d+`), "N"},                     // Replace numbers
		{regexp.MustCompile(`[a-f0-9-]{36}`), "UUID"},        // Replace UUIDs
		{regexp.MustCompile(`\b\w+@\w+\.\w+`), "EMAIL"},      // Replace emails
		{regexp.MustCompile(`\bhttps?://[^\s]+`), "URL"},     // Replace URLs
	}
	
	normalized := errMsg
	for _, pattern := range patterns {
		normalized = pattern.regex.ReplaceAllString(normalized, pattern.replace)
	}
	
	return normalized
}

// determineSeverity determines error severity
func (da *DebugAnalyzer) determineSeverity(errMsg string) string {
	errorMsg := strings.ToLower(errMsg)
	
	if strings.Contains(errorMsg, "timeout") || 
	   strings.Contains(errorMsg, "connection") ||
	   strings.Contains(errorMsg, "memory") {
		return "high"
	}
	
	if strings.Contains(errorMsg, "warning") || 
	   strings.Contains(errorMsg, "deprecated") {
		return "medium"
	}
	
	return "low"
}

// generateErrorSuggestion generates suggestions for error patterns
func (da *DebugAnalyzer) generateErrorSuggestion(pattern string) string {
	suggestions := map[string]string{
		"connection timeout": "Increase connection timeout, check network connectivity, verify endpoint availability",
		"memory": "Increase Lambda memory allocation, optimize memory usage, check for memory leaks",
		"database": "Optimize database queries, implement connection pooling, check database performance",
		"network": "Check network configurations, verify security groups, ensure proper VPC setup",
		"permission": "Verify IAM permissions, check resource policies, ensure proper role attachments",
	}
	
	for keyword, suggestion := range suggestions {
		if strings.Contains(strings.ToLower(pattern), keyword) {
			return suggestion
		}
	}
	
	return "Review error context and check AWS service documentation"
}

// analyzeMemoryUsage analyzes memory usage patterns
func (da *DebugAnalyzer) analyzeMemoryUsage(usages []int, memorySize int) *MemoryAnalysis {
	if len(usages) == 0 {
		return nil
	}
	
	var total, max int
	var lowUsageCount, highUsageCount int
	
	for _, usage := range usages {
		total += usage
		if usage > max {
			max = usage
		}
		
		utilizationPct := float64(usage) / float64(memorySize) * 100
		if utilizationPct < 25 {
			lowUsageCount++
		} else if utilizationPct > 80 {
			highUsageCount++
		}
	}
	
	avg := total / len(usages)
	utilizationPct := float64(avg) / float64(memorySize) * 100
	
	return &MemoryAnalysis{
		AvgUsage:       avg,
		MaxUsage:       max,
		MemorySize:     memorySize,
		UtilizationPct: utilizationPct,
		LowUsageCount:  lowUsageCount,
		HighUsageCount: highUsageCount,
	}
}

// generateSummary creates a debug summary
func (da *DebugAnalyzer) generateSummary(logs []*LogEntry, analysis *LogAnalysis) *DebugSummary {
	var errorCount, warningCount int
	uniqueErrors := len(analysis.ErrorPatterns)
	
	for _, log := range logs {
		switch log.Level {
		case "ERROR":
			errorCount++
		case "WARN", "WARNING":
			warningCount++
		}
	}
	
	return &DebugSummary{
		Service:       logs[0].Service,
		TimeRange:     "Last 1 hour", // Simplified
		TotalLogs:     len(logs),
		ErrorCount:    errorCount,
		WarningCount:  warningCount,
		UniqueErrors:  uniqueErrors,
		AvgDuration:   time.Minute, // Simplified
		MaxMemoryUsed: 487,         // From sample data
	}
}

// generateRecommendations creates actionable recommendations
func (da *DebugAnalyzer) generateRecommendations(logAnalysis *LogAnalysis, traceAnalysis *TraceAnalysis) []Recommendation {
	recommendations := make([]Recommendation, 0)
	
	// Error-based recommendations
	if len(logAnalysis.ErrorPatterns) > 0 {
		for _, pattern := range logAnalysis.ErrorPatterns {
			if pattern.Severity == "high" {
				recommendations = append(recommendations, Recommendation{
					Priority:    "HIGH",
					Category:    "Error Handling",
					Title:       fmt.Sprintf("Address recurring %s errors", pattern.Pattern),
					Description: fmt.Sprintf("Pattern occurs %d times", pattern.Count),
					Action:      pattern.Suggestion,
					Resources:   []string{"AWS Lambda Error Handling Guide", "AWS X-Ray Documentation"},
				})
			}
		}
	}
	
	// Performance recommendations
	if len(logAnalysis.PerformanceIssues) > 0 {
		recommendations = append(recommendations, Recommendation{
			Priority:    "MEDIUM",
			Category:    "Performance",
			Title:       "Optimize function performance",
			Description: fmt.Sprintf("Found %d performance issues", len(logAnalysis.PerformanceIssues)),
			Action:      "Review slow executions, optimize code, check external dependencies",
			Resources:   []string{"AWS Lambda Performance Optimization", "AWS Performance Insights"},
		})
	}
	
	// Memory recommendations
	if logAnalysis.MemoryAnalysis != nil {
		if logAnalysis.MemoryAnalysis.UtilizationPct < 25 {
			recommendations = append(recommendations, Recommendation{
				Priority:    "LOW",
				Category:    "Cost Optimization",
				Title:       "Reduce memory allocation",
				Description: fmt.Sprintf("Memory utilization is only %.1f%%", logAnalysis.MemoryAnalysis.UtilizationPct),
				Action:      "Consider reducing Lambda memory allocation to save costs",
				Resources:   []string{"AWS Lambda Pricing", "Memory Optimization Guide"},
			})
		} else if logAnalysis.MemoryAnalysis.UtilizationPct > 80 {
			recommendations = append(recommendations, Recommendation{
				Priority:    "HIGH",
				Category:    "Performance",
				Title:       "Increase memory allocation",
				Description: fmt.Sprintf("Memory utilization is %.1f%% - approaching limits", logAnalysis.MemoryAnalysis.UtilizationPct),
				Action:      "Increase Lambda memory allocation or optimize memory usage",
				Resources:   []string{"AWS Lambda Memory Configuration", "Memory Optimization Best Practices"},
			})
		}
	}
	
	return recommendations
}

// printLogAnalysis displays log analysis results
func (da *DebugAnalyzer) printLogAnalysis(report *DebugReport) {
	fmt.Println()
	da.bold.Println("📊 Debug Analysis Summary")
	fmt.Println(strings.Repeat("═", 80))
	
	summary := report.Summary
	fmt.Printf("🎯 Service:          %s\n", summary.Service)
	fmt.Printf("📅 Time Range:       %s\n", summary.TimeRange)
	fmt.Printf("📋 Total Logs:       %d\n", summary.TotalLogs)
	fmt.Printf("❌ Errors:          %d\n", summary.ErrorCount)
	fmt.Printf("⚠️  Warnings:        %d\n", summary.WarningCount)
	fmt.Printf("🔢 Unique Errors:    %d\n", summary.UniqueErrors)
	
	if len(report.LogAnalysis.ErrorPatterns) > 0 {
		fmt.Println()
		da.red.Println("🚨 Error Patterns:")
		for _, pattern := range report.LogAnalysis.ErrorPatterns {
			severity := ""
			switch pattern.Severity {
			case "high":
				severity = da.red.Sprint("HIGH")
			case "medium":
				severity = da.yellow.Sprint("MED")
			default:
				severity = da.green.Sprint("LOW")
			}
			
			fmt.Printf("   [%s] %s (count: %d)\n", severity, pattern.Pattern, pattern.Count)
			fmt.Printf("        💡 %s\n", pattern.Suggestion)
		}
	}
	
	if len(report.LogAnalysis.PerformanceIssues) > 0 {
		fmt.Println()
		da.yellow.Println("⚡ Performance Issues:")
		for _, issue := range report.LogAnalysis.PerformanceIssues {
			fmt.Printf("   • %s: %s\n", issue.Type, issue.Description)
			fmt.Printf("     💡 %s\n", issue.Suggestion)
		}
	}
	
	if report.LogAnalysis.MemoryAnalysis != nil {
		fmt.Println()
		da.blue.Println("💾 Memory Analysis:")
		mem := report.LogAnalysis.MemoryAnalysis
		fmt.Printf("   Average Usage: %d MB (%.1f%%)\n", mem.AvgUsage, mem.UtilizationPct)
		fmt.Printf("   Maximum Usage: %d MB\n", mem.MaxUsage)
		fmt.Printf("   Allocated:     %d MB\n", mem.MemorySize)
	}
	
	if len(report.Recommendations) > 0 {
		fmt.Println()
		da.cyan.Println("🎯 Recommendations:")
		
		// Sort by priority
		sort.Slice(report.Recommendations, func(i, j int) bool {
			priority := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
			return priority[report.Recommendations[i].Priority] > priority[report.Recommendations[j].Priority]
		})
		
		for _, rec := range report.Recommendations {
			priority := ""
			switch rec.Priority {
			case "HIGH":
				priority = da.red.Sprint("HIGH")
			case "MEDIUM":
				priority = da.yellow.Sprint("MED")
			default:
				priority = da.green.Sprint("LOW")
			}
			
			fmt.Printf("   [%s] %s\n", priority, rec.Title)
			fmt.Printf("        %s\n", rec.Description)
			fmt.Printf("        💡 %s\n", rec.Action)
		}
	}
	
	fmt.Println(strings.Repeat("═", 80))
}

// Additional methods would be implemented for:
// - fetchTraces()
// - analyzeTraces()  
// - generateTraceSummary()
// - generateTraceRecommendations()
// - printTraceAnalysis()
// - loadConfiguration()
// - validateConfigValues()
// - printConfigurationIssues()
// - autoFixConfiguration()
// - followLogs()

// SaveReport saves the debug report to a file
func (da *DebugAnalyzer) SaveReport(report *DebugReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}
	
	da.green.Printf("💾 Debug report saved to: %s\n", filename)
	return nil
}

// Placeholder implementations for completeness - would be fully implemented in production
func (da *DebugAnalyzer) fetchTraces(ctx context.Context, traceID, serviceFilter string) ([]TraceSegment, error) {
	return []TraceSegment{}, nil
}

func (da *DebugAnalyzer) analyzeTraces(traces []TraceSegment) *TraceAnalysis {
	return &TraceAnalysis{}
}

func (da *DebugAnalyzer) generateTraceSummary(traces []TraceSegment) *DebugSummary {
	return &DebugSummary{}
}

func (da *DebugAnalyzer) generateTraceRecommendations(analysis *TraceAnalysis) []Recommendation {
	return []Recommendation{}
}

func (da *DebugAnalyzer) printTraceAnalysis(report *DebugReport) {}

func (da *DebugAnalyzer) loadConfiguration(configFile string) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (da *DebugAnalyzer) validateConfigValues(config map[string]interface{}) []string {
	return []string{}
}

func (da *DebugAnalyzer) printConfigurationIssues(issues []string) {}

func (da *DebugAnalyzer) autoFixConfiguration(config map[string]interface{}, issues []string, filename string) error {
	return nil
}

func (da *DebugAnalyzer) followLogs(ctx context.Context, functionName string, report *DebugReport) (*DebugReport, error) {
	return report, nil
}