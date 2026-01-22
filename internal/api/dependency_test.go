package api

import (
	"testing"

	"github.com/zarigata/budgie/pkg/types"
)

func TestDependencyGraph_SimpleOrder(t *testing.T) {
	graph := NewDependencyGraph()

	// Create containers
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	ctrB := &types.Container{ID: "b", Name: "app-b"}
	ctrC := &types.Container{ID: "c", Name: "app-c"}

	// B depends on A, C depends on B
	graph.AddContainer(ctrA, []string{})
	graph.AddContainer(ctrB, []string{"app-a"})
	graph.AddContainer(ctrC, []string{"app-b"})

	order, err := graph.GetStartOrder()
	if err != nil {
		t.Fatalf("Failed to get start order: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("Expected 3 containers in order, got %d", len(order))
	}

	// A should come before B, B should come before C
	posA, posB, posC := -1, -1, -1
	for i, ctr := range order {
		switch ctr.Name {
		case "app-a":
			posA = i
		case "app-b":
			posB = i
		case "app-c":
			posC = i
		}
	}

	if posA > posB {
		t.Error("app-a should come before app-b")
	}
	if posB > posC {
		t.Error("app-b should come before app-c")
	}
}

func TestDependencyGraph_MultipleDependencies(t *testing.T) {
	graph := NewDependencyGraph()

	// Create containers
	ctrDB := &types.Container{ID: "db", Name: "database"}
	ctrCache := &types.Container{ID: "cache", Name: "redis"}
	ctrApp := &types.Container{ID: "app", Name: "webapp"}

	// webapp depends on both database and redis
	graph.AddContainer(ctrDB, []string{})
	graph.AddContainer(ctrCache, []string{})
	graph.AddContainer(ctrApp, []string{"database", "redis"})

	order, err := graph.GetStartOrder()
	if err != nil {
		t.Fatalf("Failed to get start order: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("Expected 3 containers in order, got %d", len(order))
	}

	// webapp should be last
	if order[len(order)-1].Name != "webapp" {
		t.Error("webapp should be last in order")
	}
}

func TestDependencyGraph_CircularDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Create circular dependency: A -> B -> C -> A
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	ctrB := &types.Container{ID: "b", Name: "app-b"}
	ctrC := &types.Container{ID: "c", Name: "app-c"}

	graph.AddContainer(ctrA, []string{"app-c"})
	graph.AddContainer(ctrB, []string{"app-a"})
	graph.AddContainer(ctrC, []string{"app-b"})

	_, err := graph.GetStartOrder()
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}
}

func TestDependencyGraph_SelfDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Container depends on itself
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	graph.AddContainer(ctrA, []string{"app-a"})

	_, err := graph.GetStartOrder()
	if err == nil {
		t.Error("Expected error for self-dependency, got nil")
	}
}

func TestDependencyGraph_MissingDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Container depends on non-existent container
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	graph.AddContainer(ctrA, []string{"missing-container"})

	_, err := graph.GetStartOrder()
	if err == nil {
		t.Error("Expected error for missing dependency, got nil")
	}
}

func TestDependencyGraph_NoDependencies(t *testing.T) {
	graph := NewDependencyGraph()

	// Create containers with no dependencies
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	ctrB := &types.Container{ID: "b", Name: "app-b"}
	ctrC := &types.Container{ID: "c", Name: "app-c"}

	graph.AddContainer(ctrA, []string{})
	graph.AddContainer(ctrB, []string{})
	graph.AddContainer(ctrC, []string{})

	order, err := graph.GetStartOrder()
	if err != nil {
		t.Fatalf("Failed to get start order: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("Expected 3 containers in order, got %d", len(order))
	}
}

func TestDependencyGraph_DiamondDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Diamond dependency:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	ctrA := &types.Container{ID: "a", Name: "app-a"}
	ctrB := &types.Container{ID: "b", Name: "app-b"}
	ctrC := &types.Container{ID: "c", Name: "app-c"}
	ctrD := &types.Container{ID: "d", Name: "app-d"}

	graph.AddContainer(ctrA, []string{})
	graph.AddContainer(ctrB, []string{"app-a"})
	graph.AddContainer(ctrC, []string{"app-a"})
	graph.AddContainer(ctrD, []string{"app-b", "app-c"})

	order, err := graph.GetStartOrder()
	if err != nil {
		t.Fatalf("Failed to get start order: %v", err)
	}

	if len(order) != 4 {
		t.Fatalf("Expected 4 containers in order, got %d", len(order))
	}

	// Find positions
	positions := make(map[string]int)
	for i, ctr := range order {
		positions[ctr.Name] = i
	}

	// A should come first
	if positions["app-a"] != 0 {
		t.Error("app-a should be first")
	}

	// D should come last
	if positions["app-d"] != 3 {
		t.Error("app-d should be last")
	}

	// B and C should come before D
	if positions["app-b"] > positions["app-d"] {
		t.Error("app-b should come before app-d")
	}
	if positions["app-c"] > positions["app-d"] {
		t.Error("app-c should come before app-d")
	}
}

func TestDependencyGraph_GetDependencies(t *testing.T) {
	graph := NewDependencyGraph()

	ctrA := &types.Container{ID: "a", Name: "app-a"}
	ctrB := &types.Container{ID: "b", Name: "app-b"}

	graph.AddContainer(ctrA, []string{})
	graph.AddContainer(ctrB, []string{"app-a"})

	deps := graph.GetDependencies("app-b")
	if len(deps) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(deps))
	}

	if deps[0] != "app-a" {
		t.Errorf("Expected dependency 'app-a', got '%s'", deps[0])
	}
}

func TestDependencyGraph_EmptyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	order, err := graph.GetStartOrder()
	if err != nil {
		t.Fatalf("Failed to get start order for empty graph: %v", err)
	}

	if len(order) != 0 {
		t.Errorf("Expected empty order for empty graph, got %d", len(order))
	}
}

func TestDependencyResolver_GetDependents(t *testing.T) {
	// Create a mock dependency resolver
	dr := &DependencyResolver{}

	dependenciesMap := map[string][]string{
		"webapp":   {"database", "redis"},
		"worker":   {"database", "rabbitmq"},
		"database": {},
		"redis":    {},
		"rabbitmq": {},
	}

	// Get dependents of database
	dependents := dr.GetDependents("database", dependenciesMap)

	if len(dependents) != 2 {
		t.Fatalf("Expected 2 dependents, got %d", len(dependents))
	}

	found := make(map[string]bool)
	for _, d := range dependents {
		found[d] = true
	}

	if !found["webapp"] {
		t.Error("webapp should depend on database")
	}
	if !found["worker"] {
		t.Error("worker should depend on database")
	}
}
