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

// GetRolesForUser gets the roles that a user has.
func (e *SyncedEnforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetRolesForUser(name, domain...)
}

// GetUsersForRole gets the users that has a role.
func (e *SyncedEnforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetUsersForRole(name, domain...)
}

// HasRoleForUser determines whether a user has a role.
func (e *SyncedEnforcer) HasRoleForUser(name string, role string, domain ...string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasRoleForUser(name, role, domain...)
}

// AddRoleForUser adds a role for a user.
// Returns false if the user already has the role (aka not affected).
func (e *SyncedEnforcer) AddRoleForUser(user string, role string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddRoleForUser(user, role, domain...)
}

// AddRolesForUser adds roles for a user.
// Returns false if the user already has the roles (aka not affected).
func (e *SyncedEnforcer) AddRolesForUser(user string, roles []string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddRolesForUser(user, roles, domain...)
}

// DeleteRoleForUser deletes a role for a user.
// Returns false if the user does not have the role (aka not affected).
func (e *SyncedEnforcer) DeleteRoleForUser(user string, role string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeleteRoleForUser(user, role, domain...)
}

// DeleteRolesForUser deletes all roles for a user.
// Returns false if the user does not have any roles (aka not affected).
func (e *SyncedEnforcer) DeleteRolesForUser(user string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeleteRolesForUser(user, domain...)
}

// DeleteUser deletes a user.
// Returns false if the user does not exist (aka not affected).
func (e *SyncedEnforcer) DeleteUser(user string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeleteUser(user)
}

// DeleteRole deletes a role.
// Returns false if the role does not exist (aka not affected).
func (e *SyncedEnforcer) DeleteRole(role string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeleteRole(role)
}

// DeletePermission deletes a permission.
// Returns false if the permission does not exist (aka not affected).
func (e *SyncedEnforcer) DeletePermission(permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeletePermission(permission...)
}

// AddPermissionForUser adds a permission for a user or role.
// Returns false if the user or role already has the permission (aka not affected).
func (e *SyncedEnforcer) AddPermissionForUser(user string, permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddPermissionForUser(user, permission...)
}

// DeletePermissionForUser deletes a permission for a user or role.
// Returns false if the user or role does not have the permission (aka not affected).
func (e *SyncedEnforcer) DeletePermissionForUser(user string, permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeletePermissionForUser(user, permission...)
}

// DeletePermissionsForUser deletes permissions for a user or role.
// Returns false if the user or role does not have any permissions (aka not affected).
func (e *SyncedEnforcer) DeletePermissionsForUser(user string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.DeletePermissionsForUser(user)
}

// GetPermissionsForUser gets permissions for a user or role.
func (e *SyncedEnforcer) GetPermissionsForUser(user string, domain ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetPermissionsForUser(user, domain...)
}

// GetNamedPermissionsForUser gets permissions for a user or role by named policy.
func (e *SyncedEnforcer) GetNamedPermissionsForUser(ptype string, user string, domain ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetNamedPermissionsForUser(ptype, user, domain...)
}

// HasPermissionForUser determines whether a user has a permission.
func (e *SyncedEnforcer) HasPermissionForUser(user string, permission ...string) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasPermissionForUser(user, permission...)
}

// GetImplicitRolesForUser gets implicit roles that a user has.
// Compared to GetRolesForUser(), this function retrieves indirect roles besides direct roles.
// For example:
// g, alice, role:admin
// g, role:admin, role:user
//
// GetRolesForUser("alice") can only get: ["role:admin"].
// But GetImplicitRolesForUser("alice") will get: ["role:admin", "role:user"].
func (e *SyncedEnforcer) GetImplicitRolesForUser(name string, domain ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetImplicitRolesForUser(name, domain...)
}

// GetImplicitPermissionsForUser gets implicit permissions for a user or role.
// Compared to GetPermissionsForUser(), this function retrieves permissions for inherited roles.
// For example:
// p, admin, data1, read
// p, alice, data2, read
// g, alice, admin
//
// GetPermissionsForUser("alice") can only get: [["alice", "data2", "read"]].
// But GetImplicitPermissionsForUser("alice") will get: [["admin", "data1", "read"], ["alice", "data2", "read"]].
func (e *SyncedEnforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.GetImplicitPermissionsForUser(user, domain...)
}

// GetNamedImplicitPermissionsForUser gets implicit permissions for a user or role by named policy.
// Compared to GetNamedPermissionsForUser(), this function retrieves permissions for inherited roles.
// For example:
// p, admin, data1, read
// p2, admin, create
// g, alice, admin
//
// GetImplicitPermissionsForUser("alice") can only get: [["admin", "data1", "read"]], whose policy is default policy "p"
// But you can specify the named policy "p2" to get: [["admin", "create"]] by    GetNamedImplicitPermissionsForUser("p2","alice")
func (e *SyncedEnforcer) GetNamedImplicitPermissionsForUser(ptype string, user string, domain ...string) ([][]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetNamedImplicitPermissionsForUser(ptype, user, domain...)
}

// GetImplicitUsersForPermission gets implicit users for a permission.
// For example:
// p, admin, data1, read
// p, bob, data1, read
// g, alice, admin
//
// GetImplicitUsersForPermission("data1", "read") will get: ["alice", "bob"].
// Note: only users will be returned, roles (2nd arg in "g") will be excluded.
func (e *SyncedEnforcer) GetImplicitUsersForPermission(permission ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetImplicitUsersForPermission(permission...)
}
