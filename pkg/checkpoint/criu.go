package checkpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
)

// CRIUManager manages CRIU (Checkpoint/Restore In Userspace) operations
type CRIUManager struct {
	criuPath string
	version  string
}

// CRIUVersion represents CRIU version information
type CRIUVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// NewCRIUManager creates a new CRIU manager
func NewCRIUManager(criuPath string) (*CRIUManager, error) {
	if criuPath == "" {
		criuPath = "/usr/sbin/criu"
	}

	// Verify CRIU exists
	if _, err := os.Stat(criuPath); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "CRIU binary not found", err).
			WithField("path", criuPath).
			WithHint("Install CRIU: apt-get install criu (Debian/Ubuntu) or yum install criu (RHEL/CentOS)")
	}

	mgr := &CRIUManager{
		criuPath: criuPath,
	}

	// Check CRIU version
	version, err := mgr.checkVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to check CRIU version: %w", err)
	}

	mgr.version = version
	log.Debugf("CRIU version: %s", version)

	return mgr, nil
}

// checkVersion checks the CRIU version
func (c *CRIUManager) checkVersion() (string, error) {
	cmd := exec.Command(c.criuPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// Check verifies CRIU functionality
func (c *CRIUManager) Check() error {
	cmd := exec.Command(c.criuPath, "check")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "CRIU check failed", err).
			WithField("output", string(output)).
			WithHint("CRIU may not be properly configured or kernel support may be missing")
	}

	log.Debug("CRIU check passed")
	return nil
}

// Dump creates a checkpoint of a container process
func (c *CRIUManager) Dump(ctx context.Context, containerID string, opts *CheckpointOptions) error {
	log.WithField("container_id", containerID).Info("Starting CRIU dump")

	// Build CRIU command arguments
	args := []string{"dump"}

	// Image directory
	if opts.ImagePath != "" {
		args = append(args, "--images-dir", opts.ImagePath)
	}

	// Work directory
	if opts.WorkDir != "" {
		args = append(args, "--work-dir", opts.WorkDir)
	} else if opts.ImagePath != "" {
		args = append(args, "--work-dir", opts.ImagePath)
	}

	// Leave running option
	if opts.LeaveRunning {
		args = append(args, "--leave-running")
	}

	// TCP established connections
	if opts.TCPEstablished {
		args = append(args, "--tcp-established")
	}

	// External UNIX sockets
	if opts.ExternalUnixSockets {
		args = append(args, "--ext-unix-sk")
	}

	// Shell job
	if opts.ShellJob {
		args = append(args, "--shell-job")
	}

	// File locks
	if opts.FileLocks {
		args = append(args, "--file-locks")
	}

	// Pre-dump for iterative checkpoint
	if opts.PreDump {
		args = append(args, "--pre-dump")

		// Parent checkpoint path for iterative dumps
		if opts.ParentPath != "" {
			args = append(args, "--prev-images-dir", opts.ParentPath)
		}
	}

	// Log file
	logFile := filepath.Join(opts.ImagePath, "dump.log")
	args = append(args, "--log-file", logFile)

	// Verbose logging
	args = append(args, "-vvv")

	// Tree ID (use container ID as tree)
	args = append(args, "--tree", containerID)

	// Add custom CRIU options
	for key, value := range opts.CRIUOpts {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s", key), value)
		} else {
			args = append(args, fmt.Sprintf("--%s", key))
		}
	}

	log.Debugf("CRIU dump command: %s %v", c.criuPath, args)

	// Execute CRIU dump
	cmd := exec.CommandContext(ctx, c.criuPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try to read log file for more details
		logContent := ""
		if logData, readErr := os.ReadFile(logFile); readErr == nil {
			logContent = string(logData)
		}

		return errors.Wrap(errors.ErrInternal, "CRIU dump failed", err).
			WithField("container_id", containerID).
			WithField("output", string(output)).
			WithField("log", logContent).
			WithHint("Check CRIU dump log for details")
	}

	log.WithField("container_id", containerID).
		WithField("output", string(output)).
		Info("CRIU dump completed successfully")

	return nil
}

