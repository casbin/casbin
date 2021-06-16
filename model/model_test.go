// Copyright 2019 The casbin Authors. All Rights Reserved.
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
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2/config"
)

var (
	basicExample = filepath.Join("..", "examples", "basic_model.conf")
	basicConfig  = &MockConfig{
		data: map[string]string{
			"request_definition::r": "sub, obj, act",
			"policy_definition::p":  "sub, obj, act",
			"policy_effect::e":      "some(where (p.eft == allow))",
			"matchers::m":           "r.sub == p.sub && r.obj == p.obj && r.act == p.act",
		},
	}
)

type MockConfig struct {
	data map[string]string
	config.ConfigInterface
}

func (mc *MockConfig) String(key string) string {
	return mc.data[key]
}

func TestNewModel(t *testing.T) {
	m := NewModel()
	if m == nil {
		t.Error("new model should not be nil")
	}
}

func TestNewModelFromFile(t *testing.T) {
	m, err := NewModelFromFile(basicExample)
	if err != nil {
		t.Errorf("model failed to load from file: %s", err)
	}
	if m == nil {
		t.Error("model should not be nil")
	}
}

func TestNewModelFromString(t *testing.T) {
	modelBytes, _ := ioutil.ReadFile(basicExample)
	modelString := string(modelBytes)
	m, err := NewModelFromString(modelString)
	if err != nil {
		t.Errorf("model faild to load from string: %s", err)
	}
	if m == nil {
		t.Error("model should not be nil")
	}
}

func TestLoadModelFromConfig(t *testing.T) {
	m := NewModel()
	err := m.loadModelFromConfig(basicConfig)
	if err != nil {
		t.Error("basic config should not return an error")
	}
	m = NewModel()
	err = m.loadModelFromConfig(&MockConfig{})
	if err == nil {
		t.Error("empty config should return error")
	} else {
		// check for missing sections in message
		for _, rs := range requiredSections {
			if !strings.Contains(err.Error(), sectionNameMap[rs]) {
				t.Errorf("section name: %s should be in message", sectionNameMap[rs])
			}
		}
	}
}

func TestHasSection(t *testing.T) {
	m := NewModel()
	_ = m.loadModelFromConfig(basicConfig)
	for _, sec := range requiredSections {
		if !m.hasSection(sec) {
			t.Errorf("%s section was expected in model", sec)
		}
	}
	m = NewModel()
	_ = m.loadModelFromConfig(&MockConfig{})
	for _, sec := range requiredSections {
		if m.hasSection(sec) {
			t.Errorf("%s section was not expected in model", sec)
		}
	}
}

func TestModel_AddDef(t *testing.T) {
	m := NewModel()
	s := "r"
	v := "sub, obj, act"
	ok := m.AddDef(s, s, v)
	if !ok {
		t.Errorf("non empty assertion should be added")
	}
	ok = m.AddDef(s, s, "")
	if ok {
		t.Errorf("empty assertion value should not be added")
	}
}

func TestModel_CopyTo(t *testing.T) {
	a := NewModel()
	a["p"] = make(AssertionMap)
	a["p"]["p"] = new(Assertion)
	a["p"]["p"].Policy = [][]string{{"1"}, {"2"}}

	b := NewModel()
	a.CopyTo(&b)

	if fmt.Sprintf("%p", a["p"]) == fmt.Sprintf("%p", b["p"]) {
		t.Fatal(`the memory address of a["p"] and b["p"] should not be equal`)
	}

	if fmt.Sprintf("%p", a["p"]["p"]) == fmt.Sprintf("%p", b["p"]["p"]) {
		t.Fatal(`the memory address of a["p"]["p"] and b["p"]["p"] should not be equal`)
	}

	a["p"]["p"].Policy = nil
	if reflect.DeepEqual(a["p"]["p"], b["p"]["p"]) {
		t.Fatal(`the a["p"]["p"] and b["p"]["p"] should not be equal`)
	}
}
