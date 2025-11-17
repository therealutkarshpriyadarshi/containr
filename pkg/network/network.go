package network

import (
	"fmt"
	"net"
	"os/exec"
)

// NetworkConfig holds network configuration for a container
type NetworkConfig struct {
	BridgeName    string
	ContainerIP   string
	GatewayIP     string
	VethHost      string
	VethContainer string
}

// SetupBridge creates a network bridge for container networking
func SetupBridge(bridgeName, gatewayIP string) error {
	// Check if bridge already exists
	if _, err := net.InterfaceByName(bridgeName); err == nil {
		return nil // Bridge already exists
	}

	// Create bridge
	cmd := exec.Command("ip", "link", "add", bridgeName, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	// Set bridge IP
	cmd = exec.Command("ip", "addr", "add", gatewayIP, "dev", bridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set bridge IP: %w", err)
	}

	// Bring bridge up
	cmd = exec.Command("ip", "link", "set", bridgeName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring bridge up: %w", err)
	}

	return nil
}

// CreateVethPair creates a virtual ethernet pair
func CreateVethPair(vethHost, vethContainer string) error {
	cmd := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}
	return nil
}

// AttachToBridge attaches a veth interface to a bridge
func AttachToBridge(bridgeName, vethName string) error {
	cmd := exec.Command("ip", "link", "set", vethName, "master", bridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach veth to bridge: %w", err)
	}

	// Bring veth up
	cmd = exec.Command("ip", "link", "set", vethName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring veth up: %w", err)
	}

	return nil
}

// MoveToNamespace moves a network interface to a network namespace
func MoveToNamespace(interfaceName string, pid int) error {
	cmd := exec.Command("ip", "link", "set", interfaceName, "netns", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move interface to namespace: %w", err)
	}
	return nil
}

// ConfigureInterface configures a network interface with an IP address
func ConfigureInterface(interfaceName, ipAddress string) error {
	// Set IP address
	cmd := exec.Command("ip", "addr", "add", ipAddress, "dev", interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", interfaceName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	return nil
}

// SetupLoopback sets up the loopback interface
func SetupLoopback() error {
	cmd := exec.Command("ip", "link", "set", "lo", "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring loopback up: %w", err)
	}
	return nil
}

// SetupDefaultRoute sets up the default route
func SetupDefaultRoute(gateway string) error {
	cmd := exec.Command("ip", "route", "add", "default", "via", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set default route: %w", err)
	}
	return nil
}

// SetupContainerNetwork sets up complete networking for a container
func SetupContainerNetwork(config *NetworkConfig) error {
	// Setup bridge
	if err := SetupBridge(config.BridgeName, config.GatewayIP); err != nil {
		return err
	}

	// Create veth pair
	if err := CreateVethPair(config.VethHost, config.VethContainer); err != nil {
		return err
	}

	// Attach host veth to bridge
	if err := AttachToBridge(config.BridgeName, config.VethHost); err != nil {
		return err
	}

	return nil
}

// EnableIPForwarding enables IP forwarding on the host
func EnableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	return nil
}

// SetupNAT sets up NAT for container networking
func SetupNAT(bridgeName string) error {
	// Add iptables rule for NAT
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "172.16.0.0/24", "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to setup NAT: %w", err)
	}

	// Allow forwarding
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", bridgeName, "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to setup forwarding: %w", err)
	}

	return nil
}
