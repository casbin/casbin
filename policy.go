package casbin

import (
	"github.com/hsluoyz/casbin/util"
	"log"
)

// Initialize the roles in RBAC.
func (model Model) BuildRoleLinks() {
	for _, ast := range model["g"] {
		ast.buildRoleLinks()
	}
}

// Print the policy.
func (model Model) PrintPolicy() {
	log.Print("Policy:")
	for key, ast := range model["p"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
	}

	for key, ast := range model["g"] {
		log.Print(key, ": ", ast.Value, ": ", ast.Policy)
	}
}

// Clear all current policy.
func (model Model) ClearPolicy() {
	for _, ast := range model["p"] {
		ast.Policy = nil
	}

	for _, ast := range model["g"] {
		ast.Policy = nil
	}
}

// Get all rules in a policy.
func (model Model) GetPolicy(sec string, ptype string) [][]string {
	return model[sec][ptype].Policy
}

// Get rules based on a field filter from a policy.
func (model Model) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValue string) [][]string {
	res := [][]string{}

	for _, v := range model[sec][ptype].Policy {
		if v[fieldIndex] == fieldValue {
			res = append(res, v)
		}
	}

	return res
}

// Determine whether a model has the specified policy rule.
func (model Model) HasPolicy(sec string, ptype string, policy []string) bool {
	for _, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			return true
		}
	}

	return false
}

// Add a policy rule to the model.
func (model Model) AddPolicy(sec string, ptype string, policy []string) bool {
	if !model.HasPolicy(sec, ptype, policy) {
		model[sec][ptype].Policy = append(model[sec][ptype].Policy, policy)
		return true
	} else {
		return false
	}
}

// Remove a policy rule from the model.
func (model Model) RemovePolicy(sec string, ptype string, policy []string) bool {
	for i, rule := range model[sec][ptype].Policy {
		if util.ArrayEquals(policy, rule) {
			model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			return true
		}
	}

	return false
}

// Remove policy rules based on a field filter from the model.
func (model Model) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValue string) bool {
	res := false
	for i := range model[sec][ptype].Policy {
		if model[sec][ptype].Policy[i][fieldIndex] == fieldValue {
			model[sec][ptype].Policy = append(model[sec][ptype].Policy[:i], model[sec][ptype].Policy[i+1:]...)
			i -= 1
			res = true
		}
	}

	return res
}

// Get all values for a field for all rules in a policy, duplicated values are removed.
func (model Model) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string {
	users := []string{}

	for _, rule := range model[sec][ptype].Policy {
		users = append(users, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&users)
	// sort.Strings(users)

	return users
}
