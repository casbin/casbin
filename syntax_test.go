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
	"testing"
)

// TestSyntaxMatcher tests that patterns like "p." inside quoted strings
// are not incorrectly escaped by EscapeAssertion.
// This addresses the bug where matchers containing strings like "a.p.p.l.e"
// would have the ".p." pattern incorrectly replaced with "_p_".
func TestSyntaxMatcher(t *testing.T) {
	e, err := NewEnforcer("examples/syntax_matcher_model.conf")
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
	// The matcher in syntax_matcher_model.conf is: m = r.sub == "a.p.p.l.e"
	// This should match when r.sub is exactly "a.p.p.l.e"
	testEnforce(t, e, "a.p.p.l.e", "file", "read", true)
	testEnforce(t, e, "other", "file", "read", false)
}
