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

// GetFilteredPolicy gets all the authorization rules in the policy, a field filter can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return e.model.GetFilteredPolicy("p", "p", fieldIndex, fieldValue)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.model.GetPolicy("g", "g")
}

// AddPolicy adds an authorization rule to the current policy.
func (e *Enforcer) AddPolicy(policy []string) {
	e.model.AddPolicy("p", "p", policy)
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(policy []string) {
	e.model.RemovePolicy("p", "p", policy)
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, a field filter can be specified.
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValue string) {
	e.model.RemoveFilteredPolicy("p", "p", fieldIndex, fieldValue)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
func (e *Enforcer) AddGroupingPolicy(policy []string) {
	e.model.AddPolicy("g", "g", policy)
	e.model.BuildRoleLinks()
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(policy []string) {
	e.model.RemovePolicy("g", "g", policy)
	e.model.BuildRoleLinks()
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, a field filter can be specified.
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValue string) {
	e.model.RemoveFilteredPolicy("g", "g", fieldIndex, fieldValue)
	e.model.BuildRoleLinks()
}

// AddSubjectAttributeFunction adds the function that gets attributes for a subject in ABAC.
func (e *Enforcer) AddSubjectAttributeFunction(function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction("subAttr", function)
}

// AddObjectAttributeFunction adds the function that gets attributes for a object in ABAC.
func (e *Enforcer) AddObjectAttributeFunction(function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction("objAttr", function)
}

// AddActionAttributeFunction adds the function that gets attributes for a object in ABAC.
func (e *Enforcer) AddActionAttributeFunction(function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction("actAttr", function)
}

// AddFunction adds a customized function.
func (e *Enforcer) AddFunction(name string, function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction(name, function)
}