// Restore restores a container from a checkpoint
func (c *CRIUManager) Restore(ctx context.Context, containerID string, opts *RestoreOptions) error {
	log.WithField("container_id", containerID).Info("Starting CRIU restore")

	// Build CRIU command arguments
	args := []string{"restore"}

	// Image directory
	if opts.ImagePath != "" {
		args = append(args, "--images-dir", opts.ImagePath)
	}

	// Work directory
	if opts.WorkDir != "" {
		args = append(args, "--work-dir", opts.WorkDir)
	} else if opts.ImagePath != "" {
		args = append(args, "--work-dir", opts.ImagePath)
	}

	// Detach option
	if opts.Detach {
		args = append(args, "--restore-detached")
	}

	// TCP established connections
	if opts.TCPEstablished {
		args = append(args, "--tcp-established")
	}

	// External UNIX sockets
	if opts.ExternalUnixSockets {
		args = append(args, "--ext-unix-sk")
	}

	// Shell job
	if opts.ShellJob {
		args = append(args, "--shell-job")
	}

	// File locks
	if opts.FileLocks {
		args = append(args, "--file-locks")
	}

	// Log file
	logFile := filepath.Join(opts.ImagePath, "restore.log")
	args = append(args, "--log-file", logFile)

	// Verbose logging
	args = append(args, "-vvv")

	// Add custom CRIU options
	for key, value := range opts.CRIUOpts {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s", key), value)
		} else {
			args = append(args, fmt.Sprintf("--%s", key))
		}
	}

	log.Debugf("CRIU restore command: %s %v", c.criuPath, args)

	// Execute CRIU restore
	cmd := exec.CommandContext(ctx, c.criuPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try to read log file for more details
		logContent := ""
		if logData, readErr := os.ReadFile(logFile); readErr == nil {
			logContent = string(logData)
		}

		return errors.Wrap(errors.ErrInternal, "CRIU restore failed", err).
			WithField("container_id", containerID).
			WithField("output", string(output)).
			WithField("log", logContent).
			WithHint("Check CRIU restore log for details")
	}

	log.WithField("container_id", containerID).
		WithField("output", string(output)).
		Info("CRIU restore completed successfully")

	return nil
}

// PreDump performs a pre-dump (iterative checkpoint)
func (c *CRIUManager) PreDump(ctx context.Context, containerID string, opts *CheckpointOptions) error {
	log.WithField("container_id", containerID).Info("Starting CRIU pre-dump")

	// Create a copy of options and enable pre-dump
	preDumpOpts := *opts
	preDumpOpts.PreDump = true
	preDumpOpts.LeaveRunning = true

	return c.Dump(ctx, containerID, &preDumpOpts)
}

// PageServerStart starts a CRIU page server for live migration
func (c *CRIUManager) PageServerStart(ctx context.Context, port int, imageDir string) (*exec.Cmd, error) {
	log.WithField("port", port).Info("Starting CRIU page server")

	args := []string{
		"page-server",
		"--images-dir", imageDir,
		"--port", strconv.Itoa(port),
		"-vvv",
	}

	cmd := exec.CommandContext(ctx, c.criuPath, args...)

	// Start the page server
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to start CRIU page server", err).
			WithField("port", port)
	}

	log.WithField("port", port).Info("CRIU page server started")
	return cmd, nil
}

// LazyPages starts lazy pages daemon for post-copy migration
func (c *CRIUManager) LazyPages(ctx context.Context, imageDir string, pageServerAddr string) (*exec.Cmd, error) {
	log.WithField("page_server", pageServerAddr).Info("Starting CRIU lazy pages")

	args := []string{
		"lazy-pages",
		"--images-dir", imageDir,
		"--page-server", pageServerAddr,
		"-vvv",
	}

	cmd := exec.CommandContext(ctx, c.criuPath, args...)

	// Start lazy pages daemon
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to start CRIU lazy pages", err).
			WithField("page_server", pageServerAddr)
	}

	log.WithField("page_server", pageServerAddr).Info("CRIU lazy pages started")
	return cmd, nil
}

// GetPID extracts the PID from checkpoint images
func (c *CRIUManager) GetPID(imageDir string) (int, error) {
	// Read inventory.img to get root PID
	inventoryPath := filepath.Join(imageDir, "inventory.img")

	data, err := os.ReadFile(inventoryPath)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInternal, "failed to read inventory file", err).
			WithField("path", inventoryPath)
	}

	var inventory map[string]interface{}
	if err := json.Unmarshal(data, &inventory); err != nil {
		return 0, errors.Wrap(errors.ErrInternal, "failed to parse inventory file", err)
	}

	// Extract PID
	if rootPID, ok := inventory["root_pid"].(float64); ok {
		return int(rootPID), nil
	}

	return 0, errors.New(errors.ErrInternal, "root_pid not found in inventory")
}

// GetVersion returns the CRIU version
func (c *CRIUManager) GetVersion() string {
	return c.version
}

// SupportsFeature checks if CRIU supports a specific feature
func (c *CRIUManager) SupportsFeature(feature string) (bool, error) {
	cmd := exec.Command(c.criuPath, "check", "--feature", feature)
	err := cmd.Run()
	return err == nil, nil
}

// GetStats reads checkpoint/restore statistics
func (c *CRIUManager) GetStats(imageDir string, statType string) (map[string]interface{}, error) {
	statsFile := filepath.Join(imageDir, fmt.Sprintf("stats-%s", statType))

	data, err := os.ReadFile(statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Stats file doesn't exist
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to read stats file", err).
			WithField("path", statsFile)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to parse stats file", err)
	}

	return stats, nil
}
