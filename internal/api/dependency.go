package api

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zarigata/budgie/pkg/types"
)

// DependencyResolver handles container dependencies for startup ordering
type DependencyResolver struct {
	manager *ContainerManager
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(manager *ContainerManager) *DependencyResolver {
	return &DependencyResolver{manager: manager}
}

// DependencyGraph represents container dependencies
type DependencyGraph struct {
	nodes      map[string][]string // containerID -> dependencies
	containers map[string]*types.Container
}

// NewDependencyGraph creates a dependency graph from containers
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes:      make(map[string][]string),
		containers: make(map[string]*types.Container),
	}
}

// AddContainer adds a container to the dependency graph
func (g *DependencyGraph) AddContainer(ctr *types.Container, dependsOn []string) {
	g.containers[ctr.Name] = ctr
	g.nodes[ctr.Name] = dependsOn
}

// GetStartOrder returns containers in dependency-resolved order (topological sort)
func (g *DependencyGraph) GetStartOrder() ([]*types.Container, error) {
	// Track visited and in-progress nodes for cycle detection
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)
	var order []*types.Container

	// Visit function for DFS
	var visit func(name string) error
	visit = func(name string) error {
		if inProgress[name] {
			return fmt.Errorf("circular dependency detected involving %s", name)
		}
		if visited[name] {
			return nil
		}

		inProgress[name] = true

		// Visit dependencies first
		for _, dep := range g.nodes[name] {
			if _, exists := g.containers[dep]; !exists {
				return fmt.Errorf("container %s depends on unknown container %s", name, dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}

		inProgress[name] = false
		visited[name] = true

		if ctr, exists := g.containers[name]; exists {
			order = append(order, ctr)
		}

		return nil
	}

	// Visit all nodes
	for name := range g.containers {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// GetDependencies returns the dependencies for a container
func (g *DependencyGraph) GetDependencies(name string) []string {
	return g.nodes[name]
}

// WaitForDependencies waits for all dependencies of a container to be running
func (dr *DependencyResolver) WaitForDependencies(ctx context.Context, ctr *types.Container, dependencies []string, timeout time.Duration) error {
	if len(dependencies) == 0 {
		return nil
	}

	logrus.Infof("Waiting for dependencies of %s: %v", ctr.Name, dependencies)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for dependencies of %s", ctr.Name)
			}

			allReady := true
			for _, depName := range dependencies {
				depCtr := dr.findByName(depName)
				if depCtr == nil {
					return fmt.Errorf("dependency %s not found", depName)
				}

				if !depCtr.IsRunning() {
					allReady = false
					logrus.Debugf("Dependency %s is not running (state: %s)", depName, depCtr.State)
					break
				}

				// Optionally check health if health check is configured
				// For now, just check running state
			}

			if allReady {
				logrus.Infof("All dependencies of %s are ready", ctr.Name)
				return nil
			}
		}
	}
}

func (dr *DependencyResolver) findByName(name string) *types.Container {
	for _, ctr := range dr.manager.List() {
		if ctr.Name == name {
			return ctr
		}
	}
	return nil
}

// ValidateDependencies checks if all dependencies exist and there are no cycles
func (dr *DependencyResolver) ValidateDependencies(containers []*types.Container, dependenciesMap map[string][]string) error {
	graph := NewDependencyGraph()

	for _, ctr := range containers {
		deps := dependenciesMap[ctr.Name]
		graph.AddContainer(ctr, deps)
	}

	_, err := graph.GetStartOrder()
	return err
}

// GetDependents returns containers that depend on the given container
func (dr *DependencyResolver) GetDependents(name string, dependenciesMap map[string][]string) []string {
	var dependents []string
	for ctrName, deps := range dependenciesMap {
		for _, dep := range deps {
			if dep == name {
				dependents = append(dependents, ctrName)
				break
			}
		}
	}
	return dependents
}
