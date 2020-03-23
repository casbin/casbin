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
	m           map[string]bool
	enableCache bool
	locker      *sync.RWMutex
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
	e.m = make(map[string]bool)
	e.locker = new(sync.RWMutex)
	return e, nil
}

// EnableCache determines whether to enable cache on Enforce(). When enableCache is enabled, cached result (true | false) will be returned for previous decisions.
func (e *CachedEnforcer) EnableCache(enableCache bool) {
	e.enableCache = enableCache
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
