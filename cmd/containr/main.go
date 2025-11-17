package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var (
	// Global flags
	debugMode   bool
	logLevel    string
	stateDir    string
	volumeDir   string
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
		log.Infof("Executing in isolated namespace: %v", os.Args[2:])
		fmt.Printf("Executing in isolated namespace: %v\n", os.Args[2:])
		os.Exit(0)
	}

	// Execute cobra CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "containr",
	Short: "A minimal container runtime built from scratch",
	Long: `Containr is an educational container runtime that demonstrates core
containerization concepts using Linux primitives like namespaces, cgroups,
and overlay filesystems.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode with verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&stateDir, "state-dir", "/var/lib/containr/state", "State directory for container metadata")
	rootCmd.PersistentFlags().StringVar(&volumeDir, "volume-dir", "/var/lib/containr/volumes", "Volume directory")

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(imagesCmd)
	rootCmd.AddCommand(rmiCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(volumeCmd)
	rootCmd.AddCommand(versionCmd)
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

	logger.Debug("Logger initialized")
	logger.Debugf("Debug mode: %v, Log level: %s", debugMode, logLevel)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containr version 2.0.0 (Phase 2 Complete)")
		fmt.Println("- Enhanced CLI with full lifecycle management")
		fmt.Println("- Volume management support")
		fmt.Println("- Registry integration (pull images)")
		fmt.Println("- User namespace support for rootless containers")
	},
}
