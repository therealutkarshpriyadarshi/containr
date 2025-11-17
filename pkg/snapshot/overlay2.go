package snapshot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Overlay2Snapshotter implements the Snapshotter interface using overlay2
type Overlay2Snapshotter struct {
	root     string
	metadata *Metadata
}

// NewOverlay2 creates a new overlay2 snapshotter
func NewOverlay2(root string) (*Overlay2Snapshotter, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return &Overlay2Snapshotter{
		root:     root,
		metadata: NewMetadata(root),
	}, nil
}

// Prepare creates an active snapshot
func (o *Overlay2Snapshotter) Prepare(ctx context.Context, key, parent string, opts ...Opt) ([]Mount, error) {
	cfg, err := buildConfig(opts)
	if err != nil {
		return nil, err
	}

	// Check if key already exists
	if o.metadata.Exists(key) {
		return nil, fmt.Errorf("snapshot %s already exists", key)
	}

	// Create snapshot directories
	snapshotDir := filepath.Join(o.root, "snapshots", key)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Create work, upper, and merged directories
	workDir := filepath.Join(snapshotDir, "work")
	upperDir := filepath.Join(snapshotDir, "upper")
	mergedDir := filepath.Join(snapshotDir, "merged")

	for _, dir := range []string{workDir, upperDir, mergedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Save metadata
	info := Info{
		Name:      key,
		Parent:    parent,
		Kind:      KindActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    cfg.labels,
	}

	if err := o.metadata.Save(info); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Build mount options
	return o.buildMounts(key, parent)
}

// View creates a read-only view of a snapshot
func (o *Overlay2Snapshotter) View(ctx context.Context, key, parent string, opts ...Opt) ([]Mount, error) {
	cfg, err := buildConfig(opts)
	if err != nil {
		return nil, err
	}

	// Check if key already exists
	if o.metadata.Exists(key) {
		return nil, fmt.Errorf("snapshot %s already exists", key)
	}

	// Verify parent exists
	if parent != "" && !o.metadata.Exists(parent) {
		return nil, fmt.Errorf("parent snapshot %s not found", parent)
	}

	// Create snapshot directory
	snapshotDir := filepath.Join(o.root, "snapshots", key)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Save metadata
	info := Info{
		Name:      key,
		Parent:    parent,
		Kind:      KindView,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    cfg.labels,
	}

	if err := o.metadata.Save(info); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Build read-only mounts
	return o.buildMounts(key, parent)
}

// Commit creates an immutable snapshot from an active snapshot
func (o *Overlay2Snapshotter) Commit(ctx context.Context, name, key string, opts ...Opt) error {
	// Load active snapshot metadata
	info, err := o.metadata.Load(key)
	if err != nil {
		return fmt.Errorf("failed to load snapshot %s: %w", key, err)
	}

	if info.Kind != KindActive {
		return fmt.Errorf("snapshot %s is not active", key)
	}

	// Create committed snapshot
	committedDir := filepath.Join(o.root, "snapshots", name)
	activeUpperDir := filepath.Join(o.root, "snapshots", key, "upper")

	if err := os.MkdirAll(committedDir, 0755); err != nil {
		return fmt.Errorf("failed to create committed snapshot directory: %w", err)
	}

	// Copy upper directory to committed snapshot
	committedDataDir := filepath.Join(committedDir, "data")
	if err := copyDir(activeUpperDir, committedDataDir); err != nil {
		return fmt.Errorf("failed to copy snapshot data: %w", err)
	}

	// Update metadata
	cfg, err := buildConfig(opts)
	if err != nil {
		return err
	}

	committedInfo := Info{
		Name:      name,
		Parent:    info.Parent,
		Kind:      KindCommitted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    cfg.labels,
	}

	if err := o.metadata.Save(committedInfo); err != nil {
		return fmt.Errorf("failed to save committed metadata: %w", err)
	}

	return nil
}

// Remove removes a snapshot
func (o *Overlay2Snapshotter) Remove(ctx context.Context, key string) error {
	// Remove snapshot directory
	snapshotDir := filepath.Join(o.root, "snapshots", key)
	if err := os.RemoveAll(snapshotDir); err != nil {
		return fmt.Errorf("failed to remove snapshot directory: %w", err)
	}

	// Remove metadata
	if err := o.metadata.Delete(key); err != nil {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	return nil
}

// Mounts returns the mounts for a snapshot
func (o *Overlay2Snapshotter) Mounts(ctx context.Context, key string) ([]Mount, error) {
	info, err := o.metadata.Load(key)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot %s: %w", key, err)
	}

	return o.buildMounts(key, info.Parent)
}

// Stat returns info about a snapshot
func (o *Overlay2Snapshotter) Stat(ctx context.Context, key string) (Info, error) {
	info, err := o.metadata.Load(key)
	if err != nil {
		return Info{}, fmt.Errorf("failed to load snapshot %s: %w", key, err)
	}

	// Calculate size if needed
	if info.Size == 0 {
		size, err := o.calculateSize(key)
		if err == nil {
			info.Size = size
		}
	}

	return info, nil
}

// Update updates snapshot metadata
func (o *Overlay2Snapshotter) Update(ctx context.Context, info Info, fieldpaths ...string) (Info, error) {
	existing, err := o.metadata.Load(info.Name)
	if err != nil {
		return Info{}, fmt.Errorf("failed to load snapshot %s: %w", info.Name, err)
	}

	// Update specified fields
	for _, field := range fieldpaths {
		switch field {
		case "labels":
			existing.Labels = info.Labels
		}
	}

	existing.UpdatedAt = time.Now()

	if err := o.metadata.Save(existing); err != nil {
		return Info{}, fmt.Errorf("failed to update metadata: %w", err)
	}

	return existing, nil
}

// Walk iterates over all snapshots
func (o *Overlay2Snapshotter) Walk(ctx context.Context, fn WalkFunc) error {
	infos, err := o.metadata.List()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	for _, info := range infos {
		if err := fn(ctx, info); err != nil {
			return err
		}
	}

	return nil
}

// Usage returns storage usage for a snapshot
func (o *Overlay2Snapshotter) Usage(ctx context.Context, key string) (Usage, error) {
	size, err := o.calculateSize(key)
	if err != nil {
		return Usage{}, err
	}

	return Usage{
		Size: size,
	}, nil
}

// Close closes the snapshotter
func (o *Overlay2Snapshotter) Close() error {
	return nil
}

// buildMounts builds overlay mounts for a snapshot
func (o *Overlay2Snapshotter) buildMounts(key, parent string) ([]Mount, error) {
	snapshotDir := filepath.Join(o.root, "snapshots", key)
	workDir := filepath.Join(snapshotDir, "work")
	upperDir := filepath.Join(snapshotDir, "upper")
	mergedDir := filepath.Join(snapshotDir, "merged")

	var lowerDirs []string

	// Build lower directory chain
	current := parent
	for current != "" {
		info, err := o.metadata.Load(current)
		if err != nil {
			break
		}

		if info.Kind == KindCommitted {
			dataDir := filepath.Join(o.root, "snapshots", current, "data")
			lowerDirs = append(lowerDirs, dataDir)
		} else if info.Kind == KindActive {
			upperDir := filepath.Join(o.root, "snapshots", current, "upper")
			lowerDirs = append(lowerDirs, upperDir)
		}

		current = info.Parent
	}

	// Build overlay options
	var options []string
	if len(lowerDirs) > 0 {
		options = append(options, "lowerdir="+strings.Join(lowerDirs, ":"))
	}
	options = append(options, "upperdir="+upperDir)
	options = append(options, "workdir="+workDir)

	return []Mount{
		{
			Type:    "overlay",
			Source:  "overlay",
			Target:  mergedDir,
			Options: options,
		},
	}, nil
}

// calculateSize calculates the size of a snapshot
func (o *Overlay2Snapshotter) calculateSize(key string) (int64, error) {
	snapshotDir := filepath.Join(o.root, "snapshots", key)

	var size int64
	err := filepath.Walk(snapshotDir, func(path string, info os.FileInfo, err error) error {
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

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
