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
	"testing"

	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// TestBackslashHandlingConsistency tests that backslashes in string literals
// within matcher expressions are handled consistently with CSV-parsed values.
// This addresses the issue where govaluate interprets escape sequences in
// string literals, but CSV parsing treats backslashes as literal characters.
func TestBackslashHandlingConsistency(t *testing.T) {
	// Test case 1: Literal string in matcher should match CSV-parsed request
	t.Run("LiteralInMatcher", func(t *testing.T) {
		m := model.NewModel()
		m.AddDef("r", "r", "sub, obj, act")
		m.AddDef("p", "p", "sub, obj, act")
		m.AddDef("e", "e", "some(where (p.eft == allow))")
		// User writes '\1\2' in matcher - should be treated as literal backslashes
		m.AddDef("m", "m", "regexMatch('\\1\\2', p.obj)")

		e, err := NewEnforcer(m, fileadapter.NewAdapter("examples/basic_policy.csv"))
		if err != nil {
			t.Fatal(err)
		}

		// Add a policy with a regex pattern containing backslashes
		// CSV format: "\\[0-9]+\\" means literal string with 4 backslashes
		_, err = e.AddPolicy("filename", "\\\\[0-9]+\\\\", "read")
		if err != nil {
			t.Fatal(err)
		}

		// This should match because '\1\2' after escaping becomes \1\2,
		// and the pattern \\[0-9]+\\ matches strings like \1\ (which is a substring of \1\2)
		result, err := e.Enforce("filename", "dummy", "read")
		if err != nil {
			t.Fatal(err)
		}

		if !result {
			t.Errorf("Expected true, got false - literal '\\1\\2' should match after escape processing")
		}
	})

	// Test case 2: Request parameter should match policy with same backslash content
	t.Run("RequestParameterVsPolicy", func(t *testing.T) {
		m := model.NewModel()
		m.AddDef("r", "r", "sub, obj, act")
		m.AddDef("p", "p", "sub, obj, act")
		m.AddDef("e", "e", "some(where (p.eft == allow))")
		m.AddDef("m", "m", "regexMatch(r.obj, p.obj)")

		e, err := NewEnforcer(m, fileadapter.NewAdapter("examples/basic_policy.csv"))
		if err != nil {
			t.Fatal(err)
		}

		// Add policy with regex pattern
		_, err = e.AddPolicy("filename", "\\\\[0-9]+\\\\", "read")
		if err != nil {
			t.Fatal(err)
		}

		// Request with backslashes - simulating CSV input "\1\2" which becomes \1\2
		result, err := e.Enforce("filename", "\\1\\2", "read")
		if err != nil {
			t.Fatal(err)
		}

		if !result {
			t.Errorf("Expected true, got false - request \\1\\2 should match regex pattern")
		}
	})

	// Test case 3: Both approaches should give the same result
	t.Run("ConsistencyBetweenLiteralAndParameter", func(t *testing.T) {
		// Create two enforcers with different matchers
		m1 := model.NewModel()
		m1.AddDef("r", "r", "sub, obj, act")
		m1.AddDef("p", "p", "sub, obj, act")
		m1.AddDef("e", "e", "some(where (p.eft == allow))")
		m1.AddDef("m", "m", "regexMatch('\\1\\2', p.obj)")

		m2 := model.NewModel()
		m2.AddDef("r", "r", "sub, obj, act")
		m2.AddDef("p", "p", "sub, obj, act")
		m2.AddDef("e", "e", "some(where (p.eft == allow))")
		m2.AddDef("m", "m", "regexMatch(r.obj, p.obj)")

		e1, err := NewEnforcer(m1, fileadapter.NewAdapter("examples/basic_policy.csv"))
		if err != nil {
			t.Fatal(err)
		}
		e2, err := NewEnforcer(m2, fileadapter.NewAdapter("examples/basic_policy.csv"))
		if err != nil {
			t.Fatal(err)
		}

		// Add same policy to both
		pattern := "\\\\[0-9]+\\\\"
		_, err = e1.AddPolicy("filename", pattern, "read")
		if err != nil {
			t.Fatal(err)
		}
		_, err = e2.AddPolicy("filename", pattern, "read")
		if err != nil {
			t.Fatal(err)
		}

		// Test with the same request
		result1, err := e1.Enforce("filename", "dummy", "read")
		if err != nil {
			t.Fatal(err)
		}

		result2, err := e2.Enforce("filename", "\\1\\2", "read")
		if err != nil {
			t.Fatal(err)
		}

		if result1 != result2 {
			t.Errorf("Inconsistent results: literal in matcher gave %v, parameter gave %v", result1, result2)
		}
	})

	// Test case 4: Simple equality check with backslashes
	t.Run("SimpleEqualityWithBackslashes", func(t *testing.T) {
		m := model.NewModel()
		m.AddDef("r", "r", "sub, obj, act")
		m.AddDef("p", "p", "sub, obj, act")
		m.AddDef("e", "e", "some(where (p.eft == allow))")
		// In Go source, '\test' (one backslash in the actual string) represents
		// what would be typed in a web form. After escape processing, it will match
		// the CSV-parsed value "\test" (one backslash).
		// Note: In Go source, we write "r.obj == '\\test'" which is a Go string
		// containing the text: r.obj == '\test' (with ONE backslash in the string content)
		m.AddDef("m", "m", "r.obj == '\\test' && p.sub == r.sub")

		e, err := NewEnforcer(m, fileadapter.NewAdapter("examples/basic_policy.csv"))
		if err != nil {
			t.Fatal(err)
		}

		_, err = e.AddPolicy("alice", "any", "read")
		if err != nil {
			t.Fatal(err)
		}

		// Request with literal backslash from CSV would be "\test"
		// In Go source, we write "\\test" which represents the string \test (one backslash)
		result, err := e.Enforce("alice", "\\test", "read")
		if err != nil {
			t.Fatal(err)
		}

		if !result {
			t.Errorf("Expected true - literal '\\test' should equal request parameter \\test")
		}
	})
}
