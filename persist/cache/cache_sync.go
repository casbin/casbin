// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	c.Lock()
	defer c.Unlock()
	c.cache[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		ttl:       ttl,
	}
	return nil
}

func (c *SyncCache) Get(key string) (bool, error) {
	c.RLock()
	res, ok := c.cache[key]
	c.RUnlock()
	if !ok {
		return false, ErrNoSuchKey
	} else {
		if res.ttl > 0 && time.Now().After(res.expiresAt) {
			c.Lock()
			defer c.Unlock()
			delete(c.cache, key)
			return false, ErrNoSuchKey
		}
		return res.value, nil
	}
}

func (c *SyncCache) Delete(key string) error {
	c.RLock()
	_, ok := c.cache[key]
	c.RUnlock()
	if !ok {
		return ErrNoSuchKey
	} else {
		c.Lock()
		defer c.Unlock()
		delete(c.cache, key)
		return nil
	}
}

func (c *SyncCache) Clear() error {
	c.Lock()
	c.cache = make(DefaultCache)
	c.Unlock()
	return nil
}

func NewSyncCache() (Cache, error) {
	cache := SyncCache{
		make(DefaultCache),
		sync.RWMutex{},
	}
	return &cache, nil
}
