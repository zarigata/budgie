package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

// Network represents a container network
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	Labels     map[string]string `json:"labels,omitempty"`
	Containers []string          `json:"containers"`
}

// NetworkManager manages container networks
type NetworkManager struct {
	networks  map[string]*Network
	dataDir   string
	statePath string
	mu        sync.RWMutex
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(dataDir string) (*NetworkManager, error) {
	nm := &NetworkManager{
		networks:  make(map[string]*Network),
		dataDir:   dataDir,
		statePath: filepath.Join(dataDir, "networks.json"),
	}

	// Ensure networks directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create network data directory: %w", err)
	}

	// Load existing networks
	if err := nm.loadState(); err != nil {
		logrus.Warnf("Failed to load network state (starting fresh): %v", err)
	}

	// Create default network if not exists
	if _, exists := nm.networks["budgie0"]; !exists {
		if err := nm.CreateNetwork("budgie0", "bridge", "172.20.0.0/16", "172.20.0.1"); err != nil {
			logrus.Warnf("Failed to create default network: %v", err)
		}
	}

	return nm, nil
}

// CreateNetwork creates a new network
func (nm *NetworkManager) CreateNetwork(name, driver, subnet, gateway string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.networks[name]; exists {
		return fmt.Errorf("network already exists: %s", name)
	}

	// Validate subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	// Validate gateway
	gwIP := net.ParseIP(gateway)
	if gwIP == nil {
		return fmt.Errorf("invalid gateway IP: %s", gateway)
	}

	if !ipNet.Contains(gwIP) {
		return fmt.Errorf("gateway %s is not within subnet %s", gateway, subnet)
	}

	network := &Network{
		ID:         generateNetworkID(),
		Name:       name,
		Driver:     driver,
		Subnet:     subnet,
		Gateway:    gateway,
		Labels:     make(map[string]string),
		Containers: []string{},
	}

	nm.networks[name] = network

	if err := nm.saveState(); err != nil {
		logrus.Errorf("Failed to save network state: %v", err)
	}

	logrus.Infof("Created network %s (%s)", name, subnet)
	return nil
}

// RemoveNetwork removes a network
func (nm *NetworkManager) RemoveNetwork(name string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	network, exists := nm.networks[name]
	if !exists {
		return fmt.Errorf("network not found: %s", name)
	}

	if len(network.Containers) > 0 {
		return fmt.Errorf("network %s is in use by %d containers", name, len(network.Containers))
	}

	if name == "budgie0" {
		return fmt.Errorf("cannot remove default network")
	}

	delete(nm.networks, name)

	if err := nm.saveState(); err != nil {
		logrus.Errorf("Failed to save network state: %v", err)
	}

	logrus.Infof("Removed network %s", name)
	return nil
}

// GetNetwork returns a network by name
func (nm *NetworkManager) GetNetwork(name string) (*Network, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	network, exists := nm.networks[name]
	if !exists {
		return nil, fmt.Errorf("network not found: %s", name)
	}

	return network, nil
}

// ListNetworks returns all networks
func (nm *NetworkManager) ListNetworks() []*Network {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	list := make([]*Network, 0, len(nm.networks))
	for _, network := range nm.networks {
		list = append(list, network)
	}
	return list
}

// ConnectContainer connects a container to a network
func (nm *NetworkManager) ConnectContainer(networkName, containerID string) (*ContainerNetworkInfo, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	network, exists := nm.networks[networkName]
	if !exists {
		return nil, fmt.Errorf("network not found: %s", networkName)
	}

	// Check if already connected
	for _, ctrID := range network.Containers {
		if ctrID == containerID {
			return nil, fmt.Errorf("container already connected to network %s", networkName)
		}
	}

	// Allocate IP address
	ip, err := nm.allocateIP(network)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	network.Containers = append(network.Containers, containerID)

	if err := nm.saveState(); err != nil {
		logrus.Errorf("Failed to save network state: %v", err)
	}

	return &ContainerNetworkInfo{
		NetworkID:   network.ID,
		NetworkName: network.Name,
		IPAddress:   ip,
		Gateway:     network.Gateway,
	}, nil
}

// DisconnectContainer disconnects a container from a network
func (nm *NetworkManager) DisconnectContainer(networkName, containerID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	network, exists := nm.networks[networkName]
	if !exists {
		return fmt.Errorf("network not found: %s", networkName)
	}

	found := false
	newContainers := make([]string, 0, len(network.Containers))
	for _, ctrID := range network.Containers {
		if ctrID == containerID {
			found = true
		} else {
			newContainers = append(newContainers, ctrID)
		}
	}

	if !found {
		return fmt.Errorf("container not connected to network %s", networkName)
	}

	network.Containers = newContainers

	if err := nm.saveState(); err != nil {
		logrus.Errorf("Failed to save network state: %v", err)
	}

	return nil
}

// ContainerNetworkInfo contains network info for a container
type ContainerNetworkInfo struct {
	NetworkID   string `json:"network_id"`
	NetworkName string `json:"network_name"`
	IPAddress   string `json:"ip_address"`
	Gateway     string `json:"gateway"`
}

func (nm *NetworkManager) allocateIP(network *Network) (string, error) {
	_, ipNet, err := net.ParseCIDR(network.Subnet)
	if err != nil {
		return "", err
	}

	// Start from .2 (gateway is typically .1)
	ip := ipNet.IP
	ip = incrementIP(ip)
	ip = incrementIP(ip)

	// Find an unused IP
	usedIPs := make(map[string]bool)
	usedIPs[network.Gateway] = true

	// TODO: Track allocated IPs per container
	// For now, allocate based on container count
	for i := 0; i < len(network.Containers); i++ {
		ip = incrementIP(ip)
	}

	if !ipNet.Contains(ip) {
		return "", fmt.Errorf("no available IPs in network %s", network.Name)
	}

	return ip.String(), nil
}

func incrementIP(ip net.IP) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] > 0 {
			break
		}
	}

	return result
}

func generateNetworkID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(i*17 + 42) // Simple deterministic ID for now
	}
	return fmt.Sprintf("%x", b)[:12]
}

func (nm *NetworkManager) loadState() error {
	data, err := os.ReadFile(nm.statePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var networks []*Network
	if err := json.Unmarshal(data, &networks); err != nil {
		return err
	}

	for _, network := range networks {
		nm.networks[network.Name] = network
	}

	return nil
}

func (nm *NetworkManager) saveState() error {
	list := make([]*Network, 0, len(nm.networks))
	for _, network := range nm.networks {
		list = append(list, network)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(nm.statePath, data, 0644)
}
