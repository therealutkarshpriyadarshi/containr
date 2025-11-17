// Package runtime provides OCI runtime specification compliance
package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Spec represents the OCI runtime specification
type Spec struct {
	Version  string   `json:"ociVersion"`
	Root     *Root    `json:"root"`
	Hostname string   `json:"hostname,omitempty"`
	Mounts   []Mount  `json:"mounts,omitempty"`
	Process  *Process `json:"process,omitempty"`
	Linux    *Linux   `json:"linux,omitempty"`
}

// Root represents the root filesystem configuration
type Root struct {
	Path     string `json:"path"`
	Readonly bool   `json:"readonly,omitempty"`
}

// Mount represents a mount configuration
type Mount struct {
	Destination string   `json:"destination"`
	Type        string   `json:"type,omitempty"`
	Source      string   `json:"source,omitempty"`
	Options     []string `json:"options,omitempty"`
}

// Process represents the process configuration
type Process struct {
	Terminal bool     `json:"terminal,omitempty"`
	User     User     `json:"user"`
	Args     []string `json:"args,omitempty"`
	Env      []string `json:"env,omitempty"`
	Cwd      string   `json:"cwd,omitempty"`
}

// User represents the user configuration
type User struct {
	UID            uint32   `json:"uid"`
	GID            uint32   `json:"gid"`
	AdditionalGids []uint32 `json:"additionalGids,omitempty"`
}

// Linux represents Linux-specific configuration
type Linux struct {
	Namespaces  []Namespace `json:"namespaces,omitempty"`
	Resources   *Resources  `json:"resources,omitempty"`
	CgroupsPath string      `json:"cgroupsPath,omitempty"`
	Seccomp     *Seccomp    `json:"seccomp,omitempty"`
}

// Namespace represents a namespace configuration
type Namespace struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

// Resources represents resource limits
type Resources struct {
	Memory *Memory `json:"memory,omitempty"`
	CPU    *CPU    `json:"cpu,omitempty"`
	Pids   *Pids   `json:"pids,omitempty"`
}

// Memory represents memory resource limits
type Memory struct {
	Limit       *int64 `json:"limit,omitempty"`
	Reservation *int64 `json:"reservation,omitempty"`
	Swap        *int64 `json:"swap,omitempty"`
}

// CPU represents CPU resource limits
type CPU struct {
	Shares *uint64 `json:"shares,omitempty"`
	Quota  *int64  `json:"quota,omitempty"`
	Period *uint64 `json:"period,omitempty"`
}

// Pids represents PID resource limits
type Pids struct {
	Limit int64 `json:"limit"`
}

// Seccomp represents seccomp configuration
type Seccomp struct {
	DefaultAction string          `json:"defaultAction"`
	Architectures []string        `json:"architectures,omitempty"`
	Syscalls      []SyscallConfig `json:"syscalls,omitempty"`
}

// SyscallConfig represents a syscall configuration
type SyscallConfig struct {
	Names  []string `json:"names"`
	Action string   `json:"action"`
}

// State represents the runtime state of a container
type State struct {
	Version     string            `json:"ociVersion"`
	ID          string            `json:"id"`
	Status      ContainerStatus   `json:"status"`
	Pid         int               `json:"pid,omitempty"`
	Bundle      string            `json:"bundle"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"createdAt,omitempty"`
}

// ContainerStatus represents the status of a container
type ContainerStatus string

const (
	// StatusCreating indicates the container is being created
	StatusCreating ContainerStatus = "creating"
	// StatusCreated indicates the container has been created
	StatusCreated ContainerStatus = "created"
	// StatusRunning indicates the container is running
	StatusRunning ContainerStatus = "running"
	// StatusStopped indicates the container has stopped
	StatusStopped ContainerStatus = "stopped"
)

// DefaultSpec returns a default OCI runtime specification
func DefaultSpec() *Spec {
	return &Spec{
		Version:  "1.0.2",
		Hostname: "containr",
		Root: &Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Process: &Process{
			Terminal: false,
			User: User{
				UID: 0,
				GID: 0,
			},
			Args: []string{"/bin/sh"},
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd: "/",
		},
		Linux: &Linux{
			Namespaces: []Namespace{
				{Type: "pid"},
				{Type: "network"},
				{Type: "ipc"},
				{Type: "uts"},
				{Type: "mount"},
			},
		},
	}
}

// LoadSpec loads an OCI spec from a file
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}

	return &spec, nil
}

// Save saves the spec to a file
func (s *Spec) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}

// Validate validates the spec
func (s *Spec) Validate() error {
	if s.Version == "" {
		return fmt.Errorf("ociVersion is required")
	}

	if s.Root == nil {
		return fmt.Errorf("root is required")
	}

	if s.Root.Path == "" {
		return fmt.Errorf("root.path is required")
	}

	if s.Process == nil {
		return fmt.Errorf("process is required")
	}

	if len(s.Process.Args) == 0 {
		return fmt.Errorf("process.args is required")
	}

	return nil
}

// LoadState loads container state from a file
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

// Save saves the state to a file
func (s *State) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
