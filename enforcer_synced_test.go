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
	"testing"
	"time"
)

func testEnforceSync(t *testing.T, e *SyncedEnforcer, sub string, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, obj, act); myRes != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestSync(t *testing.T) {
	e, _ := NewSyncedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	// Start reloading the policy every 200 ms.
	e.StartAutoLoadPolicy(time.Millisecond * 200)

	testEnforceSync(t, e, "alice", "data1", "read", true)
	testEnforceSync(t, e, "alice", "data1", "write", false)
	testEnforceSync(t, e, "alice", "data2", "read", false)
	testEnforceSync(t, e, "alice", "data2", "write", false)
	testEnforceSync(t, e, "bob", "data1", "read", false)
	testEnforceSync(t, e, "bob", "data1", "write", false)
	testEnforceSync(t, e, "bob", "data2", "read", false)
	testEnforceSync(t, e, "bob", "data2", "write", true)

	// Stop the reloading policy periodically.
	e.StopAutoLoadPolicy()
}

func TestStopAutoLoadPolicy(t *testing.T) {
	e, _ := NewSyncedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	e.StartAutoLoadPolicy(5 * time.Millisecond)
	if !e.IsAutoLoadingRunning() {
		t.Error("auto load is not running")
	}
	e.StopAutoLoadPolicy()
	// Need a moment, to exit goroutine
	time.Sleep(10 * time.Millisecond)
	if e.IsAutoLoadingRunning() {
		t.Error("auto load is still running")
	}
}
