// Package snapshot provides snapshot testing functionality for serverless tests
package snapshot

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "testing"

    "github.com/pmezard/go-difflib/difflib"
    "vasdeference"
)

// Options configures snapshot behavior
type Options struct {
    // SnapshotDir is the directory where snapshots are stored
    SnapshotDir string

    // UpdateSnapshots forces update of all snapshots
    UpdateSnapshots bool

    // Sanitizers are functions that clean dynamic data before comparison
    Sanitizers []Sanitizer

    // DiffContextLines controls how many context lines to show in diffs
    DiffContextLines int

    // JSONIndent formats JSON snapshots with indentation
    JSONIndent bool
}

// Sanitizer cleans dynamic data from snapshots
type Sanitizer func(data []byte) []byte

// Snapshot manages snapshot testing
type Snapshot struct {
    t        sfx.TestingT
    options  Options
    testName string
}

// New creates a new snapshot tester
func New(t sfx.TestingT, opts ...Options) *Snapshot {
    options := Options{
        SnapshotDir:      "testdata/snapshots",
        DiffContextLines: 3,
        JSONIndent:       true,
    }
    
    if len(opts) > 0 {
        options = opts[0]
    }
    
    // Check for update flag
    if os.Getenv("UPDATE_SNAPSHOTS") == "true" || os.Getenv("UPDATE_SNAPSHOTS") == "1" {
        options.UpdateSnapshots = true
    }
    
    // Ensure snapshot directory exists
    if err := os.MkdirAll(options.SnapshotDir, 0755); err != nil {
        t.Errorf("Failed to create snapshot directory: %v", err)
        t.FailNow()
    }
    
    // Get test name - try to extract from testing.TB if available
    testName := "unknown"
    if tb, ok := t.(testing.TB); ok {
        testName = tb.Name()
    }
    
    return &Snapshot{
        t:        t,
        options:  options,
        testName: sanitizeTestName(testName),
    }
}

// Match compares data against a snapshot
func (s *Snapshot) Match(name string, data interface{}) {
    // Convert data to bytes
    dataBytes, err := s.toBytes(data)
    if err != nil {
        s.t.Errorf("Failed to convert data to bytes: %v", err)
        s.t.FailNow()
    }
    
    // Apply sanitizers
    for _, sanitizer := range s.options.Sanitizers {
        dataBytes = sanitizer(dataBytes)
    }
    
    // Generate snapshot filename
    filename := s.getSnapshotPath(name)
    
    // Update snapshot if requested
    if s.options.UpdateSnapshots {
        s.updateSnapshot(filename, dataBytes)
        return
    }
    
    // Read existing snapshot
    expected, err := os.ReadFile(filename)
    if os.IsNotExist(err) {
        // Create new snapshot
        s.updateSnapshot(filename, dataBytes)
        return
    }
    if err != nil {
        s.t.Errorf("Failed to read snapshot: %v", err)
        s.t.FailNow()
    }
    
    // Compare snapshots
    if !bytes.Equal(expected, dataBytes) {
        diff := s.generateDiff(expected, dataBytes, filename)
        s.t.Errorf("Snapshot mismatch for %s:\n%s\n\nRun with UPDATE_SNAPSHOTS=1 to update", name, diff)
    }
}

// MatchJSON matches JSON data with automatic formatting
func (s *Snapshot) MatchJSON(name string, data interface{}) {
    // Ensure JSON formatting
    jsonBytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        s.t.Errorf("Failed to marshal JSON: %v", err)
        s.t.FailNow()
    }
    
    s.Match(name, jsonBytes)
}

// MatchLambdaResponse matches a Lambda response snapshot
func (s *Snapshot) MatchLambdaResponse(name string, payload []byte, sanitizers ...Sanitizer) {
    // Parse and format response
    var response interface{}
    if err := json.Unmarshal(payload, &response); err != nil {
        // If not JSON, use raw payload
        s.Match(name, payload)
        return
    }
    
    // Add Lambda-specific sanitizers
    allSanitizers := append(s.options.Sanitizers, sanitizers...)
    allSanitizers = append(allSanitizers, 
        SanitizeLambdaRequestID(),
        SanitizeTimestamps(),
    )
    
    snapshot := s.WithSanitizers(allSanitizers...)
    snapshot.MatchJSON(name, response)
}

