package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("cgroup")

// Cgroup represents a control group for resource management
type Cgroup struct {
	Name   string
	Parent string
}

// Config holds cgroup resource limits
type Config struct {
	Name        string
	MemoryLimit int64 // Memory limit in bytes (0 = no limit)
	CPUShares   int64 // CPU shares (relative weight, default 1024)
	PIDLimit    int64 // Maximum number of processes (0 = no limit)
}

// New creates a new cgroup
func New(config *Config) (*Cgroup, error) {
	log.WithField("cgroup_name", config.Name).Debug("Creating new cgroup")

	cg := &Cgroup{
		Name:   config.Name,
		Parent: "/sys/fs/cgroup",
	}

	log.WithFields(map[string]interface{}{
		"cgroup_name": config.Name,
		"memory_limit": config.MemoryLimit,
		"cpu_shares": config.CPUShares,
		"pid_limit": config.PIDLimit,
	}).Debug("Cgroup configuration")

	// Create cgroup directories for each controller
	controllers := []string{"memory", "cpu", "pids"}
	for _, controller := range controllers {
		cgroupPath := filepath.Join(cg.Parent, controller, cg.Name)
		log.WithField("cgroup_name", config.Name).Debugf("Creating cgroup directory: %s", cgroupPath)
		if err := os.MkdirAll(cgroupPath, 0755); err != nil {
			log.WithError(err).WithField("cgroup_name", config.Name).Error("Failed to create cgroup directory")
			// Clean up any created directories
			cg.Remove()
			return nil, errors.Wrap(errors.ErrCgroupCreate, "failed to create cgroup directory", err).
				WithField("cgroup_name", config.Name).
				WithField("cgroup_path", cgroupPath).
				WithHint("Ensure you have root privileges and cgroups v1 or v2 is mounted at /sys/fs/cgroup")
		}
	}

	log.WithField("cgroup_name", config.Name).Debug("Applying resource limits")

	// Apply resource limits
	if err := cg.applyLimits(config); err != nil {
		log.WithError(err).WithField("cgroup_name", config.Name).Error("Failed to apply limits")
		// Clean up created cgroup
		cg.Remove()
		return nil, errors.Wrap(errors.ErrCgroupApplyLimits, "failed to apply resource limits", err).
			WithField("cgroup_name", config.Name)
	}

	log.WithField("cgroup_name", config.Name).Info("Cgroup created successfully")
	return cg, nil
}

// applyLimits applies resource limits to the cgroup
func (c *Cgroup) applyLimits(config *Config) error {
	// Set memory limit
	if config.MemoryLimit > 0 {
		memoryPath := filepath.Join(c.Parent, "memory", c.Name, "memory.limit_in_bytes")
		if err := writeFile(memoryPath, strconv.FormatInt(config.MemoryLimit, 10)); err != nil {
			// Try cgroup v2 path
			memoryPath = filepath.Join(c.Parent, "memory", c.Name, "memory.max")
			if err := writeFile(memoryPath, strconv.FormatInt(config.MemoryLimit, 10)); err != nil {
				return fmt.Errorf("failed to set memory limit: %w", err)
			}
		}
	}

	// Set CPU shares
	if config.CPUShares > 0 {
		cpuPath := filepath.Join(c.Parent, "cpu", c.Name, "cpu.shares")
		if err := writeFile(cpuPath, strconv.FormatInt(config.CPUShares, 10)); err != nil {
			// Try cgroup v2 path
			cpuPath = filepath.Join(c.Parent, "cpu", c.Name, "cpu.weight")
			// cgroup v2 uses weight (1-10000), convert from shares
			weight := (config.CPUShares * 10000) / 1024
			if err := writeFile(cpuPath, strconv.FormatInt(weight, 10)); err != nil {
				return fmt.Errorf("failed to set CPU shares: %w", err)
			}
		}
	}

	// Set PID limit
	if config.PIDLimit > 0 {
		pidsPath := filepath.Join(c.Parent, "pids", c.Name, "pids.max")
		if err := writeFile(pidsPath, strconv.FormatInt(config.PIDLimit, 10)); err != nil {
			return fmt.Errorf("failed to set PID limit: %w", err)
		}
	}

	return nil
}

