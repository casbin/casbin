// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"github.com/hsluoyz/casbin/rbac"
	"log"
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
