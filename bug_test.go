package casbin

import (
	"testing"
)

// TestBugPatternInString tests that patterns like "p." inside quoted strings
// are not incorrectly escaped by EscapeAssertion.
// This addresses the bug where matchers containing strings like "a.p.p.l.e"
// would have the ".p." pattern incorrectly replaced with "_p_".
func TestBugPatternInString(t *testing.T) {
	e, err := NewEnforcer("examples/bug_test_model.conf")
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
	// The matcher in bug_test_model.conf is: m = r.sub == "a.p.p.l.e"
	// This should match when r.sub is exactly "a.p.p.l.e"
	testEnforce(t, e, "a.p.p.l.e", "file", "read", true)
	testEnforce(t, e, "other", "file", "read", false)
}
