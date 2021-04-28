// Copyright 2020 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v2/model"
)

type SampleWatcherEx struct {
	SampleWatcher
}

func (w SampleWatcherEx) UpdateForAddPolicy(sec, ptype string, params ...string) error {
	return nil
}
func (w SampleWatcherEx) UpdateForRemovePolicy(sec, ptype string, params ...string) error {
	return nil
}

func (w SampleWatcherEx) UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func (w SampleWatcherEx) UpdateForSavePolicy(model model.Model) error {
	return nil
}

func TestSetWatcherEx(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	sampleWatcherEx := SampleWatcherEx{}
	err := e.SetWatcher(sampleWatcherEx)
	if err != nil {
		t.Fatal(err)
	}

	_ = e.SavePolicy()                              // calls watcherEx.UpdateForSavePolicy()
	_, _ = e.AddPolicy("admin", "data1", "read")    // calls watcherEx.UpdateForAddPolicy()
	_, _ = e.RemovePolicy("admin", "data1", "read") // calls watcherEx.UpdateForRemovePolicy()
	_, _ = e.RemoveFilteredPolicy(1, "data1")       // calls watcherEx.UpdateForRemoveFilteredPolicy()
	_, _ = e.AddGroupingPolicy("g:admin", "data1")
	_, _ = e.RemoveGroupingPolicy("g:admin", "data1")
	_, _ = e.AddGroupingPolicy("g:admin", "data1")
	_, _ = e.RemoveFilteredGroupingPolicy(1, "data1")
}
