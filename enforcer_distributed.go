package casbin

import (
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
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
func (d *DistributedEnforcer) AddPolicySelf(shouldPersist func() bool, sec string, ptype string, rules [][]string) (effects [][]string, err error) {
	var noExistsPolicy [][]string
	for _, rule := range rules {
		if !d.model.HasPolicy(sec, ptype, rule) {
			noExistsPolicy = append(noExistsPolicy, rule)
		}
	}

	if shouldPersist() {
		if err := d.adapter.(persist.BatchAdapter).AddPolicies(sec, ptype, noExistsPolicy); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
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
func (d *DistributedEnforcer) RemovePolicySelf(shouldPersist func() bool, sec string, ptype string, rules [][]string) (effects [][]string, err error) {
	if shouldPersist() {
		if err := d.adapter.(persist.BatchAdapter).RemovePolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
	}

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
func (d *DistributedEnforcer) RemoveFilteredPolicySelf(shouldPersist func() bool, sec string, ptype string, fieldIndex int, fieldValues ...string) (effects [][]string, err error) {
	if shouldPersist() {
		if err := d.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
	}

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
func (d *DistributedEnforcer) ClearPolicySelf(shouldPersist func() bool) error {
	if shouldPersist() {
		err := d.adapter.SavePolicy(nil)
		if err != nil {
			return err
		}
	}

	d.model.ClearPolicy()

	return nil
}

// UpdatePolicySelf provides a method for dispatcher to update an authorization rule from the current policy.
func (d *DistributedEnforcer) UpdatePolicySelf(shouldPersist func() bool, sec string, ptype string, oldRule, newRule []string) (effected bool, err error) {
	if shouldPersist() {
		err := d.adapter.(persist.UpdatableAdapter).UpdatePolicy(sec, ptype, oldRule, newRule)
		if err != nil {
			return false, err
		}
	}

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
