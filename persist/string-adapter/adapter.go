// Copyright 2017 The string-adapter Authors. All Rights Reserved.
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

package string_adapter

import (
	"bytes"
	"errors"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
)

type Adapter struct {
	Line string
}

func NewAdapter(line string) *Adapter {
	return &Adapter{
		Line: line,
	}
}

func (sa *Adapter) LoadPolicy(model model.Model) error {
	if sa.Line == "" {
		return errors.New("invalid line, line cannot be empty")
	}
	strs := strings.Split(sa.Line, "\n")
	for _, str := range strs {
		if str == "" {
			continue
		}
		_ = persist.LoadPolicyLine(str, model)
	}

	return nil
}

func (sa *Adapter) SavePolicy(model model.Model) error {
	var tmp bytes.Buffer
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}
	sa.Line = strings.TrimRight(tmp.String(), "\n")
	return nil
}

func (sa *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

func (sa *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	sa.Line = ""
	return nil
}

func (sa *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}
