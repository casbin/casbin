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

type DefaultCache map[string]bool

func (c *DefaultCache) Set(key string, value bool, extra ...interface{}) error {
	(*c)[key] = value
	return nil
}

func (c *DefaultCache) Get(key string) (bool, error) {
	if res, ok := (*c)[key]; !ok {
		return false, ErrNoSuchKey
	} else {
		return res, nil
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
