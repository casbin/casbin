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

import "github.com/casbin/casbin/v2/model"

// WatcherEx is the strengthened Casbin watchers.
type WatcherEx interface {
	Watcher
	// UpdateForAddPolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.AddPolicy()
	UpdateForAddPolicy(sec, ptype string, params ...string) error
	// UpdateForRemovePolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemovePolicy()
	UpdateForRemovePolicy(sec, ptype string, params ...string) error
	// UpdateForRemoveFilteredPolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
	UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error
	// UpdateForSavePolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
	UpdateForSavePolicy(model model.Model) error
	// UpdateForAddPolicies calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.AddPolicies()
	UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error
	// UpdateForRemovePolicies calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemovePolicies()
	UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error
}
