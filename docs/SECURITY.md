# Security Guide for Containr

**Version:** 1.0
**Last Updated:** November 17, 2025
**Status:** Phase 1.2 - Security Foundations Implemented

---

## Table of Contents

- [Overview](#overview)
- [Threat Model](#threat-model)
- [Security Features](#security-features)
  - [Capabilities Management](#capabilities-management)
  - [Seccomp Profiles](#seccomp-profiles)
  - [Linux Security Modules (LSM)](#linux-security-modules-lsm)
- [Best Practices](#best-practices)
- [Security Configuration Examples](#security-configuration-examples)
- [Known Limitations](#known-limitations)
- [Reporting Security Vulnerabilities](#reporting-security-vulnerabilities)

---

## Overview

Containr implements multiple layers of defense to isolate and secure container processes. This document describes the security mechanisms, their configuration, and best practices for using containr safely.

**Important Note:** Containr is an educational container runtime designed for learning purposes. While it implements production-grade security features, it should **not** be used in production environments without thorough security review and hardening.

---

## Threat Model

### Assets to Protect

1. **Host System**: Prevent container escape and host compromise
2. **Container Isolation**: Prevent inter-container interference
3. **Data Confidentiality**: Protect sensitive data in containers
4. **Resource Availability**: Prevent resource exhaustion attacks

### Threat Actors

- **Malicious Container Processes**: Attempting to escape or escalate privileges
- **Compromised Applications**: Exploited applications trying to break isolation
- **Insider Threats**: Users with container access attempting unauthorized actions

### Attack Vectors

1. **Kernel Exploits**: Exploiting kernel vulnerabilities from containers
2. **Capability Abuse**: Misusing granted capabilities to escalate privileges
3. **Syscall Exploitation**: Using dangerous syscalls to compromise the system
4. **Resource Exhaustion**: DoS attacks via uncontrolled resource consumption
5. **Filesystem Attacks**: Attempting to access host files or escape chroot
6. **Network Attacks**: Using network access to attack host or other containers

### Mitigations

| Threat | Mitigation | Implementation |
|--------|-----------|----------------|
| Privilege Escalation | Capabilities Management | Drop dangerous capabilities by default |
| Kernel Exploits | Seccomp Filtering | Block dangerous syscalls |
| Container Escape | Namespaces + LSM | Isolate processes and enforce MAC policies |
| Resource Exhaustion | Cgroups | Enforce memory, CPU, and PID limits |
| Filesystem Escape | Mount Namespaces + Pivot Root | Isolate filesystem view |
| Network Attacks | Network Namespaces | Isolate network stack |

---

## Security Features

### Capabilities Management

Linux capabilities divide root privileges into distinct units. Containr implements fine-grained capability control to follow the **principle of least privilege**.

#### Default Capabilities

Containr grants the following capabilities by default (similar to Docker):

- `CAP_CHOWN` - Change file ownership
- `CAP_DAC_OVERRIDE` - Bypass file permission checks
- `CAP_FSETID` - Don't clear set-user-ID and set-group-ID
- `CAP_FOWNER` - Bypass permission checks on operations that normally require the filesystem UID
- `CAP_MKNOD` - Create special files
- `CAP_NET_RAW` - Use RAW and PACKET sockets
- `CAP_SETGID` - Make arbitrary manipulations of process GIDs
- `CAP_SETUID` - Make arbitrary manipulations of process UIDs
- `CAP_SETFCAP` - Set file capabilities
- `CAP_SETPCAP` - Modify process capabilities
- `CAP_NET_BIND_SERVICE` - Bind to privileged ports (<1024)
- `CAP_SYS_CHROOT` - Use chroot()
- `CAP_KILL` - Send signals to processes
- `CAP_AUDIT_WRITE` - Write to audit log

#### Dangerous Capabilities (Dropped by Default)

The following capabilities are **not** granted by default as they pose significant security risks:

- `CAP_SYS_ADMIN` - Perform system administration operations (very dangerous)
- `CAP_SYS_MODULE` - Load and unload kernel modules
- `CAP_SYS_BOOT` - Reboot the system
- `CAP_SYS_TIME` - Set system clock
- `CAP_SYS_PTRACE` - Trace arbitrary processes
- `CAP_MAC_ADMIN` - Override Mandatory Access Control
- `CAP_MAC_OVERRIDE` - Bypass MAC policy
- `CAP_SYS_RAWIO` - Perform raw I/O operations

#### Configuration

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/capabilities"

// Use default capabilities
config := &capabilities.Config{}

// Drop specific capabilities
config := &capabilities.Config{
    Drop: []capabilities.Capability{
        capabilities.CAP_NET_RAW,
        capabilities.CAP_KILL,
    },
}

// Add specific capabilities
config := &capabilities.Config{
    Add: []capabilities.Capability{
        capabilities.CAP_SYS_TIME,
    },
}

// Privileged mode (all capabilities)
config := &capabilities.Config{
    AllowAll: true,
}
```

---

### Seccomp Profiles

Seccomp (Secure Computing Mode) restricts the syscalls a process can make, reducing the kernel attack surface.

#### Default Profile

Containr's default seccomp profile:

- **Default Action**: `SCMP_ACT_ERRNO` (return EPERM for blocked syscalls)
- **Allowed Syscalls**: ~300 common syscalls needed for normal operation
- **Blocked Syscalls**: Dangerous operations like `mount`, `reboot`, `ptrace`, `kexec_load`, etc.

#### Profile Types

1. **Default Profile**: Restrictive profile blocking dangerous syscalls
2. **Docker-Default**: Compatible with Docker's default seccomp profile
3. **Unconfined**: Allow all syscalls (use with caution)
4. **Custom**: Load from JSON file

#### Configuration

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/seccomp"

// Use default profile
config := &seccomp.Config{}

// Use unconfined profile
config := &seccomp.Config{
    Profile: seccomp.UnconfinedProfile(),
}

// Load custom profile
config := &seccomp.Config{
    ProfilePath: "/path/to/custom-profile.json",
}

// Disable seccomp
config := &seccomp.Config{
    Disabled: true,
}
```

#### Custom Profile Format

Seccomp profiles use JSON format compatible with OCI/Docker:

```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": [
    "SCMP_ARCH_X86_64",
    "SCMP_ARCH_AARCH64"
  ],
  "syscalls": [
    {
      "names": ["read", "write", "open", "close"],
      "action": "SCMP_ACT_ALLOW"
    },
    {
      "names": ["mount", "umount", "reboot"],
      "action": "SCMP_ACT_ERRNO"
    }
  ]
}
```

---

### Linux Security Modules (LSM)

LSM provides Mandatory Access Control (MAC) through AppArmor or SELinux.

#### Supported LSMs

1. **AppArmor**: Path-based MAC (common on Ubuntu/Debian)
2. **SELinux**: Label-based MAC (common on RHEL/Fedora/CentOS)
3. **None**: No LSM available

#### Auto-Detection

Containr automatically detects the available LSM on the host system.

#### AppArmor

AppArmor uses profiles to define allowed operations for processes.

**Default Profile**: `containr-default`

The default profile:
- Allows basic file operations within container root
- Allows network access
- Denies access to sensitive kernel interfaces
- Restricts capabilities

**Custom Profile Example:**

```
profile my-container-profile flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  # Allow network
  network inet tcp,
  network inet udp,

  # Allow file access
  /app/** rw,
  /tmp/** rw,

  # Deny sensitive areas
  deny /sys/kernel/security/** rw,
  deny /proc/kcore r,
}
```

#### SELinux

SELinux uses contexts and policies to enforce access control.

**Default Context**: `system_u:system_r:container_t:s0`

**Custom Context:**

```go
config := &security.Config{
    LSM: security.LSMSELinux,
    ProfileName: "system_u:system_r:my_container_t:s0",
}
```

#### Configuration

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/security"

// Auto-detect and use available LSM
config := &security.Config{}

// Use specific LSM
config := &security.Config{
    LSM: security.LSMAppArmor,
    ProfileName: "my-profile",
}

// Disable LSM
config := &security.Config{
    Disabled: true,
}
```

---

## Best Practices

### 1. **Never Run Privileged Containers in Untrusted Environments**

```go
// BAD: Privileged mode disables all security features
config := &container.Config{
    Privileged: true,  // DON'T DO THIS
}

// GOOD: Use minimal required capabilities
config := &container.Config{
    Capabilities: &capabilities.Config{
        Add: []capabilities.Capability{
            capabilities.CAP_NET_ADMIN,  // Only what's needed
        },
    },
}
```

### 2. **Use Resource Limits**

Always set cgroup limits to prevent resource exhaustion:

```go
cg, _ := cgroup.New(&cgroup.Config{
    Name:        "mycontainer",
    MemoryLimit: 512 * 1024 * 1024,  // 512MB
    CPUShares:   512,                 // 50% CPU
    PIDLimit:    100,                 // Max 100 processes
})
```

### 3. **Drop Unnecessary Capabilities**

Only grant the minimum capabilities needed:

```go
config := &capabilities.Config{
    Drop: []capabilities.Capability{
        capabilities.CAP_NET_RAW,    // Drop if networking not needed
        capabilities.CAP_MKNOD,      // Drop if device creation not needed
    },
}
```

### 4. **Use Custom Seccomp Profiles**

For sensitive applications, create custom seccomp profiles that only allow required syscalls:

```go
config := &seccomp.Config{
    ProfilePath: "/etc/containr/profiles/strict.json",
}
```

### 5. **Enable LSM Profiles**

Always use AppArmor or SELinux when available:

```go
// Check LSM availability
info := security.GetLSMInfo()
if info["available"].(bool) {
    fmt.Printf("LSM available: %s\n", info["type"])
}
```

### 6. **Isolate Network**

Use network namespaces to isolate container networking:

```go
config := &container.Config{
    Isolate: true,  // Enables network isolation
}
```

### 7. **Read-Only Root Filesystem**

When possible, mount the root filesystem as read-only to prevent tampering.

### 8. **Audit and Monitor**

- Enable audit logging for security events
- Monitor for unusual syscalls or capability usage
- Set up alerts for security violations

---

## Security Configuration Examples

### Example 1: Web Server Container (Low Security)

```go
config := &container.Config{
    Command:  []string{"/usr/bin/nginx"},
    Hostname: "web-server",
    Isolate:  true,
    Capabilities: &capabilities.Config{
        Drop: []capabilities.Capability{
            capabilities.CAP_NET_RAW,
            capabilities.CAP_MKNOD,
        },
    },
    Seccomp: &seccomp.Config{
        Profile: seccomp.DefaultProfile(),
    },
    Security: &security.Config{
        LSM: security.LSMAppArmor,
        ProfileName: "web-server-profile",
    },
}
```

### Example 2: Sensitive Application (High Security)

```go
config := &container.Config{
    Command:  []string{"/app/sensitive-app"},
    Hostname: "secure-app",
    Isolate:  true,
    Capabilities: &capabilities.Config{
        // Drop all except minimal set
        Drop: []capabilities.Capability{
            capabilities.CAP_NET_RAW,
            capabilities.CAP_MKNOD,
            capabilities.CAP_SYS_CHROOT,
            capabilities.CAP_KILL,
        },
    },
    Seccomp: &seccomp.Config{
        ProfilePath: "/etc/containr/profiles/strict-app.json",
    },
    Security: &security.Config{
        LSM: security.LSMSELinux,
        ProfileName: "system_u:system_r:sensitive_app_t:s0:c0,c1",
    },
}

// Also set strict resource limits
cg, _ := cgroup.New(&cgroup.Config{
    Name:        "sensitive-app",
    MemoryLimit: 256 * 1024 * 1024,
    CPUShares:   256,
    PIDLimit:    50,
})
```

### Example 3: Development Container (Balanced)

```go
config := &container.Config{
    Command:  []string{"/bin/bash"},
    Hostname: "dev-container",
    Isolate:  true,
    Capabilities: &capabilities.Config{
        // Use defaults
    },
    Seccomp: &seccomp.Config{
        Profile: seccomp.DefaultProfile(),
    },
    Security: &security.Config{
        // Auto-detect LSM
    },
}
```

---

## Known Limitations

### Educational Tool

- **Not Production-Ready**: Containr is designed for education, not production use
- **Limited Testing**: Has not undergone extensive security auditing
- **Missing Features**: Lacks some advanced security features found in Docker/Podman

### Implementation Gaps

1. **Seccomp BPF**: Currently uses a simplified seccomp implementation; full BPF filter loading not implemented
2. **User Namespaces**: Not yet implemented (Phase 2.4)
3. **AppArmor Profile Loading**: Requires manual profile installation via `apparmor_parser`
4. **SELinux Policy Compilation**: Requires manual policy compilation via `checkpolicy`

### Platform Support

- **Linux Only**: Requires Linux kernel 3.8+ with namespace support
- **Root Required**: Most operations require root privileges (rootless mode planned for Phase 2.4)

### Security Considerations

- **Kernel Version**: Security features depend on kernel version
- **LSM Support**: AppArmor/SELinux must be enabled in kernel
- **Seccomp Support**: Requires `CONFIG_SECCOMP` and `CONFIG_SECCOMP_FILTER`

---

## Reporting Security Vulnerabilities

If you discover a security vulnerability in Containr:

1. **Do NOT** open a public GitHub issue
2. Email the maintainer directly with details
3. Allow reasonable time for a fix before public disclosure
4. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### CVE Response Process

1. **Report Received**: Acknowledge within 48 hours
2. **Triage**: Assess severity and impact within 7 days
3. **Fix Development**: Develop and test fix
4. **Disclosure**: Coordinate disclosure with reporter
5. **Release**: Publish patched version and security advisory

---

## Security Checklist

Before deploying containers with Containr:

- [ ] Review and customize capability settings
- [ ] Configure appropriate seccomp profile
- [ ] Enable and configure LSM (AppArmor/SELinux)
- [ ] Set resource limits (memory, CPU, PIDs)
- [ ] Enable network isolation
- [ ] Use read-only root filesystem when possible
- [ ] Audit security configuration
- [ ] Monitor for security events
- [ ] Keep host kernel updated
- [ ] Test container escape scenarios
- [ ] Document security requirements

---

## Additional Resources

- [Linux Capabilities Man Page](https://man7.org/linux/man-pages/man7/capabilities.7.html)
- [Seccomp Documentation](https://www.kernel.org/doc/html/latest/userspace-api/seccomp_filter.html)
- [AppArmor Documentation](https://gitlab.com/apparmor/apparmor/-/wikis/Documentation)
- [SELinux Documentation](https://github.com/SELinuxProject/selinux-notebook)
- [Container Security Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [NIST Application Container Security Guide](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-190.pdf)

---

**Remember**: Security is a continuous process. Regularly review and update security configurations as threats evolve.

**Version History:**
- v1.0 (Nov 2025): Initial security foundation implementation (Phase 1.2)
