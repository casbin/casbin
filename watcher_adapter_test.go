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
	"github.com/casbin/casbin/v2/persist"
)

// FailingAdapter is an adapter that fails AddPolicy to simulate database unique constraint errors.
type FailingAdapter struct {
	persist.Adapter
	failOnAdd bool
}

func (a *FailingAdapter) LoadPolicy(model model.Model) error {
	return nil
}

func (a *FailingAdapter) SavePolicy(model model.Model) error {
	return nil
}

func (a *FailingAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	if a.failOnAdd {
		return errors.New("unique constraint violation")
	}
	return nil
}

func (a *FailingAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *FailingAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

// TestWatcherWithFailingAdapter tests that when a watcher callback is triggered
// and the adapter fails to persist (e.g., due to unique constraints), the in-memory
// model is still updated to keep the instance in sync.
func TestWatcherWithFailingAdapter(t *testing.T) {
	// Create enforcer with failing adapter
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set adapter that will fail on AddPolicy
	failingAdapter := &FailingAdapter{failOnAdd: true}
	e.SetAdapter(failingAdapter)
	e.EnableAutoSave(true)

	// Check initial state - the policy should not exist
	hasPolicy, err := e.HasPolicy("eve", "data3", "write")
	if err != nil {
		t.Fatal(err)
	}
	if hasPolicy {
		t.Fatal("Policy should not exist initially")
	}

	// Simulate watcher callback by calling SelfAddPolicy
	// This simulates the scenario where another instance added the policy
	// and this instance receives the notification
	ok, err := e.SelfAddPolicy("p", "p", []string{"eve", "data3", "write"})
	if err != nil {
		t.Fatalf("SelfAddPolicy should not fail even when adapter fails: %v", err)
	}
	if !ok {
		t.Fatal("SelfAddPolicy should return true when policy is added to memory")
	}

	// Verify the policy was added to in-memory model despite adapter failure
	hasPolicy, err = e.HasPolicy("eve", "data3", "write")
	if err != nil {
		t.Fatal(err)
	}
	if !hasPolicy {
		t.Fatal("Policy should exist in memory after SelfAddPolicy, even though adapter failed")
	}

	// Verify enforcement works
	allowed, err := e.Enforce("eve", "data3", "write")
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("Enforcement should work with in-memory policy")
	}
}

// TestWatcherWithFailingAdapterGrouping tests the same scenario but for grouping policies.
func TestWatcherWithFailingAdapterGrouping(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	failingAdapter := &FailingAdapter{failOnAdd: true}
	e.SetAdapter(failingAdapter)
	e.EnableAutoSave(true)

	// Check initial state
	hasRole, err := e.HasGroupingPolicy("eve", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if hasRole {
		t.Fatal("Grouping policy should not exist initially")
	}

	// Simulate watcher callback for grouping policy
	ok, err := e.SelfAddPolicy("g", "g", []string{"eve", "admin"})
	if err != nil {
		t.Fatalf("SelfAddPolicy should not fail even when adapter fails: %v", err)
	}
	if !ok {
		t.Fatal("SelfAddPolicy should return true when grouping policy is added to memory")
	}

	// Verify the grouping policy was added to in-memory model
	hasRole, err = e.HasGroupingPolicy("eve", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if !hasRole {
		t.Fatal("Grouping policy should exist in memory after SelfAddPolicy")
	}

	// Verify role links were built correctly
	roles, err := e.GetRolesForUser("eve")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, role := range roles {
		if role == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("eve should have admin role")
	}
}
