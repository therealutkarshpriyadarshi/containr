package capabilities

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// Capability represents a Linux capability
type Capability string

// Linux capabilities
const (
	// Process capabilities
	CAP_CHOWN            Capability = "CAP_CHOWN"
	CAP_DAC_OVERRIDE     Capability = "CAP_DAC_OVERRIDE"
	CAP_DAC_READ_SEARCH  Capability = "CAP_DAC_READ_SEARCH"
	CAP_FOWNER           Capability = "CAP_FOWNER"
	CAP_FSETID           Capability = "CAP_FSETID"
	CAP_KILL             Capability = "CAP_KILL"
	CAP_SETGID           Capability = "CAP_SETGID"
	CAP_SETUID           Capability = "CAP_SETUID"
	CAP_SETPCAP          Capability = "CAP_SETPCAP"
	CAP_LINUX_IMMUTABLE  Capability = "CAP_LINUX_IMMUTABLE"
	CAP_NET_BIND_SERVICE Capability = "CAP_NET_BIND_SERVICE"
	CAP_NET_BROADCAST    Capability = "CAP_NET_BROADCAST"
	CAP_NET_ADMIN        Capability = "CAP_NET_ADMIN"
	CAP_NET_RAW          Capability = "CAP_NET_RAW"
	CAP_IPC_LOCK         Capability = "CAP_IPC_LOCK"
	CAP_IPC_OWNER        Capability = "CAP_IPC_OWNER"
	CAP_SYS_MODULE       Capability = "CAP_SYS_MODULE"
	CAP_SYS_RAWIO        Capability = "CAP_SYS_RAWIO"
	CAP_SYS_CHROOT       Capability = "CAP_SYS_CHROOT"
	CAP_SYS_PTRACE       Capability = "CAP_SYS_PTRACE"
	CAP_SYS_PACCT        Capability = "CAP_SYS_PACCT"
	CAP_SYS_ADMIN        Capability = "CAP_SYS_ADMIN"
	CAP_SYS_BOOT         Capability = "CAP_SYS_BOOT"
	CAP_SYS_NICE         Capability = "CAP_SYS_NICE"
	CAP_SYS_RESOURCE     Capability = "CAP_SYS_RESOURCE"
	CAP_SYS_TIME         Capability = "CAP_SYS_TIME"
	CAP_SYS_TTY_CONFIG   Capability = "CAP_SYS_TTY_CONFIG"
	CAP_MKNOD            Capability = "CAP_MKNOD"
	CAP_LEASE            Capability = "CAP_LEASE"
	CAP_AUDIT_WRITE      Capability = "CAP_AUDIT_WRITE"
	CAP_AUDIT_CONTROL    Capability = "CAP_AUDIT_CONTROL"
	CAP_SETFCAP          Capability = "CAP_SETFCAP"
	CAP_MAC_OVERRIDE     Capability = "CAP_MAC_OVERRIDE"
	CAP_MAC_ADMIN        Capability = "CAP_MAC_ADMIN"
	CAP_SYSLOG           Capability = "CAP_SYSLOG"
	CAP_WAKE_ALARM       Capability = "CAP_WAKE_ALARM"
	CAP_BLOCK_SUSPEND    Capability = "CAP_BLOCK_SUSPEND"
	CAP_AUDIT_READ       Capability = "CAP_AUDIT_READ"
)

// capabilityMap maps capability names to their kernel values
var capabilityMap = map[Capability]uintptr{
	CAP_CHOWN:            unix.CAP_CHOWN,
	CAP_DAC_OVERRIDE:     unix.CAP_DAC_OVERRIDE,
	CAP_DAC_READ_SEARCH:  unix.CAP_DAC_READ_SEARCH,
	CAP_FOWNER:           unix.CAP_FOWNER,
	CAP_FSETID:           unix.CAP_FSETID,
	CAP_KILL:             unix.CAP_KILL,
	CAP_SETGID:           unix.CAP_SETGID,
	CAP_SETUID:           unix.CAP_SETUID,
	CAP_SETPCAP:          unix.CAP_SETPCAP,
	CAP_LINUX_IMMUTABLE:  unix.CAP_LINUX_IMMUTABLE,
	CAP_NET_BIND_SERVICE: unix.CAP_NET_BIND_SERVICE,
	CAP_NET_BROADCAST:    unix.CAP_NET_BROADCAST,
	CAP_NET_ADMIN:        unix.CAP_NET_ADMIN,
	CAP_NET_RAW:          unix.CAP_NET_RAW,
	CAP_IPC_LOCK:         unix.CAP_IPC_LOCK,
	CAP_IPC_OWNER:        unix.CAP_IPC_OWNER,
	CAP_SYS_MODULE:       unix.CAP_SYS_MODULE,
	CAP_SYS_RAWIO:        unix.CAP_SYS_RAWIO,
	CAP_SYS_CHROOT:       unix.CAP_SYS_CHROOT,
	CAP_SYS_PTRACE:       unix.CAP_SYS_PTRACE,
	CAP_SYS_PACCT:        unix.CAP_SYS_PACCT,
	CAP_SYS_ADMIN:        unix.CAP_SYS_ADMIN,
	CAP_SYS_BOOT:         unix.CAP_SYS_BOOT,
	CAP_SYS_NICE:         unix.CAP_SYS_NICE,
	CAP_SYS_RESOURCE:     unix.CAP_SYS_RESOURCE,
	CAP_SYS_TIME:         unix.CAP_SYS_TIME,
	CAP_SYS_TTY_CONFIG:   unix.CAP_SYS_TTY_CONFIG,
	CAP_MKNOD:            unix.CAP_MKNOD,
	CAP_LEASE:            unix.CAP_LEASE,
	CAP_AUDIT_WRITE:      unix.CAP_AUDIT_WRITE,
	CAP_AUDIT_CONTROL:    unix.CAP_AUDIT_CONTROL,
	CAP_SETFCAP:          unix.CAP_SETFCAP,
	CAP_MAC_OVERRIDE:     unix.CAP_MAC_OVERRIDE,
	CAP_MAC_ADMIN:        unix.CAP_MAC_ADMIN,
	CAP_SYSLOG:           unix.CAP_SYSLOG,
	CAP_WAKE_ALARM:       unix.CAP_WAKE_ALARM,
	CAP_BLOCK_SUSPEND:    unix.CAP_BLOCK_SUSPEND,
	CAP_AUDIT_READ:       unix.CAP_AUDIT_READ,
}

