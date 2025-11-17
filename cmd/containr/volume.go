package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/volume"
)

// volume command (parent command)
var volumeCmd = &cobra.Command{
	Use:   "volume COMMAND",
	Short: "Manage volumes",
	Long:  "Commands to manage volumes for persistent data storage.",
}

// volume create command
var volumeCreateCmd = &cobra.Command{
	Use:   "create [flags] [VOLUME_NAME]",
	Short: "Create a volume",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("volume-create")

		volumeMgr, err := volume.NewManager(volumeDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
		}

		// Generate or use provided name
		name := ""
		if len(args) > 0 {
			name = args[0]
		} else {
			// Generate a random name if not provided
			name = fmt.Sprintf("volume-%d", os.Getpid())
		}

		log.Infof("Creating volume %s", name)

		// Create volume
		vol, err := volumeMgr.Create(name, nil)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume", err).
				WithField("name", name)
		}

		fmt.Println(vol.Name)
		log.Infof("Volume %s created at %s", vol.Name, vol.Source)
		return nil
	},
}

// volume ls command
var (
	volumeLsQuiet bool
)

var volumeLsCmd = &cobra.Command{
	Use:     "ls [flags]",
	Aliases: []string{"list"},
	Short:   "List volumes",
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeMgr, err := volume.NewManager(volumeDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
		}

		volumes, err := volumeMgr.List()
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to list volumes", err)
		}

		if volumeLsQuiet {
			// Just print names
			for _, vol := range volumes {
				fmt.Println(vol.Name)
			}
			return nil
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VOLUME NAME\tTYPE\tMOUNT POINT")

		for _, vol := range volumes {
			fmt.Fprintf(w, "%s\t%s\t%s\n", vol.Name, vol.Type, vol.Source)
		}

		w.Flush()
		return nil
	},
}

func init() {
	volumeLsCmd.Flags().BoolVarP(&volumeLsQuiet, "quiet", "q", false, "Only display volume names")
}

// volume rm command
var (
	volumeRmForce bool
)

var volumeRmCmd = &cobra.Command{
	Use:     "rm [flags] VOLUME [VOLUME...]",
	Aliases: []string{"remove"},
	Short:   "Remove one or more volumes",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("volume-rm")

		volumeMgr, err := volume.NewManager(volumeDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
		}

		for _, name := range args {
			log.Infof("Removing volume %s", name)

			if err := volumeMgr.Remove(name); err != nil {
				if volumeRmForce {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
					continue
				}
				return errors.Wrap(errors.ErrInternal, "failed to remove volume", err).
					WithField("name", name)
			}

			fmt.Println(name)
			log.Infof("Volume %s removed", name)
		}

		return nil
	},
}

func init() {
	volumeRmCmd.Flags().BoolVarP(&volumeRmForce, "force", "f", false, "Force removal")
}

// volume inspect command
var volumeInspectCmd = &cobra.Command{
	Use:   "inspect [flags] VOLUME [VOLUME...]",
	Short: "Display detailed information on one or more volumes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeMgr, err := volume.NewManager(volumeDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
		}

		var volumes []*volume.Volume
		for _, name := range args {
			vol, err := volumeMgr.Get(name)
			if err != nil {
				return errors.Wrap(errors.ErrInternal, "volume not found", err).
					WithField("name", name)
			}
			volumes = append(volumes, vol)
		}

		// JSON output
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(volumes); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to encode output", err)
		}

		return nil
	},
}

// volume prune command
var (
	volumePruneForce bool
)

var volumePruneCmd = &cobra.Command{
	Use:   "prune [flags]",
	Short: "Remove all unused volumes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !volumePruneForce {
			fmt.Print("WARNING! This will remove all volumes not used by at least one container.\n")
			fmt.Print("Are you sure you want to continue? [y/N] ")

			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		volumeMgr, err := volume.NewManager(volumeDir)
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create volume manager", err)
		}

		removed, err := volumeMgr.Prune()
		if err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to prune volumes", err)
		}

		if len(removed) == 0 {
			fmt.Println("No volumes removed")
		} else {
			fmt.Printf("Removed volumes:\n")
			for _, name := range removed {
				fmt.Println(name)
			}
			fmt.Printf("Total reclaimed space: (calculation not yet implemented)\n")
		}

		return nil
	},
}

func init() {
	volumePruneCmd.Flags().BoolVarP(&volumePruneForce, "force", "f", false, "Do not prompt for confirmation")

	// Add subcommands to volume command
	volumeCmd.AddCommand(volumeCreateCmd)
	volumeCmd.AddCommand(volumeLsCmd)
	volumeCmd.AddCommand(volumeRmCmd)
	volumeCmd.AddCommand(volumeInspectCmd)
	volumeCmd.AddCommand(volumePruneCmd)
}
