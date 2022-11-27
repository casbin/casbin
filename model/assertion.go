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

package model

import (
	"errors"
	"strings"

	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/rbac"
)

// Assertion represents an expression in a section of the model.
// For example: r = sub, obj, act
type Assertion struct {
	Key           string
	Value         string
	Tokens        []string
	Policy        [][]string
	PolicyMap     map[string]int
	RM            rbac.RoleManager
	FieldIndexMap map[string]int

	logger log.Logger
}

func (ast *Assertion) buildIncrementalRoleLinks(rm rbac.RoleManager, op PolicyOp, rules [][]string) error {
	ast.RM = rm
	count := strings.Count(ast.Value, "_")
	if count < 2 {
		return errors.New("the number of \"_\" in role definition should be at least 2")
	}

	for _, rule := range rules {
		if len(rule) < count {
			return errors.New("grouping policy elements do not meet role definition")
		}
		if len(rule) > count {
			rule = rule[:count]
		}
		switch op {
		case PolicyAdd:
			err := rm.AddLink(rule[0], rule[1], rule[2:]...)
			if err != nil {
				return err
			}
		case PolicyRemove:
			err := rm.DeleteLink(rule[0], rule[1], rule[2:]...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ast *Assertion) buildRoleLinks(rm rbac.RoleManager) error {
	ast.RM = rm
	count := strings.Count(ast.Value, "_")
	if count < 2 {
		return errors.New("the number of \"_\" in role definition should be at least 2")
	}
	for _, rule := range ast.Policy {
		if len(rule) < count {
			return errors.New("grouping policy elements do not meet role definition")
		}
		if len(rule) > count {
			rule = rule[:count]
		}
		err := ast.RM.AddLink(rule[0], rule[1], rule[2:]...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ast *Assertion) setLogger(logger log.Logger) {
	ast.logger = logger
}

func (ast *Assertion) copy() *Assertion {
	tokens := append([]string(nil), ast.Tokens...)
	policy := make([][]string, len(ast.Policy))

	for i, p := range ast.Policy {
		policy[i] = append(policy[i], p...)
	}
	policyMap := make(map[string]int)
	for k, v := range ast.PolicyMap {
		policyMap[k] = v
	}

	newAst := &Assertion{
		Key:           ast.Key,
		Value:         ast.Value,
		PolicyMap:     policyMap,
		Tokens:        tokens,
		Policy:        policy,
		FieldIndexMap: ast.FieldIndexMap,
	}

	return newAst
}
