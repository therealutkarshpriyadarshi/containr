package rbac

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestManager_CreateUser(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	user := NewUser("testuser")
	user.Roles = []*Role{GetBuiltinRole(RoleDeveloper)}

	err := mgr.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify user was created
	retrieved, err := mgr.GetUser("testuser")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", retrieved.Username)
	}
}

func TestManager_CheckPermission(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Create developer user
	user := NewUser("developer")
	user.Roles = []*Role{GetBuiltinRole(RoleDeveloper)}
	mgr.CreateUser(ctx, user)

	tests := []struct {
		name      string
		resource  string
		action    Action
		expected  bool
	}{
		{"container create allowed", "container", ActionCreate, true},
		{"container read allowed", "container", ActionRead, true},
		{"image pull allowed", "image", ActionPull, true},
		{"image push denied", "image", ActionPush, false},
		{"volume create allowed", "volume", ActionCreate, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := mgr.CheckPermission(ctx, "developer", tt.resource, tt.action)
			if err != nil {
				t.Fatalf("CheckPermission failed: %v", err)
			}

			if allowed != tt.expected {
				t.Errorf("expected %v, got %v for %s:%s", tt.expected, allowed, tt.resource, tt.action)
			}
		})
	}
}

func TestManager_EnforceQuota(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Create user with quota
	user := NewUser("quotauser")
	user.Quota = NewResourceQuota(5, 2.0, 4*1024*1024*1024, 10*1024*1024*1024)
	mgr.CreateUser(ctx, user)

	tests := []struct {
		name      string
		usage     *ResourceUsage
		expectErr bool
	}{
		{
			name: "within quota",
			usage: &ResourceUsage{
				Containers: 3,
				CPU:        1.5,
				Memory:     2 * 1024 * 1024 * 1024,
				Storage:    5 * 1024 * 1024 * 1024,
			},
			expectErr: false,
		},
		{
			name: "containers exceeded",
			usage: &ResourceUsage{
				Containers: 6,
				CPU:        1.0,
				Memory:     2 * 1024 * 1024 * 1024,
				Storage:    5 * 1024 * 1024 * 1024,
			},
			expectErr: true,
		},
		{
			name: "CPU exceeded",
			usage: &ResourceUsage{
				Containers: 3,
				CPU:        2.5,
				Memory:     2 * 1024 * 1024 * 1024,
				Storage:    5 * 1024 * 1024 * 1024,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.EnforceQuota(ctx, "quotauser", tt.usage)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestManager_GrantRevokeRole(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Create user with viewer role
	user := NewUser("testuser")
	user.Roles = []*Role{GetBuiltinRole(RoleViewer)}
	mgr.CreateUser(ctx, user)

	// Grant developer role
	err := mgr.GrantRole(ctx, "testuser", RoleDeveloper)
	if err != nil {
		t.Fatalf("GrantRole failed: %v", err)
	}

	// Verify user has both roles
	retrieved, _ := mgr.GetUser("testuser")
	if len(retrieved.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(retrieved.Roles))
	}

	// Revoke viewer role
	err = mgr.RevokeRole(ctx, "testuser", RoleViewer)
	if err != nil {
		t.Fatalf("RevokeRole failed: %v", err)
	}

	// Verify only developer role remains
	retrieved, _ = mgr.GetUser("testuser")
	if len(retrieved.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(retrieved.Roles))
	}
	if retrieved.Roles[0].Name != RoleDeveloper {
		t.Errorf("expected developer role, got %s", retrieved.Roles[0].Name)
	}
}

func TestRole_HasPermission(t *testing.T) {
	role := GetBuiltinRole(RoleDeveloper)

	tests := []struct {
		resource string
		action   Action
		expected bool
	}{
		{"container", ActionCreate, true},
		{"container", ActionRead, true},
		{"image", ActionPull, true},
		{"image", ActionPush, false},
		{"volume", ActionCreate, true},
		{"network", ActionCreate, false},
	}

	for _, tt := range tests {
		result := role.HasPermission(tt.resource, tt.action)
		if result != tt.expected {
			t.Errorf("HasPermission(%s, %s) = %v, expected %v",
				tt.resource, tt.action, result, tt.expected)
		}
	}
}

func TestResourceQuota_Check(t *testing.T) {
	quota := NewResourceQuota(10, 4.0, 8*1024*1024*1024, 50*1024*1024*1024)

	tests := []struct {
		name      string
		usage     *ResourceUsage
		expectErr bool
	}{
		{
			name: "all within limits",
			usage: &ResourceUsage{
				Containers: 5,
				CPU:        2.0,
				Memory:     4 * 1024 * 1024 * 1024,
				Storage:    25 * 1024 * 1024 * 1024,
			},
			expectErr: false,
		},
		{
			name: "containers over limit",
			usage: &ResourceUsage{
				Containers: 11,
				CPU:        2.0,
				Memory:     4 * 1024 * 1024 * 1024,
				Storage:    25 * 1024 * 1024 * 1024,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := quota.Check(tt.usage)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestAuditor_Log(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	auditor, err := NewAuditor(logPath)
	if err != nil {
		t.Fatalf("NewAuditor failed: %v", err)
	}
	defer auditor.Close()

	ctx := context.Background()

	event := AuditEvent{
		User:     "testuser",
		Action:   "container.create",
		Resource: "mycontainer",
		Result:   "success",
	}

	err = auditor.Log(ctx, event)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Verify log file exists and has content
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("audit log file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("audit log file is empty")
	}
}

// setupTestManager creates a test RBAC manager
func setupTestManager(t *testing.T, tmpDir string) *Manager {
	configPath := filepath.Join(tmpDir, "rbac.yaml")
	storePath := filepath.Join(tmpDir, "store")

	// Create default config file
	config := defaultConfig()
	data, _ := yaml.Marshal(config)
	os.WriteFile(configPath, data, 0644)

	mgr, err := NewManager(configPath, storePath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	return mgr
}
