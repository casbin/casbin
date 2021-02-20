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

// WatcherUpdatable is the strengthen for Casbin watchers.
type WatcherUpdatable interface {
	Watcher
	// UPdateForUpdaticy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.UpdatePolicy()
	UpdateForUpdatePolicy(oldRule, newRule []string) error
	// UPdateForUpdaticies calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.UpdatePolicies()
	UpdateForUpdatePolicies(oldRules, newRules [][]string) error
}