// MatchDynamoDBItems matches DynamoDB items with attribute value handling
func (s *Snapshot) MatchDynamoDBItems(name string, items []map[string]interface{}) {
    // Convert DynamoDB items to readable format
    readable := make([]map[string]interface{}, len(items))
    for i, item := range items {
        readable[i] = make(map[string]interface{})
        for k, v := range item {
            readable[i][k] = v
        }
    }
    
    // Add DynamoDB-specific sanitizers
    sanitizers := append(s.options.Sanitizers,
        SanitizeTimestamps(),
        SanitizeDynamoDBMetadata(),
    )
    
    snapshot := s.WithSanitizers(sanitizers...)
    snapshot.MatchJSON(name, readable)
}

// MatchStepFunctionExecution matches Step Functions execution output
func (s *Snapshot) MatchStepFunctionExecution(name string, output interface{}) {
    // Add Step Functions-specific sanitizers
    sanitizers := append(s.options.Sanitizers,
        SanitizeExecutionArn(),
        SanitizeTimestamps(),
        SanitizeUUIDs(),
    )
    
    snapshot := s.WithSanitizers(sanitizers...)
    snapshot.MatchJSON(name, output)
}

// WithSanitizers returns a new snapshot with additional sanitizers
func (s *Snapshot) WithSanitizers(sanitizers ...Sanitizer) *Snapshot {
    newSnapshot := *s
    newSnapshot.options.Sanitizers = append(
        append([]Sanitizer{}, s.options.Sanitizers...),
        sanitizers...,
    )
    return &newSnapshot
}

// Helper methods

func (s *Snapshot) toBytes(data interface{}) ([]byte, error) {
    switch v := data.(type) {
    case []byte:
        return v, nil
    case string:
        return []byte(v), nil
    case io.Reader:
        return io.ReadAll(v)
    default:
        if s.options.JSONIndent {
            return json.MarshalIndent(data, "", "  ")
        }
        return json.Marshal(data)
    }
}

func (s *Snapshot) getSnapshotPath(name string) string {
    filename := fmt.Sprintf("%s_%s.snap", s.testName, sanitizeSnapshotName(name))
    return filepath.Join(s.options.SnapshotDir, filename)
}

func (s *Snapshot) updateSnapshot(filename string, data []byte) {
    err := os.WriteFile(filename, data, 0644)
    if err != nil {
        s.t.Errorf("Failed to write snapshot: %v", err)
        s.t.FailNow()
    }
}

func (s *Snapshot) generateDiff(expected, actual []byte, filename string) string {
    diff := difflib.UnifiedDiff{
        A:        difflib.SplitLines(string(expected)),
        B:        difflib.SplitLines(string(actual)),
        FromFile: filename,
        ToFile:   "actual",
        Context:  s.options.DiffContextLines,
    }
    
    text, _ := difflib.GetUnifiedDiffString(diff)
    return text
}

// Sanitizer functions

// SanitizeTimestamps replaces timestamp patterns with placeholders
func SanitizeTimestamps() Sanitizer {
    return func(data []byte) []byte {
        // ISO 8601 timestamps
        re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`)
        data = re.ReplaceAll(data, []byte("<TIMESTAMP>"))
        
        // Unix timestamps (10-13 digits)
        re = regexp.MustCompile(`"(timestamp|createdAt|updatedAt|lastModified)":\s*\d{10,13}`)
        data = re.ReplaceAll(data, []byte(`"$1": "<TIMESTAMP>"`))
        
        return data
    }
}

// SanitizeUUIDs replaces UUID patterns with placeholders
func SanitizeUUIDs() Sanitizer {
    return func(data []byte) []byte {
        re := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
        return re.ReplaceAll(data, []byte("<UUID>"))
    }
}

// SanitizeLambdaRequestID replaces Lambda request IDs
func SanitizeLambdaRequestID() Sanitizer {
    return func(data []byte) []byte {
        re := regexp.MustCompile(`"(requestId|RequestId)":\s*"[^"]+?"`)
        return re.ReplaceAll(data, []byte(`"$1": "<REQUEST_ID>"`))
    }
}

// SanitizeExecutionArn replaces Step Functions execution ARNs
func SanitizeExecutionArn() Sanitizer {
    return func(data []byte) []byte {
        re := regexp.MustCompile(`arn:aws:states:[^:]+:\d+:execution:[^:]+:[^"]+`)
        return re.ReplaceAll(data, []byte("<EXECUTION_ARN>"))
    }
}

// SanitizeDynamoDBMetadata removes DynamoDB metadata fields
func SanitizeDynamoDBMetadata() Sanitizer {
    return func(data []byte) []byte {
        // Remove consumed capacity, scanned count, etc.
        re := regexp.MustCompile(`"(ConsumedCapacity|ScannedCount|Count)":\s*[^,}]+,?`)
        return re.ReplaceAll(data, []byte(""))
    }
}

