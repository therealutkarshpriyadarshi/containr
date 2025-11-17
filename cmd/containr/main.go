package main

import (
	"fmt"
	"os"

	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
)

func main() {
	// Check if we're being called as a child process
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// We're in the child process - do setup
		if err := container.SetupChild(); err != nil {
			fmt.Fprintf(os.Stderr, "Child setup failed: %v\n", err)
			os.Exit(1)
		}

		// Execute the actual command
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "No command specified\n")
			os.Exit(1)
		}

		// Execute the command passed to us
		// In a real implementation, we'd use exec.Command here
		// For now, just demonstrate that we're isolated
		fmt.Printf("Executing in isolated namespace: %v\n", os.Args[2:])
		os.Exit(0)
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if err := runCommand(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// Parse flags
	hostname := "container"
	isolate := true

	// Create container configuration
	config := &container.Config{
		Command:  args,
		Hostname: hostname,
		Isolate:  isolate,
	}

	// Create and run container
	c := container.New(config)

	fmt.Printf("Starting container %s...\n", c.ID)
	fmt.Printf("Namespaces: UTS, PID, Mount, IPC, Network\n")
	fmt.Printf("Command: %v\n", c.Command)
	fmt.Println("---")

	if err := c.RunWithSetup(); err != nil {
		return err
	}

	fmt.Printf("\nContainer %s exited\n", c.ID)
	return nil
}

func printUsage() {
	fmt.Println("containr - A simple container runtime built from Linux primitives")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  containr run <command> [args...]  Run a command in a container")
	fmt.Println("  containr help                     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  containr run /bin/sh")
	fmt.Println("  containr run /bin/bash -c 'echo Hello from container'")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Process isolation with namespaces (UTS, PID, Mount, IPC, Network)")
	fmt.Println("  - Resource limits with cgroups")
	fmt.Println("  - Filesystem isolation")
	fmt.Println()
	fmt.Println("Note: This tool requires root privileges to create namespaces")
}
