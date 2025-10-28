// Copyright 2018 The casbin Authors. All Rights Reserved.
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
	"testing"
	"time"
)

func testEnforceCache(t *testing.T, e *CachedEnforcer, sub string, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, obj, act); myRes != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestCache(t *testing.T) {
	e, _ := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	// The cache is enabled by default for NewCachedEnforcer.

	testEnforceCache(t, e, "alice", "data1", "read", true)
	testEnforceCache(t, e, "alice", "data1", "write", false)
	testEnforceCache(t, e, "alice", "data2", "read", false)
	testEnforceCache(t, e, "alice", "data2", "write", false)

	// The cache is enabled, calling RemovePolicy, LoadPolicy or RemovePolicies will
	// also operate cached items.
	_, _ = e.RemovePolicy("alice", "data1", "read")

	testEnforceCache(t, e, "alice", "data1", "read", false)
	testEnforceCache(t, e, "alice", "data1", "write", false)
	testEnforceCache(t, e, "alice", "data2", "read", false)
	testEnforceCache(t, e, "alice", "data2", "write", false)

	e, _ = NewCachedEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testEnforceCache(t, e, "alice", "data1", "read", true)
	testEnforceCache(t, e, "bob", "data2", "write", true)
	testEnforceCache(t, e, "alice", "data2", "read", true)
	testEnforceCache(t, e, "alice", "data2", "write", true)

	_, _ = e.RemovePolicies([][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	})

	testEnforceCache(t, e, "alice", "data1", "read", false)
	testEnforceCache(t, e, "bob", "data2", "write", false)
	testEnforceCache(t, e, "alice", "data2", "read", true)
	testEnforceCache(t, e, "alice", "data2", "write", true)

	e, _ = NewCachedEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	testEnforceCache(t, e, "alice", "data1", "read", true)
	testEnforceCache(t, e, "bob", "data2", "write", true)
	testEnforceCache(t, e, "alice", "data2", "read", true)
	testEnforceCache(t, e, "alice", "data2", "write", true)

	e.ClearPolicy()

	testEnforceCache(t, e, "alice", "data1", "read", false)
	testEnforceCache(t, e, "bob", "data2", "write", false)
	testEnforceCache(t, e, "alice", "data2", "read", false)
	testEnforceCache(t, e, "alice", "data2", "write", false)
}

// TestCacheNeverExpires verifies that cache entries never expire when expireTime is set to 0 or negative.
// This is useful in multi-instance scenarios where you want to avoid lock contention and recalculation overhead,
// and instead manually trigger LoadPolicy() to refresh cache when policies change.
func TestCacheNeverExpires(t *testing.T) {
	e, _ := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	
	// Set cache to never expire (0 or negative duration)
	e.SetExpireTime(0)
	
	// First enforcement creates cache entry
	testEnforceCache(t, e, "alice", "data1", "read", true)
	
	// Wait a bit to ensure time has passed
	time.Sleep(10 * time.Millisecond)
	
	// Cache should still be valid (never expires)
	testEnforceCache(t, e, "alice", "data1", "read", true)
	
	// Remove the policy from the underlying model
	_, _ = e.Enforcer.RemovePolicy("alice", "data1", "read")
	
	// Cache still returns true because it hasn't been invalidated
	testEnforceCache(t, e, "alice", "data1", "read", true)
	
	// Manually invalidate cache (simulating notification from another instance)
	_ = e.InvalidateCache()
	
	// Now the cache is cleared, so it should return false
	testEnforceCache(t, e, "alice", "data1", "read", false)
}

// TestCacheWithExpiration verifies that cache entries expire after the specified duration.
func TestCacheWithExpiration(t *testing.T) {
	e, _ := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	
	// Set cache to expire after 50ms
	e.SetExpireTime(50 * time.Millisecond)
	
	// First enforcement creates cache entry
	testEnforceCache(t, e, "alice", "data1", "read", true)
	
	// Immediately check - should hit cache
	testEnforceCache(t, e, "alice", "data1", "read", true)
	
	// Wait for cache to expire
	time.Sleep(60 * time.Millisecond)
	
	// Remove the policy from the underlying model
	_, _ = e.Enforcer.RemovePolicy("alice", "data1", "read")
	
	// Cache has expired, so it should re-evaluate and return false
	testEnforceCache(t, e, "alice", "data1", "read", false)
}

// TestCacheLoadPolicyClearsCache verifies that LoadPolicy() clears the cache.
// This is important for multi-instance scenarios where one instance notifies others
// to reload policies when changes occur.
func TestCacheLoadPolicyClearsCache(t *testing.T) {
	e, _ := NewCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	
	// Set cache to never expire
	e.SetExpireTime(0)
	
	// Create cache entries
	testEnforceCache(t, e, "alice", "data1", "read", true)
	testEnforceCache(t, e, "alice", "data1", "write", false)
	
	// Add a new policy
	_, _ = e.AddPolicy("alice", "data2", "read")
	
	// Cache doesn't have this entry yet
	testEnforceCache(t, e, "alice", "data2", "read", true)
	
	// LoadPolicy clears cache and reloads from source (which doesn't have alice,data2,read)
	_ = e.LoadPolicy()
	
	// After reload, the added policy is gone (since it wasn't in the file)
	// and cache is cleared, so it re-evaluates
	testEnforceCache(t, e, "alice", "data2", "read", false)
	
	// Original policies still work
	testEnforceCache(t, e, "alice", "data1", "read", true)
}
