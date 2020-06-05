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
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/rbac"
	"github.com/casbin/casbin/v2/util"
)

type PolicyOp int

const (
	PolicyAdd PolicyOp = iota
	PolicyRemove
)

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (model Model) BuildIncrementalRoleLinks(rm rbac.RoleManager, op PolicyOp, sec string, ptype string, rules [][]string) error {
	if sec == "g" {
		return model[sec][ptype].buildIncrementalRoleLinks(rm, op, rules)
	}
	return nil
}

// BuildRoleLinks initializes the roles in RBAC.
func (model Model) BuildRoleLinks(rm rbac.RoleManager) error {
	for _, ast := range model["g"] {
		err := ast.buildRoleLinks(rm)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintPolicy prints the policy to log.
func (model Model) PrintPolicy() {
	log.LogPrint("Policy:")
	for key, ast := range model["p"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy)
	}

	for key, ast := range model["g"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy)
	}
}

// ClearPolicy clears all current policy.
func (model Model) ClearPolicy() {
	for _, ast := range model["p"] {
		ast.Policy = nil
	}

	for _, ast := range model["g"] {
		ast.Policy = nil
	}
}

// GetPolicy gets all rules in a policy.
func (model Model) GetPolicy(sec string, ptype string) [][]string {
	return model[sec][ptype].Policy
}

// GetFilteredPolicy gets rules based on field filters from a policy.
func (model Model) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string {
	res := [][]string{}

	for _, rule := range model[sec][ptype].Policy {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			res = append(res, rule)
		}
	}

	return res
}

// HasPolicy determines whether a model has the specified policy rule.
func (model Model) HasPolicy(sec string, ptype string, rule []string) bool {
	for _, r := range model[sec][ptype].Policy {
		if util.ArrayEquals(rule, r) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (model Model) AddPolicy(sec string, ptype string, rule []string) bool {
	if !model.HasPolicy(sec, ptype, rule) {
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, rule)
		return true
	}
	return false
}

// AddPolicies adds policy rules to the model.
func (model Model) AddPolicies(sec string, ptype string, rules [][]string) bool {
	for i := 0; i < len(rules); i++ {
		if model.HasPolicy(sec, ptype, rules[i]) {
			return false
		}
	}

	for i := 0; i < len(rules); i++ {
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, rules[i])
	}

	return true
}

// RemovePolicy removes a policy rule from the model.
func (model Model) RemovePolicy(sec string, ptype string, rule []string) bool {
	for i, r := range model[sec][ptype].Policy {
		if util.ArrayEquals(rule, r) {
			model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			return true
		}
	}

	return false
}

// RemovePolicies removes policy rules from the model.
func (model Model) RemovePolicies(sec string, ptype string, rules [][]string) bool {
OUTER:
	for j := 0; j < len(rules); j++ {
		for _, r := range model[sec][ptype].Policy {
			if util.ArrayEquals(rules[j], r) {
				continue OUTER
			}
		}
		return false
	}

	for j := 0; j < len(rules); j++ {
		for i, r := range model[sec][ptype].Policy {
			if util.ArrayEquals(rules[j], r) {
				model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			}
		}
	}
	return true
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model.
func (model Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string) {
	var tmp [][]string
	var effects [][]string
	res := false
	for _, rule := range model[sec][ptype].Policy {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			effects = append(effects, rule)
			res = true
		} else {
			tmp = append(tmp, rule)
		}
	}

	model[sec][ptype].Policy = tmp
	return res, effects
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (model Model) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	values := []string{}

	for _, rule := range model[sec][ptype].Policy {
		values = append(values, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}

// GetValuesForFieldInPolicyAllTypes gets all values for a field for all rules in a policy of all ptypes, duplicated values are removed.
func (model Model) GetValuesForFieldInPolicyAllTypes(sec string, fieldIndex int) []string {
	values := []string{}

	for ptype := range model[sec] {
		values = append(values, model.GetValuesForFieldInPolicy(sec, ptype, fieldIndex)...)
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}
