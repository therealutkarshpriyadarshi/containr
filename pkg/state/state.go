package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("state")

// ContainerState represents the state of a container
type ContainerState string

const (
	// StateCreated means the container has been created but not started
	StateCreated ContainerState = "created"
	// StateRunning means the container is currently running
	StateRunning ContainerState = "running"
	// StateStopped means the container has been stopped
	StateStopped ContainerState = "stopped"
	// StateExited means the container has exited
	StateExited ContainerState = "exited"
	// StatePaused means the container is paused
	StatePaused ContainerState = "paused"
)

// Container represents stored container information
type Container struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	State       ContainerState         `json:"state"`
	PID         int                    `json:"pid,omitempty"`
	ExitCode    int                    `json:"exit_code,omitempty"`
	Created     time.Time              `json:"created"`
	Started     time.Time              `json:"started,omitempty"`
	Finished    time.Time              `json:"finished,omitempty"`
	Image       string                 `json:"image,omitempty"`
	Command     []string               `json:"command"`
	Hostname    string                 `json:"hostname,omitempty"`
	RootFS      string                 `json:"rootfs,omitempty"`
	WorkingDir  string                 `json:"working_dir,omitempty"`
	Env         []string               `json:"env,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Volumes     []Volume               `json:"volumes,omitempty"`
	NetworkMode string                 `json:"network_mode,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Volume represents a mounted volume
type Volume struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ReadOnly    bool   `json:"read_only"`
	Type        string `json:"type"` // bind, volume, tmpfs
}

// Store manages container state persistence
type Store struct {
	root  string
	mu    sync.RWMutex
	cache map[string]*Container
}

const (
	// DefaultRoot is the default state directory
	DefaultRoot = "/var/lib/containr/state"
)

// NewStore creates a new state store
func NewStore(root string) (*Store, error) {
	if root == "" {
		root = DefaultRoot
	}

	// Create root directory if it doesn't exist
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create state directory", err).
			WithField("path", root).
			WithHint("Ensure you have write permissions to the directory")
	}

	store := &Store{
		root:  root,
		cache: make(map[string]*Container),
	}

	// Load existing containers into cache
	if err := store.loadCache(); err != nil {
		log.WithError(err).Warn("Failed to load cache, starting with empty cache")
	}

	return store, nil
}

// Save persists container state to disk
func (s *Store) Save(container *Container) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.containerPath(container.ID)
	dir := filepath.Dir(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create container state directory", err).
			WithField("container_id", container.ID)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to marshal container state", err).
			WithField("container_id", container.ID)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to write container state", err).
			WithField("container_id", container.ID).
			WithField("path", path)
	}

	// Update cache
	s.cache[container.ID] = container

	log.WithField("container_id", container.ID).Debug("Container state saved")
	return nil
}

// Load retrieves container state from disk
func (s *Store) Load(id string) (*Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	if container, ok := s.cache[id]; ok {
		return container, nil
	}

	return s.loadFromDisk(id)
}

// loadFromDisk loads container state from disk
func (s *Store) loadFromDisk(id string) (*Container, error) {
	path := s.containerPath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrContainerNotFound, "container not found").
				WithField("container_id", id)
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to read container state", err).
			WithField("container_id", id).
			WithField("path", path)
	}

	var container Container
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to unmarshal container state", err).
			WithField("container_id", id)
	}

	return &container, nil
}

// Delete removes container state from disk
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.containerPath(id)

	if err := os.RemoveAll(filepath.Dir(path)); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.ErrInternal, "failed to delete container state", err).
			WithField("container_id", id)
	}

	// Remove from cache
	delete(s.cache, id)

	log.WithField("container_id", id).Debug("Container state deleted")
	return nil
}

// List returns all containers
func (s *Store) List() ([]*Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var containers []*Container

	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return containers, nil
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to read state directory", err).
			WithField("path", s.root)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		container, err := s.loadFromDisk(id)
		if err != nil {
			log.WithError(err).WithField("container_id", id).Warn("Failed to load container state")
			continue
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// ListByState returns containers in a specific state
func (s *Store) ListByState(state ContainerState) ([]*Container, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var filtered []*Container
	for _, container := range all {
		if container.State == state {
			filtered = append(filtered, container)
		}
	}

	return filtered, nil
}

// Exists checks if a container exists
func (s *Store) Exists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.cache[id]; ok {
		return true
	}

	path := s.containerPath(id)
	_, err := os.Stat(path)
	return err == nil
}

// containerPath returns the path to a container's state file
func (s *Store) containerPath(id string) string {
	return filepath.Join(s.root, id, "state.json")
}

// loadCache loads all containers into memory cache
func (s *Store) loadCache() error {
	containers, err := s.List()
	if err != nil {
		return err
	}

	for _, container := range containers {
		s.cache[container.ID] = container
	}

	log.Debugf("Loaded %d containers into cache", len(containers))
	return nil
}

// FindByName finds a container by name
func (s *Store) FindByName(name string) (*Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	for _, container := range s.cache {
		if container.Name == name {
			return container, nil
		}
	}

	// Search on disk
	containers, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.Name == name {
			return container, nil
		}
	}

	return nil, errors.New(errors.ErrContainerNotFound, fmt.Sprintf("container with name '%s' not found", name)).
		WithField("name", name)
}
