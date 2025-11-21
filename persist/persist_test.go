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

package persist_test

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

func TestPersist(t *testing.T) {
	// No tests yet
}

func testRuleCount(t *testing.T, model model.Model, expected int, sec string, ptype string, tag string) {
	t.Helper()

	ruleCount := len(model[sec][ptype].Policy)
	if ruleCount != expected {
		t.Errorf("[%s] rule count: %d, expected %d", tag, ruleCount, expected)
	}
}

func TestDuplicateRuleInAdapter(t *testing.T) {
	e, _ := casbin.NewEnforcer("../examples/basic_model.conf")

	_, _ = e.AddPolicy("alice", "data1", "read")
	_, _ = e.AddPolicy("alice", "data1", "read")

	testRuleCount(t, e.GetModel(), 1, "p", "p", "AddPolicy")

	e.ClearPolicy()

	// simulate adapter.LoadPolicy with duplicate rules
	_ = persist.LoadPolicyArray([]string{"p", "alice", "data1", "read"}, e.GetModel())
	_ = persist.LoadPolicyArray([]string{"p", "alice", "data1", "read"}, e.GetModel())

	testRuleCount(t, e.GetModel(), 1, "p", "p", "LoadPolicyArray")
}

func TestStringToNullable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected persist.NullableString
	}{
		{
			name:  "non-empty string",
			input: "alice",
			expected: persist.NullableString{
				Value: "alice",
				Valid: true,
			},
		},
		{
			name:  "empty string",
			input: "",
			expected: persist.NullableString{
				Value: "",
				Valid: true, // Empty strings should be valid
			},
		},
		{
			name:  "whitespace",
			input: " ",
			expected: persist.NullableString{
				Value: " ",
				Valid: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := persist.StringToNullable(tt.input)
			if result.Value != tt.expected.Value {
				t.Errorf("StringToNullable(%q) Value = %q, expected %q", tt.input, result.Value, tt.expected.Value)
			}
			if result.Valid != tt.expected.Valid {
				t.Errorf("StringToNullable(%q) Valid = %v, expected %v", tt.input, result.Valid, tt.expected.Valid)
			}
		})
	}
}

func TestNullableToString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		valid    bool
		expected string
	}{
		{
			name:     "valid non-empty string",
			value:    "alice",
			valid:    true,
			expected: "alice",
		},
		{
			name:     "valid empty string",
			value:    "",
			valid:    true,
			expected: "", // Valid empty string should remain empty
		},
		{
			name:     "invalid (NULL)",
			value:    "",
			valid:    false,
			expected: "", // NULL should become empty string
		},
		{
			name:     "valid whitespace",
			value:    " ",
			valid:    true,
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := persist.NullableToString(tt.value, tt.valid)
			if result != tt.expected {
				t.Errorf("NullableToString(%q, %v) = %q, expected %q", tt.value, tt.valid, result, tt.expected)
			}
		})
	}
}

func TestEmptyStringPreservation(t *testing.T) {
	// Test round-trip conversion
	testCases := []string{"", "alice", " ", "data1"}

	for _, original := range testCases {
		t.Run("round-trip: "+original, func(t *testing.T) {
			// Simulate saving to database
			nullable := persist.StringToNullable(original)

			// Simulate loading from database
			recovered := persist.NullableToString(nullable.Value, nullable.Valid)

			if recovered != original {
				t.Errorf("Round-trip failed: original=%q, recovered=%q", original, recovered)
			}
		})
	}
}

func TestDatabaseAdapterEmptyStringUsage(t *testing.T) {
	// This test demonstrates how database adapters should use the helper functions
	// to correctly handle empty strings in policy rules

	// Example policy rule with an empty string field
	originalRule := []string{"alice", "", "read"}

	// Simulate converting rule for database storage
	var dbFields []persist.NullableString
	for _, field := range originalRule {
		dbFields = append(dbFields, persist.StringToNullable(field))
	}

	// Verify all fields are marked as valid (including empty string)
	for i, dbField := range dbFields {
		if !dbField.Valid {
			t.Errorf("Field %d should be valid, got Valid=%v", i, dbField.Valid)
		}
	}

	// Verify empty string is preserved
	if dbFields[1].Value != "" {
		t.Errorf("Empty string field should have empty Value, got %q", dbFields[1].Value)
	}

	// Simulate loading rule from database
	var recoveredRule []string
	for _, dbField := range dbFields {
		recoveredRule = append(recoveredRule, persist.NullableToString(dbField.Value, dbField.Valid))
	}

	// Verify round-trip preservation
	if len(recoveredRule) != len(originalRule) {
		t.Errorf("Rule length mismatch: got %d, expected %d", len(recoveredRule), len(originalRule))
	}

	for i := range originalRule {
		if recoveredRule[i] != originalRule[i] {
			t.Errorf("Field %d mismatch: got %q, expected %q", i, recoveredRule[i], originalRule[i])
		}
	}
}
