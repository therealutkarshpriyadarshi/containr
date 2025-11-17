package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// EventType represents the type of container event
type EventType string

const (
	// Container lifecycle events
	EventContainerCreate  EventType = "container:create"
	EventContainerStart   EventType = "container:start"
	EventContainerStop    EventType = "container:stop"
	EventContainerRestart EventType = "container:restart"
	EventContainerRemove  EventType = "container:remove"
	EventContainerDie     EventType = "container:die"
	EventContainerKill    EventType = "container:kill"
	EventContainerPause   EventType = "container:pause"
	EventContainerUnpause EventType = "container:unpause"

	// Container health events
	EventContainerHealthy   EventType = "container:health_status:healthy"
	EventContainerUnhealthy EventType = "container:health_status:unhealthy"

	// Resource events
	EventContainerOOM      EventType = "container:oom"
	EventContainerThrottle EventType = "container:throttle"

	// Network events
	EventNetworkCreate     EventType = "network:create"
	EventNetworkRemove     EventType = "network:remove"
	EventNetworkConnect    EventType = "network:connect"
	EventNetworkDisconnect EventType = "network:disconnect"

	// Volume events
	EventVolumeCreate  EventType = "volume:create"
	EventVolumeRemove  EventType = "volume:remove"
	EventVolumeMount   EventType = "volume:mount"
	EventVolumeUnmount EventType = "volume:unmount"

	// Image events
	EventImagePull   EventType = "image:pull"
	EventImageRemove EventType = "image:remove"
	EventImageImport EventType = "image:import"
	EventImageExport EventType = "image:export"
)

// Event represents a container-related event
type Event struct {
	Type        EventType         `json:"type"`
	Timestamp   time.Time         `json:"timestamp"`
	ContainerID string            `json:"container_id,omitempty"`
	Image       string            `json:"image,omitempty"`
	Name        string            `json:"name,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// EventListener is a function that receives events
type EventListener func(*Event)

// EventManager manages container events
type EventManager struct {
	listeners []EventListener
	events    []*Event
	eventFile string
	mu        sync.RWMutex
	log       *logger.Logger
}

// NewEventManager creates a new event manager
func NewEventManager(stateDir string) (*EventManager, error) {
	log := logger.New("events")

	eventsDir := filepath.Join(stateDir, "events")
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create events directory: %w", err)
	}

	em := &EventManager{
		listeners: make([]EventListener, 0),
		events:    make([]*Event, 0),
		eventFile: filepath.Join(eventsDir, "events.log"),
		log:       log,
	}

	// Load existing events
	if err := em.loadEvents(); err != nil {
		log.Warnf("Failed to load existing events: %v", err)
	}

	return em, nil
}

// Emit emits a new event
func (em *EventManager) Emit(event *Event) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Store event
	em.events = append(em.events, event)

	// Persist event
	if err := em.persistEvent(event); err != nil {
		em.log.Warnf("Failed to persist event: %v", err)
	}

	// Notify listeners
	for _, listener := range em.listeners {
		go listener(event)
	}

	em.log.Debugf("Event emitted: %s for %s", event.Type, event.ContainerID)
}

// Subscribe adds an event listener
func (em *EventManager) Subscribe(listener EventListener) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.listeners = append(em.listeners, listener)
}

// GetEvents returns events, optionally filtered
func (em *EventManager) GetEvents(filters EventFilters) []*Event {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var filtered []*Event
	for _, event := range em.events {
		if filters.Matches(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// EventFilters holds filters for querying events
type EventFilters struct {
	Since       *time.Time
	Until       *time.Time
	ContainerID string
	Types       []EventType
	Limit       int
}

// Matches checks if an event matches the filters
func (f *EventFilters) Matches(event *Event) bool {
	// Check time range
	if f.Since != nil && event.Timestamp.Before(*f.Since) {
		return false
	}
	if f.Until != nil && event.Timestamp.After(*f.Until) {
		return false
	}

	// Check container ID
	if f.ContainerID != "" && event.ContainerID != f.ContainerID {
		return false
	}

	// Check event types
	if len(f.Types) > 0 {
		matched := false
		for _, t := range f.Types {
			if event.Type == t {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// persistEvent appends an event to the event log file
func (em *EventManager) persistEvent(event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(em.eventFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}

// loadEvents loads events from the event log file
func (em *EventManager) loadEvents() error {
	data, err := os.ReadFile(em.eventFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := string(data)
	for _, line := range filepath.SplitList(lines) {
		if line == "" {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			em.log.Warnf("Failed to parse event: %v", err)
			continue
		}

		em.events = append(em.events, &event)
	}

	return nil
}

// CreateEvent creates a new event
func CreateEvent(eventType EventType, containerID, image, name string, attrs map[string]string) *Event {
	return &Event{
		Type:        eventType,
		Timestamp:   time.Now(),
		ContainerID: containerID,
		Image:       image,
		Name:        name,
		Attributes:  attrs,
	}
}

// CreateErrorEvent creates an error event
func CreateErrorEvent(eventType EventType, containerID string, err error) *Event {
	return &Event{
		Type:        eventType,
		Timestamp:   time.Now(),
		ContainerID: containerID,
		Error:       err.Error(),
	}
}

// FormatEvent formats an event for display
func FormatEvent(e *Event) string {
	base := fmt.Sprintf("%s %s", e.Timestamp.Format(time.RFC3339), e.Type)

	if e.ContainerID != "" {
		base += fmt.Sprintf(" container=%s", e.ContainerID)
	}

	if e.Name != "" {
		base += fmt.Sprintf(" name=%s", e.Name)
	}

	if e.Image != "" {
		base += fmt.Sprintf(" image=%s", e.Image)
	}

	if e.Error != "" {
		base += fmt.Sprintf(" error=%s", e.Error)
	}

	for k, v := range e.Attributes {
		base += fmt.Sprintf(" %s=%s", k, v)
	}

	return base
}
