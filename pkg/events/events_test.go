package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	event := CreateEvent(EventContainerStart, "container123", "alpine", "mycontainer", map[string]string{
		"test": "value",
	})

	if event.Type != EventContainerStart {
		t.Errorf("Type = %v, want %v", event.Type, EventContainerStart)
	}

	if event.ContainerID != "container123" {
		t.Errorf("ContainerID = %s, want container123", event.ContainerID)
	}

	if event.Image != "alpine" {
		t.Errorf("Image = %s, want alpine", event.Image)
	}

	if event.Name != "mycontainer" {
		t.Errorf("Name = %s, want mycontainer", event.Name)
	}

	if event.Attributes["test"] != "value" {
		t.Errorf("Attributes[test] = %s, want value", event.Attributes["test"])
	}
}

func TestEventFilters(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		filters EventFilters
		event   *Event
		want    bool
	}{
		{
			name:    "no filters matches all",
			filters: EventFilters{},
			event: &Event{
				Type:        EventContainerStart,
				Timestamp:   now,
				ContainerID: "abc",
			},
			want: true,
		},
		{
			name: "since filter matches",
			filters: EventFilters{
				Since: &past,
			},
			event: &Event{
				Timestamp: now,
			},
			want: true,
		},
		{
			name: "since filter doesn't match",
			filters: EventFilters{
				Since: &future,
			},
			event: &Event{
				Timestamp: now,
			},
			want: false,
		},
		{
			name: "until filter matches",
			filters: EventFilters{
				Until: &future,
			},
			event: &Event{
				Timestamp: now,
			},
			want: true,
		},
		{
			name: "until filter doesn't match",
			filters: EventFilters{
				Until: &past,
			},
			event: &Event{
				Timestamp: now,
			},
			want: false,
		},
		{
			name: "container ID matches",
			filters: EventFilters{
				ContainerID: "abc123",
			},
			event: &Event{
				ContainerID: "abc123",
			},
			want: true,
		},
		{
			name: "container ID doesn't match",
			filters: EventFilters{
				ContainerID: "xyz",
			},
			event: &Event{
				ContainerID: "abc123",
			},
			want: false,
		},
		{
			name: "event type matches",
			filters: EventFilters{
				Types: []EventType{EventContainerStart, EventContainerStop},
			},
			event: &Event{
				Type: EventContainerStart,
			},
			want: true,
		},
		{
			name: "event type doesn't match",
			filters: EventFilters{
				Types: []EventType{EventContainerStop},
			},
			event: &Event{
				Type: EventContainerStart,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filters.Matches(tt.event)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventManager(t *testing.T) {
	// Create temporary state directory
	tmpDir, err := os.MkdirTemp("", "containr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateDir := filepath.Join(tmpDir, "state")

	// Create event manager
	em, err := NewEventManager(stateDir)
	if err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	// Test emitting events
	event1 := CreateEvent(EventContainerStart, "container1", "alpine", "test1", nil)
	em.Emit(event1)

	event2 := CreateEvent(EventContainerStop, "container2", "ubuntu", "test2", nil)
	em.Emit(event2)

	// Test getting all events
	events := em.GetEvents(EventFilters{})
	if len(events) != 2 {
		t.Errorf("GetEvents() returned %d events, want 2", len(events))
	}

	// Test filtering by container ID
	events = em.GetEvents(EventFilters{
		ContainerID: "container1",
	})
	if len(events) != 1 {
		t.Errorf("GetEvents(container1) returned %d events, want 1", len(events))
	}

	// Test filtering by event type
	events = em.GetEvents(EventFilters{
		Types: []EventType{EventContainerStart},
	})
	if len(events) != 1 {
		t.Errorf("GetEvents(start) returned %d events, want 1", len(events))
	}
}

func TestEventSubscription(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "containr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateDir := filepath.Join(tmpDir, "state")

	em, err := NewEventManager(stateDir)
	if err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	// Subscribe to events
	received := make(chan *Event, 1)
	em.Subscribe(func(e *Event) {
		received <- e
	})

	// Emit event
	event := CreateEvent(EventContainerStart, "container1", "alpine", "test", nil)
	em.Emit(event)

	// Wait for event
	select {
	case e := <-received:
		if e.ContainerID != "container1" {
			t.Errorf("Received event with container ID %s, want container1", e.ContainerID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}
