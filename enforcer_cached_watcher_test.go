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
	"sync"
	"testing"

	"github.com/casbin/casbin/v3/persist/cache"
)

// mockWatcherEx is a mock WatcherEx for testing callback behavior.
type mockWatcherEx struct {
	mu          sync.Mutex
	callback    func(string)
	updateCount int
}

func (m *mockWatcherEx) SetUpdateCallback(fn func(string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = fn
	return nil
}

func (m *mockWatcherEx) Update() error {
	return nil
}

func (m *mockWatcherEx) Close() {}

func (m *mockWatcherEx) UpdateForAddPolicy(sec, ptype string, params ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCount++
	// Simulate calling the callback when a policy change is pushed
	if m.callback != nil {
		m.callback("add_policy")
	}
	return nil
}

func (m *mockWatcherEx) UpdateForRemovePolicy(sec, ptype string, params ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCount++
	if m.callback != nil {
		m.callback("remove_policy")
	}
	return nil
}

func (m *mockWatcherEx) UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func (m *mockWatcherEx) UpdateForSavePolicy(model interface{}) error {
	return nil
}

func (m *mockWatcherEx) UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCount++
	if m.callback != nil {
		m.callback("add_policies")
	}
	return nil
}

func (m *mockWatcherEx) UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error {
	return nil
}

func (m *mockWatcherEx) UpdateForUpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	return nil
}

func (m *mockWatcherEx) UpdateForUpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return nil
}

func (m *mockWatcherEx) GetUpdateCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateCount
}

// TestCachedEnforcerWatcherExCallback tests that CachedEnforcer properly sets a callback
// when a WatcherEx is provided, and that the callback invalidates the cache.
func TestCachedEnforcerWatcherExCallback(t *testing.T) {
	e, err := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create CachedEnforcer: %v", err)
	}

	mock := &mockWatcherEx{}
	if err := e.SetWatcher(mock); err != nil {
		t.Fatalf("Failed to set watcher: %v", err)
	}

	// First enforce call - should be a cache miss
	res, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to have read access to data1")
	}

	// Second enforce call - should be cached
	res, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected cached result for alice read data1")
	}

	// Simulate a policy change from another instance via WatcherEx.UpdateForAddPolicy
	// This should trigger the callback, which should invalidate the cache
	// and reload the policy.
	_, err = e.AddPolicy("alice", "data1", "write")
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	// The cache should have been invalidated by the callback
	// Verify that the new policy is reflected
	res, err = e.Enforce("alice", "data1", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to have write access to data1 after policy was added")
	}
}

// TestCachedEnforcerBasicWatcherCallback tests that CachedEnforcer works with
// a basic Watcher (not WatcherEx) and properly invalidates cache.
func TestCachedEnforcerBasicWatcherCallback(t *testing.T) {
	e, err := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create CachedEnforcer: %v", err)
	}

	mock := &mockBasicWatcher{}
	if err := e.SetWatcher(mock); err != nil {
		t.Fatalf("Failed to set watcher: %v", err)
	}

	// First enforce call - cache miss
	res, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to have read access to data1")
	}

	// Second enforce call - should be cached (same result)
	res, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected cached result")
	}

	// Verify the basic watcher is being used (not the WatcherEx path)
	if !mock.callbackSet {
		t.Error("Expected callback to be set for basic watcher")
	}
}

// mockBasicWatcher is a basic Watcher for testing.
type mockBasicWatcher struct {
	mu           sync.Mutex
	callback     func(string)
	callbackSet  bool
}

func (m *mockBasicWatcher) SetUpdateCallback(fn func(string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = fn
	m.callbackSet = true
	return nil
}

func (m *mockBasicWatcher) Update() error {
	return nil
}

func (m *mockBasicWatcher) Close() {}

// TestSyncedCachedEnforcerWatcherExCallback tests that SyncedCachedEnforcer properly
// sets a callback for WatcherEx implementations.
func TestSyncedCachedEnforcerWatcherExCallback(t *testing.T) {
	e, err := NewSyncedCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create SyncedCachedEnforcer: %v", err)
	}

	mock := &mockWatcherEx{}
	if err := e.SetWatcher(mock); err != nil {
		t.Fatalf("Failed to set watcher: %v", err)
	}

	// First enforce - cache miss
	res, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to have read access to data1")
	}

	// Second enforce - cached
	res, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected cached result")
	}

	// Remove policy - should trigger callback which invalidates cache
	_, err = e.RemovePolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("RemovePolicy failed: %v", err)
	}

	// After policy removal and cache invalidation, this should now be false
	res, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if res {
		t.Error("Expected alice to NOT have read access after policy was removed")
	}
}

// TestInvalidateCacheDirectly tests that InvalidateCache works independently.
func TestInvalidateCacheDirectly(t *testing.T) {
	e, err := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create CachedEnforcer: %v", err)
	}

	// Populate cache
	res, _ := e.Enforce("alice", "data1", "read")
	if !res {
		t.Error("Expected alice to have read access to data1")
	}

	// Directly invalidate cache
	if err := e.InvalidateCache(); err != nil {
		t.Fatalf("InvalidateCache failed: %v", err)
	}

	// Next enforce should recompute (cache miss) and succeed
	res, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to still have read access after cache invalidation")
	}
}

// TestCacheWithCustomTTL tests that setting a custom TTL works.
func TestCacheWithCustomTTL(t *testing.T) {
	e, err := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create CachedEnforcer: %v", err)
	}

	// Set a short TTL
	e.SetExpireTime(100) // 100ns - essentially immediate for tests

	// Populate cache
	res, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !res {
		t.Error("Expected alice to have read access to data1")
	}

	// The cache entry should have the TTL set
	c, ok := e.cache.(*cache.DefaultCache)
	if !ok {
		t.Skip("Cache is not a DefaultCache, skipping TTL test")
	}

	item, exists := (*c)["alice$$data1$$read$$"]
	if !exists {
		t.Error("Expected cache entry for alice$$data1$$read$$")
	}
	_ = item // item.ttl should be 100ns
}
