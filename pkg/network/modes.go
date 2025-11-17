package network

import (
	"fmt"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// NetworkMode represents different network isolation modes
type NetworkMode string

const (
	// NetworkModeNone - No networking
	NetworkModeNone NetworkMode = "none"
	// NetworkModeHost - Use host network (no isolation)
	NetworkModeHost NetworkMode = "host"
	// NetworkModeBridge - Bridge networking (default)
	NetworkModeBridge NetworkMode = "bridge"
	// NetworkModeContainer - Share network with another container
	NetworkModeContainer NetworkMode = "container"
)

// ParseNetworkMode parses a network mode string
func ParseNetworkMode(mode string) (NetworkMode, error) {
	switch mode {
	case "none":
		return NetworkModeNone, nil
	case "host":
		return NetworkModeHost, nil
	case "bridge", "":
		return NetworkModeBridge, nil
	default:
		// Check if it's container:<id> format
		if len(mode) > 10 && mode[:10] == "container:" {
			return NetworkModeContainer, nil
		}
		return "", fmt.Errorf("invalid network mode: %s (valid: none, host, bridge, container:<id>)", mode)
	}
}

// GetNetworkNamespaceFlags returns the appropriate namespace flags for the network mode
func GetNetworkNamespaceFlags(mode NetworkMode) (bool, error) {
	log := logger.New("network-mode")

	switch mode {
	case NetworkModeNone:
		// Create new network namespace but don't configure it
		log.Debug("Network mode: none - creating isolated network namespace")
		return true, nil
	case NetworkModeHost:
		// Don't create network namespace - use host network
		log.Debug("Network mode: host - using host network namespace")
		return false, nil
	case NetworkModeBridge:
		// Create new network namespace and configure bridge
		log.Debug("Network mode: bridge - creating bridge network")
		return true, nil
	case NetworkModeContainer:
		// Will join another container's network namespace
		log.Debug("Network mode: container - will share network namespace")
		return false, nil
	default:
		return false, fmt.Errorf("unknown network mode: %s", mode)
	}
}

// SetupNetworkForMode sets up networking based on the specified mode
func SetupNetworkForMode(mode NetworkMode, config *NetworkConfig, containerID string) error {
	log := logger.New("network-mode")
	log.Infof("Setting up network for mode: %s", mode)

	switch mode {
	case NetworkModeNone:
		// Just bring up loopback
		log.Debug("Setting up loopback only (network mode: none)")
		return SetupLoopback()

	case NetworkModeHost:
		// Nothing to do - using host network
		log.Debug("Using host network - no setup needed")
		return nil

	case NetworkModeBridge:
		// Full bridge network setup
		log.Debug("Setting up bridge network")
		if config == nil {
			return fmt.Errorf("network config required for bridge mode")
		}
		return SetupContainerNetwork(config)

	case NetworkModeContainer:
		// Network namespace sharing handled by container package
		log.Debug("Container network sharing - handled by namespace setup")
		return nil

	default:
		return fmt.Errorf("unsupported network mode: %s", mode)
	}
}

// ValidateNetworkMode validates the network mode and returns any error
func ValidateNetworkMode(mode NetworkMode, containerID string) error {
	switch mode {
	case NetworkModeNone, NetworkModeHost, NetworkModeBridge:
		return nil
	case NetworkModeContainer:
		if containerID == "" {
			return fmt.Errorf("container:<id> network mode requires container ID")
		}
		// TODO: Validate that the target container exists
		return nil
	default:
		return fmt.Errorf("invalid network mode: %s", mode)
	}
}
