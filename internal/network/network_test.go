package network

import (
	"os"
	"testing"
)

func TestNetworkManager_CreateNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Test creating a network
	err = nm.CreateNetwork("test-net", "bridge", "172.25.0.0/16", "172.25.0.1")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	// Verify network exists
	net, err := nm.GetNetwork("test-net")
	if err != nil {
		t.Fatalf("Failed to get network: %v", err)
	}

	if net.Name != "test-net" {
		t.Errorf("Network name mismatch: got %s, want test-net", net.Name)
	}

	if net.Subnet != "172.25.0.0/16" {
		t.Errorf("Subnet mismatch: got %s, want 172.25.0.0/16", net.Subnet)
	}

	if net.Gateway != "172.25.0.1" {
		t.Errorf("Gateway mismatch: got %s, want 172.25.0.1", net.Gateway)
	}

	if net.Driver != "bridge" {
		t.Errorf("Driver mismatch: got %s, want bridge", net.Driver)
	}
}

func TestNetworkManager_DefaultNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Default network should be created
	net, err := nm.GetNetwork("budgie0")
	if err != nil {
		t.Fatalf("Default network not created: %v", err)
	}

	if net.Subnet != "172.20.0.0/16" {
		t.Errorf("Default subnet mismatch: got %s, want 172.20.0.0/16", net.Subnet)
	}
}

func TestNetworkManager_DuplicateNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	err = nm.CreateNetwork("dup-net", "bridge", "172.26.0.0/16", "172.26.0.1")
	if err != nil {
		t.Fatalf("Failed to create first network: %v", err)
	}

	// Try to create duplicate - should fail
	err = nm.CreateNetwork("dup-net", "bridge", "172.27.0.0/16", "172.27.0.1")
	if err == nil {
		t.Error("Expected error when creating duplicate network, got nil")
	}
}

func TestNetworkManager_InvalidSubnet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Invalid subnet format
	err = nm.CreateNetwork("invalid-net", "bridge", "not-a-subnet", "172.28.0.1")
	if err == nil {
		t.Error("Expected error for invalid subnet, got nil")
	}
}

func TestNetworkManager_GatewayOutsideSubnet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Gateway outside subnet
	err = nm.CreateNetwork("outside-gw", "bridge", "172.29.0.0/16", "192.168.1.1")
	if err == nil {
		t.Error("Expected error for gateway outside subnet, got nil")
	}
}

func TestNetworkManager_RemoveNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	err = nm.CreateNetwork("remove-net", "bridge", "172.30.0.0/16", "172.30.0.1")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	err = nm.RemoveNetwork("remove-net")
	if err != nil {
		t.Fatalf("Failed to remove network: %v", err)
	}

	// Verify it's gone
	_, err = nm.GetNetwork("remove-net")
	if err == nil {
		t.Error("Expected error when getting removed network, got nil")
	}
}

func TestNetworkManager_CannotRemoveDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Should not be able to remove default network
	err = nm.RemoveNetwork("budgie0")
	if err == nil {
		t.Error("Expected error when removing default network, got nil")
	}
}

func TestNetworkManager_ConnectContainer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	containerID := "container123"

	info, err := nm.ConnectContainer("budgie0", containerID)
	if err != nil {
		t.Fatalf("Failed to connect container: %v", err)
	}

	if info.NetworkName != "budgie0" {
		t.Errorf("Network name mismatch: got %s, want budgie0", info.NetworkName)
	}

	if info.IPAddress == "" {
		t.Error("IP address should not be empty")
	}

	if info.Gateway != "172.20.0.1" {
		t.Errorf("Gateway mismatch: got %s, want 172.20.0.1", info.Gateway)
	}

	// Verify container is in network
	net, _ := nm.GetNetwork("budgie0")
	found := false
	for _, ctr := range net.Containers {
		if ctr == containerID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Container not found in network")
	}
}

func TestNetworkManager_DisconnectContainer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	containerID := "container456"

	_, err = nm.ConnectContainer("budgie0", containerID)
	if err != nil {
		t.Fatalf("Failed to connect container: %v", err)
	}

	err = nm.DisconnectContainer("budgie0", containerID)
	if err != nil {
		t.Fatalf("Failed to disconnect container: %v", err)
	}

	// Verify container is not in network
	net, _ := nm.GetNetwork("budgie0")
	for _, ctr := range net.Containers {
		if ctr == containerID {
			t.Error("Container should not be in network after disconnect")
		}
	}
}

func TestNetworkManager_ListNetworks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	// Create additional networks
	nm.CreateNetwork("net1", "bridge", "172.31.0.0/16", "172.31.0.1")
	nm.CreateNetwork("net2", "bridge", "172.32.0.0/16", "172.32.0.1")

	networks := nm.ListNetworks()

	// Should have default + 2 created
	if len(networks) != 3 {
		t.Errorf("Expected 3 networks, got %d", len(networks))
	}

	names := make(map[string]bool)
	for _, n := range networks {
		names[n.Name] = true
	}

	for _, expected := range []string{"budgie0", "net1", "net2"} {
		if !names[expected] {
			t.Errorf("Network %s not found in list", expected)
		}
	}
}

func TestNetworkManager_CannotRemoveNetworkWithContainers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nm, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create network manager: %v", err)
	}

	err = nm.CreateNetwork("busy-net", "bridge", "172.33.0.0/16", "172.33.0.1")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	// Connect a container
	_, err = nm.ConnectContainer("busy-net", "container789")
	if err != nil {
		t.Fatalf("Failed to connect container: %v", err)
	}

	// Try to remove - should fail
	err = nm.RemoveNetwork("busy-net")
	if err == nil {
		t.Error("Expected error when removing network with containers, got nil")
	}
}

func TestNetworkManager_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "budgie-network-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create network manager and add a network
	nm1, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create first network manager: %v", err)
	}

	err = nm1.CreateNetwork("persist-net", "bridge", "172.34.0.0/16", "172.34.0.1")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	// Create a new network manager (simulating restart)
	nm2, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second network manager: %v", err)
	}

	// Verify the network persisted
	net, err := nm2.GetNetwork("persist-net")
	if err != nil {
		t.Fatalf("Failed to get persisted network: %v", err)
	}

	if net.Subnet != "172.34.0.0/16" {
		t.Errorf("Persisted subnet mismatch: got %s, want 172.34.0.0/16", net.Subnet)
	}
}
