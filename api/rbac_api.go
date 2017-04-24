package api

// Get roles for a user.
func (e *Enforcer) GetRolesForUser(name string) []string {
	return e.model["g"]["g"].RM.GetRoles(name)
}

// Add a role for a user.
func (e *Enforcer) AddRoleForUser(user string, role string) {
	e.AddGroupingPolicy([]string{user, role})
}

// Delete all roles for a user.
func (e *Enforcer) DeleteRolesForUser(user string) {
	e.RemoveFilteredGroupingPolicy(0, user)
}

// Delete a user.
func (e *Enforcer) DeleteUser(user string) {
	e.RemoveFilteredGroupingPolicy(0, user)
}

// Delete a role.
func (e *Enforcer) DeleteRole(role string) {
	e.RemoveFilteredGroupingPolicy(1, role)
	e.RemoveFilteredPolicy(0, role)
}

// Delete a permission.
func (e *Enforcer) DeletePermission(permission string) {
	e.RemoveFilteredPolicy(1, permission)
}

// Add a permission for a user or role.
func (e *Enforcer) AddPermissionForUser(user string, permission string) {
	e.AddPolicy([]string{user, permission})
}

// Delete permissions for a user or role.
func (e *Enforcer) DeletePermissionsForUser(user string) {
	e.RemoveFilteredPolicy(0, user)
}

// Get permissions for a user or role.
func (e *Enforcer) GetPermissionsForUser(user string) []string {
	res := []string{}

	policy := e.GetFilteredPolicy(0, user)
	for _, rule := range policy {
		res = append(res, rule[1])
	}

	return res
}