// AddProcess adds a process to the cgroup
func (c *Cgroup) AddProcess(pid int) error {
	log.WithFields(map[string]interface{}{
		"cgroup_name": c.Name,
		"pid": pid,
	}).Debug("Adding process to cgroup")

	controllers := []string{"memory", "cpu", "pids"}
	for _, controller := range controllers {
		// Try cgroup v1
		procsPath := filepath.Join(c.Parent, controller, c.Name, "cgroup.procs")
		if err := writeFile(procsPath, strconv.Itoa(pid)); err != nil {
			// Try cgroup v2
			procsPath = filepath.Join(c.Parent, controller, c.Name, "cgroup.procs")
			if err := writeFile(procsPath, strconv.Itoa(pid)); err != nil {
				log.WithError(err).WithFields(map[string]interface{}{
					"cgroup_name": c.Name,
					"pid": pid,
					"controller": controller,
				}).Error("Failed to add process to cgroup")
				return errors.Wrap(errors.ErrCgroupAddProcess, "failed to add process to cgroup", err).
					WithField("cgroup_name", c.Name).
					WithField("pid", pid).
					WithHint("Ensure the process exists and you have root privileges")
			}
		}
	}

	log.WithFields(map[string]interface{}{
		"cgroup_name": c.Name,
		"pid": pid,
	}).Info("Process added to cgroup successfully")
	return nil
}

// Remove removes the cgroup
func (c *Cgroup) Remove() error {
	log.WithField("cgroup_name", c.Name).Debug("Removing cgroup")

	var removeErrors []error
	controllers := []string{"memory", "cpu", "pids"}
	for _, controller := range controllers {
		cgroupPath := filepath.Join(c.Parent, controller, c.Name)
		log.WithField("cgroup_name", c.Name).Debugf("Removing cgroup path: %s", cgroupPath)
		if err := os.Remove(cgroupPath); err != nil && !os.IsNotExist(err) {
			log.WithError(err).WithField("cgroup_name", c.Name).Warn("Failed to remove cgroup path")
			removeErrors = append(removeErrors, err)
		}
	}

	if len(removeErrors) > 0 {
		log.WithField("cgroup_name", c.Name).Warnf("Cgroup removal completed with %d errors", len(removeErrors))
		return errors.Wrap(errors.ErrCgroupRemove, "failed to remove cgroup completely", removeErrors[0]).
			WithField("cgroup_name", c.Name).
			WithHint("Some cgroup directories may not have been removed. Check if processes are still running in the cgroup.")
	}

	log.WithField("cgroup_name", c.Name).Info("Cgroup removed successfully")
	return nil
}

// GetStats returns resource usage statistics
func (c *Cgroup) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Read memory usage
	memoryUsagePath := filepath.Join(c.Parent, "memory", c.Name, "memory.usage_in_bytes")
	if data, err := os.ReadFile(memoryUsagePath); err == nil {
		if usage, err := strconv.ParseInt(string(data), 10, 64); err == nil {
			stats.MemoryUsage = usage
		}
	} else {
		// Try cgroup v2
		memoryUsagePath = filepath.Join(c.Parent, "memory", c.Name, "memory.current")
		if data, err := os.ReadFile(memoryUsagePath); err == nil {
			if usage, err := strconv.ParseInt(string(data), 10, 64); err == nil {
				stats.MemoryUsage = usage
			}
		}
	}

	return stats, nil
}

// Stats holds cgroup statistics
type Stats struct {
	MemoryUsage int64
	CPUUsage    int64
	PIDCount    int64
}

// writeFile writes data to a file
func writeFile(path, data string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	return err
}

// AddCurrentProcess adds the current process to the cgroup
func (c *Cgroup) AddCurrentProcess() error {
	return c.AddProcess(syscall.Getpid())
}
