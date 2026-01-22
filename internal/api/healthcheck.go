package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zarigata/budgie/pkg/types"
)

// HealthStatus represents the health status of a container
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusStarting  HealthStatus = "starting"
	HealthStatusNone      HealthStatus = "none"
)

// ContainerHealth tracks the health status of a container
type ContainerHealth struct {
	Status        HealthStatus  `json:"status"`
	FailingStreak int           `json:"failing_streak"`
	Log           []HealthLog   `json:"log"`
	mu            sync.Mutex
}

// HealthLog records a health check result
type HealthLog struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	ExitCode int       `json:"exit_code"`
	Output   string    `json:"output"`
}

// HealthCheckMonitor monitors container health and triggers restarts for unhealthy containers
type HealthCheckMonitor struct {
	manager        *ContainerManager
	restartMonitor *RestartMonitor
	health         map[string]*ContainerHealth
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	httpClient     *http.Client
}

// NewHealthCheckMonitor creates a new health check monitor
func NewHealthCheckMonitor(manager *ContainerManager, restartMonitor *RestartMonitor) *HealthCheckMonitor {
	return &HealthCheckMonitor{
		manager:        manager,
		restartMonitor: restartMonitor,
		health:         make(map[string]*ContainerHealth),
		stopChan:       make(chan struct{}),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start begins health check monitoring
func (hm *HealthCheckMonitor) Start() {
	hm.wg.Add(1)
	go hm.monitor()
	logrus.Info("Health check monitor started")
}

// Stop stops the health check monitor
func (hm *HealthCheckMonitor) Stop() {
	close(hm.stopChan)
	hm.wg.Wait()
	logrus.Info("Health check monitor stopped")
}

// GetHealth returns the health status for a container
func (hm *HealthCheckMonitor) GetHealth(containerID string) *ContainerHealth {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.health[containerID]
}

func (hm *HealthCheckMonitor) monitor() {
	defer hm.wg.Done()

	// Check health every second, but each container has its own interval
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.checkAllContainers()
		}
	}
}

func (hm *HealthCheckMonitor) checkAllContainers() {
	containers := hm.manager.List()

	for _, ctr := range containers {
		// Skip containers without health check config
		if ctr.Health == nil || ctr.Health.Path == "" {
			continue
		}

		// Only check running containers
		if ctr.State != types.StateRunning {
			continue
		}

		// Get or create health status
		health := hm.getOrCreateHealth(ctr.ID)

		// Check if it's time to run health check
		if hm.shouldRunHealthCheck(ctr, health) {
			go hm.runHealthCheck(ctr, health)
		}
	}
}

func (hm *HealthCheckMonitor) getOrCreateHealth(containerID string) *ContainerHealth {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if health, exists := hm.health[containerID]; exists {
		return health
	}

	health := &ContainerHealth{
		Status: HealthStatusStarting,
		Log:    make([]HealthLog, 0),
	}
	hm.health[containerID] = health
	return health
}

func (hm *HealthCheckMonitor) shouldRunHealthCheck(ctr *types.Container, health *ContainerHealth) bool {
	health.mu.Lock()
	defer health.mu.Unlock()

	// If no logs, always run
	if len(health.Log) == 0 {
		return true
	}

	// Check interval
	lastCheck := health.Log[len(health.Log)-1]
	interval := ctr.Health.Interval
	if interval == 0 {
		interval = 30 * time.Second // Default interval
	}

	return time.Since(lastCheck.End) >= interval
}

func (hm *HealthCheckMonitor) runHealthCheck(ctr *types.Container, health *ContainerHealth) {
	start := time.Now()

	// Determine health check URL
	var checkURL string
	for _, port := range ctr.Ports {
		if port.HostPort > 0 {
			checkURL = fmt.Sprintf("http://localhost:%d%s", port.HostPort, ctr.Health.Path)
			break
		}
	}

	if checkURL == "" {
		// No exposed port, skip health check
		return
	}

	timeout := ctr.Health.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		hm.recordHealthCheck(ctr, health, start, 1, err.Error())
		return
	}

	resp, err := hm.httpClient.Do(req)
	if err != nil {
		hm.recordHealthCheck(ctr, health, start, 1, err.Error())
		return
	}
	defer resp.Body.Close()

	// Consider 2xx status codes as healthy
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		hm.recordHealthCheck(ctr, health, start, 0, fmt.Sprintf("HTTP %d", resp.StatusCode))
	} else {
		hm.recordHealthCheck(ctr, health, start, 1, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
}

func (hm *HealthCheckMonitor) recordHealthCheck(ctr *types.Container, health *ContainerHealth, start time.Time, exitCode int, output string) {
	health.mu.Lock()
	defer health.mu.Unlock()

	log := HealthLog{
		Start:    start,
		End:      time.Now(),
		ExitCode: exitCode,
		Output:   output,
	}

	// Keep only last 5 health check results
	health.Log = append(health.Log, log)
	if len(health.Log) > 5 {
		health.Log = health.Log[1:]
	}

	retries := ctr.Health.Retries
	if retries == 0 {
		retries = 3 // Default retries
	}

	if exitCode == 0 {
		// Health check passed
		health.FailingStreak = 0
		health.Status = HealthStatusHealthy
		logrus.Debugf("Container %s health check passed", ctr.ShortID())
	} else {
		// Health check failed
		health.FailingStreak++
		logrus.Warnf("Container %s health check failed (%d/%d): %s",
			ctr.ShortID(), health.FailingStreak, retries, output)

		if health.FailingStreak >= retries {
			health.Status = HealthStatusUnhealthy
			hm.handleUnhealthy(ctr)
		}
	}
}

func (hm *HealthCheckMonitor) handleUnhealthy(ctr *types.Container) {
	logrus.Warnf("Container %s is unhealthy, triggering restart", ctr.ShortID())

	// Mark container as failed to trigger restart policy
	hm.manager.mu.Lock()
	ctr.State = types.StateFailed
	hm.manager.mu.Unlock()

	// Save state
	if err := hm.manager.saveState(); err != nil {
		logrus.Errorf("Failed to save state after marking container unhealthy: %v", err)
	}

	// The RestartMonitor will pick up the failed state and restart if policy allows
}

// RemoveHealth removes health tracking for a container (called when container is removed)
func (hm *HealthCheckMonitor) RemoveHealth(containerID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.health, containerID)
}

// ResetHealth resets health tracking for a container (called on container start)
func (hm *HealthCheckMonitor) ResetHealth(containerID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.health[containerID] = &ContainerHealth{
		Status: HealthStatusStarting,
		Log:    make([]HealthLog, 0),
	}
}
