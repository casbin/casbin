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

import "time"

type cacheItem struct {
	value     bool
	expiresAt time.Time
	ttl       time.Duration
}

type DefaultCache map[string]cacheItem

func (c *DefaultCache) Set(key string, value bool, extra ...interface{}) error {
	ttl := time.Duration(-1)
	if len(extra) > 0 {
		ttl = extra[0].(time.Duration)
	}
	(*c)[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		ttl:       ttl,
	}
	return nil
}

func (c *DefaultCache) Get(key string) (bool, error) {
	if res, ok := (*c)[key]; !ok {
		return false, ErrNoSuchKey
	} else {
		if res.ttl > 0 && time.Now().After(res.expiresAt) {
			delete(*c, key)
			return false, ErrNoSuchKey
		}
		return res.value, nil
	}
}

func (c *DefaultCache) Delete(key string) error {
	if _, ok := (*c)[key]; !ok {
		return ErrNoSuchKey
	} else {
		delete(*c, key)
		return nil
	}
}

func (c *DefaultCache) Clear() error {
	*c = make(DefaultCache)
	return nil
}

func NewDefaultCache() (Cache, error) {
	cache := make(DefaultCache)
	return &cache, nil
}
