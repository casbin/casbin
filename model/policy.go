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
	"github.com/casbin/casbin/util"
	"log"
)

// BuildRoleLinks initializes the roles in RBAC.
func (model Model) BuildRoleLinks() {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

// PrintPolicy prints the policy to log.
func (model Model) PrintPolicy() {
	log.Print("Policy:")
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
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

// GetFilteredPolicy gets rules based on a field filter from a policy.
func (model Model) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValue string) [][]string {
	res := [][]string{}

	for _, v := range model[sec][ptype].Policy {
		if v[fieldIndex] == fieldValue {
			res = append(res, v)
		}
	}

	return res
}

// HasPolicy determines whether a model has the specified policy rule.
func (model Model) HasPolicy(sec string, ptype string, policy []string) bool {
	for _, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			return true
		}
	}

	return false
}

// AddPolicy adds a policy rule to the model.
func (model Model) AddPolicy(sec string, ptype string, policy []string) bool {
	if !model.HasPolicy(sec, ptype, policy) {
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, policy)
		return true
	}
	return false
}

// RemovePolicy removes a policy rule from the model.
func (model Model) RemovePolicy(sec string, ptype string, policy []string) bool {
	for i, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			return true
		}
	}

	return false
}

// RemoveFilteredPolicy removes policy rules based on a field filter from the model.
func (model Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValue string) bool {
	tmp := [][]string{}
	res := false
	for _, rule := range model[sec][ptype].Policy {
		if rule[fieldIndex] != fieldValue {
			tmp = append(tmp, rule)
		} else {
			res = true
		}
	}

	model[sec][ptype].Policy = tmp
	return res
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (model Model) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	users := []string{}

	for _, rule := range model[sec][ptype].Policy {
		users = append(users, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&users)
	// sort.Strings(users)

	return users
}
