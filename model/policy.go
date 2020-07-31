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

	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/rbac"
	"github.com/casbin/casbin/v2/util"
)

type (
	PolicyOp int
)

const (
	PolicyAdd PolicyOp = iota
	PolicyRemove
)

const DefaultSep = ","

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
		ast.PolicyMap = map[string]int{}
	}

	for _, ast := range model["g"] {
		ast.Policy = nil
		ast.PolicyMap = map[string]int{}
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
	_, ok := model[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	return ok
}

// HasPolicies determines whether a model has any of the specified policies. If one is found we return false.
func (model Model) HasPolicies(sec string, ptype string, rules [][]string) bool {
	for i := 0; i < len(rules); i++ {
		if model.HasPolicy(sec, ptype, rules[i]) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (model Model) AddPolicy(sec string, ptype string, rule []string) {
	model[sec][ptype].Policy = append(model[sec][ptype].Policy, rule)
	model[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)] = len(model[sec][ptype].Policy) - 1
}

// AddPolicies adds policy rules to the model.
func (model Model) AddPolicies(sec string, ptype string, rules [][]string) {
	for _, rule := range rules {
		hashKey := strings.Join(rule, DefaultSep)
		_, ok := model[sec][ptype].PolicyMap[hashKey]
		if ok {
			continue
		}
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, rule)
		model[sec][ptype].PolicyMap[hashKey] = len(model[sec][ptype].Policy) - 1
	}
}

// RemovePolicy removes a policy rule from the model.
func (model Model) RemovePolicy(sec string, ptype string, rule []string) bool {
	index, ok := model[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	if !ok {
		return false
	}

	model[sec][ptype].Policy = append(model[sec][ptype].Policy[:index], model[sec][ptype].Policy[index+1:]...)
	delete(model[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
	for i := index; i < len(model[sec][ptype].Policy); i++ {
		model[sec][ptype].PolicyMap[strings.Join(model[sec][ptype].Policy[i], DefaultSep)] = i
	}

	return true
}

// RemovePolicies removes policy rules from the model.
func (model Model) RemovePolicies(sec string, ptype string, rules [][]string) bool {
	for _, rule := range rules {
		index, ok := model[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
		if !ok {
			continue
		}

		model[sec][ptype].Policy = append(model[sec][ptype].Policy[:index], model[sec][ptype].Policy[index+1:]...)
		delete(model[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
		for i := index; i < len(model[sec][ptype].Policy); i++ {
			model[sec][ptype].PolicyMap[strings.Join(model[sec][ptype].Policy[i], DefaultSep)] = i
		}
	}
	return true
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model.
func (model Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string) {
	var tmp [][]string
	var effects [][]string
	res := false
	firstIndex := -1

	if len(fieldValues) == 0 {
		return false, effects
	}

	for index, rule := range model[sec][ptype].Policy {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			if firstIndex == -1 {
				firstIndex = index
			}
			delete(model[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
			effects = append(effects, rule)
			res = true
		} else {
			tmp = append(tmp, rule)
		}
	}

	if firstIndex != -1 {
		model[sec][ptype].Policy = tmp
		for i := firstIndex; i < len(model[sec][ptype].Policy); i++ {
			model[sec][ptype].PolicyMap[strings.Join(model[sec][ptype].Policy[i], DefaultSep)] = i
		}
	}

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
