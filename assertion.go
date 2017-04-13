package casbin

import "log"

type Assertion struct {
	key    string
	value  string
	tokens []string
	policy [][]string
	rm     *RoleManager
}

func (ast *Assertion) buildRoleLinks() {
	ast.rm = newRoleManager(1)
	for _, policy_role := range ast.policy {
		ast.rm.addLink(policy_role[0], policy_role[1])
	}

	log.Print("Role links for: " + ast.key)
	ast.rm.printRoles()
}
