package hotreload

import (
	"context"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	config := &Config{
		ContainerID: "test-container",
		WatchPaths:  []string{"/app"},
		IgnorePatterns: []string{"*.tmp"},
		Debounce:    100 * time.Millisecond,
	}

	watcher := NewWatcher(config)
	if watcher == nil {
		t.Fatal("Expected watcher to be created")
	}

	if watcher.containerID != config.ContainerID {
		t.Errorf("Expected container ID %s, got %s", config.ContainerID, watcher.containerID)
	}
}

func TestWatcherStartStop(t *testing.T) {
	config := &Config{
		ContainerID: "test-container",
		WatchPaths:  []string{"/app"},
	}

	watcher := NewWatcher(config)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := watcher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = watcher.Stop()
	if err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}
}

func TestShouldIgnore(t *testing.T) {
	watcher := &Watcher{
		ignorePatterns: []string{"*.tmp", "*.log"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/app/file.tmp", true},
		{"/app/file.log", true},
		{"/app/file.go", false},
		{"/app/test.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := watcher.shouldIgnore(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.path, result)
			}
		})
	}
}

func TestSyncer(t *testing.T) {
	syncer := NewSyncer("test-container")
	if syncer == nil {
		t.Fatal("Expected syncer to be created")
	}

	pair := &SyncPair{
		HostPath:      "/host/app",
		ContainerPath: "/app",
		Direction:     SyncDirectionBidirectional,
	}

	syncer.AddSyncPair(pair)

	ctx := context.Background()
	err := syncer.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}

func TestReloadManager(t *testing.T) {
	manager := NewReloadManager("test-container", StrategyRestart)
	if manager == nil {
		t.Fatal("Expected reload manager to be created")
	}

	if manager.strategy != StrategyRestart {
		t.Errorf("Expected strategy %s, got %s", StrategyRestart, manager.strategy)
	}

	ctx := context.Background()
	err := manager.Reload(ctx)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	stats := manager.GetStats()
	if stats["reloadCount"].(int) != 1 {
		t.Errorf("Expected reload count 1, got %d", stats["reloadCount"])
	}
}

func TestDevEnvironment(t *testing.T) {
	config := &Config{
		ContainerID: "test-container",
		WatchPaths:  []string{"/app"},
	}

	env := NewDevEnvironment("test-container", config)
	if env == nil {
		t.Fatal("Expected dev environment to be created")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := env.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start dev environment: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = env.Stop()
	if err != nil {
		t.Fatalf("Failed to stop dev environment: %v", err)
	}
}
