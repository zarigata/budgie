package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/budgie/budgie/internal/bundle"
	"github.com/budgie/budgie/internal/discovery"
	"github.com/budgie/budgie/internal/sync"
)

const (
	testTimeout = 30 * time.Second
)

// TestChirpDiscovery tests that two containers can discover each other
func TestChirpDiscovery(t *testing.T) {
	if os.Getenv("BUDGIE_E2E") != "1" {
		t.Skip("Skipping E2E test (set BUDGIE_E2E=1 to run)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create two discovery services to simulate two nodes
	disc1 := discovery.NewDiscoveryService()
	disc2 := discovery.NewDiscoveryService()

	// Load test bundle 1
	bundle1, err := bundle.Parse(filepath.Join("..", "fixtures", "test1.bun"))
	if err != nil {
		t.Fatalf("Failed to parse test1.bun: %v", err)
	}

	// Create container from bundle
	ctr1 := bundle1.ToContainer(filepath.Join("..", "fixtures", "test1.bun"))
	ctr1.ID = "test1container123456789012345678901234567890123456789012"

	// Announce container 1
	if err := disc1.AnnounceContainer(ctr1); err != nil {
		t.Fatalf("Failed to announce container 1: %v", err)
	}
	defer disc1.Shutdown()

	// Give mDNS time to propagate
	time.Sleep(2 * time.Second)

	// Discover containers from service 2
	containers, err := disc2.DiscoverContainers(5 * time.Second)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Verify we found at least one container
	if len(containers) == 0 {
		t.Error("Expected to discover at least one container, got none")
	}

	// Look for our test container
	found := false
	for _, c := range containers {
		if c.Name == ctr1.Name {
			found = true
			if c.ID != ctr1.ID {
				t.Errorf("Container ID mismatch: got %s, want %s", c.ID, ctr1.ID)
			}
			break
		}
	}

	if !found {
		t.Errorf("Test container %s not found in discovery results", ctr1.Name)
	}

	_ = ctx // Use context for potential future async operations
}

// TestVolumeSync tests volume synchronization between containers
func TestVolumeSync(t *testing.T) {
	if os.Getenv("BUDGIE_E2E") != "1" {
		t.Skip("Skipping E2E test (set BUDGIE_E2E=1 to run)")
	}

	// Create temp directories for source and destination
	srcDir, err := os.MkdirTemp("", "budgie-sync-src-*")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "budgie-sync-dst-*")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create test files in source
	testFiles := map[string]string{
		"file1.txt":         "Hello, World!",
		"subdir/file2.txt":  "Nested file content",
		"subdir/file3.json": `{"key": "value"}`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(srcDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Start sync server
	server, err := sync.NewServer(0) // Use any available port
	if err != nil {
		t.Fatalf("Failed to create sync server: %v", err)
	}
	defer server.Stop()

	server.RegisterVolume("test-container", dstDir)
	go server.Start()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create sync manager and sync files
	srcManager, err := sync.NewSyncManager(srcDir)
	if err != nil {
		t.Fatalf("Failed to create source sync manager: %v", err)
	}

	// Get server address
	addr := server.Addr().String()
	if err := srcManager.SyncVolume(srcDir, addr); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify files were synced
	for path, expectedContent := range testFiles {
		dstPath := filepath.Join(dstDir, path)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read synced file %s: %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s: got %q, want %q", path, string(content), expectedContent)
		}
	}
}

// TestSyncServer tests the sync server accepts connections
func TestSyncServer(t *testing.T) {
	if os.Getenv("BUDGIE_E2E") != "1" {
		t.Skip("Skipping E2E test (set BUDGIE_E2E=1 to run)")
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "budgie-sync-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Start sync server on any available port
	server, err := sync.NewServer(0)
	if err != nil {
		t.Fatalf("Failed to create sync server: %v", err)
	}
	defer server.Stop()

	server.RegisterVolume("test", tmpDir)
	go server.Start()

	// Verify server is listening
	addr := server.Addr()
	if addr == nil {
		t.Fatal("Server has no address")
	}

	t.Logf("Sync server listening on %s", addr.String())
}

// TestBundleParsing tests that bundle files can be parsed correctly
func TestBundleParsing(t *testing.T) {
	testCases := []struct {
		name     string
		file     string
		wantErr  bool
		wantName string
	}{
		{
			name:     "test1 bundle",
			file:     filepath.Join("..", "fixtures", "test1.bun"),
			wantErr:  false,
			wantName: "test-app-1",
		},
		{
			name:     "test2 bundle",
			file:     filepath.Join("..", "fixtures", "test2.bun"),
			wantErr:  false,
			wantName: "test-app-2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := bundle.Parse(tc.file)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if b.Name != tc.wantName {
				t.Errorf("Name mismatch: got %s, want %s", b.Name, tc.wantName)
			}

			// Verify required fields
			if b.Version == "" {
				t.Error("Version should not be empty")
			}
			if len(b.Ports) == 0 {
				t.Error("At least one port mapping required")
			}
		})
	}
}

// TestDiscoveryService tests the discovery service
func TestDiscoveryService(t *testing.T) {
	disc := discovery.NewDiscoveryService()
	if disc == nil {
		t.Fatal("NewDiscoveryService returned nil")
	}

	// Shutdown should not panic on empty service
	if err := disc.Shutdown(); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}
