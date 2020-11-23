package casbin

import (
	"github.com/casbin/casbin/v2/model"
)

// DistributedEnforcer wraps SyncedEnforcer for dispatcher.
type DistributedEnforcer struct {
	*SyncedEnforcer
}

func NewDistributedEnforcer(params ...interface{}) (*DistributedEnforcer, error) {
	e := &DistributedEnforcer{}
	var err error
	e.SyncedEnforcer, err = NewSyncedEnforcer(params...)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// AddPolicySelf provides a method for dispatcher to add authorization rules to the current policy.
// The function returns the rules affected and error.
func (d *DistributedEnforcer) AddPolicySelf(sec string, ptype string, rules [][]string) (effects [][]string, err error) {
	var noExistsPolicy [][]string
	for _, rule := range rules {
		if !d.HasPolicy(sec, ptype, rule) {
			noExistsPolicy = append(noExistsPolicy, rule)
		}
	}

	d.model.AddPolicies(sec, ptype, noExistsPolicy)

	if sec == "g" {
		err := d.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, noExistsPolicy)
		if err != nil {
			return noExistsPolicy, err
		}
	}

	return rules, nil
}

// RemovePolicySelf provides a method for dispatcher to remove policies from current policy.
// The function returns the rules affected and error.
func (d *DistributedEnforcer) RemovePolicySelf(sec string, ptype string, rules [][]string) (effects [][]string, err error) {
	d.model.RemovePolicies(sec, ptype, rules)

	if sec == "g" {
		err := d.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, rules)
		if err != nil {
			return rules, err
		}
	}

	return rules, err
}

// RemoveFilteredPolicySelf provides a method for dispatcher to remove an authorization rule from the current policy, field filters can be specified.
// The function returns the rules affected and error.
func (d *DistributedEnforcer) RemoveFilteredPolicySelf(sec string, ptype string, fieldIndex int, fieldValues ...string) (effects [][]string, err error) {
	_, effects = d.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)

	if sec == "g" {
		err := d.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, effects)
		if err != nil {
			return effects, err
		}
	}

	return effects, nil
}

// ClearPolicySelf provides a method for dispatcher to clear all rules from the current policy.
func (d *DistributedEnforcer) ClearPolicySelf() error {
	d.model.ClearPolicy()

	return nil
}

// UpdatePolicySelf provides a method for dispatcher to update an authorization rule from the current policy.
func (d *DistributedEnforcer) UpdatePolicySelf(sec string, ptype string, oldRule, newRule []string) (effected bool, err error) {
	ruleUpdated := d.model.UpdatePolicy(sec, ptype, oldRule, newRule)
	if !ruleUpdated {
		return ruleUpdated, nil
	}

	if sec == "g" {
		err := d.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, [][]string{oldRule}) // remove the old rule
		if err != nil {
			return ruleUpdated, err
		}
		err = d.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, [][]string{newRule}) // add the new rule
		if err != nil {
			return ruleUpdated, err
		}
	}

	return ruleUpdated, nil
}
