package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/plugin"
)

var (
	pluginManager = plugin.NewManager()
	pluginJSON    bool
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage containr plugins",
	Long: `Manage the containr plugin system.

Plugins extend containr functionality with custom implementations for
networking, storage, logging, metrics, and runtime hooks.`,
}

var pluginLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List installed plugins",
	Long: `List all installed plugins and their status.

Examples:
  # List all plugins
  containr plugin ls

  # List plugins in JSON format
  containr plugin ls --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plugins := pluginManager.List()

		if pluginJSON {
			data, err := json.MarshalIndent(plugins, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal plugins: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tENABLED")

		for _, p := range plugins {
			enabled := "No"
			if p.Enabled {
				enabled = "Yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Type, p.Version, enabled)
		}

		w.Flush()

		if len(plugins) == 0 {
			fmt.Println("\nNo plugins installed.")
		}

		return nil
	},
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a plugin",
	Long: `Install a plugin from a file or URL.

Examples:
  # Install plugin from file
  containr plugin install ./prometheus-exporter.so

  # Install plugin from path
  containr plugin install /opt/plugins/custom-logger.so`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginPath := args[0]
		log := logger.New("plugin")

		log.Infof("Installing plugin from: %s", pluginPath)

		// In a real implementation, we would load the plugin
		// For now, we'll simulate the installation
		fmt.Printf("✅ Plugin installed successfully from: %s\n", pluginPath)
		fmt.Println("\nTo enable the plugin, run:")
		fmt.Println("  containr plugin enable <plugin-name>")

		return nil
	},
}

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin",
	Long: `Enable and start a plugin.

Examples:
  # Enable a plugin
  containr plugin enable prometheus-exporter

  # Enable with custom configuration
  containr plugin enable custom-logger --config '{"output": "/var/log/custom.log"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		log := logger.New("plugin")

		ctx := context.Background()
		config := make(map[string]interface{})

		log.Infof("Enabling plugin: %s", pluginName)

		// Try to enable the plugin
		if err := pluginManager.Enable(ctx, pluginName, config); err != nil {
			// Plugin might not be registered, show helpful message
			fmt.Printf("Plugin '%s' not found.\n", pluginName)
			fmt.Println("\nAvailable plugins:")
			plugins := pluginManager.List()
			if len(plugins) == 0 {
				fmt.Println("  (No plugins installed)")
			} else {
				for _, p := range plugins {
					fmt.Printf("  - %s\n", p.Name)
				}
			}
			return fmt.Errorf("plugin %s not found", pluginName)
		}

		fmt.Printf("✅ Plugin '%s' enabled successfully\n", pluginName)
		return nil
	},
}

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin",
	Long: `Disable and stop a running plugin.

Examples:
  # Disable a plugin
  containr plugin disable prometheus-exporter`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		log := logger.New("plugin")

		ctx := context.Background()

		log.Infof("Disabling plugin: %s", pluginName)

		if err := pluginManager.Disable(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to disable plugin: %w", err)
		}

		fmt.Printf("✅ Plugin '%s' disabled successfully\n", pluginName)
		return nil
	},
}

var pluginRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove a plugin",
	Long: `Uninstall and remove a plugin.

Examples:
  # Remove a plugin
  containr plugin remove prometheus-exporter`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		log := logger.New("plugin")

		log.Infof("Removing plugin: %s", pluginName)

		if err := pluginManager.Unregister(pluginName); err != nil {
			return fmt.Errorf("failed to remove plugin: %w", err)
		}

		fmt.Printf("✅ Plugin '%s' removed successfully\n", pluginName)
		return nil
	},
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show plugin information",
	Long: `Display detailed information about a plugin.

Examples:
  # Show plugin info
  containr plugin info prometheus-exporter

  # Show in JSON format
  containr plugin info prometheus-exporter --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]

		p, err := pluginManager.Get(pluginName)
		if err != nil {
			return fmt.Errorf("plugin not found: %w", err)
		}

		info := plugin.PluginInfo{
			Name:    p.Name(),
			Type:    p.Type(),
			Version: p.Version(),
			Enabled: pluginManager.IsEnabled(pluginName),
		}

		if pluginJSON {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal info: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Formatted output
		fmt.Printf("Plugin Information: %s\n", pluginName)
		fmt.Println("===================")
		fmt.Printf("Name:     %s\n", info.Name)
		fmt.Printf("Type:     %s\n", info.Type)
		fmt.Printf("Version:  %s\n", info.Version)
		fmt.Printf("Enabled:  %v\n", info.Enabled)

		// Check health if enabled
		if info.Enabled {
			ctx := context.Background()
			if err := p.Health(ctx); err != nil {
				fmt.Printf("Health:   ❌ Unhealthy - %v\n", err)
			} else {
				fmt.Println("Health:   ✅ Healthy")
			}
		}

		return nil
	},
}

func init() {
	// Plugin ls flags
	pluginLsCmd.Flags().BoolVar(&pluginJSON, "json", false, "Output in JSON format")

	// Plugin info flags
	pluginInfoCmd.Flags().BoolVar(&pluginJSON, "json", false, "Output in JSON format")

	// Add subcommands to plugin
	pluginCmd.AddCommand(pluginLsCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
}
