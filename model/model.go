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
	"container/list"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2/config"
	"github.com/casbin/casbin/v2/constant"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/util"
)

// Model represents the whole access control model.
type Model map[string]AssertionMap

// AssertionMap is the collection of assertions, can be "r", "p", "g", "e", "m".
type AssertionMap map[string]*Assertion

const defaultDomain string = ""
const defaultSeparator = "::"

var sectionNameMap = map[string]string{
	"r": "request_definition",
	"p": "policy_definition",
	"g": "role_definition",
	"e": "policy_effect",
	"m": "matchers",
}

// Minimal required sections for a model to be valid
var requiredSections = []string{"r", "p", "e", "m"}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) bool {
	value := cfg.String(sectionNameMap[sec] + "::" + key)
	return model.AddDef(sec, key, value)
}

// AddDef adds an assertion to the model.
func (model Model) AddDef(sec string, key string, value string) bool {
	if value == "" {
		return false
	}

	ast := Assertion{}
	ast.Key = key
	ast.Value = value
	ast.PolicyMap = make(map[string]int)
	ast.FieldIndexMap = make(map[string]int)
	ast.setLogger(model.GetLogger())

	if sec == "r" || sec == "p" {
		ast.Tokens = strings.Split(ast.Value, ",")
		for i := range ast.Tokens {
			ast.Tokens[i] = key + "_" + strings.TrimSpace(ast.Tokens[i])
		}
	} else if sec == "g" {
		ast.Tokens = strings.Split(ast.Value, ",")
	} else {
		ast.Value = util.RemoveComments(util.EscapeAssertion(ast.Value))
	}

	if sec == "m" && strings.Contains(ast.Value, "in") {
		ast.Value = strings.Replace(strings.Replace(ast.Value, "[", "(", -1), "]", ")", -1)
	}

	_, ok := model[sec]
	if !ok {
		model[sec] = make(AssertionMap)
	}

	model[sec][key] = &ast
	return true
}

func getKeySuffix(i int) string {
	if i == 1 {
		return ""
	}

	return strconv.Itoa(i)
}

func loadSection(model Model, cfg config.ConfigInterface, sec string) {
	i := 1
	for {
		if !loadAssertion(model, cfg, sec, sec+getKeySuffix(i)) {
			break
		} else {
			i++
		}
	}
}

// SetLogger sets the model's logger.
func (model Model) SetLogger(logger log.Logger) {
	for _, astMap := range model {
		for _, ast := range astMap {
			ast.logger = logger
		}
	}
	model["logger"] = AssertionMap{"logger": &Assertion{logger: logger}}
}

// GetLogger returns the model's logger.
func (model Model) GetLogger() log.Logger {
	return model["logger"]["logger"].logger
}

// NewModel creates an empty model.
func NewModel() Model {
	m := make(Model)
	m.SetLogger(&log.DefaultLogger{})

	return m
}

