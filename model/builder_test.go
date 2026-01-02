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

package model

import (
	"strings"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Error("NewBuilder() should return a non-nil builder")
	}
}

func TestBuilderBasicModel(t *testing.T) {
	// Build a basic ACL model programmatically
	m, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(AllowOverride).
		Matcher(And(Eq("sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify request definition
	if m["r"]["r"].Value != "sub, obj, act" {
		t.Errorf("Request definition incorrect: got %q, want %q", m["r"]["r"].Value, "sub, obj, act")
	}

	// Verify policy definition
	if m["p"]["p"].Value != "sub, obj, act" {
		t.Errorf("Policy definition incorrect: got %q, want %q", m["p"]["p"].Value, "sub, obj, act")
	}

	// Verify effect
	if m["e"]["e"].Value != AllowOverride {
		t.Errorf("Effect incorrect: got %q, want %q", m["e"]["e"].Value, AllowOverride)
	}

	// Verify matcher (note: matcher is transformed by AddDef)
	expectedMatcher := "r_sub == p_sub && r_obj == p_obj && r_act == p_act"
	if m["m"]["m"].Value != expectedMatcher {
		t.Errorf("Matcher incorrect: got %q, want %q", m["m"]["m"].Value, expectedMatcher)
	}
}

func TestBuilderRBACModel(t *testing.T) {
	// Build an RBAC model programmatically
	m, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		RoleDefinition("_", "_").
		Effect(AllowOverride).
		Matcher(And(G("r.sub", "p.sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify role definition
	if m["g"]["g"].Value != "_, _" {
		t.Errorf("Role definition incorrect: got %q, want %q", m["g"]["g"].Value, "_, _")
	}

	// Verify matcher includes role matching (note: matcher is transformed by AddDef)
	expectedMatcher := "g(r_sub, p_sub) && r_obj == p_obj && r_act == p_act"
	if m["m"]["m"].Value != expectedMatcher {
		t.Errorf("Matcher incorrect: got %q, want %q", m["m"]["m"].Value, expectedMatcher)
	}
}

func TestBuilderEquivalentToFileModel(t *testing.T) {
	// Load model from file
	fileModel, err := NewModelFromFile("../examples/basic_model.conf")
	if err != nil {
		t.Fatalf("Failed to load model from file: %v", err)
	}

	// Build equivalent model programmatically
	builtModel, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(AllowOverride).
		Matcher(And(Eq("sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Compare key components
	if fileModel["r"]["r"].Value != builtModel["r"]["r"].Value {
		t.Errorf("Request definitions don't match: file=%q, built=%q",
			fileModel["r"]["r"].Value, builtModel["r"]["r"].Value)
	}

	if fileModel["p"]["p"].Value != builtModel["p"]["p"].Value {
		t.Errorf("Policy definitions don't match: file=%q, built=%q",
			fileModel["p"]["p"].Value, builtModel["p"]["p"].Value)
	}

	if fileModel["e"]["e"].Value != builtModel["e"]["e"].Value {
		t.Errorf("Effects don't match: file=%q, built=%q",
			fileModel["e"]["e"].Value, builtModel["e"]["e"].Value)
	}

	if fileModel["m"]["m"].Value != builtModel["m"]["m"].Value {
		t.Errorf("Matchers don't match: file=%q, built=%q",
			fileModel["m"]["m"].Value, builtModel["m"]["m"].Value)
	}
}

func TestBuilderRBACEquivalentToFile(t *testing.T) {
	// Load RBAC model from file
	fileModel, err := NewModelFromFile("../examples/rbac_model.conf")
	if err != nil {
		t.Fatalf("Failed to load model from file: %v", err)
	}

	// Build equivalent RBAC model programmatically
	builtModel, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		RoleDefinition("_", "_").
		Effect(AllowOverride).
		Matcher(And(G("r.sub", "p.sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Compare all sections
	if fileModel["r"]["r"].Value != builtModel["r"]["r"].Value {
		t.Errorf("Request definitions don't match: file=%q, built=%q",
			fileModel["r"]["r"].Value, builtModel["r"]["r"].Value)
	}

	if fileModel["p"]["p"].Value != builtModel["p"]["p"].Value {
		t.Errorf("Policy definitions don't match: file=%q, built=%q",
			fileModel["p"]["p"].Value, builtModel["p"]["p"].Value)
	}

	if fileModel["g"]["g"].Value != builtModel["g"]["g"].Value {
		t.Errorf("Role definitions don't match: file=%q, built=%q",
			fileModel["g"]["g"].Value, builtModel["g"]["g"].Value)
	}

	if fileModel["e"]["e"].Value != builtModel["e"]["e"].Value {
		t.Errorf("Effects don't match: file=%q, built=%q",
			fileModel["e"]["e"].Value, builtModel["e"]["e"].Value)
	}

	if fileModel["m"]["m"].Value != builtModel["m"]["m"].Value {
		t.Errorf("Matchers don't match: file=%q, built=%q",
			fileModel["m"]["m"].Value, builtModel["m"]["m"].Value)
	}
}

func TestBuilderMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *Builder
		wantErr string
	}{
		{
			name: "missing request",
			builder: func() *Builder {
				return NewBuilder().
					Policy("sub", "obj", "act").
					Effect(AllowOverride).
					Matcher(Eq("sub"))
			},
			wantErr: "request definition is required",
		},
		{
			name: "missing policy",
			builder: func() *Builder {
				return NewBuilder().
					Request("sub", "obj", "act").
					Effect(AllowOverride).
					Matcher(Eq("sub"))
			},
			wantErr: "policy definition is required",
		},
		{
			name: "missing effect",
			builder: func() *Builder {
				return NewBuilder().
					Request("sub", "obj", "act").
					Policy("sub", "obj", "act").
					Matcher(Eq("sub"))
			},
			wantErr: "effect definition is required",
		},
		{
			name: "missing matcher",
			builder: func() *Builder {
				return NewBuilder().
					Request("sub", "obj", "act").
					Policy("sub", "obj", "act").
					Effect(AllowOverride)
			},
			wantErr: "matcher definition is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder().Build()
			if err == nil {
				t.Error("Build() should return an error")
			} else if err.Error() != tt.wantErr {
				t.Errorf("Build() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBuilderHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "Eq helper",
			fn:       func() string { return Eq("sub") },
			expected: "r.sub == p.sub",
		},
		{
			name:     "G helper",
			fn:       func() string { return G("r.sub", "p.sub") },
			expected: "g(r.sub, p.sub)",
		},
		{
			name:     "And helper",
			fn:       func() string { return And(Eq("sub"), Eq("obj")) },
			expected: "r.sub == p.sub && r.obj == p.obj",
		},
		{
			name:     "Or helper",
			fn:       func() string { return Or(Eq("sub"), Eq("obj")) },
			expected: "r.sub == p.sub || r.obj == p.obj",
		},
		{
			name: "Complex expression",
			// Note: And() and Or() are simple string concatenators.
			// Users should be aware of operator precedence when combining them.
			// For explicit grouping, use custom matcher strings with parentheses.
			fn:       func() string { return And(G("r.sub", "p.sub"), Or(Eq("obj"), Eq("act"))) },
			expected: "g(r.sub, p.sub) && r.obj == p.obj || r.act == p.act",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuilderChaining(t *testing.T) {
	// Test that builder methods return the same builder instance for chaining
	b1 := NewBuilder()
	b2 := b1.Request("sub")
	b3 := b2.Policy("sub")
	b4 := b3.Effect(AllowOverride)
	b5 := b4.Matcher(Eq("sub"))

	if b1 != b2 || b2 != b3 || b3 != b4 || b4 != b5 {
		t.Error("Builder methods should return the same instance for chaining")
	}
}

func TestBuilderDifferentEffects(t *testing.T) {
	effects := []string{
		AllowOverride,
		DenyOverride,
		AllowAndDeny,
		Priority,
		SubjectPriority,
	}

	for _, effect := range effects {
		m, err := NewBuilder().
			Request("sub", "obj", "act").
			Policy("sub", "obj", "act").
			Effect(effect).
			Matcher(Eq("sub")).
			Build()

		if err != nil {
			t.Fatalf("Build() with effect %q failed: %v", effect, err)
		}

		if m["e"]["e"].Value != effect {
			t.Errorf("Effect incorrect: got %q, want %q", m["e"]["e"].Value, effect)
		}
	}
}

func TestBuilderCustomMatcher(t *testing.T) {
	customMatcher := "r.sub == p.sub && keyMatch(r.obj, p.obj)"

	m, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(AllowOverride).
		Matcher(customMatcher).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Note: matcher is transformed by AddDef
	expectedMatcher := "r_sub == p_sub && keyMatch(r_obj, p_obj)"
	if m["m"]["m"].Value != expectedMatcher {
		t.Errorf("Custom matcher incorrect: got %q, want %q", m["m"]["m"].Value, expectedMatcher)
	}
}

func TestBuilderWithoutRoles(t *testing.T) {
	// Test building a model without role definition
	m, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(AllowOverride).
		Matcher(And(Eq("sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify that "g" section doesn't exist or is empty
	if m["g"] != nil && len(m["g"]) > 0 {
		t.Error("Model should not have role definition when not specified")
	}
}

func TestBuilderToText(t *testing.T) {
	// Test that a built model can be converted to text format
	m, err := NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(AllowOverride).
		Matcher(And(Eq("sub"), Eq("obj"), Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	text := m.ToText()
	if text == "" {
		t.Error("ToText() should return non-empty string")
	}

	// Verify the text contains expected sections
	if !strings.Contains(text, "[request_definition]") {
		t.Error("ToText() should contain [request_definition]")
	}
	if !strings.Contains(text, "[policy_definition]") {
		t.Error("ToText() should contain [policy_definition]")
	}
	if !strings.Contains(text, "[policy_effect]") {
		t.Error("ToText() should contain [policy_effect]")
	}
	if !strings.Contains(text, "[matchers]") {
		t.Error("ToText() should contain [matchers]")
	}

	// Verify the model can be loaded back from the text
	m2, err := NewModelFromString(text)
	if err != nil {
		t.Fatalf("Failed to load model from text: %v", err)
	}

	// Compare the two models
	if m["r"]["r"].Value != m2["r"]["r"].Value {
		t.Errorf("Request definitions don't match: original=%q, reloaded=%q",
			m["r"]["r"].Value, m2["r"]["r"].Value)
	}
}
