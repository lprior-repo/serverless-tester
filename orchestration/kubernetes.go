// Kubernetes integration for container orchestration testing
package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// KubernetesOrchestrator provides Kubernetes-based testing orchestration
type KubernetesOrchestrator struct {
	config     *KubernetesConfig
	client     KubernetesClient
	namespace  string
	logger     Logger
	
	// Job management
	jobs       map[string]*TestJob
	jobHistory []JobExecution
}

// KubernetesConfig contains Kubernetes cluster configuration
type KubernetesConfig struct {
	KubeconfigPath    string            `json:"kubeconfig_path"`
	ClusterEndpoint   string            `json:"cluster_endpoint"`
	Namespace         string            `json:"namespace"`
	ServiceAccount    string            `json:"service_account"`
	ImageRegistry     string            `json:"image_registry"`
	DefaultResources  ResourceRequests  `json:"default_resources"`
	NetworkPolicy     *NetworkPolicy    `json:"network_policy,omitempty"`
	SecurityContext   *SecurityContext  `json:"security_context,omitempty"`
	NodeSelector      map[string]string `json:"node_selector,omitempty"`
	Tolerations       []Toleration      `json:"tolerations,omitempty"`
	Affinity          *Affinity         `json:"affinity,omitempty"`
}

// ResourceRequests defines CPU and memory requests/limits
type ResourceRequests struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

// NetworkPolicy defines network policies for test jobs
type NetworkPolicy struct {
	Name           string                  `json:"name"`
	PodSelector    map[string]string       `json:"pod_selector"`
	PolicyTypes    []string                `json:"policy_types"`
	Ingress        []NetworkPolicyRule     `json:"ingress,omitempty"`
	Egress         []NetworkPolicyRule     `json:"egress,omitempty"`
}

// NetworkPolicyRule defines ingress/egress rules
type NetworkPolicyRule struct {
	From     []NetworkPolicyPeer `json:"from,omitempty"`
	To       []NetworkPolicyPeer `json:"to,omitempty"`
	Ports    []NetworkPolicyPort `json:"ports,omitempty"`
}

// NetworkPolicyPeer defines network policy peers
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"pod_selector,omitempty"`
	NamespaceSelector map[string]string `json:"namespace_selector,omitempty"`
}

// NetworkPolicyPort defines network policy ports
type NetworkPolicyPort struct {
	Protocol string `json:"protocol,omitempty"`
	Port     string `json:"port,omitempty"`
}

// SecurityContext defines security constraints for test jobs
type SecurityContext struct {
	RunAsUser          *int64 `json:"run_as_user,omitempty"`
	RunAsGroup         *int64 `json:"run_as_group,omitempty"`
	RunAsNonRoot       *bool  `json:"run_as_non_root,omitempty"`
	ReadOnlyRootFS     *bool  `json:"read_only_root_fs,omitempty"`
	AllowPrivilegeEscalation *bool `json:"allow_privilege_escalation,omitempty"`
}

// Toleration defines node taints that pods can tolerate
type Toleration struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// Affinity defines node affinity rules
type Affinity struct {
	NodeAffinity    *NodeAffinity    `json:"node_affinity,omitempty"`
	PodAffinity     *PodAffinity     `json:"pod_affinity,omitempty"`
	PodAntiAffinity *PodAntiAffinity `json:"pod_anti_affinity,omitempty"`
}

// NodeAffinity defines node affinity rules
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution *NodeSelector `json:"required_during_scheduling_ignored_during_execution,omitempty"`
}

// NodeSelector defines node selection requirements
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"node_selector_terms"`
}

// NodeSelectorTerm defines node selector terms
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"match_expressions,omitempty"`
}

// NodeSelectorRequirement defines node selector requirements
type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

// PodAffinity defines pod affinity rules
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"required_during_scheduling_ignored_during_execution,omitempty"`
}

