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

// TestLoadPolicyAdapter is a custom adapter for testing race conditions
type TestLoadPolicyAdapter struct {
	policies [][]string
	mu       sync.RWMutex
	loadDelay time.Duration
}

func NewTestLoadPolicyAdapter(policies [][]string) *TestLoadPolicyAdapter {
	return &TestLoadPolicyAdapter{
		policies: policies,
		loadDelay: 0,
	}
}

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

func (a *TestLoadPolicyAdapter) SavePolicy(model model.Model) error {
	return nil
}

func (a *TestLoadPolicyAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *TestLoadPolicyAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *TestLoadPolicyAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func (a *TestLoadPolicyAdapter) UpdatePolicies(newPolicies [][]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.policies = newPolicies
}

func (a *TestLoadPolicyAdapter) SetLoadDelay(delay time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loadDelay = delay
}

// TestConcurrentLoadPolicyAndModify tests the race condition where LoadPolicy
// can overwrite concurrent policy modifications
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
			if err := e.LoadPolicy(); err != nil {
				errors <- err
			}
		}()
		
		// Small delay to ensure LoadPolicy starts first
		time.Sleep(2 * time.Millisecond)
		
		// Goroutine 2: Add a policy while LoadPolicy is in progress
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.AddPolicy("dave", "data4", "write")
			if err != nil {
				errors <- err
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
			
			// Check which policy is missing
			hasAlice := false
			hasBob := false
			hasCharlie := false
			hasDave := false
			for _, p := range policies {
				if len(p) == 3 {
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
			}
			
			if !hasDave {
				t.Error("Race condition detected: AddPolicy was lost due to concurrent LoadPolicy")
			}
			t.Logf("hasAlice=%v, hasBob=%v, hasCharlie=%v, hasDave=%v", 
				hasAlice, hasBob, hasCharlie, hasDave)
		}
	}
}

// TestMultipleConcurrentLoadPolicy tests multiple concurrent LoadPolicy calls
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
			if err := e.LoadPolicy(); err != nil {
				errors <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Fatal(err)
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
