package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// Network represents a container network
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`  // "bridge", "host", "none"
	Subnet     string            `json:"subnet"`  // e.g., "172.16.0.0/24"
	Gateway    string            `json:"gateway"` // e.g., "172.16.0.1"
	Bridge     string            `json:"bridge"`  // Bridge interface name
	Created    time.Time         `json:"created"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
	Containers []string          `json:"containers"` // Container IDs using this network
}

// NetworkManager manages container networks
type NetworkManager struct {
	networks    map[string]*Network
	stateDir    string
	mu          sync.RWMutex
	ipAllocator *IPAllocator
}

// IPAllocator manages IP address allocation for networks
type IPAllocator struct {
	allocated map[string]bool // IP -> allocated
	subnet    *net.IPNet
	gateway   net.IP
	mu        sync.Mutex
}

// NewIPAllocator creates a new IP allocator for a subnet
func NewIPAllocator(subnet, gateway string) (*IPAllocator, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	gw := net.ParseIP(gateway)
	if gw == nil {
		return nil, fmt.Errorf("invalid gateway IP: %s", gateway)
	}

	return &IPAllocator{
		allocated: make(map[string]bool),
		subnet:    ipnet,
		gateway:   gw,
	}, nil
}

// AllocateIP allocates the next available IP address
func (a *IPAllocator) AllocateIP() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Start from the first IP after gateway
	ip := make(net.IP, len(a.subnet.IP))
	copy(ip, a.subnet.IP)

	// Skip network address and gateway
	for i := 2; i < 254; i++ {
		ip[len(ip)-1] = byte(i)
		if !a.subnet.Contains(ip) {
			continue
		}

		ipStr := ip.String()
		if ipStr == a.gateway.String() {
			continue
		}

		if !a.allocated[ipStr] {
			a.allocated[ipStr] = true
			return ipStr + "/24", nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in subnet")
}

// ReleaseIP releases an allocated IP address
func (a *IPAllocator) ReleaseIP(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.allocated, ip)
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(stateDir string) (*NetworkManager, error) {
	log := logger.New("network-manager")

	nm := &NetworkManager{
		networks: make(map[string]*Network),
		stateDir: filepath.Join(stateDir, "networks"),
	}

	// Create networks directory
	if err := os.MkdirAll(nm.stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create networks directory: %w", err)
	}

	// Load existing networks
	if err := nm.loadNetworks(); err != nil {
		log.Warnf("Failed to load existing networks: %v", err)
	}

	// Ensure default bridge network exists
	if _, exists := nm.networks["bridge"]; !exists {
		log.Info("Creating default bridge network")
		if _, err := nm.CreateNetwork("bridge", "bridge", "172.16.0.0/24", "172.16.0.1", nil); err != nil {
			log.Warnf("Failed to create default bridge network: %v", err)
		}
	}

	return nm, nil
}

// CreateNetwork creates a new network
func (nm *NetworkManager) CreateNetwork(name, driver, subnet, gateway string, labels map[string]string) (*Network, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	log := logger.New("network-manager")

	// Check if network already exists
	for _, net := range nm.networks {
		if net.Name == name {
			return nil, fmt.Errorf("network with name %s already exists", name)
		}
	}

	// Generate network ID
	id := generateID(name)

	// Create IP allocator
	allocator, err := NewIPAllocator(subnet, gateway)
	if err != nil {
		return nil, err
	}

	network := &Network{
		ID:         id,
		Name:       name,
		Driver:     driver,
		Subnet:     subnet,
		Gateway:    gateway,
		Bridge:     fmt.Sprintf("cbr-%s", id[:8]),
		Created:    time.Now(),
		Labels:     labels,
		Options:    make(map[string]string),
		Containers: []string{},
	}

	// Setup bridge if driver is bridge
	if driver == "bridge" {
		log.Infof("Setting up bridge %s for network %s", network.Bridge, name)
		if err := SetupBridge(network.Bridge, gateway); err != nil {
			return nil, fmt.Errorf("failed to setup bridge: %w", err)
		}

		// Setup NAT
		if err := SetupNAT(network.Bridge); err != nil {
			log.Warnf("Failed to setup NAT: %v", err)
		}
	}

	nm.networks[id] = network
	nm.ipAllocator = allocator

	// Persist network
	if err := nm.saveNetwork(network); err != nil {
		log.Warnf("Failed to persist network: %v", err)
	}

	log.Infof("Created network: %s (ID: %s, Subnet: %s)", name, id, subnet)
	return network, nil
}

// GetNetwork retrieves a network by name or ID
func (nm *NetworkManager) GetNetwork(nameOrID string) (*Network, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Try by ID first
	if net, exists := nm.networks[nameOrID]; exists {
		return net, nil
	}

	// Try by name
	for _, net := range nm.networks {
		if net.Name == nameOrID {
			return net, nil
		}
	}

	return nil, fmt.Errorf("network not found: %s", nameOrID)
}

// ListNetworks returns all networks
func (nm *NetworkManager) ListNetworks() []*Network {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	networks := make([]*Network, 0, len(nm.networks))
	for _, net := range nm.networks {
		networks = append(networks, net)
	}
	return networks
}

// RemoveNetwork removes a network
func (nm *NetworkManager) RemoveNetwork(nameOrID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	log := logger.New("network-manager")

	net, err := nm.GetNetwork(nameOrID)
	if err != nil {
		return err
	}

	// Check if network is in use
	if len(net.Containers) > 0 {
		return fmt.Errorf("network is in use by %d container(s)", len(net.Containers))
	}

	// Remove bridge if it exists
	if net.Driver == "bridge" {
		log.Infof("Removing bridge %s", net.Bridge)
		// Best effort cleanup
		_ = removeBridge(net.Bridge)
	}

	delete(nm.networks, net.ID)

	// Remove persisted state
	statePath := filepath.Join(nm.stateDir, net.ID+".json")
	if err := os.Remove(statePath); err != nil {
		log.Warnf("Failed to remove network state file: %v", err)
	}

	log.Infof("Removed network: %s", net.Name)
	return nil
}

// saveNetwork persists a network to disk
func (nm *NetworkManager) saveNetwork(net *Network) error {
	data, err := json.MarshalIndent(net, "", "  ")
	if err != nil {
		return err
	}

	statePath := filepath.Join(nm.stateDir, net.ID+".json")
	return os.WriteFile(statePath, data, 0644)
}

// loadNetworks loads all persisted networks
func (nm *NetworkManager) loadNetworks() error {
	files, err := os.ReadDir(nm.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(nm.stateDir, file.Name()))
		if err != nil {
			continue
		}

		var net Network
		if err := json.Unmarshal(data, &net); err != nil {
			continue
		}

		nm.networks[net.ID] = &net
	}

	return nil
}

// removeBridge removes a bridge interface
func removeBridge(bridgeName string) error {
	// Bring bridge down
	exec.Command("ip", "link", "set", bridgeName, "down").Run()

	// Delete bridge
	cmd := exec.Command("ip", "link", "delete", bridgeName)
	return cmd.Run()
}

// generateID generates a simple ID from a name
func generateID(name string) string {
	return fmt.Sprintf("%s-%d", name, time.Now().Unix())
}
