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

// GetUsersForRoleInDomain gets the users that has a role inside a domain. Add by Gordon
func (e *SyncedEnforcer) GetUsersForRoleInDomain(name string, domain string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetUsersForRoleInDomain(name, domain)
}

// GetRolesForUserInDomain gets the roles that a user has inside a domain.
func (e *SyncedEnforcer) GetRolesForUserInDomain(name string, domain string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetRolesForUserInDomain(name, domain)
}

// GetPermissionsForUserInDomain gets permissions for a user or role inside a domain.
func (e *SyncedEnforcer) GetPermissionsForUserInDomain(user string, domain string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetPermissionsForUserInDomain(user, domain)
}

// AddRoleForUserInDomain adds a role for a user inside a domain.
// Returns false if the user already has the role (aka not affected).
func (e *SyncedEnforcer) AddRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddRoleForUserInDomain(user, role, domain)
}

// DeleteRoleForUserInDomain deletes a role for a user inside a domain.
// Returns false if the user does not have the role (aka not affected).
func (e *SyncedEnforcer) DeleteRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeleteRoleForUserInDomain(user, role, domain)
}