// PodAntiAffinity defines pod anti-affinity rules
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"required_during_scheduling_ignored_during_execution,omitempty"`
}

// PodAffinityTerm defines pod affinity terms
type PodAffinityTerm struct {
	LabelSelector *LabelSelector `json:"label_selector,omitempty"`
	TopologyKey   string         `json:"topology_key"`
}

// LabelSelector defines label selection criteria
type LabelSelector struct {
	MatchLabels      map[string]string `json:"match_labels,omitempty"`
	MatchExpressions []LabelSelectorRequirement `json:"match_expressions,omitempty"`
}

// LabelSelectorRequirement defines label selector requirements
type LabelSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

// TestJob represents a Kubernetes job for testing
type TestJob struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Image           string            `json:"image"`
	Command         []string          `json:"command"`
	Args            []string          `json:"args"`
	Env             map[string]string `json:"env"`
	Resources       ResourceRequests  `json:"resources"`
	RestartPolicy   string            `json:"restart_policy"`
	BackoffLimit    *int32            `json:"backoff_limit,omitempty"`
	ActiveDeadline  *int64            `json:"active_deadline,omitempty"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	NodeSelector    map[string]string `json:"node_selector,omitempty"`
	Tolerations     []Toleration      `json:"tolerations,omitempty"`
	Affinity        *Affinity         `json:"affinity,omitempty"`
	SecurityContext *SecurityContext  `json:"security_context,omitempty"`
	
	// Test-specific fields
	TestType        TestJobType       `json:"test_type"`
	TestConfig      map[string]interface{} `json:"test_config"`
	ExpectedDuration time.Duration    `json:"expected_duration"`
	MaxRetries      int              `json:"max_retries"`
	DependsOn       []string         `json:"depends_on"`
	
	// State
	Status          JobStatus        `json:"status"`
	CreatedAt       time.Time        `json:"created_at"`
	StartedAt       *time.Time       `json:"started_at,omitempty"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
	FailedAt        *time.Time       `json:"failed_at,omitempty"`
}

// TestJobType defines the type of test job
type TestJobType string

const (
	TestJobTypeUnitTest        TestJobType = "unit_test"
	TestJobTypeIntegrationTest TestJobType = "integration_test"
	TestJobTypeE2ETest         TestJobType = "e2e_test"
	TestJobTypeLoadTest        TestJobType = "load_test"
	TestJobTypeSecurityTest    TestJobType = "security_test"
	TestJobTypePerformanceTest TestJobType = "performance_test"
	TestJobTypeChaosTest       TestJobType = "chaos_test"
	TestJobTypeCustom          TestJobType = "custom"
)

// JobStatus defines the status of a test job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusTimeout   JobStatus = "timeout"
)

// JobExecution represents an execution of a test job
type JobExecution struct {
	JobID          string            `json:"job_id"`
	ExecutionID    string            `json:"execution_id"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        *time.Time        `json:"end_time,omitempty"`
	Duration       time.Duration     `json:"duration"`
	Status         JobStatus         `json:"status"`
	ExitCode       *int32            `json:"exit_code,omitempty"`
	Logs           []string          `json:"logs"`
	Metrics        *JobMetrics       `json:"metrics,omitempty"`
	Resources      *ResourceUsage    `json:"resources,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	Artifacts      []JobArtifact     `json:"artifacts"`
}

// JobMetrics contains performance metrics for a job execution
type JobMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    int64   `json:"memory_usage"`
	NetworkIn      int64   `json:"network_in"`
	NetworkOut     int64   `json:"network_out"`
	DiskRead       int64   `json:"disk_read"`
	DiskWrite      int64   `json:"disk_write"`
	TestsRun       int     `json:"tests_run"`
	TestsPassed    int     `json:"tests_passed"`
	TestsFailed    int     `json:"tests_failed"`
	TestsSkipped   int     `json:"tests_skipped"`
	Coverage       float64 `json:"coverage"`
}

// ResourceUsage contains resource usage information
type ResourceUsage struct {
	CPURequest     string    `json:"cpu_request"`
	CPULimit       string    `json:"cpu_limit"`
	CPUUsage       string    `json:"cpu_usage"`
	MemoryRequest  string    `json:"memory_request"`
	MemoryLimit    string    `json:"memory_limit"`
	MemoryUsage    string    `json:"memory_usage"`
	NetworkPolicy  string    `json:"network_policy"`
	StorageUsage   string    `json:"storage_usage"`
	EphemeralStorage string  `json:"ephemeral_storage"`
}

// JobArtifact represents an artifact produced by a test job
type JobArtifact struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"` // log, report, coverage, binary, etc.
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	URL         string    `json:"url,omitempty"`
}