// DefaultCapabilities returns a safe default set of capabilities
// Similar to Docker's default capability set
func DefaultCapabilities() []Capability {
	return []Capability{
		CAP_CHOWN,
		CAP_DAC_OVERRIDE,
		CAP_FSETID,
		CAP_FOWNER,
		CAP_MKNOD,
		CAP_NET_RAW,
		CAP_SETGID,
		CAP_SETUID,
		CAP_SETFCAP,
		CAP_SETPCAP,
		CAP_NET_BIND_SERVICE,
		CAP_SYS_CHROOT,
		CAP_KILL,
		CAP_AUDIT_WRITE,
	}
}

// AllCapabilities returns all available Linux capabilities
func AllCapabilities() []Capability {
	caps := make([]Capability, 0, len(capabilityMap))
	for cap := range capabilityMap {
		caps = append(caps, cap)
	}
	return caps
}

// Config represents capability configuration for a container
type Config struct {
	// Add capabilities to add beyond the default set
	Add []Capability
	// Drop capabilities to drop from the default set
	Drop []Capability
	// AllowAll grants all capabilities (privileged mode)
	AllowAll bool
}

// Resolve returns the final set of capabilities after applying add/drop rules
func (c *Config) Resolve() ([]Capability, error) {
	var caps []Capability

	if c.AllowAll {
		return AllCapabilities(), nil
	}

	// Start with default capabilities
	caps = DefaultCapabilities()

	// Remove dropped capabilities
	if len(c.Drop) > 0 {
		caps = removeCapabilities(caps, c.Drop)
	}

	// Add requested capabilities
	if len(c.Add) > 0 {
		caps = addCapabilities(caps, c.Add)
	}

	// Validate all capabilities exist
	for _, cap := range caps {
		if _, ok := capabilityMap[cap]; !ok {
			return nil, fmt.Errorf("unknown capability: %s", cap)
		}
	}

	return caps, nil
}

// Apply applies the capability configuration to the current process
func (c *Config) Apply() error {
	caps, err := c.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve capabilities: %w", err)
	}

	// Drop all capabilities first
	if err := dropAllCapabilities(); err != nil {
		return fmt.Errorf("failed to drop all capabilities: %w", err)
	}

	// Add back the allowed capabilities
	for _, cap := range caps {
		if err := addCapability(cap); err != nil {
			return fmt.Errorf("failed to add capability %s: %w", cap, err)
		}
	}

	return nil
}

// dropAllCapabilities drops all capabilities from the current process
func dropAllCapabilities() error {
	// Drop all capabilities from all capability sets
	for i := uintptr(0); i <= unix.CAP_LAST_CAP; i++ {
		// Drop from effective set
		if err := unix.Prctl(unix.PR_CAPBSET_DROP, i, 0, 0, 0); err != nil {
			// Ignore EINVAL - capability doesn't exist on this kernel
			if err != unix.EINVAL {
				return fmt.Errorf("failed to drop capability %d: %w", i, err)
			}
		}
	}
	return nil
}

// addCapability adds a single capability to the current process
func addCapability(cap Capability) error {
	capValue, ok := capabilityMap[cap]
	if !ok {
		return fmt.Errorf("unknown capability: %s", cap)
	}

	// Set capability in the bounding set
	if err := unix.Prctl(unix.PR_CAP_AMBIENT, unix.PR_CAP_AMBIENT_RAISE, capValue, 0, 0); err != nil {
		// If ambient caps not supported, just return success
		// The capability will be inherited if it's in the permitted set
		if err != unix.EINVAL && err != unix.EPERM {
			return fmt.Errorf("failed to raise ambient capability: %w", err)
		}
	}

	return nil
}

// removeCapabilities removes capabilities from a list
func removeCapabilities(caps []Capability, remove []Capability) []Capability {
	removeMap := make(map[Capability]bool)
	for _, cap := range remove {
		removeMap[cap] = true
	}

	result := make([]Capability, 0, len(caps))
	for _, cap := range caps {
		if !removeMap[cap] {
			result = append(result, cap)
		}
	}
	return result
}

// addCapabilities adds capabilities to a list (avoiding duplicates)
func addCapabilities(caps []Capability, add []Capability) []Capability {
	capMap := make(map[Capability]bool)
	for _, cap := range caps {
		capMap[cap] = true
	}

	for _, cap := range add {
		if !capMap[cap] {
			caps = append(caps, cap)
			capMap[cap] = true
		}
	}
	return caps
}

// ParseCapability parses a capability string
func ParseCapability(s string) (Capability, error) {
	// Normalize to uppercase and ensure CAP_ prefix
	s = strings.ToUpper(s)
	if !strings.HasPrefix(s, "CAP_") {
		s = "CAP_" + s
	}

	cap := Capability(s)
	if _, ok := capabilityMap[cap]; !ok {
		return "", fmt.Errorf("unknown capability: %s", s)
	}
	return cap, nil
}

// String returns the string representation of a capability
func (c Capability) String() string {
	return string(c)
}
