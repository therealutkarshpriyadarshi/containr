package volume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("volume")

// VolumeType represents the type of volume
type VolumeType string

const (
	// TypeBind represents a bind mount
	TypeBind VolumeType = "bind"
	// TypeVolume represents a named volume
	TypeVolume VolumeType = "volume"
	// TypeTmpfs represents a tmpfs mount
	TypeTmpfs VolumeType = "tmpfs"
)

// Volume represents a volume or mount
type Volume struct {
	Name        string                 `json:"name,omitempty"`
	Type        VolumeType             `json:"type"`
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`
	ReadOnly    bool                   `json:"read_only"`
	Options     []string               `json:"options,omitempty"`
	Created     time.Time              `json:"created"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Manager manages volumes
type Manager struct {
	root  string
	mu    sync.RWMutex
	cache map[string]*Volume
}

const (
	// DefaultRoot is the default volume directory
	DefaultRoot = "/var/lib/containr/volumes"
)

// NewManager creates a new volume manager
func NewManager(root string) (*Manager, error) {
	if root == "" {
		root = DefaultRoot
	}

	// Create root directory if it doesn't exist
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create volume directory", err).
			WithField("path", root).
			WithHint("Ensure you have write permissions to the directory")
	}

	manager := &Manager{
		root:  root,
		cache: make(map[string]*Volume),
	}

	// Load existing volumes into cache
	if err := manager.loadCache(); err != nil {
		log.WithError(err).Warn("Failed to load cache, starting with empty cache")
	}

	return manager, nil
}

// Create creates a new named volume
func (m *Manager) Create(name string, options map[string]string) (*Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if volume already exists
	if _, exists := m.cache[name]; exists {
		return nil, errors.New(errors.ErrInternal, fmt.Sprintf("volume '%s' already exists", name)).
			WithField("name", name)
	}

	// Create volume directory
	volumePath := filepath.Join(m.root, name)
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create volume directory", err).
			WithField("name", name).
			WithField("path", volumePath)
	}

	volume := &Volume{
		Name:    name,
		Type:    TypeVolume,
		Source:  volumePath,
		Created: time.Now(),
		Metadata: map[string]interface{}{
			"options": options,
		},
	}

	// Save volume metadata
	if err := m.saveVolume(volume); err != nil {
		// Clean up on error
		os.RemoveAll(volumePath)
		return nil, err
	}

	// Add to cache
	m.cache[name] = volume

	log.WithField("name", name).WithField("path", volumePath).Info("Volume created")
	return volume, nil
}

// Remove removes a volume
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, exists := m.cache[name]
	if !exists {
		return errors.New(errors.ErrInternal, fmt.Sprintf("volume '%s' not found", name)).
			WithField("name", name)
	}

	// Remove volume directory
	if err := os.RemoveAll(volume.Source); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.ErrInternal, "failed to remove volume", err).
			WithField("name", name).
			WithField("path", volume.Source)
	}

	// Remove metadata file
	metadataPath := filepath.Join(m.root, name+".json")
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.ErrInternal, "failed to remove volume metadata", err).
			WithField("name", name)
	}

	// Remove from cache
	delete(m.cache, name)

	log.WithField("name", name).Info("Volume removed")
	return nil
}

// Get retrieves a volume by name
func (m *Manager) Get(name string) (*Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volume, exists := m.cache[name]
	if !exists {
		return nil, errors.New(errors.ErrInternal, fmt.Sprintf("volume '%s' not found", name)).
			WithField("name", name)
	}

	return volume, nil
}

// List returns all volumes
func (m *Manager) List() ([]*Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volumes := make([]*Volume, 0, len(m.cache))
	for _, volume := range m.cache {
		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// Exists checks if a volume exists
func (m *Manager) Exists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.cache[name]
	return exists
}

// Mount mounts a volume to a destination
func (m *Manager) Mount(volume *Volume, containerRoot string) error {
	destination := filepath.Join(containerRoot, volume.Destination)

	log.WithFields(map[string]interface{}{
		"type":        volume.Type,
		"source":      volume.Source,
		"destination": destination,
		"read_only":   volume.ReadOnly,
	}).Debug("Mounting volume")

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destination, 0755); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create mount destination", err).
			WithField("destination", destination)
	}

	switch volume.Type {
	case TypeBind:
		return m.mountBind(volume, destination)
	case TypeVolume:
		return m.mountBind(volume, destination)
	case TypeTmpfs:
		return m.mountTmpfs(volume, destination)
	default:
		return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("unsupported volume type: %s", volume.Type)).
			WithField("type", volume.Type)
	}
}

