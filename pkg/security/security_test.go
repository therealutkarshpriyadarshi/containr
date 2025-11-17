package security

import (
	"os"
	"testing"
)

func TestDetectLSM(t *testing.T) {
	lsm := DetectLSM()

	// LSM should be one of the known types
	validLSMs := map[LSMType]bool{
		LSMAppArmor: true,
		LSMSELinux:  true,
		LSMNone:     true,
	}

	if !validLSMs[lsm] {
		t.Errorf("DetectLSM returned invalid LSM type: %s", lsm)
	}

	t.Logf("Detected LSM: %s", lsm)
}

func TestIsAppArmorEnabled(t *testing.T) {
	enabled := isAppArmorEnabled()
	t.Logf("AppArmor enabled: %v", enabled)

	// If AppArmor is enabled, the directory should exist
	if enabled {
		if _, err := os.Stat("/sys/kernel/security/apparmor"); err != nil {
			if _, err := os.Stat("/sys/module/apparmor"); err != nil {
				t.Error("AppArmor reported as enabled but directories don't exist")
			}
		}
	}
}

func TestIsSELinuxEnabled(t *testing.T) {
	enabled := isSELinuxEnabled()
	t.Logf("SELinux enabled: %v", enabled)

	// If SELinux is enabled, the directory should exist
	if enabled {
		if _, err := os.Stat("/sys/fs/selinux"); err != nil {
			if _, err := os.Stat("/selinux"); err != nil {
				t.Error("SELinux reported as enabled but directories don't exist")
			}
		}
	}
}

func TestConfig_Apply(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "default configuration",
			config: &Config{
				LSM: DetectLSM(),
			},
			wantErr: false,
		},
		{
			name: "disabled LSM",
			config: &Config{
				Disabled: true,
			},
			wantErr: false,
		},
		{
			name: "AppArmor with custom profile",
			config: &Config{
				LSM:         LSMAppArmor,
				ProfileName: "test-profile",
			},
			wantErr: false,
		},
		{
			name: "SELinux with custom context",
			config: &Config{
				LSM:         LSMSELinux,
				ProfileName: "system_u:system_r:container_t:s0",
			},
			wantErr: false,
		},
		{
			name: "no LSM",
			config: &Config{
				LSM: LSMNone,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Apply()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetLSMInfo(t *testing.T) {
	info := GetLSMInfo()

	if info == nil {
		t.Fatal("GetLSMInfo returned nil")
	}

	// Check required fields
	if _, ok := info["type"]; !ok {
		t.Error("GetLSMInfo missing 'type' field")
	}

	if _, ok := info["available"]; !ok {
		t.Error("GetLSMInfo missing 'available' field")
	}

	lsmType := info["type"].(string)
	available := info["available"].(bool)

	t.Logf("LSM Info: type=%s, available=%v", lsmType, available)

	// If LSM is available, additional fields should be present
	if available {
		if lsmType == string(LSMAppArmor) {
			if _, ok := info["apparmor_enabled"]; !ok {
				t.Error("GetLSMInfo missing 'apparmor_enabled' field for AppArmor")
			}
		} else if lsmType == string(LSMSELinux) {
			if _, ok := info["selinux_enabled"]; !ok {
				t.Error("GetLSMInfo missing 'selinux_enabled' field for SELinux")
			}
		}
	}
}

func TestDefaultAppArmorProfile(t *testing.T) {
	profile := DefaultAppArmorProfile()

	if profile == "" {
		t.Fatal("DefaultAppArmorProfile returned empty string")
	}

	// Check for essential elements
	essentialElements := []string{
		"profile containr-default",
		"capability",
		"deny",
		"network",
	}

	for _, element := range essentialElements {
		if !contains(profile, element) {
			t.Errorf("DefaultAppArmorProfile missing element: %s", element)
		}
	}

	t.Logf("AppArmor profile length: %d bytes", len(profile))
}

func TestDefaultSELinuxPolicy(t *testing.T) {
	policy := DefaultSELinuxPolicy()

	if policy == "" {
		t.Fatal("DefaultSELinuxPolicy returned empty string")
	}

	// Check for essential elements
	essentialElements := []string{
		"module containr",
		"type container_t",
		"allow",
		"neverallow",
	}

	for _, element := range essentialElements {
		if !contains(policy, element) {
			t.Errorf("DefaultSELinuxPolicy missing element: %s", element)
		}
	}

	t.Logf("SELinux policy length: %d bytes", len(policy))
}

func TestLSMTypes(t *testing.T) {
	// Test LSM type constants
	if LSMAppArmor != "apparmor" {
		t.Errorf("LSMAppArmor = %s, want apparmor", LSMAppArmor)
	}

	if LSMSELinux != "selinux" {
		t.Errorf("LSMSELinux = %s, want selinux", LSMSELinux)
	}

	if LSMNone != "none" {
		t.Errorf("LSMNone = %s, want none", LSMNone)
	}
}

func TestSetAppArmorProfile(t *testing.T) {
	config := &Config{
		ProfileName: "test-profile",
	}

	// This will likely fail in test environment without AppArmor
	// but should not panic
	err := config.setAppArmorProfile("test-profile")

	// We don't check for error because it may fail in test environment
	// The important thing is that it doesn't panic
	t.Logf("setAppArmorProfile result: %v", err)
}

func TestSetSELinuxContext(t *testing.T) {
	config := &Config{
		ProfileName: "system_u:system_r:container_t:s0",
	}

	// This will likely fail in test environment without SELinux
	// but should not panic
	err := config.setSELinuxContext("system_u:system_r:container_t:s0")

	// We don't check for error because it may fail in test environment
	// The important thing is that it doesn't panic
	t.Logf("setSELinuxContext result: %v", err)
}

func TestConfigWithProfilePath(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := tmpDir + "/test-profile"

	// Create a dummy profile file
	if err := os.WriteFile(profilePath, []byte("test profile content"), 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	config := &Config{
		LSM:         LSMAppArmor,
		ProfilePath: profilePath,
	}

	// Apply should not error (though it may not actually set the profile)
	err := config.Apply()
	if err != nil {
		t.Logf("Config.Apply() with profile path: %v", err)
	}
}

func TestAutoDetectLSM(t *testing.T) {
	config := &Config{
		// LSM not specified, should auto-detect
	}

	err := config.Apply()
	if err != nil {
		t.Errorf("Config.Apply() with auto-detect failed: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkDetectLSM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DetectLSM()
	}
}

func BenchmarkGetLSMInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetLSMInfo()
	}
}

func BenchmarkDefaultAppArmorProfile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultAppArmorProfile()
	}
}
