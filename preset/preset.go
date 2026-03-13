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

package preset

import (
	"github.com/casbin/casbin/v3/model"
)

const rbacModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// RBAC returns a standard RBAC model.
// This is equivalent to a model.conf with:
//
//	[request_definition]
//	r = sub, obj, act
//
//	[policy_definition]
//	p = sub, obj, act
//
//	[role_definition]
//	g = _, _
//
//	[policy_effect]
//	e = some(where (p.eft == allow))
//
//	[matchers]
//	m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
func RBAC() model.Model {
	// Model creation from this hardcoded valid text should never fail.
	// If it does, it indicates a programming error in the preset definition.
	m, err := model.NewModelFromString(rbacModelText)
	if err != nil {
		panic("preset: failed to create RBAC model: " + err.Error())
	}
	return m
}
