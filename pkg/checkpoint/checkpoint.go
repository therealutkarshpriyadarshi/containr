package checkpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("checkpoint")

// Checkpointer defines the interface for container checkpointing
type Checkpointer interface {
	// Checkpoint creates a checkpoint of a running container
	Checkpoint(ctx context.Context, containerID string, opts *CheckpointOptions) (*Checkpoint, error)

	// Restore restores a container from a checkpoint
	Restore(ctx context.Context, checkpointID string, opts *RestoreOptions) error

	// List returns all checkpoints
	List(ctx context.Context) ([]*Checkpoint, error)

	// Get retrieves a specific checkpoint
	Get(ctx context.Context, checkpointID string) (*Checkpoint, error)

	// Delete removes a checkpoint
	Delete(ctx context.Context, checkpointID string) error

	// Migrate performs live migration to another host
	Migrate(ctx context.Context, containerID string, opts *MigrationOptions) error
}

// CheckpointOptions holds options for checkpoint creation
type CheckpointOptions struct {
	// Name is the checkpoint name
	Name string

	// Leave container running after checkpoint
	LeaveRunning bool

	// Include TCP connections
	TCPEstablished bool

	// Include external UNIX sockets
	ExternalUnixSockets bool

	// Include shell jobs
	ShellJob bool

	// File locks handling
	FileLocks bool

	// Image path for checkpoint files
	ImagePath string

	// Work directory for CRIU
	WorkDir string

	// Pre-dump - iterative checkpoint
	PreDump bool

	// Parent checkpoint path for iterative checkpoints
	ParentPath string

	// Additional CRIU options
	CRIUOpts map[string]string
}

// RestoreOptions holds options for checkpoint restoration
type RestoreOptions struct {
	// Container name for restored container
	Name string

	// Detach after restore
	Detach bool

	// TCP connections handling
	TCPEstablished bool

	// External UNIX sockets
	ExternalUnixSockets bool

	// Shell job restoration
	ShellJob bool

	// File locks handling
	FileLocks bool

	// Image path
	ImagePath string

	// Work directory
	WorkDir string

	// Additional CRIU options
	CRIUOpts map[string]string
}

