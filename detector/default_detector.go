// Copyright 2025 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package detector

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/rbac"
)

// rangeableRM is an interface for role managers that support iterating over all role links.
// This is used to build the adjacency graph for cycle detection.
type rangeableRM interface {
	Range(func(name1, name2 string, domain ...string) bool)
}

// DefaultDetector is the default implementation of the Detector interface.
// It uses depth-first search (DFS) to detect cycles in role inheritance.
type DefaultDetector struct{}

// NewDefaultDetector creates a new instance of DefaultDetector.
func NewDefaultDetector() *DefaultDetector {
	return &DefaultDetector{}
}

// Check checks whether the current status of the passed-in RoleManager contains logical errors (e.g., cycles in role inheritance).
// It uses DFS to traverse the role graph and detect cycles.
// Returns nil if no cycle is found, otherwise returns an error with a description of the cycle.
func (d *DefaultDetector) Check(rm rbac.RoleManager) error {
	// Build the adjacency graph by exploring all roles
	graph, err := d.buildGraph(rm)
	if err != nil {
		return err
	}

	// Run DFS to detect cycles
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for role := range graph {
		if !visited[role] {
			if cycle := d.detectCycle(role, graph, visited, recursionStack, []string{}); cycle != nil {
				return fmt.Errorf("Cycle detected: %s", strings.Join(cycle, " -> "))
			}
		}
	}

	return nil
}

// buildGraph builds an adjacency list representation of the role inheritance graph.
// It uses the Range method (via type assertion) to iterate through all role links.
func (d *DefaultDetector) buildGraph(rm rbac.RoleManager) (map[string][]string, error) {
	graph := make(map[string][]string)

	// Try to cast to a RoleManager implementation that supports Range
	// This works with RoleManagerImpl and similar implementations
	rrm, ok := rm.(rangeableRM)
	if !ok {
		// Return an error if the RoleManager doesn't support Range iteration
		return nil, fmt.Errorf("RoleManager does not support Range iteration, cannot detect cycles")
	}

	// Use Range method to build the graph directly
	rrm.Range(func(name1, name2 string, domain ...string) bool {
		// Initialize empty slice for name1 if it doesn't exist
		if graph[name1] == nil {
			graph[name1] = []string{}
		}
		// Add the link: name1 -> name2
		graph[name1] = append(graph[name1], name2)

		// Ensure name2 exists in graph even if it has no outgoing edges
		if graph[name2] == nil {
			graph[name2] = []string{}
		}
		return true
	})
	return graph, nil
}

// detectCycle performs DFS to detect cycles in the role graph.
// Returns a slice representing the cycle path if found, nil otherwise.
func (d *DefaultDetector) detectCycle(
	role string,
	graph map[string][]string,
	visited map[string]bool,
	recursionStack map[string]bool,
	path []string,
) []string {
	// Mark the current role as visited and add to recursion stack
	visited[role] = true
	recursionStack[role] = true
	path = append(path, role)

	// Visit all neighbors (parent roles)
	for _, neighbor := range graph[role] {
		if !visited[neighbor] {
			// Recursively visit unvisited neighbor
			if cycle := d.detectCycle(neighbor, graph, visited, recursionStack, path); cycle != nil {
				return cycle
			}
		} else if recursionStack[neighbor] {
			// Back edge found - cycle detected
			// Find where the cycle starts in the path
			cycleStart := -1
			for i, p := range path {
				if p == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				// Build the cycle path
				cyclePath := append([]string{}, path[cycleStart:]...)
				cyclePath = append(cyclePath, neighbor)
				return cyclePath
			}
		}
	}

	// Remove from recursion stack before returning
	recursionStack[role] = false
	return nil
}
