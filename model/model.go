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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/casbin/casbin/config"
	"github.com/casbin/casbin/log"
	"github.com/casbin/casbin/util"
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

var sectionIdentifiers = func() (identifiers []string) {
	for k := range sectionNameMap {
		identifiers = append(identifiers, k)
	}
	return
}

func loadAssertion(model Model, cfg config.ConfigInterface, sec string, key string) bool {
	value := cfg.String(sectionNameMap[sec] + "::" + key)
	return model.AddDef(sec, key, value)
}

// AddDef adds an assertion to the model.
func (model Model) AddDef(sec string, key string, value string) bool {
	ast := Assertion{}
	ast.Key = key
	ast.Value = value

	if ast.Value == "" {
		return false
	}

	if sec == "r" || sec == "p" {
		ast.Tokens = strings.Split(ast.Value, ", ")
		for i := range ast.Tokens {
			ast.Tokens[i] = key + "_" + ast.Tokens[i]
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

func loadAllSections(model Model, cfg config.ConfigInterface) {
	for _, sec := range sectionIdentifiers() {
		loadSection(model, cfg, sec)
	}
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

// Load will attempt to load a model from either a file or text
func (model Model) Load(pathOrText string) (err error) {
	var cfg config.ConfigInterface
	if strings.ToLower(filepath.Ext(pathOrText)) == ".conf" {
		cfg, err = config.NewConfig(pathOrText)
	} else {
		cfg, err = config.NewConfigFromText(pathOrText)
	}
	if err != nil {
		return
	}
	loadAllSections(model, cfg)
	return
}

// LoadModel loads the model from model CONF file.
func (model Model) LoadModel(path string) {
	cfg, err := config.NewConfig(path)
	if err != nil {
		panic(err)
	}
	loadAllSections(model, cfg)
}

// LoadModelFromText loads the model from the text.
func (model Model) LoadModelFromText(text string) {
	cfg, err := config.NewConfigFromText(text)
	if err != nil {
		panic(err)
	}
	loadAllSections(model, cfg)
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