// KubernetesClient interface for Kubernetes operations
type KubernetesClient interface {
	// Job operations
	CreateJob(ctx context.Context, job *TestJob) error
	GetJob(ctx context.Context, namespace, name string) (*TestJob, error)
	DeleteJob(ctx context.Context, namespace, name string) error
	ListJobs(ctx context.Context, namespace string, labels map[string]string) ([]*TestJob, error)
	WatchJob(ctx context.Context, namespace, name string) (<-chan JobEvent, error)
	
	// Pod operations
	GetJobPods(ctx context.Context, namespace, jobName string) ([]Pod, error)
	GetPodLogs(ctx context.Context, namespace, podName string, follow bool) (io.ReadCloser, error)
	GetPodMetrics(ctx context.Context, namespace, podName string) (*JobMetrics, error)
	
	// Resource operations
	GetResourceUsage(ctx context.Context, namespace string) (*NamespaceResourceUsage, error)
	
	// Cluster operations
	GetClusterInfo() (*ClusterInfo, error)
	GetNodeResources() ([]*NodeResources, error)
}

// JobEvent represents a job state change event
type JobEvent struct {
	Type      string    `json:"type"` // ADDED, MODIFIED, DELETED
	Job       *TestJob  `json:"job"`
	Timestamp time.Time `json:"timestamp"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Phase     string            `json:"phase"`
	NodeName  string            `json:"node_name"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	StartedAt *time.Time        `json:"started_at,omitempty"`
}

