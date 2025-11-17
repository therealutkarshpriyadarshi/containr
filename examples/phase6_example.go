package main

import (
	"context"
	"fmt"
	"log"

	"github.com/therealutkarshpriyadarshi/containr/pkg/build"
	"github.com/therealutkarshpriyadarshi/containr/pkg/cri"
	"github.com/therealutkarshpriyadarshi/containr/pkg/plugin"
	"github.com/therealutkarshpriyadarshi/containr/pkg/snapshot"
)

// This example demonstrates the Phase 6 features:
// 1. Plugin System
// 2. Snapshot Support
// 3. CRI Integration
// 4. Build Engine

// Example 1: Plugin System
func examplePluginSystem() {
	fmt.Println("=== Plugin System Example ===")

	// Create plugin manager
	pm := plugin.NewManager()

	// Create a custom metrics plugin
	metricsPlugin := &CustomMetricsPlugin{
		BasePlugin: plugin.NewBasePlugin("custom-metrics", plugin.MetricsPlugin, "1.0.0"),
	}

	// Register plugin
	if err := pm.Register(metricsPlugin); err != nil {
		log.Fatal(err)
	}

	// Enable plugin
	ctx := context.Background()
	config := map[string]interface{}{
		"port": 9090,
		"path": "/metrics",
	}

	if err := pm.Enable(ctx, "custom-metrics", config); err != nil {
		log.Fatal(err)
	}

	// List plugins
	plugins := pm.List()
	for _, p := range plugins {
		fmt.Printf("Plugin: %s (type: %s, version: %s, enabled: %v)\n",
			p.Name, p.Type, p.Version, p.Enabled)
	}

	// Health check
	healthResults := pm.HealthCheck(ctx)
	for name, err := range healthResults {
		if err != nil {
			fmt.Printf("Plugin %s is unhealthy: %v\n", name, err)
		} else {
			fmt.Printf("Plugin %s is healthy\n", name)
		}
	}

	fmt.Println()
}

// CustomMetricsPlugin is an example metrics plugin
type CustomMetricsPlugin struct {
	*plugin.BasePlugin
}

func (p *CustomMetricsPlugin) Start(ctx context.Context) error {
	fmt.Println("Starting custom metrics plugin...")
	// In a real implementation, start metrics server here
	return nil
}

func (p *CustomMetricsPlugin) Stop(ctx context.Context) error {
	fmt.Println("Stopping custom metrics plugin...")
	return nil
}

