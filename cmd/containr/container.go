package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/state"
)

// create command
var createCmd = &cobra.Command{
	Use:   "create [flags] IMAGE [COMMAND] [ARG...]",
	Short: "Create a new container",
	Long:  "Creates a container but does not start it.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Create command - Phase 2 implementation")
		fmt.Println("This command creates a container without starting it")
		fmt.Printf("Image: %s\n", args[0])
		if len(args) > 1 {
			fmt.Printf("Command: %v\n", args[1:])
		}
		return nil
	},
}

// start command
var startCmd = &cobra.Command{
	Use:   "start CONTAINER [CONTAINER...]",
	Short: "Start one or more stopped containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("start")

		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		for _, nameOrID := range args {
			log.Infof("Starting container %s", nameOrID)

			// Load container state
			container, err := loadContainerByNameOrID(store, nameOrID)
			if err != nil {
				return err
			}

			if container.State == state.StateRunning {
				fmt.Printf("Container %s is already running\n", nameOrID)
				continue
			}

			// TODO: Implement actual container start logic
			fmt.Printf("Starting container %s...\n", container.Name)

			// Update state
			container.State = state.StateRunning
			container.Started = time.Now()
			if err := store.Save(container); err != nil {
				return errors.Wrap(errors.ErrInternal, "failed to update container state", err)
			}

			fmt.Printf("Container %s started\n", container.Name)
		}

		return nil
	},
}

// stop command
var stopCmd = &cobra.Command{
	Use:   "stop CONTAINER [CONTAINER...]",
	Short: "Stop one or more running containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("stop")

		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		for _, nameOrID := range args {
			log.Infof("Stopping container %s", nameOrID)

			// Load container state
			container, err := loadContainerByNameOrID(store, nameOrID)
			if err != nil {
				return err
			}

			if container.State != state.StateRunning {
				fmt.Printf("Container %s is not running\n", nameOrID)
				continue
			}

			// TODO: Implement actual container stop logic (send SIGTERM, wait, SIGKILL)
			fmt.Printf("Stopping container %s...\n", container.Name)

			// Update state
			container.State = state.StateStopped
			container.Finished = time.Now()
			if err := store.Save(container); err != nil {
				return errors.Wrap(errors.ErrInternal, "failed to update container state", err)
			}

			fmt.Printf("Container %s stopped\n", container.Name)
		}

		return nil
	},
}

// rm command
var (
	rmForce bool
	rmAll   bool
)

var rmCmd = &cobra.Command{
	Use:   "rm [flags] CONTAINER [CONTAINER...]",
	Short: "Remove one or more containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("rm")

		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		for _, nameOrID := range args {
			log.Infof("Removing container %s", nameOrID)

			// Load container state
			container, err := loadContainerByNameOrID(store, nameOrID)
			if err != nil {
				if !rmForce {
					return err
				}
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				continue
			}

			// Check if container is running
			if container.State == state.StateRunning && !rmForce {
				return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("container %s is running", nameOrID)).
					WithHint("Use --force to remove a running container")
			}

			// TODO: Implement cleanup (unmount volumes, remove rootfs, etc.)

			// Delete state
			if err := store.Delete(container.ID); err != nil {
				return errors.Wrap(errors.ErrInternal, "failed to remove container", err).
					WithField("container_id", container.ID)
			}

			fmt.Printf("Removed container %s\n", container.Name)
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force removal of running containers")
	rmCmd.Flags().BoolVarP(&rmAll, "all", "a", false, "Remove all containers")
}

// ps command
var (
	psAll    bool
	psQuiet  bool
	psNoTrunc bool
)

var psCmd = &cobra.Command{
	Use:   "ps [flags]",
	Short: "List containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := state.NewStore(stateDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
		}

		var containers []*state.Container
		if psAll {
			containers, err = store.List()
		} else {
			containers, err = store.ListByState(state.StateRunning)
		}

		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to list containers", err)
		}

		if psQuiet {
			// Just print IDs
			for _, c := range containers {
				fmt.Println(c.ID)
			}
			return nil
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tNAME")

		for _, c := range containers {
			containerID := c.ID
			if !psNoTrunc && len(containerID) > 12 {
				containerID = containerID[:12]
			}

			image := c.Image
			if image == "" {
				image = "<none>"
			}

			command := formatCommand(c.Command)
			created := formatDuration(time.Since(c.Created))
			status := formatStatus(c)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				containerID, image, command, created, status, c.Name)
		}

		w.Flush()
		return nil
	},
}

func init() {
	psCmd.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers (default shows just running)")
	psCmd.Flags().BoolVarP(&psQuiet, "quiet", "q", false, "Only display container IDs")
	psCmd.Flags().BoolVar(&psNoTrunc, "no-trunc", false, "Don't truncate output")
}

// logs command
var (
	logsFollow bool
	logsTail   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [flags] CONTAINER",
	Short: "Fetch the logs of a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Logs for container %s:\n", args[0])
		fmt.Println("(Log collection not yet implemented)")
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", -1, "Number of lines to show from the end of the logs")
}

// exec command
var (
	execDetach      bool
	execInteractive bool
	execTTY         bool
	execUser        string
	execWorkdir     string
)

var execCmd = &cobra.Command{
	Use:   "exec [flags] CONTAINER COMMAND [ARG...]",
	Short: "Execute a command in a running container",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerID := args[0]
		command := args[1:]

		fmt.Printf("Executing command in container %s: %v\n", containerID, command)
		fmt.Println("(Exec not yet fully implemented)")
		return nil
	},
}

func init() {
	execCmd.Flags().BoolVarP(&execDetach, "detach", "d", false, "Detached mode")
	execCmd.Flags().BoolVarP(&execInteractive, "interactive", "i", false, "Keep STDIN open")
	execCmd.Flags().BoolVarP(&execTTY, "tty", "t", false, "Allocate a pseudo-TTY")
	execCmd.Flags().StringVarP(&execUser, "user", "u", "", "Username or UID")
	execCmd.Flags().StringVarP(&execWorkdir, "workdir", "w", "", "Working directory")
}

// Helper functions

func loadContainerByNameOrID(store *state.Store, nameOrID string) (*state.Container, error) {
	// Try to load by ID first
	container, err := store.Load(nameOrID)
	if err == nil {
		return container, nil
	}

	// Try to load by name
	container, err = store.FindByName(nameOrID)
	if err != nil {
		return nil, errors.New(errors.ErrContainerNotFound, fmt.Sprintf("container '%s' not found", nameOrID)).
			WithField("container", nameOrID)
	}

	return container, nil
}

func formatCommand(command []string) string {
	if len(command) == 0 {
		return ""
	}
	if len(command) == 1 {
		return command[0]
	}
	return fmt.Sprintf("%s ...", command[0])
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	}
	return fmt.Sprintf("%d days ago", int(d.Hours()/24))
}

func formatStatus(c *state.Container) string {
	switch c.State {
	case state.StateRunning:
		if !c.Started.IsZero() {
			return fmt.Sprintf("Up %s", formatDuration(time.Since(c.Started)))
		}
		return "Up"
	case state.StateExited:
		return fmt.Sprintf("Exited (%d)", c.ExitCode)
	case state.StateStopped:
		return "Stopped"
	case state.StateCreated:
		return "Created"
	default:
		return string(c.State)
	}
}
