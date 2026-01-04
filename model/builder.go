// Copyright 2026 The casbin Authors. All Rights Reserved.
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
)

// Builder provides a programmatic way to construct Casbin models.
// It allows creating models without needing a model.conf file.
//
// Example usage:
//
//	m, _ := model.New().
//	    Request("sub", "obj", "act").
//	    Policy("sub", "obj", "act").
//	    Role("_", "_").
//	    Effect("some(where (p.eft == allow))").
//	    Matcher("g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act").
//	    Build()
//
// The resulting model is equivalent to one loaded from a model.conf file.
type Builder struct {
	requestDef string
	policyDef  string
	roleDef    string
	effectDef  string
	matcherDef string
}

// New creates a new model builder.
func New() *Builder {
	return &Builder{}
}

// Request sets the request definition with the provided fields.
// Example: Request("sub", "obj", "act")
func (b *Builder) Request(fields ...string) *Builder {
	b.requestDef = strings.Join(fields, ", ")
	return b
}

// Policy sets the policy definition with the provided fields.
// Example: Policy("sub", "obj", "act")
func (b *Builder) Policy(fields ...string) *Builder {
	b.policyDef = strings.Join(fields, ", ")
	return b
}

// Role sets the role definition.
// Example: Role("_", "_") for basic RBAC
func (b *Builder) Role(fields ...string) *Builder {
	b.roleDef = strings.Join(fields, ", ")
	return b
}

// Effect sets the policy effect.
// Example: Effect("some(where (p.eft == allow))")
func (b *Builder) Effect(effect string) *Builder {
	b.effectDef = effect
	return b
}

// Matcher sets the matcher expression.
// Example: Matcher("r.sub == p.sub && r.obj == p.obj && r.act == p.act")
func (b *Builder) Matcher(matcher string) *Builder {
	b.matcherDef = matcher
	return b
}

// Build creates a Model from the builder configuration.
func (b *Builder) Build() (Model, error) {
	m := NewModel()

	if b.requestDef != "" {
		m.AddDef("r", "r", b.requestDef)
	}

	if b.policyDef != "" {
		m.AddDef("p", "p", b.policyDef)
	}

	if b.roleDef != "" {
		m.AddDef("g", "g", b.roleDef)
	}

	if b.effectDef != "" {
		m.AddDef("e", "e", b.effectDef)
	}

	if b.matcherDef != "" {
		m.AddDef("m", "m", b.matcherDef)
	}

	return m, nil
}

// ToString returns the model as a string in CONF format.
func (b *Builder) ToString() string {
	var sb strings.Builder

	sb.WriteString("[request_definition]\n")
	if b.requestDef != "" {
		sb.WriteString("r = ")
		sb.WriteString(b.requestDef)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("[policy_definition]\n")
	if b.policyDef != "" {
		sb.WriteString("p = ")
		sb.WriteString(b.policyDef)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	if b.roleDef != "" {
		sb.WriteString("[role_definition]\n")
		sb.WriteString("g = ")
		sb.WriteString(b.roleDef)
		sb.WriteString("\n\n")
	}

	sb.WriteString("[policy_effect]\n")
	if b.effectDef != "" {
		sb.WriteString("e = ")
		sb.WriteString(b.effectDef)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("[matchers]\n")
	if b.matcherDef != "" {
		sb.WriteString("m = ")
		sb.WriteString(b.matcherDef)
		sb.WriteString("\n")
	}

	return sb.String()
}
