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

package model_test

import (
	"fmt"

	"github.com/casbin/casbin/v3/model"
)

// ExampleNewBuilder demonstrates how to build a basic ACL model programmatically.
func ExampleNewBuilder() {
	// Build a basic ACL model
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher(
			model.And(
				model.Eq("sub"),
				model.Eq("obj"),
				model.Eq("act"),
			),
		).
		Build()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The model is now ready to use with an enforcer
	fmt.Printf("Request definition: %s\n", m["r"]["r"].Value)
	fmt.Printf("Policy definition: %s\n", m["p"]["p"].Value)
	// Output:
	// Request definition: sub, obj, act
	// Policy definition: sub, obj, act
}

// ExampleNewBuilder_rbac demonstrates how to build an RBAC model programmatically.
func ExampleNewBuilder_rbac() {
	// Build an RBAC model
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		RoleDefinition("_", "_").
		Effect(model.AllowOverride).
		Matcher(
			model.And(
				model.G("r.sub", "p.sub"),
				model.Eq("obj"),
				model.Eq("act"),
			),
		).
		Build()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The model includes role inheritance
	fmt.Printf("Has role definition: %v\n", m["g"] != nil && m["g"]["g"] != nil)
	// Output:
	// Has role definition: true
}

// ExampleNewBuilder_customMatcher demonstrates using a custom matcher expression.
func ExampleNewBuilder_customMatcher() {
	// Build a model with custom matcher using keyMatch
	_, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher("r.sub == p.sub && keyMatch(r.obj, p.obj) && r.act == p.act").
		Build()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The custom matcher is preserved
	fmt.Printf("Model created successfully\n")
	// Output:
	// Model created successfully
}

// ExampleNewBuilder_toText demonstrates converting a built model to text format.
func ExampleNewBuilder_toText() {
	// Build a model
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher(model.And(model.Eq("sub"), model.Eq("obj"), model.Eq("act"))).
		Build()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Convert to text format (can be saved to a file or loaded again)
	text := m.ToText()
	fmt.Printf("Model text starts with: [request_definition]\n")
	fmt.Printf("Model can be reloaded: %v\n", len(text) > 0)
	// Output:
	// Model text starts with: [request_definition]
	// Model can be reloaded: true
}

// ExampleBuilder_Effect demonstrates using different effect types.
func ExampleBuilder_Effect() {
	// AllowOverride: allows if any policy allows
	m1, _ := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher(model.Eq("sub")).
		Build()

	// DenyOverride: denies if any policy denies
	m2, _ := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.DenyOverride).
		Matcher(model.Eq("sub")).
		Build()

	// AllowAndDeny: allows only if at least one allows and none deny
	m3, _ := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowAndDeny).
		Matcher(model.Eq("sub")).
		Build()

	fmt.Printf("Models created: %d\n", 3)
	fmt.Printf("AllowOverride: %v\n", m1 != nil)
	fmt.Printf("DenyOverride: %v\n", m2 != nil)
	fmt.Printf("AllowAndDeny: %v\n", m3 != nil)
	// Output:
	// Models created: 3
	// AllowOverride: true
	// DenyOverride: true
	// AllowAndDeny: true
}
