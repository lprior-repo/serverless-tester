// Package orchestration provides cloud-native orchestration capabilities
// Multi-cloud abstraction layer for AWS, Azure, and GCP services
package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CloudProvider defines the supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
	ProviderGCP   CloudProvider = "gcp"
)

// ResourceType defines common resource types across cloud providers
type ResourceType string

const (
	ResourceTypeFunction    ResourceType = "function"
	ResourceTypeDatabase    ResourceType = "database"
	ResourceTypeQueue       ResourceType = "queue"
	ResourceTypeStorage     ResourceType = "storage"
	ResourceTypePubSub      ResourceType = "pubsub"
	ResourceTypeOrchestrator ResourceType = "orchestrator"
	ResourceTypeMonitoring  ResourceType = "monitoring"
	ResourceTypeNetwork     ResourceType = "network"
	ResourceTypeCompute     ResourceType = "compute"
)

// CloudResource represents a generic cloud resource
type CloudResource struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     CloudProvider          `json:"provider"`
	Type         ResourceType           `json:"type"`
	Region       string                 `json:"region"`
	Tags         map[string]string      `json:"tags"`
	Properties   map[string]interface{} `json:"properties"`
	Status       ResourceStatus         `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ResourceStatus represents the status of a cloud resource
type ResourceStatus string

const (
	StatusPending    ResourceStatus = "pending"
	StatusCreating   ResourceStatus = "creating"
	StatusActive     ResourceStatus = "active"
	StatusUpdating   ResourceStatus = "updating"
	StatusDeleting   ResourceStatus = "deleting"
	StatusDeleted    ResourceStatus = "deleted"
	StatusFailed     ResourceStatus = "failed"
	StatusUnknown    ResourceStatus = "unknown"
)

// CloudServiceProvider defines the interface for cloud-specific operations
type CloudServiceProvider interface {
	// Resource lifecycle operations
	CreateResource(ctx context.Context, spec *ResourceSpec) (*CloudResource, error)
	GetResource(ctx context.Context, resourceID string) (*CloudResource, error)
	UpdateResource(ctx context.Context, resourceID string, spec *ResourceSpec) (*CloudResource, error)
	DeleteResource(ctx context.Context, resourceID string) error
	ListResources(ctx context.Context, filters *ResourceFilters) ([]*CloudResource, error)
	
	// Resource status and monitoring
	GetResourceStatus(ctx context.Context, resourceID string) (ResourceStatus, error)
	WaitForResourceReady(ctx context.Context, resourceID string, timeout time.Duration) error
	
	// Provider-specific operations
	GetProviderInfo() *ProviderInfo
	ValidateResourceSpec(spec *ResourceSpec) error
	
	// Cost and optimization
	GetResourceCosts(ctx context.Context, resourceID string, period time.Duration) (*CostInfo, error)
	GetOptimizationRecommendations(ctx context.Context, resourceID string) ([]*OptimizationRecommendation, error)
}

// ResourceSpec defines the specification for creating/updating resources
type ResourceSpec struct {
	Name         string                 `json:"name"`
	Type         ResourceType           `json:"type"`
	Region       string                 `json:"region"`
	Tags         map[string]string      `json:"tags"`
	Properties   map[string]interface{} `json:"properties"`
	Dependencies []string               `json:"dependencies"`
}

// ResourceFilters defines filters for listing resources
type ResourceFilters struct {
	Provider CloudProvider     `json:"provider,omitempty"`
	Type     ResourceType      `json:"type,omitempty"`
	Region   string            `json:"region,omitempty"`
	Status   ResourceStatus    `json:"status,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// ProviderInfo contains information about a cloud provider
type ProviderInfo struct {
	Name            string            `json:"name"`
	Provider        CloudProvider     `json:"provider"`
	SupportedTypes  []ResourceType    `json:"supported_types"`
	SupportedRegions []string         `json:"supported_regions"`
	Features        map[string]bool   `json:"features"`
	Limits          map[string]int    `json:"limits"`
}

// CostInfo represents cost information for a resource
type CostInfo struct {
	ResourceID    string    `json:"resource_id"`
	Currency      string    `json:"currency"`
	TotalCost     float64   `json:"total_cost"`
	DailyCost     float64   `json:"daily_cost"`
	Period        string    `json:"period"`
	Breakdown     []CostItem `json:"breakdown"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CostItem represents individual cost components
type CostItem struct {
	Service     string  `json:"service"`
	Component   string  `json:"component"`
	Amount      float64 `json:"amount"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TotalCost   float64 `json:"total_cost"`
}

