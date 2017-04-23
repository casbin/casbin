package api

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
