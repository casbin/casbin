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

package internal

import (
	"github.com/casbin/casbin/v3/api"
	"github.com/casbin/casbin/v3/errors"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/casbin/casbin/v3/rbac"
)

var _ api.PolicyManager = &policyManager{}

type policyManager struct {
	model         model.Model
	adapter       persist.Adapter
	rm            rbac.RoleManager
	shouldPersist func() bool
}

const (
	notImplemented = "not implemented"
)

// NewPolicyManager is the constructor for PolicyManager
func NewPolicyManager(model model.Model, adapter persist.Adapter, rm rbac.RoleManager, shouldPersist func() bool) api.PolicyManager {
	return &policyManager{
		model:         model,
		adapter:       adapter,
		rm:            rm,
		shouldPersist: shouldPersist,
	}
}

// AddPolicies adds rules to model and adapter.
func (p *policyManager) AddPolicies(sec string, ptype string, rules [][]string) ([][]string, error) {
	rules = p.model.FilterNotExistPolicy(sec, ptype, rules)
	if len(rules) == 0 {
		return nil, nil
	}

	if p.shouldPersist() {
		if err := p.adapter.(persist.BatchAdapter).AddPolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
	}

	p.model.AddPolicies(sec, ptype, rules)

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyAdd, "g", ptype, rules)
		if err != nil {
			return rules, err
		}
	}

	return rules, nil
}

// RemovePolicies removes rules from model and adapter.
func (p *policyManager) RemovePolicies(sec string, ptype string, rules [][]string) ([][]string, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	if p.shouldPersist() {
		if err := p.adapter.(persist.BatchAdapter).RemovePolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
	}

	rulesRemoved := p.model.RemovePolicies(sec, ptype, rules)
	if !rulesRemoved {
		return rules, nil
	}

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyRemove, "g", ptype, rules)
		if err != nil {
			return rules, err
		}
	}

	return rules, nil
}

// RemoveFilteredPolicy removes rules based on field filters from model and adapter.
func (p *policyManager) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	if len(fieldValues) == 0 {
		return nil, errors.INVALID_FIELDVAULES_PARAMETER
	}

	if p.shouldPersist() {
		if err := p.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
	}

	ruleRemoved, effects := p.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	if !ruleRemoved {
		return effects, nil
	}

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyRemove, "g", ptype, effects)
		if err != nil {
			return effects, err
		}
	}

	return effects, nil
}
