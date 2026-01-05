package casbin

import (
"testing"
)

func TestDebugRM(t *testing.T) {
e, err := NewEnforcer("examples/rbac_model.conf")
if err != nil {
t.Fatalf("Failed to create enforcer: %v", err)
}

t.Logf("After NewEnforcer (with no policy file):")
if assertion, ok := e.model["g"]["g"]; ok {
t.Logf("  model['g']['g'].RM == nil: %v", assertion.RM == nil)
t.Logf("  rmMap['g'] == nil: %v", e.rmMap["g"] == nil)
if e.rmMap["g"] != nil && assertion.RM != nil {
t.Logf("  rmMap['g'] == model['g']['g'].RM: %v", e.rmMap["g"] == assertion.RM)
}
}

// Now add a grouping policy
_, err = e.AddGroupingPolicy("alice", "admin")
if err != nil {
t.Fatalf("Failed to add grouping policy: %v", err)
}

t.Logf("After AddGroupingPolicy:")
if assertion, ok := e.model["g"]["g"]; ok {
t.Logf("  model['g']['g'].RM == nil: %v", assertion.RM == nil)
t.Logf("  rmMap['g'] == nil: %v", e.rmMap["g"] == nil)
if e.rmMap["g"] != nil && assertion.RM != nil {
t.Logf("  rmMap['g'] == model['g']['g'].RM: %v", e.rmMap["g"] == assertion.RM)
}
}
}
