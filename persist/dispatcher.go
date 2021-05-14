// Copyright 2020 The casbin Authors. All Rights Reserved.
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

// Dispatcher is the interface for Casbin dispatcher
type Dispatcher interface {
	// AddPolicies adds policies rule to all instance.
	AddPolicies(sec string, ptype string, rules [][]string) error
	// RemovePolicies removes policies rule from all instance.
	RemovePolicies(sec string, ptype string, rules [][]string) error
	// RemoveFilteredPolicy removes policy rules that match the filter from all instance.
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error
	// ClearPolicy clears all current policy in all instances
	ClearPolicy() error
	// UpdatePolicy updates policy rule from all instance.
	UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error
	// UpdatePolicies updates some policy rules from all instance
	UpdatePolicies(sec string, ptype string, oldrules, newRules [][]string) error
	// UpdateFilteredPolicies deletes old rules and adds new rules.
	UpdateFilteredPolicies(sec string, ptype string, oldRules [][]string, newRules [][]string) error
}
