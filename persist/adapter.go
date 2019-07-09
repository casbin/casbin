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

package persist

import (
	"strings"

	"github.com/casbin/casbin/model"
)

// LoadPolicyLine loads a text line as a policy rule to model.
func LoadPolicyLine(line string, model model.Model) {
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}
	sec, ptype, rules := policyLineToSlice(line)
	model[sec][ptype].Policy = append(model[sec][ptype].Policy, rules[1:])
}

// LoadPolicyLineUnique loads a text line as a policy rule to a model at most once.
func LoadPolicyLineUnique(line string, model model.Model) {
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}
	model.AddPolicy(policyLineToSlice(line))
}

// policyLineToSlice cleans and slices a policy line string.
func policyLineToSlice(line string) (sec string, ptype string, rules []string) {
	tokens := strings.Split(line, ",")
	for i := 0; i < len(tokens); i++ {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	ptype = tokens[0]
	sec = ptype[:1]
	tokens = tokens[1:]
	return
}

// Adapter is the interface for Casbin adapters.
type Adapter interface {
	// LoadPolicy loads all policy rules from the storage.
	LoadPolicy(model model.Model) error
	// SavePolicy saves all policy rules to the storage.
	SavePolicy(model model.Model) error

	// AddPolicy adds a policy rule to the storage.
	// This is part of the Auto-Save feature.
	AddPolicy(sec string, ptype string, rule []string) error
	// RemovePolicy removes a policy rule from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicy(sec string, ptype string, rule []string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
	// This is part of the Auto-Save feature.
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error
}
