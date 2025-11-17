package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/therealutkarshpriyadarshi/containr/pkg/capabilities"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/namespace"
	"github.com/therealutkarshpriyadarshi/containr/pkg/seccomp"
	"github.com/therealutkarshpriyadarshi/containr/pkg/security"
)

var log = logger.New("container")

// Container represents a container instance
type Container struct {
	ID           string
	RootFS       string
	Command      []string
	WorkingDir   string
	Hostname     string
	Namespaces   []namespace.NamespaceType
	Capabilities *capabilities.Config
	Seccomp      *seccomp.Config
	Security     *security.Config
}

// Config holds container configuration
type Config struct {
	RootFS       string
	Command      []string
	WorkingDir   string
	Hostname     string
	Isolate      bool // Enable full isolation
	Capabilities *capabilities.Config
	Seccomp      *seccomp.Config
	Security     *security.Config
	Privileged   bool // Run in privileged mode (disables security restrictions)
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

	// Set up default security configurations if not provided
	capConfig := config.Capabilities
	if capConfig == nil {
		if config.Privileged {
			// Privileged mode: allow all capabilities
			capConfig = &capabilities.Config{
				AllowAll: true,
			}
		} else {
			// Default: use default safe capability set
			capConfig = &capabilities.Config{}
		}
	}

	seccompConfig := config.Seccomp
	if seccompConfig == nil {
		if config.Privileged {
			// Privileged mode: disable seccomp
			seccompConfig = &seccomp.Config{
				Disabled: true,
			}
		} else {
			// Default: use default restrictive profile
			seccompConfig = &seccomp.Config{}
		}
	}

	securityConfig := config.Security
	if securityConfig == nil {
		if config.Privileged {
			// Privileged mode: disable LSM
			securityConfig = &security.Config{
				Disabled: true,
			}
		} else {
			// Default: auto-detect and use LSM
			securityConfig = &security.Config{}
		}
	}

	return &Container{
		ID:           id,
		RootFS:       config.RootFS,
		Command:      config.Command,
		WorkingDir:   config.WorkingDir,
		Hostname:     config.Hostname,
		Namespaces:   namespaces,
		Capabilities: capConfig,
		Seccomp:      seccompConfig,
		Security:     securityConfig,
	}
}

// Run executes the container
func (c *Container) Run() error {
	log.WithField("container_id", c.ID).Debug("Starting container execution")

	// Get namespace flags
	flags := namespace.GetNamespaceFlags(c.Namespaces...)
	log.WithField("container_id", c.ID).Debugf("Namespace flags: %v", flags)

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
		log.WithField("container_id", c.ID).Debugf("Setting working directory: %s", c.WorkingDir)
		cmd.Dir = c.WorkingDir
	}

	// Start the process
	log.WithField("container_id", c.ID).Debug("Starting container process")
	if err := cmd.Start(); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Error("Failed to start container")
		return errors.Wrap(errors.ErrContainerStart, "failed to start container process", err).
			WithField("container_id", c.ID).
			WithHint("Ensure the command exists and you have sufficient privileges")
	}

	log.WithField("container_id", c.ID).Debugf("Container process started with PID: %d", cmd.Process.Pid)

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Warn("Container exited with error")
		if exitErr, ok := err.(*exec.ExitError); ok {
			return errors.New(errors.ErrContainerStart, fmt.Sprintf("container exited with status %d", exitErr.ExitCode())).
				WithField("container_id", c.ID).
				WithField("exit_code", exitErr.ExitCode())
		}
		return errors.Wrap(errors.ErrContainerStart, "container error", err).
			WithField("container_id", c.ID)
	}

	log.WithField("container_id", c.ID).Info("Container completed successfully")
	return nil
}

