// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package casbin

import (
	"context"
	"sync"
	"testing"
)

func TestSyncedContextEnforcer(t *testing.T) {
	e, err := NewSyncedContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewSyncedContextEnforcer failed: %v", err)
	}
	e.EnableAutoSave(false)

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := e.AddPolicyCtx(ctx, "eve", "data3", "read"); err != nil {
					t.Error(err)
				}
				if _, err := e.RemovePolicyCtx(ctx, "eve", "data3", "read"); err != nil {
					t.Error(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if err := e.LoadPolicyCtx(ctx); err != nil {
					t.Error(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := e.Enforce("alice", "data2", "read"); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()

	testEnforceSync(t, e.SyncedEnforcer, "alice", "data1", "read", true)
	testEnforceSync(t, e.SyncedEnforcer, "alice", "data2", "read", true)
	testEnforceSync(t, e.SyncedEnforcer, "eve", "data3", "read", false)
}