// NamespaceResourceUsage represents resource usage for a namespace
type NamespaceResourceUsage struct {
	Namespace     string `json:"namespace"`
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	CPUUsage      string `json:"cpu_usage"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
	MemoryUsage   string `json:"memory_usage"`
	PodCount      int    `json:"pod_count"`
}

// ClusterInfo contains information about the Kubernetes cluster
type ClusterInfo struct {
	Version       string `json:"version"`
	NodeCount     int    `json:"node_count"`
	NamespaceCount int   `json:"namespace_count"`
	PodCount      int    `json:"pod_count"`
	ServiceCount  int    `json:"service_count"`
}

// NodeResources represents resource information for a cluster node
type NodeResources struct {
	Name              string            `json:"name"`
	Labels            map[string]string `json:"labels"`
	Allocatable       ResourceList      `json:"allocatable"`
	Capacity          ResourceList      `json:"capacity"`
	AllocatedPods     int               `json:"allocated_pods"`
	MaxPods           int               `json:"max_pods"`
	Ready             bool              `json:"ready"`
}

// ResourceList represents a list of resources
type ResourceList struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	EphemeralStorage string `json:"ephemeral_storage"`
	Pods             string `json:"pods"`
}

// NewKubernetesOrchestrator creates a new Kubernetes orchestrator
func NewKubernetesOrchestrator(config *KubernetesConfig, client KubernetesClient, logger Logger) *KubernetesOrchestrator {
	return &KubernetesOrchestrator{
		config:     config,
		client:     client,
		namespace:  config.Namespace,
		logger:     logger,
		jobs:       make(map[string]*TestJob),
		jobHistory: make([]JobExecution, 0),
	}
}

// SubmitTestJob submits a test job to Kubernetes
func (ko *KubernetesOrchestrator) SubmitTestJob(ctx context.Context, job *TestJob) (*JobExecution, error) {
	ko.logger.Info("Submitting test job", "job_id", job.ID, "name", job.Name, "type", job.TestType)
	
	// Set defaults from configuration
	ko.applyDefaultsToJob(job)
	
	// Validate job specification
	if err := ko.validateJob(job); err != nil {
		return nil, fmt.Errorf("job validation failed: %w", err)
	}
	
	// Check dependencies
	if err := ko.checkJobDependencies(ctx, job); err != nil {
		return nil, fmt.Errorf("dependency check failed: %w", err)
	}
	
	// Create job execution record
	execution := &JobExecution{
		JobID:       job.ID,
		ExecutionID: fmt.Sprintf("%s-%d", job.ID, time.Now().Unix()),
		StartTime:   time.Now(),
		Status:      JobStatusPending,
		Artifacts:   make([]JobArtifact, 0),
	}
	
	// Submit to Kubernetes
	if err := ko.client.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes job: %w", err)
	}
	
	// Store job and execution
	ko.jobs[job.ID] = job
	ko.jobHistory = append(ko.jobHistory, *execution)
	
	// Start monitoring the job
	go ko.monitorJob(ctx, job, execution)
	
	return execution, nil
}

// SubmitTestJobBatch submits multiple test jobs as a batch
func (ko *KubernetesOrchestrator) SubmitTestJobBatch(ctx context.Context, jobs []*TestJob, config *BatchConfig) ([]*JobExecution, error) {
	ko.logger.Info("Submitting test job batch", "job_count", len(jobs))
	
	if config == nil {
		config = &BatchConfig{
			MaxConcurrent: 5,
			FailFast:      false,
		}
	}
	
	executions := make([]*JobExecution, 0, len(jobs))
	errors := make([]error, 0)
	
	// Use semaphore to limit concurrent job submissions
	semaphore := make(chan struct{}, config.MaxConcurrent)
	resultChan := make(chan batchResult, len(jobs))
	
	// Submit jobs concurrently
	for _, job := range jobs {
		go func(j *TestJob) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			execution, err := ko.SubmitTestJob(ctx, j)
			resultChan <- batchResult{execution: execution, err: err}
		}(job)
	}
	
	// Collect results
	for i := 0; i < len(jobs); i++ {
		result := <-resultChan
		if result.err != nil {
			errors = append(errors, result.err)
			if config.FailFast {
				break
			}
		} else {
			executions = append(executions, result.execution)
		}
	}
	
	if len(errors) > 0 {
		return executions, fmt.Errorf("batch submission failed with %d errors: %v", len(errors), errors[0])
	}
	
	return executions, nil
}

// BatchConfig defines configuration for batch job submission
type BatchConfig struct {
	MaxConcurrent int  `json:"max_concurrent"`
	FailFast      bool `json:"fail_fast"`
}

type batchResult struct {
	execution *JobExecution
	err       error
}

// WaitForJob waits for a job to complete
func (ko *KubernetesOrchestrator) WaitForJob(ctx context.Context, jobID string, timeout time.Duration) (*JobExecution, error) {
	ko.logger.Info("Waiting for job completion", "job_id", jobID, "timeout", timeout)
	
	job, exists := ko.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	// Create context with timeout
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Watch for job completion
	eventChan, err := ko.client.WatchJob(waitCtx, job.Namespace, job.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to watch job: %w", err)
	}
	
	for {
		select {
		case event := <-eventChan:
			if event.Job.Status == JobStatusCompleted || event.Job.Status == JobStatusFailed {
				// Find and return the job execution
				for i := range ko.jobHistory {
					if ko.jobHistory[i].JobID == jobID {
						return &ko.jobHistory[i], nil
					}
				}
			}
		case <-waitCtx.Done():
			return nil, fmt.Errorf("timeout waiting for job %s", jobID)
		}
	}
}

// GetJobLogs retrieves logs for a test job
func (ko *KubernetesOrchestrator) GetJobLogs(ctx context.Context, jobID string) ([]string, error) {
	job, exists := ko.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	// Get pods for the job
	pods, err := ko.client.GetJobPods(ctx, job.Namespace, job.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get job pods: %w", err)
	}
	
	allLogs := make([]string, 0)
	
	for _, pod := range pods {
		logReader, err := ko.client.GetPodLogs(ctx, pod.Namespace, pod.Name, false)
		if err != nil {
			ko.logger.Error("Failed to get pod logs", "pod", pod.Name, "error", err)
			continue
		}
		defer logReader.Close()
		
		logData, err := io.ReadAll(logReader)
		if err != nil {
			ko.logger.Error("Failed to read pod logs", "pod", pod.Name, "error", err)
			continue
		}
		
		if len(logData) > 0 {
			logs := strings.Split(string(logData), "\n")
			allLogs = append(allLogs, logs...)
		}
	}
	
	return allLogs, nil
}

// CancelJob cancels a running test job
func (ko *KubernetesOrchestrator) CancelJob(ctx context.Context, jobID string) error {
	ko.logger.Info("Cancelling job", "job_id", jobID)
	
	job, exists := ko.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}
	
	// Delete the Kubernetes job
	if err := ko.client.DeleteJob(ctx, job.Namespace, job.Name); err != nil {
		return fmt.Errorf("failed to delete Kubernetes job: %w", err)
	}
	
	// Update job status
	job.Status = JobStatusCancelled
	
	// Update execution history
	for i := range ko.jobHistory {
		if ko.jobHistory[i].JobID == jobID {
			ko.jobHistory[i].Status = JobStatusCancelled
			now := time.Now()
			ko.jobHistory[i].EndTime = &now
			ko.jobHistory[i].Duration = now.Sub(ko.jobHistory[i].StartTime)
			break
		}
	}
	
	return nil
}

// ListJobs lists all test jobs
func (ko *KubernetesOrchestrator) ListJobs(ctx context.Context, filters *JobFilters) ([]*TestJob, error) {
	labels := make(map[string]string)
	if filters != nil && filters.Labels != nil {
		labels = filters.Labels
	}
	
	jobs, err := ko.client.ListJobs(ctx, ko.namespace, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	
	// Apply additional filters
	if filters != nil {
		jobs = ko.applyJobFilters(jobs, filters)
	}
	
	return jobs, nil
}

// JobFilters defines filters for listing jobs
type JobFilters struct {
	Status    JobStatus         `json:"status,omitempty"`
	TestType  TestJobType       `json:"test_type,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	StartTime *time.Time        `json:"start_time,omitempty"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
}

