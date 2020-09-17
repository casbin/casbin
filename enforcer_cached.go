// Copyright 2018 The casbin Authors. All Rights Reserved.
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

package casbin

import (
	"strings"
	"sync"
)

// CachedEnforcer wraps Enforcer and provides decision cache
type CachedEnforcer struct {
	*Enforcer
	m              map[string]bool
	enableCache    bool
	autoClearCache bool
	locker         *sync.RWMutex
}

// NewCachedEnforcer creates a cached enforcer via file or DB.
func NewCachedEnforcer(params ...interface{}) (*CachedEnforcer, error) {
	e := &CachedEnforcer{}
	var err error
	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	e.enableCache = true
	e.autoClearCache = true
	e.m = make(map[string]bool)
	e.locker = new(sync.RWMutex)
	return e, nil
}

// EnableCache determines whether to enable cache on Enforce(). When enableCache is enabled, cached result (true | false) will be returned for previous decisions.
func (e *CachedEnforcer) EnableCache(enableCache bool) {
	e.enableCache = enableCache
}

// AutoClear determines whether to clear cache after any operation for policy. if AutoClear is true, cache will be clear after update policy.
func (e *CachedEnforcer) AutoClearCache(autoClearCache bool) {
	e.autoClearCache = autoClearCache
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
// if rvals is not string , ingore the cache
func (e *CachedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	if !e.enableCache {
		return e.Enforcer.Enforce(rvals...)
	}

	var key strings.Builder
	for _, rval := range rvals {
		if val, ok := rval.(string); ok {
			key.WriteString(val)
			key.WriteString("$$")
		} else {
			return e.Enforcer.Enforce(rvals...)
		}
	}

	if res, ok := e.getCachedResult(key.String()); ok {
		return res, nil
	}
	res, err := e.Enforcer.Enforce(rvals...)
	if err != nil {
		return false, err
	}

	e.setCachedResult(key.String(), res)
	return res, nil
}

func (e *CachedEnforcer) getCachedResult(key string) (res bool, ok bool) {
	e.locker.RLock()
	defer e.locker.RUnlock()
	res, ok = e.m[key]
	return
}

func (e *CachedEnforcer) setCachedResult(key string, res bool) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.m[key] = res
}

// InvalidateCache deletes all the existing cached decisions.
func (e *CachedEnforcer) InvalidateCache() {
	e.m = make(map[string]bool)
}

/* Clear cache after update Policy*/
// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *CachedEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	ok, err := e.AddNamedPolicy("p", params...)
	if e.autoClearCache {
		e.InvalidateCache()
	}
	return ok, err
}

// AddPolicies adds authorization rules to the current policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding rule by adding the new rule.
func (e *CachedEnforcer) AddPolicies(rules [][]string) (bool, error) {
	ok, err := e.AddNamedPolicies("p", rules)
	if e.autoClearCache {
		e.InvalidateCache()
	}
	return ok, err
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *CachedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	ok, err := e.RemoveNamedPolicy("p", params...)
	e.InvalidateCache()
	return ok, err
}

// RemovePolicies removes authorization rules from the current policy.
func (e *CachedEnforcer) RemovePolicies(rules [][]string) (bool, error) {
	ok, err := e.RemoveNamedPolicies("p", rules)
	if e.autoClearCache {
		e.InvalidateCache()
	}
	return ok, err
}
