package checkpoint

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
)

// MigrationStrategy defines the migration strategy
type MigrationStrategy string

const (
	// MigrationStrategyPreCopy performs iterative pre-dump before final checkpoint
	MigrationStrategyPreCopy MigrationStrategy = "pre-copy"

	// MigrationStrategyPostCopy performs checkpoint and lazy restore
	MigrationStrategyPostCopy MigrationStrategy = "post-copy"

	// MigrationStrategyStopAndCopy stops container, checkpoint, and restore
	MigrationStrategyStopAndCopy MigrationStrategy = "stop-and-copy"
)

// MigrationOptions holds options for container migration
type MigrationOptions struct {
	// Target host address
	TargetHost string

	// Target port
	TargetPort int

	// Migration strategy
	Strategy MigrationStrategy

	// Pre-dump iterations for pre-copy migration
	PreDumpIterations int

	// Pre-dump interval
	PreDumpInterval time.Duration

	// Page server port for post-copy migration
	PageServerPort int

	// Compression enabled
	Compression bool

	// Encryption enabled
	Encryption bool

	// Bandwidth limit (bytes per second, 0 = unlimited)
	BandwidthLimit int64

	// TCP established connections
	TCPEstablished bool

	// Timeout for migration
	Timeout time.Duration

	// Additional checkpoint options
	CheckpointOpts *CheckpointOptions

	// Additional restore options
	RestoreOpts *RestoreOptions
}

// MigrationState represents the state of a migration
type MigrationState string

const (
	// MigrationStateInitializing means migration is being initialized
	MigrationStateInitializing MigrationState = "initializing"

	// MigrationStatePreDumping means pre-dump iterations are running
	MigrationStatePreDumping MigrationState = "pre-dumping"

	// MigrationStateCheckpointing means final checkpoint is being created
	MigrationStateCheckpointing MigrationState = "checkpointing"

	// MigrationStateTransferring means checkpoint data is being transferred
	MigrationStateTransferring MigrationState = "transferring"

	// MigrationStateRestoring means container is being restored on target
	MigrationStateRestoring MigrationState = "restoring"

	// MigrationStateCompleted means migration completed successfully
	MigrationStateCompleted MigrationState = "completed"

	// MigrationStateFailed means migration failed
	MigrationStateFailed MigrationState = "failed"
)

// MigrationProgress represents migration progress information
type MigrationProgress struct {
	State           MigrationState `json:"state"`
	CurrentIteration int            `json:"current_iteration"`
	TotalIterations  int            `json:"total_iterations"`
	BytesTransferred int64          `json:"bytes_transferred"`
	TotalBytes       int64          `json:"total_bytes"`
	StartTime        time.Time      `json:"start_time"`
	Duration         time.Duration  `json:"duration"`
	Error            string         `json:"error,omitempty"`
}

// Migrator handles container migration operations
type Migrator struct {
	criuManager CRIUInterface
}

// NewMigrator creates a new migrator
func NewMigrator(criuManager CRIUInterface) *Migrator {
	return &Migrator{
		criuManager: criuManager,
	}
}

// Migrate performs container migration to another host
func (m *Migrator) Migrate(ctx context.Context, containerID string, opts *MigrationOptions) error {
	log.WithField("container_id", containerID).
		WithField("target", opts.TargetHost).
		WithField("strategy", opts.Strategy).
		Info("Starting container migration")

	// Set defaults
	if opts.Strategy == "" {
		opts.Strategy = MigrationStrategyPreCopy
	}

	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Minute
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Perform migration based on strategy
	switch opts.Strategy {
	case MigrationStrategyPreCopy:
		return m.migratePreCopy(timeoutCtx, containerID, opts)
	case MigrationStrategyPostCopy:
		return m.migratePostCopy(timeoutCtx, containerID, opts)
	case MigrationStrategyStopAndCopy:
		return m.migrateStopAndCopy(timeoutCtx, containerID, opts)
	default:
		return errors.New(errors.ErrInternal, fmt.Sprintf("unknown migration strategy: %s", opts.Strategy))
	}
}

