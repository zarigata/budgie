// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/zarigata/budgie/internal/api"
	"github.com/zarigata/budgie/internal/config"
	"github.com/zarigata/budgie/internal/runtime"
	"github.com/zarigata/budgie/pkg/types"
)

// CommandContext holds common resources needed by commands
type CommandContext struct {
	Runtime runtime.Runtime
	Manager *api.ContainerManager
	Config  *config.Config
	DataDir string
}

// NewCommandContext initializes the common command context with runtime and manager.
// This reduces boilerplate in individual commands.
func NewCommandContext() (*CommandContext, error) {
	cfg := config.Get()

	rt, err := runtime.GetDefaultRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize runtime: %w", err)
	}

	dataDir := cfg.DataDir
	if envDir := os.Getenv("BUDGIE_DATA_DIR"); envDir != "" {
		dataDir = envDir
	}

	manager, err := api.NewContainerManager(rt, dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container manager: %w", err)
	}

	return &CommandContext{
		Runtime: rt,
		Manager: manager,
		Config:  cfg,
		DataDir: dataDir,
	}, nil
}

// FindContainer finds a container by ID prefix or name.
// It returns an error if no container is found or if the ID is ambiguous.
func FindContainer(manager *api.ContainerManager, idOrName string) (*types.Container, error) {
	// Try exact match first
	if ctr, err := manager.Get(idOrName); err == nil {
		return ctr, nil
	}

	// Try prefix match and name match
	containers := manager.List()
	var matches []*types.Container

	for _, ctr := range containers {
		if strings.HasPrefix(ctr.ID, idOrName) || ctr.Name == idOrName {
			matches = append(matches, ctr)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no such container: %s", idOrName)
	}

	if len(matches) > 1 {
		var ids []string
		for _, m := range matches {
			ids = append(ids, m.ShortID())
		}
		return nil, fmt.Errorf("ambiguous container ID %q, matches: %s", idOrName, strings.Join(ids, ", "))
	}

	return matches[0], nil
}

// FindContainers finds multiple containers by ID prefixes or names.
// Returns found containers and a slice of errors for those not found.
func FindContainers(manager *api.ContainerManager, idsOrNames []string) ([]*types.Container, []error) {
	var found []*types.Container
	var errors []error

	for _, idOrName := range idsOrNames {
		ctr, err := FindContainer(manager, idOrName)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", idOrName, err))
		} else {
			found = append(found, ctr)
		}
	}

	return found, errors
}

// MustFindContainer finds a container or returns a user-friendly error.
// Use this when a container is required for the command to proceed.
func (ctx *CommandContext) MustFindContainer(idOrName string) (*types.Container, error) {
	return FindContainer(ctx.Manager, idOrName)
}

// GetDataDir returns the configured data directory, checking environment override.
func GetDataDir() string {
	if envDir := os.Getenv("BUDGIE_DATA_DIR"); envDir != "" {
		return envDir
	}
	cfg := config.Get()
	return cfg.DataDir
}

// FormatContainerID formats a container ID for display (short form).
func FormatContainerID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// RequireRunning returns an error if the container is not running.
func RequireRunning(ctr *types.Container) error {
	if !ctr.IsRunning() {
		return fmt.Errorf("container %s is not running (state: %s)", ctr.ShortID(), ctr.State)
	}
	return nil
}

// RequireStopped returns an error if the container is running.
func RequireStopped(ctr *types.Container) error {
	if ctr.IsRunning() {
		return fmt.Errorf("container %s is running, stop it first or use --force", ctr.ShortID())
	}
	return nil
}
