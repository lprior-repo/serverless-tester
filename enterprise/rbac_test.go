// Package enterprise - RBAC (Role-Based Access Control) framework tests
// Following strict TDD methodology for enterprise access control
package enterprise

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
)

// Test Models for RBAC Framework

type Permission struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Conditions  map[string]string `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	ParentRoles []string     `json:"parentRoles,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type Subject struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // user, service, group
	Name       string            `json:"name"`
	Roles      []string          `json:"roles"`
	Groups     []string          `json:"groups,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Tenant     string            `json:"tenant,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

type AccessRequest struct {
	Subject      Subject           `json:"subject"`
	Resource     string            `json:"resource"`
	Action       string            `json:"action"`
	Context      map[string]string `json:"context,omitempty"`
	RequestTime  time.Time         `json:"requestTime"`
	ClientIP     string            `json:"clientIp,omitempty"`
	UserAgent    string            `json:"userAgent,omitempty"`
}

type AccessDecision struct {
	Allowed     bool              `json:"allowed"`
	Reason      string            `json:"reason"`
	PolicyID    string            `json:"policyId,omitempty"`
	Permissions []Permission      `json:"permissions,omitempty"`
	Conditions  []string          `json:"conditions,omitempty"`
	TTL         time.Duration     `json:"ttl,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Policy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []PolicyRule      `json:"rules"`
	Priority    int               `json:"priority"`
	Enabled     bool              `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type PolicyRule struct {
	ID         string            `json:"id"`
	Effect     string            `json:"effect"` // ALLOW or DENY
	Subjects   []string          `json:"subjects"`
	Resources  []string          `json:"resources"`
	Actions    []string          `json:"actions"`
	Conditions map[string]string `json:"conditions,omitempty"`
	Priority   int               `json:"priority"`
}

type AuditLog struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Subject     Subject           `json:"subject"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Decision    string            `json:"decision"` // ALLOW or DENY
	Reason      string            `json:"reason"`
	PolicyID    string            `json:"policyId,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
	ClientIP    string            `json:"clientIp,omitempty"`
	UserAgent   string            `json:"userAgent,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RBAC Engine Interface
type RBACEngine interface {
	// Permission Management
	CreatePermission(ctx context.Context, permission Permission) error
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	ListPermissions(ctx context.Context, resource string) ([]Permission, error)
	DeletePermission(ctx context.Context, permissionID string) error
	
	// Role Management
	CreateRole(ctx context.Context, role Role) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, role Role) error
	DeleteRole(ctx context.Context, roleID string) error
	AddPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	
	// Subject Management
	CreateSubject(ctx context.Context, subject Subject) error
	GetSubject(ctx context.Context, subjectID string) (*Subject, error)
	ListSubjects(ctx context.Context, tenant string) ([]Subject, error)
	UpdateSubject(ctx context.Context, subject Subject) error
	DeleteSubject(ctx context.Context, subjectID string) error
	AssignRole(ctx context.Context, subjectID, roleID string) error
	RevokeRole(ctx context.Context, subjectID, roleID string) error
	
	// Policy Management
	CreatePolicy(ctx context.Context, policy Policy) error
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)
	ListPolicies(ctx context.Context) ([]Policy, error)
	UpdatePolicy(ctx context.Context, policy Policy) error
	DeletePolicy(ctx context.Context, policyID string) error
	
	// Access Control
	CheckAccess(ctx context.Context, request AccessRequest) (*AccessDecision, error)
	BatchCheckAccess(ctx context.Context, requests []AccessRequest) ([]AccessDecision, error)
	GetEffectivePermissions(ctx context.Context, subjectID string) ([]Permission, error)
	
	// Auditing
	GetAuditLogs(ctx context.Context, subjectID string, from, to time.Time) ([]AuditLog, error)
	GetAccessReport(ctx context.Context, resource string, from, to time.Time) (map[string]interface{}, error)
}

// TDD RED Phase - Write failing tests first

func TestRBACPermissionManagement(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err, "Should create RBAC engine successfully")

	t.Run("create_permission", func(t *testing.T) {
		permission := Permission{
			ID:          "perm-read-users",
			Name:        "Read Users",
			Description: "Allows reading user information",
			Resource:    "users",
			Action:      "read",
			Conditions: map[string]string{
				"department": "engineering",
			},
		}

		err := engine.CreatePermission(context.Background(), permission)
		require.NoError(t, err, "Should create permission successfully")

		// Verify permission was created
		retrieved, err := engine.GetPermission(context.Background(), permission.ID)
		require.NoError(t, err, "Should retrieve created permission")
		assert.Equal(t, permission.ID, retrieved.ID, "Permission ID should match")
		assert.Equal(t, permission.Name, retrieved.Name, "Permission name should match")
		assert.Equal(t, permission.Resource, retrieved.Resource, "Resource should match")
		assert.Equal(t, permission.Action, retrieved.Action, "Action should match")
	})

	t.Run("list_permissions_by_resource", func(t *testing.T) {
		// Create multiple permissions for the same resource
		permissions := []Permission{
			{ID: "perm-read-orders", Name: "Read Orders", Resource: "orders", Action: "read"},
			{ID: "perm-write-orders", Name: "Write Orders", Resource: "orders", Action: "write"},
			{ID: "perm-delete-orders", Name: "Delete Orders", Resource: "orders", Action: "delete"},
		}

		for _, perm := range permissions {
			err := engine.CreatePermission(context.Background(), perm)
			require.NoError(t, err, "Should create permission %s", perm.ID)
		}

		// List permissions for the resource
		retrieved, err := engine.ListPermissions(context.Background(), "orders")
		require.NoError(t, err, "Should list permissions successfully")
		assert.Len(t, retrieved, 3, "Should return 3 permissions")

		// Verify all permissions are for the correct resource
		for _, perm := range retrieved {
			assert.Equal(t, "orders", perm.Resource, "All permissions should be for orders resource")
		}
	})

	t.Run("delete_permission", func(t *testing.T) {
		permission := Permission{
			ID:       "perm-temp",
			Name:     "Temporary Permission",
			Resource: "temp",
			Action:   "temp",
		}

		err := engine.CreatePermission(context.Background(), permission)
		require.NoError(t, err, "Should create temporary permission")

		err = engine.DeletePermission(context.Background(), permission.ID)
		require.NoError(t, err, "Should delete permission successfully")

		// Verify permission was deleted
		_, err = engine.GetPermission(context.Background(), permission.ID)
		assert.Error(t, err, "Should not find deleted permission")
	})
}

func TestRBACRoleManagement(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	// Create some permissions first
	readPerm := Permission{ID: "perm-read", Name: "Read", Resource: "data", Action: "read"}
	writePerm := Permission{ID: "perm-write", Name: "Write", Resource: "data", Action: "write"}
	
	err = engine.CreatePermission(context.Background(), readPerm)
	require.NoError(t, err)
	err = engine.CreatePermission(context.Background(), writePerm)
	require.NoError(t, err)

	t.Run("create_role_with_permissions", func(t *testing.T) {
		role := Role{
			ID:          "role-developer",
			Name:        "Developer",
			Description: "Software Developer Role",
			Permissions: []Permission{readPerm, writePerm},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := engine.CreateRole(context.Background(), role)
		require.NoError(t, err, "Should create role successfully")

		// Verify role was created
		retrieved, err := engine.GetRole(context.Background(), role.ID)
		require.NoError(t, err, "Should retrieve created role")
		assert.Equal(t, role.ID, retrieved.ID, "Role ID should match")
		assert.Equal(t, role.Name, retrieved.Name, "Role name should match")
		assert.Len(t, retrieved.Permissions, 2, "Role should have 2 permissions")
	})

	t.Run("add_permission_to_role", func(t *testing.T) {
		// Create a role without permissions
		role := Role{
			ID:          "role-viewer",
			Name:        "Viewer",
			Description: "Read-only access",
			Permissions: []Permission{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := engine.CreateRole(context.Background(), role)
		require.NoError(t, err, "Should create role")

		// Add permission to role
		err = engine.AddPermissionToRole(context.Background(), role.ID, readPerm.ID)
		require.NoError(t, err, "Should add permission to role")

		// Verify permission was added
		retrieved, err := engine.GetRole(context.Background(), role.ID)
		require.NoError(t, err, "Should retrieve updated role")
		assert.Len(t, retrieved.Permissions, 1, "Role should have 1 permission")
		assert.Equal(t, readPerm.ID, retrieved.Permissions[0].ID, "Should have correct permission")
	})

	t.Run("role_hierarchy", func(t *testing.T) {
		// Create parent role
		parentRole := Role{
			ID:          "role-senior-dev",
			Name:        "Senior Developer",
			Description: "Senior Development Role",
			Permissions: []Permission{readPerm, writePerm},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := engine.CreateRole(context.Background(), parentRole)
		require.NoError(t, err, "Should create parent role")

		// Create child role that inherits from parent
		childRole := Role{
			ID:          "role-junior-dev",
			Name:        "Junior Developer", 
			Description: "Junior Development Role",
			Permissions: []Permission{readPerm},
			ParentRoles: []string{parentRole.ID},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err = engine.CreateRole(context.Background(), childRole)
		require.NoError(t, err, "Should create child role")

		// Verify hierarchy
		retrieved, err := engine.GetRole(context.Background(), childRole.ID)
		require.NoError(t, err, "Should retrieve child role")
		assert.Contains(t, retrieved.ParentRoles, parentRole.ID, "Child role should reference parent")
	})
}

func TestRBACSubjectManagement(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	// Create a role first
	role := Role{
		ID:          "role-user",
		Name:        "User",
		Description: "Basic User Role",
		Permissions: []Permission{{ID: "perm-read", Name: "Read", Resource: "profile", Action: "read"}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = engine.CreateRole(context.Background(), role)
	require.NoError(t, err)

	t.Run("create_user_subject", func(t *testing.T) {
		subject := Subject{
			ID:   "user-john-doe",
			Type: "user",
			Name: "John Doe",
			Roles: []string{},
			Attributes: map[string]string{
				"department": "engineering",
				"level":      "senior",
			},
			Tenant:    "enterprise",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := engine.CreateSubject(context.Background(), subject)
		require.NoError(t, err, "Should create subject successfully")

		// Verify subject was created
		retrieved, err := engine.GetSubject(context.Background(), subject.ID)
		require.NoError(t, err, "Should retrieve created subject")
		assert.Equal(t, subject.ID, retrieved.ID, "Subject ID should match")
		assert.Equal(t, subject.Name, retrieved.Name, "Subject name should match")
		assert.Equal(t, subject.Type, retrieved.Type, "Subject type should match")
		assert.Equal(t, subject.Tenant, retrieved.Tenant, "Tenant should match")
	})

	t.Run("assign_role_to_subject", func(t *testing.T) {
		subject := Subject{
			ID:        "user-jane-smith",
			Type:      "user",
			Name:      "Jane Smith",
			Roles:     []string{},
			Tenant:    "enterprise",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := engine.CreateSubject(context.Background(), subject)
		require.NoError(t, err, "Should create subject")

		// Assign role to subject
		err = engine.AssignRole(context.Background(), subject.ID, role.ID)
		require.NoError(t, err, "Should assign role to subject")

		// Verify role was assigned
		retrieved, err := engine.GetSubject(context.Background(), subject.ID)
		require.NoError(t, err, "Should retrieve updated subject")
		assert.Contains(t, retrieved.Roles, role.ID, "Subject should have assigned role")
	})

	t.Run("list_subjects_by_tenant", func(t *testing.T) {
		// Create subjects in different tenants
		subjects := []Subject{
			{ID: "user-tenant1-1", Type: "user", Name: "User 1", Tenant: "tenant1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "user-tenant1-2", Type: "user", Name: "User 2", Tenant: "tenant1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "user-tenant2-1", Type: "user", Name: "User 3", Tenant: "tenant2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		for _, subj := range subjects {
			err := engine.CreateSubject(context.Background(), subj)
			require.NoError(t, err, "Should create subject %s", subj.ID)
		}

		// List subjects for tenant1
		retrieved, err := engine.ListSubjects(context.Background(), "tenant1")
		require.NoError(t, err, "Should list subjects successfully")
		assert.Len(t, retrieved, 2, "Should return 2 subjects for tenant1")

		for _, subj := range retrieved {
			assert.Equal(t, "tenant1", subj.Tenant, "All subjects should belong to tenant1")
		}
	})
}

func TestRBACAccessControl(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	// Setup test data
	setupRBACTestData(t, engine)

	t.Run("check_access_allowed", func(t *testing.T) {
		request := AccessRequest{
			Subject: Subject{
				ID:   "user-developer",
				Type: "user",
				Name: "Developer User",
				Roles: []string{"role-developer"},
				Attributes: map[string]string{
					"department": "engineering",
				},
			},
			Resource:    "users",
			Action:      "read",
			Context: map[string]string{
				"department": "engineering",
			},
			RequestTime: time.Now(),
			ClientIP:    "192.168.1.100",
		}

		decision, err := engine.CheckAccess(context.Background(), request)
		require.NoError(t, err, "Should check access successfully")
		assert.True(t, decision.Allowed, "Access should be allowed")
		assert.NotEmpty(t, decision.Reason, "Should provide reason for decision")
		assert.Greater(t, len(decision.Permissions), 0, "Should include matching permissions")
	})

	t.Run("check_access_denied", func(t *testing.T) {
		request := AccessRequest{
			Subject: Subject{
				ID:   "user-viewer",
				Type: "user",
				Name: "Viewer User",
				Roles: []string{"role-viewer"},
			},
			Resource:    "users",
			Action:      "delete",
			RequestTime: time.Now(),
		}

		decision, err := engine.CheckAccess(context.Background(), request)
		require.NoError(t, err, "Should check access successfully")
		assert.False(t, decision.Allowed, "Access should be denied")
		assert.NotEmpty(t, decision.Reason, "Should provide reason for denial")
	})

	t.Run("batch_access_check", func(t *testing.T) {
		requests := []AccessRequest{
			{
				Subject: Subject{ID: "user-developer", Type: "user", Roles: []string{"role-developer"}},
				Resource: "users", Action: "read", RequestTime: time.Now(),
			},
			{
				Subject: Subject{ID: "user-developer", Type: "user", Roles: []string{"role-developer"}},
				Resource: "users", Action: "write", RequestTime: time.Now(),
			},
			{
				Subject: Subject{ID: "user-developer", Type: "user", Roles: []string{"role-developer"}},
				Resource: "users", Action: "delete", RequestTime: time.Now(),
			},
		}

		decisions, err := engine.BatchCheckAccess(context.Background(), requests)
		require.NoError(t, err, "Should perform batch access check")
		assert.Len(t, decisions, len(requests), "Should return decision for each request")

		// Developer should have read and write but not delete
		assert.True(t, decisions[0].Allowed, "Read should be allowed")
		assert.True(t, decisions[1].Allowed, "Write should be allowed")
		assert.False(t, decisions[2].Allowed, "Delete should be denied")
	})

	t.Run("get_effective_permissions", func(t *testing.T) {
		permissions, err := engine.GetEffectivePermissions(context.Background(), "user-developer")
		require.NoError(t, err, "Should get effective permissions")
		assert.Greater(t, len(permissions), 0, "Should have effective permissions")

		// Verify specific permissions exist
		hasReadPermission := false
		hasWritePermission := false
		
		for _, perm := range permissions {
			if perm.Action == "read" && perm.Resource == "users" {
				hasReadPermission = true
			}
			if perm.Action == "write" && perm.Resource == "users" {
				hasWritePermission = true
			}
		}

		assert.True(t, hasReadPermission, "Should have read permission")
		assert.True(t, hasWritePermission, "Should have write permission")
	})
}

func TestRBACPolicyManagement(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	t.Run("create_policy", func(t *testing.T) {
		policy := Policy{
			ID:          "policy-developer-access",
			Name:        "Developer Access Policy",
			Description: "Defines access rules for developers",
			Rules: []PolicyRule{
				{
					ID:        "rule-dev-read",
					Effect:    "ALLOW",
					Subjects:  []string{"role-developer"},
					Resources: []string{"users", "projects"},
					Actions:   []string{"read", "write"},
					Priority:  100,
				},
				{
					ID:        "rule-dev-deny-delete",
					Effect:    "DENY",
					Subjects:  []string{"role-developer"},
					Resources: []string{"users"},
					Actions:   []string{"delete"},
					Priority:  200, // Higher priority = more specific
				},
			},
			Priority:  10,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := engine.CreatePolicy(context.Background(), policy)
		require.NoError(t, err, "Should create policy successfully")

		// Verify policy was created
		retrieved, err := engine.GetPolicy(context.Background(), policy.ID)
		require.NoError(t, err, "Should retrieve created policy")
		assert.Equal(t, policy.ID, retrieved.ID, "Policy ID should match")
		assert.Equal(t, policy.Name, retrieved.Name, "Policy name should match")
		assert.Len(t, retrieved.Rules, 2, "Policy should have 2 rules")
		assert.True(t, retrieved.Enabled, "Policy should be enabled")
	})

	t.Run("policy_rule_priority", func(t *testing.T) {
		// Test that higher priority rules override lower priority ones
		policy := Policy{
			ID:   "policy-priority-test",
			Name: "Priority Test Policy",
			Rules: []PolicyRule{
				{
					ID:       "rule-low-priority",
					Effect:   "ALLOW",
					Subjects: []string{"role-test"},
					Resources: []string{"test-resource"},
					Actions:  []string{"test-action"},
					Priority: 100,
				},
				{
					ID:       "rule-high-priority",
					Effect:   "DENY",
					Subjects: []string{"role-test"},
					Resources: []string{"test-resource"},
					Actions:  []string{"test-action"},
					Priority: 200,
				},
			},
			Priority:  10,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := engine.CreatePolicy(context.Background(), policy)
		require.NoError(t, err, "Should create priority test policy")

		retrieved, err := engine.GetPolicy(context.Background(), policy.ID)
		require.NoError(t, err, "Should retrieve policy")

		// Verify rules are ordered by priority (highest first)
		assert.Greater(t, retrieved.Rules[1].Priority, retrieved.Rules[0].Priority,
			"Rules should be ordered by priority")
	})
}

func TestRBACMultiTenantIsolation(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	t.Run("tenant_isolation", func(t *testing.T) {
		// Create subjects in different tenants
		tenant1Subject := Subject{
			ID: "user-tenant1", Type: "user", Name: "Tenant 1 User",
			Tenant: "tenant1", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		tenant2Subject := Subject{
			ID: "user-tenant2", Type: "user", Name: "Tenant 2 User",
			Tenant: "tenant2", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}

		err := engine.CreateSubject(context.Background(), tenant1Subject)
		require.NoError(t, err)
		err = engine.CreateSubject(context.Background(), tenant2Subject)
		require.NoError(t, err)

		// Verify tenant isolation
		tenant1Subjects, err := engine.ListSubjects(context.Background(), "tenant1")
		require.NoError(t, err)
		assert.Len(t, tenant1Subjects, 1, "Tenant1 should have 1 subject")
		assert.Equal(t, "user-tenant1", tenant1Subjects[0].ID)

		tenant2Subjects, err := engine.ListSubjects(context.Background(), "tenant2")
		require.NoError(t, err)
		assert.Len(t, tenant2Subjects, 1, "Tenant2 should have 1 subject")
		assert.Equal(t, "user-tenant2", tenant2Subjects[0].ID)
	})

	t.Run("cross_tenant_access_denied", func(t *testing.T) {
		// Test that users in one tenant cannot access resources in another tenant
		request := AccessRequest{
			Subject: Subject{
				ID: "user-tenant1", Type: "user", Tenant: "tenant1",
				Roles: []string{"role-admin"},
			},
			Resource: "tenant2/data",
			Action:   "read",
			Context: map[string]string{
				"target_tenant": "tenant2",
			},
			RequestTime: time.Now(),
		}

		decision, err := engine.CheckAccess(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, decision.Allowed, "Cross-tenant access should be denied")
		assert.Contains(t, decision.Reason, "tenant", "Reason should mention tenant isolation")
	})
}

func TestRBACContextualAccess(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	engine, err := NewRBACEngine()
	require.NoError(t, err)

	t.Run("time_based_access", func(t *testing.T) {
		// Create a permission with time-based conditions
		permission := Permission{
			ID:       "perm-business-hours",
			Name:     "Business Hours Access",
			Resource: "sensitive-data",
			Action:   "read",
			Conditions: map[string]string{
				"time_start": "09:00",
				"time_end":   "17:00",
				"weekdays":   "true",
			},
		}

		err := engine.CreatePermission(context.Background(), permission)
		require.NoError(t, err)

		// Test access during business hours
		request := AccessRequest{
			Subject: Subject{ID: "user-test", Type: "user"},
			Resource: "sensitive-data",
			Action:   "read",
			Context: map[string]string{
				"current_time": "14:00",
				"current_day":  "tuesday",
			},
			RequestTime: time.Now(),
		}

		decision, err := engine.CheckAccess(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, decision.Allowed, "Access should be allowed during business hours")

		// Test access outside business hours
		request.Context["current_time"] = "20:00"
		decision, err = engine.CheckAccess(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, decision.Allowed, "Access should be denied outside business hours")
	})

	t.Run("location_based_access", func(t *testing.T) {
		request := AccessRequest{
			Subject: Subject{ID: "user-remote", Type: "user"},
			Resource: "internal-systems",
			Action:   "access",
			Context: map[string]string{
				"client_location": "external",
				"vpn_connected":   "false",
			},
			ClientIP:    "203.0.113.1", // External IP
			RequestTime: time.Now(),
		}

		decision, err := engine.CheckAccess(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, decision.Allowed, "External access should be denied")
		assert.Contains(t, decision.Reason, "location", "Reason should mention location restriction")
	})
}

// Helper Functions

func setupRBACTestData(t *testing.T, engine RBACEngine) {
	t.Helper()

	// Create permissions
	permissions := []Permission{
		{ID: "perm-users-read", Name: "Read Users", Resource: "users", Action: "read"},
		{ID: "perm-users-write", Name: "Write Users", Resource: "users", Action: "write"},
		{ID: "perm-users-delete", Name: "Delete Users", Resource: "users", Action: "delete"},
	}

	for _, perm := range permissions {
		err := engine.CreatePermission(context.Background(), perm)
		require.NoError(t, err, "Should create permission %s", perm.ID)
	}

	// Create roles
	roles := []Role{
		{
			ID: "role-developer", Name: "Developer", Description: "Developer Role",
			Permissions: []Permission{permissions[0], permissions[1]}, // read, write
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: "role-viewer", Name: "Viewer", Description: "Viewer Role",
			Permissions: []Permission{permissions[0]}, // read only
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	for _, role := range roles {
		err := engine.CreateRole(context.Background(), role)
		require.NoError(t, err, "Should create role %s", role.ID)
	}

	// Create subjects
	subjects := []Subject{
		{
			ID: "user-developer", Type: "user", Name: "Developer User",
			Roles: []string{"role-developer"}, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: "user-viewer", Type: "user", Name: "Viewer User",
			Roles: []string{"role-viewer"}, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	for _, subject := range subjects {
		err := engine.CreateSubject(context.Background(), subject)
		require.NoError(t, err, "Should create subject %s", subject.ID)
	}
}

// Mock constructor function (should fail in RED phase)
func NewRBACEngine() (RBACEngine, error) {
	return nil, fmt.Errorf("RBAC engine not implemented yet")
}