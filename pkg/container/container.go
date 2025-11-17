package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/therealutkarshpriyadarshi/containr/pkg/namespace"
)

// Container represents a container instance
type Container struct {
	ID         string
	RootFS     string
	Command    []string
	WorkingDir string
	Hostname   string
	Namespaces []namespace.NamespaceType
}

// Config holds container configuration
type Config struct {
	RootFS     string
	Command    []string
	WorkingDir string
	Hostname   string
	Isolate    bool // Enable full isolation
}

// New creates a new container instance
func New(config *Config) *Container {
	id := generateID()

	namespaces := []namespace.NamespaceType{
		namespace.UTS,
		namespace.PID,
		namespace.Mount,
	}

	if config.Isolate {
		namespaces = append(namespaces,
			namespace.IPC,
			namespace.Network,
		)
	}

	return &Container{
		ID:         id,
		RootFS:     config.RootFS,
		Command:    config.Command,
		WorkingDir: config.WorkingDir,
		Hostname:   config.Hostname,
		Namespaces: namespaces,
	}
}

// Run executes the container
func (c *Container) Run() error {
	// Get namespace flags
	flags := namespace.GetNamespaceFlags(c.Namespaces...)

	// Create the command
	cmd := exec.Command(c.Command[0], c.Command[1:]...)

	// Set up namespace isolation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

	// Set up standard streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set working directory
	if c.WorkingDir != "" {
		cmd.Dir = c.WorkingDir
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container exited with error: %w", err)
	}

	return nil
}

// RunWithSetup executes the container with additional setup in the child process
func (c *Container) RunWithSetup() error {
	// Get namespace flags
	flags := namespace.GetNamespaceFlags(c.Namespaces...)

	// Create the command
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, c.Command...)...)

	// Set up namespace isolation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

	// Set up standard streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pass container config via environment
	cmd.Env = []string{
		fmt.Sprintf("CONTAINER_ID=%s", c.ID),
		fmt.Sprintf("CONTAINER_ROOTFS=%s", c.RootFS),
		fmt.Sprintf("CONTAINER_HOSTNAME=%s", c.Hostname),
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Container process exited with non-zero status
			return fmt.Errorf("container exited with status %d", exitErr.ExitCode())
		}
		return fmt.Errorf("container error: %w", err)
	}

	return nil
}

// SetupChild performs setup inside the child process namespace
func SetupChild() error {
	fmt.Printf("Container starting (PID: %d)\n", os.Getpid())

	// Set hostname if specified
	hostname := os.Getenv("CONTAINER_HOSTNAME")
	if hostname != "" {
		if err := syscall.Sethostname([]byte(hostname)); err != nil {
			return fmt.Errorf("failed to set hostname: %w", err)
		}
	}

	// Mount proc filesystem
	if err := mountProc(); err != nil {
		return fmt.Errorf("failed to mount /proc: %w", err)
	}

	return nil
}

// mountProc mounts the /proc filesystem
func mountProc() error {
	// Create /proc directory if it doesn't exist
	if err := os.MkdirAll("/proc", 0755); err != nil {
		return err
	}

	// Mount proc
	return syscall.Mount("proc", "/proc", "proc", 0, "")
}

// generateID generates a simple container ID
func generateID() string {
	// In a real implementation, use UUID or similar
	return fmt.Sprintf("container-%d", os.Getpid())
}

// SetupRootFS sets up the root filesystem for the container
func (c *Container) SetupRootFS() error {
	if c.RootFS == "" {
		return nil
	}

	// Ensure root filesystem exists
	if _, err := os.Stat(c.RootFS); os.IsNotExist(err) {
		return fmt.Errorf("root filesystem does not exist: %s", c.RootFS)
	}

	// Create mount point
	mountPoint := filepath.Join("/tmp", c.ID, "rootfs")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Bind mount the root filesystem
	if err := syscall.Mount(c.RootFS, mountPoint, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind mount rootfs: %w", err)
	}

	return nil
}
