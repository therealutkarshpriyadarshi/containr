package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Kind represents the kind of snapshot
type Kind int

const (
	// KindActive represents an active snapshot that can be written to
	KindActive Kind = iota
	// KindCommitted represents a committed (immutable) snapshot
	KindCommitted
	// KindView represents a read-only view of a snapshot
	KindView
)

func (k Kind) String() string {
	switch k {
	case KindActive:
		return "active"
	case KindCommitted:
		return "committed"
	case KindView:
		return "view"
	default:
		return "unknown"
	}
}

// Info contains metadata about a snapshot
type Info struct {
	Name      string            `json:"name"`
	Parent    string            `json:"parent,omitempty"`
	Kind      Kind              `json:"kind"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Labels    map[string]string `json:"labels,omitempty"`
	Size      int64             `json:"size,omitempty"`
}

// Mount represents a filesystem mount point
type Mount struct {
	Type    string   `json:"type"`
	Source  string   `json:"source"`
	Target  string   `json:"target,omitempty"`
	Options []string `json:"options,omitempty"`
}

// Usage represents storage usage information
type Usage struct {
	Inodes int64 `json:"inodes"`
	Size   int64 `json:"size"`
}

// WalkFunc is called for each snapshot during a walk operation
type WalkFunc func(context.Context, Info) error

// Snapshotter defines the interface for snapshot operations
type Snapshotter interface {
	// Prepare creates an active snapshot identified by key
	// The key can be thought of as a transaction identifier for creating a new snapshot
	Prepare(ctx context.Context, key, parent string, opts ...Opt) ([]Mount, error)

	// View creates a read-only view of a snapshot
	View(ctx context.Context, key, parent string, opts ...Opt) ([]Mount, error)

	// Commit creates an immutable snapshot from an active snapshot
	// The key is the active snapshot identifier, name is the committed snapshot name
	Commit(ctx context.Context, name, key string, opts ...Opt) error

	// Remove removes a snapshot by key or name
	Remove(ctx context.Context, key string) error

	// Mounts returns the mounts for an active snapshot transaction
	Mounts(ctx context.Context, key string) ([]Mount, error)

	// Stat returns info about a snapshot by name or key
	Stat(ctx context.Context, key string) (Info, error)

	// Update updates the info for a snapshot
	Update(ctx context.Context, info Info, fieldpaths ...string) (Info, error)

	// Walk iterates over all snapshots
	Walk(ctx context.Context, fn WalkFunc) error

	// Usage returns storage usage for a snapshot
	Usage(ctx context.Context, key string) (Usage, error)

	// Close closes the snapshotter and cleans up resources
	Close() error
}

// Opt is a functional option for snapshot operations
type Opt func(*config) error

type config struct {
	labels map[string]string
}

// WithLabels adds labels to a snapshot
func WithLabels(labels map[string]string) Opt {
	return func(c *config) error {
		c.labels = labels
		return nil
	}
}

func buildConfig(opts []Opt) (*config, error) {
	cfg := &config{
		labels: make(map[string]string),
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Metadata manages snapshot metadata persistence
type Metadata struct {
	root string
}

// NewMetadata creates a new metadata manager
func NewMetadata(root string) *Metadata {
	return &Metadata{root: root}
}

// Save saves snapshot info to disk
func (m *Metadata) Save(info Info) error {
	path := filepath.Join(m.root, "metadata", info.Name+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// Load loads snapshot info from disk
func (m *Metadata) Load(name string) (Info, error) {
	path := filepath.Join(m.root, "metadata", name+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return Info{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	var info Info
	if err := json.Unmarshal(data, &info); err != nil {
		return Info{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return info, nil
}

// Delete removes snapshot metadata
func (m *Metadata) Delete(name string) error {
	path := filepath.Join(m.root, "metadata", name+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}
	return nil
}

// List returns all snapshot metadata
func (m *Metadata) List() ([]Info, error) {
	metadataDir := filepath.Join(m.root, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var infos []Info
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		info, err := m.Load(name)
		if err != nil {
			continue // Skip corrupted metadata
		}

		infos = append(infos, info)
	}

	return infos, nil
}

// Exists checks if snapshot metadata exists
func (m *Metadata) Exists(name string) bool {
	path := filepath.Join(m.root, "metadata", name+".json")
	_, err := os.Stat(path)
	return err == nil
}