// NewModelFromFile creates a model from a .CONF file.
func NewModelFromFile(path string) (Model, error) {
	m := NewModel()

	err := m.LoadModel(path)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// NewModelFromString creates a model from a string which contains model text.
func NewModelFromString(text string) (Model, error) {
	m := NewModel()

	err := m.LoadModelFromText(text)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// LoadModel loads the model from model CONF file.
func (model Model) LoadModel(path string) error {
	cfg, err := config.NewConfig(path)
	if err != nil {
		return err
	}

	return model.loadModelFromConfig(cfg)
}

// LoadModelFromText loads the model from the text.
func (model Model) LoadModelFromText(text string) error {
	cfg, err := config.NewConfigFromText(text)
	if err != nil {
		return err
	}

	return model.loadModelFromConfig(cfg)
}

func (model Model) loadModelFromConfig(cfg config.ConfigInterface) error {
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

func (model Model) hasSection(sec string) bool {
	section := model[sec]
	return section != nil
}

// PrintModel prints the model to the log.
func (model Model) PrintModel() {
	if !model.GetLogger().IsEnabled() {
		return
	}

	var modelInfo [][]string
	for k, v := range model {
		if k == "logger" {
			continue
		}

		for i, j := range v {
			modelInfo = append(modelInfo, []string{k, i, j.Value})
		}
	}

	model.GetLogger().LogModel(modelInfo)
}

func (model Model) SortPoliciesBySubjectHierarchy() error {
	if model["e"]["e"].Value != constant.SubjectPriorityEffect {
		return nil
	}
	subIndex := 0
	for ptype, assertion := range model["p"] {
		domainIndex, err := model.GetFieldIndex(ptype, constant.DomainIndex)
		if err != nil {
			domainIndex = -1
		}
		policies := assertion.Policy
		subjectHierarchyMap, err := getSubjectHierarchyMap(model["g"]["g"].Policy)
		if err != nil {
			return err
		}
		sort.SliceStable(policies, func(i, j int) bool {
			domain1, domain2 := defaultDomain, defaultDomain
			if domainIndex != -1 {
				domain1 = policies[i][domainIndex]
				domain2 = policies[j][domainIndex]
			}
			name1, name2 := getNameWithDomain(domain1, policies[i][subIndex]), getNameWithDomain(domain2, policies[j][subIndex])
			p1 := subjectHierarchyMap[name1]
			p2 := subjectHierarchyMap[name2]
			return p1 > p2
		})
		for i, policy := range assertion.Policy {
			assertion.PolicyMap[strings.Join(policy, ",")] = i
		}
	}
	return nil
}

func getSubjectHierarchyMap(policies [][]string) (map[string]int, error) {
	subjectHierarchyMap := make(map[string]int)
	// Tree structure of role
	policyMap := make(map[string][]string)
	for _, policy := range policies {
		if len(policy) < 2 {
			return nil, errors.New("policy g expect 2 more params")
		}
		domain := defaultDomain
		if len(policy) != 2 {
			domain = policy[2]
		}
		child := getNameWithDomain(domain, policy[0])
		parent := getNameWithDomain(domain, policy[1])
		policyMap[parent] = append(policyMap[parent], child)
		if _, ok := subjectHierarchyMap[child]; !ok {
			subjectHierarchyMap[child] = 0
		}
		if _, ok := subjectHierarchyMap[parent]; !ok {
			subjectHierarchyMap[parent] = 0
		}
		subjectHierarchyMap[child] = 1
	}
	// Use queues for levelOrder
	queue := list.New()
	for k, v := range subjectHierarchyMap {
		root := k
		if v != 0 {
			continue
		}
		lv := 0
		queue.PushBack(root)
		for queue.Len() != 0 {
			sz := queue.Len()
			for i := 0; i < sz; i++ {
				node := queue.Front()
				queue.Remove(node)
				nodeValue := node.Value.(string)
				subjectHierarchyMap[nodeValue] = lv
				if _, ok := policyMap[nodeValue]; ok {
					for _, child := range policyMap[nodeValue] {
						queue.PushBack(child)
					}
				}
			}
			lv++
		}
	}
	return subjectHierarchyMap, nil
}

func getNameWithDomain(domain string, name string) string {
	return domain + defaultSeparator + name
}

func (model Model) SortPoliciesByPriority() error {
	for ptype, assertion := range model["p"] {
		priorityIndex, err := model.GetFieldIndex(ptype, constant.PriorityIndex)
		if err != nil {
			continue
		}
		policies := assertion.Policy
		sort.SliceStable(policies, func(i, j int) bool {
			p1, err := strconv.Atoi(policies[i][priorityIndex])
			if err != nil {
				return true
			}
			p2, err := strconv.Atoi(policies[j][priorityIndex])
			if err != nil {
				return true
			}
			return p1 < p2
		})
		for i, policy := range assertion.Policy {
			assertion.PolicyMap[strings.Join(policy, ",")] = i
		}
	}
	return nil
}

func (model Model) ToText() string {
	tokenPatterns := make(map[string]string)

	pPattern, rPattern := regexp.MustCompile("^p_"), regexp.MustCompile("^r_")
	for _, ptype := range []string{"r", "p"} {
		for _, token := range model[ptype][ptype].Tokens {
			tokenPatterns[token] = rPattern.ReplaceAllString(pPattern.ReplaceAllString(token, "p."), "r.")
		}
	}
	if strings.Contains(model["e"]["e"].Value, "p_eft") {
		tokenPatterns["p_eft"] = "p.eft"
	}
	s := strings.Builder{}
	writeString := func(sec string) {
		for ptype := range model[sec] {
			value := model[sec][ptype].Value
			for tokenPattern, newToken := range tokenPatterns {
				value = strings.Replace(value, tokenPattern, newToken, -1)
			}
			s.WriteString(fmt.Sprintf("%s = %s\n", sec, value))
		}
	}
	s.WriteString("[request_definition]\n")
	writeString("r")
	s.WriteString("[policy_definition]\n")
	writeString("p")
	if _, ok := model["g"]; ok {
		s.WriteString("[role_definition]\n")
		for ptype := range model["g"] {
			s.WriteString(fmt.Sprintf("%s = %s\n", ptype, model["g"][ptype].Value))
		}
	}
	s.WriteString("[policy_effect]\n")
	writeString("e")
	s.WriteString("[matchers]\n")
	writeString("m")
	return s.String()
}

func (model Model) Copy() Model {
	newModel := NewModel()

	for sec, m := range model {
		newAstMap := make(AssertionMap)
		for ptype, ast := range m {
			newAstMap[ptype] = ast.copy()
		}
		newModel[sec] = newAstMap
	}

	newModel.SetLogger(model.GetLogger())
	return newModel
}

func (model Model) GetFieldIndex(ptype string, field string) (int, error) {
	assertion := model["p"][ptype]
	if index, ok := assertion.FieldIndexMap[field]; ok {
		return index, nil
	}
	pattern := fmt.Sprintf("%s_"+field, ptype)
	index := -1
	for i, token := range assertion.Tokens {
		if token == pattern {
			index = i
			break
		}
	}
	if index == -1 {
		return index, fmt.Errorf(field + " index is not set, please use enforcer.SetFieldIndex() to set index")
	}
	assertion.FieldIndexMap[field] = index
	return index, nil
}
