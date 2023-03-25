package cache

import (
	"sync"
	"time"
)

type SyncCache struct {
	cache DefaultCache
	sync.RWMutex
}

func (c *SyncCache) Set(key string, value bool, extra ...interface{}) error {
	ttl := time.Duration(-1)
	if len(extra) > 0 {
		ttl = extra[0].(time.Duration)
	}
	(*c).Lock()
	defer (*c).Unlock()
	(*c).cache[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		ttl:       ttl,
	}
	return nil
}

func (c *SyncCache) Get(key string) (bool, error) {
	(*c).RLock()
	res, ok := (*c).cache[key]
	(*c).RUnlock()
	if !ok {
		return false, ErrNoSuchKey
	} else {
		if res.ttl > 0 && time.Now().After(res.expiresAt) {
			(*c).Lock()
			defer (*c).Unlock()
			delete((*c).cache, key)
			return false, ErrNoSuchKey
		}
		return res.value, nil
	}
}

func (c *SyncCache) Delete(key string) error {
	(*c).RLock()
	_, ok := (*c).cache[key]
	(*c).RUnlock()
	if !ok {
		return ErrNoSuchKey
	} else {
		(*c).Lock()
		defer (*c).Unlock()
		delete((*c).cache, key)
		return nil
	}
}

func (c *SyncCache) Clear() error {
	*c = SyncCache{
		make(DefaultCache),
		sync.RWMutex{},
	}
	return nil
}

func NewSyncCache() (Cache, error) {
	cache := SyncCache{
		make(DefaultCache),
		sync.RWMutex{},
	}
	return &cache, nil
}
