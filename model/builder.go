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
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/constant"
)

// Builder provides a programmatic way to construct Casbin models.
type Builder struct {
	requestFields []string
	policyFields  []string
	roleFields    []string
	effect        string
	matcher       string
}

// NewBuilder creates a new model builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Request defines the request definition fields.
// Example: Request("sub", "obj", "act").
func (b *Builder) Request(fields ...string) *Builder {
	b.requestFields = fields
	return b
}

// Policy defines the policy definition fields.
// Example: Policy("sub", "obj", "act").
func (b *Builder) Policy(fields ...string) *Builder {
	b.policyFields = fields
	return b
}

// RoleDefinition defines the role definition.
// Example: RoleDefinition("_", "_") for basic RBAC.
// Example: RoleDefinition("_", "_", "_") for RBAC with domains.
func (b *Builder) RoleDefinition(fields ...string) *Builder {
	b.roleFields = fields
	return b
}

// Effect defines the policy effect.
// Use constants from constant package or provide custom effect expression.
func (b *Builder) Effect(effect string) *Builder {
	b.effect = effect
	return b
}

// Matcher defines the matcher expression.
// Can be a string expression or use helper functions like And(), Or(), Eq().
func (b *Builder) Matcher(matcher string) *Builder {
	b.matcher = matcher
	return b
}

// Build constructs and returns a standard Casbin Model.
func (b *Builder) Build() (Model, error) {
	if len(b.requestFields) == 0 {
		return nil, fmt.Errorf("request definition is required")
	}
	if len(b.policyFields) == 0 {
		return nil, fmt.Errorf("policy definition is required")
	}
	if b.effect == "" {
		return nil, fmt.Errorf("effect definition is required")
	}
	if b.matcher == "" {
		return nil, fmt.Errorf("matcher definition is required")
	}

	m := NewModel()

	// Add request definition
	requestDef := strings.Join(b.requestFields, ", ")
	m.AddDef("r", "r", requestDef)

	// Add policy definition
	policyDef := strings.Join(b.policyFields, ", ")
	m.AddDef("p", "p", policyDef)

	// Add role definition if specified
	if len(b.roleFields) > 0 {
		roleDef := strings.Join(b.roleFields, ", ")
		m.AddDef("g", "g", roleDef)
	}

	// Add effect
	m.AddDef("e", "e", b.effect)

	// Add matcher
	m.AddDef("m", "m", b.matcher)

	return m, nil
}

// Helper functions for building matcher expressions

// Eq creates an equality matcher for a field.
// Example: Eq("sub") generates "r.sub == p.sub".
func Eq(field string) string {
	return fmt.Sprintf("r.%s == p.%s", field, field)
}

// G creates a role matching expression.
// Example: G("r.sub", "p.sub") generates "g(r.sub, p.sub)".
func G(role1, role2 string) string {
	return fmt.Sprintf("g(%s, %s)", role1, role2)
}

// And combines multiple expressions with logical AND.
// Example: And(Eq("sub"), Eq("obj")) generates "r.sub == p.sub && r.obj == p.obj"
// Note: This is a simple string concatenator. Be aware of operator precedence
// when combining with Or(). For explicit grouping, use parentheses in custom matcher strings.
func And(expressions ...string) string {
	return strings.Join(expressions, " && ")
}

// Or combines multiple expressions with logical OR.
// Example: Or(Eq("sub"), Eq("obj")) generates "r.sub == p.sub || r.obj == p.obj"
// Note: This is a simple string concatenator. Be aware of operator precedence
// when combining with And(). For explicit grouping, use parentheses in custom matcher strings.
func Or(expressions ...string) string {
	return strings.Join(expressions, " || ")
}

// Effect constants for convenience.
const (
	// AllowOverride allows if any policy allows (some(where (p.eft == allow))).
	AllowOverride = constant.AllowOverrideEffect
	// DenyOverride denies if any policy denies (!some(where (p.eft == deny))).
	DenyOverride = constant.DenyOverrideEffect
	// AllowAndDeny allows only if at least one allows and none deny.
	AllowAndDeny = constant.AllowAndDenyEffect
	// Priority uses priority-based effect (priority(p.eft) || deny).
	Priority = constant.PriorityEffect
	// SubjectPriority uses subject priority-based effect (subjectPriority(p.eft) || deny).
	SubjectPriority = constant.SubjectPriorityEffect
)
