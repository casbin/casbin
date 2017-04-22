package api

// Get the roles assigned to a subject.
func (e *Enforcer) GetRoles(name string) []string {
	return e.GetRolesForPolicyType("g", name)
}

// Get the roles assigned to a subject, policy type can be specified.
func (e *Enforcer) GetRolesForPolicyType(ptype string, name string) []string {
	return e.model["g"][ptype].RM.GetRoles(name)
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
	return e.GetPolicyForPolicyType("p")
}

// Get all the authorization rules in the policy, policy type can be specified.
func (e *Enforcer) GetPolicyForPolicyType(ptype string) [][]string {
	return e.model.GetPolicy("p", ptype)
}

// Get all the authorization rules in the policy, a field filter can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValue string) [][]string {
	return e.GetFilteredPolicyForPolicyType("p", fieldIndex, fieldValue)
}

// Get all the authorization rules in the policy, a field filter can be specified, policy type can be specified.
func (e *Enforcer) GetFilteredPolicyForPolicyType(ptype string, fieldIndex int, fieldValue string) [][]string {
	return e.model.GetFilteredPolicy("p", ptype, fieldIndex, fieldValue)
}

// Get all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.GetGroupingPolicyForPolicyType("g")
}

// Get all the role inheritance rules in the policy, policy type can be specified.
func (e *Enforcer) GetGroupingPolicyForPolicyType(ptype string) [][]string {
	return e.model.GetPolicy("g", ptype)
}

// Add an authorization rule to the current policy.
func (e *Enforcer) AddPolicy(policy []string) {
	e.AddPolicyForPolicyType("p", policy)
}

// Remove an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(policy []string) {
	e.RemovePolicyForPolicyType("p", policy)
}

// Add an authorization rule to the current policy, policy type can be specified.
func (e *Enforcer) AddPolicyForPolicyType(ptype string, policy []string) {
	e.model.AddPolicy("p", ptype, policy)
}

// Remove an authorization rule from the current policy, policy type can be specified.
func (e *Enforcer) RemovePolicyForPolicyType(ptype string, policy []string) {
	e.model.RemovePolicy("p", ptype, policy)
}

// Add a role inheritance rule to the current policy.
func (e *Enforcer) AddGroupingPolicy(policy []string) {
	e.AddGroupingPolicyForPolicyType("g", policy)
	e.model.BuildRoleLinks()
}

// Remove a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(policy []string) {
	e.RemoveGroupingPolicyForPolicyType("g", policy)
	e.model.BuildRoleLinks()
}

// Add a role inheritance rule to the current policy, policy type can be specified.
func (e *Enforcer) AddGroupingPolicyForPolicyType(ptype string, policy []string) {
	e.model.AddPolicy("g", ptype, policy)
	e.model.BuildRoleLinks()
}

// Remove a role inheritance rule from the current policy, policy type can be specified.
func (e *Enforcer) RemoveGroupingPolicyForPolicyType(ptype string, policy []string) {
	e.model.RemovePolicy("g", ptype, policy)
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
