// Copyright 2025 The casbin Authors. All Rights Reserved.
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

package detector

import "github.com/casbin/casbin/v3/rbac"

// Detector defines the interface of a policy consistency checker, currently used to detect RBAC inheritance cycles.
type Detector interface {
	// Check checks whether the current status of the passed-in RoleManager contains logical errors (e.g., cycles in role inheritance).
	// param: rm RoleManager instance
	// return: If an error is found, return a descriptive error; otherwise return nil.
	Check(rm rbac.RoleManager) error
}
