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

package fileadapter

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/casbin/casbin/util"
)

// Adapter is the file adapter for Casbin.
// It can load policy from file or save policy to file.
type Adapter struct {
	filePath string
}

// NewAdapter is the constructor for Adapter.
func NewAdapter(filePath string) *Adapter {
	return &Adapter{filePath: filePath}
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(m model.Model) error {
	if a.filePath == "" {
		return errors.New("invalid file path, file path cannot be empty")
	}

	return a.loadPolicyFile(m, persist.LoadPolicyLine)
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(m model.Model) error {
	if a.filePath == "" {
		return errors.New("invalid file path, file path cannot be empty")
	}

	var tmp bytes.Buffer

	for ptype, ast := range m["p"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	for ptype, ast := range m["g"] {
		for _, rule := range ast.Policy {
			tmp.WriteString(ptype + ", ")
			tmp.WriteString(util.ArrayToString(rule))
			tmp.WriteString("\n")
		}
	}

	return a.savePolicyFile(strings.TrimRight(tmp.String(), "\n"))
}

func (a *Adapter) loadPolicyFile(m model.Model, handler func(string, model.Model)) error {
	f, err := os.Open(a.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		handler(line, m)
	}
	return scanner.Err()
}

func (a *Adapter) savePolicyFile(text string) error {
	f, err := os.Create(a.filePath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	// error intentionally ignored
	w.WriteString(text)
	w.Flush()
	f.Close()
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}
