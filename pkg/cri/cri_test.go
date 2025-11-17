package cri

import (
	"context"
	"testing"
)

func TestCRIService_RunPodSandbox(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	config := &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{
			Name:      "test-pod",
			UID:       "uid-123",
			Namespace: "default",
		},
		Hostname: "test-host",
		Labels:   map[string]string{"app": "test"},
	}

	podID, err := service.RunPodSandbox(ctx, config)
	if err != nil {
		t.Fatalf("Failed to run pod sandbox: %v", err)
	}

	if podID == "" {
		t.Fatal("Expected non-empty pod ID")
	}

	// Verify pod was created
	status, err := service.PodSandboxStatus(ctx, podID)
	if err != nil {
		t.Fatalf("Failed to get pod status: %v", err)
	}

	if status.Metadata.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got '%s'", status.Metadata.Name)
	}

	if status.State != PodSandboxReady {
		t.Errorf("Expected pod state Ready, got %v", status.State)
	}
}

func TestCRIService_StopPodSandbox(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create a pod
	config := &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	}
	podID, _ := service.RunPodSandbox(ctx, config)

	// Stop the pod
	err := service.StopPodSandbox(ctx, podID)
	if err != nil {
		t.Fatalf("Failed to stop pod: %v", err)
	}

	// Verify pod is stopped
	status, _ := service.PodSandboxStatus(ctx, podID)
	if status.State != PodSandboxNotReady {
		t.Errorf("Expected pod state NotReady, got %v", status.State)
	}
}

func TestCRIService_RemovePodSandbox(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create and remove a pod
	config := &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	}
	podID, _ := service.RunPodSandbox(ctx, config)

	err := service.RemovePodSandbox(ctx, podID)
	if err != nil {
		t.Fatalf("Failed to remove pod: %v", err)
	}

	// Verify pod is removed
	_, err = service.PodSandboxStatus(ctx, podID)
	if err == nil {
		t.Fatal("Expected error when getting removed pod status")
	}
}

func TestCRIService_ListPodSandbox(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create multiple pods
	for i := 0; i < 3; i++ {
		config := &PodSandboxConfig{
			Metadata: &PodSandboxMetadata{
				Name: string(rune('A' + i)),
			},
		}
		service.RunPodSandbox(ctx, config)
	}

	// List all pods
	pods, err := service.ListPodSandbox(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(pods) != 3 {
		t.Errorf("Expected 3 pods, got %d", len(pods))
	}

	// Test filtering by state
	readyState := PodSandboxReady
	filter := &PodSandboxFilter{State: &readyState}
	pods, err = service.ListPodSandbox(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list filtered pods: %v", err)
	}

	if len(pods) != 3 {
		t.Errorf("Expected 3 ready pods, got %d", len(pods))
	}
}

func TestCRIService_CreateContainer(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create a pod first
	podConfig := &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	}
	podID, _ := service.RunPodSandbox(ctx, podConfig)

	// Create a container
	containerConfig := &ContainerConfig{
		Metadata: &ContainerMetadata{Name: "test-container"},
		Image:    &ImageSpec{Image: "alpine:latest"},
		Command:  []string{"/bin/sh"},
		Labels:   map[string]string{"app": "test"},
	}

	containerID, err := service.CreateContainer(ctx, podID, containerConfig)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	if containerID == "" {
		t.Fatal("Expected non-empty container ID")
	}

	// Verify container was created
	status, err := service.ContainerStatus(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to get container status: %v", err)
	}

	if status.Metadata.Name != "test-container" {
		t.Errorf("Expected container name 'test-container', got '%s'", status.Metadata.Name)
	}

	if status.State != ContainerCreated {
		t.Errorf("Expected container state Created, got %v", status.State)
	}
}

func TestCRIService_StartContainer(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create pod and container
	podID, _ := service.RunPodSandbox(ctx, &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	})

	containerID, _ := service.CreateContainer(ctx, podID, &ContainerConfig{
		Metadata: &ContainerMetadata{Name: "test-container"},
		Image:    &ImageSpec{Image: "alpine:latest"},
	})

	// Start container
	err := service.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Verify container is running
	status, _ := service.ContainerStatus(ctx, containerID)
	if status.State != ContainerRunning {
		t.Errorf("Expected container state Running, got %v", status.State)
	}
}