// RunWithSetup executes the container with additional setup in the child process
func (c *Container) RunWithSetup() error {
	log.WithField("container_id", c.ID).Debug("Starting container with setup")

	// Get namespace flags
	flags := namespace.GetNamespaceFlags(c.Namespaces...)
	log.WithField("container_id", c.ID).Debugf("Namespace flags: %v", flags)

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

	log.WithField("container_id", c.ID).Debug("Starting container child process")

	// Start the process
	if err := cmd.Start(); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Error("Failed to start container")
		return errors.Wrap(errors.ErrContainerStart, "failed to start container child process", err).
			WithField("container_id", c.ID).
			WithHint("Ensure you have root privileges and the executable is accessible")
	}

	log.WithField("container_id", c.ID).Debugf("Container child process started with PID: %d", cmd.Process.Pid)

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Warn("Container exited with error")
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Container process exited with non-zero status
			return errors.New(errors.ErrContainerStart, fmt.Sprintf("container exited with status %d", exitErr.ExitCode())).
				WithField("container_id", c.ID).
				WithField("exit_code", exitErr.ExitCode())
		}
		return errors.Wrap(errors.ErrContainerStart, "container error", err).
			WithField("container_id", c.ID)
	}

	log.WithField("container_id", c.ID).Info("Container completed successfully")
	return nil
}

// SetupChild performs setup inside the child process namespace
func SetupChild() error {
	childLog := logger.New("child-setup")
	pid := os.Getpid()

	childLog.WithField("pid", pid).Debug("Child process starting")
	fmt.Printf("Container starting (PID: %d)\n", pid)

	// Set hostname if specified
	hostname := os.Getenv("CONTAINER_HOSTNAME")
	if hostname != "" {
		childLog.WithField("hostname", hostname).Debug("Setting hostname")
		if err := syscall.Sethostname([]byte(hostname)); err != nil {
			childLog.WithError(err).Error("Failed to set hostname")
			return errors.Wrap(errors.ErrNamespaceSetup, "failed to set hostname", err).
				WithField("hostname", hostname).
				WithHint("Ensure you are running with root privileges")
		}
		childLog.WithField("hostname", hostname).Info("Hostname set successfully")
	}

	// Mount proc filesystem
	childLog.Debug("Mounting /proc filesystem")
	if err := mountProc(); err != nil {
		childLog.WithError(err).Error("Failed to mount /proc")
		return errors.Wrap(errors.ErrRootFSMount, "failed to mount /proc", err).
			WithHint("Ensure you have CAP_SYS_ADMIN capability")
	}
	childLog.Info("/proc filesystem mounted successfully")

	return nil
}

