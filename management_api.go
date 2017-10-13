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
	return e.GetAllNamedSubjects("p")
}

// GetAllNamedSubjects gets the list of subjects that show up in the current named policy.
func (e *Enforcer) GetAllNamedSubjects(name string) []string {
	return e.model.GetValuesForFieldInPolicy("p", name, 0)
}

// GetAllObjects gets the list of objects that show up in the current policy.
func (e *Enforcer) GetAllObjects() []string {
	return e.GetAllNamedObjects("p")
}

// GetAllNamedObjects gets the list of objects that show up in the current named policy.
func (e *Enforcer) GetAllNamedObjects(name string) []string {
	return e.model.GetValuesForFieldInPolicy("p", name, 1)
}

// GetAllActions gets the list of actions that show up in the current policy.
func (e *Enforcer) GetAllActions() []string {
	return e.GetAllNamedActions("p")
}

// GetAllNamedActions gets the list of actions that show up in the current named policy.
func (e *Enforcer) GetAllNamedActions(name string) []string {
	return e.model.GetValuesForFieldInPolicy("p", name, 2)
}

// GetAllRoles gets the list of roles that show up in the current policy.
func (e *Enforcer) GetAllRoles() []string {
	return e.GetAllNamedRoles("g")
}

// GetAllNamedRoles gets the list of roles that show up in the current named policy.
func (e *Enforcer) GetAllNamedRoles(name string) []string {
	return e.model.GetValuesForFieldInPolicy("g", name, 1)
}

// GetPolicy gets all the authorization rules in the policy.
func (e *Enforcer) GetPolicy() [][]string {
	return e.GetNamedPolicy("p")
}

// GetFilteredPolicy gets all the authorization rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.GetFilteredNamedPolicy("p", fieldIndex, fieldValues...)
}

// GetNamedPolicy gets all the authorization rules in the named policy.
func (e *Enforcer) GetNamedPolicy(name string) [][]string {
	return e.model.GetPolicy("p", name)
}

// GetFilteredNamedPolicy gets all the authorization rules in the named policy, field filters can be specified.
func (e *Enforcer) GetFilteredNamedPolicy(name string, fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("p", name, fieldIndex, fieldValues...)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.GetNamedGroupingPolicy("g")
}

// GetFilteredGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.GetFilteredNamedGroupingPolicy("g", fieldIndex, fieldValues...)
}

// GetNamedGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetNamedGroupingPolicy(name string) [][]string {
	return e.model.GetPolicy("g", name)
}

// GetFilteredNamedGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredNamedGroupingPolicy(name string, fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("g", name, fieldIndex, fieldValues...)
}

// HasPolicy determines whether an authorization rule exists.
func (e *Enforcer) HasPolicy(params ...interface{}) bool {
	return e.HasNamedPolicy("p", params...)
}

// HasNamedPolicy determines whether a named authorization rule exists.
func (e *Enforcer) HasNamedPolicy(name string, params ...interface{}) bool {
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		return e.model.HasPolicy("p", name, params[0].([]string))
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("p", name, policy)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddPolicy(params ...interface{}) bool {
	return e.AddNamedPolicy("p", params...)
}

// AddNamedPolicy adds an authorization rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddNamedPolicy(name string, params ...interface{}) bool {
	ruleAdded := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleAdded = e.addPolicy("p", name, params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleAdded = e.addPolicy("p", name, policy)
	}

	return ruleAdded
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(params ...interface{}) bool {
	return e.RemoveNamedPolicy("p", params...)
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) bool {
	return e.RemoveFilteredNamedPolicy("p", fieldIndex, fieldValues...)
}

// RemoveNamedPolicy removes an authorization rule from the current named policy.
func (e *Enforcer) RemoveNamedPolicy(name string, params ...interface{}) bool {
	ruleRemoved := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleRemoved = e.removePolicy("p", name, params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleRemoved = e.removePolicy("p", name, policy)
	}

	return ruleRemoved
}

// RemoveFilteredNamedPolicy removes an authorization rule from the current named policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredNamedPolicy(name string, fieldIndex int, fieldValues ...string) bool {
	ruleRemoved := e.removeFilteredPolicy("p", name, fieldIndex, fieldValues...)
	return ruleRemoved
}

// HasGroupingPolicy determines whether a role inheritance rule exists.
func (e *Enforcer) HasGroupingPolicy(params ...interface{}) bool {
	return e.HasNamedGroupingPolicy("g", params...)
}

// HasNamedGroupingPolicy determines whether a named role inheritance rule exists.
func (e *Enforcer) HasNamedGroupingPolicy(name string, params ...interface{}) bool {
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		return e.model.HasPolicy("g", name, params[0].([]string))
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("g", name, policy)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddGroupingPolicy(params ...interface{}) bool {
	return e.AddNamedGroupingPolicy("g", params...)
}

// AddNamedGroupingPolicy adds a named role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddNamedGroupingPolicy(name string, params ...interface{}) bool {
	ruleAdded := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleAdded = e.addPolicy("g", name, params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleAdded = e.addPolicy("g", name, policy)
	}

	e.model.BuildRoleLinks(e.rmc)
	return ruleAdded
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(params ...interface{}) bool {
	return e.RemoveNamedGroupingPolicy("g", params...)
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) bool {
	return e.RemoveFilteredNamedGroupingPolicy("g", fieldIndex, fieldValues...)
}

// RemoveNamedGroupingPolicy removes a role inheritance rule from the current named policy.
func (e *Enforcer) RemoveNamedGroupingPolicy(name string, params ...interface{}) bool {
	ruleRemoved := false
	if len(params) == 1 && reflect.TypeOf(params[0]).Kind() == reflect.Slice {
		ruleRemoved = e.removePolicy("g", name, params[0].([]string))
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleRemoved = e.removePolicy("g", name, policy)
	}

	e.model.BuildRoleLinks(e.rmc)
	return ruleRemoved
}

// RemoveFilteredNamedGroupingPolicy removes a role inheritance rule from the current named policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredNamedGroupingPolicy(name string, fieldIndex int, fieldValues ...string) bool {
	ruleRemoved := e.removeFilteredPolicy("g", name, fieldIndex, fieldValues...)
	e.model.BuildRoleLinks(e.rmc)
	return ruleRemoved
}

// AddFunction adds a customized function.
func (e *Enforcer) AddFunction(name string, function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction(name, function)
}
