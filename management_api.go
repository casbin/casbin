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

package casbin

import (
	"reflect"
)

// GetAllSubjects gets the list of subjects that show up in the current policy.
func (e *Enforcer) GetAllSubjects() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 0)
}

// GetAllObjects gets the list of objects that show up in the current policy.
func (e *Enforcer) GetAllObjects() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 1)
}

// GetAllActions gets the list of actions that show up in the current policy.
func (e *Enforcer) GetAllActions() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 2)
}

// GetAllRoles gets the list of roles that show up in the current policy.
func (e *Enforcer) GetAllRoles() []string {
	return e.model.GetValuesForFieldInPolicy("g", "g", 1)
}

// GetPolicy gets all the authorization rules in the policy.
func (e *Enforcer) GetPolicy() [][]string {
	return e.model.GetPolicy("p", "p")
}

// GetFilteredPolicy gets all the authorization rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("p", "p", fieldIndex, fieldValues...)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.model.GetPolicy("g", "g")
}

// GetFilteredGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("g", "g", fieldIndex, fieldValues...)
}

// HasPolicy determines whether an authorization rule exists.
func (e *Enforcer) HasPolicy(params ...interface{}) bool {
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		return e.model.HasPolicy("p", "p", params[0].([]string))
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("p", "p", policy)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddPolicy(params ...interface{}) bool {
	ruleAdded := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleAdded = e.addPolicy("p", "p", params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleAdded = e.addPolicy("p", "p", policy)
	}

	return ruleAdded
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(params ...interface{}) bool {
	ruleRemoved := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleRemoved = e.removePolicy("p", "p", params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleRemoved = e.removePolicy("p", "p", policy)
	}

	return ruleRemoved
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) bool {
	ruleRemoved := e.removeFilteredPolicy("p", "p", fieldIndex, fieldValues...)
	return ruleRemoved
}

// HasGroupingPolicy determines whether a role inheritance rule exists.
func (e *Enforcer) HasGroupingPolicy(params ...interface{}) bool {
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		return e.model.HasPolicy("g", "g", params[0].([]string))
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("g", "g", policy)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddGroupingPolicy(params ...interface{}) bool {
	ruleAdded := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleAdded = e.addPolicy("g", "g", params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleAdded = e.addPolicy("g", "g", policy)
	}

	e.model.BuildRoleLinks(e.rmc)
	return ruleAdded
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(params ...interface{}) bool {
	ruleRemoved := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleRemoved = e.removePolicy("g", "g", params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleRemoved = e.removePolicy("g", "g", policy)
	}

	e.model.BuildRoleLinks(e.rmc)
	return ruleRemoved
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) bool {
	ruleRemoved := e.removeFilteredPolicy("g", "g", fieldIndex, fieldValues...)
	e.model.BuildRoleLinks(e.rmc)
	return ruleRemoved
}

// AddFunction adds a customized function.
func (e *Enforcer) AddFunction(name string, function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction(name, function)
}
