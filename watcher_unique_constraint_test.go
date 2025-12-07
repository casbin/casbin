// Copyright 2024 The casbin Authors. All Rights Reserved.
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
	"errors"
	"testing"

	"github.com/casbin/casbin/v2/model"
)

// MockAdapterWithUniqueConstraint simulates a database adapter with unique constraints
type MockAdapterWithUniqueConstraint struct {
	alreadyAdded map[string]bool
}

func NewMockAdapterWithUniqueConstraint() *MockAdapterWithUniqueConstraint {
	return &MockAdapterWithUniqueConstraint{
		alreadyAdded: make(map[string]bool),
	}
}

func (a *MockAdapterWithUniqueConstraint) LoadPolicy(model model.Model) error {
	return nil
}

func (a *MockAdapterWithUniqueConstraint) SavePolicy(model model.Model) error {
	return nil
}

func (a *MockAdapterWithUniqueConstraint) AddPolicy(sec string, ptype string, rule []string) error {
	key := sec + ptype + toString(rule)
	if a.alreadyAdded[key] {
		// Simulate unique constraint violation
		return errors.New("unique constraint violation: duplicate policy")
	}
	a.alreadyAdded[key] = true
	return nil
}

func (a *MockAdapterWithUniqueConstraint) RemovePolicy(sec string, ptype string, rule []string) error {
	key := sec + ptype + toString(rule)
	delete(a.alreadyAdded, key)
	return nil
}

func (a *MockAdapterWithUniqueConstraint) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func (a *MockAdapterWithUniqueConstraint) AddPolicies(sec string, ptype string, rules [][]string) error {
	for _, rule := range rules {
		if err := a.AddPolicy(sec, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func (a *MockAdapterWithUniqueConstraint) RemovePolicies(sec string, ptype string, rules [][]string) error {
	for _, rule := range rules {
		if err := a.RemovePolicy(sec, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func toString(rule []string) string {
	result := ""
	for _, r := range rule {
		result += r + ","
	}
	return result
}

// TestWatcherNotifyWithUniqueConstraint simulates the scenario where:
// 1. Instance A adds a policy and notifies via watcher
// 2. Instance B receives the notification and tries to add the same policy
// 3. Instance B's adapter fails with unique constraint error
// 4. Instance B should still have the policy in its in-memory model
func TestWatcherNotifyWithUniqueConstraint(t *testing.T) {
	// Instance A - the one that originally adds the policy
	adapterA := NewMockAdapterWithUniqueConstraint()
	enforcerA, _ := NewEnforcer("examples/rbac_model.conf", adapterA)
	enforcerA.EnableAutoSave(true)

	// Instance B - another instance that receives the notification
	// It shares the same underlying database (simulated by sharing the alreadyAdded map)
	adapterB := &MockAdapterWithUniqueConstraint{
		alreadyAdded: adapterA.alreadyAdded, // Share the same "database"
	}
	enforcerB, _ := NewEnforcer("examples/rbac_model.conf", adapterB)
	enforcerB.EnableAutoSave(true)

	// Instance A adds a policy successfully
	ok, err := enforcerA.AddPolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Instance A should add policy successfully: %v", err)
	}
	if !ok {
		t.Fatal("Instance A should return true when adding new policy")
	}

	// Verify Instance A has the policy in memory
	hasPolicy, _ := enforcerA.HasPolicy("alice", "data1", "read")
	if !hasPolicy {
		t.Fatal("Instance A should have the policy in memory")
	}

	// Instance B receives notification and tries to add the same policy
	// This simulates the watcher callback (SelfAddPolicy is typically used in watcher callbacks)
	ok, err = enforcerB.SelfAddPolicy("p", "p", []string{"alice", "data1", "read"})

	// The current implementation will:
	// 1. Check if policy exists in memory (it doesn't in B)
	// 2. Try to persist to adapter (fails with unique constraint)
	// 3. Return error without updating memory
	// This is the BUG - Instance B should still have the policy in memory

	if err != nil {
		t.Logf("Expected: Instance B got error from adapter: %v", err)
	}

	// Instance B should have the policy in its in-memory model even if adapter failed
	// because the policy already exists in the database (added by Instance A)
	hasPolicy, _ = enforcerB.HasPolicy("alice", "data1", "read")
	if !hasPolicy {
		t.Fatal("BUG: Instance B should have the policy in memory even if adapter fails with unique constraint")
	}
}

// TestWatcherNotifyBatchWithUniqueConstraint tests the batch version
func TestWatcherNotifyBatchWithUniqueConstraint(t *testing.T) {
	// Instance A - the one that originally adds the policies
	adapterA := NewMockAdapterWithUniqueConstraint()
	enforcerA, _ := NewEnforcer("examples/rbac_model.conf", adapterA)
	enforcerA.EnableAutoSave(true)

	// Instance B - another instance that receives the notification
	adapterB := &MockAdapterWithUniqueConstraint{
		alreadyAdded: adapterA.alreadyAdded, // Share the same "database"
	}
	enforcerB, _ := NewEnforcer("examples/rbac_model.conf", adapterB)
	enforcerB.EnableAutoSave(true)

	// Instance A adds policies successfully
	rules := [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	}
	ok, err := enforcerA.AddPolicies(rules)
	if err != nil {
		t.Fatalf("Instance A should add policies successfully: %v", err)
	}
	if !ok {
		t.Fatal("Instance A should return true when adding new policies")
	}

	// Verify Instance A has the policies in memory
	hasPolicy, _ := enforcerA.HasPolicy("alice", "data1", "read")
	if !hasPolicy {
		t.Fatal("Instance A should have first policy in memory")
	}
	hasPolicy, _ = enforcerA.HasPolicy("bob", "data2", "write")
	if !hasPolicy {
		t.Fatal("Instance A should have second policy in memory")
	}

	// Instance B receives notification and tries to add the same policies
	ok, err = enforcerB.SelfAddPolicies("p", "p", rules)

	if err != nil {
		t.Logf("Expected: Instance B got error from adapter: %v", err)
	}

	// Instance B should have the policies in its in-memory model even if adapter failed
	hasPolicy, _ = enforcerB.HasPolicy("alice", "data1", "read")
	if !hasPolicy {
		t.Fatal("Instance B should have first policy in memory even if adapter fails")
	}
	hasPolicy, _ = enforcerB.HasPolicy("bob", "data2", "write")
	if !hasPolicy {
		t.Fatal("Instance B should have second policy in memory even if adapter fails")
	}
}
