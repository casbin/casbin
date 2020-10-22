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

package model

import (
	"strings"
	"sync"

	"github.com/casbin/casbin/v3/util"
)

// Policy represents an policy rule
type Policy struct {
	data  [][]string
	index map[string]int
	mutex *sync.RWMutex
}

// NewPolicy creates a policy
func NewPolicy() *Policy {
	return &Policy{
		data:  [][]string{},
		index: map[string]int{},
		mutex: &sync.RWMutex{},
	}
}

// toIndexKey returns key of index.
func toIndexKey(rule []string) string {
	return strings.Join(rule, ",")
}

// addIndex adds the give rule and value to index.
func (p *Policy) addIndex(rule []string, value int) {
	p.index[toIndexKey(rule)] = value
}

// removeIndex removes the give rule from index.
func (p *Policy) removeIndex(rule []string) {
	delete(p.index, toIndexKey(rule))
}

// getIndex returns the index, and whether the rule exists in index.
func (p *Policy) getIndex(rule []string) (int, bool) {
	value, found := p.index[toIndexKey(rule)]
	return value, found
}

// addPolicy adds a rule and returns whether it was successful.
func (p *Policy) addPolicy(rule []string) bool {
	_, found := p.getIndex(rule)
	if found {
		return false
	}

	p.data = append(p.data, rule)
	p.addIndex(rule, len(p.data)-1)

	return true
}

// removePolicy removes a rule and returns whether it was successful.
func (p *Policy) removePolicy(rule []string) bool {
	index, found := p.getIndex(rule)
	if !found {
		return false
	}

	p.removeIndex(rule)
	p.data = append(p.data[:index], p.data[index+1:]...)
	for i := index; i < len(p.data); i++ {
		p.addIndex(p.data[i], i)
	}

	return true
}

// hasIndex checks whether the rule exists.
func (p *Policy) hasIndex(rule []string) bool {
	_, found := p.getIndex(rule)
	return found
}

// GetPolicy gets all rules.
func (p *Policy) GetPolicy() [][]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var res [][]string
	for _, v := range p.data {
		temp := make([]string, len(v))
		copy(temp, v)
		res = append(res, temp)
	}

	return res
}

// GetFilteredPolicy gets rules based on field filters.
func (p *Policy) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var res [][]string

	for _, rule := range p.data {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			temp := make([]string, len(rule))
			copy(temp, rule)
			res = append(res, temp)
		}
	}

	return res
}

// AddPolicies adds the given policy rule and returns whether it was successful.
func (p *Policy) AddPolicy(rule []string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.addPolicy(rule)
}

// AddPolicies adds the given policy rules and returns the effected policy rules.
func (p *Policy) AddPolicies(rules [][]string) [][]string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var effects [][]string

	for _, rule := range rules {
		if p.addPolicy(rule) {
			effects = append(effects, rule)
		}
	}

	return effects
}

// RemovePolicy removes the given policy rule and returns whether it was successful.
func (p *Policy) RemovePolicy(rule []string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removePolicy(rule)
}

// RemovePolicies removes the given policy rules and returns the effected policy rules.
func (p *Policy) RemovePolicies(rules [][]string) [][]string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var effects [][]string

	for _, rule := range rules {
		if p.removePolicy(rule) {
			effects = append(effects, rule)
		}
	}

	return effects
}

// RemoveFilteredPolicy removes policy rules based on field filters from the model and returns the affected policy.
func (p *Policy) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var tmp [][]string
	var effects [][]string
	firstIndex := -1

	if len(fieldValues) == 0 {
		return effects
	}

	for index, rule := range p.data {
		matched := true
		for i, fieldValue := range fieldValues {
			if fieldValue != "" && rule[fieldIndex+i] != fieldValue {
				matched = false
				break
			}
		}

		if matched {
			if firstIndex == -1 {
				firstIndex = index
			}
			p.removeIndex(rule)
			effects = append(effects, rule)
		} else {
			tmp = append(tmp, rule)
		}
	}

	if firstIndex != -1 {
		p.data = tmp
		for i := firstIndex; i < len(p.data); i++ {
			p.addIndex(p.data[i], i)
		}
	}

	return effects
}

// ClearPolicy clears all current policy.
func (p *Policy) ClearPolicy() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.data = make([][]string, 0)
	p.index = make(map[string]int)
}

// HasPolicy determines whether a model has the specified policy rule.
func (p *Policy) HasPolicy(rule []string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.hasIndex(rule)
}

// GetValuesForFieldInPolicy gets all values for a field for all rules in a policy, duplicated values are removed.
func (p *Policy) GetValuesForFieldInPolicy(fieldIndex int) []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var values []string

	for _, rule := range p.data {
		values = append(values, rule[fieldIndex])
	}

	util.ArrayRemoveDuplicates(&values)
	return values
}

// FilterNotExistsPolicy returns the policy that exist in the model by checking the given rules.
func (p *Policy) FilterExistsPolicy(rules [][]string) [][]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var res [][]string

	for _, rule := range rules {
		if p.hasIndex(rule) {
			res = append(res, rule)
		}
	}

	return res
}

// FilterNotExistsPolicy returns the policy that not exist in the model by checking the given rules.
func (p *Policy) FilterNotExistsPolicy(rules [][]string) [][]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var res [][]string

	for _, rule := range rules {
		if !p.hasIndex(rule) {
			res = append(res, rule)
		}
	}

	return res
}