// Checkpoint represents a container checkpoint
type Checkpoint struct {
	ID           string                 `json:"id"`
	ContainerID  string                 `json:"container_id"`
	Name         string                 `json:"name"`
	Created      time.Time              `json:"created"`
	State        CheckpointState        `json:"state"`
	Size         int64                  `json:"size"`
	ImagePath    string                 `json:"image_path"`
	Options      *CheckpointOptions     `json:"options"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ParentID     string                 `json:"parent_id,omitempty"`
	PreDumpCount int                    `json:"pre_dump_count"`
}

// CheckpointState represents the state of a checkpoint
type CheckpointState string

const (
	// CheckpointStateCreating means checkpoint is being created
	CheckpointStateCreating CheckpointState = "creating"

	// CheckpointStateReady means checkpoint is ready for use
	CheckpointStateReady CheckpointState = "ready"

	// CheckpointStateFailed means checkpoint creation failed
	CheckpointStateFailed CheckpointState = "failed"

	// CheckpointStateMigrating means checkpoint is being migrated
	CheckpointStateMigrating CheckpointState = "migrating"
)

// CRIUInterface defines the interface for CRIU operations
type CRIUInterface interface {
	Dump(ctx context.Context, containerID string, opts *CheckpointOptions) error
	Restore(ctx context.Context, containerID string, opts *RestoreOptions) error
	PreDump(ctx context.Context, containerID string, opts *CheckpointOptions) error
	Check() error
	GetVersion() string
}

// Manager manages container checkpoints
type Manager struct {
	criuManager  CRIUInterface
	stateStore   *StateStore
	checkpoints  map[string]*Checkpoint
	mu           sync.RWMutex
	storePath    string
	criuPath     string
}

// Config holds checkpoint manager configuration
type Config struct {
	// StorePath is the path to store checkpoints
	StorePath string

	// CRIUPath is the path to CRIU binary
	CRIUPath string

	// EnablePreDump enables pre-dump optimization
	EnablePreDump bool

	// DefaultOptions are default checkpoint options
	DefaultOptions *CheckpointOptions
}

// NewManager creates a new checkpoint manager
func NewManager(config *Config) (*Manager, error) {
	if config.StorePath == "" {
		config.StorePath = "/var/lib/containr/checkpoints"
	}

	if config.CRIUPath == "" {
		config.CRIUPath = "/usr/sbin/criu"
	}

	// Verify CRIU is available
	if _, err := os.Stat(config.CRIUPath); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "CRIU binary not found", err).
			WithField("path", config.CRIUPath).
			WithHint("Install CRIU: https://criu.org/Installation")
	}

	// Create store directory
	if err := os.MkdirAll(config.StorePath, 0700); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create checkpoint store directory", err).
			WithField("path", config.StorePath)
	}

	criuManager, err := NewCRIUManager(config.CRIUPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRIU manager: %w", err)
	}

	stateStore, err := NewStateStore(filepath.Join(config.StorePath, "state"))
	if err != nil {
		return nil, fmt.Errorf("failed to create state store: %w", err)
	}

	mgr := &Manager{
		criuManager: criuManager,
		stateStore:  stateStore,
		checkpoints: make(map[string]*Checkpoint),
		storePath:   config.StorePath,
		criuPath:    config.CRIUPath,
	}

	// Load existing checkpoints
	if err := mgr.loadCheckpoints(); err != nil {
		log.WithError(err).Warn("Failed to load checkpoints, starting with empty cache")
	}

	return mgr, nil
}

// Checkpoint creates a checkpoint of a running container
func (m *Manager) Checkpoint(ctx context.Context, containerID string, opts *CheckpointOptions) (*Checkpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.WithField("container_id", containerID).Info("Creating checkpoint")

	// Generate checkpoint ID
	checkpointID := generateCheckpointID(containerID)

	// Set default options
	if opts == nil {
		opts = &CheckpointOptions{}
	}

	// Set checkpoint name if not provided
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("checkpoint-%s", checkpointID[:8])
	}

	// Set image path
	if opts.ImagePath == "" {
		opts.ImagePath = filepath.Join(m.storePath, checkpointID, "images")
	}

	// Create checkpoint directory
	if err := os.MkdirAll(opts.ImagePath, 0700); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create checkpoint directory", err).
			WithField("container_id", containerID)
	}

	// Create checkpoint object
	checkpoint := &Checkpoint{
		ID:          checkpointID,
		ContainerID: containerID,
		Name:        opts.Name,
		Created:     time.Now(),
		State:       CheckpointStateCreating,
		ImagePath:   opts.ImagePath,
		Options:     opts,
		Metadata:    make(map[string]interface{}),
	}

	// Save checkpoint state
	m.checkpoints[checkpointID] = checkpoint
	if err := m.saveCheckpoint(checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint state: %w", err)
	}

	// Perform checkpoint using CRIU
	if err := m.criuManager.Dump(ctx, containerID, opts); err != nil {
		checkpoint.State = CheckpointStateFailed
		m.saveCheckpoint(checkpoint)
		return nil, errors.Wrap(errors.ErrInternal, "CRIU dump failed", err).
			WithField("container_id", containerID).
			WithField("checkpoint_id", checkpointID)
	}

	// Calculate checkpoint size
	size, err := calculateDirSize(opts.ImagePath)
	if err != nil {
		log.WithError(err).Warn("Failed to calculate checkpoint size")
	} else {
		checkpoint.Size = size
	}

	// Update checkpoint state
	checkpoint.State = CheckpointStateReady
	if err := m.saveCheckpoint(checkpoint); err != nil {
		return nil, fmt.Errorf("failed to update checkpoint state: %w", err)
	}

	log.WithField("checkpoint_id", checkpointID).
		WithField("container_id", containerID).
		Info("Checkpoint created successfully")

	return checkpoint, nil
}

// Restore restores a container from a checkpoint
func (m *Manager) Restore(ctx context.Context, checkpointID string, opts *RestoreOptions) error {
	m.mu.RLock()
	checkpoint, exists := m.checkpoints[checkpointID]
	m.mu.RUnlock()

	if !exists {
		return errors.New(errors.ErrInternal, "checkpoint not found").
			WithField("checkpoint_id", checkpointID)
	}

	if checkpoint.State != CheckpointStateReady {
		return errors.New(errors.ErrInternal, fmt.Sprintf("checkpoint not ready: %s", checkpoint.State)).
			WithField("checkpoint_id", checkpointID)
	}

	log.WithField("checkpoint_id", checkpointID).Info("Restoring checkpoint")

	// Set default options
	if opts == nil {
		opts = &RestoreOptions{}
	}

	// Use checkpoint image path if not provided
	if opts.ImagePath == "" {
		opts.ImagePath = checkpoint.ImagePath
	}

	// Perform restore using CRIU
	if err := m.criuManager.Restore(ctx, checkpoint.ContainerID, opts); err != nil {
		return errors.Wrap(errors.ErrInternal, "CRIU restore failed", err).
			WithField("checkpoint_id", checkpointID)
	}

	log.WithField("checkpoint_id", checkpointID).Info("Checkpoint restored successfully")
	return nil
}

// List returns all checkpoints
func (m *Manager) List(ctx context.Context) ([]*Checkpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checkpoints := make([]*Checkpoint, 0, len(m.checkpoints))
	for _, cp := range m.checkpoints {
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, nil
}

// Get retrieves a specific checkpoint
func (m *Manager) Get(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checkpoint, exists := m.checkpoints[checkpointID]
	if !exists {
		return nil, errors.New(errors.ErrInternal, "checkpoint not found").
			WithField("checkpoint_id", checkpointID)
	}

	return checkpoint, nil
}

// Delete removes a checkpoint
func (m *Manager) Delete(ctx context.Context, checkpointID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	checkpoint, exists := m.checkpoints[checkpointID]
	if !exists {
		return errors.New(errors.ErrInternal, "checkpoint not found").
			WithField("checkpoint_id", checkpointID)
	}

	log.WithField("checkpoint_id", checkpointID).Info("Deleting checkpoint")

	// Remove checkpoint directory
	checkpointDir := filepath.Dir(checkpoint.ImagePath)
	if err := os.RemoveAll(checkpointDir); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to remove checkpoint directory", err).
			WithField("checkpoint_id", checkpointID)
	}

	// Remove from cache
	delete(m.checkpoints, checkpointID)

	// Remove from state store
	if err := m.stateStore.Delete(checkpointID); err != nil {
		log.WithError(err).Warn("Failed to delete checkpoint from state store")
	}

	log.WithField("checkpoint_id", checkpointID).Info("Checkpoint deleted successfully")
	return nil
}

// Migrate performs live migration to another host
func (m *Manager) Migrate(ctx context.Context, containerID string, opts *MigrationOptions) error {
	migrator := NewMigrator(m.criuManager)
	return migrator.Migrate(ctx, containerID, opts)
}

// Close closes the checkpoint manager
func (m *Manager) Close() error {
	return nil
}

// loadCheckpoints loads existing checkpoints from disk
func (m *Manager) loadCheckpoints() error {
	entries, err := os.ReadDir(m.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		checkpointID := entry.Name()
		if checkpointID == "state" {
			continue
		}

		checkpoint, err := m.loadCheckpoint(checkpointID)
		if err != nil {
			log.WithError(err).
				WithField("checkpoint_id", checkpointID).
				Warn("Failed to load checkpoint")
			continue
		}

		m.checkpoints[checkpointID] = checkpoint
	}

	log.Debugf("Loaded %d checkpoints", len(m.checkpoints))
	return nil
}

// loadCheckpoint loads a checkpoint from disk
func (m *Manager) loadCheckpoint(checkpointID string) (*Checkpoint, error) {
	path := filepath.Join(m.storePath, checkpointID, "checkpoint.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}

	return &checkpoint, nil
}

// saveCheckpoint saves a checkpoint to disk
func (m *Manager) saveCheckpoint(checkpoint *Checkpoint) error {
	checkpointDir := filepath.Join(m.storePath, checkpoint.ID)
	if err := os.MkdirAll(checkpointDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(checkpointDir, "checkpoint.json")

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// generateCheckpointID generates a unique checkpoint ID
func generateCheckpointID(containerID string) string {
	timestamp := time.Now().UnixNano()
	prefix := containerID
	if len(containerID) > 12 {
		prefix = containerID[:12]
	}
	return fmt.Sprintf("ckpt-%s-%d", prefix, timestamp)
}

// calculateDirSize calculates the size of a directory
func calculateDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