// migratePreCopy performs pre-copy migration with iterative pre-dumps
func (m *Migrator) migratePreCopy(ctx context.Context, containerID string, opts *MigrationOptions) error {
	log.Info("Starting pre-copy migration")

	// Set defaults
	if opts.PreDumpIterations == 0 {
		opts.PreDumpIterations = 3
	}
	if opts.PreDumpInterval == 0 {
		opts.PreDumpInterval = 5 * time.Second
	}

	// Create temporary directory for checkpoints
	tempDir, err := os.MkdirTemp("", "containr-migration-*")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create temp directory", err)
	}
	defer os.RemoveAll(tempDir)

	// Perform iterative pre-dumps
	var parentPath string
	for i := 0; i < opts.PreDumpIterations; i++ {
		log.WithField("iteration", i+1).Info("Performing pre-dump")

		iterationDir := filepath.Join(tempDir, fmt.Sprintf("pre-dump-%d", i))
		if err := os.MkdirAll(iterationDir, 0700); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create iteration directory", err)
		}

		preDumpOpts := &CheckpointOptions{
			ImagePath:   iterationDir,
			PreDump:     true,
			ParentPath:  parentPath,
			LeaveRunning: true,
		}

		if opts.CheckpointOpts != nil {
			preDumpOpts.TCPEstablished = opts.CheckpointOpts.TCPEstablished
			preDumpOpts.ExternalUnixSockets = opts.CheckpointOpts.ExternalUnixSockets
		}

		if err := m.criuManager.PreDump(ctx, containerID, preDumpOpts); err != nil {
			return errors.Wrap(errors.ErrInternal, "pre-dump failed", err).
				WithField("iteration", i+1)
		}

		// Transfer pre-dump data to target
		if err := m.transferCheckpoint(ctx, iterationDir, opts.TargetHost, opts.TargetPort, opts); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to transfer pre-dump data", err).
				WithField("iteration", i+1)
		}

		parentPath = iterationDir

		// Wait before next iteration
		if i < opts.PreDumpIterations-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(opts.PreDumpInterval):
			}
		}
	}

	// Perform final checkpoint
	log.Info("Performing final checkpoint")
	finalDir := filepath.Join(tempDir, "final")
	if err := os.MkdirAll(finalDir, 0700); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create final directory", err)
	}

	finalOpts := &CheckpointOptions{
		ImagePath:  finalDir,
		ParentPath: parentPath,
		LeaveRunning: false, // Stop container on final checkpoint
	}

	if opts.CheckpointOpts != nil {
		finalOpts.TCPEstablished = opts.CheckpointOpts.TCPEstablished
		finalOpts.ExternalUnixSockets = opts.CheckpointOpts.ExternalUnixSockets
		finalOpts.FileLocks = opts.CheckpointOpts.FileLocks
		finalOpts.ShellJob = opts.CheckpointOpts.ShellJob
	}

	if err := m.criuManager.Dump(ctx, containerID, finalOpts); err != nil {
		return errors.Wrap(errors.ErrInternal, "final checkpoint failed", err)
	}

	// Transfer final checkpoint to target
	if err := m.transferCheckpoint(ctx, finalDir, opts.TargetHost, opts.TargetPort, opts); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to transfer final checkpoint", err)
	}

	log.Info("Pre-copy migration completed successfully")
	return nil
}

// migratePostCopy performs post-copy migration with lazy pages
func (m *Migrator) migratePostCopy(ctx context.Context, containerID string, opts *MigrationOptions) error {
	log.Info("Starting post-copy migration")

	// Post-copy migration requires CRIU page server which is not yet implemented
	// in this version. Use pre-copy or stop-and-copy instead.
	return errors.New(errors.ErrInternal, "post-copy migration is not yet implemented").
		WithHint("Use pre-copy or stop-and-copy migration strategy instead")
}

// migrateStopAndCopy performs simple stop-and-copy migration
func (m *Migrator) migrateStopAndCopy(ctx context.Context, containerID string, opts *MigrationOptions) error {
	log.Info("Starting stop-and-copy migration")

	// Create temporary directory for checkpoint
	tempDir, err := os.MkdirTemp("", "containr-migration-*")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create temp directory", err)
	}
	defer os.RemoveAll(tempDir)

	checkpointDir := filepath.Join(tempDir, "checkpoint")
	if err := os.MkdirAll(checkpointDir, 0700); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create checkpoint directory", err)
	}

	// Perform checkpoint (container will be stopped)
	log.Info("Checkpointing container")

	checkpointOpts := &CheckpointOptions{
		ImagePath:    checkpointDir,
		LeaveRunning: false,
	}

	if opts.CheckpointOpts != nil {
		checkpointOpts.TCPEstablished = opts.CheckpointOpts.TCPEstablished
		checkpointOpts.ExternalUnixSockets = opts.CheckpointOpts.ExternalUnixSockets
		checkpointOpts.FileLocks = opts.CheckpointOpts.FileLocks
		checkpointOpts.ShellJob = opts.CheckpointOpts.ShellJob
	}

	if err := m.criuManager.Dump(ctx, containerID, checkpointOpts); err != nil {
		return errors.Wrap(errors.ErrInternal, "checkpoint failed", err)
	}

	// Transfer checkpoint to target
	if err := m.transferCheckpoint(ctx, checkpointDir, opts.TargetHost, opts.TargetPort, opts); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to transfer checkpoint", err)
	}

	log.Info("Stop-and-copy migration completed successfully")
	return nil
}

