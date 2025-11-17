package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/network"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage container networks",
	Long:  `Create and manage container networks for networking isolation and communication.`,
}

var networkCreateCmd = &cobra.Command{
	Use:   "create [OPTIONS] NETWORK",
	Short: "Create a network",
	Long:  `Create a new container network with the specified configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		networkName := args[0]

		// Get flags
		driver, _ := cmd.Flags().GetString("driver")
		subnet, _ := cmd.Flags().GetString("subnet")
		gateway, _ := cmd.Flags().GetString("gateway")
		labels, _ := cmd.Flags().GetStringToString("label")

		// Create network manager
		nm, err := network.NewNetworkManager(stateDir)
		if err != nil {
			return fmt.Errorf("failed to create network manager: %w", err)
		}

		// Create network
		net, err := nm.CreateNetwork(networkName, driver, subnet, gateway, labels)
		if err != nil {
			return fmt.Errorf("failed to create network: %w", err)
		}

		fmt.Printf("Network created: %s (ID: %s)\n", net.Name, net.ID)
		return nil
	},
}

var networkLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List networks",
	Long:  `List all container networks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		quiet, _ := cmd.Flags().GetBool("quiet")
		format, _ := cmd.Flags().GetString("format")

		// Create network manager
		nm, err := network.NewNetworkManager(stateDir)
		if err != nil {
			return fmt.Errorf("failed to create network manager: %w", err)
		}

		// List networks
		networks := nm.ListNetworks()

		if quiet {
			// Just print IDs
			for _, net := range networks {
				fmt.Println(net.ID)
			}
			return nil
		}

		if format == "json" {
			// JSON output
			data, err := json.MarshalIndent(networks, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSUBNET\tGATEWAY")

		for _, net := range networks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				net.ID[:12],
				net.Name,
				net.Driver,
				net.Subnet,
				net.Gateway,
			)
		}

		w.Flush()
		return nil
	},
}

var networkRmCmd = &cobra.Command{
	Use:   "rm NETWORK [NETWORK...]",
	Short: "Remove one or more networks",
	Long:  `Remove one or more container networks by name or ID.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create network manager
		nm, err := network.NewNetworkManager(stateDir)
		if err != nil {
			return fmt.Errorf("failed to create network manager: %w", err)
		}

		// Remove each network
		for _, nameOrID := range args {
			if err := nm.RemoveNetwork(nameOrID); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing network %s: %v\n", nameOrID, err)
				continue
			}
			fmt.Printf("Network removed: %s\n", nameOrID)
		}

		return nil
	},
}

var networkInspectCmd = &cobra.Command{
	Use:   "inspect NETWORK [NETWORK...]",
	Short: "Display detailed information on one or more networks",
	Long:  `Display detailed network information in JSON format.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create network manager
		nm, err := network.NewNetworkManager(stateDir)
		if err != nil {
			return fmt.Errorf("failed to create network manager: %w", err)
		}

		// Inspect each network
		results := make([]*network.Network, 0)
		for _, nameOrID := range args {
			net, err := nm.GetNetwork(nameOrID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error inspecting network %s: %v\n", nameOrID, err)
				continue
			}
			results = append(results, net)
		}

		// Output as JSON
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var networkConnectCmd = &cobra.Command{
	Use:   "connect NETWORK CONTAINER",
	Short: "Connect a container to a network",
	Long:  `Connect a running container to an existing network.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement network connect
		return fmt.Errorf("network connect not yet implemented")
	},
}

var networkDisconnectCmd = &cobra.Command{
	Use:   "disconnect NETWORK CONTAINER",
	Short: "Disconnect a container from a network",
	Long:  `Disconnect a container from a network.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement network disconnect
		return fmt.Errorf("network disconnect not yet implemented")
	},
}

func init() {
	// Network create flags
	networkCreateCmd.Flags().String("driver", "bridge", "Driver to manage the network (bridge, host, none)")
	networkCreateCmd.Flags().String("subnet", "172.18.0.0/24", "Subnet in CIDR format")
	networkCreateCmd.Flags().String("gateway", "172.18.0.1", "Gateway for the subnet")
	networkCreateCmd.Flags().StringToString("label", nil, "Set metadata on the network")

	// Network ls flags
	networkLsCmd.Flags().BoolP("quiet", "q", false, "Only display network IDs")
	networkLsCmd.Flags().String("format", "", "Format output (json)")

	// Add subcommands
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkLsCmd)
	networkCmd.AddCommand(networkRmCmd)
	networkCmd.AddCommand(networkInspectCmd)
	networkCmd.AddCommand(networkConnectCmd)
	networkCmd.AddCommand(networkDisconnectCmd)
}
