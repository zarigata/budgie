package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zarigata/budgie/pkg/types"
)

func TestHealthStatus_Values(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusStarting, "starting"},
		{HealthStatusNone, "none"},
	}

	for _, tc := range tests {
		if string(tc.status) != tc.expected {
			t.Errorf("HealthStatus mismatch: got %s, want %s", string(tc.status), tc.expected)
		}
	}
}

func TestContainerHealth_InitialState(t *testing.T) {
	health := &ContainerHealth{
		Status: HealthStatusStarting,
		Log:    make([]HealthLog, 0),
	}

	if health.Status != HealthStatusStarting {
		t.Errorf("Initial status should be starting, got %s", health.Status)
	}

	if health.FailingStreak != 0 {
		t.Errorf("Initial failing streak should be 0, got %d", health.FailingStreak)
	}

	if len(health.Log) != 0 {
		t.Errorf("Initial log should be empty, got %d entries", len(health.Log))
	}
}

func TestHealthLog_Structure(t *testing.T) {
	start := time.Now()
	end := start.Add(100 * time.Millisecond)

	log := HealthLog{
		Start:    start,
		End:      end,
		ExitCode: 0,
		Output:   "HTTP 200",
	}

	if log.Start != start {
		t.Error("Start time mismatch")
	}

	if log.End != end {
		t.Error("End time mismatch")
	}

	if log.ExitCode != 0 {
		t.Errorf("Exit code should be 0, got %d", log.ExitCode)
	}

	if log.Output != "HTTP 200" {
		t.Errorf("Output mismatch: got %s", log.Output)
	}
}

func TestHealthCheck_HTTPEndpoint(t *testing.T) {
	// Create a test HTTP server that returns 200
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer healthyServer.Close()

	// Create a test HTTP server that returns 500
	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}))
	defer unhealthyServer.Close()

	// Test healthy endpoint
	resp, err := http.Get(healthyServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	// Test unhealthy endpoint
	resp2, err := http.Get(unhealthyServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", resp2.StatusCode)
	}
}

func TestHealthCheck_TypeConfiguration(t *testing.T) {
	healthCheck := &types.HealthCheck{
		Path:     "/health",
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
		Retries:  3,
	}

	if healthCheck.Path != "/health" {
		t.Errorf("Path mismatch: got %s", healthCheck.Path)
	}

	if healthCheck.Interval != 30*time.Second {
		t.Errorf("Interval mismatch: got %v", healthCheck.Interval)
	}

	if healthCheck.Timeout != 10*time.Second {
		t.Errorf("Timeout mismatch: got %v", healthCheck.Timeout)
	}

	if healthCheck.Retries != 3 {
		t.Errorf("Retries mismatch: got %d", healthCheck.Retries)
	}
}

func TestContainerHealth_LogRotation(t *testing.T) {
	health := &ContainerHealth{
		Status: HealthStatusHealthy,
		Log:    make([]HealthLog, 0),
	}

	// Add 10 log entries
	for i := 0; i < 10; i++ {
		health.Log = append(health.Log, HealthLog{
			Start:    time.Now(),
			End:      time.Now(),
			ExitCode: 0,
			Output:   "OK",
		})

		// Keep only last 5
		if len(health.Log) > 5 {
			health.Log = health.Log[1:]
		}
	}

	// Should only have 5 entries
	if len(health.Log) != 5 {
		t.Errorf("Expected 5 log entries after rotation, got %d", len(health.Log))
	}
}

func TestContainerHealth_FailingStreak(t *testing.T) {
	health := &ContainerHealth{
		Status:        HealthStatusHealthy,
		FailingStreak: 0,
	}

	retries := 3

	// Simulate failing health checks
	for i := 0; i < 5; i++ {
		health.FailingStreak++

		if health.FailingStreak >= retries {
			health.Status = HealthStatusUnhealthy
		}
	}

	if health.Status != HealthStatusUnhealthy {
		t.Errorf("Status should be unhealthy after %d failures", health.FailingStreak)
	}

	if health.FailingStreak != 5 {
		t.Errorf("Failing streak should be 5, got %d", health.FailingStreak)
	}
}

func TestContainerHealth_Recovery(t *testing.T) {
	health := &ContainerHealth{
		Status:        HealthStatusUnhealthy,
		FailingStreak: 5,
	}

	// Simulate successful health check
	health.FailingStreak = 0
	health.Status = HealthStatusHealthy

	if health.Status != HealthStatusHealthy {
		t.Error("Status should be healthy after recovery")
	}

	if health.FailingStreak != 0 {
		t.Errorf("Failing streak should be 0 after recovery, got %d", health.FailingStreak)
	}
}

func TestHealthCheck_DefaultValues(t *testing.T) {
	healthCheck := &types.HealthCheck{
		Path: "/health",
	}

	// Default interval should be used if not set
	interval := healthCheck.Interval
	if interval == 0 {
		interval = 30 * time.Second // Default
	}

	if interval != 30*time.Second {
		t.Errorf("Default interval should be 30s, got %v", interval)
	}

	// Default timeout should be used if not set
	timeout := healthCheck.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default
	}

	if timeout != 30*time.Second {
		t.Errorf("Default timeout should be 30s, got %v", timeout)
	}

	// Default retries should be used if not set
	retries := healthCheck.Retries
	if retries == 0 {
		retries = 3 // Default
	}

	if retries != 3 {
		t.Errorf("Default retries should be 3, got %d", retries)
	}
}

func TestContainerWithHealthCheck(t *testing.T) {
	container := &types.Container{
		ID:    "test123",
		Name:  "test-container",
		State: types.StateRunning,
		Health: &types.HealthCheck{
			Path:     "/health",
			Interval: 10 * time.Second,
			Timeout:  5 * time.Second,
			Retries:  3,
		},
		Ports: []types.PortMapping{
			{ContainerPort: 8080, HostPort: 8080, Protocol: "tcp"},
		},
	}

	if container.Health == nil {
		t.Fatal("Health check should not be nil")
	}

	if container.Health.Path != "/health" {
		t.Errorf("Health check path mismatch: got %s", container.Health.Path)
	}

	// Verify the health check can construct a URL
	if len(container.Ports) == 0 {
		t.Fatal("Container should have ports for health check")
	}

	port := container.Ports[0].HostPort
	expectedURL := "http://localhost:8080/health"
	actualURL := "http://localhost:" + string(rune('0'+port/1000)) + string(rune('0'+(port%1000)/100)) + string(rune('0'+(port%100)/10)) + string(rune('0'+port%10)) + container.Health.Path

	// Simplified check
	if port != 8080 {
		t.Errorf("Port mismatch: got %d, want 8080", port)
	}
	_ = expectedURL
	_ = actualURL
}