// OptimizationRecommendation provides recommendations for resource optimization
type OptimizationRecommendation struct {
	Type            string                 `json:"type"`
	Priority        string                 `json:"priority"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	EstimatedSaving float64                `json:"estimated_saving"`
	Actions         []string               `json:"actions"`
	Properties      map[string]interface{} `json:"properties"`
}

// MultiCloudManager manages resources across multiple cloud providers
type MultiCloudManager struct {
	providers map[CloudProvider]CloudServiceProvider
	registry  *ResourceRegistry
	logger    Logger
}

// ResourceRegistry maintains a registry of all managed resources
type ResourceRegistry struct {
	resources map[string]*CloudResource
	indexes   map[string]map[string][]string // provider -> type -> resource_ids
}

// Logger interface for cloud operations logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewMultiCloudManager creates a new multi-cloud manager
func NewMultiCloudManager(logger Logger) *MultiCloudManager {
	return &MultiCloudManager{
		providers: make(map[CloudProvider]CloudServiceProvider),
		registry:  newResourceRegistry(),
		logger:    logger,
	}
}

// RegisterProvider registers a cloud service provider
func (mcm *MultiCloudManager) RegisterProvider(provider CloudProvider, svc CloudServiceProvider) error {
	if svc == nil {
		return fmt.Errorf("provider service cannot be nil")
	}
	
	mcm.providers[provider] = svc
	mcm.logger.Info("Cloud provider registered", "provider", provider)
	return nil
}

// GetProvider returns a registered cloud service provider
func (mcm *MultiCloudManager) GetProvider(provider CloudProvider) (CloudServiceProvider, error) {
	svc, exists := mcm.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not registered", provider)
	}
	return svc, nil
}

// ListProviders returns all registered providers
func (mcm *MultiCloudManager) ListProviders() []CloudProvider {
	providers := make([]CloudProvider, 0, len(mcm.providers))
	for provider := range mcm.providers {
		providers = append(providers, provider)
	}
	return providers
}

// CreateResource creates a resource on the specified provider
func (mcm *MultiCloudManager) CreateResource(ctx context.Context, provider CloudProvider, spec *ResourceSpec) (*CloudResource, error) {
	svc, err := mcm.GetProvider(provider)
	if err != nil {
		return nil, err
	}
	
	// Validate resource specification
	if err := svc.ValidateResourceSpec(spec); err != nil {
		return nil, fmt.Errorf("resource spec validation failed: %w", err)
	}
	
	mcm.logger.Info("Creating resource", "provider", provider, "name", spec.Name, "type", spec.Type)
	
	// Create the resource
	resource, err := svc.CreateResource(ctx, spec)
	if err != nil {
		mcm.logger.Error("Resource creation failed", "provider", provider, "name", spec.Name, "error", err)
		return nil, err
	}
	
	// Register in local registry
	mcm.registry.addResource(resource)
	
	mcm.logger.Info("Resource created successfully", "provider", provider, "resource_id", resource.ID)
	return resource, nil
}

// GetResource retrieves a resource by ID and provider
func (mcm *MultiCloudManager) GetResource(ctx context.Context, provider CloudProvider, resourceID string) (*CloudResource, error) {
	svc, err := mcm.GetProvider(provider)
	if err != nil {
		return nil, err
	}
	
	resource, err := svc.GetResource(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	
	// Update local registry
	mcm.registry.addResource(resource)
	
	return resource, nil
}

// DeleteResource deletes a resource by ID and provider
func (mcm *MultiCloudManager) DeleteResource(ctx context.Context, provider CloudProvider, resourceID string) error {
	svc, err := mcm.GetProvider(provider)
	if err != nil {
		return err
	}
	
	mcm.logger.Info("Deleting resource", "provider", provider, "resource_id", resourceID)
	
	if err := svc.DeleteResource(ctx, resourceID); err != nil {
		mcm.logger.Error("Resource deletion failed", "provider", provider, "resource_id", resourceID, "error", err)
		return err
	}
	
	// Remove from local registry
	mcm.registry.removeResource(resourceID)
	
	mcm.logger.Info("Resource deleted successfully", "provider", provider, "resource_id", resourceID)
	return nil
}

// ListAllResources lists all resources across all providers
func (mcm *MultiCloudManager) ListAllResources(ctx context.Context, filters *ResourceFilters) ([]*CloudResource, error) {
	allResources := make([]*CloudResource, 0)
	
	for provider, svc := range mcm.providers {
		// Apply provider filter if specified
		if filters != nil && filters.Provider != "" && filters.Provider != provider {
			continue
		}
		
		resources, err := svc.ListResources(ctx, filters)
		if err != nil {
			mcm.logger.Error("Failed to list resources", "provider", provider, "error", err)
			continue // Continue with other providers
		}
		
		allResources = append(allResources, resources...)
	}
	
	return allResources, nil
}

// GetMultiCloudCosts retrieves cost information across all providers
func (mcm *MultiCloudManager) GetMultiCloudCosts(ctx context.Context, period time.Duration) (*MultiCloudCostInfo, error) {
	costInfo := &MultiCloudCostInfo{
		Period:        period.String(),
		ProviderCosts: make(map[CloudProvider]*ProviderCostInfo),
		Currency:      "USD", // Default currency
		UpdatedAt:     time.Now(),
	}
	
	var totalCost float64
	
	for provider, svc := range mcm.providers {
		providerCost := &ProviderCostInfo{
			Provider:  provider,
			Resources: make(map[string]*CostInfo),
		}
		
		// Get all resources for this provider
		resources, err := svc.ListResources(ctx, nil)
		if err != nil {
			mcm.logger.Error("Failed to list resources for cost calculation", "provider", provider, "error", err)
			continue
		}
		
		var providerTotalCost float64
		
		for _, resource := range resources {
			costs, err := svc.GetResourceCosts(ctx, resource.ID, period)
			if err != nil {
				mcm.logger.Error("Failed to get resource costs", "provider", provider, "resource_id", resource.ID, "error", err)
				continue
			}
			
			providerCost.Resources[resource.ID] = costs
			providerTotalCost += costs.TotalCost
		}
		
		providerCost.TotalCost = providerTotalCost
		costInfo.ProviderCosts[provider] = providerCost
		totalCost += providerTotalCost
	}
	
	costInfo.TotalCost = totalCost
	return costInfo, nil
}

// GetOptimizationRecommendations gets optimization recommendations across all providers
func (mcm *MultiCloudManager) GetOptimizationRecommendations(ctx context.Context) ([]*MultiCloudOptimizationRecommendation, error) {
	recommendations := make([]*MultiCloudOptimizationRecommendation, 0)
	
	for provider, svc := range mcm.providers {
		// Get all resources for this provider
		resources, err := svc.ListResources(ctx, nil)
		if err != nil {
			mcm.logger.Error("Failed to list resources for optimization", "provider", provider, "error", err)
			continue
		}
		
		for _, resource := range resources {
			recs, err := svc.GetOptimizationRecommendations(ctx, resource.ID)
			if err != nil {
				mcm.logger.Error("Failed to get optimization recommendations", "provider", provider, "resource_id", resource.ID, "error", err)
				continue
			}
			
			for _, rec := range recs {
				multiCloudRec := &MultiCloudOptimizationRecommendation{
					Provider:     provider,
					ResourceID:   resource.ID,
					ResourceName: resource.Name,
					Recommendation: rec,
				}
				recommendations = append(recommendations, multiCloudRec)
			}
		}
	}
	
	return recommendations, nil
}

// MultiCloudCostInfo represents cost information across all providers
type MultiCloudCostInfo struct {
	TotalCost     float64                            `json:"total_cost"`
	Currency      string                             `json:"currency"`
	Period        string                             `json:"period"`
	ProviderCosts map[CloudProvider]*ProviderCostInfo `json:"provider_costs"`
	UpdatedAt     time.Time                          `json:"updated_at"`
}

// ProviderCostInfo represents cost information for a specific provider
type ProviderCostInfo struct {
	Provider  CloudProvider             `json:"provider"`
	TotalCost float64                   `json:"total_cost"`
	Resources map[string]*CostInfo      `json:"resources"`
}

// MultiCloudOptimizationRecommendation represents optimization recommendations across providers
type MultiCloudOptimizationRecommendation struct {
	Provider       CloudProvider                `json:"provider"`
	ResourceID     string                       `json:"resource_id"`
	ResourceName   string                       `json:"resource_name"`
	Recommendation *OptimizationRecommendation  `json:"recommendation"`
}

// ResourceRegistry implementation

func newResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		resources: make(map[string]*CloudResource),
		indexes:   make(map[string]map[string][]string),
	}
}

func (rr *ResourceRegistry) addResource(resource *CloudResource) {
	rr.resources[resource.ID] = resource
	
	// Update indexes
	providerKey := string(resource.Provider)
	typeKey := string(resource.Type)
	
	if rr.indexes[providerKey] == nil {
		rr.indexes[providerKey] = make(map[string][]string)
	}
	if rr.indexes[providerKey][typeKey] == nil {
		rr.indexes[providerKey][typeKey] = make([]string, 0)
	}
	
	// Add resource ID to index if not already present
	found := false
	for _, id := range rr.indexes[providerKey][typeKey] {
		if id == resource.ID {
			found = true
			break
		}
	}
	if !found {
		rr.indexes[providerKey][typeKey] = append(rr.indexes[providerKey][typeKey], resource.ID)
	}
}

func (rr *ResourceRegistry) removeResource(resourceID string) {
	resource, exists := rr.resources[resourceID]
	if !exists {
		return
	}
	
	// Remove from main registry
	delete(rr.resources, resourceID)
	
	// Remove from indexes
	providerKey := string(resource.Provider)
	typeKey := string(resource.Type)
	
	if rr.indexes[providerKey] != nil && rr.indexes[providerKey][typeKey] != nil {
		for i, id := range rr.indexes[providerKey][typeKey] {
			if id == resourceID {
				// Remove from slice
				rr.indexes[providerKey][typeKey] = append(
					rr.indexes[providerKey][typeKey][:i],
					rr.indexes[providerKey][typeKey][i+1:]...,
				)
				break
			}
		}
	}
}

// GetAllResources returns all resources from the registry
func (rr *ResourceRegistry) GetAllResources() []*CloudResource {
	resources := make([]*CloudResource, 0, len(rr.resources))
	for _, resource := range rr.resources {
		resources = append(resources, resource)
	}
	return resources
}

// GetResourcesByProvider returns resources filtered by provider
func (rr *ResourceRegistry) GetResourcesByProvider(provider CloudProvider) []*CloudResource {
	resources := make([]*CloudResource, 0)
	
	providerKey := string(provider)
	if typeMap, exists := rr.indexes[providerKey]; exists {
		for _, resourceIDs := range typeMap {
			for _, resourceID := range resourceIDs {
				if resource, exists := rr.resources[resourceID]; exists {
					resources = append(resources, resource)
				}
			}
		}
	}
	
	return resources
}

// GetResourcesByType returns resources filtered by type across all providers
func (rr *ResourceRegistry) GetResourcesByType(resourceType ResourceType) []*CloudResource {
	resources := make([]*CloudResource, 0)
	
	for _, resource := range rr.resources {
		if resource.Type == resourceType {
			resources = append(resources, resource)
		}
	}
	
	return resources
}

// MarshalJSON implements custom JSON marshaling for CloudResource
func (cr *CloudResource) MarshalJSON() ([]byte, error) {
	type Alias CloudResource
	return json.Marshal(&struct {
		*Alias
		CreatedAtISO string `json:"created_at_iso"`
		UpdatedAtISO string `json:"updated_at_iso"`
	}{
		Alias:        (*Alias)(cr),
		CreatedAtISO: cr.CreatedAt.Format(time.RFC3339),
		UpdatedAtISO: cr.UpdatedAt.Format(time.RFC3339),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for CloudResource
func (cr *CloudResource) UnmarshalJSON(data []byte) error {
	type Alias CloudResource
	aux := &struct {
		*Alias
		CreatedAtISO string `json:"created_at_iso"`
		UpdatedAtISO string `json:"updated_at_iso"`
	}{
		Alias: (*Alias)(cr),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// Parse timestamps if provided
	if aux.CreatedAtISO != "" {
		if t, err := time.Parse(time.RFC3339, aux.CreatedAtISO); err == nil {
			cr.CreatedAt = t
		}
	}
	if aux.UpdatedAtISO != "" {
		if t, err := time.Parse(time.RFC3339, aux.UpdatedAtISO); err == nil {
			cr.UpdatedAt = t
		}
	}
	
	return nil
}