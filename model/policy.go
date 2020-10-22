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
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/rbac"
	"github.com/casbin/casbin/v3/util"
)

type (
	PolicyOp int
)

const (
	PolicyAdd PolicyOp = iota
	PolicyRemove
)

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (m Model) BuildIncrementalRoleLinks(rm rbac.RoleManager, op PolicyOp, sec string, ptype string, rules [][]string) error {
	if sec == "g" {
		return m[sec][ptype].BuildIncrementalRoleLinks(rm, op, rules)
	}
	return nil
}

// BuildRoleLinks initializes the roles in RBAC.
func (m Model) BuildRoleLinks(rm rbac.RoleManager) error {
	for _, ast := range m["g"] {
		err := ast.BuildRoleLinks(rm)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintPolicy prints the policy to log.
func (m Model) PrintPolicy() {
	log.LogPrint("Policy:")
	for key, ast := range m["p"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy.GetPolicy())
	}

	for key, ast := range m["g"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy.GetPolicy())
	}
}

// ClearPolicy clears all current policy.
func (m Model) ClearPolicy() {
	for _, ast := range m["p"] {
		ast.Policy.ClearPolicy()
	}

	for _, ast := range m["g"] {
		ast.Policy.ClearPolicy()
	}
}

// GetPolicy gets all rules in a policy.
func (m Model) GetPolicy(sec string, ptype string) [][]string {
	return m[sec][ptype].Policy.GetPolicy()
}

// GetFilteredPolicy gets rules based on field filters from a policy.
func (m Model) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return m[sec][ptype].Policy.GetFilteredPolicy(fieldIndex, fieldValues...)
}

// HasPolicy determines whether a model has the specified policy rule.
func (m Model) HasPolicy(sec string, ptype string, rule []string) bool {
	return m[sec][ptype].Policy.HasPolicy(rule)
}

// HasPolicies determines whether a model has any of the specified policies. If one is found we return false.
func (m Model) HasPolicies(sec string, ptype string, rules [][]string) bool {
	for i := 0; i < len(rules); i++ {
		if m.HasPolicy(sec, ptype, rules[i]) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (m Model) AddPolicy(sec string, ptype string, rule []string) bool {
	return m[sec][ptype].Policy.AddPolicy(rule)
}

// AddPolicies adds policy rules to the model.
func (m Model) AddPolicies(sec string, ptype string, rules [][]string) [][]string {
	return m[sec][ptype].Policy.AddPolicies(rules)
}

// RemovePolicy removes a policy rule from the model.
func (m Model) RemovePolicy(sec string, ptype string, rule []string) bool {
	return m[sec][ptype].Policy.RemovePolicy(rule)
}

// RemovePolicies removes policy rules from the model.
func (m Model) RemovePolicies(sec string, ptype string, rules [][]string) [][]string {
	return m[sec][ptype].Policy.RemovePolicies(rules)
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model.
func (m Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return m[sec][ptype].Policy.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (m Model) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	return m[sec][ptype].Policy.GetValuesForFieldInPolicy(fieldIndex)
}

// GetValuesForFieldInPolicyAllTypes gets all values for a field for all rules in a policy of all ptypes, duplicated values are removed.
func (m Model) GetValuesForFieldInPolicyAllTypes(sec string, fieldIndex int) []string {
	var values []string

	for ptype := range m[sec] {
		values = append(values, m.GetValuesForFieldInPolicy(sec, ptype, fieldIndex)...)
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}

// FilterNotExistsPolicy returns the policy that exist in the model by checking the given rules.
func (m Model) FilterExistsPolicy(sec string, ptype string, rules [][]string) [][]string {
	return m[sec][ptype].Policy.FilterExistsPolicy(rules)
}

// FilterNotExistsPolicy returns the policy that not exist in the model by checking the given rules.
func (m Model) FilterNotExistsPolicy(sec string, ptype string, rules [][]string) [][]string {
	return m[sec][ptype].Policy.FilterNotExistsPolicy(rules)
}
