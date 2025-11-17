package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
)

// StateStore manages checkpoint state persistence
type StateStore struct {
	root  string
	mu    sync.RWMutex
	cache map[string]*ContainerCheckpointState
}

// ContainerCheckpointState represents the complete state of a checkpoint
type ContainerCheckpointState struct {
	ID              string                 `json:"id"`
	ContainerID     string                 `json:"container_id"`
	ContainerName   string                 `json:"container_name,omitempty"`
	CheckpointName  string                 `json:"checkpoint_name"`
	Created         time.Time              `json:"created"`
	Updated         time.Time              `json:"updated"`
	Status          string                 `json:"status"`
	ImagePath       string                 `json:"image_path"`
	Size            int64                  `json:"size"`
	ProcessState    *ProcessState          `json:"process_state,omitempty"`
	FileSystemState *FileSystemState       `json:"filesystem_state,omitempty"`
	NetworkState    *NetworkState          `json:"network_state,omitempty"`
	MemoryState     *MemoryState           `json:"memory_state,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Tags            []string               `json:"tags,omitempty"`
	ParentID        string                 `json:"parent_id,omitempty"`
	Children        []string               `json:"children,omitempty"`
}

// ProcessState represents the process state in a checkpoint
type ProcessState struct {
	PID          int               `json:"pid"`
	PPID         int               `json:"ppid"`
	Command      []string          `json:"command"`
	WorkingDir   string            `json:"working_dir"`
	Environment  []string          `json:"environment"`
	User         string            `json:"user"`
	Group        string            `json:"group"`
	Capabilities []string          `json:"capabilities"`
	OpenFiles    []OpenFileInfo    `json:"open_files,omitempty"`
	Threads      []ThreadInfo      `json:"threads,omitempty"`
	Signals      *SignalInfo       `json:"signals,omitempty"`
}

// OpenFileInfo represents an open file descriptor
type OpenFileInfo struct {
	FD    int    `json:"fd"`
	Path  string `json:"path"`
	Flags int    `json:"flags"`
	Pos   int64  `json:"pos"`
	Type  string `json:"type"` // regular, socket, pipe, etc.
}

// ThreadInfo represents thread state
type ThreadInfo struct {
	TID          int    `json:"tid"`
	State        string `json:"state"`
	CPUAffinity  []int  `json:"cpu_affinity,omitempty"`
}

// SignalInfo represents signal state
type SignalInfo struct {
	Pending []int          `json:"pending,omitempty"`
	Blocked []int          `json:"blocked,omitempty"`
	Handlers map[int]string `json:"handlers,omitempty"`
}

// FileSystemState represents filesystem-related state
type FileSystemState struct {
	RootFS       string        `json:"rootfs"`
	WorkingDir   string        `json:"working_dir"`
	Mounts       []MountInfo   `json:"mounts"`
	FileLocks    []FileLockInfo `json:"file_locks,omitempty"`
}

// MountInfo represents a mount point
type MountInfo struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Type        string   `json:"type"`
	Options     []string `json:"options"`
	Flags       int      `json:"flags"`
}

// FileLockInfo represents a file lock
type FileLockInfo struct {
	Path string `json:"path"`
	Type string `json:"type"` // read, write, exclusive
	PID  int    `json:"pid"`
}

// NetworkState represents network-related state
type NetworkState struct {
	Hostname     string           `json:"hostname"`
	Interfaces   []InterfaceInfo  `json:"interfaces"`
	Routes       []RouteInfo      `json:"routes,omitempty"`
	Connections  []ConnectionInfo `json:"connections,omitempty"`
	Sockets      []SocketInfo     `json:"sockets,omitempty"`
}

// InterfaceInfo represents a network interface
type InterfaceInfo struct {
	Name       string   `json:"name"`
	Index      int      `json:"index"`
	MAC        string   `json:"mac"`
	Addresses  []string `json:"addresses"`
	MTU        int      `json:"mtu"`
	Flags      []string `json:"flags"`
}

// RouteInfo represents a routing table entry
type RouteInfo struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
}

// ConnectionInfo represents a network connection
type ConnectionInfo struct {
	Protocol    string `json:"protocol"` // tcp, udp, unix
	LocalAddr   string `json:"local_addr"`
	LocalPort   int    `json:"local_port,omitempty"`
	RemoteAddr  string `json:"remote_addr,omitempty"`
	RemotePort  int    `json:"remote_port,omitempty"`
	State       string `json:"state"`
	PID         int    `json:"pid"`
	FD          int    `json:"fd"`
}

// SocketInfo represents a socket
type SocketInfo struct {
	Type    string `json:"type"` // unix, inet, inet6
	Path    string `json:"path,omitempty"`
	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty"`
	PID     int    `json:"pid"`
	FD      int    `json:"fd"`
}

