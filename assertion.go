package casbin

import "log"

type assertion struct {
	key    string
	value  string
	tokens []string
	policy [][]string
	rm     *RoleManager
}

func (ast *assertion) buildRoleLinks() {
	ast.rm = newRoleManager(1)
	for _, rule := range ast.policy {
		ast.rm.addLink(rule[0], rule[1])
	}

	log.Print("Role links for: " + ast.key)
	ast.rm.printRoles()
}
