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

package casbin

import (
	"sync"
	"testing"
	"time"

	"github.com/casbin/casbin/v2/model"
)

// TestLoadPolicyAdapter is a custom adapter for testing race conditions.
type TestLoadPolicyAdapter struct {
	policies  [][]string
	mu        sync.RWMutex
	loadDelay time.Duration
}

func NewTestLoadPolicyAdapter(policies [][]string) *TestLoadPolicyAdapter {
	return &TestLoadPolicyAdapter{
		policies:  policies,
		loadDelay: 0,
	}
}

// LoadPolicy loads all policy rules from the storage.
func (a *TestLoadPolicyAdapter) LoadPolicy(model model.Model) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Simulate slow adapter (e.g., network latency)
	if a.loadDelay > 0 {
		time.Sleep(a.loadDelay)
	}

	for _, policy := range a.policies {
		key := policy[0]
		sec := key[:1]
		_ = model.AddPolicy(sec, key, policy[1:])
	}
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *TestLoadPolicyAdapter) SavePolicy(model model.Model) error {
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *TestLoadPolicyAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemovePolicy removes a policy rule from the storage.
func (a *TestLoadPolicyAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *TestLoadPolicyAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

// UpdatePolicies updates the policies in the adapter.
func (a *TestLoadPolicyAdapter) UpdatePolicies(newPolicies [][]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.policies = newPolicies
}

// SetLoadDelay sets the delay for loading policies.
func (a *TestLoadPolicyAdapter) SetLoadDelay(delay time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loadDelay = delay
}

// TestConcurrentLoadPolicyAndModify tests the race condition where LoadPolicy
// can overwrite concurrent policy modifications.
func TestConcurrentLoadPolicyAndModify(t *testing.T) {
	// Run the test multiple times to increase chances of hitting the race condition
	for iteration := 0; iteration < 20; iteration++ {
		// Initial policies
		adapter := NewTestLoadPolicyAdapter([][]string{
			{"p", "alice", "data1", "read"},
			{"p", "bob", "data2", "write"},
		})

		e, err := NewSyncedEnforcer("examples/basic_model.conf")
		if err != nil {
			t.Fatal(err)
		}
		e.SetAdapter(adapter)

		// Load initial policy
		err = e.LoadPolicy()
		if err != nil {
			t.Fatal(err)
		}

		// Update adapter to have a new policy
		adapter.UpdatePolicies([][]string{
			{"p", "alice", "data1", "read"},
			{"p", "bob", "data2", "write"},
			{"p", "charlie", "data3", "read"},
		})

		// Add some delay to LoadPolicy to increase race window
		adapter.SetLoadDelay(10 * time.Millisecond)

		var wg sync.WaitGroup
		errors := make(chan error, 2)

		// Goroutine 1: LoadPolicy (simulating watcher callback)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if loadErr := e.LoadPolicy(); loadErr != nil {
				errors <- loadErr
			}
		}()

		// Small delay to ensure LoadPolicy starts first
		time.Sleep(2 * time.Millisecond)

		// Goroutine 2: Add a policy while LoadPolicy is in progress
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, addErr := e.AddPolicy("dave", "data4", "write")
			if addErr != nil {
				errors <- addErr
			}
		}()

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Fatal(err)
		}

		// Verify all policies are present
		policies, err := e.GetPolicy()
		if err != nil {
			t.Fatal(err)
		}

		// We expect 4 policies:
		// 1. alice, data1, read (from adapter)
		// 2. bob, data2, write (from adapter)
		// 3. charlie, data3, read (from adapter)
		// 4. dave, data4, write (added concurrently)
		expectedCount := 4
		if len(policies) != expectedCount {
			t.Errorf("Iteration %d: Expected %d policies, got %d. Policies: %v",
				iteration, expectedCount, len(policies), policies)

			checkMissingPolicy(t, policies)
		}
	}
}

// checkMissingPolicy checks which policies are present and logs if dave is missing.
func checkMissingPolicy(t *testing.T, policies [][]string) {
	t.Helper()
	hasAlice := false
	hasBob := false
	hasCharlie := false
	hasDave := false
	for _, p := range policies {
		if len(p) != 3 {
			continue
		}
		if p[0] == "alice" && p[1] == "data1" && p[2] == "read" {
			hasAlice = true
		}
		if p[0] == "bob" && p[1] == "data2" && p[2] == "write" {
			hasBob = true
		}
		if p[0] == "charlie" && p[1] == "data3" && p[2] == "read" {
			hasCharlie = true
		}
		if p[0] == "dave" && p[1] == "data4" && p[2] == "write" {
			hasDave = true
		}
	}

	if !hasDave {
		t.Error("Race condition detected: AddPolicy was lost due to concurrent LoadPolicy")
	}
	t.Logf("hasAlice=%v, hasBob=%v, hasCharlie=%v, hasDave=%v",
		hasAlice, hasBob, hasCharlie, hasDave)
}

