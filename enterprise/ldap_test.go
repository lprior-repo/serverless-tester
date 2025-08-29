// Package enterprise - LDAP/Active Directory integration tests
// Following strict TDD methodology for directory service integration
package enterprise

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
)

// Test Models for LDAP/AD Integration

type LDAPConfiguration struct {
	ServerURL       string            `json:"serverUrl"`
	BaseDN          string            `json:"baseDn"`
	BindDN          string            `json:"bindDn"`
	BindPassword    string            `json:"bindPassword"`
	UserSearchBase  string            `json:"userSearchBase"`
	GroupSearchBase string            `json:"groupSearchBase"`
	UserFilter      string            `json:"userFilter"`
	GroupFilter     string            `json:"groupFilter"`
	Attributes      map[string]string `json:"attributes"`
	TLSConfig       *TLSConfig        `json:"tlsConfig,omitempty"`
	ConnectionPool  *PoolConfig       `json:"connectionPool,omitempty"`
}

type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	CertificateFile    string `json:"certificateFile,omitempty"`
	KeyFile            string `json:"keyFile,omitempty"`
	CAFile             string `json:"caFile,omitempty"`
}

type PoolConfig struct {
	MaxConnections    int           `json:"maxConnections"`
	MaxIdleTime       time.Duration `json:"maxIdleTime"`
	ConnectionTimeout time.Duration `json:"connectionTimeout"`
	ReadTimeout       time.Duration `json:"readTimeout"`
}

type DirectoryUser struct {
	DN             string            `json:"dn"`
	Username       string            `json:"username"`
	Email          string            `json:"email"`
	FirstName      string            `json:"firstName"`
	LastName       string            `json:"lastName"`
	DisplayName    string            `json:"displayName"`
	Department     string            `json:"department"`
	Title          string            `json:"title"`
	Manager        string            `json:"manager"`
	Groups         []string          `json:"groups"`
	Attributes     map[string]string `json:"attributes"`
	AccountEnabled bool              `json:"accountEnabled"`
	LastLogon      *time.Time        `json:"lastLogon,omitempty"`
	PasswordExpiry *time.Time        `json:"passwordExpiry,omitempty"`
}

