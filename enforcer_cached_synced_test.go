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
	"sync"
	"testing"
	"time"
)

func testSyncEnforceCache(t *testing.T, e *SyncedCachedEnforcer, sub string, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, obj, act); myRes != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestSyncCache(t *testing.T) {
	e, _ := NewSyncedCachedEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	e.expireTime = time.Millisecond
	// The cache is enabled by default for NewCachedEnforcer.
	g := sync.WaitGroup{}
	goThread := 1000
	g.Add(goThread)
	for i := 0; i < goThread; i++ {
		go func() {
			_, _ = e.AddPolicy("alice", "data2", "read")
			testSyncEnforceCache(t, e, "alice", "data2", "read", true)
			if e.InvalidateCache() != nil {
				panic("never reached")
			}
			g.Done()
		}()
	}
	g.Wait()
	_, _ = e.RemovePolicy("alice", "data2", "read")

	testSyncEnforceCache(t, e, "alice", "data1", "read", true)
	time.Sleep(time.Millisecond * 2) // coverage for expire
	testSyncEnforceCache(t, e, "alice", "data1", "read", true)

	testSyncEnforceCache(t, e, "alice", "data1", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", false)
	testSyncEnforceCache(t, e, "alice", "data2", "write", false)
	// The cache is enabled, calling RemovePolicy, LoadPolicy or RemovePolicies will
	// also operate cached items.
	_, _ = e.RemovePolicy("alice", "data1", "read")

	testSyncEnforceCache(t, e, "alice", "data1", "read", false)
	testSyncEnforceCache(t, e, "alice", "data1", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", false)
	testSyncEnforceCache(t, e, "alice", "data2", "write", false)

	e, _ = NewSyncedCachedEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testSyncEnforceCache(t, e, "alice", "data1", "read", true)
	testSyncEnforceCache(t, e, "bob", "data2", "write", true)
	testSyncEnforceCache(t, e, "alice", "data2", "read", true)
	testSyncEnforceCache(t, e, "alice", "data2", "write", true)

	_, _ = e.RemovePolicies([][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	})

	testSyncEnforceCache(t, e, "alice", "data1", "read", false)
	testSyncEnforceCache(t, e, "bob", "data2", "write", false)
	testSyncEnforceCache(t, e, "alice", "data2", "read", true)
	testSyncEnforceCache(t, e, "alice", "data2", "write", true)
}
