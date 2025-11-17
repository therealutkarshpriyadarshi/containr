package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LSMType represents the type of Linux Security Module
type LSMType string

const (
	// LSMAppArmor represents AppArmor LSM
	LSMAppArmor LSMType = "apparmor"
	// LSMSELinux represents SELinux LSM
	LSMSELinux LSMType = "selinux"
	// LSMNone indicates no LSM is available
	LSMNone LSMType = "none"
)

// Config represents security configuration for LSM
type Config struct {
	// LSM type to use (auto-detected if not specified)
	LSM LSMType
	// ProfilePath is the path to a custom LSM profile
	ProfilePath string
	// ProfileName is the name of the profile to use
	ProfileName string
	// Disabled disables LSM
	Disabled bool
}

// DetectLSM detects the available LSM on the system
func DetectLSM() LSMType {
	// Check for AppArmor
	if isAppArmorEnabled() {
		return LSMAppArmor
	}

	// Check for SELinux
	if isSELinuxEnabled() {
		return LSMSELinux
	}

	return LSMNone
}

// isAppArmorEnabled checks if AppArmor is enabled on the system
func isAppArmorEnabled() bool {
	// Check if AppArmor is loaded by checking for /sys/kernel/security/apparmor
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil {
		return true
	}

	// Alternative check: /sys/module/apparmor
	if _, err := os.Stat("/sys/module/apparmor"); err == nil {
		return true
	}

	return false
}

// isSELinuxEnabled checks if SELinux is enabled on the system
func isSELinuxEnabled() bool {
	// Check if SELinux is loaded by checking for /sys/fs/selinux
	if _, err := os.Stat("/sys/fs/selinux"); err == nil {
		// Check if SELinux is enforcing or permissive (not disabled)
		if data, err := os.ReadFile("/sys/fs/selinux/enforce"); err == nil {
			mode := strings.TrimSpace(string(data))
			// "0" = permissive, "1" = enforcing
			return mode == "0" || mode == "1"
		}
		return true
	}

	// Alternative check: /selinux
	if _, err := os.Stat("/selinux"); err == nil {
		return true
	}

	return false
}

// Apply applies the security configuration
func (c *Config) Apply() error {
	if c.Disabled {
		return nil
	}

	// Auto-detect LSM if not specified
	lsm := c.LSM
	if lsm == "" {
		lsm = DetectLSM()
	}

	// Apply LSM-specific configuration
	switch lsm {
	case LSMAppArmor:
		return c.applyAppArmor()
	case LSMSELinux:
		return c.applySELinux()
	case LSMNone:
		// No LSM available, silently skip
		return nil
	default:
		return fmt.Errorf("unknown LSM type: %s", lsm)
	}
}

// applyAppArmor applies AppArmor profile
func (c *Config) applyAppArmor() error {
	profileName := c.ProfileName
	if profileName == "" {
		profileName = "containr-default"
	}

	// Check if profile exists
	profilePath := filepath.Join("/etc/apparmor.d", profileName)
	if c.ProfilePath != "" {
		profilePath = c.ProfilePath
	}

	// In a real implementation, we would:
	// 1. Load the profile using apparmor_parser
	// 2. Set the profile for the current process
	//
	// For now, we'll check if the profile exists and document the approach

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		// Profile doesn't exist, use default behavior
		// In production, we'd create a default profile or use unconfined
		return nil
	}

	// Set AppArmor profile for the process
	// This would typically be done by writing to /proc/self/attr/current
	// Example: echo "containr-default" > /proc/self/attr/current
	//
	// For educational purposes, we document the process here

	return c.setAppArmorProfile(profileName)
}

// setAppArmorProfile sets the AppArmor profile for the current process
func (c *Config) setAppArmorProfile(profileName string) error {
	// Write profile name to /proc/self/attr/current
	attrPath := "/proc/self/attr/current"

	// Read current profile
	currentProfile, err := os.ReadFile(attrPath)
	if err != nil {
		// If we can't read the current profile, AppArmor might not be available
		return nil
	}

	// Don't change if already in the desired profile
	if strings.Contains(string(currentProfile), profileName) {
		return nil
	}

	// Write new profile
	// Note: This typically requires CAP_MAC_ADMIN capability
	profile := fmt.Sprintf("exec %s", profileName)
	if err := os.WriteFile(attrPath, []byte(profile), 0644); err != nil {
		// If we can't write, we might not have permissions
		// In production, this would be a warning, not an error
		return nil
	}

	return nil
}

