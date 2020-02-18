package casbin

// GetRolesForUser gets the roles that a user has.
func (e *CachedEnforcer) GetRolesForUser(name string) ([]string, error) {
	return e.api.GetRolesForUser(name)
}

// GetUsersForRole gets the users that has a role.
func (e *CachedEnforcer) GetUsersForRole(name string) ([]string, error) {
	return e.api.GetUsersForRole(name)
}

// HasRoleForUser determines whether a user has a role.
func (e *CachedEnforcer) HasRoleForUser(name string, role string) (bool, error) {
	return e.api.HasRoleForUser(name, role)
}

// AddRoleForUser adds a role for a user.
// Returns false if the user already has the role (aka not affected).
func (e *CachedEnforcer) AddRoleForUser(user string, role string) (bool, error) {
	return e.api.AddRoleForUser(user, role)
}

// DeleteRoleForUser deletes a role for a user.
// Returns false if the user does not have the role (aka not affected).
func (e *CachedEnforcer) DeleteRoleForUser(user string, role string) (bool, error) {
	return e.api.DeleteRoleForUser(user, role)
}

// DeleteRolesForUser deletes all roles for a user.
// Returns false if the user does not have any roles (aka not affected).
func (e *CachedEnforcer) DeleteRolesForUser(user string) (bool, error) {
	return e.api.DeleteRolesForUser(user)
}

// DeleteUser deletes a user.
// Returns false if the user does not exist (aka not affected).
func (e *CachedEnforcer) DeleteUser(user string) (bool, error) {
	return e.api.DeleteUser(user)
}

// DeleteRole deletes a role.
// Returns false if the role does not exist (aka not affected).
func (e *CachedEnforcer) DeleteRole(role string) (bool, error) {
	return e.api.DeleteRole(role)
}

// DeletePermission deletes a permission.
// Returns false if the permission does not exist (aka not affected).
func (e *CachedEnforcer) DeletePermission(permission ...string) (bool, error) {
	return e.api.DeletePermission(permission...)
}

// AddPermissionForUser adds a permission for a user or role.
// Returns false if the user or role already has the permission (aka not affected).
func (e *CachedEnforcer) AddPermissionForUser(user string, permission ...string) (bool, error) {
	return e.api.AddPermissionForUser(user, permission...)
}

// DeletePermissionForUser deletes a permission for a user or role.
// Returns false if the user or role does not have the permission (aka not affected).
func (e *CachedEnforcer) DeletePermissionForUser(user string, permission ...string) (bool, error) {
	return e.api.DeletePermissionForUser(user, permission...)
}

// DeletePermissionsForUser deletes permissions for a user or role.
// Returns false if the user or role does not have any permissions (aka not affected).
func (e *CachedEnforcer) DeletePermissionsForUser(user string) (bool, error) {
	return e.api.DeletePermissionsForUser(user)
}

// GetPermissionsForUser gets permissions for a user or role.
func (e *CachedEnforcer) GetPermissionsForUser(user string) [][]string {
	return e.api.GetPermissionsForUser(user)
}

// HasPermissionForUser determines whether a user has a permission.
func (e *CachedEnforcer) HasPermissionForUser(user string, permission ...string) bool {
	return e.api.HasPermissionForUser(user, permission...)
}
