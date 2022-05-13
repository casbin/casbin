// Copyright 2017 The casbin Authors. All Rights Reserved.
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
	"sync/atomic"
	"time"
)

// SyncedEnforcer wraps Enforcer and provides synchronized access
type SyncedEnforcer struct {
	*Enforcer
	stopAutoLoad    chan struct{}
	autoLoadRunning int32
}

// NewSyncedEnforcer creates a synchronized enforcer via file or DB.
func NewSyncedEnforcer(params ...interface{}) (*SyncedEnforcer, error) {
	e := &SyncedEnforcer{}
	var err error
	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}
	e.Enforcer.shouldLock = true

	e.stopAutoLoad = make(chan struct{}, 1)
	e.autoLoadRunning = 0
	return e, nil
}

// IsAutoLoadingRunning check if SyncedEnforcer is auto loading policies
func (e *SyncedEnforcer) IsAutoLoadingRunning() bool {
	return atomic.LoadInt32(&(e.autoLoadRunning)) != 0
}

// StartAutoLoadPolicy starts a go routine that will every specified duration call LoadPolicy
func (e *SyncedEnforcer) StartAutoLoadPolicy(d time.Duration) {
	// Don't start another goroutine if there is already one running
	if !atomic.CompareAndSwapInt32(&e.autoLoadRunning, 0, 1) {
		return
	}

	ticker := time.NewTicker(d)
	go func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt32(&(e.autoLoadRunning), int32(0))
		}()
		n := 1
		for {
			select {
			case <-ticker.C:
				// error intentionally ignored
				_ = e.LoadPolicy()
				// Uncomment this line to see when the policy is loaded.
				// log.Print("Load policy for time: ", n)
				n++
			case <-e.stopAutoLoad:
				return
			}
		}
	}()
}

// StopAutoLoadPolicy causes the go routine to exit.
func (e *SyncedEnforcer) StopAutoLoadPolicy() {
	if e.IsAutoLoadingRunning() {
		e.stopAutoLoad <- struct{}{}
	}
}

// LoadPolicy reloads the policy from file/database.
func (e *SyncedEnforcer) LoadPolicy() error {
	e.m.RLock()
	cleanedNewModel := e.model.Copy()
	e.m.RUnlock()
	newModel, err := e.prepareNewModel(cleanedNewModel)
	if err != nil {
		return err
	}

	e.m.Lock()
	defer e.m.Unlock()
	if e.autoBuildRoleLinks {
		if err := e.tryBuildingRoleLinksWithModel(newModel); err != nil {
			return err
		}
	}
	e.model = newModel
	return nil
}