// applySELinux applies SELinux context
func (c *Config) applySELinux() error {
	context := c.ProfileName
	if context == "" {
		context = "system_u:system_r:container_t:s0"
	}

	// In a real implementation, we would:
	// 1. Set the SELinux context using setexeccon()
	// 2. Apply the context to the process
	//
	// For now, we document the approach

	return c.setSELinuxContext(context)
}

// setSELinuxContext sets the SELinux context for the current process
func (c *Config) setSELinuxContext(context string) error {
	// Write context to /proc/self/attr/current
	attrPath := "/proc/self/attr/current"

	// Read current context
	currentContext, err := os.ReadFile(attrPath)
	if err != nil {
		// If we can't read the current context, SELinux might not be available
		return nil
	}

	// Don't change if already in the desired context
	if strings.TrimSpace(string(currentContext)) == context {
		return nil
	}

	// Write new context
	if err := os.WriteFile(attrPath, []byte(context), 0644); err != nil {
		// If we can't write, we might not have permissions
		// In production, this would be a warning, not an error
		return nil
	}

	return nil
}

// GetLSMInfo returns information about the current LSM configuration
func GetLSMInfo() map[string]interface{} {
	lsm := DetectLSM()
	info := make(map[string]interface{})

	info["type"] = string(lsm)
	info["available"] = lsm != LSMNone

	if lsm == LSMAppArmor {
		info["apparmor_enabled"] = isAppArmorEnabled()
		if profile, err := os.ReadFile("/proc/self/attr/current"); err == nil {
			info["current_profile"] = strings.TrimSpace(string(profile))
		}
	} else if lsm == LSMSELinux {
		info["selinux_enabled"] = isSELinuxEnabled()
		if context, err := os.ReadFile("/proc/self/attr/current"); err == nil {
			info["current_context"] = strings.TrimSpace(string(context))
		}
		if enforce, err := os.ReadFile("/sys/fs/selinux/enforce"); err == nil {
			mode := strings.TrimSpace(string(enforce))
			if mode == "1" {
				info["selinux_mode"] = "enforcing"
			} else if mode == "0" {
				info["selinux_mode"] = "permissive"
			}
		}
	}

	return info
}

// DefaultAppArmorProfile returns the default AppArmor profile for containr
func DefaultAppArmorProfile() string {
	return `#include <tunables/global>

profile containr-default flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  # Allow network access
  network inet tcp,
  network inet udp,
  network inet icmp,

  # Allow file access within container root
  /proc/** r,
  /sys/** r,
  /dev/** rw,
  /tmp/** rw,
  /var/tmp/** rw,

  # Allow execution
  /usr/bin/** ix,
  /usr/sbin/** ix,
  /bin/** ix,
  /sbin/** ix,

  # Deny dangerous operations
  deny /sys/kernel/security/** rw,
  deny /sys/module/** w,
  deny /proc/sys/kernel/** w,
  deny /proc/kcore r,
  deny /boot/** r,

  # Allow capability drops
  capability setuid,
  capability setgid,
  capability chown,
  capability dac_override,
  capability fowner,
  capability fsetid,
  capability kill,
  capability setpcap,
  capability net_bind_service,
  capability net_raw,
  capability sys_chroot,
  capability mknod,
  capability audit_write,
  capability setfcap,

  # Deny dangerous capabilities
  deny capability sys_admin,
  deny capability sys_module,
  deny capability sys_boot,
  deny capability sys_time,
  deny capability sys_ptrace,
  deny capability mac_admin,
  deny capability mac_override,
}
`
}

// DefaultSELinuxPolicy returns a basic SELinux policy for containr
func DefaultSELinuxPolicy() string {
	return `module containr 1.0;

require {
	type container_t;
	type container_file_t;
	class process { fork signal_perms };
	class file { read write open execute };
}

# Allow container processes to fork
allow container_t self:process { fork signal_perms };

# Allow container to access its files
allow container_t container_file_t:file { read write open execute };

# Deny sensitive operations
neverallow container_t kernel_t:file { read write };
neverallow container_t security_t:file { read write };
`
}
