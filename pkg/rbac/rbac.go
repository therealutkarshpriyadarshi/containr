package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Manager manages RBAC for containr
type Manager struct {
	config    *Config
	users     map[string]*User
	roles     map[string]*Role
	auditor   *Auditor
	mu        sync.RWMutex
	storePath string
}

// Config holds RBAC configuration
type Config struct {
	Enabled       bool              `yaml:"enabled"`
	DefaultRole   string            `yaml:"default_role"`
	Roles         map[string]*Role  `yaml:"roles"`
	DefaultQuotas map[string]*ResourceQuota `yaml:"quotas"`
	Authentication AuthConfig      `yaml:"authentication"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Providers []AuthProvider `yaml:"providers"`
}

// AuthProvider represents an authentication provider
type AuthProvider struct {
	Type     string            `yaml:"type"`
	Enabled  bool              `yaml:"enabled"`
	Config   map[string]string `yaml:"config"`
}

// NewManager creates a new RBAC manager
func NewManager(configPath string, storePath string) (*Manager, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	auditor, err := NewAuditor(filepath.Join(storePath, "audit.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create auditor: %w", err)
	}

	mgr := &Manager{
		config:    config,
		users:     make(map[string]*User),
		roles:     config.Roles,
		auditor:   auditor,
		storePath: storePath,
	}

	// Load existing users
	if err := mgr.loadUsers(); err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	return mgr, nil
}

// CreateUser creates a new user
func (m *Manager) CreateUser(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Username]; exists {
		return fmt.Errorf("user %s already exists", user.Username)
	}

	// Apply default role if no roles specified
	if len(user.Roles) == 0 {
		if defaultRole, ok := m.roles[m.config.DefaultRole]; ok {
			user.Roles = []*Role{defaultRole}
		}
	}

	// Apply default quota if not specified
	if user.Quota == nil {
		if defaultQuota, ok := m.config.DefaultQuotas["default"]; ok {
			user.Quota = defaultQuota
		}
	}

	m.users[user.Username] = user

	// Audit log
	m.auditor.Log(ctx, AuditEvent{
		User:     "admin",
		Action:   "user.create",
		Resource: user.Username,
		Result:   "success",
	})

	return m.saveUsers()
}

// GetUser retrieves a user by username
func (m *Manager) GetUser(username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[username]
	if !ok {
		return nil, fmt.Errorf("user %s not found", username)
	}

	return user, nil
}

// DeleteUser deletes a user
func (m *Manager) DeleteUser(ctx context.Context, username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[username]; !exists {
		return fmt.Errorf("user %s not found", username)
	}

	delete(m.users, username)

	// Audit log
	m.auditor.Log(ctx, AuditEvent{
		User:     "admin",
		Action:   "user.delete",
		Resource: username,
		Result:   "success",
	})

	return m.saveUsers()
}

// ListUsers returns all users
func (m *Manager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	return users
}

// GrantRole grants a role to a user
func (m *Manager) GrantRole(ctx context.Context, username string, roleName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[username]
	if !ok {
		return fmt.Errorf("user %s not found", username)
	}

	role, ok := m.roles[roleName]
	if !ok {
		return fmt.Errorf("role %s not found", roleName)
	}

	// Check if user already has role
	for _, r := range user.Roles {
		if r.Name == roleName {
			return fmt.Errorf("user %s already has role %s", username, roleName)
		}
	}

	user.Roles = append(user.Roles, role)

	// Audit log
	m.auditor.Log(ctx, AuditEvent{
		User:     "admin",
		Action:   "role.grant",
		Resource: fmt.Sprintf("%s:%s", username, roleName),
		Result:   "success",
	})

	return m.saveUsers()
}

// RevokeRole revokes a role from a user
func (m *Manager) RevokeRole(ctx context.Context, username string, roleName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[username]
	if !ok {
		return fmt.Errorf("user %s not found", username)
	}

	newRoles := make([]*Role, 0)
	found := false

	for _, r := range user.Roles {
		if r.Name != roleName {
			newRoles = append(newRoles, r)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user %s does not have role %s", username, roleName)
	}

	user.Roles = newRoles

	// Audit log
	m.auditor.Log(ctx, AuditEvent{
		User:     "admin",
		Action:   "role.revoke",
		Resource: fmt.Sprintf("%s:%s", username, roleName),
		Result:   "success",
	})

	return m.saveUsers()
}

// CheckPermission checks if a user has permission for an action
func (m *Manager) CheckPermission(ctx context.Context, username string, resource string, action Action) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[username]
	if !ok {
		return false, fmt.Errorf("user %s not found", username)
	}

	// Check all user roles
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			// Check wildcard resource
			if perm.Resource == "*" || perm.Resource == resource {
				// Check wildcard action
				for _, a := range perm.Actions {
					if a == "*" || a == action {
						// Audit log
						m.auditor.Log(ctx, AuditEvent{
							User:     username,
							Action:   fmt.Sprintf("permission.check.%s.%s", resource, action),
							Resource: resource,
							Result:   "allowed",
						})
						return true, nil
					}
				}
			}
		}
	}

	// Audit log - denied
	m.auditor.Log(ctx, AuditEvent{
		User:     username,
		Action:   fmt.Sprintf("permission.check.%s.%s", resource, action),
		Resource: resource,
		Result:   "denied",
	})

	return false, nil
}

// EnforceQuota checks if user is within quota limits
func (m *Manager) EnforceQuota(ctx context.Context, username string, usage *ResourceUsage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[username]
	if !ok {
		return fmt.Errorf("user %s not found", username)
	}

	if user.Quota == nil {
		return nil // No quota enforcement
	}

	quota := user.Quota

	// Check container count
	if usage.Containers >= quota.MaxContainers {
		return fmt.Errorf("container quota exceeded: %d/%d", usage.Containers, quota.MaxContainers)
	}

	// Check CPU
	if usage.CPU > quota.MaxCPU {
		return fmt.Errorf("CPU quota exceeded: %.2f/%.2f", usage.CPU, quota.MaxCPU)
	}

	// Check memory
	if usage.Memory > quota.MaxMemory {
		return fmt.Errorf("memory quota exceeded: %d/%d", usage.Memory, quota.MaxMemory)
	}

	// Check storage
	if usage.Storage > quota.MaxStorage {
		return fmt.Errorf("storage quota exceeded: %d/%d", usage.Storage, quota.MaxStorage)
	}

	return nil
}

// SetQuota sets resource quota for a user
func (m *Manager) SetQuota(ctx context.Context, username string, quota *ResourceQuota) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[username]
	if !ok {
		return fmt.Errorf("user %s not found", username)
	}

	user.Quota = quota

	// Audit log
	m.auditor.Log(ctx, AuditEvent{
		User:     "admin",
		Action:   "quota.set",
		Resource: username,
		Result:   "success",
	})

	return m.saveUsers()
}

// Close closes the RBAC manager
func (m *Manager) Close() error {
	return m.auditor.Close()
}

// loadConfig loads RBAC configuration from file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// defaultConfig returns default RBAC configuration
func defaultConfig() *Config {
	return &Config{
		Enabled:     true,
		DefaultRole: "viewer",
		Roles: map[string]*Role{
			"admin": {
				Name: "admin",
				Permissions: []*Permission{
					{
						Resource: "*",
						Actions:  []Action{"*"},
					},
				},
			},
			"developer": {
				Name: "developer",
				Permissions: []*Permission{
					{
						Resource: "container",
						Actions:  []Action{"create", "read", "update", "delete"},
					},
					{
						Resource: "image",
						Actions:  []Action{"read", "pull"},
					},
					{
						Resource: "volume",
						Actions:  []Action{"create", "read", "delete"},
					},
				},
			},
			"operator": {
				Name: "operator",
				Permissions: []*Permission{
					{
						Resource: "container",
						Actions:  []Action{"read", "start", "stop"},
					},
				},
			},
			"viewer": {
				Name: "viewer",
				Permissions: []*Permission{
					{
						Resource: "*",
						Actions:  []Action{"read"},
					},
				},
			},
		},
		DefaultQuotas: map[string]*ResourceQuota{
			"default": {
				MaxContainers: 5,
				MaxCPU:        2.0,
				MaxMemory:     4 * 1024 * 1024 * 1024, // 4Gi
				MaxStorage:    10 * 1024 * 1024 * 1024, // 10Gi
			},
		},
	}
}

// saveUsers persists users to disk
func (m *Manager) saveUsers() error {
	usersFile := filepath.Join(m.storePath, "users.json")

	data, err := json.MarshalIndent(m.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	if err := os.MkdirAll(m.storePath, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	if err := os.WriteFile(usersFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write users file: %w", err)
	}

	return nil
}

// loadUsers loads users from disk
func (m *Manager) loadUsers() error {
	usersFile := filepath.Join(m.storePath, "users.json")

	data, err := os.ReadFile(usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No users file yet
		}
		return fmt.Errorf("failed to read users file: %w", err)
	}

	if err := json.Unmarshal(data, &m.users); err != nil {
		return fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return nil
}
