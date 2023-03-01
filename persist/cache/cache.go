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

import "errors"

var ErrNoSuchKey = errors.New("there's no such key existing in cache")

type Cache interface {
	// Set puts key and value into cache.
	// First parameter for extra should be time.Time object denoting expected survival time.
	// If survival time equals 0 or less, the key will always be survival.
	Set(key string, value bool, extra ...interface{}) error

	// Get returns result for key,
	// If there's no such key existing in cache,
	// ErrNoSuchKey will be returned.
	Get(key string) (bool, error)

	// Delete will remove the specific key in cache.
	// If there's no such key existing in cache,
	// ErrNoSuchKey will be returned.
	Delete(key string) error

	// Clear deletes all the items stored in cache.
	Clear() error
}
