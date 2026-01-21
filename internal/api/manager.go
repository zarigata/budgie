package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/budgie/budgie/internal/runtime"
	"github.com/budgie/budgie/pkg/types"
	"github.com/sirupsen/logrus"
)

type ContainerManager struct {
	runtime  runtime.Runtime
	containers map[string]*types.Container
	mu        sync.RWMutex
	statePath string
}

func NewContainerManager(rt runtime.Runtime, dataDir string) (*ContainerManager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	cm := &ContainerManager{
		runtime:   rt,
		containers: make(map[string]*types.Container),
		statePath:  filepath.Join(dataDir, "state.json"),
	}

	if err := cm.loadState(); err != nil {
		logrus.Warnf("Failed to load state: %v", err)
	}

	return cm, nil
}

func (m *ContainerManager) Create(ctx context.Context, ctr *types.Container) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.containers[ctr.ID]; exists {
		return fmt.Errorf("container already exists: %s", ctr.ID)
	}

	if err := m.runtime.Create(ctx, ctr); err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	ctr.State = types.StateCreated
	m.containers[ctr.ID] = ctr

	if err := m.saveState(); err != nil {
		logrus.Warnf("Failed to save state: %v", err)
	}

	return nil
}

func (m *ContainerManager) Start(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctr, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	if ctr.State != types.StateCreated && ctr.State != types.StateStopped {
		return fmt.Errorf("container is not in startable state: %s", ctr.State)
	}

	if err := m.runtime.Start(ctx, id); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	ctr.State = types.StateRunning
	ctr.StartedAt = time.Now()

	if err := m.saveState(); err != nil {
		logrus.Warnf("Failed to save state: %v", err)
	}

	return nil
}

func (m *ContainerManager) Stop(ctx context.Context, id string, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctr, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	if ctr.State != types.StateRunning {
		return fmt.Errorf("container is not running: %s", id)
	}

	if err := m.runtime.Stop(ctx, id, timeout); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	ctr.State = types.StateStopped
	ctr.ExitedAt = time.Now()

	if err := m.saveState(); err != nil {
		logrus.Warnf("Failed to save state: %v", err)
	}

	return nil
}

func (m *ContainerManager) Remove(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctr, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container not found: %s", id)
	}

	if ctr.State == types.StateRunning {
		return fmt.Errorf("cannot remove running container: %s", id)
	}

	if err := m.runtime.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	delete(m.containers, id)

	if err := m.saveState(); err != nil {
		logrus.Warnf("Failed to save state: %v", err)
	}

	return nil
}

func (m *ContainerManager) Get(id string) (*types.Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctr, exists := m.containers[id]
	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	return ctr, nil
}

func (m *ContainerManager) List() []*types.Container {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*types.Container, 0, len(m.containers))
	for _, ctr := range m.containers {
		list = append(list, ctr)
	}

	return list
}

func (m *ContainerManager) loadState() error {
	data, err := os.ReadFile(m.statePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var containers []*types.Container
	if err := json.Unmarshal(data, &containers); err != nil {
		return err
	}

	for _, ctr := range containers {
		m.containers[ctr.ID] = ctr
	}

	return nil
}

func (m *ContainerManager) saveState() error {
	list := make([]*types.Container, 0, len(m.containers))
	for _, ctr := range m.containers {
		list = append(list, ctr)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.statePath, data, 0644)
}
