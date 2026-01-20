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

package rbac

import (
	"github.com/casbin/casbin/v3"
)

// AssignRole assigns a role to a user.
// This is a convenience wrapper around AddRoleForUser.
// Returns false if the user already has the role (aka not affected).
func AssignRole(e *casbin.Enforcer, user string, role string) (bool, error) {
	return e.AddRoleForUser(user, role)
}

// Grant grants a permission to a subject (user or role).
// This is a convenience wrapper around AddPolicy.
// Returns false if the permission already exists (aka not affected).
func Grant(e *casbin.Enforcer, subject string, object string, action string) (bool, error) {
	return e.AddPolicy(subject, object, action)
}
