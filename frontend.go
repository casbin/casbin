// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"bytes"
	"encoding/json"
)

 // CasbinJsGetPermissionForUser 生成 casbin.js 前端所需的数据载荷。
 // 返回内容包含：
 // - m: 模型的文本表示（model.ToText()）
 // - p: 所有策略规则，格式为 [ptype, ...params]
 // - g: 所有分组/角色继承规则，格式为 [ptype, ...params]
 // 说明：参数 user 仅为兼容保留，不在服务端用于过滤，前端可结合 casbin.js 进行用户态判定。
func CasbinJsGetPermissionForUser(e IEnforcer, user string) (string, error) {
	model := e.GetModel()
	m := map[string]interface{}{}

	m["m"] = model.ToText()

	pRules := [][]string{}
	for ptype := range model["p"] {
		policies, err := model.GetPolicy("p", ptype)
		if err != nil {
			return "", err
		}
		for _, rules := range policies {
			pRules = append(pRules, append([]string{ptype}, rules...))
		}
	}
	m["p"] = pRules

	gRules := [][]string{}
	for ptype := range model["g"] {
		policies, err := model.GetPolicy("g", ptype)
		if err != nil {
			return "", err
		}
		for _, rules := range policies {
			gRules = append(gRules, append([]string{ptype}, rules...))
		}
	}
	m["g"] = gRules

	result := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(result)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(m)
	return result.String(), err
}
