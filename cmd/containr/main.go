package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var (
	debugMode   bool
	logLevel    string
	logToStderr bool
)

func main() {
	// Check if we're being called as a child process
	if len(os.Args) > 1 && os.Args[1] == "child" {
		// We're in the child process - do setup
		log := logger.New("child")
		if err := container.SetupChild(); err != nil {
			log.WithError(err).Error("Child setup failed")
			fmt.Fprintf(os.Stderr, "Child setup failed: %v\n", err)
			os.Exit(1)
		}

		// Execute the actual command
		if len(os.Args) < 3 {
			log.Error("No command specified")
			fmt.Fprintf(os.Stderr, "No command specified\n")
			os.Exit(1)
		}

		// Execute the command passed to us
		// In a real implementation, we'd use exec.Command here
		// For now, just demonstrate that we're isolated
		log.Infof("Executing in isolated namespace: %v", os.Args[2:])
		fmt.Printf("Executing in isolated namespace: %v\n", os.Args[2:])
		os.Exit(0)
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Define global flags
	globalFlags := flag.NewFlagSet("global", flag.ExitOnError)
	globalFlags.BoolVar(&debugMode, "debug", false, "Enable debug mode with verbose output")
	globalFlags.StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error)")
	globalFlags.BoolVar(&logToStderr, "log-stderr", false, "Log to stderr instead of stdout")

	switch command {
	case "run":
		// Parse run-specific flags
		runFlags := flag.NewFlagSet("run", flag.ExitOnError)
		runFlags.BoolVar(&debugMode, "debug", false, "Enable debug mode with verbose output")
		runFlags.StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error)")

		// Parse flags
		if err := runFlags.Parse(os.Args[2:]); err != nil {
			handleError(errors.Wrap(errors.ErrInvalidArgument, "failed to parse flags", err))
		}

		// Initialize logger
		initLogger()

		// Get remaining args after flags
		args := runFlags.Args()
		if err := runCommand(args); err != nil {
			handleError(err)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// initLogger initializes the logger with the specified settings
func initLogger() {
	log := logger.GetLogger()

	// Set log level
	if debugMode {
		log.SetLevel(logger.DebugLevel)
	} else {
		switch logLevel {
		case "debug":
			log.SetLevel(logger.DebugLevel)
		case "info":
			log.SetLevel(logger.InfoLevel)
		case "warn":
			log.SetLevel(logger.WarnLevel)
		case "error":
			log.SetLevel(logger.ErrorLevel)
		default:
			log.SetLevel(logger.InfoLevel)
		}
	}

	// Set output
	if logToStderr {
		log.SetOutput(os.Stderr)
	}

	logger.Debug("Logger initialized")
	logger.Debugf("Debug mode: %v, Log level: %s", debugMode, logLevel)
}

// handleError handles errors with proper formatting
func handleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a ContainrError
	if ce, ok := err.(*errors.ContainrError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", ce.GetFullMessage())
		logger.WithFields(map[string]interface{}{
			"code":   ce.Code,
			"fields": ce.Fields,
		}).Error(ce.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		logger.Error(err)
	}
	os.Exit(1)
}

func runCommand(args []string) error {
	log := logger.New("run")

	if len(args) == 0 {
		return errors.New(errors.ErrInvalidArgument, "no command specified").
			WithHint("Usage: containr run [flags] <command> [args...]")
	}

	log.Debugf("Running command: %v", args)

	// Parse flags
	hostname := "container"
	isolate := true

	log.Debugf("Container configuration: hostname=%s, isolate=%v", hostname, isolate)

	// Create container configuration
	config := &container.Config{
		Command:  args,
		Hostname: hostname,
		Isolate:  isolate,
	}

	// Create and run container
	c := container.New(config)

	log.WithField("container_id", c.ID).Info("Starting container")
	fmt.Printf("Starting container %s...\n", c.ID)

	if debugMode {
		fmt.Printf("Namespaces: UTS, PID, Mount, IPC, Network\n")
		fmt.Printf("Command: %v\n", c.Command)
		fmt.Println("---")
	}

	if err := c.RunWithSetup(); err != nil {
		log.WithError(err).WithField("container_id", c.ID).Error("Container execution failed")
		return errors.Wrap(errors.ErrContainerStart, "failed to run container", err)
	}

	log.WithField("container_id", c.ID).Info("Container exited successfully")
	fmt.Printf("\nContainer %s exited\n", c.ID)
	return nil
}

func printUsage() {
	fmt.Println("containr - A simple container runtime built from Linux primitives")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  containr run [flags] <command> [args...]  Run a command in a container")
	fmt.Println("  containr help                              Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --debug            Enable debug mode with verbose output")
	fmt.Println("  --log-level        Set log level (debug, info, warn, error) [default: info]")
	fmt.Println("  --log-stderr       Log to stderr instead of stdout")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  containr run /bin/sh")
	fmt.Println("  containr run /bin/bash -c 'echo Hello from container'")
	fmt.Println("  containr run --debug /bin/sh")
	fmt.Println("  containr run --log-level debug /bin/bash")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Process isolation with namespaces (UTS, PID, Mount, IPC, Network)")
	fmt.Println("  - Resource limits with cgroups")
	fmt.Println("  - Filesystem isolation")
	fmt.Println("  - Security features (capabilities, seccomp, AppArmor/SELinux)")
	fmt.Println("  - Structured logging with configurable levels")
	fmt.Println("  - Comprehensive error handling with hints")
	fmt.Println()
	fmt.Println("Note: This tool requires root privileges to create namespaces")
}
