package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/state"
	"github.com/therealutkarshpriyadarshi/containr/pkg/volume"
)

var (
	// Run command flags
	runName        string
	runHostname    string
	runMemory      string
	runCPUs        string
	runPIDs        int
	runNetwork     string
	runVolumes     []string
	runEnv         []string
	runWorkdir     string
	runUser        string
	runPrivileged  bool
	runDetach      bool
	runInteractive bool
	runTTY         bool
	runRM          bool
	runImage       string
)

var runCmd = &cobra.Command{
	Use:   "run [flags] IMAGE [COMMAND] [ARG...]",
	Short: "Run a command in a new container",
	Long: `Creates and starts a new container from an image.

If the image is not available locally, it will be pulled from the registry.`,
	Example: `  containr run alpine /bin/sh
  containr run --name mycontainer alpine:latest /bin/sh
  containr run -v /host/path:/container/path alpine /bin/sh
  containr run --memory 100m --cpus 0.5 alpine /bin/sh
  containr run -d --name webserver nginx
  containr run --rm alpine /bin/sh -c "echo hello"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runContainer,
}

func init() {
	runCmd.Flags().StringVar(&runName, "name", "", "Assign a name to the container")
	runCmd.Flags().StringVar(&runHostname, "hostname", "", "Container hostname")
	runCmd.Flags().StringVar(&runMemory, "memory", "", "Memory limit (e.g., 100m, 1g)")
	runCmd.Flags().StringVar(&runCPUs, "cpus", "", "Number of CPUs")
	runCmd.Flags().IntVar(&runPIDs, "pids", 0, "PID limit")
	runCmd.Flags().StringVar(&runNetwork, "network", "bridge", "Network mode (none, host, bridge)")
	runCmd.Flags().StringSliceVarP(&runVolumes, "volume", "v", []string{}, "Bind mount a volume (source:destination[:ro])")
	runCmd.Flags().StringSliceVarP(&runEnv, "env", "e", []string{}, "Set environment variables")
	runCmd.Flags().StringVarP(&runWorkdir, "workdir", "w", "", "Working directory inside the container")
	runCmd.Flags().StringVarP(&runUser, "user", "u", "", "Username or UID")
	runCmd.Flags().BoolVar(&runPrivileged, "privileged", false, "Run in privileged mode")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run container in background")
	runCmd.Flags().BoolVarP(&runInteractive, "interactive", "i", false, "Keep STDIN open")
	runCmd.Flags().BoolVarP(&runTTY, "tty", "t", false, "Allocate a pseudo-TTY")
	runCmd.Flags().BoolVar(&runRM, "rm", false, "Automatically remove container when it exits")
}

func runContainer(cmd *cobra.Command, args []string) error {
	log := logger.New("run")

	// Parse arguments
	imageName := args[0]
	command := []string{"/bin/sh"}
	if len(args) > 1 {
		command = args[1:]
	}

	// Generate container ID and name
	containerID := uuid.New().String()[:12]
	if runName == "" {
		runName = "containr-" + containerID
	}

	if runHostname == "" {
		runHostname = runName
	}

	log.Infof("Creating container %s (%s) from image %s", runName, containerID, imageName)

	// Create state store
	store, err := state.NewStore(stateDir)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create state store", err)
	}

	// Check if container with same name exists
	if existing, _ := store.FindByName(runName); existing != nil {
		return errors.New(errors.ErrContainerAlreadyExists, fmt.Sprintf("container with name '%s' already exists", runName)).
			WithField("name", runName).
			WithHint("Use a different name or remove the existing container with 'containr rm'")
	}

	// Parse volumes
	volumeMgr, err := volume.NewManager(volumeDir)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
	}

	var volumes []state.Volume
	for _, volStr := range runVolumes {
		vol, err := volume.ParseVolumeString(volStr, volumeMgr)
		if err != nil {
			return errors.Wrap(errors.ErrInvalidArgument, "invalid volume specification", err).
				WithField("volume", volStr)
		}

		volumes = append(volumes, state.Volume{
			Source:      vol.Source,
			Destination: vol.Destination,
			ReadOnly:    vol.ReadOnly,
			Type:        string(vol.Type),
		})
	}

	// Create container state
	containerState := &state.Container{
		ID:          containerID,
		Name:        runName,
		State:       state.StateCreated,
		Created:     time.Now(),
		Image:       imageName,
		Command:     command,
		Hostname:    runHostname,
		WorkingDir:  runWorkdir,
		Env:         runEnv,
		Volumes:     volumes,
		NetworkMode: runNetwork,
		Labels:      make(map[string]string),
	}

	// Save container state
	if err := store.Save(containerState); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to save container state", err)
	}

	// Create container configuration
	config := &container.Config{
		Command:    command,
		Hostname:   runHostname,
		Isolate:    runNetwork != "host",
		WorkingDir: runWorkdir,
		Privileged: runPrivileged,
	}

	// Create and run container
	c := container.New(config)

	log.WithField("container_id", c.ID).Info("Starting container")
	if !runDetach {
		fmt.Printf("Starting container %s...\n", runName)
	}

	// Update state to running
	containerState.State = state.StateRunning
	containerState.Started = time.Now()
	if err := store.Save(containerState); err != nil {
		log.WithError(err).Warn("Failed to update container state")
	}

	// Run container
	var runErr error
	if runDetach {
		// TODO: Implement background execution
		fmt.Printf("Detached mode not yet fully implemented\n")
		fmt.Printf("Container %s started\n", runName)
	} else {
		if debugMode {
			fmt.Printf("Namespaces: UTS, PID, Mount, IPC, Network\n")
			fmt.Printf("Command: %v\n", c.Command)
			fmt.Println("---")
		}

		runErr = c.RunWithSetup()
	}

	// Update state based on result
	if runErr != nil {
		containerState.State = state.StateExited
		containerState.ExitCode = 1
		containerState.Finished = time.Now()
		store.Save(containerState)

		if runRM {
			store.Delete(containerID)
		}

		return errors.Wrap(errors.ErrContainerStart, "container failed", runErr).
			WithField("container_id", containerID)
	}

	// Container completed successfully
	containerState.State = state.StateExited
	containerState.ExitCode = 0
	containerState.Finished = time.Now()
	if err := store.Save(containerState); err != nil {
		log.WithError(err).Warn("Failed to update container state")
	}

	if !runDetach {
		fmt.Printf("\nContainer %s exited\n", runName)
	}

	// Auto-remove if --rm flag is set
	if runRM {
		if err := store.Delete(containerID); err != nil {
			log.WithError(err).Warn("Failed to remove container")
		} else {
			log.Info("Container removed")
		}
	}

	return nil
}
