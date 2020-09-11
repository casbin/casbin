package internal

import (
	"github.com/casbin/casbin/v3/errors"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/casbin/casbin/v3/rbac"
)

// PolicyManager is the policy manager for model and adapter
type PolicyManager interface {
	AddPolicy(sec string, ptype string, rule []string, shouldPersist bool) (bool, error)
	AddPolicies(sec string, ptype string, rules [][]string, shouldPersist bool) (bool, error)
	RemovePolicy(sec string, ptype string, rule []string, shouldPersist bool) (bool, error)
	RemovePolicies(sec string, ptype string, rules [][]string, shouldPersist bool) (bool, error)
	RemoveFilteredPolicy(sec string, ptype string, shouldPersist bool, fieldIndex int, fieldValues ...string) (bool, [][]string, error)
}

type policyManager struct {
	model   *model.Model
	adapter persist.Adapter
	rm      rbac.RoleManager
}

const (
	notImplemented = "not implemented"
)

// NewPolicyManager is the constructor for InternalController
func NewPolicyManager(model *model.Model, adapter persist.Adapter, rm rbac.RoleManager) PolicyManager {
	return &policyManager{
		model:   model,
		adapter: adapter,
		rm:      rm,
	}
}

// AddPolicy adds a rule to model and adapter.
func (p *policyManager) AddPolicy(sec string, ptype string, rule []string, shouldPersist bool) (bool, error) {
	if p.model.HasPolicy(sec, ptype, rule) {
		return false, nil
	}

	if shouldPersist {
		if err := p.adapter.AddPolicy(sec, ptype, rule); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	p.model.AddPolicy(sec, ptype, rule)

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyAdd, "g", ptype, [][]string{rule})
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

// AddPolicies adds rules to model and adapter.
func (p *policyManager) AddPolicies(sec string, ptype string, rules [][]string, shouldPersist bool) (bool, error) {
	if p.model.HasPolicies(sec, ptype, rules) {
		return false, nil
	}

	if shouldPersist {
		if err := p.adapter.(persist.BatchAdapter).AddPolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	p.model.AddPolicies(sec, ptype, rules)

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyAdd, "g", ptype, rules)
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

// RemovePolicy removes a rule from model and adapter.
func (p *policyManager) RemovePolicy(sec string, ptype string, rule []string, shouldPersist bool) (bool, error) {
	if shouldPersist {
		if err := p.adapter.RemovePolicy(sec, ptype, rule); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	ruleRemoved := p.model.RemovePolicy(sec, ptype, rule)
	if !ruleRemoved {
		return ruleRemoved, nil
	}

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyRemove, "g", ptype, [][]string{rule})
		if err != nil {
			return ruleRemoved, err
		}
	}

	return ruleRemoved, nil
}

// RemovePolicies removes rules from model and adapter.
func (p *policyManager) RemovePolicies(sec string, ptype string, rules [][]string, shouldPersist bool) (bool, error) {
	if !p.model.HasPolicies(sec, ptype, rules) {
		return false, nil
	}

	if shouldPersist {
		if err := p.adapter.(persist.BatchAdapter).RemovePolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	rulesRemoved := p.model.RemovePolicies(sec, ptype, rules)
	if !rulesRemoved {
		return rulesRemoved, nil
	}

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyRemove, "g", ptype, rules)
		if err != nil {
			return rulesRemoved, err
		}
	}

	return rulesRemoved, nil
}

// RemoveFilteredPolicy removes rules based on field filters from model and adapter.
func (p *policyManager) RemoveFilteredPolicy(sec string, ptype string, shouldPersist bool, fieldIndex int, fieldValues ...string) (bool, [][]string, error) {
	if len(fieldValues) == 0 {
		return false, nil, errors.INVALID_FIELDVAULES_PARAMETER
	}

	if shouldPersist {
		if err := p.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return false, nil, err
			}
		}
	}

	ruleRemoved, effects := p.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	if !ruleRemoved {
		return ruleRemoved, effects, nil
	}

	if sec == "g" {
		err := p.model.BuildIncrementalRoleLinks(p.rm, model.PolicyRemove, "g", ptype, effects)
		if err != nil {
			return ruleRemoved, effects, err
		}
	}

	return ruleRemoved, effects, nil
}