// Example 2: Snapshot Support
func exampleSnapshotSupport() {
	fmt.Println("=== Snapshot Support Example ===")

	// Create snapshotter
	snapshotter, err := snapshot.NewOverlay2("/tmp/containr/snapshots")
	if err != nil {
		log.Fatal(err)
	}
	defer snapshotter.Close()

	ctx := context.Background()

	// Create a snapshot
	labels := map[string]string{
		"app":     "myapp",
		"version": "1.0",
	}

	mounts, err := snapshotter.Prepare(ctx, "snapshot-1", "", snapshot.WithLabels(labels))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created snapshot with %d mounts\n", len(mounts))

	// Get snapshot info
	info, err := snapshotter.Stat(ctx, "snapshot-1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Snapshot: %s (kind: %s, created: %s)\n",
		info.Name, info.Kind, info.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Labels: %v\n", info.Labels)

	// Commit snapshot
	err = snapshotter.Commit(ctx, "snapshot-1-committed", "snapshot-1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Snapshot committed successfully")

	// List all snapshots
	fmt.Println("All snapshots:")
	snapshotter.Walk(ctx, func(ctx context.Context, info snapshot.Info) error {
		fmt.Printf("  - %s (kind: %s)\n", info.Name, info.Kind)
		return nil
	})

	fmt.Println()
}

// Example 3: CRI Integration
func exampleCRIIntegration() {
	fmt.Println("=== CRI Integration Example ===")

	// Create CRI service
	criService := cri.NewCRIService()

	ctx := context.Background()

	// Create pod sandbox
	podConfig := &cri.PodSandboxConfig{
		Metadata: &cri.PodSandboxMetadata{
			Name:      "example-pod",
			UID:       "uid-12345",
			Namespace: "default",
			Attempt:   0,
		},
		Hostname: "example-host",
		Labels: map[string]string{
			"app": "example",
		},
	}

	podID, err := criService.RunPodSandbox(ctx, podConfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created pod sandbox: %s\n", podID)

	// Create container
	containerConfig := &cri.ContainerConfig{
		Metadata: &cri.ContainerMetadata{
			Name:    "example-container",
			Attempt: 0,
		},
		Image: &cri.ImageSpec{
			Image: "alpine:latest",
		},
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "echo Hello from container"},
		Labels: map[string]string{
			"component": "app",
		},
	}

	containerID, err := criService.CreateContainer(ctx, podID, containerConfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created container: %s\n", containerID)

	// Start container
	err = criService.StartContainer(ctx, containerID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Container started")

	// Get container status
	status, err := criService.ContainerStatus(ctx, containerID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Container state: %v\n", status.State)

	// List containers
	containers, err := criService.ListContainers(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total containers: %d\n", len(containers))

	// Pull image
	imageSpec := &cri.ImageSpec{Image: "nginx:latest"}
	imageID, err := criService.PullImage(ctx, imageSpec, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pulled image: %s\n", imageID)

	// List images
	images, err := criService.ListImages(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total images: %d\n", len(images))

	fmt.Println()
}

// Example 4: Build Engine
func exampleBuildEngine() {
	fmt.Println("=== Build Engine Example ===")

	// Create a simple Dockerfile
	// In a real scenario, you would read this from a file
	parser := build.NewParser()

	// For this example, we'll just demonstrate the build context
	ctx := context.Background()

	// Create build context
	buildContext, err := build.NewBuildContext(".")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Build context hash: %s\n", buildContext.Hash()[:16])

	// Create build cache
	cache, err := build.NewBuildCache("/tmp/containr/cache", true)
	if err != nil {
		log.Fatal(err)
	}

	// Example cache key
	cacheKey := build.CacheKey{
		ParentHash:  "parent123",
		Instruction: "RUN echo test",
		ContextHash: buildContext.Hash(),
		BuildArgs:   map[string]string{"VERSION": "1.0"},
	}

	fmt.Printf("Cache key hash: %s\n", cacheKey.Hash()[:16])

	// Check cache (will miss on first run)
	if layer, found := cache.Lookup(cacheKey); found {
		fmt.Printf("Cache hit! Layer: %s\n", layer.ID)
	} else {
		fmt.Println("Cache miss - would build new layer")

		// Create a new layer (simulated)
		layer := &build.Layer{
			ID:      "layer-abc123",
			Parent:  "parent123",
			Command: "RUN echo test",
			Size:    1024,
		}

		// Store in cache
		if err := cache.Store(cacheKey, layer); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Layer stored in cache")
	}

	// Demonstrate Dockerfile parser
	fmt.Println("\nExample Dockerfile instructions:")
	instructions := []string{
		"FROM alpine:latest",
		"RUN apk add --no-cache git",
		"WORKDIR /app",
		"COPY . .",
		"CMD [\"/app/start.sh\"]",
	}

	for _, instr := range instructions {
		fmt.Printf("  %s\n", instr)
	}

	fmt.Println()

	// Note: Full build would be done with:
	// dockerfile, _ := parser.ParseFile("Dockerfile")
	// config := &build.BuildConfig{Tags: []string{"myapp:latest"}}
	// builder, _ := build.NewBuilder(dockerfile, buildContext, config)
	// manifest, _ := builder.Build(ctx)

	_ = ctx
	_ = parser
}

func main() {
	fmt.Println("Containr Phase 6 Features Demo")
	fmt.Println("===============================\n")

	// Run examples
	examplePluginSystem()
	exampleSnapshotSupport()
	exampleCRIIntegration()
	exampleBuildEngine()

	fmt.Println("Demo complete!")
}
