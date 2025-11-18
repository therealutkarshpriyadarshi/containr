package gitops

import (
	"context"
	"testing"
	"time"
)

func TestNewController(t *testing.T) {
	config := &Config{
		Repository: "https://github.com/example/repo",
		Branch:     "main",
		Interval:   30 * time.Second,
	}

	controller := NewController(config)
	if controller == nil {
		t.Fatal("Expected controller to be created")
	}

	if controller.repository != config.Repository {
		t.Errorf("Expected repository %s, got %s", config.Repository, controller.repository)
	}
}

func TestDeploymentOperations(t *testing.T) {
	config := &Config{
		Repository: "https://github.com/example/repo",
		Branch:     "main",
		Interval:   30 * time.Second,
	}

	controller := NewController(config)
	ctx := context.Background()

	deployment := &Deployment{
		Name:      "test-app",
		Namespace: "default",
		Image:     "nginx:latest",
		Revision:  "abc123",
		Spec: &DeploymentSpec{
			Replicas: 3,
			Image:    "nginx:latest",
		},
	}

	// Test apply deployment
	err := controller.applyDeployment(ctx, deployment)
	if err != nil {
		t.Fatalf("Failed to apply deployment: %v", err)
	}

	// Test get deployment
	retrieved, err := controller.GetDeployment("test-app")
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	if retrieved.Name != deployment.Name {
		t.Errorf("Expected name %s, got %s", deployment.Name, retrieved.Name)
	}

	// Test list deployments
	deployments := controller.ListDeployments()
	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}

	// Test delete deployment
	err = controller.DeleteDeployment(ctx, "test-app")
	if err != nil {
		t.Fatalf("Failed to delete deployment: %v", err)
	}

	deployments = controller.ListDeployments()
	if len(deployments) != 0 {
		t.Errorf("Expected 0 deployments after delete, got %d", len(deployments))
	}
}

func TestPipelineExecutor(t *testing.T) {
	executor := NewPipelineExecutor()
	if executor == nil {
		t.Fatal("Expected executor to be created")
	}

	pipeline := &Pipeline{
		Name: "test-pipeline",
		Stages: []*Stage{
			{
				Name: "build",
				Steps: []*Step{
					{
						Name:    "compile",
						Command: "make",
						Args:    []string{"build"},
					},
				},
			},
			{
				Name: "test",
				Steps: []*Step{
					{
						Name:    "unit-tests",
						Command: "make",
						Args:    []string{"test"},
					},
				},
			},
		},
		Timeout: 30 * time.Minute,
	}

	ctx := context.Background()
	err := executor.Execute(ctx, pipeline)
	if err != nil {
		t.Fatalf("Failed to execute pipeline: %v", err)
	}
}

func TestReconciler(t *testing.T) {
	config := &Config{
		Repository: "https://github.com/example/repo",
		Branch:     "main",
		Interval:   30 * time.Second,
	}

	controller := NewController(config)
	reconciler := NewReconciler(controller)

	if reconciler == nil {
		t.Fatal("Expected reconciler to be created")
	}

	deployment := &Deployment{
		Name:     "test-app",
		Revision: "abc123",
	}

	ctx := context.Background()
	err := reconciler.Reconcile(ctx, deployment)
	if err != nil {
		t.Fatalf("Failed to reconcile: %v", err)
	}
}
