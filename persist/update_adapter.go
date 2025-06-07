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

// UpdatableAdapter is the interface for Casbin adapters with add update policy function.
type UpdatableAdapter interface {
	Adapter
	// UpdatePolicy updates a policy rule from storage.
	// This is part of the Auto-Save feature.
	UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error
	// UpdatePolicies updates some policy rules to storage, like db, redis.
	UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error
	// UpdateFilteredPolicies deletes old rules and adds new rules.
	UpdateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error)
}
