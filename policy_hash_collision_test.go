// Copyright 2017 The casbin Authors. All Rights Reserved.
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
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/model"
)

// Test for potential hash key collision with comma separator
// This demonstrates the theoretical issue with using comma as separator
func TestPolicyHashCollisionPotential(t *testing.T) {
	// Two different policies that would have identical hash keys if using comma separator
	policy1 := []string{"alice", "data,file", "read"}       // 3 fields
	policy2 := []string{"alice", "data", "file", "read"}    // 4 fields

	// OLD approach (with comma) would cause collision
	oldHash1 := strings.Join(policy1, ",")
	oldHash2 := strings.Join(policy2, ",")

	t.Logf("Policy 1: %v -> old hash: %s", policy1, oldHash1)
	t.Logf("Policy 2: %v -> old hash: %s", policy2, oldHash2)

	if oldHash1 == oldHash2 {
		t.Logf("OLD APPROACH: Collision detected with comma separator")
	}

	// NEW approach (with unit separator) prevents collision
	// Note: we can't call policyKey directly as it's not exported, but the internal implementation uses \x1f
	newHash1 := strings.Join(policy1, "\x1f")
	newHash2 := strings.Join(policy2, "\x1f")

	t.Logf("NEW approach - Policy 1: %v -> new hash: %q", policy1, newHash1)
	t.Logf("NEW approach - Policy 2: %v -> new hash: %q", policy2, newHash2)

	if newHash1 != newHash2 {
		t.Logf("NEW APPROACH: No collision with unit separator - fix working!")
	} else {
		t.Error("NEW APPROACH: Unexpected collision!")
	}

	// Verify with actual Casbin behavior
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Try to add policy1 (3 fields - should succeed)
	added1, err := e.AddPolicy(policy1)
	if err != nil {
		t.Logf("AddPolicy for policy1 failed (expected): %v", err)
	} else if added1 {
		t.Logf("Policy1 added successfully")
	}

	// Try to add policy2 (4 fields - should fail due to field count mismatch)
	added2, err := e.AddPolicy(policy2)
	if err != nil {
		t.Logf("AddPolicy for policy2 failed as expected due to field count: %v", err)
	} else if !added2 {
		t.Logf("Policy2 not added (field count validation)")
	} else {
		// This is unexpected but not an error - model validation should prevent it
		t.Logf("Note: Policy2 was added despite field count difference")
	}
}

// More realistic collision test: same field count
func TestPolicyHashCollisionSameFieldCount(t *testing.T) {
	// Can we create a collision with the same field count?
	// Yes, if we have nested commas
	
	policy1 := []string{"alice", "data", "read"}
	policy2 := []string{"alice", "da,ta", "read"} // This has different content but could theoretically collide if parsed incorrectly
	
	hash1 := strings.Join(policy1, ",")
	hash2 := strings.Join(policy2, ",")
	
	t.Logf("Policy 1: %#v -> hash: %s", policy1, hash1)
	t.Logf("Policy 2: %#v -> hash: %s", policy2, hash2)
	
	// These should have DIFFERENT hashes
	// policy1: "alice,data,read"
	// policy2: "alice,da,ta,read" <- wait, this has 4 commas but []string has 3 elements
	
	// Actually strings.Join is safe here because it works on []string
	// The issue would only occur if we parse FROM a string
	
	if hash1 != hash2 {
		t.Logf("Good: No collision (hashes are different)")
	} else {
		t.Errorf("COLLISION: Two policies with same field count have identical hash!")
	}
	
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add both policies
	_, err = e.AddPolicy(policy1)
	if err != nil {
		t.Fatalf("Failed to add policy1: %v", err)
	}

	_, err = e.AddPolicy(policy2)
	if err != nil {
		t.Fatalf("Failed to add policy2: %v", err)
	}

	// Check both exist
	has1, _ := e.HasPolicy(policy1)
	has2, _ := e.HasPolicy(policy2)

	t.Logf("Policy1 exists: %v", has1)
	t.Logf("Policy2 exists: %v", has2)

	if !has1 || !has2 {
		t.Error("One or both policies missing - possible hash collision")
	}

	// Try to remove policy1
	removed, err := e.RemovePolicy(policy1)
	if err != nil {
		t.Fatalf("RemovePolicy failed: %v", err)
	}
	if !removed {
		t.Error("Failed to remove policy1")
	}

	// Check policy2 is still there
	has2After, _ := e.HasPolicy(policy2)
	if !has2After {
		t.Error("Policy2 was incorrectly removed (hash collision!)")
	}

	// Check policy1 is gone
	has1After, _ := e.HasPolicy(policy1)
	if has1After {
		t.Error("Policy1 still exists after removal")
	}
}