// GetClusterStatus returns the status of the Kubernetes cluster
func (ko *KubernetesOrchestrator) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	clusterInfo, err := ko.client.GetClusterInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	
	nodeResources, err := ko.client.GetNodeResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get node resources: %w", err)
	}
	
	namespaceUsage, err := ko.client.GetResourceUsage(ctx, ko.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace resource usage: %w", err)
	}
	
	return &ClusterStatus{
		ClusterInfo:     clusterInfo,
		NodeResources:   nodeResources,
		NamespaceUsage:  namespaceUsage,
		ActiveJobs:      ko.countActiveJobs(),
		TotalJobs:       len(ko.jobs),
		LastUpdated:     time.Now(),
	}, nil
}

// ClusterStatus represents the overall status of the Kubernetes cluster
type ClusterStatus struct {
	ClusterInfo     *ClusterInfo             `json:"cluster_info"`
	NodeResources   []*NodeResources         `json:"node_resources"`
	NamespaceUsage  *NamespaceResourceUsage  `json:"namespace_usage"`
	ActiveJobs      int                      `json:"active_jobs"`
	TotalJobs       int                      `json:"total_jobs"`
	LastUpdated     time.Time                `json:"last_updated"`
}

// Helper methods

func (ko *KubernetesOrchestrator) applyDefaultsToJob(job *TestJob) {
	if job.Namespace == "" {
		job.Namespace = ko.config.Namespace
	}
	
	if job.ServiceAccount == "" && ko.config.ServiceAccount != "" {
		// Would set service account in job spec
	}
	
	// Apply default resources
	if job.Resources.CPURequest == "" {
		job.Resources.CPURequest = ko.config.DefaultResources.CPURequest
	}
	if job.Resources.CPULimit == "" {
		job.Resources.CPULimit = ko.config.DefaultResources.CPULimit
	}
	if job.Resources.MemoryRequest == "" {
		job.Resources.MemoryRequest = ko.config.DefaultResources.MemoryRequest
	}
	if job.Resources.MemoryLimit == "" {
		job.Resources.MemoryLimit = ko.config.DefaultResources.MemoryLimit
	}
	
	// Apply default node selector
	if len(job.NodeSelector) == 0 && len(ko.config.NodeSelector) > 0 {
		job.NodeSelector = make(map[string]string)
		for k, v := range ko.config.NodeSelector {
			job.NodeSelector[k] = v
		}
	}
	
	// Apply default tolerations
	if len(job.Tolerations) == 0 && len(ko.config.Tolerations) > 0 {
		job.Tolerations = make([]Toleration, len(ko.config.Tolerations))
		copy(job.Tolerations, ko.config.Tolerations)
	}
	
	// Apply default affinity
	if job.Affinity == nil && ko.config.Affinity != nil {
		job.Affinity = ko.config.Affinity
	}
	
	// Apply default security context
	if job.SecurityContext == nil && ko.config.SecurityContext != nil {
		job.SecurityContext = ko.config.SecurityContext
	}
	
	// Set defaults for test job specific fields
	if job.RestartPolicy == "" {
		job.RestartPolicy = "Never"
	}
	
	if job.BackoffLimit == nil {
		backoffLimit := int32(3)
		job.BackoffLimit = &backoffLimit
	}
	
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
}

func (ko *KubernetesOrchestrator) validateJob(job *TestJob) error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}
	
	if job.Image == "" {
		return fmt.Errorf("job image is required")
	}
	
	if job.TestType == "" {
		return fmt.Errorf("job test type is required")
	}
	
	// Validate resource requests/limits
	if err := ko.validateResources(&job.Resources); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}
	
	return nil
}

func (ko *KubernetesOrchestrator) validateResources(resources *ResourceRequests) error {
	// Validate CPU and memory format
	// This would include proper validation of Kubernetes resource format
	return nil
}