// MemoryState represents memory-related state
type MemoryState struct {
	TotalPages     int64            `json:"total_pages"`
	DirtyPages     int64            `json:"dirty_pages"`
	SharedPages    int64            `json:"shared_pages"`
	PageSize       int64            `json:"page_size"`
	MemoryRegions  []MemoryRegion   `json:"memory_regions,omitempty"`
	Statistics     *MemoryStats     `json:"statistics,omitempty"`
}

// MemoryRegion represents a memory region
type MemoryRegion struct {
	Start       uint64   `json:"start"`
	End         uint64   `json:"end"`
	Permissions string   `json:"permissions"` // rwx
	Type        string   `json:"type"`        // heap, stack, anonymous, file-backed
	Path        string   `json:"path,omitempty"`
	Offset      uint64   `json:"offset,omitempty"`
}

// MemoryStats represents memory statistics
type MemoryStats struct {
	RSS        int64 `json:"rss"`         // Resident set size
	VSZ        int64 `json:"vsz"`         // Virtual memory size
	Shared     int64 `json:"shared"`      // Shared memory
	SwapUsed   int64 `json:"swap_used"`   // Swap usage
	PageFaults int64 `json:"page_faults"` // Page faults
}

// NewStateStore creates a new state store
func NewStateStore(root string) (*StateStore, error) {
	// Create root directory
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create state directory", err).
			WithField("path", root)
	}

	store := &StateStore{
		root:  root,
		cache: make(map[string]*ContainerCheckpointState),
	}

	// Load existing states
	if err := store.loadCache(); err != nil {
		log.WithError(err).Warn("Failed to load state cache")
	}

	return store, nil
}

// Save persists checkpoint state to disk
func (s *StateStore) Save(state *ContainerCheckpointState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state.Updated = time.Now()

	path := s.statePath(state.ID)
	dir := filepath.Dir(path)

	// Create directory
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create state directory", err).
			WithField("checkpoint_id", state.ID)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to marshal checkpoint state", err).
			WithField("checkpoint_id", state.ID)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to write checkpoint state", err).
			WithField("checkpoint_id", state.ID)
	}

	// Update cache
	s.cache[state.ID] = state

	return nil
}

// Load retrieves checkpoint state from disk
func (s *StateStore) Load(id string) (*ContainerCheckpointState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	if state, ok := s.cache[id]; ok {
		return state, nil
	}

	return s.loadFromDisk(id)
}

// loadFromDisk loads checkpoint state from disk
func (s *StateStore) loadFromDisk(id string) (*ContainerCheckpointState, error) {
	path := s.statePath(id)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrInternal, "checkpoint state not found").
				WithField("checkpoint_id", id)
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to read checkpoint state", err).
			WithField("checkpoint_id", id)
	}

	var state ContainerCheckpointState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to unmarshal checkpoint state", err).
			WithField("checkpoint_id", id)
	}

	return &state, nil
}

// Delete removes checkpoint state from disk
func (s *StateStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.statePath(id)

	if err := os.RemoveAll(filepath.Dir(path)); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.ErrInternal, "failed to delete checkpoint state", err).
			WithField("checkpoint_id", id)
	}

	// Remove from cache
	delete(s.cache, id)

	return nil
}