// mountBind performs a bind mount
func (m *Manager) mountBind(volume *Volume, destination string) error {
	// Verify source exists
	if _, err := os.Stat(volume.Source); err != nil {
		return errors.Wrap(errors.ErrInternal, "source path does not exist", err).
			WithField("source", volume.Source).
			WithHint("Ensure the source path exists and is accessible")
	}

	// Perform bind mount
	flags := syscall.MS_BIND | syscall.MS_REC
	if volume.ReadOnly {
		flags |= syscall.MS_RDONLY
	}

	if err := syscall.Mount(volume.Source, destination, "", uintptr(flags), ""); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to bind mount", err).
			WithField("source", volume.Source).
			WithField("destination", destination).
			WithHint("Ensure you have CAP_SYS_ADMIN capability")
	}

	// If read-only, remount with read-only flag
	if volume.ReadOnly {
		flags = syscall.MS_BIND | syscall.MS_REMOUNT | syscall.MS_RDONLY
		if err := syscall.Mount(volume.Source, destination, "", uintptr(flags), ""); err != nil {
			// Try to unmount on error
			syscall.Unmount(destination, 0)
			return errors.Wrap(errors.ErrInternal, "failed to remount as read-only", err).
				WithField("destination", destination)
		}
	}

	log.WithFields(map[string]interface{}{
		"source":      volume.Source,
		"destination": destination,
		"read_only":   volume.ReadOnly,
	}).Info("Bind mount successful")

	return nil
}

// mountTmpfs mounts a tmpfs filesystem
func (m *Manager) mountTmpfs(volume *Volume, destination string) error {
	// Parse options for size
	options := "mode=0755"
	if len(volume.Options) > 0 {
		for _, opt := range volume.Options {
			options += "," + opt
		}
	}

	if err := syscall.Mount("tmpfs", destination, "tmpfs", 0, options); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to mount tmpfs", err).
			WithField("destination", destination).
			WithField("options", options).
			WithHint("Ensure you have CAP_SYS_ADMIN capability")
	}

	log.WithFields(map[string]interface{}{
		"destination": destination,
		"options":     options,
	}).Info("Tmpfs mount successful")

	return nil
}

// Unmount unmounts a volume
func (m *Manager) Unmount(destination string) error {
	log.WithField("destination", destination).Debug("Unmounting volume")

	if err := syscall.Unmount(destination, 0); err != nil && err != syscall.EINVAL {
		return errors.Wrap(errors.ErrInternal, "failed to unmount volume", err).
			WithField("destination", destination)
	}

	log.WithField("destination", destination).Info("Volume unmounted")
	return nil
}

// ParseVolumeString parses a volume string like "src:dest:ro" or "name:/path"
func ParseVolumeString(volumeStr string, volumeManager *Manager) (*Volume, error) {
	parts := splitVolumeString(volumeStr)

	if len(parts) < 2 {
		return nil, errors.New(errors.ErrInvalidArgument, "invalid volume format").
			WithField("volume", volumeStr).
			WithHint("Use format: source:destination[:ro] or volume_name:destination[:ro]")
	}

	source := parts[0]
	destination := parts[1]
	readOnly := len(parts) > 2 && parts[2] == "ro"

	volume := &Volume{
		Source:      source,
		Destination: destination,
		ReadOnly:    readOnly,
	}

	// Determine volume type
	if filepath.IsAbs(source) {
		// Absolute path = bind mount
		volume.Type = TypeBind
	} else if volumeManager != nil && volumeManager.Exists(source) {
		// Named volume
		namedVol, err := volumeManager.Get(source)
		if err != nil {
			return nil, err
		}
		volume.Type = TypeVolume
		volume.Name = source
		volume.Source = namedVol.Source
	} else {
		// Assume bind mount with relative path
		volume.Type = TypeBind
		absPath, err := filepath.Abs(source)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInvalidArgument, "failed to resolve path", err).
				WithField("path", source)
		}
		volume.Source = absPath
	}

	return volume, nil
}

// splitVolumeString splits a volume string by ':'
func splitVolumeString(s string) []string {
	var parts []string
	var current string
	var escaped bool

	for _, ch := range s {
		if escaped {
			current += string(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == ':' {
			parts = append(parts, current)
			current = ""
			continue
		}

		current += string(ch)
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// saveVolume saves volume metadata to disk
func (m *Manager) saveVolume(volume *Volume) error {
	metadataPath := filepath.Join(m.root, volume.Name+".json")

	data, err := json.MarshalIndent(volume, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to marshal volume metadata", err).
			WithField("name", volume.Name)
	}

	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to write volume metadata", err).
			WithField("name", volume.Name).
			WithField("path", metadataPath)
	}

	return nil
}

// loadCache loads all volumes into memory cache
func (m *Manager) loadCache() error {
	entries, err := os.ReadDir(m.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(errors.ErrInternal, "failed to read volume directory", err).
			WithField("path", m.root)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		metadataPath := filepath.Join(m.root, entry.Name())
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			log.WithError(err).WithField("path", metadataPath).Warn("Failed to read volume metadata")
			continue
		}

		var volume Volume
		if err := json.Unmarshal(data, &volume); err != nil {
			log.WithError(err).WithField("path", metadataPath).Warn("Failed to unmarshal volume metadata")
			continue
		}

		m.cache[volume.Name] = &volume
	}

	log.Debugf("Loaded %d volumes into cache", len(m.cache))
	return nil
}

// Prune removes unused volumes
func (m *Manager) Prune() ([]string, error) {
	// In a full implementation, this would check which volumes are in use
	// For now, we'll just return an empty list
	log.Info("Volume prune requested (not yet implemented)")
	return []string{}, nil
}