func (ko *KubernetesOrchestrator) checkJobDependencies(ctx context.Context, job *TestJob) error {
	for _, depJobID := range job.DependsOn {
		depJob, exists := ko.jobs[depJobID]
		if !exists {
			return fmt.Errorf("dependency job not found: %s", depJobID)
		}
		
		if depJob.Status != JobStatusCompleted {
			return fmt.Errorf("dependency job %s not completed (status: %s)", depJobID, depJob.Status)
		}
	}
	
	return nil
}

func (ko *KubernetesOrchestrator) monitorJob(ctx context.Context, job *TestJob, execution *JobExecution) {
	// Monitor job progress and update execution record
	eventChan, err := ko.client.WatchJob(ctx, job.Namespace, job.Name)
	if err != nil {
		ko.logger.Error("Failed to watch job", "job_id", job.ID, "error", err)
		return
	}
	
	for event := range eventChan {
		// Update job status
		job.Status = event.Job.Status
		
		// Update execution record
		for i := range ko.jobHistory {
			if ko.jobHistory[i].JobID == job.ID {
				ko.jobHistory[i].Status = event.Job.Status
				
				if event.Job.Status == JobStatusRunning && ko.jobHistory[i].StartTime.IsZero() {
					now := time.Now()
					ko.jobHistory[i].StartTime = now
					job.StartedAt = &now
				}
				
				if event.Job.Status == JobStatusCompleted || event.Job.Status == JobStatusFailed {
					now := time.Now()
					ko.jobHistory[i].EndTime = &now
					ko.jobHistory[i].Duration = now.Sub(ko.jobHistory[i].StartTime)
					
					if event.Job.Status == JobStatusCompleted {
						job.CompletedAt = &now
					} else {
						job.FailedAt = &now
					}
				}
				break
			}
		}
		
		if event.Job.Status == JobStatusCompleted || event.Job.Status == JobStatusFailed {
			// Collect final metrics and logs
			ko.collectJobResults(ctx, job)
			break
		}
	}
}

func (ko *KubernetesOrchestrator) collectJobResults(ctx context.Context, job *TestJob) {
	// Get job pods
	pods, err := ko.client.GetJobPods(ctx, job.Namespace, job.Name)
	if err != nil {
		ko.logger.Error("Failed to get job pods for result collection", "job_id", job.ID, "error", err)
		return
	}
	
	// Collect metrics and logs from each pod
	for _, pod := range pods {
		// Get pod metrics
		metrics, err := ko.client.GetPodMetrics(ctx, pod.Namespace, pod.Name)
		if err != nil {
			ko.logger.Error("Failed to get pod metrics", "pod", pod.Name, "error", err)
		}
		
		// Update execution record with metrics
		for i := range ko.jobHistory {
			if ko.jobHistory[i].JobID == job.ID {
				if metrics != nil {
					ko.jobHistory[i].Metrics = metrics
				}
				break
			}
		}
	}
}

func (ko *KubernetesOrchestrator) applyJobFilters(jobs []*TestJob, filters *JobFilters) []*TestJob {
	filtered := make([]*TestJob, 0)
	
	for _, job := range jobs {
		if filters.Status != "" && job.Status != filters.Status {
			continue
		}
		
		if filters.TestType != "" && job.TestType != filters.TestType {
			continue
		}
		
		if filters.StartTime != nil && job.CreatedAt.Before(*filters.StartTime) {
			continue
		}
		
		if filters.EndTime != nil && job.CreatedAt.After(*filters.EndTime) {
			continue
		}
		
		filtered = append(filtered, job)
	}
	
	return filtered
}

func (ko *KubernetesOrchestrator) countActiveJobs() int {
	count := 0
	for _, job := range ko.jobs {
		if job.Status == JobStatusPending || job.Status == JobStatusRunning {
			count++
		}
	}
	return count
}

// MarshalJSON implements custom JSON marshaling for TestJob
func (tj *TestJob) MarshalJSON() ([]byte, error) {
	type Alias TestJob
	return json.Marshal(&struct {
		*Alias
		CreatedAtISO   string `json:"created_at_iso"`
		StartedAtISO   string `json:"started_at_iso,omitempty"`
		CompletedAtISO string `json:"completed_at_iso,omitempty"`
		FailedAtISO    string `json:"failed_at_iso,omitempty"`
	}{
		Alias:          (*Alias)(tj),
		CreatedAtISO:   tj.CreatedAt.Format(time.RFC3339),
		StartedAtISO:   formatTimePtr(tj.StartedAt),
		CompletedAtISO: formatTimePtr(tj.CompletedAt),
		FailedAtISO:    formatTimePtr(tj.FailedAt),
	})
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}