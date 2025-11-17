package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/state"
)

// inspect command
var (
	inspectFormat string
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] CONTAINER [CONTAINER...]",
	Short: "Display detailed information on one or more containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		var containers []*state.Container
		for _, nameOrID := range args {
			container, err := loadContainerByNameOrID(store, nameOrID)
			if err != nil {
				return err
			}
			containers = append(containers, container)
		}

		// Format output
		if inspectFormat != "" {
			// TODO: Implement custom format templates
			fmt.Println("Custom format not yet implemented")
			return nil
		}

		// JSON output
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(containers); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to encode output", err)
		}

		return nil
	},
}

func init() {
	inspectCmd.Flags().StringVarP(&inspectFormat, "format", "f", "", "Format the output using a Go template")
}

// stats command
var (
	statsAll    bool
	statsNoStream bool
)

var statsCmd = &cobra.Command{
	Use:   "stats [flags] [CONTAINER...]",
	Short: "Display a live stream of container(s) resource usage statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		var containers []*state.Container
		if len(args) == 0 {
			// Show stats for all running containers
			containers, err = store.ListByState(state.StateRunning)
			if err != nil {
				return errors.Wrap(errors.ErrInternal, "failed to list containers", err)
			}
		} else {
			// Show stats for specified containers
			for _, nameOrID := range args {
				container, err := loadContainerByNameOrID(store, nameOrID)
				if err != nil {
					return err
				}
				containers = append(containers, container)
			}
		}

		if len(containers) == 0 {
			fmt.Println("No running containers")
			return nil
		}

		// Print header
		fmt.Printf("%-15s %-15s %-10s %-10s %-10s %-10s %-10s\n",
			"CONTAINER ID", "NAME", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O")

		// TODO: Implement actual stats collection from cgroups
		for _, c := range containers {
			containerID := c.ID
			if len(containerID) > 12 {
				containerID = containerID[:12]
			}

			fmt.Printf("%-15s %-15s %-10s %-10s %-10s %-10s %-10s\n",
				containerID, c.Name, "--", "--", "--", "--", "--")
		}

		fmt.Println("\n(Real-time stats collection not yet implemented)")
		return nil
	},
}

func init() {
	statsCmd.Flags().BoolVarP(&statsAll, "all", "a", false, "Show all containers (default shows just running)")
	statsCmd.Flags().BoolVar(&statsNoStream, "no-stream", false, "Disable streaming stats and only pull the first result")
}

// top command
var topCmd = &cobra.Command{
	Use:   "top CONTAINER [ps OPTIONS]",
	Short: "Display the running processes of a container",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		containerID := args[0]
		container, err := loadContainerByNameOrID(store, containerID)
		if err != nil {
			return err
		}

		fmt.Printf("Processes in container %s:\n", container.Name)
		fmt.Println("UID\tPID\tPPID\tC\tSTIME\tTTY\tTIME\tCMD")
		fmt.Println("(Process listing not yet implemented)")

		// TODO: Read from /proc to list processes in container's PID namespace
		return nil
	},
}
