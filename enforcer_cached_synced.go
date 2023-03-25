package casbin

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v2/persist/cache"
)

// SyncCachedEnforcer wraps Enforcer and provides decision sync cache
type SyncCachedEnforcer struct {
	*Enforcer
	expireTime  time.Duration
	cache       cache.Cache
	enableCache int32
	locker      *sync.RWMutex
}

// NewSyncedCachedEnforcer creates a sync cached enforcer via file or DB.
func NewSyncedCachedEnforcer(params ...interface{}) (*SyncCachedEnforcer, error) {
	e := &SyncCachedEnforcer{}
	var err error
	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	e.enableCache = 1
	e.cache, _ = cache.NewSyncCache()
	e.locker = new(sync.RWMutex)
	return e, nil
}

// EnableCache determines whether to enable cache on Enforce(). When enableCache is enabled, cached result (true | false) will be returned for previous decisions.
func (e *SyncCachedEnforcer) EnableCache(enableCache bool) {
	var enabled int32
	if enableCache {
		enabled = 1
	}
	atomic.StoreInt32(&e.enableCache, enabled)
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
// if rvals is not string , ingore the cache
func (e *SyncCachedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	if atomic.LoadInt32(&e.enableCache) == 0 {
		e.locker.RLock()
		defer e.locker.RUnlock()
		return e.Enforcer.Enforce(rvals...)
	}

	key, ok := e.getKey(rvals...)
	if !ok {
		e.locker.RLock()
		defer e.locker.RUnlock()
		return e.Enforcer.Enforce(rvals...)
	}

	if res, err := e.getCachedResult(key); err == nil {
		return res, nil
	} else if err != cache.ErrNoSuchKey {
		return res, err
	}

	e.locker.RLock()
	res, err := e.Enforcer.Enforce(rvals...)
	e.locker.RUnlock()
	if err != nil {
		return false, err
	}

	err = e.setCachedResult(key, res, e.expireTime)
	return res, err
}

func (e *SyncCachedEnforcer) LoadPolicy() error {
	if atomic.LoadInt32(&e.enableCache) != 0 {
		if err := e.cache.Clear(); err != nil {
			return err
		}
	}
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.Enforcer.LoadPolicy()
}

func (e *SyncCachedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	if atomic.LoadInt32(&e.enableCache) != 0 {
		key, ok := e.getKey(params...)
		if ok {
			if err := e.cache.Delete(key); err != nil && err != cache.ErrNoSuchKey {
				return false, err
			}
		}
	}
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.Enforcer.RemovePolicy(params...)
}

func (e *SyncCachedEnforcer) RemovePolicies(rules [][]string) (bool, error) {
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
	e.locker.Lock()
	defer e.locker.Unlock()
	return e.Enforcer.RemovePolicies(rules)
}

func (e *SyncCachedEnforcer) getCachedResult(key string) (res bool, err error) {
	return e.cache.Get(key)
}

func (e *SyncCachedEnforcer) SetExpireTime(expireTime time.Duration) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.expireTime = expireTime
}

func (e *SyncCachedEnforcer) SetCache(c cache.Cache) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.cache = c
}

func (e *SyncCachedEnforcer) setCachedResult(key string, res bool, extra ...interface{}) error {
	return e.cache.Set(key, res, extra...)
}

func (e *SyncCachedEnforcer) getKey(params ...interface{}) (string, bool) {
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

// InvalidateCache deletes all the existing cached decisions.
func (e *SyncCachedEnforcer) InvalidateCache() error {
	return e.cache.Clear()
}
