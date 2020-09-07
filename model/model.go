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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
	"github.com/casbin/casbin/v3/rbac"

	"github.com/casbin/casbin/v3/config"
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/util"
)

// Model is an interface that manager the whole access control model.
type Model interface {
	AddDef(sec string, key string, value string) bool
	AddPolicy(sec string, ptype string, rule []string)
	AddPolicies(sec string, ptype string, rules [][]string)
	BuildRoleLinks(rm rbac.RoleManager) error
	BuildIncrementalRoleLinks(rm rbac.RoleManager, op PolicyOp, sec string, ptype string, rules [][]string) error
	ClearPolicy()
	GenerateFunctions(fm FunctionMap) map[string]govaluate.ExpressionFunction
	GetPtypes(sec string) []string
	GetTokens(sec string, ptype string) map[string]int
	GetMatcher() string
	GetPolicy(sec string, ptype string) [][]string
	GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) [][]string
	GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) []string
	GetValuesForFieldInPolicyAllTypes(sec string, fieldIndex int) []string
	GetEffectExpression() string
	GetRoleManager(sec string, ptype string) rbac.RoleManager
	HasPolicy(sec string, ptype string, rule []string) bool
	HasPolicies(sec string, ptype string, rules [][]string) bool
	PrintPolicy()
	PrintModel()
	RemovePolicy(sec string, ptype string, rule []string) bool
	RemovePolicies(sec string, ptype string, rules [][]string) bool
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string)
}

// DefaultModel provides a default implementation.
type DefaultModel struct {
	data  map[string]AssertionMap
	mutex sync.RWMutex
}

var _ Model = (*DefaultModel)(nil)

// AssertionMap is the collection of assertions, can be "r", "p", "g", "e", "m".
type AssertionMap map[string]*Assertion

var sectionNameMap = map[string]string{
	"r": "request_definition",
	"p": "policy_definition",
	"g": "role_definition",
	"e": "policy_effect",
	"m": "matchers",
}

// Minimal required sections for a model to be valid
var requiredSections = []string{"r", "p", "e", "m"}

func loadAssertion(model *DefaultModel, cfg config.ConfigInterface, sec string, key string) bool {
	value := cfg.String(sectionNameMap[sec] + "::" + key)
	return model.addDef(sec, key, value)
}

// AddDef adds an assertion to the model.
func (model *DefaultModel) AddDef(sec string, key string, value string) bool {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	return model.addDef(sec, key, value)
}

func (model *DefaultModel) addDef(sec string, key string, value string) bool {
	if value == "" {
		return false
	}

	ast := Assertion{}
	ast.Key = key
	ast.Value = value
	ast.PolicyMap = make(map[string]int)

	if sec == "r" || sec == "p" {
		ast.Tokens = strings.Split(ast.Value, ",")
		for i := range ast.Tokens {
			ast.Tokens[i] = key + "_" + strings.TrimSpace(ast.Tokens[i])
		}
	} else {
		ast.Value = util.RemoveComments(util.EscapeAssertion(ast.Value))
	}

	_, ok := model.data[sec]
	if !ok {
		model.data[sec] = make(AssertionMap)
	}

	model.data[sec][key] = &ast
	return true
}

func getKeySuffix(i int) string {
	if i == 1 {
		return ""
	}

	return strconv.Itoa(i)
}

func loadSection(model *DefaultModel, cfg config.ConfigInterface, sec string) {
	i := 1
	for {
		if !loadAssertion(model, cfg, sec, sec+getKeySuffix(i)) {
			break
		} else {
			i++
		}
	}
}
// newDefaultModel creates an empty model.
func newDefaultModel() *DefaultModel {
	m := new(DefaultModel)
	m.data = make(map[string]AssertionMap)
	return m
}

// NewModel creates an empty model.
func NewModel() Model {
	return newDefaultModel()
}

// NewModelFromFile creates a model from a .CONF file.
func NewModelFromFile(path string) (Model, error) {
	m := newDefaultModel()
	cfg, err := config.NewConfig(path)
	if err != nil {
		return nil, err
	}
	err = m.loadModelFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return m, err
}

// NewModelFromString creates a model from a string which contains model text.
func NewModelFromString(text string) (Model, error) {
	m := newDefaultModel()
	err := m.LoadModelFromText(text)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// LoadModel loads the model from model CONF file.
func (model *DefaultModel) LoadModel(path string) error {
	cfg, err := config.NewConfig(path)
	if err != nil {
		return err
	}

	return model.loadModelFromConfig(cfg)
}

// LoadModelFromText loads the model from the text.
func (model *DefaultModel) LoadModelFromText(text string) error {
	cfg, err := config.NewConfigFromText(text)
	if err != nil {
		return err
	}

	return model.loadModelFromConfig(cfg)
}

func (model *DefaultModel) loadModelFromConfig(cfg config.ConfigInterface) error {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	for s := range sectionNameMap {
		loadSection(model, cfg, s)
	}
	ms := make([]string, 0)
	for _, rs := range requiredSections {
		if !model.hasSection(rs) {
			ms = append(ms, sectionNameMap[rs])
		}
	}
	if len(ms) > 0 {
		return fmt.Errorf("missing required sections: %s", strings.Join(ms, ","))
	}
	return nil
}

func (model *DefaultModel) hasSection(sec string) bool {
	section := model.data[sec]
	return section != nil
}

// PrintModel prints the model to the log.
func (model *DefaultModel) PrintModel() {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	log.LogPrint("DefaultModel:")
	for k, v := range model.data {
		for i, j := range v {
			log.LogPrintf("%s.%s: %s", k, i, j.Value)
		}
	}
}

// GetMatcher gets the matcher.
func (model *DefaultModel) GetMatcher() string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	return model.data["m"]["m"].Value
}

// GetEffectExpression gets the effect expression.
func (model *DefaultModel) GetEffectExpression() string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	return model.data["e"]["e"].Value
}

// GetRoleManager gets the current role manager used in ptype.
func (model *DefaultModel) GetRoleManager(sec string, ptype string) rbac.RoleManager {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	return model.data[sec][ptype].RM
}

// GetTokens returns a map with all the tokens
func (model *DefaultModel) GetTokens(sec string, ptype string) map[string]int {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	tokens := make(map[string]int, len(model.data[sec][ptype].Tokens))
	for i, token := range model.data[sec][ptype].Tokens {
		tokens[token] = i
	}

	return tokens
}

// GetPtypes returns a slice for all ptype
func (model *DefaultModel) GetPtypes(sec string) []string {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	var res []string
	for k := range model.data[sec] {
		res = append(res, k)
	}
	return res
}

// GenerateFunctions return a map with all the functions
func (model *DefaultModel) GenerateFunctions(fm FunctionMap) map[string]govaluate.ExpressionFunction {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	functions := fm.GetFunctions()

	if _, ok := model.data["g"]; ok {
		for key, ast := range model.data["g"] {
			rm := ast.RM
			functions[key] = util.GenerateGFunction(rm)
		}
	}
	return functions
}