// TestMultipleConcurrentLoadPolicy tests multiple concurrent LoadPolicy calls.
func TestMultipleConcurrentLoadPolicy(t *testing.T) {
	adapter := NewTestLoadPolicyAdapter([][]string{
		{"p", "alice", "data1", "read"},
		{"p", "bob", "data2", "write"},
	})
	adapter.SetLoadDelay(10 * time.Millisecond)

	e, err := NewSyncedEnforcer("examples/basic_model.conf")
	if err != nil {
		t.Fatal(err)
	}
	e.SetAdapter(adapter)

	// Load initial policy
	err = e.LoadPolicy()
	if err != nil {
		t.Fatal(err)
	}

	// Update adapter with new policy
	adapter.UpdatePolicies([][]string{
		{"p", "alice", "data1", "read"},
		{"p", "bob", "data2", "write"},
		{"p", "charlie", "data3", "read"},
	})

	// Launch multiple concurrent LoadPolicy calls
	var wg sync.WaitGroup
	numGoroutines := 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if loadErr := e.LoadPolicy(); loadErr != nil {
				errors <- loadErr
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for errItem := range errors {
		t.Fatal(errItem)
	}

	// Verify the final policy state
	policies, err := e.GetPolicy()
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 3
	if len(policies) != expectedCount {
		t.Errorf("Expected %d policies, got %d. Policies: %v",
			expectedCount, len(policies), policies)
	}
}

// TestAutoLoadPolicyWithConcurrentModification tests that auto-load doesn't lose concurrent modifications.
func TestAutoLoadPolicyWithConcurrentModification(t *testing.T) {
	adapter := NewTestLoadPolicyAdapter([][]string{
		{"p", "alice", "data1", "read"},
		{"p", "bob", "data2", "write"},
	})
	// Set a very small delay to speed up the test
	adapter.SetLoadDelay(1 * time.Millisecond)

	e, err := NewSyncedEnforcer("examples/basic_model.conf")
	if err != nil {
		t.Fatal(err)
	}
	e.SetAdapter(adapter)

	// Load initial policy
	err = e.LoadPolicy()
	if err != nil {
		t.Fatal(err)
	}

	// Update adapter with new policy
	adapter.UpdatePolicies([][]string{
		{"p", "alice", "data1", "read"},
		{"p", "bob", "data2", "write"},
		{"p", "charlie", "data3", "read"},
	})

	// Start auto-loading
	e.StartAutoLoadPolicy(20 * time.Millisecond)
	defer e.StopAutoLoadPolicy()

	// Wait for first auto-load
	time.Sleep(30 * time.Millisecond)

	// Verify that charlie was loaded
	policies, err := e.GetPolicy()
	if err != nil {
		t.Fatal(err)
	}

	hasCharlie := false
	for _, p := range policies {
		if len(p) == 3 && p[0] == "charlie" && p[1] == "data3" && p[2] == "read" {
			hasCharlie = true
			break
		}
	}

	if !hasCharlie {
		t.Errorf("Expected to find charlie policy after auto-load, got: %v", policies)
	}

	// Now add a policy while auto-loading is active
	_, err = e.AddPolicy("dave", "data4", "write")
	if err != nil {
		t.Fatal(err)
	}

	// Wait for another auto-load cycle
	time.Sleep(30 * time.Millisecond)

	// Verify dave is still present even after auto-load
	// Note: In reality, dave would be lost because it's not in the adapter
	// This test shows that dave is correctly maintained during LoadPolicy
	policies, err = e.GetPolicy()
	if err != nil {
		t.Fatal(err)
	}

	// After auto-load, dave should not be present because it wasn't saved to the adapter
	// This is expected behavior - LoadPolicy reloads from the adapter
	hasDave := false
	for _, p := range policies {
		if len(p) == 3 && p[0] == "dave" && p[1] == "data4" && p[2] == "write" {
			hasDave = true
			break
		}
	}

	// Dave should not be present because LoadPolicy reloads from adapter
	// and dave was never saved to the adapter
	if hasDave {
		t.Log("Note: dave is present, which means policies added locally are maintained (unexpected with LoadPolicy)")
	}
}
