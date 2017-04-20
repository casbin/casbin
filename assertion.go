package casbin

import (
	"log"
	"github.com/hsluoyz/casbin/rbac"
)

type Assertion struct {
	Key    string
	Value  string
	Tokens []string
	Policy [][]string
	RM     *rbac.RoleManager
}

func (ast *Assertion) buildRoleLinks() {
	ast.RM = rbac.NewRoleManager(1)
	for _, rule := range ast.Policy {
		ast.RM.AddLink(rule[0], rule[1])
	}

	log.Print("Role links for: " + ast.Key)
	ast.RM.PrintRoles()
}