// ApplySecurity applies security configurations (capabilities, seccomp, LSM)
func (c *Container) ApplySecurity() error {
	log.WithField("container_id", c.ID).Debug("Applying security configurations")

	// Apply LSM (AppArmor/SELinux) first
	if c.Security != nil {
		log.WithField("container_id", c.ID).Debug("Applying LSM configuration")
		if err := c.Security.Apply(); err != nil {
			log.WithError(err).WithField("container_id", c.ID).Error("Failed to apply LSM configuration")
			return errors.Wrap(errors.ErrSecurityLSM, "failed to apply LSM configuration", err).
				WithField("container_id", c.ID).
				WithHint("Check if AppArmor or SELinux is properly configured on your system")
		}
		log.WithField("container_id", c.ID).Info("LSM configuration applied successfully")
	}

	// Apply seccomp profile
	if c.Seccomp != nil {
		log.WithField("container_id", c.ID).Debug("Applying seccomp profile")
		if err := c.Seccomp.Apply(); err != nil {
			log.WithError(err).WithField("container_id", c.ID).Error("Failed to apply seccomp profile")
			return errors.Wrap(errors.ErrSecuritySeccomp, "failed to apply seccomp profile", err).
				WithField("container_id", c.ID).
				WithHint("Ensure seccomp is supported by your kernel")
		}
		log.WithField("container_id", c.ID).Info("Seccomp profile applied successfully")
	}

	// Apply capabilities last (drop/add as needed)
	if c.Capabilities != nil {
		log.WithField("container_id", c.ID).Debug("Applying capabilities configuration")
		if err := c.Capabilities.Apply(); err != nil {
			log.WithError(err).WithField("container_id", c.ID).Error("Failed to apply capabilities")
			return errors.Wrap(errors.ErrSecurityCapabilities, "failed to apply capabilities", err).
				WithField("container_id", c.ID).
				WithHint("Ensure you have the necessary privileges to modify capabilities")
		}
		log.WithField("container_id", c.ID).Info("Capabilities applied successfully")
	}

	log.WithField("container_id", c.ID).Info("All security configurations applied successfully")
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
	log.WithField("container_id", c.ID).Debug("Setting up root filesystem")

	if c.RootFS == "" {
		log.WithField("container_id", c.ID).Debug("No root filesystem specified, skipping setup")
		return nil
	}

	log.WithField("container_id", c.ID).Debugf("Root filesystem path: %s", c.RootFS)

	// Ensure root filesystem exists
	if _, err := os.Stat(c.RootFS); os.IsNotExist(err) {
		log.WithError(err).WithField("container_id", c.ID).Error("Root filesystem does not exist")
		return errors.Wrap(errors.ErrRootFSNotFound, "root filesystem does not exist", err).
			WithField("container_id", c.ID).
			WithField("rootfs_path", c.RootFS).
			WithHint("Ensure the root filesystem path is correct and accessible")
	}

	// Create mount point
	mountPoint := filepath.Join("/tmp", c.ID, "rootfs")
	log.WithField("container_id", c.ID).Debugf("Creating mount point: %s", mountPoint)
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Error("Failed to create mount point")
		return errors.Wrap(errors.ErrRootFSMount, "failed to create mount point", err).
			WithField("container_id", c.ID).
			WithField("mount_point", mountPoint)
	}

	// Bind mount the root filesystem
	log.WithField("container_id", c.ID).Debugf("Bind mounting %s to %s", c.RootFS, mountPoint)
	if err := syscall.Mount(c.RootFS, mountPoint, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Error("Failed to bind mount rootfs")
		// Clean up mount point on error
		os.RemoveAll(mountPoint)
		return errors.Wrap(errors.ErrRootFSMount, "failed to bind mount rootfs", err).
			WithField("container_id", c.ID).
			WithField("rootfs_path", c.RootFS).
			WithField("mount_point", mountPoint).
			WithHint("Ensure you have CAP_SYS_ADMIN capability and the mount point is accessible")
	}

	log.WithField("container_id", c.ID).Info("Root filesystem mounted successfully")
	return nil
}

// Cleanup cleans up container resources
func (c *Container) Cleanup() error {
	log.WithField("container_id", c.ID).Debug("Cleaning up container resources")

	var cleanupErrors []error

	// Unmount root filesystem if it was mounted
	if c.RootFS != "" {
		mountPoint := filepath.Join("/tmp", c.ID, "rootfs")
		log.WithField("container_id", c.ID).Debugf("Unmounting root filesystem: %s", mountPoint)
		if err := syscall.Unmount(mountPoint, 0); err != nil && !os.IsNotExist(err) {
			log.WithError(err).WithField("container_id", c.ID).Warn("Failed to unmount root filesystem")
			cleanupErrors = append(cleanupErrors, errors.Wrap(errors.ErrRootFSUnmount, "failed to unmount rootfs", err))
		}

		// Remove mount point directory
		log.WithField("container_id", c.ID).Debugf("Removing mount point directory: %s", mountPoint)
		if err := os.RemoveAll(filepath.Join("/tmp", c.ID)); err != nil {
			log.WithError(err).WithField("container_id", c.ID).Warn("Failed to remove mount point directory")
			cleanupErrors = append(cleanupErrors, err)
		}
	}

	if len(cleanupErrors) > 0 {
		log.WithField("container_id", c.ID).Warnf("Cleanup completed with %d errors", len(cleanupErrors))
		return errors.New(errors.ErrInternal, fmt.Sprintf("cleanup encountered %d errors", len(cleanupErrors))).
			WithField("container_id", c.ID)
	}

	log.WithField("container_id", c.ID).Info("Cleanup completed successfully")
	return nil
}
