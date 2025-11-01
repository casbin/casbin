package casbin

import (
"testing"
)

func TestBugPatternInString(t *testing.T) {
e, err := NewEnforcer("examples/bug_test_model.conf")
if err != nil {
t.Fatalf("Error: %v\n", err)
}
testEnforce(t, e, "a.p.p.l.e", "file", "read", true)
}
