package main

import (
	"fmt"
	"os"

	"github.com/therealutkarshpriyadarshi/containr/pkg/cgroup"
	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
)

// Simple example demonstrating container creation with resource limits
func main() {
	// This example requires root privileges
	if os.Geteuid() != 0 {
		fmt.Println("This example requires root privileges. Run with sudo.")
		os.Exit(1)
	}

	// Create cgroup for resource limits
	cgConfig := &cgroup.Config{
		Name:        "example-container",
		MemoryLimit: 100 * 1024 * 1024, // 100 MB
		CPUShares:   512,               // Half of default CPU shares
		PIDLimit:    100,               // Max 100 processes
	}

	cg, err := cgroup.New(cgConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cgroup: %v\n", err)
		os.Exit(1)
	}
	defer cg.Remove()

	// Add current process to cgroup
	if err := cg.AddCurrentProcess(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add process to cgroup: %v\n", err)
		os.Exit(1)
	}

	// Create container configuration
	config := &container.Config{
		Command:  []string{"/bin/sh", "-c", "echo 'Hello from isolated container!'; hostname; ps aux"},
		Hostname: "mycontainer",
		Isolate:  true,
	}

	// Create and run container
	c := container.New(config)

	fmt.Println("Starting container with:")
	fmt.Printf("  Memory limit: 100 MB\n")
	fmt.Printf("  CPU shares: 512\n")
	fmt.Printf("  PID limit: 100\n")
	fmt.Printf("  Hostname: mycontainer\n")
	fmt.Println()

	if err := c.RunWithSetup(); err != nil {
		fmt.Fprintf(os.Stderr, "Container error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nContainer exited successfully")
}