// transferCheckpoint transfers checkpoint data to target host
func (m *Migrator) transferCheckpoint(ctx context.Context, checkpointDir string, targetHost string, targetPort int, opts *MigrationOptions) error {
	log.WithField("target", targetHost).
		WithField("port", targetPort).
		Info("Transferring checkpoint data")

	// Connect to target
	addr := fmt.Sprintf("%s:%d", targetHost, targetPort)
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to connect to target", err).
			WithField("address", addr)
	}
	defer conn.Close()

	// Create a rate-limited writer if bandwidth limit is set
	var writer io.Writer = conn
	if opts.BandwidthLimit > 0 {
		writer = newRateLimitedWriter(conn, opts.BandwidthLimit)
	}

	// Transfer files
	err = m.transferDirectory(ctx, checkpointDir, writer)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to transfer checkpoint files", err)
	}

	log.Info("Checkpoint data transferred successfully")
	return nil
}

// transferDirectory transfers a directory over a writer
func (m *Migrator) transferDirectory(ctx context.Context, dirPath string, writer io.Writer) error {
	// Walk through directory and transfer files
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Write header: path length, path, data length, data
		header := fmt.Sprintf("%d:%s:%d:", len(relPath), relPath, len(data))
		if _, err := writer.Write([]byte(header)); err != nil {
			return err
		}

		// Write data
		if _, err := writer.Write(data); err != nil {
			return err
		}

		return nil
	})
}

// rateLimitedWriter implements rate-limited writing
type rateLimitedWriter struct {
	writer      io.Writer
	limit       int64
	bucket      int64
	lastRefill  time.Time
	mu          chan struct{}
}

// newRateLimitedWriter creates a new rate-limited writer
func newRateLimitedWriter(writer io.Writer, limitBytesPerSec int64) *rateLimitedWriter {
	return &rateLimitedWriter{
		writer:     writer,
		limit:      limitBytesPerSec,
		bucket:     limitBytesPerSec,
		lastRefill: time.Now(),
		mu:         make(chan struct{}, 1),
	}
}

// Write writes data with rate limiting
func (w *rateLimitedWriter) Write(p []byte) (n int, err error) {
	w.mu <- struct{}{}
	defer func() { <-w.mu }()

	// Refill bucket based on elapsed time
	now := time.Now()
	elapsed := now.Sub(w.lastRefill)
	refill := int64(elapsed.Seconds() * float64(w.limit))
	w.bucket += refill
	if w.bucket > w.limit {
		w.bucket = w.limit
	}
	w.lastRefill = now

	// Write in chunks based on available bucket
	written := 0
	for written < len(p) {
		// Wait if bucket is empty
		if w.bucket <= 0 {
			time.Sleep(100 * time.Millisecond)
			now := time.Now()
			elapsed := now.Sub(w.lastRefill)
			refill := int64(elapsed.Seconds() * float64(w.limit))
			w.bucket += refill
			if w.bucket > w.limit {
				w.bucket = w.limit
			}
			w.lastRefill = now
			continue
		}

		// Write chunk
		chunkSize := int64(len(p) - written)
		if chunkSize > w.bucket {
			chunkSize = w.bucket
		}

		n, err := w.writer.Write(p[written : written+int(chunkSize)])
		if err != nil {
			return written, err
		}

		written += n
		w.bucket -= int64(n)
	}

	return written, nil
}

// GetProgress returns current migration progress
func (m *Migrator) GetProgress(migrationID string) (*MigrationProgress, error) {
	// This would be implemented with actual progress tracking
	// For now, return a placeholder
	return &MigrationProgress{
		State:     MigrationStateCompleted,
		StartTime: time.Now(),
	}, nil
}