// SanitizeCustom creates a custom sanitizer with regex
func SanitizeCustom(pattern, replacement string) Sanitizer {
    re := regexp.MustCompile(pattern)
    return func(data []byte) []byte {
        return re.ReplaceAll(data, []byte(replacement))
    }
}

// SanitizeField replaces specific field values
func SanitizeField(fieldName, replacement string) Sanitizer {
    return func(data []byte) []byte {
        pattern := fmt.Sprintf(`"%s":\s*"[^"]*"`, fieldName)
        re := regexp.MustCompile(pattern)
        return re.ReplaceAll(data, []byte(fmt.Sprintf(`"%s": "%s"`, fieldName, replacement)))
    }
}

// Helper functions

func sanitizeTestName(name string) string {
    // Remove test prefix and sanitize
    name = strings.TrimPrefix(name, "Test")
    name = strings.ReplaceAll(name, "/", "_")
    name = strings.ReplaceAll(name, " ", "_")
    return name
}

func sanitizeSnapshotName(name string) string {
    // Replace non-alphanumeric characters
    re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
    return re.ReplaceAllString(name, "_")
}

// Global snapshot instance for simple usage
var defaultSnapshot *Snapshot

// Match uses a default snapshot instance
func Match(t sfx.TestingT, name string, data interface{}) {
    if defaultSnapshot == nil || defaultSnapshot.t != t {
        defaultSnapshot = New(t)
    }
    defaultSnapshot.Match(name, data)
}

// MatchJSON uses a default snapshot instance for JSON
func MatchJSON(t sfx.TestingT, name string, data interface{}) {
    if defaultSnapshot == nil || defaultSnapshot.t != t {
        defaultSnapshot = New(t)
    }
    defaultSnapshot.MatchJSON(name, data)
}

// NewE creates a new snapshot tester with error return (Terratest pattern)
func NewE(t sfx.TestingT, opts ...Options) (*Snapshot, error) {
    options := Options{
        SnapshotDir:      "testdata/snapshots",
        DiffContextLines: 3,
        JSONIndent:       true,
    }
    
    if len(opts) > 0 {
        options = opts[0]
    }
    
    // Check for update flag
    if os.Getenv("UPDATE_SNAPSHOTS") == "true" || os.Getenv("UPDATE_SNAPSHOTS") == "1" {
        options.UpdateSnapshots = true
    }
    
    // Ensure snapshot directory exists
    if err := os.MkdirAll(options.SnapshotDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
    }
    
    // Get test name - try to extract from testing.TB if available
    testName := "unknown"
    if tb, ok := t.(testing.TB); ok {
        testName = tb.Name()
    }
    
    return &Snapshot{
        t:        t,
        options:  options,
        testName: sanitizeTestName(testName),
    }, nil
}

// MatchE compares data against a snapshot with error return
func (s *Snapshot) MatchE(name string, data interface{}) error {
    // Convert data to bytes
    dataBytes, err := s.toBytes(data)
    if err != nil {
        return fmt.Errorf("failed to convert data to bytes: %w", err)
    }
    
    // Apply sanitizers
    for _, sanitizer := range s.options.Sanitizers {
        dataBytes = sanitizer(dataBytes)
    }
    
    // Generate snapshot filename
    filename := s.getSnapshotPath(name)
    
    // Update snapshot if requested
    if s.options.UpdateSnapshots {
        return s.updateSnapshotE(filename, dataBytes)
    }
    
    // Read existing snapshot
    expected, err := os.ReadFile(filename)
    if os.IsNotExist(err) {
        // Create new snapshot
        return s.updateSnapshotE(filename, dataBytes)
    }
    if err != nil {
        return fmt.Errorf("failed to read snapshot: %w", err)
    }
    
    // Compare snapshots
    if !bytes.Equal(expected, dataBytes) {
        diff := s.generateDiff(expected, dataBytes, filename)
        return fmt.Errorf("snapshot mismatch for %s:\n%s\n\nRun with UPDATE_SNAPSHOTS=1 to update", name, diff)
    }
    
    return nil
}

// MatchJSONE matches JSON data with error return
func (s *Snapshot) MatchJSONE(name string, data interface{}) error {
    // Ensure JSON formatting
    jsonBytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %w", err)
    }
    
    return s.MatchE(name, jsonBytes)
}

func (s *Snapshot) updateSnapshotE(filename string, data []byte) error {
    return os.WriteFile(filename, data, 0644)
}