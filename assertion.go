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
	ast.RM = NewRoleManager(1)
	for _, rule := range ast.Policy {
		ast.RM.AddLink(rule[0], rule[1])
	}

	log.Print("Role links for: " + ast.Key)
	ast.RM.PrintRoles()
}
