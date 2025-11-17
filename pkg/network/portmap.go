package network

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// PortMapping represents a port mapping configuration
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string // "tcp" or "udp"
	HostIP        string // Optional: bind to specific IP (default: 0.0.0.0)
}

// ParsePortMapping parses a port mapping string like "8080:80/tcp" or "8080:80"
func ParsePortMapping(portStr string) (*PortMapping, error) {
	log := logger.New("portmap")

	// Default values
	pm := &PortMapping{
		Protocol: "tcp",
		HostIP:   "0.0.0.0",
	}

	// Split by protocol
	parts := strings.Split(portStr, "/")
	portPart := parts[0]
	if len(parts) > 1 {
		pm.Protocol = strings.ToLower(parts[1])
		if pm.Protocol != "tcp" && pm.Protocol != "udp" {
			return nil, fmt.Errorf("invalid protocol: %s (must be tcp or udp)", pm.Protocol)
		}
	}

	// Parse ports (host:container or just port)
	portParts := strings.Split(portPart, ":")
	if len(portParts) == 1 {
		// Same port for host and container
		port, err := strconv.Atoi(portParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", portParts[0])
		}
		pm.HostPort = port
		pm.ContainerPort = port
	} else if len(portParts) == 2 {
		// Different host and container ports
		hostPort, err := strconv.Atoi(portParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid host port: %s", portParts[0])
		}
		containerPort, err := strconv.Atoi(portParts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid container port: %s", portParts[1])
		}
		pm.HostPort = hostPort
		pm.ContainerPort = containerPort
	} else {
		return nil, fmt.Errorf("invalid port mapping format: %s", portStr)
	}

	log.Debugf("Parsed port mapping: %s -> %+v", portStr, pm)
	return pm, nil
}

// SetupPortMapping sets up iptables rules for port forwarding
func SetupPortMapping(pm *PortMapping, containerIP string) error {
	log := logger.New("portmap")
	log.Infof("Setting up port mapping: %s:%d -> %s:%d/%s",
		pm.HostIP, pm.HostPort, containerIP, pm.ContainerPort, pm.Protocol)

	// Add DNAT rule to forward incoming traffic to container
	cmd := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING",
		"-p", pm.Protocol,
		"-d", pm.HostIP,
		"--dport", strconv.Itoa(pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add DNAT rule: %w (output: %s)", err, string(output))
	}

	// Add MASQUERADE rule for locally generated traffic
	cmd = exec.Command("iptables", "-t", "nat", "-A", "OUTPUT",
		"-p", pm.Protocol,
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add OUTPUT rule: %w (output: %s)", err, string(output))
	}

	log.Infof("Port mapping configured successfully")
	return nil
}

// RemovePortMapping removes iptables rules for port forwarding
func RemovePortMapping(pm *PortMapping, containerIP string) error {
	log := logger.New("portmap")
	log.Infof("Removing port mapping: %s:%d -> %s:%d/%s",
		pm.HostIP, pm.HostPort, containerIP, pm.ContainerPort, pm.Protocol)

	// Remove DNAT rule
	cmd := exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
		"-p", pm.Protocol,
		"-d", pm.HostIP,
		"--dport", strconv.Itoa(pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort))

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warnf("Failed to remove DNAT rule: %v (output: %s)", err, string(output))
	}

	// Remove OUTPUT rule
	cmd = exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
		"-p", pm.Protocol,
		"-d", "127.0.0.1",
		"--dport", strconv.Itoa(pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", containerIP, pm.ContainerPort))

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warnf("Failed to remove OUTPUT rule: %v (output: %s)", err, string(output))
	}

	return nil
}

// SetupPortMappings sets up multiple port mappings
func SetupPortMappings(mappings []*PortMapping, containerIP string) error {
	for _, pm := range mappings {
		if err := SetupPortMapping(pm, containerIP); err != nil {
			return err
		}
	}
	return nil
}

// RemovePortMappings removes multiple port mappings
func RemovePortMappings(mappings []*PortMapping, containerIP string) {
	for _, pm := range mappings {
		RemovePortMapping(pm, containerIP)
	}
}
