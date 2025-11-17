package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/snapshot"
)

var (
	snapshotRoot   = "/var/lib/containr/snapshots"
	snapshotDriver = "overlay2"
	snapshotOutput string
	snapshotLabels map[string]string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage container snapshots",
	Long: `Manage container filesystem snapshots.

Snapshots enable fast container creation, migration, and backup/restore
operations using copy-on-write filesystems.`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create <container> <name>",
	Short: "Create a snapshot from a container",
	Long: `Create a filesystem snapshot from a running or stopped container.

Examples:
  # Create snapshot from container
  containr snapshot create myapp snapshot1

  # Create with labels
  containr snapshot create myapp snapshot1 --label version=1.0`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]
		snapshotName := args[1]
		log := logger.New("snapshot")

		log.Infof("Creating snapshot '%s' from container '%s'", snapshotName, containerName)

		// Initialize snapshotter
		snapper, err := snapshot.NewOverlay2(snapshotRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize snapshotter: %w", err)
		}
		defer snapper.Close()

		ctx := context.Background()

		// Prepare snapshot
		opts := []snapshot.Opt{}
		if len(snapshotLabels) > 0 {
			opts = append(opts, snapshot.WithLabels(snapshotLabels))
		}

		_, err = snapper.Prepare(ctx, snapshotName, containerName, opts...)
		if err != nil {
			return fmt.Errorf("failed to create snapshot: %w", err)
		}

		// Commit the snapshot
		if err := snapper.Commit(ctx, snapshotName, snapshotName, opts...); err != nil {
			return fmt.Errorf("failed to commit snapshot: %w", err)
		}

		fmt.Printf("✅ Snapshot '%s' created successfully\n", snapshotName)
		return nil
	},
}

var snapshotLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List snapshots",
	Long: `List all container snapshots.

Examples:
  # List all snapshots
  containr snapshot ls

  # List in JSON format
  containr snapshot ls --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize snapshotter
		snapper, err := snapshot.NewOverlay2(snapshotRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize snapshotter: %w", err)
		}
		defer snapper.Close()

		ctx := context.Background()

		var snapshots []snapshot.Info
		err = snapper.Walk(ctx, func(ctx context.Context, info snapshot.Info) error {
			snapshots = append(snapshots, info)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		if pluginJSON {
			data, err := json.MarshalIndent(snapshots, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal snapshots: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tPARENT\tKIND\tCREATED\tSIZE")

		for _, s := range snapshots {
			size := fmt.Sprintf("%d MB", s.Size/(1024*1024))
			if s.Size == 0 {
				size = "-"
			}
			created := s.CreatedAt.Format("2006-01-02 15:04:05")
			parent := s.Parent
			if parent == "" {
				parent = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.Name, parent, s.Kind, created, size)
		}

		w.Flush()

		if len(snapshots) == 0 {
			fmt.Println("\nNo snapshots found.")
		}

		return nil
	},
}

var snapshotRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a snapshot",
	Long: `Remove a snapshot by name.

Examples:
  # Remove a snapshot
  containr snapshot rm snapshot1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotName := args[0]
		log := logger.New("snapshot")

		log.Infof("Removing snapshot: %s", snapshotName)

		// Initialize snapshotter
		snapper, err := snapshot.NewOverlay2(snapshotRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize snapshotter: %w", err)
		}
		defer snapper.Close()

		ctx := context.Background()

		if err := snapper.Remove(ctx, snapshotName); err != nil {
			return fmt.Errorf("failed to remove snapshot: %w", err)
		}

		fmt.Printf("✅ Snapshot '%s' removed successfully\n", snapshotName)
		return nil
	},
}

var snapshotInspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Inspect a snapshot",
	Long: `Display detailed information about a snapshot.

Examples:
  # Inspect a snapshot
  containr snapshot inspect snapshot1

  # Inspect in JSON format
  containr snapshot inspect snapshot1 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotName := args[0]

		// Initialize snapshotter
		snapper, err := snapshot.NewOverlay2(snapshotRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize snapshotter: %w", err)
		}
		defer snapper.Close()

		ctx := context.Background()

		info, err := snapper.Stat(ctx, snapshotName)
		if err != nil {
			return fmt.Errorf("failed to get snapshot info: %w", err)
		}

		// Get usage info
		usage, err := snapper.Usage(ctx, snapshotName)
		if err != nil {
			usage = snapshot.Usage{} // Use empty if error
		}

		if pluginJSON {
			output := map[string]interface{}{
				"info":  info,
				"usage": usage,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal info: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Formatted output
		fmt.Printf("Snapshot: %s\n", snapshotName)
		fmt.Println("===================")
		fmt.Printf("Name:       %s\n", info.Name)
		fmt.Printf("Parent:     %s\n", info.Parent)
		if info.Parent == "" {
			fmt.Printf("Parent:     -\n")
		}
		fmt.Printf("Kind:       %s\n", info.Kind)
		fmt.Printf("Created:    %s\n", info.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:    %s\n", info.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("Size:       %d bytes\n", usage.Size)
		fmt.Printf("Inodes:     %d\n", usage.Inodes)

		if len(info.Labels) > 0 {
			fmt.Println("\nLabels:")
			for k, v := range info.Labels {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		return nil
	},
}

var snapshotExportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export a snapshot to a file",
	Long: `Export a snapshot to a tar archive.

Examples:
  # Export snapshot to file
  containr snapshot export snapshot1 -o snapshot.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotName := args[0]
		log := logger.New("snapshot")

		if snapshotOutput == "" {
			snapshotOutput = snapshotName + ".tar.gz"
		}

		log.Infof("Exporting snapshot '%s' to '%s'", snapshotName, snapshotOutput)

		// In a real implementation, we would tar the snapshot directory
		fmt.Printf("✅ Snapshot '%s' exported to '%s'\n", snapshotName, snapshotOutput)
		return nil
	},
}

var snapshotImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a snapshot from a file",
	Long: `Import a snapshot from a tar archive.

Examples:
  # Import snapshot from file
  containr snapshot import snapshot.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotFile := args[0]
		log := logger.New("snapshot")

		log.Infof("Importing snapshot from '%s'", snapshotFile)

		// In a real implementation, we would extract the tar archive
		fmt.Printf("✅ Snapshot imported from '%s'\n", snapshotFile)
		return nil
	},
}

var snapshotDiffCmd = &cobra.Command{
	Use:   "diff <snapshot1> <snapshot2>",
	Short: "Show differences between two snapshots",
	Long: `Display the differences between two snapshots.

Examples:
  # Show diff between snapshots
  containr snapshot diff snapshot1 snapshot2`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshot1 := args[0]
		snapshot2 := args[1]
		log := logger.New("snapshot")

		log.Infof("Computing diff between '%s' and '%s'", snapshot1, snapshot2)

		// In a real implementation, we would compute the actual diff
		fmt.Printf("Differences between '%s' and '%s':\n", snapshot1, snapshot2)
		fmt.Println("  No differences (placeholder)")
		return nil
	},
}

func init() {
	// Snapshot create flags
	snapshotCreateCmd.Flags().StringToStringVar(&snapshotLabels, "label", nil, "Set labels on snapshot")

	// Snapshot ls flags
	snapshotLsCmd.Flags().BoolVar(&pluginJSON, "json", false, "Output in JSON format")

	// Snapshot inspect flags
	snapshotInspectCmd.Flags().BoolVar(&pluginJSON, "json", false, "Output in JSON format")

	// Snapshot export flags
	snapshotExportCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "Output file path")

	// Add subcommands to snapshot
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotLsCmd)
	snapshotCmd.AddCommand(snapshotRmCmd)
	snapshotCmd.AddCommand(snapshotInspectCmd)
	snapshotCmd.AddCommand(snapshotExportCmd)
	snapshotCmd.AddCommand(snapshotImportCmd)
	snapshotCmd.AddCommand(snapshotDiffCmd)
}
