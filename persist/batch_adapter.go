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

// BatchAdapter is the interface for Casbin adapters with multiple add and remove policy functions.
type BatchAdapter interface {
	Adapter
	// AddPolicies adds policy rules to the storage.
	// This is part of the Auto-Save feature.
	AddPolicies(sec string, ptype string, rules [][]string) error
	// RemovePolicies removes policy rules from the storage.
	// This is part of the Auto-Save feature.
	RemovePolicies(sec string, ptype string, rules [][]string) error
}
