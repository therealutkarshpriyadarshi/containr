package namespace

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Namespace types that can be isolated
type NamespaceType int

const (
	UTS     NamespaceType = 1 << iota // Hostname and domain name
	IPC                               // Inter-process communication
	PID                               // Process IDs
	Mount                             // Mount points
	Network                           // Network devices, stacks, ports
	User                              // User and group IDs
)

// Config holds namespace configuration
type Config struct {
	Flags      int      // Namespace flags to use
	Command    string   // Command to execute
	Args       []string // Command arguments
	WorkingDir string   // Working directory
}

// CreateNamespaces creates a new process with specified namespaces
func CreateNamespaces(config *Config) error {
	// Create the command with namespace isolation
	cmd := exec.Command(config.Command, config.Args...)

	// Set up namespace flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(config.Flags),
	}

	// Set up standard streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command in namespace: %w", err)
	}

	return nil
}

// GetNamespaceFlags converts namespace types to syscall flags
func GetNamespaceFlags(types ...NamespaceType) int {
	flags := 0
	for _, t := range types {
		switch t {
		case UTS:
			flags |= syscall.CLONE_NEWUTS
		case IPC:
			flags |= syscall.CLONE_NEWIPC
		case PID:
			flags |= syscall.CLONE_NEWPID
		case Mount:
			flags |= syscall.CLONE_NEWNS
		case Network:
			flags |= syscall.CLONE_NEWNET
		case User:
			flags |= syscall.CLONE_NEWUSER
		}
	}
	return flags
}

// Reexec re-executes the current process with the "init" argument
// This is necessary for proper PID namespace initialization
func Reexec() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		// We're in the re-executed process
		return
	}

	// Re-execute ourselves
	cmd := exec.Command("/proc/self/exe", append([]string{"init"}, os.Args[1:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error re-executing: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