func TestCRIService_StopContainer(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create and start container
	podID, _ := service.RunPodSandbox(ctx, &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	})

	containerID, _ := service.CreateContainer(ctx, podID, &ContainerConfig{
		Metadata: &ContainerMetadata{Name: "test-container"},
		Image:    &ImageSpec{Image: "alpine:latest"},
	})

	service.StartContainer(ctx, containerID)

	// Stop container
	err := service.StopContainer(ctx, containerID, 10)
	if err != nil {
		t.Fatalf("Failed to stop container: %v", err)
	}

	// Verify container is stopped
	status, _ := service.ContainerStatus(ctx, containerID)
	if status.State != ContainerExited {
		t.Errorf("Expected container state Exited, got %v", status.State)
	}

	if status.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", status.ExitCode)
	}
}

func TestCRIService_RemoveContainer(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create container
	podID, _ := service.RunPodSandbox(ctx, &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	})

	containerID, _ := service.CreateContainer(ctx, podID, &ContainerConfig{
		Metadata: &ContainerMetadata{Name: "test-container"},
		Image:    &ImageSpec{Image: "alpine:latest"},
	})

	// Remove container
	err := service.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to remove container: %v", err)
	}

	// Verify container is removed
	_, err = service.ContainerStatus(ctx, containerID)
	if err == nil {
		t.Fatal("Expected error when getting removed container status")
	}
}

func TestCRIService_ListContainers(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Create pod
	podID, _ := service.RunPodSandbox(ctx, &PodSandboxConfig{
		Metadata: &PodSandboxMetadata{Name: "test-pod"},
	})

	// Create multiple containers
	for i := 0; i < 3; i++ {
		service.CreateContainer(ctx, podID, &ContainerConfig{
			Metadata: &ContainerMetadata{Name: string(rune('A' + i))},
			Image:    &ImageSpec{Image: "alpine:latest"},
		})
	}

	// List all containers
	containers, err := service.ListContainers(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	if len(containers) != 3 {
		t.Errorf("Expected 3 containers, got %d", len(containers))
	}
}

func TestCRIService_PullImage(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	imageSpec := &ImageSpec{Image: "alpine:latest"}

	imageID, err := service.PullImage(ctx, imageSpec, nil)
	if err != nil {
		t.Fatalf("Failed to pull image: %v", err)
	}

	if imageID == "" {
		t.Fatal("Expected non-empty image ID")
	}

	// Verify image was pulled
	image, err := service.ImageStatus(ctx, imageSpec)
	if err != nil {
		t.Fatalf("Failed to get image status: %v", err)
	}

	if len(image.RepoTags) == 0 {
		t.Error("Expected at least one repo tag")
	}
}

func TestCRIService_ListImages(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	// Pull some images
	images := []string{"alpine:latest", "ubuntu:latest", "nginx:latest"}
	for _, img := range images {
		service.PullImage(ctx, &ImageSpec{Image: img}, nil)
	}

	// List all images
	list, err := service.ListImages(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list images: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 images, got %d", len(list))
	}

	// Test filtering
	filter := &ImageFilter{Image: &ImageSpec{Image: "alpine:latest"}}
	list, err = service.ListImages(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to list filtered images: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 image, got %d", len(list))
	}
}

func TestCRIService_RemoveImage(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	imageSpec := &ImageSpec{Image: "alpine:latest"}

	// Pull and remove image
	service.PullImage(ctx, imageSpec, nil)

	err := service.RemoveImage(ctx, imageSpec)
	if err != nil {
		t.Fatalf("Failed to remove image: %v", err)
	}

	// Verify image is removed
	_, err = service.ImageStatus(ctx, imageSpec)
	if err == nil {
		t.Fatal("Expected error when getting removed image status")
	}
}

func TestCRIService_ImageFsInfo(t *testing.T) {
	service := NewCRIService()
	ctx := context.Background()

	fsInfo, err := service.ImageFsInfo(ctx)
	if err != nil {
		t.Fatalf("Failed to get filesystem info: %v", err)
	}

	if len(fsInfo) == 0 {
		t.Error("Expected at least one filesystem info")
	}

	if fsInfo[0].UsedBytes == nil {
		t.Error("Expected used bytes to be set")
	}
}
