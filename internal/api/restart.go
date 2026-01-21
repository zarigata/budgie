package api

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zarigata/budgie/pkg/types"
)

// RestartMonitor monitors containers and restarts them according to their restart policy
type RestartMonitor struct {
	manager    *ContainerManager
	stopChan   chan struct{}
	wg         sync.WaitGroup
	interval   time.Duration
	maxBackoff time.Duration
}

// NewRestartMonitor creates a new restart monitor
func NewRestartMonitor(manager *ContainerManager) *RestartMonitor {
	return &RestartMonitor{
		manager:    manager,
		stopChan:   make(chan struct{}),
		interval:   5 * time.Second,
		maxBackoff: 5 * time.Minute,
	}
}

// Start begins monitoring containers for restart
func (rm *RestartMonitor) Start() {
	rm.wg.Add(1)
	go rm.monitor()
	logrus.Info("Restart monitor started")
}

// Stop stops the restart monitor
func (rm *RestartMonitor) Stop() {
	close(rm.stopChan)
	rm.wg.Wait()
	logrus.Info("Restart monitor stopped")
}

func (rm *RestartMonitor) monitor() {
	defer rm.wg.Done()

	ticker := time.NewTicker(rm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.checkContainers()
		}
	}
}

func (rm *RestartMonitor) checkContainers() {
	containers := rm.manager.List()

	for _, ctr := range containers {
		if rm.shouldRestart(ctr) {
			rm.restartContainer(ctr)
		}
	}
}

func (rm *RestartMonitor) shouldRestart(ctr *types.Container) bool {
	// Only consider stopped/failed containers
	if ctr.State != types.StateStopped && ctr.State != types.StateFailed {
		return false
	}

	// Check restart policy
	if ctr.RestartPolicy == nil {
		return false
	}

	switch ctr.RestartPolicy.Name {
	case "no":
		return false

	case "always":
		return true

	case "on-failure":
		// Check if max retries exceeded
		if ctr.RestartPolicy.MaximumRetryCount > 0 {
			if ctr.RestartCount >= ctr.RestartPolicy.MaximumRetryCount {
				return false
			}
		}
		// Only restart on failure (non-zero exit)
		return ctr.State == types.StateFailed

	case "unless-stopped":
		// Restart unless explicitly stopped by user
		// We track this by checking if the container was stopped gracefully
		// For now, we assume all stopped containers should restart
		return true

	default:
		return false
	}
}

func (rm *RestartMonitor) restartContainer(ctr *types.Container) {
	ctx := context.Background()

	// Calculate backoff based on restart count
	backoff := rm.calculateBackoff(ctr.RestartCount)

	// Check if enough time has passed since last exit
	if time.Since(ctr.ExitedAt) < backoff {
		return
	}

	logrus.Infof("Restarting container %s (attempt %d)", ctr.ShortID(), ctr.RestartCount+1)

	// Update restart count
	rm.manager.mu.Lock()
	ctr.RestartCount++
	rm.manager.mu.Unlock()

	// Start the container
	if err := rm.manager.Start(ctx, ctr.ID); err != nil {
		logrus.Errorf("Failed to restart container %s: %v", ctr.ShortID(), err)
		return
	}

	logrus.Infof("Container %s restarted successfully", ctr.ShortID())
}

func (rm *RestartMonitor) calculateBackoff(restartCount int) time.Duration {
	// Exponential backoff: 1s, 2s, 4s, 8s, ... up to maxBackoff
	backoff := time.Second * time.Duration(1<<uint(restartCount))
	if backoff > rm.maxBackoff {
		backoff = rm.maxBackoff
	}
	return backoff
}

// ResetRestartCount resets the restart count for a container (called on manual start)
func (rm *RestartMonitor) ResetRestartCount(containerID string) {
	rm.manager.mu.Lock()
	defer rm.manager.mu.Unlock()

	if ctr, exists := rm.manager.containers[containerID]; exists {
		ctr.RestartCount = 0
	}
}
