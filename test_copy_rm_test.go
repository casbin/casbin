package casbin

import (
"testing"
)

// TestModelCopyPreservesRM tests that copying a model preserves RM and CondRM fields
func TestModelCopyPreservesRM(t *testing.T) {
e, err := NewEnforcer("examples/rbac_model.conf")
if err != nil {
t.Fatalf("Failed to create enforcer: %v", err)
}

// Verify RM is set in original model
originalAssertion := e.model["g"]["g"]
if originalAssertion.RM == nil {
t.Fatal("RM should not be nil in original model")
}
originalRM := originalAssertion.RM

// Copy the model
copiedModel := e.model.Copy()

// Verify RM is preserved in copied model
copiedAssertion := copiedModel["g"]["g"]
if copiedAssertion.RM == nil {
t.Error("RM should not be nil in copied model (this was the bug)")
}

if copiedAssertion.RM != originalRM {
t.Error("RM should be the same object in copied model")
}
}

// TestEnforceAfterModelCopyWithoutBuildRoleLinks tests that enforce works
// even if model is copied but BuildRoleLinks is not called
func TestEnforceAfterModelCopyWithoutBuildRoleLinks(t *testing.T) {
e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
if err != nil {
t.Fatalf("Failed to create enforcer: %v", err)
}

// Verify alice can access data2 initially
ok, err := e.Enforce("alice", "data2", "read")
if err != nil {
t.Fatalf("Enforce failed: %v", err)
}
if !ok {
t.Error("alice should have read access to data2")
}

// Copy the model and replace (simulating some scenario where model is replaced)
copiedModel := e.model.Copy()
e.model = copiedModel

// Try to enforce again - this should still work because RM is preserved
ok, err = e.Enforce("alice", "data2", "read")
if err != nil {
t.Fatalf("Enforce failed after model copy: %v (this demonstrates the bug)", err)
}
if !ok {
t.Error("alice should still have read access to data2 after model copy")
}
}
