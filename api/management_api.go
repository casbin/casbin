package api

// Get the roles assigned to a subject.
func (e *Enforcer) GetRoles(name string) []string {
	return e.model["g"]["g"].RM.GetRoles(name)
}

// Get the list of subjects that show up in the current policy.
func (e *Enforcer) GetAllSubjects() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 0)
}

// Get the list of objects that show up in the current policy.
func (e *Enforcer) GetAllObjects() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 1)
}

// Get the list of actions that show up in the current policy.
func (e *Enforcer) GetAllActions() []string {
	return e.model.GetValuesForFieldInPolicy("p", "p", 2)
}

// Get the list of roles that show up in the current policy.
func (e *Enforcer) GetAllRoles() []string {
	return e.model.GetValuesForFieldInPolicy("g", "g", 1)
}

// Get all the authorization rules in the policy.
func (e *Enforcer) GetPolicy() [][]string {
	return e.model.GetPolicy("p", "p")
}

// Get all the authorization rules in the policy, a field filter can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return e.model.GetFilteredPolicy("p", "p", fieldIndex, fieldValue)
}

// Get all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.model.GetPolicy("g", "g")
}

// Add an authorization rule to the current policy.
func (e *Enforcer) AddPolicy(policy []string) {
	e.model.AddPolicy("p", "p", policy)
}

// Remove an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(policy []string) {
	e.model.RemovePolicy("p", "p", policy)
}

// Remove an authorization rule from the current policy, a field filter can be specified.
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValue string) {
	e.model.RemoveFilteredPolicy("p", "p", fieldIndex, fieldValue)
}

// Add a role inheritance rule to the current policy.
func (e *Enforcer) AddGroupingPolicy(policy []string) {
	e.model.AddPolicy("g", "g", policy)
	e.model.BuildRoleLinks()
}

// Remove a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(policy []string) {
	e.model.RemovePolicy("g", "g", policy)
	e.model.BuildRoleLinks()
}

// Remove a role inheritance rule from the current policy, a field filter can be specified.
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValue string) {
	e.model.RemoveFilteredPolicy("g", "g", fieldIndex, fieldValue)
	e.model.BuildRoleLinks()
}

// Add the function that gets attributes for a subject in ABAC.
func (e *Enforcer) AddSubjectAttributeFunction(function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction("subAttr", function)
}

// Add the function that gets attributes for a object in ABAC.
func (e *Enforcer) AddObjectAttributeFunction(function func(args ...interface{}) (interface{}, error)) {
	e.fm.AddFunction("objAttr", function)
}
