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

package util

import (
	"strings"
	"testing"
)

func TestTransformBlockMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "non-block matcher",
			input:    "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act",
			expected: "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act",
		},
		{
			name: "simple block with let statements",
			input: `{
				let role_match = g(r.sub, p.sub)
				let obj_match = r.obj == p.obj
				let act_match = r.act == p.act
				return role_match && obj_match && act_match
			}`,
			expected: "(g(r.sub, p.sub)) && (r.obj == p.obj) && (r.act == p.act)",
		},
		{
			name: "block with nested let expressions",
			input: `{
				let role_match = g(r.sub, p.sub)
				let obj_direct_match = r.obj == p.obj
				let obj_inherit_match = g2(r.obj, p.obj)
				let obj_match = obj_direct_match || obj_inherit_match
				let act_match = r.act == p.act
				return role_match && obj_match && act_match
			}`,
			expected: "(g(r.sub, p.sub)) && ((r.obj == p.obj) || (g2(r.obj, p.obj))) && (r.act == p.act)",
		},
		{
			name: "block with single early return",
			input: `{
				let role_match = g(r.sub, p.sub)
				if !role_match {
					return false
				}
				return r.obj == p.obj
			}`,
			expected: "((!(g(r.sub, p.sub))) && (false)) || (!(!(g(r.sub, p.sub))) && (r.obj == p.obj))",
		},
		{
			name: "block with multiple early returns",
			input: `{
				let role_match = g(r.sub, p.sub)
				if !role_match {
					return false
				}
				if r.act != p.act {
					return false
				}
				if r.obj == p.obj {
					return true
				}
				if g2(r.obj, p.obj) {
					return true
				}
				return false
			}`,
			expected: "((!(g(r.sub, p.sub))) && (false)) || (!(!(g(r.sub, p.sub))) && (((r.act != p.act) && (false)) || (!(r.act != p.act) && (((r.obj == p.obj) && (true)) || (!(r.obj == p.obj) && (((g2(r.obj, p.obj)) && (true)) || (!(g2(r.obj, p.obj)) && (false))))))))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformBlockMatcher(tt.input)
			// Normalize whitespace for comparison
			result = strings.Join(strings.Fields(result), " ")
			expected := strings.Join(strings.Fields(tt.expected), " ")

			if result != expected {
				t.Errorf("TransformBlockMatcher() = %v, want %v", result, expected)
			}
		})
	}
}

func TestTransformBlockMatcherEdgeCases(t *testing.T) {
	// Test block with only return
	input := `{ return true }`
	result := TransformBlockMatcher(input)
	expected := "true"
	if strings.TrimSpace(result) != expected {
		t.Errorf("Block with only return should be transformed to %v, got: %v", expected, result)
	}

	// Test expression without braces
	input = "r.sub == p.sub"
	result = TransformBlockMatcher(input)
	if result != input {
		t.Errorf("Expression without braces should be unchanged, got: %v", result)
	}
}