// List returns all checkpoint states
func (s *StateStore) List() ([]*ContainerCheckpointState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var states []*ContainerCheckpointState

	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return states, nil
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to read state directory", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		state, err := s.loadFromDisk(id)
		if err != nil {
			log.WithError(err).WithField("checkpoint_id", id).Warn("Failed to load checkpoint state")
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

// ListByContainer returns checkpoint states for a specific container
func (s *StateStore) ListByContainer(containerID string) ([]*ContainerCheckpointState, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var filtered []*ContainerCheckpointState
	for _, state := range all {
		if state.ContainerID == containerID {
			filtered = append(filtered, state)
		}
	}

	return filtered, nil
}

// Exists checks if a checkpoint state exists
func (s *StateStore) Exists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.cache[id]; ok {
		return true
	}

	path := s.statePath(id)
	_, err := os.Stat(path)
	return err == nil
}

// statePath returns the path to a checkpoint state file
func (s *StateStore) statePath(id string) string {
	return filepath.Join(s.root, id, "state.json")
}

// loadCache loads all checkpoint states into memory cache
func (s *StateStore) loadCache() error {
	states, err := s.List()
	if err != nil {
		return err
	}

	for _, state := range states {
		s.cache[state.ID] = state
	}

	log.Debugf("Loaded %d checkpoint states into cache", len(states))
	return nil
}

// UpdateStatus updates the status of a checkpoint state
func (s *StateStore) UpdateStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.cache[id]
	if !ok {
		loadedState, err := s.loadFromDisk(id)
		if err != nil {
			return err
		}
		state = loadedState
	}

	state.Status = status
	state.Updated = time.Now()

	return s.Save(state)
}

// AddMetadata adds metadata to a checkpoint state
func (s *StateStore) AddMetadata(id string, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.cache[id]
	if !ok {
		loadedState, err := s.loadFromDisk(id)
		if err != nil {
			return err
		}
		state = loadedState
	}

	if state.Metadata == nil {
		state.Metadata = make(map[string]interface{})
	}

	state.Metadata[key] = value
	state.Updated = time.Now()

	return s.Save(state)
}

// AddTag adds a tag to a checkpoint state
func (s *StateStore) AddTag(id string, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.cache[id]
	if !ok {
		loadedState, err := s.loadFromDisk(id)
		if err != nil {
			return err
		}
		state = loadedState
	}

	// Check if tag already exists
	for _, t := range state.Tags {
		if t == tag {
			return nil // Tag already exists
		}
	}

	state.Tags = append(state.Tags, tag)
	state.Updated = time.Now()

	return s.Save(state)
}

// RemoveTag removes a tag from a checkpoint state
func (s *StateStore) RemoveTag(id string, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.cache[id]
	if !ok {
		loadedState, err := s.loadFromDisk(id)
		if err != nil {
			return err
		}
		state = loadedState
	}

	// Filter out the tag
	var newTags []string
	for _, t := range state.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	state.Tags = newTags
	state.Updated = time.Now()

	return s.Save(state)
}

// FindByTag finds checkpoint states with a specific tag
func (s *StateStore) FindByTag(tag string) ([]*ContainerCheckpointState, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var filtered []*ContainerCheckpointState
	for _, state := range all {
		for _, t := range state.Tags {
			if t == tag {
				filtered = append(filtered, state)
				break
			}
		}
	}

	return filtered, nil
}

// GetStatistics returns checkpoint statistics
func (s *StateStore) GetStatistics() (*Statistics, error) {
	states, err := s.List()
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		Total: len(states),
		ByStatus: make(map[string]int),
		TotalSize: 0,
	}

	for _, state := range states {
		stats.ByStatus[state.Status]++
		stats.TotalSize += state.Size
	}

	return stats, nil
}

// Statistics represents checkpoint statistics
type Statistics struct {
	Total     int            `json:"total"`
	ByStatus  map[string]int `json:"by_status"`
	TotalSize int64          `json:"total_size"`
}

// Export exports checkpoint state to a portable format
func (s *StateStore) Export(id string) ([]byte, error) {
	state, err := s.Load(id)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to marshal checkpoint state", err)
	}

	return data, nil
}

// Import imports checkpoint state from a portable format
func (s *StateStore) Import(data []byte) (*ContainerCheckpointState, error) {
	var state ContainerCheckpointState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to unmarshal checkpoint state", err)
	}

	// Verify required fields
	if state.ID == "" {
		return nil, errors.New(errors.ErrInternal, "checkpoint state missing ID")
	}

	if state.ContainerID == "" {
		return nil, errors.New(errors.ErrInternal, "checkpoint state missing container ID")
	}

	// Save imported state
	if err := s.Save(&state); err != nil {
		return nil, fmt.Errorf("failed to save imported state: %w", err)
	}

	return &state, nil
}