type DirectoryGroup struct {
	DN          string            `json:"dn"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Members     []string          `json:"members"`
	Attributes  map[string]string `json:"attributes"`
	GroupType   string            `json:"groupType"`
}

type SearchResult struct {
	Users  []DirectoryUser  `json:"users"`
	Groups []DirectoryGroup `json:"groups"`
	Count  int              `json:"count"`
	HasMore bool            `json:"hasMore"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

type DirectoryAuthResult struct {
	Success      bool              `json:"success"`
	User         *DirectoryUser    `json:"user,omitempty"`
	ErrorCode    string            `json:"errorCode,omitempty"`
	ErrorMessage string            `json:"errorMessage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LDAP Client Interface
type LDAPClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Authenticate(ctx context.Context, username, password string) (*DirectoryAuthResult, error)
	SearchUsers(ctx context.Context, filter string, attributes []string) (*SearchResult, error)
	SearchGroups(ctx context.Context, filter string, attributes []string) (*SearchResult, error)
	GetUser(ctx context.Context, username string) (*DirectoryUser, error)
	GetGroup(ctx context.Context, groupName string) (*DirectoryGroup, error)
	GetUserGroups(ctx context.Context, username string) ([]DirectoryGroup, error)
	ValidateCredentials(ctx context.Context, username, password string) error
}

// TDD RED Phase - Write failing tests first

func TestLDAPClientConnection(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	config := LDAPConfiguration{
		ServerURL:       "ldap://ad.enterprise.com:389",
		BaseDN:          "DC=enterprise,DC=com",
		BindDN:          "CN=service-account,OU=Service Accounts,DC=enterprise,DC=com",
		BindPassword:    "SecureServicePassword123",
		UserSearchBase:  "OU=Users,DC=enterprise,DC=com",
		GroupSearchBase: "OU=Groups,DC=enterprise,DC=com",
		UserFilter:      "(&(objectClass=user)(sAMAccountName=%s))",
		GroupFilter:     "(&(objectClass=group)(cn=%s))",
		Attributes: map[string]string{
			"username":    "sAMAccountName",
			"email":       "mail",
			"firstName":   "givenName",
			"lastName":    "sn",
			"displayName": "displayName",
			"department":  "department",
			"title":       "title",
		},
	}

	client, err := NewLDAPClient(config)
	require.NoError(t, err, "Should create LDAP client successfully")

	t.Run("successful_connection", func(t *testing.T) {
		err := client.Connect(context.Background())
		require.NoError(t, err, "Should connect to LDAP server successfully")

		defer func() {
			err := client.Disconnect()
			assert.NoError(t, err, "Should disconnect cleanly")
		}()
	})

	t.Run("connection_with_invalid_credentials", func(t *testing.T) {
		invalidConfig := config
		invalidConfig.BindPassword = "WrongPassword"
		
		invalidClient, err := NewLDAPClient(invalidConfig)
		require.NoError(t, err, "Should create client with invalid config")

		err = invalidClient.Connect(context.Background())
		assert.Error(t, err, "Should fail to connect with invalid credentials")
		assert.Contains(t, err.Error(), "authentication", "Error should indicate authentication failure")
	})
}

func TestLDAPUserAuthentication(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	config := createTestLDAPConfig()
	client, err := NewLDAPClient(config)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)
	defer client.Disconnect()

	t.Run("successful_user_authentication", func(t *testing.T) {
		result, err := client.Authenticate(context.Background(), "john.doe", "UserPassword123")
		require.NoError(t, err, "Authentication should succeed")
		assert.True(t, result.Success, "Authentication result should be successful")
		assert.NotNil(t, result.User, "Should return user information")
		assert.Equal(t, "john.doe", result.User.Username, "Username should match")
		assert.Equal(t, "john.doe@enterprise.com", result.User.Email, "Email should be populated")
		assert.Equal(t, "John", result.User.FirstName, "First name should be populated")
		assert.Equal(t, "Doe", result.User.LastName, "Last name should be populated")
	})

	t.Run("failed_user_authentication", func(t *testing.T) {
		result, err := client.Authenticate(context.Background(), "john.doe", "WrongPassword")
		require.NoError(t, err, "Should handle authentication failure gracefully")
		assert.False(t, result.Success, "Authentication should fail")
		assert.Nil(t, result.User, "Should not return user information on failure")
		assert.NotEmpty(t, result.ErrorCode, "Should provide error code")
		assert.NotEmpty(t, result.ErrorMessage, "Should provide error message")
	})

	t.Run("authentication_with_disabled_account", func(t *testing.T) {
		result, err := client.Authenticate(context.Background(), "disabled.user", "AnyPassword")
		require.NoError(t, err, "Should handle disabled account gracefully")
		assert.False(t, result.Success, "Authentication should fail for disabled account")
		assert.Equal(t, "account_disabled", result.ErrorCode, "Should indicate account is disabled")
	})
}

func TestLDAPUserSearch(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	config := createTestLDAPConfig()
	client, err := NewLDAPClient(config)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)
	defer client.Disconnect()

	t.Run("search_users_by_department", func(t *testing.T) {
		filter := "department=Engineering"
		attributes := []string{"username", "email", "department", "title"}

		result, err := client.SearchUsers(context.Background(), filter, attributes)
		require.NoError(t, err, "User search should succeed")
		assert.Greater(t, result.Count, 0, "Should find users in Engineering department")
		assert.Greater(t, len(result.Users), 0, "Should return user records")

		// Validate returned user data
		for _, user := range result.Users {
			assert.NotEmpty(t, user.Username, "Username should be populated")
			assert.NotEmpty(t, user.Email, "Email should be populated")
			assert.Equal(t, "Engineering", user.Department, "Department should match filter")
		}
	})

	t.Run("search_users_with_pagination", func(t *testing.T) {
		filter := "*" // Search all users
		attributes := []string{"username", "email"}

		result, err := client.SearchUsers(context.Background(), filter, attributes)
		require.NoError(t, err, "User search should succeed")
		assert.Greater(t, result.Count, 0, "Should find users")
		
		if result.HasMore {
			assert.NotEmpty(t, result.NextPageToken, "Should provide pagination token")
		}
	})

	t.Run("get_specific_user", func(t *testing.T) {
		user, err := client.GetUser(context.Background(), "john.doe")
		require.NoError(t, err, "Should get user successfully")
		assert.Equal(t, "john.doe", user.Username, "Username should match")
		assert.Equal(t, "john.doe@enterprise.com", user.Email, "Email should be correct")
		assert.True(t, user.AccountEnabled, "Account should be enabled")
		assert.Greater(t, len(user.Groups), 0, "User should belong to groups")
	})
}

func TestLDAPGroupOperations(t *testing.T) {
	// RED: This test should fail initially  
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	config := createTestLDAPConfig()
	client, err := NewLDAPClient(config)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)
	defer client.Disconnect()

	t.Run("search_groups", func(t *testing.T) {
		filter := "groupType=Security"
		attributes := []string{"name", "description", "members"}

		result, err := client.SearchGroups(context.Background(), filter, attributes)
		require.NoError(t, err, "Group search should succeed")
		assert.Greater(t, result.Count, 0, "Should find security groups")

		for _, group := range result.Groups {
			assert.NotEmpty(t, group.Name, "Group name should be populated")
			assert.Equal(t, "Security", group.GroupType, "Group type should match filter")
		}
	})

	t.Run("get_specific_group", func(t *testing.T) {
		group, err := client.GetGroup(context.Background(), "Developers")
		require.NoError(t, err, "Should get group successfully")
		assert.Equal(t, "Developers", group.Name, "Group name should match")
		assert.Greater(t, len(group.Members), 0, "Group should have members")
		assert.NotEmpty(t, group.Description, "Group should have description")
	})

	t.Run("get_user_groups", func(t *testing.T) {
		groups, err := client.GetUserGroups(context.Background(), "john.doe")
		require.NoError(t, err, "Should get user groups successfully")
		assert.Greater(t, len(groups), 0, "User should belong to groups")

		// Check for expected groups
		groupNames := make([]string, len(groups))
		for i, group := range groups {
			groupNames[i] = group.Name
		}
		assert.Contains(t, groupNames, "Developers", "User should be in Developers group")
		assert.Contains(t, groupNames, "Domain Users", "User should be in Domain Users group")
	})
}

func TestLDAPPooledConnections(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	config := createTestLDAPConfig()
	config.ConnectionPool = &PoolConfig{
		MaxConnections:    10,
		MaxIdleTime:       5 * time.Minute,
		ConnectionTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
	}

	client, err := NewLDAPClient(config)
	require.NoError(t, err, "Should create LDAP client with pool config")

	t.Run("concurrent_operations", func(t *testing.T) {
		err = client.Connect(context.Background())
		require.NoError(t, err)
		defer client.Disconnect()

		// Test concurrent user lookups
		usernames := []string{"john.doe", "jane.smith", "bob.wilson", "alice.johnson"}
		results := make(chan error, len(usernames))

		for _, username := range usernames {
			go func(user string) {
				_, err := client.GetUser(context.Background(), user)
				results <- err
			}(username)
		}

		// Wait for all goroutines to complete
		for i := 0; i < len(usernames); i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent user lookup should succeed")
		}
	})

	t.Run("connection_pool_limits", func(t *testing.T) {
		err = client.Connect(context.Background())
		require.NoError(t, err)
		defer client.Disconnect()

		// Test that pool respects connection limits
		// This would be more complex in a real implementation
		pooledClient, ok := client.(PooledLDAPClient)
		require.True(t, ok, "Client should implement PooledLDAPClient interface")
		
		poolStats := pooledClient.GetConnectionPoolStats()
		assert.LessOrEqual(t, poolStats.ActiveConnections, config.ConnectionPool.MaxConnections,
			"Active connections should not exceed pool limit")
	})
}

func TestActiveDirectorySpecificFeatures(t *testing.T) {
	// RED: This test should fail initially
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	// Active Directory specific configuration
	config := LDAPConfiguration{
		ServerURL:       "ldap://dc.enterprise.com:389",
		BaseDN:          "DC=enterprise,DC=com",
		BindDN:          "service-account@enterprise.com",
		BindPassword:    "SecureServicePassword123",
		UserSearchBase:  "CN=Users,DC=enterprise,DC=com",
		GroupSearchBase: "CN=Builtin,DC=enterprise,DC=com",
		UserFilter:      "(&(objectCategory=person)(objectClass=user)(sAMAccountName=%s))",
		GroupFilter:     "(&(objectClass=group)(cn=%s))",
		Attributes: map[string]string{
			"username":       "sAMAccountName",
			"userPrincipal":  "userPrincipalName",
			"email":          "mail",
			"firstName":      "givenName",
			"lastName":       "sn",
			"displayName":    "displayName",
			"department":     "department",
			"title":          "title",
			"manager":        "manager",
			"lastLogon":      "lastLogon",
			"accountControl": "userAccountControl",
		},
	}

	client, err := NewActiveDirectoryClient(config)
	require.NoError(t, err, "Should create AD client successfully")

	t.Run("ad_authentication_with_upn", func(t *testing.T) {
		err = client.Connect(context.Background())
		require.NoError(t, err)
		defer client.Disconnect()

		// Test authentication with User Principal Name
		result, err := client.Authenticate(context.Background(), "john.doe@enterprise.com", "UserPassword123")
		require.NoError(t, err, "AD authentication with UPN should succeed")
		assert.True(t, result.Success, "Authentication should be successful")
		assert.Equal(t, "john.doe@enterprise.com", result.User.Email, "UPN should match email")
	})

	t.Run("ad_nested_group_membership", func(t *testing.T) {
		err = client.Connect(context.Background())
		require.NoError(t, err)
		defer client.Disconnect()

		// Test nested group resolution
		groups, err := client.GetUserGroups(context.Background(), "john.doe")
		require.NoError(t, err, "Should get user groups including nested")
		
		// Should include both direct and nested groups
		groupNames := make([]string, len(groups))
		for i, group := range groups {
			groupNames[i] = group.Name
		}
		
		assert.Contains(t, groupNames, "Developers", "Should include direct group membership")
		assert.Contains(t, groupNames, "Engineering", "Should include parent group from nesting")
	})

	t.Run("ad_account_status_check", func(t *testing.T) {
		err = client.Connect(context.Background())
		require.NoError(t, err)
		defer client.Disconnect()

		user, err := client.GetUser(context.Background(), "john.doe")
		require.NoError(t, err, "Should get user successfully")

		assert.True(t, user.AccountEnabled, "Account should be enabled")
		assert.NotNil(t, user.LastLogon, "Should have last logon time")
		
		// Test password expiry information
		if user.PasswordExpiry != nil {
			assert.True(t, user.PasswordExpiry.After(time.Now()), "Password should not be expired")
		}
	})
}

// Helper Functions and Mock Data

func createTestLDAPConfig() LDAPConfiguration {
	return LDAPConfiguration{
		ServerURL:       "ldap://test-ad.enterprise.com:389",
		BaseDN:          "DC=enterprise,DC=com",
		BindDN:          "CN=test-service,OU=Service Accounts,DC=enterprise,DC=com",
		BindPassword:    "TestServicePassword123",
		UserSearchBase:  "OU=Users,DC=enterprise,DC=com",
		GroupSearchBase: "OU=Groups,DC=enterprise,DC=com",
		UserFilter:      "(&(objectClass=user)(sAMAccountName=%s))",
		GroupFilter:     "(&(objectClass=group)(cn=%s))",
		Attributes: map[string]string{
			"username":    "sAMAccountName",
			"email":       "mail",
			"firstName":   "givenName",
			"lastName":    "sn",
			"displayName": "displayName",
			"department":  "department",
			"title":       "title",
		},
		TLSConfig: &TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: true,
		},
	}
}

// Interface definitions that need to be implemented
type ConnectionPoolStats struct {
	ActiveConnections int `json:"activeConnections"`
	IdleConnections   int `json:"idleConnections"`
	TotalConnections  int `json:"totalConnections"`
}

// Extended interface for connection pool statistics
type PooledLDAPClient interface {
	LDAPClient
	GetConnectionPoolStats() *ConnectionPoolStats
}

// Mock constructor functions are now implemented in ldap.go