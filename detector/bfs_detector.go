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
	"container/list"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/rbac"
)

// BFSDetector is a topological sort-based implementation of the Detector interface.
// It uses Kahn's algorithm (a topological sorting algorithm with BFS-like queue operations) to detect cycles in role inheritance.
type BFSDetector struct{}

// NewBFSDetector creates a new instance of BFSDetector.
func NewBFSDetector() *BFSDetector {
	return &BFSDetector{}
}

// Check checks whether the current status of the passed-in RoleManager contains logical errors (e.g., cycles in role inheritance).
// It uses Kahn's algorithm (topological sort) to detect cycles.
// Returns nil if no cycle is found, otherwise returns an error with a description of the cycle.
func (d *BFSDetector) Check(rm rbac.RoleManager) error {
	// Defensive nil check to prevent runtime panics
	if rm == nil {
		return fmt.Errorf("role manager cannot be nil")
	}

	// Build the adjacency graph by exploring all roles
	graph, err := d.buildGraph(rm)
	if err != nil {
		return err
	}

	// Run BFS-based cycle detection using Kahn's algorithm
	return d.detectCycle(graph)
}

// buildGraph builds an adjacency list representation of the role inheritance graph.
// It uses the Range method (via type assertion) to iterate through all role links.
func (d *BFSDetector) buildGraph(rm rbac.RoleManager) (graph map[string][]string, err error) {
	graph = make(map[string][]string)

	// Recover from any panics during Range iteration (e.g., nil pointer dereferences)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("RoleManager is not properly initialized: %v", r)
		}
	}()

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

// detectCycle performs cycle detection using Kahn's algorithm (topological sort).
// If a topological sort is possible, there's no cycle. If not all nodes are processed, there's a cycle.
// Returns nil if no cycle is found, otherwise returns an error describing the cycle.
func (d *BFSDetector) detectCycle(graph map[string][]string) error {
	if len(graph) == 0 {
		return nil
	}

	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for node := range graph {
		if _, exists := inDegree[node]; !exists {
			inDegree[node] = 0
		}
		for _, neighbor := range graph[node] {
			inDegree[neighbor]++
		}
	}

	// Use container/list for efficient O(1) enqueue/dequeue operations
	queue := list.New()
	for node, degree := range inDegree {
		if degree == 0 {
			queue.PushBack(node)
		}
	}

	// Process nodes with BFS
	processedCount := 0
	for queue.Len() > 0 {
		// Dequeue from front - O(1) operation
		element := queue.Front()
		current := element.Value.(string)
		queue.Remove(element)
		processedCount++

		// Reduce in-degree of neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue.PushBack(neighbor)
			}
		}
	}

	// If not all nodes were processed, there's a cycle
	if processedCount < len(graph) {
		// Find nodes that are part of the cycle
		cycleNodes := []string{}
		for node, degree := range inDegree {
			if degree > 0 {
				cycleNodes = append(cycleNodes, node)
			}
		}
		return fmt.Errorf("cycle detected involving nodes: %s", strings.Join(cycleNodes, ", "))
	}

	return nil
}
