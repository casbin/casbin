package casbin

import "log"

type Assertion struct {
	Key    string
	Value  string
	Tokens []string
	Policy [][]string
	RM     *RoleManager
}

func (ast *Assertion) buildRoleLinks() {
	ast.RM = newRoleManager(1)
	for _, rule := range ast.Policy {
		ast.RM.addLink(rule[0], rule[1])
	}

	log.Print("Role links for: " + ast.Key)
	ast.RM.printRoles()
}
