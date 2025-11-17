package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// RootFS manages container root filesystems
type RootFS struct {
	Path       string
	MountPoint string
	Layers     []string // For overlay filesystem
}

// Config holds rootfs configuration
type Config struct {
	Path       string
	MountPoint string
	UseOverlay bool
	Layers     []string
}

// New creates a new RootFS instance
func New(config *Config) *RootFS {
	return &RootFS{
		Path:       config.Path,
		MountPoint: config.MountPoint,
		Layers:     config.Layers,
	}
}

// Setup prepares the root filesystem
func (r *RootFS) Setup() error {
	// Create mount point if it doesn't exist
	if err := os.MkdirAll(r.MountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	if len(r.Layers) > 0 {
		// Use overlay filesystem
		return r.setupOverlay()
	}

	// Simple bind mount
	return r.setupBind()
}

// setupBind sets up a simple bind mount
func (r *RootFS) setupBind() error {
	if err := syscall.Mount(r.Path, r.MountPoint, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind mount: %w", err)
	}
	return nil
}

// setupOverlay sets up an overlay filesystem with multiple layers
func (r *RootFS) setupOverlay() error {
	// Create overlay directories
	overlayDir := filepath.Join("/tmp", "overlay")
	upperDir := filepath.Join(overlayDir, "upper")
	workDir := filepath.Join(overlayDir, "work")

	for _, dir := range []string{upperDir, workDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create overlay directory: %w", err)
		}
	}

	// Build overlay mount options
	// Format: lowerdir=layer1:layer2:...,upperdir=upper,workdir=work
	lowerDirs := ""
	for i, layer := range r.Layers {
		if i > 0 {
			lowerDirs += ":"
		}
		lowerDirs += layer
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDirs, upperDir, workDir)

	// Mount overlay filesystem
	if err := syscall.Mount("overlay", r.MountPoint, "overlay", 0, options); err != nil {
		return fmt.Errorf("failed to mount overlay: %w", err)
	}

	return nil
}

// Teardown cleans up the root filesystem
func (r *RootFS) Teardown() error {
	if err := syscall.Unmount(r.MountPoint, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount: %w", err)
	}
	return nil
}

// PivotRoot changes the root filesystem of the calling process
func (r *RootFS) PivotRoot() error {
	// Create old root directory
	oldRoot := filepath.Join(r.MountPoint, ".pivot_root")
	if err := os.MkdirAll(oldRoot, 0755); err != nil {
		return fmt.Errorf("failed to create old root directory: %w", err)
	}

	// Pivot root
	if err := syscall.PivotRoot(r.MountPoint, oldRoot); err != nil {
		return fmt.Errorf("failed to pivot root: %w", err)
	}

	// Change to new root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("failed to change to new root: %w", err)
	}

	// Unmount old root
	oldRoot = "/.pivot_root"
	if err := syscall.Unmount(oldRoot, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount old root: %w", err)
	}

	// Remove old root directory
	if err := os.Remove(oldRoot); err != nil {
		return fmt.Errorf("failed to remove old root: %w", err)
	}

	return nil
}

// MountProc mounts the /proc filesystem
func MountProc() error {
	procPath := "/proc"

	// Check if already mounted
	if _, err := os.Stat(filepath.Join(procPath, "self")); err == nil {
		return nil
	}

	// Create /proc if it doesn't exist
	if err := os.MkdirAll(procPath, 0755); err != nil {
		return fmt.Errorf("failed to create /proc: %w", err)
	}

	// Mount proc
	if err := syscall.Mount("proc", procPath, "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /proc: %w", err)
	}

	return nil
}

// MountSys mounts the /sys filesystem
func MountSys() error {
	sysPath := "/sys"

	// Create /sys if it doesn't exist
	if err := os.MkdirAll(sysPath, 0755); err != nil {
		return fmt.Errorf("failed to create /sys: %w", err)
	}

	// Mount sysfs
	if err := syscall.Mount("sysfs", sysPath, "sysfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /sys: %w", err)
	}

	return nil
}

// MountDev mounts essential device nodes
func MountDev() error {
	devPath := "/dev"

	// Create /dev if it doesn't exist
	if err := os.MkdirAll(devPath, 0755); err != nil {
		return fmt.Errorf("failed to create /dev: %w", err)
	}

	// Mount tmpfs on /dev
	if err := syscall.Mount("tmpfs", devPath, "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("failed to mount /dev: %w", err)
	}

	// Create essential device nodes
	devices := []struct {
		path string
		mode uint32
		dev  int
	}{
		{"/dev/null", syscall.S_IFCHR | 0666, makedev(1, 3)},
		{"/dev/zero", syscall.S_IFCHR | 0666, makedev(1, 5)},
		{"/dev/random", syscall.S_IFCHR | 0666, makedev(1, 8)},
		{"/dev/urandom", syscall.S_IFCHR | 0666, makedev(1, 9)},
	}

	for _, device := range devices {
		if err := syscall.Mknod(device.path, device.mode, device.dev); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create device %s: %w", device.path, err)
		}
	}

	return nil
}

// makedev creates a device number
func makedev(major, minor int) int {
	return (major << 8) | minor
}
