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

const DefaultSep = ","

// BuildIncrementalRoleLinks provides incremental build the role inheritance relations.
func (model *DefaultModel) BuildIncrementalRoleLinks(rm rbac.RoleManager, op PolicyOp, sec string, ptype string, rules [][]string) error {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	if sec == "g" {
		return model.data[sec][ptype].buildIncrementalRoleLinks(rm, op, rules)
	}
	return nil
}

// BuildRoleLinks initializes the roles in RBAC.
func (model *DefaultModel) BuildRoleLinks(rm rbac.RoleManager) error {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	for _, ast := range model.data["g"] {
		err := ast.buildRoleLinks(rm)
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintPolicy prints the policy to log.
func (model *DefaultModel) PrintPolicy() {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	log.LogPrint("Policy:")
	for key, ast := range model.data["p"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy)
	}

	for key, ast := range model.data["g"] {
		log.LogPrint(key, ": ", ast.Value, ": ", ast.Policy)
	}
}

// ClearPolicy clears all current policy.
func (model *DefaultModel) ClearPolicy() {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	for _, ast := range model.data["p"] {
		ast.Policy = nil
		ast.PolicyMap = map[string]int{}
	}

	for _, ast := range model.data["g"] {
		ast.Policy = nil
		ast.PolicyMap = map[string]int{}
	}
}

// GetPolicy gets all rules in a policy.
func (model *DefaultModel) GetPolicy(sec string, ptype string) [][]string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	var res [][]string
	for _, v := range model.data[sec][ptype].Policy {
		temp := make([]string, len(v))
		copy(temp, v)
		res = append(res, temp)
	}
	return res
}

// GetFilteredPolicy gets rules based on field filters from a policy.
func (model *DefaultModel) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	res := [][]string{}

	for _, rule := range model.data[sec][ptype].Policy {
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
func (model *DefaultModel) HasPolicy(sec string, ptype string, rule []string) bool {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	_, ok := model.data[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	return ok
}

// HasPolicies determines whether a model has any of the specified policies. If one is found we return false.
func (model *DefaultModel) HasPolicies(sec string, ptype string, rules [][]string) bool {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	for i := 0; i < len(rules); i++ {
		if model.HasPolicy(sec, ptype, rules[i]) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (model *DefaultModel) AddPolicy(sec string, ptype string, rule []string) {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	model.data[sec][ptype].Policy = append(model.data[sec][ptype].Policy, rule)
	model.data[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)] = len(model.data[sec][ptype].Policy) - 1
}

// AddPolicies adds policy rules to the model.
func (model *DefaultModel) AddPolicies(sec string, ptype string, rules [][]string) {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	for _, rule := range rules {
		hashKey := strings.Join(rule, DefaultSep)
		_, ok := model.data[sec][ptype].PolicyMap[hashKey]
		if ok {
			continue
		}
		model.data[sec][ptype].Policy = append(model.data[sec][ptype].Policy, rule)
		model.data[sec][ptype].PolicyMap[hashKey] = len(model.data[sec][ptype].Policy) - 1
	}
}

// RemovePolicy removes a policy rule from the model.
func (model *DefaultModel) RemovePolicy(sec string, ptype string, rule []string) bool {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	index, ok := model.data[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
	if !ok {
		return false
	}

	model.data[sec][ptype].Policy = append(model.data[sec][ptype].Policy[:index], model.data[sec][ptype].Policy[index+1:]...)
	delete(model.data[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
	for i := index; i < len(model.data[sec][ptype].Policy); i++ {
		model.data[sec][ptype].PolicyMap[strings.Join(model.data[sec][ptype].Policy[i], DefaultSep)] = i
	}

	return true
}

// RemovePolicies removes policy rules from the model.
func (model *DefaultModel) RemovePolicies(sec string, ptype string, rules [][]string) bool {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	for _, rule := range rules {
		index, ok := model.data[sec][ptype].PolicyMap[strings.Join(rule, DefaultSep)]
		if !ok {
			continue
		}

		model.data[sec][ptype].Policy = append(model.data[sec][ptype].Policy[:index], model.data[sec][ptype].Policy[index+1:]...)
		delete(model.data[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
		for i := index; i < len(model.data[sec][ptype].Policy); i++ {
			model.data[sec][ptype].PolicyMap[strings.Join(model.data[sec][ptype].Policy[i], DefaultSep)] = i
		}
	}
	return true
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model.
func (model *DefaultModel) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string) {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	var tmp [][]string
	var effects [][]string
	res := false
	firstIndex := -1

	if len(fieldValues) == 0 {
		return false, effects
	}

	for index, rule := range model.data[sec][ptype].Policy {
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
			delete(model.data[sec][ptype].PolicyMap, strings.Join(rule, DefaultSep))
			effects = append(effects, rule)
			res = true
		} else {
			tmp = append(tmp, rule)
		}
	}

	if firstIndex != -1 {
		model.data[sec][ptype].Policy = tmp
		for i := firstIndex; i < len(model.data[sec][ptype].Policy); i++ {
			model.data[sec][ptype].PolicyMap[strings.Join(model.data[sec][ptype].Policy[i], DefaultSep)] = i
		}
	}

	return res, effects
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (model *DefaultModel) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	values := []string{}

	for _, rule := range model.data[sec][ptype].Policy {
		values = append(values, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}

// GetValuesForFieldInPolicyAllTypes gets all values for a field for all rules in a policy of all ptypes, duplicated values are removed.
func (model *DefaultModel) GetValuesForFieldInPolicyAllTypes(sec string, fieldIndex int) []string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	values := []string{}

	for ptype := range model.data[sec] {
		values = append(values, model.GetValuesForFieldInPolicy(sec, ptype, fieldIndex)...)
	}

	util.ArrayRemoveDuplicates(&values)

	return values
}
