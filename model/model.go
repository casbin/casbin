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

	"github.com/casbin/casbin/v2/config"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/util"
)

// Model represents the whole access control model.
type Model map[string]AssertionMap

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

	if sec == "r" || sec == "p" {
		ast.Tokens = strings.Split(ast.Value, ",")
		for i := range ast.Tokens {
			ast.Tokens[i] = key + "_" + strings.TrimSpace(ast.Tokens[i])
		}
	} else {
		ast.Value = util.RemoveComments(util.EscapeAssertion(ast.Value))
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

// NewModel creates an empty model.
func NewModel() Model {
	m := make(Model)
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
	log.LogPrint("Model:")
	for k, v := range model {
		for i, j := range v {
			log.LogPrintf("%s.%s: %s", k, i, j.Value)
		}
	}
}
