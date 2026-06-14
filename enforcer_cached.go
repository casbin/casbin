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
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v3/persist"
	"github.com/casbin/casbin/v3/persist/cache"
)

// CachedEnforcer wraps Enforcer and provides decision cache.
type CachedEnforcer struct {
	*Enforcer
	expireTime  time.Duration
	cache       cache.Cache
	enableCache int32
	locker      *sync.RWMutex
}

type CacheableParam interface {
	GetCacheKey() string
}

// NewCachedEnforcer creates a cached enforcer via file or DB.
func NewCachedEnforcer(params ...interface{}) (*CachedEnforcer, error) {
	e := &CachedEnforcer{}
	var err error
	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	e.enableCache = 1
	e.cache, _ = cache.NewDefaultCache()
	e.locker = new(sync.RWMutex)
	return e, nil
}

// EnableCache determines whether to enable cache on Enforce(). When enableCache is enabled, cached result (true | false) will be returned for previous decisions.
func (e *CachedEnforcer) EnableCache(enableCache bool) {
	var enabled int32
	if enableCache {
		enabled = 1
	}
	atomic.StoreInt32(&e.enableCache, enabled)
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
// if rvals is not string , ignore the cache.
func (e *CachedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	if atomic.LoadInt32(&e.enableCache) == 0 {
		return e.Enforcer.Enforce(rvals...)
	}

	key, ok := e.getKey(rvals...)
	if !ok {
		return e.Enforcer.Enforce(rvals...)
	}

	if res, err := e.getCachedResult(key); err == nil {
		return res, nil
	} else if err != cache.ErrNoSuchKey {
		return res, err
	}

	res, err := e.Enforcer.Enforce(rvals...)
	if err != nil {
		return false, err
	}

	err = e.setCachedResult(key, res, e.expireTime)
	return res, err
}

// LoadPolicy reloads the policy from file/database and clears the cache.
func (e *CachedEnforcer) LoadPolicy() error {
	if atomic.LoadInt32(&e.enableCache) != 0 {
		if err := e.cache.Clear(); err != nil {
			return err
		}
	}
	return e.Enforcer.LoadPolicy()
}

// SetWatcher sets the current watcher for the CachedEnforcer.
// It overrides the base Enforcer.SetWatcher to ensure that:
// 1) For WatcherEx implementations (e.g., Redis watcher), a proper callback is set
//    that calls CachedEnforcer.InvalidateCache() + LoadPolicy() for efficient cache clearing.
// 2) For basic Watcher implementations, the callback calls LoadPolicy() which clears cache.
func (e *CachedEnforcer) SetWatcher(watcher persist.Watcher) error {
	e.watcher = watcher
	if _, ok := watcher.(persist.WatcherEx); ok {
		// For WatcherEx, set a callback that invalidates cache on any policy change.
		// The callback is invoked by the watcher implementation (e.g., Redis pub/sub subscriber)
		// when another instance modifies the policy.
		return watcher.SetUpdateCallback(func(string) {
			// First invalidate the cache to prevent stale reads
			if atomic.LoadInt32(&e.enableCache) != 0 {
				_ = e.InvalidateCache()
			}
			// Then reload the policy from the persistence layer
			_ = e.LoadPolicy()
		})
	}
	// For basic Watcher, the default callback is sufficient since
	// LoadPolicy() on CachedEnforcer already clears the cache.
	return watcher.SetUpdateCallback(func(string) { _ = e.LoadPolicy() })
}

// RemovePolicy removes an authorization rule from the current policy.
// It also removes the corresponding cache entry.
func (e *CachedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	if atomic.LoadInt32(&e.enableCache) != 0 {
		key, ok := e.getKey(params...)
		if ok {
			if err := e.cache.Delete(key); err != nil && err != cache.ErrNoSuchKey {
				return false, err
			}
		}
	}
	return e.Enforcer.RemovePolicy(params...)
}

func (e *CachedEnforcer) RemovePolicies(rules [][]string) (bool, error) {
	if len(rules) != 0 {
		if atomic.LoadInt32(&e.enableCache) != 0 {
			irule := make([]interface{}, len(rules[0]))
			for _, rule := range rules {
				for i, param := range rule {
					irule[i] = param
				}
				key, _ := e.getKey(irule...)
				if err := e.cache.Delete(key); err != nil && err != cache.ErrNoSuchKey {
					return false, err
				}
			}
		}
	}
	return e.Enforcer.RemovePolicies(rules)
}

func (e *CachedEnforcer) getCachedResult(key string) (res bool, err error) {
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.cache.Get(key)
}

// SetExpireTime sets the cache expiration time (TTL).
// Use 0 or negative duration to make cache entries never expire.
// This is useful in multi-instance scenarios where you want to avoid lock contention
// and recalculation overhead, and instead manually trigger LoadPolicy() or InvalidateCache()
// when policies change.
func (e *CachedEnforcer) SetExpireTime(expireTime time.Duration) {
	e.expireTime = expireTime
}

// SetCache sets the cache implementation.
func (e *CachedEnforcer) SetCache(c cache.Cache) {
	e.cache = c
}

func (e *CachedEnforcer) setCachedResult(key string, res bool, extra ...interface{}) error {
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.cache.Set(key, res, extra...)
}

func (e *CachedEnforcer) getKey(params ...interface{}) (string, bool) {
	return GetCacheKey(params...)
}

// InvalidateCache deletes all the existing cached decisions.
func (e *CachedEnforcer) InvalidateCache() error {
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.cache.Clear()
}

func GetCacheKey(params ...interface{}) (string, bool) {
	key := strings.Builder{}
	for _, param := range params {
		switch typedParam := param.(type) {
		case string:
			key.WriteString(typedParam)
		case CacheableParam:
			key.WriteString(typedParam.GetCacheKey())
		default:
			return "", false
		}
		key.WriteString("$$")
	}
	return key.String(), true
}

// ClearPolicy clears all policy and the cache.
func (e *CachedEnforcer) ClearPolicy() {
	if atomic.LoadInt32(&e.enableCache) != 0 {
		if err := e.cache.Clear(); err != nil {
			// Logger has been removed - error is ignored
			return
		}
	}
	e.Enforcer.ClearPolicy()
}
